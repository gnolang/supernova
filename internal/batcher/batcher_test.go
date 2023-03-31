package batcher

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"testing"

	core_types "github.com/gnolang/gno/pkgs/bft/rpc/core/types"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gnolang/supernova/internal/common"
	"github.com/stretchr/testify/assert"
)

// generateRandomData generates random 32B chunks
func generateRandomData(t *testing.T, count int) [][]byte {
	t.Helper()

	data := make([][]byte, count)

	for i := 0; i < count; i++ {
		buf := make([]byte, 32)

		_, err := rand.Read(buf)
		if err != nil {
			t.Fatalf("unable to generate random data, %v", err)
		}

		data[i] = buf
	}

	return data
}

// generateTestTransactions generates test transactions
func generateTestTransactions(count int) []*std.Tx {
	data := make([]*std.Tx, count)

	for i := 0; i < count; i++ {
		data[i] = &std.Tx{
			Memo: fmt.Sprintf("tx-%d", i),
		}
	}

	return data
}

func TestBatcher_BatchTransactions(t *testing.T) {
	t.Parallel()

	var (
		numTxs    = 100
		batchSize = 20
		txs       = generateTestTransactions(numTxs)
		txHashes  = generateRandomData(t, numTxs)

		broadcastTxs = make([][]byte, 0)
		currIndex    = 0

		mockBatch = &mockBatch{
			addTxBroadcastFn: func(tx []byte) error {
				broadcastTxs = append(broadcastTxs, tx)

				return nil
			},
			executeFn: func() ([]interface{}, error) {
				res := make([]any, batchSize)

				for i := 0; i < batchSize; i++ {
					res[i] = &core_types.ResultBroadcastTx{
						Hash: txHashes[currIndex],
					}

					currIndex++
				}

				return res, nil
			},
		}
		mockClient = &mockClient{
			createBatchFn: func() common.Batch {
				return mockBatch
			},
		}
	)

	// Create the batcher
	b := NewBatcher(mockClient)

	// Batch the transactions
	res, err := b.BatchTransactions(txs, batchSize)
	if err != nil {
		t.Fatalf("unable to batch transactions, %v", err)
	}

	assert.NotNil(t, res)

	if len(res.TxHashes) != numTxs {
		t.Fatalf("invalid tx hashes returned, %d", len(res.TxHashes))
	}

	for index, txHash := range txHashes {
		assert.True(t, bytes.Equal(txHash, txHashes[index]))
	}
}
