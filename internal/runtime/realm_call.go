package runtime

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/supernova/internal/signer"
)

const (
	methodName = "SayHello"
)

type realmCall struct {
	realmPath string
}

func newRealmCall() *realmCall {
	return &realmCall{}
}

func (r *realmCall) Initialize(account std.Account, key crypto.PrivKey, chainID string) ([]*std.Tx, error) {
	// Get absolute path to folder
	deployPathAbs, err := filepath.Abs(realmLocation)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve absolute path, %w", err)
	}

	// The Realm needs to be deployed before
	// it can be interacted with
	r.realmPath = fmt.Sprintf("%s/stress_%d", realmPathPrefix, time.Now().Unix())

	// Construct the transaction
	msg := vm.MsgAddPackage{
		Creator: account.GetAddress(),
		Package: gnolang.ReadMemPackage(
			deployPathAbs,
			r.realmPath,
		),
	}

	tx := &std.Tx{
		Msgs: []std.Msg{msg},
		Fee:  defaultDeployTxFee,
	}

	// Sign it
	cfg := signer.SignCfg{
		ChainID:       chainID,
		AccountNumber: account.GetAccountNumber(),
		Sequence:      account.GetSequence(),
	}

	if err := signer.SignTx(tx, key, cfg); err != nil {
		return nil, fmt.Errorf("unable to sign initialize transaction, %w", err)
	}

	return []*std.Tx{tx}, nil
}

func (r *realmCall) ConstructTransactions(
	keys []crypto.PrivKey,
	accounts []std.Account,
	transactions uint64,
	chainID string,
) ([]*std.Tx, error) {
	getMsgFn := func(creator std.Account, index int) std.Msg {
		return vm.MsgCall{
			Caller:  creator.GetAddress(),
			PkgPath: r.realmPath,
			Func:    methodName,
			Args:    []string{fmt.Sprintf("Account-%d", index)},
		}
	}

	return constructTransactions(
		keys,
		accounts,
		transactions,
		chainID,
		getMsgFn,
	)
}
