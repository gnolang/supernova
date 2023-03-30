package internal

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/gnolang/gno/pkgs/bft/rpc/client"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/supernova/internal/batcher"
	"github.com/gnolang/supernova/internal/common"
	"github.com/gnolang/supernova/internal/distributor"
	"github.com/gnolang/supernova/internal/runtime"
	"github.com/gnolang/supernova/internal/signer"
	"github.com/schollz/progressbar/v3"
)

var (
	errInvalidURL          = errors.New("invalid node URL specified")
	errInvalidMnemonic     = errors.New("invalid Mnemonic specified")
	errInvalidMode         = errors.New("invalid mode specified")
	errInvalidSubaccounts  = errors.New("invalid number of subaccounts specified")
	errInvalidTransactions = errors.New("invalid number of transactions specified")
	errInvalidBatchSize    = errors.New("invalid batch size specified")
)

var (
	urlRegex = regexp.MustCompile(`(https?://.*)(:(\d*)\/?(.*))?`)
)

type Pipeline struct {
	ctx context.Context

	cfg     *Config
	keybase keys.Keybase
	cli     client.Client
}

// NewPipeline creates a new pipeline instance
func NewPipeline(ctx context.Context, cfg *Config) *Pipeline {
	return &Pipeline{
		ctx:     ctx,
		cfg:     cfg,
		keybase: keys.NewInMemory(),
		cli:     client.NewHTTP(cfg.URL, ""),
	}
}

func (p *Pipeline) Execute() error {
	// Register the accounts with the keybase
	fmt.Printf("\nGenerating sub-accounts...\n")

	accounts := make([]keys.Info, p.cfg.SubAccounts+1)
	bar := progressbar.Default(int64(p.cfg.SubAccounts+1), "sub-accounts added")

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
			return fmt.Errorf("unable to create account with keybase, %w", err)
		}

		accounts[i] = info
		_ = bar.Add(1)
	}

	var (
		txSigner       = signer.NewKeybaseSigner(p.keybase, p.cfg.ChainID)
		requestBatcher = batcher.NewBatcher(client.NewHTTP(p.cfg.URL, ""))
		accountStore   = newStore(p.cli)
		txBroadcaster  = newBroadcaster(p.cli)
	)

	setRuntime := runtime.GetRuntime(runtime.Type(p.cfg.Mode), txSigner)

	if runtime.Type(p.cfg.Mode) == runtime.RealmCall {
		fmt.Printf("\n✨ Starting Predeployment Procedure ✨\n\n")

		// Get the deployer account
		deployer, err := accountStore.GetAccount(accounts[0].GetAddress().String())
		if err != nil {
			return fmt.Errorf("unable to fetch deployer account, %w", err)
		}

		// Get the predeploy transactions
		predeployTxs, err := setRuntime.Initialize(deployer)
		if err != nil {
			return fmt.Errorf("unable to initialize runtime, %w", err)
		}

		bar := progressbar.Default(int64(len(predeployTxs)), "predeployed txs")

		// Execute the predeploy transactions
		for _, tx := range predeployTxs {
			if err := txBroadcaster.BroadcastTxWithCommit(tx); err != nil {
				return fmt.Errorf("unable to broadcast predeploy tx, %w", err)
			}

			_ = bar.Add(1)
		}

		fmt.Printf("✅ Successfully predeployed %d transactions\n", len(predeployTxs))
	}

	// Distribution //

	fmt.Printf("\n💸 Starting Fund Distribution 💸\n\n")

	runAccounts, err := distributor.NewDistributor(
		txBroadcaster,
		accountStore,
		txSigner,
	).Distribute(
		accounts,
		p.cfg.Transactions,
	)
	if err != nil {
		return fmt.Errorf("unable to distribute funds, %w", err)
	}

	// Runtime //

	fmt.Printf("\n🔨 Constructing Transactions 🔨\n\n")

	txs, err := setRuntime.ConstructTransactions(runAccounts, p.cfg.Transactions)
	if err != nil {
		return fmt.Errorf("unable to construct transactions, %w", err)
	}

	// Batcher //
	fmt.Printf("\n📦 Batching Transactions 📦\n\n")

	_, err = requestBatcher.BatchTransactions(txs, int(p.cfg.BatchSize))
	if err != nil {
		return fmt.Errorf("unable to batch transactions %w", err)
	}

	// Collector //

	// TODO

	return nil
}
