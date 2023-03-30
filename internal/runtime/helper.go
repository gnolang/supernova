package runtime

import (
	"fmt"

	"github.com/gnolang/gno/gnoland"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gnolang/supernova/internal/common"
	"github.com/schollz/progressbar/v3"
)

// msgFn defines the transaction message constructor
type msgFn func(creator *gnoland.GnoAccount, index int) std.Msg

// constructTransactions constructs and signs the transactions
// using the passed in message generator and signer
func constructTransactions(
	signer Signer,
	accounts []*gnoland.GnoAccount,
	transactions uint64,
	getMsg msgFn,
) ([]*std.Tx, error) {
	var (
		txs = make([]*std.Tx, transactions)

		// A local nonce map is updated to avoid unnecessary calls
		// for fetching the fresh info from the chain every time
		// an account is used
		nonceMap = make(map[uint64]uint64) // accountNumber -> nonce
	)

	fmt.Printf("\n🔨 Constructing Transactions 🔨\n\n")

	bar := progressbar.Default(int64(transactions), "constructing txs")

	for i := 0; i < int(transactions); i++ {
		// Generate the transaction
		creator := accounts[i%len(accounts)]

		tx := &std.Tx{
			Msgs: []std.Msg{getMsg(creator, i)},
			Fee:  defaultDeployTxFee,
		}

		// Fetch the next account nonce
		nonce, found := nonceMap[creator.AccountNumber]
		if !found {
			nonce = creator.Sequence
			nonceMap[creator.AccountNumber] = nonce
		}

		// Sign the transaction
		if err := signer.SignTx(tx, creator, nonce, common.EncryptPassword); err != nil {
			return nil, fmt.Errorf("unable to sign transaction, %w", err)
		}

		// Increase the creator nonce locally
		nonceMap[creator.AccountNumber] = nonce + 1

		// Mark the transaction as ready
		txs[i] = tx
		_ = bar.Add(1)
	}

	fmt.Printf("✅ Successfully constructed %d transactions\n", transactions)

	return txs, nil
}
