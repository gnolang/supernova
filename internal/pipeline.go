package internal

import (
	"fmt"
	"time"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/supernova/internal/batcher"
	"github.com/gnolang/supernova/internal/client"
	"github.com/gnolang/supernova/internal/collector"
	"github.com/gnolang/supernova/internal/distributor"
	"github.com/gnolang/supernova/internal/runtime"
	"github.com/gnolang/supernova/internal/signer"
	"github.com/schollz/progressbar/v3"
)

type pipelineClient interface {
	distributor.Client
	batcher.Client
	collector.Client
}

// Pipeline is the central run point
// for the stress test
type Pipeline struct {
	cfg *Config        // the run configuration
	cli pipelineClient // HTTP client connection
}

// NewPipeline creates a new pipeline instance
func NewPipeline(cfg *Config) (*Pipeline, error) {
	var (
		cli *client.Client
		err error
	)

	// Check which kind of client to create
	if httpRegex.MatchString(cfg.URL) {
		cli, err = client.NewHTTPClient(cfg.URL)
	} else {
		cli, err = client.NewWSClient(cfg.URL)
	}

	if err != nil {
		return nil, fmt.Errorf("unable to create RPC client, %w", err)
	}

	return &Pipeline{
		cfg: cfg,
		cli: cli,
	}, nil
}

// Execute runs the entire pipeline process
func (p *Pipeline) Execute() error {
	var (
		mode = runtime.Type(p.cfg.Mode)

		txBatcher   = batcher.NewBatcher(p.cli)
		txCollector = collector.NewCollector(p.cli)
		txRuntime   = runtime.GetRuntime(mode)
	)

	// Initialize the accounts for the runtime
	accounts := p.initializeAccounts()

	// Predeploy any pending transactions
	if err := prepareRuntime(
		mode,
		accounts[0],
		p.cfg.ChainID,
		p.cli,
		txRuntime,
	); err != nil {
		return err
	}

	// Extract the addresses
	addresses := make([]crypto.Address, 0, len(accounts[1:]))
	for _, account := range accounts[1:] {
		addresses = append(addresses, account.PubKey().Address())
	}

	// Distribute the funds to sub-accounts
	runAccounts, err := distributor.NewDistributor(p.cli).Distribute(
		accounts[0],
		addresses,
		p.cfg.Transactions,
		p.cfg.ChainID,
	)
	if err != nil {
		return fmt.Errorf("unable to distribute funds, %w", err)
	}

	// Find which keys belong to the run accounts (not all initial accounts are run accounts)
	runKeys := make([]crypto.PrivKey, 0, len(runAccounts))

	for _, runAccount := range runAccounts {
		for _, account := range accounts[1:] {
			if account.PubKey().Address() == runAccount.GetAddress() {
				runKeys = append(runKeys, account)
			}
		}
	}

	// Construct the transactions using the runtime
	txs, err := txRuntime.ConstructTransactions(
		runKeys,
		runAccounts,
		p.cfg.Transactions,
		p.cfg.ChainID,
		p.cli.EstimateGas,
	)
	if err != nil {
		return fmt.Errorf("unable to construct transactions, %w", err)
	}

	// Send the signed transactions in batches
	batchStart := time.Now()

	batchResult, err := txBatcher.BatchTransactions(txs, int(p.cfg.BatchSize))
	if err != nil {
		return fmt.Errorf("unable to batch transactions %w", err)
	}

	// Collect the transaction results
	runResult, err := txCollector.GetRunResult(
		batchResult.TxHashes,
		batchResult.StartBlock,
		batchStart,
	)
	if err != nil {
		return fmt.Errorf("unable to collect transactions, %w", err)
	}

	// Display [+ save the results]
	return p.handleResults(runResult)
}

// initializeAccounts initializes the accounts needed for the stress test run
func (p *Pipeline) initializeAccounts() []crypto.PrivKey {
	fmt.Printf("\n🧮 Initializing Accounts 🧮\n\n")
	fmt.Printf("Generating sub-accounts...\n")

	var (
		accounts = make([]crypto.PrivKey, p.cfg.SubAccounts+1)
		bar      = progressbar.Default(int64(p.cfg.SubAccounts+1), "accounts initialized")

		seed = bip39.NewSeed(p.cfg.Mnemonic, "")
	)

	// Register the accounts with the keybase
	for i := 0; i < int(p.cfg.SubAccounts)+1; i++ {
		accounts[i] = signer.GenerateKeyFromSeed(seed, uint32(i))
		_ = bar.Add(1) //nolint:errcheck // No need to check
	}

	fmt.Printf("✅ Successfully generated %d accounts\n", len(accounts))

	return accounts
}

// handleResults displays the results in the terminal,
// and saves them to disk if an output path was specified
func (p *Pipeline) handleResults(runResult *collector.RunResult) error {
	// Display the results in the terminal
	displayResults(runResult)

	// Check if the results need to be saved to disk
	if p.cfg.Output == "" {
		// No disk save necessary
		return nil
	}

	fmt.Printf("\n💾 Saving Results 💾\n\n")

	if err := saveResults(runResult, p.cfg.Output); err != nil {
		return fmt.Errorf("unable to save results, %w", err)
	}

	fmt.Printf("✅ Successfully saved results to %s\n", p.cfg.Output)

	return nil
}

// prepareRuntime prepares the runtime by pre-deploying
// any pending transactions
func prepareRuntime(
	mode runtime.Type,
	deployerKey crypto.PrivKey,
	chainID string,
	cli pipelineClient,
	txRuntime runtime.Runtime,
) error {
	if mode != runtime.RealmCall {
		return nil
	}

	fmt.Printf("\n✨ Starting Predeployment Procedure ✨\n\n")

	// Get the deployer account
	deployer, err := cli.GetAccount(deployerKey.PubKey().Address().String())
	if err != nil {
		return fmt.Errorf("unable to fetch deployer account, %w", err)
	}

	// Get the predeploy transactions
	predeployTxs, err := txRuntime.Initialize(
		deployer,
		deployerKey,
		chainID,
		cli.EstimateGas,
	)
	if err != nil {
		return fmt.Errorf("unable to initialize runtime, %w", err)
	}

	bar := progressbar.Default(int64(len(predeployTxs)), "predeployed txs")

	// Execute the predeploy transactions
	for _, tx := range predeployTxs {
		if err := cli.BroadcastTransaction(tx); err != nil {
			return fmt.Errorf("unable to broadcast predeploy tx, %w", err)
		}

		_ = bar.Add(1) //nolint:errcheck // No need to check
	}

	fmt.Printf("✅ Successfully predeployed %d transactions\n", len(predeployTxs))

	return nil
}
