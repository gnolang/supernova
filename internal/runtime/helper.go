package runtime

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/supernova/internal/signer"
	"github.com/schollz/progressbar/v3"
)

// msgFn defines the transaction message constructor
type msgFn func(creator std.Account, index int) std.Msg

// constructTransactions constructs and signs the transactions
// using the passed in message generator and signer
func constructTransactions(
	keys []crypto.PrivKey,
	accounts []std.Account,
	transactions uint64,
	chainID string,
	getMsg msgFn,
	estimateFn EstimateGasFn,
) ([]*std.Tx, error) {
	var (
		txs = make([]*std.Tx, transactions)

		// A local nonce map is updated to avoid unnecessary calls
		// for fetching the fresh info from the chain every time
		// an account is used
		nonceMap = make(map[uint64]uint64) // accountNumber -> nonce
	)

	fmt.Printf("\n🔨 Estimating Gas 🔨\n\n")

	// Estimate the fee for the transaction batch
	txFee := defaultDeployTxFee

	// Construct the first tx
	// Generate the transaction
	var (
		creator    = accounts[0]
		creatorKey = keys[0]
	)

	tx := &std.Tx{
		Msgs: []std.Msg{getMsg(creator, 0)},
		Fee:  txFee,
	}

	// Sign the transaction
	cfg := signer.SignCfg{
		ChainID:       chainID,
		AccountNumber: creator.GetAccountNumber(),
		Sequence:      creator.GetSequence(),
	}

	if err := signer.SignTx(tx, creatorKey, cfg); err != nil {
		return nil, fmt.Errorf("unable to sign transaction, %w", err)
	}

	gasWanted, err := estimateFn(tx)
	if err != nil {
		return nil, fmt.Errorf("unable to estimate gas, %w", err)
	}

	// Clear the old signatures, because they need
	// to be regenerated
	clear(tx.Signatures)

	// Use the estimated gas limit
	txFee.GasWanted = gasWanted + 10_000 // 10k gas buffer

	if err = signer.SignTx(tx, creatorKey, cfg); err != nil {
		return nil, fmt.Errorf("unable to sign transaction, %w", err)
	}

	fmt.Printf("\n🔨 Constructing Transactions 🔨\n\n")

	bar := progressbar.Default(int64(transactions), "constructing txs")

	for i := 0; i < int(transactions); i++ {
		// Generate the transaction
		var (
			creator       = accounts[i%len(accounts)]
			creatorKey    = keys[i%len(accounts)]
			accountNumber = creator.GetAccountNumber()
		)

		tx := &std.Tx{
			Msgs: []std.Msg{getMsg(creator, i)},
			Fee:  txFee,
		}

		// Fetch the next account nonce
		nonce, found := nonceMap[creator.GetAccountNumber()]
		if !found {
			nonce = creator.GetSequence()
			nonceMap[creator.GetAccountNumber()] = nonce
		}

		// Sign the transaction
		cfg := signer.SignCfg{
			ChainID:       chainID,
			AccountNumber: accountNumber,
			Sequence:      nonce,
		}

		if err := signer.SignTx(tx, creatorKey, cfg); err != nil {
			return nil, fmt.Errorf("unable to sign transaction, %w", err)
		}

		// Increase the creator nonce locally
		nonceMap[accountNumber] = nonce + 1

		// Mark the transaction as ready
		txs[i] = tx
		_ = bar.Add(1) //nolint:errcheck // No need to check
	}

	fmt.Printf("✅ Successfully constructed %d transactions\n", transactions)

	return txs, nil
}
