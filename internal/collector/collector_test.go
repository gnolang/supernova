package collector

import (
	"context"
	"crypto/rand"
	"testing"
	"time"

	core_types "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/tmhash"
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

func TestCollector_GetRunResults(t *testing.T) {
	t.Parallel()

	numTxs := 100
	startTime := time.Now()
	blockTimes := make([]time.Time, numTxs)

	for i := 0; i < numTxs; i++ {
		if i == 0 {
			blockTimes[i] = startTime
		}

		blockTimes[i] = startTime.Add(time.Duration(i) * time.Second)
	}

	txs := generateRandomData(t, numTxs)
	txHashes := make([][]byte, numTxs)

	for i := 0; i < numTxs; i++ {
		txHashes[i] = tmhash.Sum(txs[i])
	}

	var (
		gasLimit = int64(1000)
		gasUsed  = int64(100)

		mockClient = &mockClient{
			getBlockFn: func(ctx context.Context, height *int64) (*core_types.ResultBlock, error) {
				if *height > int64(numTxs) {
					t.Fatalf("invalid height requested")
				}

				return &core_types.ResultBlock{
					BlockMeta: &types.BlockMeta{
						Header: types.Header{
							Height: *height,
							Time:   blockTimes[*height-1],
							NumTxs: 1,
						},
					},
					Block: &types.Block{
						Data: types.Data{
							Txs: []types.Tx{
								txs[*height-1],
							},
						},
					},
				}, nil
			},
			getLatestBlockHeightFn: func(ctx context.Context) (int64, error) {
				return int64(numTxs), nil
			},
			getBlockGasLimitFn: func(ctx context.Context, height int64) (int64, error) {
				if height > int64(numTxs) {
					t.Fatalf("invalid height requested")
				}

				return gasLimit, nil
			},
			getBlockGasUsedFn: func(ctx context.Context, height int64) (int64, error) {
				if height > int64(numTxs) {
					t.Fatalf("invalid height requested")
				}

				return gasUsed, nil
			},
		}
	)

	// Create the collector
	c := NewCollector(context.Background(), mockClient)
	c.requestTimeout = time.Second * 0

	// Collect the results
	result, err := c.GetRunResult(txHashes, 1, startTime)
	if err != nil {
		t.Fatalf("unable to get run results, %v", err)
	}

	if result == nil {
		t.Fatal("result should not be nil")
	}

	assert.NotZero(t, result.AverageTPS)
	assert.Len(t, result.Blocks, numTxs)

	for index, block := range result.Blocks {
		assert.Equal(t, int64(index+1), block.Number)
		assert.Equal(t, gasUsed, block.GasUsed)
		assert.Equal(t, gasLimit, block.GasLimit)
		assert.Equal(t, int64(1), block.Transactions)
	}
}
