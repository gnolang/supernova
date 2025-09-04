package collector

import (
	"context"
	"time"

	core_types "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
)

type Client interface {
	GetBlock(ctx context.Context, height *int64) (*core_types.ResultBlock, error)
	GetBlockGasUsed(ctx context.Context, height int64) (int64, error)
	GetBlockGasLimit(ctx context.Context, height int64) (int64, error)
	GetLatestBlockHeight(ctx context.Context) (int64, error)
}

// RunResult is the complete test-run result
type RunResult struct {
	Blocks     []*BlockResult `json:"blocks"`
	AverageTPS float64        `json:"averageTPS"`
}

// BlockResult is the single-block test run result
type BlockResult struct {
	Time         time.Time `json:"created"`
	Number       int64     `json:"blockNumber"`
	Transactions int64     `json:"numTransactions"`
	GasUsed      int64     `json:"gasUsed"`
	GasLimit     int64     `json:"gasLimit"`
}
