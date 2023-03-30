package runtime

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/gnolang/gno/gnoland"
	"github.com/gnolang/gno/pkgs/gnolang"
	"github.com/gnolang/gno/pkgs/sdk/vm"
	"github.com/gnolang/gno/pkgs/std"
)

type commonDeployment struct {
	signer Signer

	deployDir        string
	deployPathPrefix string
}

func newCommonDeployment(signer Signer, deployDir, deployPrefix string) *commonDeployment {
	return &commonDeployment{
		signer:           signer,
		deployDir:        deployDir,
		deployPathPrefix: deployPrefix,
	}
}

func (c *commonDeployment) Initialize(_ *gnoland.GnoAccount) ([]*std.Tx, error) {
	// No extra setup needed for this runtime type
	return nil, nil
}

func (c *commonDeployment) ConstructTransactions(
	accounts []*gnoland.GnoAccount,
	transactions uint64,
) ([]*std.Tx, error) {
	// Get absolute path to folder
	deployPathAbs, err := filepath.Abs(c.deployDir)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve absolute path, %w", err)
	}

	var (
		timestamp = time.Now().Unix()
		getMsgFn  = func(creator *gnoland.GnoAccount, index int) std.Msg {
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
		c.signer,
		accounts,
		transactions,
		getMsgFn,
	)
}
