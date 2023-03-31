package batcher

import (
	"github.com/gnolang/supernova/internal/common"
)

type Client interface {
	CreateBatch() common.Batch
	GetLatestBlockHeight() (int64, error)
}

// TxBatchResult contains batching results
type TxBatchResult struct {
	TxHashes   [][]byte // the tx hashes
	StartBlock int64    // the initial block for querying
}
