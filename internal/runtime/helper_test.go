package runtime

import (
	"testing"

	"github.com/gnolang/gno/gnoland"
	"github.com/gnolang/gno/pkgs/sdk/vm"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/stretchr/testify/assert"
)

// generateAccounts generates mock gno accounts
func generateAccounts(count int) []*gnoland.GnoAccount {
	accounts := make([]*gnoland.GnoAccount, count)

	for i := 0; i < count; i++ {
		accounts[i] = &gnoland.GnoAccount{
			BaseAccount: std.BaseAccount{
				AccountNumber: uint64(i),
			},
		}
	}

	return accounts
}

func TestHelper_ConstructTransactions(t *testing.T) {
	t.Parallel()

	accounts := generateAccounts(10)
	nonceMap := make(map[uint64]uint64, len(accounts))

	// Initialize the nonce map
	for _, account := range accounts {
		nonceMap[account.AccountNumber] = 0
	}

	var (
		transactions  = uint64(100)
		capturedSigns = make([]*std.Tx, 0, transactions)
		msg           = vm.MsgAddPackage{}

		mockSigner = &mockSigner{
			signTxFn: func(tx *std.Tx, account *gnoland.GnoAccount, nonce uint64, _ string) error {
				// Make sure the nonce is being incremented
				if nonce != nonceMap[account.AccountNumber] {
					t.Fatalf("invalid nonce for account")
				}

				capturedSigns = append(capturedSigns, tx)
				nonceMap[account.AccountNumber]++

				return nil
			},
		}
		getMsgFn = func(creator *gnoland.GnoAccount, index int) std.Msg {
			return msg
		}
	)

	txs, err := constructTransactions(mockSigner, accounts, transactions, getMsgFn)
	if err != nil {
		t.Fatalf("unable to construct transactions, %v", err)
	}

	if len(txs) != int(transactions) {
		t.Fatalf("invalid number of transactions, %d", len(txs))
	}

	// Make sure each transaction was signed
	if len(capturedSigns) != int(transactions) {
		t.Fatalf("invalid number of transactions signed, %d", len(txs))
	}

	// Make sure the constructed transactions are valid
	for index, tx := range txs {
		// Make sure the fee is valid
		assert.Equal(t, defaultDeployTxFee, tx.Fee)

		// Make sure the message is valid
		if len(tx.Msgs) != 1 {
			t.Fatalf("invalid number of transaction messages, %d", len(tx.Msgs))
		}

		assert.Equal(t, msg, tx.Msgs[0])

		assert.Equal(t, capturedSigns[index], tx)
	}
}
