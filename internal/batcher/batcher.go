package batcher

import (
	"errors"
	"fmt"

	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/bft/rpc/client"
	core_types "github.com/gnolang/gno/pkgs/bft/rpc/core/types"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/schollz/progressbar/v3"
)

type Batcher struct {
	cli *client.HTTP
}

// NewBatcher creates a new Batcher instance
func NewBatcher(cli *client.HTTP) *Batcher {
	return &Batcher{
		cli: cli,
	}
}

type TxBatchResults struct {
	TxHashes   [][]byte
	StartBlock int64
}

func (b *Batcher) BatchTransactions(txs []*std.Tx, batchSize int) (*TxBatchResults, error) {
	// Note the current latest block
	latest, err := b.getLatestBlock()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch latest block %w", err)
	}

	fmt.Printf("Latest block number: %d\n", latest)

	// Marshal the transactions
	fmt.Printf("\nPreparing transactions...\n")

	preparedTxs, err := prepareTransactions(txs)
	if err != nil {
		return nil, fmt.Errorf("unable to batch transactions, %w", err)
	}

	// Generate the batches
	batches := generateBatches[[]byte](preparedTxs, batchSize)
	numBatches := len(batches)

	fmt.Printf("\nGenerating batches...\n")

	bar := progressbar.Default(int64(numBatches), "batches generated")

	readyBatches := make([]*client.BatchHTTP, numBatches)

	for index, batch := range batches {
		cliBatch := b.cli.NewBatch()

		for _, tx := range batch {
			// Append the transaction
			if _, err = cliBatch.BroadcastTxSync(tx); err != nil {
				return nil, fmt.Errorf("unable to prepare transaction, %w", err)
			}
		}

		readyBatches[index] = cliBatch

		_ = bar.Add(1)
	}

	// Execute the batch requests.
	// Batch requests need to be sent out sequentially
	// to preserve account sequence order
	fmt.Printf("\nSending batches...\n")

	bar = progressbar.Default(int64(numBatches), "batches sent")

	batchResults := make([][]any, numBatches)

	for index, readyBatch := range readyBatches {
		batchResult, err := readyBatch.Send()
		if err != nil {
			return nil, fmt.Errorf("unable to batch request, %w", err)
		}

		batchResults[index] = batchResult

		_ = bar.Add(1)
	}

	// Parse the results.
	// Parsing is done in a separate loop to not hinder
	// the batch send speed (as txs need to be parsed sequentially)
	fmt.Printf("\nParsing batch results...\n")

	bar = progressbar.Default(int64(len(txs)), "results parsed")

	txHashes := make([][]byte, len(txs))
	index := 0

	for _, batchResult := range batchResults {
		// For each batch, extract the transaction hashes
		for _, txResultRaw := range batchResult {
			txResult, ok := txResultRaw.(*core_types.ResultBroadcastTx)
			if !ok {
				return nil, errors.New("invalid result type returned")
			}

			// Check the errors
			if txResult.Error != nil {
				return nil, fmt.Errorf(
					"error when parsing transaction %s, %w",
					txResult.Hash,
					txResult.Error,
				)
			}

			txHashes[index] = txResult.Hash
			index++

			_ = bar.Add(1)
		}
	}

	fmt.Printf("✅ Successfully sent %d txs in %d batches\n", len(txs), numBatches)

	return &TxBatchResults{
		TxHashes:   txHashes,
		StartBlock: latest,
	}, nil
}

// prepareTransactions marshals the transactions into amino binary
func prepareTransactions(txs []*std.Tx) ([][]byte, error) {
	marshalledTxs := make([][]byte, len(txs))
	bar := progressbar.Default(int64(len(txs)), "txs prepared")

	for index, tx := range txs {
		txBin, err := amino.Marshal(tx)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal tx, %w", err)
		}

		marshalledTxs[index] = txBin

		_ = bar.Add(1)
	}

	return marshalledTxs, nil
}

// getLatestBlock fetches the latest block height from the chain
func (b *Batcher) getLatestBlock() (int64, error) {
	status, err := b.cli.Status()
	if err != nil {
		return 0, fmt.Errorf("unable to fetch status, %w", err)
	}

	return status.SyncInfo.LatestBlockHeight, nil
}
