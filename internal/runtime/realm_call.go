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
	var (
		txs = make([]*std.Tx, transactions)

		// A local nonce map is updated to avoid unnecessary calls
		// for fetching the fresh info from the chain every time
		// an account is used
		nonceMap = make(map[uint64]uint64) // accountNumber -> nonce
	)

	for i := 0; i < int(transactions); i++ {
		// Generate the transaction
		creator := accounts[i%len(accounts)]

		msg := vm.MsgCall{
			Caller:  creator.Address,
			PkgPath: r.realmPath,
			Func:    methodName,
			Args:    []string{fmt.Sprintf("Account-%d", i)},
		}

		tx := &std.Tx{
			Msgs: []std.Msg{msg},
			Fee:  std.NewFee(600000, common.DefaultGasFee),
		}

		// Fetch the next account nonce
		nonce, found := nonceMap[creator.AccountNumber]
		if !found {
			nonce = creator.Sequence
			nonceMap[creator.AccountNumber] = nonce
		}

		// Sign the transaction
		if err := r.signer.SignTx(tx, creator, nonce, common.EncryptPassword); err != nil {
			return nil, fmt.Errorf("unable to sign call transaction, %w", err)
		}

		// Increase the creator nonce locally
		nonceMap[creator.AccountNumber] = nonce + 1

		// Mark the transaction as ready
		txs[i] = tx
	}

	return txs, nil
}
