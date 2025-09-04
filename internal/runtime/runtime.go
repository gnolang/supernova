package runtime

import (
	"context"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

const (
	realmPathPrefix   = "gno.land/r"
	packagePathPrefix = "gno.land/p"
)

// EstimateGasFn is the gas estimation callback
type EstimateGasFn func(ctx context.Context, tx *std.Tx) (int64, error)

// SignFn is the tx signing callback
type SignFn func(tx *std.Tx) error

// Runtime is the base interface for all runtime
// implementations.
//
// The runtime's job is to prepare the transactions for the stress test (generate + sign),
// and to predeploy (initialize) any infrastructure (package)
type Runtime interface {
	// Initialize prepares any infrastructure transactions that are required
	// to be executed before the stress test runs, if any
	Initialize(
		account std.Account,
		signFn SignFn,
		estimateFn EstimateGasFn,
		currentMaxGas int64,
		gasPrice std.GasPrice,
	) ([]*std.Tx, error)
	// CalculateRuntimeCosts calculates the amount of funds
	// each account needs to have in order to participate in the
	// stress test run
	CalculateRuntimeCosts(
		account std.Account,
		estimateFn EstimateGasFn,
		signFn SignFn,
		currentMaxGas int64,
		gasPrice std.GasPrice,
		transactions uint64,
	) (std.Coin, error)

	// ConstructTransactions generates and signs the required transactions
	// that will be used in the stress test
	ConstructTransactions(
		keys []crypto.PrivKey,
		accounts []std.Account,
		transactions uint64,
		maxGas int64,
		gasPrice std.GasPrice,
		chainID string,
		estimateFn EstimateGasFn,
	) ([]*std.Tx, error)
}

// GetRuntime fetches the specified runtime, if any
func GetRuntime(ctx context.Context, runtimeType Type) Runtime {
	switch runtimeType {
	case RealmCall:
		return newRealmCall(ctx)
	case RealmDeployment:
		return newRealmDeployment(ctx)
	case PackageDeployment:
		return newPackageDeployment(ctx)
	default:
		return nil
	}
}
