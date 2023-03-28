package runtime

import (
	"fmt"
	"time"

	"github.com/gnolang/gno/gnoland"
	"github.com/gnolang/gno/pkgs/gnolang"
	"github.com/gnolang/gno/pkgs/sdk/vm"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gnolang/supernova/internal/signer"
)

type commonDeployment struct {
	signer signer.Signer

	deployPrefix string
}

func newCommonDeployment(signer signer.Signer, prefix string) *commonDeployment {
	return &commonDeployment{
		signer:       signer,
		deployPrefix: prefix,
	}
}

func (c *commonDeployment) Initialize(_ *gnoland.GnoAccount) error {
	// No extra setup needed for this runtime type
	return nil
}

func (c *commonDeployment) ConstructTransactions(
	accounts []*gnoland.GnoAccount,
	transactions uint64,
) ([]*std.Tx, error) {
	var (
		timestamp = time.Now().Unix()
		getMsgFn  = func(creator *gnoland.GnoAccount, index int) std.Msg {
			memPkg := gnolang.ReadMemPackage(
				fmt.Sprintf("./scripts/%s", c.deployPrefix),
				fmt.Sprintf("gno.land/%s/demo/stress-%d-%d", c.deployPrefix, timestamp, index),
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
