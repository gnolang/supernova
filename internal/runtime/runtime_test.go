package runtime

import (
	"context"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/supernova/internal/common"
	testutils "github.com/gnolang/supernova/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			r, err := GetRuntime(context.Background(), testCase.mode)
			if err != nil {
				t.Fatalf("unable to get runtime: %v", err)
			}

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
	r, err := GetRuntime(context.Background(), RealmCall)
	if err != nil {
		t.Fatalf("unable to get runtime: %v", err)
	}

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

func TestRuntime_Mixed_InitializeWithRealmCall(t *testing.T) {
	t.Parallel()

	config := &MixConfig{
		Ratios: []mixRatio{
			{RealmCall, 70},
			{RealmDeployment, 30},
		},
	}

	ctx := WithMixConfig(context.Background(), config)
	r, err := GetRuntime(ctx, Mixed)
	require.NoError(t, err)

	accounts := generateAccounts(1)

	initialTxs, err := r.Initialize(
		accounts[0],
		func(_ *std.Tx) error { return nil },
		func(_ context.Context, _ *std.Tx) (int64, error) { return 1_000_000, nil },
		1_000_000,
		common.DefaultGasPrice,
	)
	require.NoError(t, err)
	require.Len(t, initialTxs, 1)

	// Verify it's a realm deployment
	msg, ok := initialTxs[0].Msgs[0].(vm.MsgAddPackage)
	require.True(t, ok)
	assert.Contains(t, msg.Package.Path, realmPathPrefix)
	assert.Len(t, msg.Package.Files, 2)
}

func TestRuntime_Mixed_InitializeWithoutRealmCall(t *testing.T) {
	t.Parallel()

	config := &MixConfig{
		Ratios: []mixRatio{
			{RealmDeployment, 60},
			{PackageDeployment, 40},
		},
	}

	ctx := WithMixConfig(context.Background(), config)
	r, err := GetRuntime(ctx, Mixed)
	require.NoError(t, err)

	accounts := generateAccounts(1)

	initialTxs, err := r.Initialize(
		accounts[0],
		func(_ *std.Tx) error { return nil },
		func(_ context.Context, _ *std.Tx) (int64, error) { return 1_000_000, nil },
		1_000_000,
		common.DefaultGasPrice,
	)
	require.NoError(t, err)
	assert.Nil(t, initialTxs)
}

func TestRuntime_Mixed_ConstructTransactions_TypeDistribution(t *testing.T) {
	t.Parallel()

	config := &MixConfig{
		Ratios: []mixRatio{
			{RealmCall, 70},
			{RealmDeployment, 20},
			{PackageDeployment, 10},
		},
	}

	ctx := WithMixConfig(context.Background(), config)
	r, err := GetRuntime(ctx, Mixed)
	require.NoError(t, err)

	var (
		transactions = uint64(100)
		accounts     = generateAccounts(10)
		accountKeys  = testutils.GenerateAccounts(t, 10)
	)

	// Initialize first to set up realmPath for REALM_CALL
	_, err = r.Initialize(
		accounts[0],
		func(_ *std.Tx) error { return nil },
		func(_ context.Context, _ *std.Tx) (int64, error) { return 1_000_000, nil },
		1_000_000,
		common.DefaultGasPrice,
	)
	require.NoError(t, err)

	txs, err := r.ConstructTransactions(
		accountKeys,
		accounts,
		transactions,
		1_000_000,
		common.DefaultGasPrice,
		"dummy",
		func(_ context.Context, _ *std.Tx) (int64, error) { return 1_000_000, nil },
	)
	require.NoError(t, err)
	require.Len(t, txs, int(transactions))

	// Count transaction types
	var realmCalls, realmDeploys, pkgDeploys int

	for _, tx := range txs {
		require.Len(t, tx.Msgs, 1)
		switch msg := tx.Msgs[0].(type) {
		case vm.MsgCall:
			realmCalls++
		case vm.MsgAddPackage:
			if assert.NotNil(t, msg.Package) {
				if len(msg.Package.Path) > 0 && msg.Package.Path[:len(packagePathPrefix)] == packagePathPrefix {
					pkgDeploys++
				} else {
					realmDeploys++
				}
			}
		}
	}

	assert.Equal(t, 70, realmCalls)
	assert.Equal(t, 20, realmDeploys)
	assert.Equal(t, 10, pkgDeploys)
}

func TestRuntime_Mixed_ConstructTransactions_NonceManagement(t *testing.T) {
	t.Parallel()

	config := &MixConfig{
		Ratios: []mixRatio{
			{RealmCall, 50},
			{RealmDeployment, 50},
		},
	}

	ctx := WithMixConfig(context.Background(), config)
	r, err := GetRuntime(ctx, Mixed)
	require.NoError(t, err)

	var (
		transactions = uint64(20)
		accounts     = generateAccounts(2)
		accountKeys  = testutils.GenerateAccounts(t, 2)
	)

	// Initialize to set up realmPath
	_, err = r.Initialize(
		accounts[0],
		func(_ *std.Tx) error { return nil },
		func(_ context.Context, _ *std.Tx) (int64, error) { return 1_000_000, nil },
		1_000_000,
		common.DefaultGasPrice,
	)
	require.NoError(t, err)

	txs, err := r.ConstructTransactions(
		accountKeys,
		accounts,
		transactions,
		1_000_000,
		common.DefaultGasPrice,
		"dummy",
		func(_ context.Context, _ *std.Tx) (int64, error) { return 1_000_000, nil },
	)
	require.NoError(t, err)
	require.Len(t, txs, int(transactions))

	// Verify each transaction was signed
	for i, tx := range txs {
		assert.Len(t, tx.Signatures, 1, "tx %d should have exactly 1 signature", i)
	}
}

func TestRuntime_Mixed_RealmCallUsesPredeployedPath(t *testing.T) {
	t.Parallel()

	config := &MixConfig{
		Ratios: []mixRatio{
			{RealmCall, 50},
			{PackageDeployment, 50},
		},
	}

	ctx := WithMixConfig(context.Background(), config)
	r, err := GetRuntime(ctx, Mixed)
	require.NoError(t, err)

	accounts := generateAccounts(1)

	// Initialize sets the realmPath
	_, err = r.Initialize(
		accounts[0],
		func(_ *std.Tx) error { return nil },
		func(_ context.Context, _ *std.Tx) (int64, error) { return 1_000_000, nil },
		1_000_000,
		common.DefaultGasPrice,
	)
	require.NoError(t, err)

	// Access the mixed runtime to verify realmPath is set and used
	mr := r.(*mixedRuntime)
	assert.NotEmpty(t, mr.realmPath)
	assert.Contains(t, mr.realmPath, realmPathPrefix)

	// Verify getMsgForType produces a MsgCall targeting the predeployed path
	msg := mr.getMsgForType(RealmCall, accounts[0], 0)
	callMsg, ok := msg.(vm.MsgCall)
	require.True(t, ok)
	assert.Equal(t, mr.realmPath, callMsg.PkgPath)
	assert.Equal(t, methodName, callMsg.Func)
}

func TestRuntime_Mixed_DeploymentsHaveUniquePaths(t *testing.T) {
	t.Parallel()

	config := &MixConfig{
		Ratios: []mixRatio{
			{RealmDeployment, 50},
			{PackageDeployment, 50},
		},
	}

	ctx := WithMixConfig(context.Background(), config)
	r, err := GetRuntime(ctx, Mixed)
	require.NoError(t, err)

	mr := r.(*mixedRuntime)
	accounts := generateAccounts(1)

	paths := make(map[string]bool)

	for i := range 5 {
		msg := mr.getMsgForType(RealmDeployment, accounts[0], i)
		deployMsg := msg.(vm.MsgAddPackage)
		assert.False(t, paths[deployMsg.Package.Path], "duplicate realm path at index %d", i)
		paths[deployMsg.Package.Path] = true
	}

	for i := range 5 {
		msg := mr.getMsgForType(PackageDeployment, accounts[0], i)
		deployMsg := msg.(vm.MsgAddPackage)
		assert.False(t, paths[deployMsg.Package.Path], "duplicate package path at index %d", i)
		paths[deployMsg.Package.Path] = true
	}
}
