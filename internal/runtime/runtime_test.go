package runtime

import (
	"context"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/supernova/internal/common"
	testutils "github.com/gnolang/supernova/internal/testing"
	"github.com/stretchr/testify/assert"
)

// verifyDeployTxCommon does common transaction verification
func verifyDeployTxCommon(t *testing.T, tx *std.Tx, expectedPrefix string) {
	t.Helper()

	if len(tx.Msgs) != 1 {
		t.Fatalf("invalid number of tx messages, %d", len(tx.Msgs))
	}

	msg := tx.Msgs[0]

	vmMsg, ok := msg.(vm.MsgAddPackage)
	if !ok {
		t.Fatal("invalid tx message type")
	}

	// Make sure the deploy params are valid
	assert.Contains(t, vmMsg.Package.Path, expectedPrefix)
	assert.Len(t, vmMsg.Package.Files, 2)
	assert.NotNil(t, vmMsg.Creator)
	assert.Nil(t, vmMsg.Send)

	// Make sure the fee is valid
	assert.Equal(
		t,
		common.CalculateFeeInRatio(1_000_000+gasBuffer, common.DefaultGasPrice),
		tx.Fee,
	)
}

func TestRuntime_CommonDeployment(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name           string
		mode           Type
		expectedPrefix string
	}{
		{
			"Realm Deployment",
			RealmDeployment,
			realmPathPrefix,
		},
		{
			"Package Deployment",
			PackageDeployment,
			packagePathPrefix,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				transactions = uint64(100)
				accounts     = generateAccounts(10)
				accountKeys  = testutils.GenerateAccounts(t, 10)
			)

			// Get the runtime
			r := GetRuntime(context.Background(), testCase.mode)

			// Make sure there is no initialization logic
			initialTxs, err := r.Initialize(
				accounts[0],
				func(_ *std.Tx) error {
					return nil
				},
				func(_ context.Context, _ *std.Tx) (int64, error) {
					return 1_000_000, nil
				},
				1_000_000,
				common.DefaultGasPrice,
			)

			assert.Nil(t, initialTxs)
			assert.Nil(t, err)

			// Construct the transactions
			txs, err := r.ConstructTransactions(
				accountKeys,
				accounts,
				transactions,
				1_000_000,
				common.DefaultGasPrice,
				"dummy",
				func(_ context.Context, _ *std.Tx) (int64, error) {
					return 1_000_000, nil
				},
			)
			if err != nil {
				t.Fatalf("unable to construct transactions, %v", err)
			}

			// Make sure they were constructed properly
			if len(txs) != int(transactions) {
				t.Fatalf("invalid number of transactions constructed, %d", len(txs))
			}

			for _, tx := range txs {
				verifyDeployTxCommon(t, tx, testCase.expectedPrefix)
			}
		})
	}
}

func TestRuntime_RealmCall(t *testing.T) {
	t.Parallel()

	var (
		transactions = uint64(100)
		accounts     = generateAccounts(11)
		accountKeys  = testutils.GenerateAccounts(t, 11)
	)

	// Get the runtime
	r := GetRuntime(context.Background(), RealmCall)

	// Make sure the initialization logic is present
	initialTxs, err := r.Initialize(
		accounts[0],
		func(_ *std.Tx) error {
			return nil
		},
		func(_ context.Context, _ *std.Tx) (int64, error) {
			return 1_000_000, nil
		},
		1_000_000,
		common.DefaultGasPrice,
	)
	if err != nil {
		t.Fatalf("unable to generate init transactions, %v", err)
	}

	if len(initialTxs) != 1 {
		t.Fatalf("invalid number of initial transactions, %d", len(initialTxs))
	}

	for _, tx := range initialTxs {
		verifyDeployTxCommon(t, tx, realmPathPrefix)
	}

	// Construct the transactions
	txs, err := r.ConstructTransactions(
		accountKeys[1:],
		accounts[1:],
		transactions,
		1_000_000,
		common.DefaultGasPrice,
		"dummy",
		func(_ context.Context, _ *std.Tx) (int64, error) {
			return 1_000_000, nil
		},
	)
	if err != nil {
		t.Fatalf("unable to construct transactions, %v", err)
	}

	// Make sure they were constructed properly
	if len(txs) != int(transactions) {
		t.Fatalf("invalid number of transactions constructed, %d", len(txs))
	}

	for _, tx := range txs {
		if len(tx.Msgs) != 1 {
			t.Fatalf("invalid number of tx messages, %d", len(tx.Msgs))
		}

		msg := tx.Msgs[0]

		vmMsg, ok := msg.(vm.MsgCall)
		if !ok {
			t.Fatal("invalid tx message type")
		}

		// Make sure the call params are valid
		assert.Equal(t, vmMsg.Func, methodName)
		assert.NotNil(t, vmMsg.Caller)
		assert.Nil(t, vmMsg.Send)

		if len(vmMsg.Args) != 1 {
			t.Fatalf("invalid number of arguments provided for call")
		}

		assert.Contains(t, vmMsg.Args[0], "Account")

		// Make sure the fee is valid
		assert.Equal(
			t,
			common.CalculateFeeInRatio(1_000_000+gasBuffer, common.DefaultGasPrice),
			tx.Fee,
		)
	}
}
