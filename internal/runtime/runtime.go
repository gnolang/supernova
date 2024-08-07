package runtime

import (
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/supernova/internal/common"
)

const (
	realmPathPrefix   = "gno.land/r"
	packagePathPrefix = "gno.land/p"
)

var defaultDeployTxFee = std.NewFee(500000, common.DefaultGasFee)

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
		key crypto.PrivKey,
		chainID string,
	) ([]*std.Tx, error)

	// ConstructTransactions generates and signs the required transactions
	// that will be used in the stress test
	ConstructTransactions(
		keys []crypto.PrivKey,
		accounts []std.Account,
		transactions uint64,
		chainID string,
	) ([]*std.Tx, error)
}

// GetRuntime fetches the specified runtime, if any
func GetRuntime(runtimeType Type) Runtime {
	switch runtimeType {
	case RealmCall:
		return newRealmCall()
	case RealmDeployment:
		return newRealmDeployment()
	case PackageDeployment:
		return newPackageDeployment()
	default:
		return nil
	}
}
