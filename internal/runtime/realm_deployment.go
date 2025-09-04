package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type realmDeployment struct {
	ctx context.Context
}

func newRealmDeployment(ctx context.Context) *realmDeployment {
	return &realmDeployment{
		ctx: ctx,
	}
}

func (c *realmDeployment) Initialize(
	_ std.Account,
	_ SignFn,
	_ EstimateGasFn,
	_ int64,
	_ std.GasPrice,
) ([]*std.Tx, error) {
	// No extra setup needed for this runtime type
	return nil, nil
}

func (c *realmDeployment) CalculateRuntimeCosts(
	account std.Account,
	estimateFn EstimateGasFn,
	signFn SignFn,
	currentMaxGas int64,
	gasPrice std.GasPrice,
	transactions uint64,
) (std.Coin, error) {
	return calculateRuntimeCosts(
		c.ctx,
		account,
		transactions,
		currentMaxGas,
		gasPrice,
		c.getMsgFn,
		signFn,
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
	maxGas int64,
	gasPrice std.GasPrice,
	chainID string,
	estimateFn EstimateGasFn,
) ([]*std.Tx, error) {
	return constructTransactions(
		c.ctx,
		keys,
		accounts,
		transactions,
		maxGas,
		gasPrice,
		chainID,
		c.getMsgFn,
		estimateFn,
	)
}
