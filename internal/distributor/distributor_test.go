package distributor

import (
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/supernova/internal/common"
	testutils "github.com/gnolang/supernova/internal/testing"
	"github.com/stretchr/testify/assert"
)

func TestDistributor_Distribute(t *testing.T) {
	t.Parallel()

	var (
		numTx      = uint64(1000)
		singleCost = std.Coin{
			Denom:  common.Denomination,
			Amount: int64(numTx) * 100_00,
		}
	)

	getAccount := func(address string, accounts []crypto.PrivKey) crypto.PrivKey {
		for _, account := range accounts {
			if address == account.PubKey().Address().String() {
				return account
			}
		}

		return nil
	}

	t.Run("all accounts funded", func(t *testing.T) {
		t.Parallel()

		var (
			accounts = testutils.GenerateAccounts(t, 10)

			mockClient = &mockClient{
				getAccountFn: func(address string) (*gnoland.GnoAccount, error) {
					acc := getAccount(address, accounts)
					if acc == nil {
						t.Fatal("invalid account requested")
					}

					return &gnoland.GnoAccount{
						BaseAccount: *std.NewBaseAccount(
							acc.PubKey().Address(),
							std.NewCoins(singleCost),
							nil,
							0,
							0,
						),
					}, nil
				},
			}
		)

		d := NewDistributor(
			mockClient,
		)

		// Extract the addresses
		addresses := make([]crypto.Address, 0, len(accounts[1:]))
		for _, account := range accounts[1:] {
			addresses = append(addresses, account.PubKey().Address())
		}

		readyAccounts, err := d.Distribute(accounts[0], addresses, "dummy", common.DefaultGasPrice, singleCost)
		if err != nil {
			t.Fatalf("unable to distribute funds, %v", err)
		}

		// Make sure all accounts are funded
		// (the distributor does not participate in the run, hence -1)
		assert.Len(t, readyAccounts, len(accounts)-1)

		// Make sure the accounts match
		for index, account := range accounts[1:] {
			assert.Equal(t, account.PubKey().Address().String(), readyAccounts[index].GetAddress().String())
		}
	})

	t.Run("insufficient distributor funds", func(t *testing.T) {
		t.Parallel()

		emptyBalance := std.Coin{
			Denom:  common.Denomination,
			Amount: 0,
		}

		var (
			accounts = testutils.GenerateAccounts(t, 10)

			mockClient = &mockClient{
				getAccountFn: func(address string) (*gnoland.GnoAccount, error) {
					acc := getAccount(address, accounts)
					if acc == nil {
						t.Fatal("invalid account requested")
					}

					return &gnoland.GnoAccount{
						BaseAccount: *std.NewBaseAccount(
							acc.PubKey().Address(),
							std.NewCoins(emptyBalance),
							nil,
							0,
							0,
						),
					}, nil
				},
			}
		)

		d := NewDistributor(
			mockClient,
		)

		// Extract the addresses
		addresses := make([]crypto.Address, 0, len(accounts[1:]))
		for _, account := range accounts[1:] {
			addresses = append(addresses, account.PubKey().Address())
		}

		readyAccounts, err := d.Distribute(accounts[0], addresses, "dummy", common.DefaultGasPrice, singleCost)

		assert.Nil(t, readyAccounts)
		assert.ErrorIs(t, err, errInsufficientFunds)
	})

	t.Run("fund all short accounts", func(t *testing.T) {
		t.Parallel()

		emptyBalance := std.Coin{
			Denom:  common.Denomination,
			Amount: 0,
		}

		var (
			accounts           = testutils.GenerateAccounts(t, 10)
			capturedBroadcasts = make([]*std.Tx, 0)

			mockClient = &mockClient{
				getAccountFn: func(address string) (*gnoland.GnoAccount, error) {
					acc := getAccount(address, accounts)
					if acc == nil {
						t.Fatal("invalid account requested")
					}

					if acc.Equals(accounts[0]) {
						sendCost := common.CalculateFeeInRatio(100_000, common.DefaultGasPrice)

						return &gnoland.GnoAccount{
							BaseAccount: *std.NewBaseAccount(
								acc.PubKey().Address(),
								std.NewCoins(std.Coin{
									Denom:  common.Denomination,
									Amount: int64(numTx) * sendCost.GasFee.Add(singleCost).Amount,
								}),
								nil,
								0,
								0,
							),
						}, nil
					}

					return &gnoland.GnoAccount{
						BaseAccount: *std.NewBaseAccount(
							acc.PubKey().Address(),
							std.NewCoins(emptyBalance),
							nil,
							0,
							0,
						),
					}, nil
				},
				broadcastTransactionFn: func(tx *std.Tx) error {
					capturedBroadcasts = append(capturedBroadcasts, tx)

					return nil
				},
			}
		)

		d := NewDistributor(
			mockClient,
		)

		// Extract the addresses
		addresses := make([]crypto.Address, 0, len(accounts[1:]))
		for _, account := range accounts[1:] {
			addresses = append(addresses, account.PubKey().Address())
		}

		readyAccounts, err := d.Distribute(accounts[0], addresses, "dummy", common.DefaultGasPrice, singleCost)
		if err != nil {
			t.Fatalf("unable to distribute funds, %v", err)
		}

		// Make sure all accounts are funded
		// (the distributor does not participate in the run, hence -1)
		assert.Len(t, readyAccounts, len(accounts)-1)

		// Make sure the accounts match
		for index, account := range accounts[1:] {
			assert.Equal(t, account.PubKey().Address(), readyAccounts[index].GetAddress())
		}

		// Check the broadcast transactions
		assert.Len(t, capturedBroadcasts, len(accounts)-1)

		sendType := bank.MsgSend{}.Type()

		for _, tx := range capturedBroadcasts {
			if len(tx.Msgs) != 1 {
				t.Fatal("invalid number of messages")
			}

			assert.Equal(t, sendType, tx.Msgs[0].Type())
		}
	})
}
