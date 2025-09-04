package distributor

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/supernova/internal/common"
	"github.com/gnolang/supernova/internal/signer"
	"github.com/schollz/progressbar/v3"
)

var errInsufficientFunds = errors.New("insufficient distributor funds")

type Client interface {
	GetAccount(ctx context.Context, address string) (*gnoland.GnoAccount, error)
	BroadcastTransaction(ctx context.Context, tx *std.Tx) error
	EstimateGas(ctx context.Context, tx *std.Tx) (int64, error)
	FetchGasPrice(ctx context.Context) (std.GasPrice, error)
}

// Distributor is the process
// that manages sub-account distributions
type Distributor struct {
	cli Client
	ctx context.Context
}

// NewDistributor creates a new instance of the distributor
func NewDistributor(
	ctx context.Context,
	cli Client,
) *Distributor {
	return &Distributor{
		cli: cli,
		ctx: ctx,
	}
}

// Distribute distributes the funds from the base account
// (account 0 in the mnemonic) to other subaccounts
func (d *Distributor) Distribute(
	distributor crypto.PrivKey,
	accounts []crypto.Address,
	chainID string,
	gasPrice std.GasPrice,
	calculatedRuntimeCost std.Coin,
) ([]std.Account, error) {
	fmt.Printf("\n💸 Starting Fund Distribution 💸\n\n")

	fmt.Printf(
		"Calculated sub-account cost as %d %s\n",
		calculatedRuntimeCost.Amount,
		calculatedRuntimeCost.Denom,
	)

	// Fund the accounts
	return d.fundAccounts(distributor, accounts, calculatedRuntimeCost, chainID, gasPrice)
}

// fundAccounts attempts to fund accounts that have missing funds,
// and returns the accounts that can participate in the stress test
func (d *Distributor) fundAccounts(
	distributorKey crypto.PrivKey,
	accounts []crypto.Address,
	singleRunCost std.Coin,
	chainID string,
	gasPrice std.GasPrice,
) ([]std.Account, error) {
	type shortAccount struct {
		missingFunds std.Coin
		address      crypto.Address
	}

	var (
		// Accounts that are ready (funded) for the run
		readyAccounts = make([]std.Account, 0, len(accounts))

		// Accounts that need funding
		shortAccounts = make([]shortAccount, 0, len(accounts))
	)

	// Check if there are any accounts that need to be funded
	// before the stress test starts
	for _, account := range accounts {
		// Fetch the account balance
		subAccount, err := d.cli.GetAccount(d.ctx, account.String())
		if err != nil {
			return nil, fmt.Errorf("unable to fetch sub-account, %w", err)
		}

		// Check if it has enough funds for the run
		if subAccount.Coins.AmountOf(common.Denomination) < singleRunCost.Amount {
			// Mark the account as needing a top-up
			shortAccounts = append(shortAccounts, shortAccount{
				address: account,
				missingFunds: std.Coin{
					Denom:  common.Denomination,
					Amount: singleRunCost.Amount - subAccount.Coins.AmountOf(common.Denomination),
				},
			})

			continue
		}

		// The account is cleared for the stress test
		readyAccounts = append(readyAccounts, subAccount)
	}

	// Check if funding is even necessary
	if len(shortAccounts) == 0 {
		// All accounts are already funded
		fmt.Printf("✅ All %d accounts are already funded\n", len(readyAccounts))

		return readyAccounts, nil
	}

	// Sort the short accounts so the ones with
	// the lowest missing funds are funded first
	sort.Slice(shortAccounts, func(i, j int) bool {
		return shortAccounts[i].missingFunds.IsLT(shortAccounts[j].missingFunds)
	})

	// Figure out how many accounts can actually be funded
	distributor, err := d.cli.GetAccount(d.ctx, distributorKey.PubKey().Address().String())
	if err != nil {
		return nil, fmt.Errorf("unable to fetch distributor account, %w", err)
	}

	var (
		distributorBalance = distributor.Coins
		fundableIndex      = 0
		defaultFee         = common.CalculateFeeInRatio(100_000, gasPrice)
	)

	for _, account := range shortAccounts {
		// The transfer cost is the single run cost (missing balance) + approximate transfer cost
		transferCost := std.NewCoins(defaultFee.GasFee.Add(account.missingFunds))

		if distributorBalance.IsAllLT(transferCost) {
			// Distributor does not have any more funds
			// to cover the run cost
			break
		}

		fundableIndex++

		distributorBalance.Sub(transferCost)
	}

	if fundableIndex == 0 {
		// The distributor does not have funds to fund
		// any account for the stress test
		fmt.Printf(
			"❌ Distributor cannot fund any account, balance is %d %s\n",
			distributorBalance.AmountOf(common.Denomination),
			common.Denomination,
		)

		return nil, errInsufficientFunds
	}

	// Locally keep track of the nonce, so
	// there is no need to re-fetch the account again
	// before signing a future tx
	nonce := distributor.Sequence

	fmt.Printf("Funding %d accounts...\n", len(shortAccounts))
	bar := progressbar.Default(int64(len(shortAccounts)), "funding short accounts")

	for _, account := range shortAccounts {
		// Generate the transaction
		tx := &std.Tx{
			Msgs: []std.Msg{
				bank.MsgSend{
					FromAddress: distributor.GetAddress(),
					ToAddress:   account.address,
					Amount:      std.NewCoins(account.missingFunds),
				},
			},
			Fee: defaultFee,
		}

		cfg := signer.SignCfg{
			ChainID:       chainID,
			AccountNumber: distributor.AccountNumber,
			Sequence:      nonce,
		}

		// Sign the transaction
		if err := signer.SignTx(tx, distributorKey, cfg); err != nil {
			return nil, fmt.Errorf("unable to sign transaction, %w", err)
		}

		// Update the local nonce
		nonce++

		// Broadcast the tx and wait for it to be committed
		if err := d.cli.BroadcastTransaction(d.ctx, tx); err != nil {
			return nil, fmt.Errorf("unable to broadcast tx with commit, %w", err)
		}

		// Since accounts can be uninitialized on the node, after the
		// transfer they will have acquired a storage slot, and need
		// to be re-fetched for their data (Sequence + Account Number)
		nodeAccount, err := d.cli.GetAccount(d.ctx, account.address.String())
		if err != nil {
			return nil, fmt.Errorf("unable to fetch account, %w", err)
		}

		// Mark the account as funded
		readyAccounts = append(readyAccounts, nodeAccount)

		_ = bar.Add(1) //nolint:errcheck // No need to check
	}

	fmt.Printf("✅ Successfully funded %d accounts\n", len(shortAccounts))

	return readyAccounts, nil
}
