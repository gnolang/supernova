package batcher

import (
	"errors"
	"fmt"
	"math"

	"github.com/gnolang/gno/pkgs/amino"
	core_types "github.com/gnolang/gno/pkgs/bft/rpc/core/types"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gnolang/supernova/internal/common"
	"github.com/schollz/progressbar/v3"
)

// Batcher batches signed transactions
// to the Gno Tendermint node
type Batcher struct {
	cli Client
}

// NewBatcher creates a new Batcher instance
func NewBatcher(cli Client) *Batcher {
	return &Batcher{
		cli: cli,
	}
}

// BatchTransactions batches provided transactions using the
// specified batch size
func (b *Batcher) BatchTransactions(txs []*std.Tx, batchSize int) (*TxBatchResult, error) {
	fmt.Printf("\n📦 Batching Transactions 📦\n\n")

	// Note the current latest block
	latest, err := b.cli.GetLatestBlockHeight()
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
	readyBatches, err := b.generateBatches(preparedTxs, batchSize)
	if err != nil {
		return nil, fmt.Errorf("unable to generate batches, %w", err)
	}

	// Execute the batch requests.
	// Batch requests need to be sent out sequentially
	// to preserve account sequence order
	batchResults, err := sendBatches(readyBatches)
	if err != nil {
		return nil, fmt.Errorf("unable to send batches, %w", err)
	}

	// Parse the results
	txHashes, err := parseBatchResults(batchResults, len(txs))
	if err != nil {
		return nil, fmt.Errorf("unable to parse batch results, %w", err)
	}

	fmt.Printf("✅ Successfully sent %d txs in %d batches\n", len(txs), len(readyBatches))

	return &TxBatchResult{
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

// generateBatches generates batches of transactions
func (b *Batcher) generateBatches(txs [][]byte, batchSize int) ([]common.Batch, error) {
	var (
		batches      = generateBatches(txs, batchSize)
		numBatches   = len(batches)
		readyBatches = make([]common.Batch, numBatches)
	)

	fmt.Printf("\nGenerating batches...\n")

	bar := progressbar.Default(int64(numBatches), "batches generated")

	for index, batch := range batches {
		cliBatch := b.cli.CreateBatch()

		for _, tx := range batch {
			// Append the transaction
			if err := cliBatch.AddTxBroadcast(tx); err != nil {
				return nil, fmt.Errorf("unable to prepare transaction, %w", err)
			}
		}

		readyBatches[index] = cliBatch

		_ = bar.Add(1)
	}

	return readyBatches, nil
}

// sendBatches sends the prepared batch requests
func sendBatches(readyBatches []common.Batch) ([][]any, error) {
	var (
		numBatches   = len(readyBatches)
		batchResults = make([][]any, numBatches)
	)

	fmt.Printf("\nSending batches...\n")

	bar := progressbar.Default(int64(numBatches), "batches sent")

	for index, readyBatch := range readyBatches {
		batchResult, err := readyBatch.Execute()
		if err != nil {
			return nil, fmt.Errorf("unable to batch request, %w", err)
		}

		batchResults[index] = batchResult

		_ = bar.Add(1)
	}

	fmt.Printf("✅ Successfully sent %d batches\n", numBatches)

	return batchResults, nil
}

// parseBatchResults extracts transaction hashes
// from batch results
func parseBatchResults(batchResults [][]any, numTx int) ([][]byte, error) {
	var (
		txHashes = make([][]byte, numTx)
		index    = 0
	)

	fmt.Printf("\nParsing batch results...\n")

	bar := progressbar.Default(int64(numTx), "results parsed")

	// Parsing is done in a separate loop to not hinder
	// the batch send speed (as txs need to be parsed sequentially)
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

	fmt.Printf("✅ Successfully parsed %d batch results\n", len(batchResults))

	return txHashes, nil
}

// generateBatches generates data batches based on passed in params
func generateBatches(items [][]byte, batchSize int) [][][]byte {
	numBatches := int(math.Ceil(float64(len(items)) / float64(batchSize)))
	if numBatches == 0 {
		numBatches = 1
	}

	batches := make([][][]byte, numBatches)
	for i := 0; i < numBatches; i++ {
		batches[i] = make([][]byte, 0)
	}

	currentBatch := 0
	for _, item := range items {
		batches[currentBatch] = append(batches[currentBatch], item)

		if len(batches[currentBatch])%batchSize == 0 {
			currentBatch++
		}
	}

	return batches
}
