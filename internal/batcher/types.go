package batcher

import (
	"context"

	"github.com/gnolang/supernova/internal/common"
)

type Client interface {
	CreateBatch() common.Batch
	GetLatestBlockHeight(ctx context.Context) (int64, error)
}

// TxBatchResult contains batching results
type TxBatchResult struct {
	TxHashes   [][]byte // the tx hashes
	StartBlock int64    // the initial block for querying
}
