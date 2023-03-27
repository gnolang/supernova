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
		txs       = make([]*std.Tx, transactions)

		// A local nonce map is updated to avoid unnecessary calls
		// for fetching the fresh info from the chain every time
		// an account is used
		nonceMap = make(map[uint64]uint64) // accountNumber -> nonce
	)

	for i := 0; i < int(transactions); i++ {
		// Generate the transaction
		creator := accounts[i%len(accounts)]

		memPkg := gnolang.ReadMemPackage(
			fmt.Sprintf("./scripts/%s", c.deployPrefix),
			fmt.Sprintf("gno.land/%s/demo/stress-%d-%d", c.deployPrefix, timestamp, i),
		)

		msg := vm.MsgAddPackage{
			Creator: creator.GetAddress(),
			Package: memPkg,
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
		if err := c.signer.SignTx(tx, creator, nonce, common.EncryptPassword); err != nil {
			return nil, fmt.Errorf("unable to sign deploy transaction, %w", err)
		}

		// Increase the creator nonce locally
		nonceMap[creator.AccountNumber] = nonce + 1

		// Mark the transaction as ready
		txs[i] = tx
	}

	return txs, nil
}
