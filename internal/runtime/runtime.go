package runtime

import (
	"context"
	"errors"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var (
	errMissingMixConfig = errors.New("mix config is required for MIXED mode")
	errUnknownRuntime   = errors.New("unknown runtime type")
)

const (
	realmPathPrefix   = "gno.land/r"
	packagePathPrefix = "gno.land/p"
)

type mixConfigKey struct{}

// WithMixConfig attaches the mix config to the context
func WithMixConfig(ctx context.Context, config *MixConfig) context.Context {
	return context.WithValue(ctx, mixConfigKey{}, config)
}

// GetMixConfig retrieves the mix config from the context, if any
func GetMixConfig(ctx context.Context) *MixConfig {
	config, _ := ctx.Value(mixConfigKey{}).(*MixConfig)
	return config
}

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
// Returns an error if the runtime type is unknown or if the mix config is missing for mixed runtimes
func GetRuntime(ctx context.Context, runtimeType Type) (Runtime, error) {
	switch runtimeType {
	case RealmCall:
		return newRealmCall(ctx), nil
	case RealmDeployment:
		return newRealmDeployment(ctx), nil
	case PackageDeployment:
		return newPackageDeployment(ctx), nil
	case Mixed:
		config := GetMixConfig(ctx)
		if config == nil {
			return nil, errMissingMixConfig
		}
		return newMixedRuntime(ctx, config), nil
	default:
		return nil, errUnknownRuntime
	}
}
