package collector

import "time"

// RunResult is the complete test-run result
type RunResult struct {
	AverageTPS int            `json:"averageTPS"`
	Blocks     []*BlockResult `json:"blocks"`
}

// BlockResult is the single-block test run result
type BlockResult struct {
	Number       int64     `json:"blockNumber"`
	Time         time.Time `json:"created"`
	Transactions int64     `json:"numTransactions"`
	GasUsed      int64     `json:"gasUsed"`
	GasLimit     int64     `json:"gasLimit"`
}
