package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/supernova/internal/common"
)

const methodName = "SayHello"

type realmCall struct {
	realmPath string
	ctx       context.Context
}

func newRealmCall(ctx context.Context) *realmCall {
	return &realmCall{
		ctx: ctx,
	}
}

func (r *realmCall) Initialize(
	account std.Account,
	signFn SignFn,
	estimateFn EstimateGasFn,
	currentMaxGas int64,
	gasPrice std.GasPrice,
) ([]*std.Tx, error) {
	// The Realm needs to be deployed before
	// it can be interacted with
	r.realmPath = fmt.Sprintf(
		"%s/%s/stress_%d",
		realmPathPrefix,
		account.GetAddress().String(),
		time.Now().Unix(),
	)

	// Construct the transaction
	msg := vm.MsgAddPackage{
		Creator: account.GetAddress(),
		Package: &std.MemPackage{
			Name: packageName,
			Path: r.realmPath,
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
		},
	}

	tx := &std.Tx{
		Msgs: []std.Msg{msg},
		// passing in the maximum block gas, this is just a simulation
		Fee: common.CalculateFeeInRatio(currentMaxGas, gasPrice),
	}

	err := signFn(tx)
	if err != nil {
		return nil, fmt.Errorf("unable to sign initialize transaction, %w", err)
	}

	// Estimate the gas for the initial tx
	gasWanted, err := estimateFn(r.ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("unable to estimate gas: %w", err)
	}

	// Wipe the signatures, because we will change the fee,
	// and cause the previous ones to be invalid
	tx.Signatures = make([]std.Signature, 0)
	tx.Fee = common.CalculateFeeInRatio(gasWanted+gasBuffer, gasPrice) // buffer with 10k gas

	err = signFn(tx)
	if err != nil {
		return nil, fmt.Errorf("unable to sign initialize transaction, %w", err)
	}

	return []*std.Tx{tx}, nil
}

func (r *realmCall) CalculateRuntimeCosts(
	account std.Account,
	estimateFn EstimateGasFn,
	signFn SignFn,
	currentMaxGas int64,
	gasPrice std.GasPrice,
	transactions uint64,
) (std.Coin, error) {
	return calculateRuntimeCosts(
		r.ctx,
		account,
		transactions,
		currentMaxGas,
		gasPrice,
		r.getMsgFn,
		signFn,
		estimateFn,
	)
}

func (r *realmCall) ConstructTransactions(
	keys []crypto.PrivKey,
	accounts []std.Account,
	transactions uint64,
	maxGas int64,
	gasPrice std.GasPrice,
	chainID string,
	estimateFn EstimateGasFn,
) ([]*std.Tx, error) {
	return constructTransactions(
		r.ctx,
		keys,
		accounts,
		transactions,
		maxGas,
		gasPrice,
		chainID,
		r.getMsgFn,
		estimateFn,
	)
}

func (r *realmCall) getMsgFn(creator std.Account, index int) std.Msg {
	return vm.MsgCall{
		Caller:  creator.GetAddress(),
		PkgPath: r.realmPath,
		Func:    methodName,
		Args:    []string{fmt.Sprintf("Account-%d", index)},
	}
}
