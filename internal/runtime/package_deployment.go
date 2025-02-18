package runtime

import (
	"fmt"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type packageDeployment struct{}

func newPackageDeployment() *packageDeployment {
	return &packageDeployment{}
}

func (c *packageDeployment) Initialize(
	_ std.Account,
	_ crypto.PrivKey,
	_ string,
	_ EstimateGasFn,
) ([]*std.Tx, error) {
	// No extra setup needed for this runtime type
	return nil, nil
}

func (c *packageDeployment) ConstructTransactions(
	keys []crypto.PrivKey,
	accounts []std.Account,
	transactions uint64,
	chainID string,
	estimateFn EstimateGasFn,
) ([]*std.Tx, error) {
	var (
		timestamp = time.Now().Unix()

		getMsgFn = func(creator std.Account, index int) std.Msg {
			memPkg := &gnovm.MemPackage{
				Name: packageName,
				Path: fmt.Sprintf(
					"%s/%s/stress_%d_%d",
					packagePathPrefix,
					creator.GetAddress().String(),
					timestamp,
					index,
				),
				Files: []*gnovm.MemFile{
					{
						Name: packageFileName,
						Body: packageBody,
					},
				},
			}

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
		estimateFn,
	)
}
