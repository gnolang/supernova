package runtime

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type commonDeployment struct {
	deployDir        string
	deployPathPrefix string
}

func newCommonDeployment(deployDir, deployPrefix string) *commonDeployment {
	return &commonDeployment{
		deployDir:        deployDir,
		deployPathPrefix: deployPrefix,
	}
}

func (c *commonDeployment) Initialize(_ std.Account, _ crypto.PrivKey, _ string) ([]*std.Tx, error) {
	// No extra setup needed for this runtime type
	return nil, nil
}

func (c *commonDeployment) ConstructTransactions(
	keys []crypto.PrivKey,
	accounts []std.Account,
	transactions uint64,
	chainID string,
) ([]*std.Tx, error) {
	// Get absolute path to folder
	deployPathAbs, err := filepath.Abs(c.deployDir)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve absolute path, %w", err)
	}

	var (
		timestamp = time.Now().Unix()

		getMsgFn = func(creator std.Account, index int) std.Msg {
			memPkg := gnolang.ReadMemPackage(
				deployPathAbs,
				fmt.Sprintf("%s/stress_%d_%d", c.deployPathPrefix, timestamp, index),
			)

			return vm.MsgAddPackage{
				Creator: creator.GetAddress(),
				Package: memPkg,
			}
		}
	)

	return constructTransactions(
		keys,
		accounts,
		transactions,
		chainID,
		getMsgFn,
	)
}
