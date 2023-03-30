package runtime

import (
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/gnolang/gno/pkgs/sdk/vm"
	"github.com/gnolang/gno/pkgs/std"
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
	assert.Len(t, vmMsg.Package.Files, 1)
	assert.NotNil(t, vmMsg.Creator)
	assert.Nil(t, vmMsg.Deposit)

	// Make sure the fee is valid
	assert.Equal(t, tx.Fee, defaultDeployTxFee)
}

// moveToRoot sets the current working
// test directory to the project root.
// This is used because of fixed .gno files in
// ./scripts
func moveToRoot(t *testing.T) {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to get caller information")
	}

	dir := path.Join(path.Dir(filename), "../..")
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("unable to change to root dir, %v", err)
	}
}

func TestRuntime_CommonDeployment(t *testing.T) {
	t.Parallel()

	// Change the working directory to root
	moveToRoot(t)

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
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				transactions = uint64(100)
				accounts     = generateAccounts(10)
			)

			// Get the runtime
			r := GetRuntime(testCase.mode, &mockSigner{})

			// Make sure there is no initialization logic
			initialTxs, err := r.Initialize(nil)

			assert.Nil(t, initialTxs)
			assert.Nil(t, err)

			// Construct the transactions
			txs, err := r.ConstructTransactions(accounts, transactions)
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

	// Change the working directory to root
	moveToRoot(t)

	var (
		transactions = uint64(100)
		accounts     = generateAccounts(11)
	)

	// Get the runtime
	r := GetRuntime(RealmCall, &mockSigner{})

	// Make sure the initialization logic is present
	initialTxs, err := r.Initialize(accounts[0])
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
	txs, err := r.ConstructTransactions(accounts, transactions)
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
		assert.Equal(t, tx.Fee, defaultDeployTxFee)
	}
}
