package runtime

import (
	"fmt"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type realmDeployment struct{}

func newRealmDeployment() *realmDeployment {
	return &realmDeployment{}
}

func (c *realmDeployment) Initialize(
	_ std.Account,
	_ crypto.PrivKey,
	_ string,
	_ EstimateGasFn,
) ([]*std.Tx, error) {
	// No extra setup needed for this runtime type
	return nil, nil
}

func (c *realmDeployment) CalculateRuntimeCosts(
	account std.Account,
	key crypto.PrivKey,
	chainID string,
	estimateFn EstimateGasFn,
	transactions uint64,
) (std.Coin, error) {
	return calculateRuntimeCosts(
		key,
		account,
		transactions,
		chainID,
		c.getMsgFn,
		estimateFn,
	)
}

func (c *realmDeployment) getMsgFn(creator std.Account, index int) std.Msg {
	timestamp := time.Now().Unix()
	memPkg := &std.MemPackage{
		Name: packageName,
		Path: fmt.Sprintf(
			"%s/%s/stress_%d_%d",
			realmPathPrefix,
			creator.GetAddress().String(),
			timestamp,
			index,
		),
		Files: []*std.MemFile{
			{
				Name: gnomodFileName,
				Body: gnomodBody,
			},
			{
				Name: realmFileName,
				Body: realmBody,
			},
		},
	}

	return vm.MsgAddPackage{
		Creator: creator.GetAddress(),
		Package: memPkg,
	}
}

func (c *realmDeployment) ConstructTransactions(
	keys []crypto.PrivKey,
	accounts []std.Account,
	transactions uint64,
	chainID string,
	estimateFn EstimateGasFn,
) ([]*std.Tx, error) {
	return constructTransactions(
		keys,
		accounts,
		transactions,
		chainID,
		c.getMsgFn,
		estimateFn,
	)
}
