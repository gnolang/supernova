package internal

import (
	"fmt"
	"time"

	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/supernova/internal/batcher"
	"github.com/gnolang/supernova/internal/client"
	"github.com/gnolang/supernova/internal/collector"
	"github.com/gnolang/supernova/internal/common"
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

type pipelineSigner interface {
	distributor.Signer
}

// Pipeline is the central run point
// for the stress test
type Pipeline struct {
	cfg *Config // the run configuration

	keybase keys.Keybase   // relevant keybase
	cli     pipelineClient // HTTP client connection
	signer  pipelineSigner // the transaction signer
}

// NewPipeline creates a new pipeline instance
func NewPipeline(cfg *Config) *Pipeline {
	kb := keys.NewInMemory()

	return &Pipeline{
		cfg:     cfg,
		keybase: kb,
		cli:     client.NewHTTPClient(cfg.URL),
		signer:  signer.NewKeybaseSigner(kb, cfg.ChainID),
	}
}

// Execute runs the entire pipeline process
func (p *Pipeline) Execute() error {
	var (
		mode = runtime.Type(p.cfg.Mode)

		txBatcher   = batcher.NewBatcher(p.cli)
		txCollector = collector.NewCollector(p.cli)
		txRuntime   = runtime.GetRuntime(mode, p.signer)
	)

	// Initialize the accounts for the runtime
	accounts, err := p.initializeAccounts()
	if err != nil {
		return err
	}

	// Predeploy any pending transactions
	if err := prepareRuntime(mode, accounts, p.cli, txRuntime); err != nil {
		return err
	}

	// Distribute the funds to sub-accounts
	runAccounts, err := distributor.NewDistributor(p.cli, p.signer).Distribute(
		accounts,
		p.cfg.Transactions,
	)
	if err != nil {
		return fmt.Errorf("unable to distribute funds, %w", err)
	}

	// Construct the transactions using the runtime
	txs, err := txRuntime.ConstructTransactions(runAccounts, p.cfg.Transactions)
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
func (p *Pipeline) initializeAccounts() ([]keys.Info, error) {
	fmt.Printf("\n🧮 Initializing Accounts 🧮\n\n")

	fmt.Printf("Generating sub-accounts...\n")

	var (
		accounts = make([]keys.Info, p.cfg.SubAccounts+1)
		bar      = progressbar.Default(int64(p.cfg.SubAccounts+1), "accounts initialized")
	)

	// Register the accounts with the keybase
	for i := 0; i < int(p.cfg.SubAccounts)+1; i++ {
		info, err := p.keybase.CreateAccount(
			fmt.Sprintf("%s%d", common.KeybasePrefix, i),
			p.cfg.Mnemonic,
			"",
			common.EncryptPassword,
			uint32(0),
			uint32(i),
		)
		if err != nil {
			return nil, fmt.Errorf("unable to create account with keybase, %w", err)
		}

		accounts[i] = info
		_ = bar.Add(1)
	}

	fmt.Printf("✅ Successfully generated %d accounts\n", len(accounts))

	return accounts, nil
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
	accounts []keys.Info,
	cli pipelineClient,
	txRuntime runtime.Runtime,
) error {
	if mode != runtime.RealmCall {
		return nil
	}

	fmt.Printf("\n✨ Starting Predeployment Procedure ✨\n\n")

	// Get the deployer account
	deployer, err := cli.GetAccount(accounts[0].GetAddress().String())
	if err != nil {
		return fmt.Errorf("unable to fetch deployer account, %w", err)
	}

	// Get the predeploy transactions
	predeployTxs, err := txRuntime.Initialize(deployer)
	if err != nil {
		return fmt.Errorf("unable to initialize runtime, %w", err)
	}

	bar := progressbar.Default(int64(len(predeployTxs)), "predeployed txs")

	// Execute the predeploy transactions
	for _, tx := range predeployTxs {
		if err := cli.BroadcastTransaction(tx); err != nil {
			return fmt.Errorf("unable to broadcast predeploy tx, %w", err)
		}

		_ = bar.Add(1)
	}

	fmt.Printf("✅ Successfully predeployed %d transactions\n", len(predeployTxs))

	return nil
}
