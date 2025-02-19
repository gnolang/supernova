package runtime

import (
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/supernova/internal/common"
	testutils "github.com/gnolang/supernova/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateAccounts generates mock gno accounts
func generateAccounts(count int) []std.Account {
	accounts := make([]std.Account, count)

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

	var (
		accounts    = generateAccounts(10)
		accountKeys = testutils.GenerateAccounts(t, 10)
		nonceMap    = make(map[uint64]uint64, len(accounts))
	)

	// Initialize the nonce map
	for _, account := range accounts {
		nonceMap[account.GetAccountNumber()] = 0
	}

	var (
		transactions = uint64(100)
		msg          = vm.MsgAddPackage{}

		getMsgFn = func(_ std.Account, _ int) std.Msg {
			return msg
		}
	)

	txs, err := constructTransactions(
		accountKeys,
		accounts,
		transactions,
		"dummy",
		getMsgFn,
		func(_ *std.Tx) (int64, error) {
			return 1_000_000, nil
		},
	)
	require.NoError(t, err)

	assert.Len(t, txs, int(transactions))

	// Make sure the constructed transactions are valid
	for _, tx := range txs {
		// Make sure the fee is valid
		assert.Equal(
			t,
			common.CalculateFeeInRatio(1_000_000+gasBuffer, common.DefaultGasPrice),
			tx.Fee,
		)

		// Make sure the message is valid
		if len(tx.Msgs) != 1 {
			t.Fatalf("invalid number of transaction messages, %d", len(tx.Msgs))
		}

		assert.Equal(t, msg, tx.Msgs[0])
		assert.NotEmpty(t, tx.Msgs[0].GetSigners())
	}
}
