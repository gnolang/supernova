package runtime

import (
	"fmt"
	"time"

	"github.com/gnolang/gno/gnoland"
	"github.com/gnolang/gno/pkgs/gnolang"
	"github.com/gnolang/gno/pkgs/sdk/vm"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gnolang/supernova/internal/common"
	"github.com/gnolang/supernova/internal/signer"
)

const (
	methodName = "SayHello"
)

type realmCall struct {
	signer signer.Signer

	realmPath string
}

func newRealmCall(signer signer.Signer) *realmCall {
	return &realmCall{
		signer: signer,
	}
}

func (r *realmCall) Initialize(account *gnoland.GnoAccount) error {
	// The Realm needs to be deployed before
	// it can be interacted with
	realmPath := fmt.Sprintf("gno.land/r/demo/stress-%d", time.Now().Unix())

	memPkg := gnolang.ReadMemPackage(
		"./scripts/r",
		realmPath,
	)

	msg := vm.MsgAddPackage{
		Creator: account.GetAddress(),
		Package: memPkg,
	}

	tx := &std.Tx{
		Msgs: []std.Msg{msg},
		Fee:  std.NewFee(600000, common.DefaultGasFee),
	}

	if err := r.signer.SignTx(tx, account, account.Sequence, common.EncryptPassword); err != nil {
		return fmt.Errorf("unable to sign initialize transaction, %w", err)
	}

	// TODO Broadcast

	return nil
}

func (r *realmCall) ConstructTransactions(
	accounts []*gnoland.GnoAccount,
	transactions uint64,
) ([]*std.Tx, error) {
	getMsgFn := func(creator *gnoland.GnoAccount, index int) std.Msg {
		return vm.MsgCall{
			Caller:  creator.Address,
			PkgPath: r.realmPath,
			Func:    methodName,
			Args:    []string{fmt.Sprintf("Account-%d", index)},
		}
	}

	return constructTransactions(
		r.signer,
		accounts,
		transactions,
		getMsgFn,
	)
}
