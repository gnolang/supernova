package distributor

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/gnoland"
	"github.com/gnolang/gno/pkgs/crypto/bip39"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/sdk/bank"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gnolang/supernova/internal/common"
	"github.com/stretchr/testify/assert"
)

// generateMnemonic generates a new BIP39 mnemonic
func generateMnemonic(t *testing.T) string {
	t.Helper()

	// Generate the entropy seed
	entropySeed, err := bip39.NewEntropy(256)
	if err != nil {
		t.Fatalf("unable to generate entropy seed, %v", err)
	}

	// Generate the actual mnemonic
	mnemonic, err := bip39.NewMnemonic(entropySeed[:])
	if err != nil {
		t.Fatalf("unable to generate mnemonic, %v", err)
	}

	return mnemonic
}

// generateAccounts generates mock keybase accounts
func generateAccounts(t *testing.T, count int) []keys.Info {
	t.Helper()

	kb := keys.NewInMemory()
	accounts := make([]keys.Info, count)
	mnemonic := generateMnemonic(t)

	for i := 0; i < count; i++ {
		info, err := kb.CreateAccount(
			fmt.Sprintf("%s%d", common.KeybasePrefix, i),
			mnemonic,
			"",
			common.EncryptPassword,
			uint32(0),
			uint32(i),
		)
		if err != nil {
			t.Fatalf("unable to create account with keybase, %v", err)
		}

		accounts[i] = info
	}

	return accounts
}

func TestDistributor_Distribute(t *testing.T) {
	t.Parallel()

	var (
		numTx      = uint64(1000)
		singleCost = calculateRuntimeCosts(int64(numTx))
	)

	getAccount := func(address string, accounts []keys.Info) keys.Info {
		for _, account := range accounts {
			if address == account.GetAddress().String() {
				return account
			}
		}

		return nil
	}

	t.Run("all accounts funded", func(t *testing.T) {
		t.Parallel()

		var (
			accounts = generateAccounts(t, 10)

			mockClient = &mockClient{
				getAccountFn: func(address string) (*gnoland.GnoAccount, error) {
					acc := getAccount(address, accounts)
					if acc == nil {
						t.Fatal("invalid account requested")
					}

					return &gnoland.GnoAccount{
						BaseAccount: *std.NewBaseAccount(
							acc.GetAddress(),
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
			&mockSigner{},
		)

		readyAccounts, err := d.Distribute(accounts, numTx)
		if err != nil {
			t.Fatalf("unable to distribute funds, %v", err)
		}

		// Make sure all accounts are funded
		// (the distributor does not participate in the run, hence -1)
		assert.Len(t, readyAccounts, len(accounts)-1)

		// Make sure the accounts match
		for index, account := range accounts[1:] {
			assert.Equal(t, account.GetAddress().String(), readyAccounts[index].GetAddress().String())
		}
	})

	t.Run("insufficient distributor funds", func(t *testing.T) {
		t.Parallel()

		emptyBalance := std.Coin{
			Denom:  common.Denomination,
			Amount: 0,
		}

		var (
			accounts = generateAccounts(t, 10)

			mockClient = &mockClient{
				getAccountFn: func(address string) (*gnoland.GnoAccount, error) {
					acc := getAccount(address, accounts)
					if acc == nil {
						t.Fatal("invalid account requested")
					}

					return &gnoland.GnoAccount{
						BaseAccount: *std.NewBaseAccount(
							acc.GetAddress(),
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
			&mockSigner{},
		)

		readyAccounts, err := d.Distribute(accounts, numTx)

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
			accounts           = generateAccounts(t, 10)
			capturedBroadcasts = make([]*std.Tx, 0)
			capturedNonce      = uint64(0)

			mockClient = &mockClient{
				getAccountFn: func(address string) (*gnoland.GnoAccount, error) {
					acc := getAccount(address, accounts)
					if acc == nil {
						t.Fatal("invalid account requested")
					}

					if acc.GetName() == fmt.Sprintf("%s%d", common.KeybasePrefix, 0) {
						return &gnoland.GnoAccount{
							BaseAccount: *std.NewBaseAccount(
								acc.GetAddress(),
								std.NewCoins(std.Coin{
									Denom:  common.Denomination,
									Amount: int64(numTx) * common.DefaultGasFee.Add(singleCost).Amount,
								}),
								nil,
								0,
								0,
							),
						}, nil
					}

					return &gnoland.GnoAccount{
						BaseAccount: *std.NewBaseAccount(
							acc.GetAddress(),
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
			mockSigner = &mockSigner{
				signTxFn: func(
					tx *std.Tx,
					account *gnoland.GnoAccount,
					nonce uint64,
					passphrase string,
				) error {
					if acc := getAccount(account.GetAddress().String(), accounts); acc == nil {
						t.Fatal("invalid account")
					}

					if passphrase != common.EncryptPassword {
						t.Fatal("invalid passphrase")
					}

					if nonce != capturedNonce {
						t.Fatal("nonce not incrementing")
					}

					capturedNonce++

					return nil
				},
			}
		)

		d := NewDistributor(
			mockClient,
			mockSigner,
		)

		readyAccounts, err := d.Distribute(accounts, numTx)
		if err != nil {
			t.Fatalf("unable to distribute funds, %v", err)
		}

		// Make sure all accounts are funded
		// (the distributor does not participate in the run, hence -1)
		assert.Len(t, readyAccounts, len(accounts)-1)

		// Make sure the accounts match
		for index, account := range accounts[1:] {
			assert.Equal(t, account.GetAddress().String(), readyAccounts[index].GetAddress().String())
		}

		// Check the broadcast transactions
		if len(capturedBroadcasts) != len(accounts)-1 {
			t.Fatal("invalid number of transactions broadcast")
		}

		sendType := bank.MsgSend{}.Type()
		for _, tx := range capturedBroadcasts {
			if len(tx.Msgs) != 1 {
				t.Fatal("invalid number of messages")
			}

			assert.Equal(t, sendType, tx.Msgs[0].Type())
		}
	})
}
