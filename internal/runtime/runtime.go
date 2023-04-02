package runtime

import (
	"github.com/gnolang/gno/gnoland"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gnolang/supernova/internal/common"
)

const (
	realmLocation   = "./scripts/r"
	packageLocation = "./scripts/p"

	realmPathPrefix   = "gno.land/r/demo"
	packagePathPrefix = "gno.land/p/demo"
)

var (
	defaultDeployTxFee = std.NewFee(165000, common.DefaultGasFee)
)

// Runtime is the base interface for all runtime
// implementations.
//
// The runtime's job is to prepare the transactions for the stress test (generate + sign),
// and to predeploy (initialize) any infrastructure (package)
type Runtime interface {
	// Initialize prepares any infrastructure transactions that are required
	// to be executed before the stress test runs, if any
	Initialize(*gnoland.GnoAccount) ([]*std.Tx, error)

	// ConstructTransactions generates and signs the required transactions
	// that will be used in the stress test
	ConstructTransactions(accounts []*gnoland.GnoAccount, transactions uint64) ([]*std.Tx, error)
}

type Signer interface {
	SignTx(tx *std.Tx, account *gnoland.GnoAccount, nonce uint64, passphrase string) error
}

// GetRuntime fetches the specified runtime, if any
func GetRuntime(runtimeType Type, signer Signer) Runtime {
	switch runtimeType {
	case RealmCall:
		return newRealmCall(signer)
	case RealmDeployment:
		return newCommonDeployment(signer, realmLocation, realmPathPrefix)
	case PackageDeployment:
		return newCommonDeployment(signer, packageLocation, packagePathPrefix)
	default:
		return nil
	}
}
