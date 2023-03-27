package runtime

import (
	"github.com/gnolang/gno/gnoland"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gnolang/supernova/internal/signer"
)

// Runtime is the base interface for all runtime
// implementations.
//
// The runtime's job is to prepare the transactions for the stress test (generate + sign),
// and to predeploy (initialize) any infrastructure (package)
type Runtime interface {
	// Initialize prepares any infrastructure (predeploys packages), if any
	Initialize(*gnoland.GnoAccount) error

	// ConstructTransactions generates and signs the required transactions
	// that will be used in the stress test
	ConstructTransactions(accounts []*gnoland.GnoAccount, transactions uint64) ([]*std.Tx, error)
}

// GetRuntime fetches the specified runtime, if any
func GetRuntime(runtimeType Type, signer signer.Signer) Runtime {
	switch runtimeType {
	case RealmCall:
		return newRealmCall(signer)
	case RealmDeployment:
		return newCommonDeployment(signer, "r")
	case PackageDeployment:
		return newCommonDeployment(signer, "p")
	default:
		return nil
	}
}
