package collector

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/gnolang/gno/pkgs/bft/rpc/client"
	"github.com/schollz/progressbar/v3"
)

type Collector struct {
	cli client.Client
}

func NewCollector(cli client.Client) *Collector {
	return &Collector{
		cli: cli,
	}
}

type RunResult struct {
	AverageTPS int            `json:"averageTPS"`
	Blocks     []*BlockResult `json:"blocks"`
}

type BlockResult struct {
	Number       int64     `json:"blockNumber"`
	Time         time.Time `json:"created"`
	Transactions int64     `json:"numTransactions"`
	GasUsed      int64     `json:"gasUsed"`
	GasLimit     int64     `json:"gasLimit"`
}

func (c *Collector) CollectTransactions(txHashes [][]byte, startBlock int64) (*RunResult, error) {
	blockResults := make([]*BlockResult, 0)
	timeout := time.After(5 * time.Minute)
	start := startBlock

	txMap := make(map[string]struct{})
	processed := 0

	for _, tx := range txHashes {
		txMap[string(tx)] = struct{}{}
	}

	bar := progressbar.Default(int64(len(txHashes)), "txs collected")

	for {
		if processed >= len(txHashes) {
			if processed > len(txHashes) {
				fmt.Printf("\nwtf %d\n", processed)
			}

			break
		}

		select {
		case <-timeout:
			return nil, errors.New("collector timed out")
		case <-time.After(2 * time.Second):
			status, err := c.cli.Status()
			if err != nil {
				return nil, fmt.Errorf("unable to fetch node status, %w", err)
			}

			latest := status.SyncInfo.LatestBlockHeight
			if latest < start {
				continue
			}

			for i := start; i <= latest; i++ {
				relevantBlock := false

				block, err := c.cli.Block(&i)
				if err != nil {
					return nil, fmt.Errorf("unable to fetch block, %w", err)
				}

				for _, tx := range block.Block.Txs {
					txHash := tx.Hash()

					if _, ok := txMap[string(txHash)]; ok {
						processed++

						relevantBlock = true
						_ = bar.Add(1)
					}
				}

				if !relevantBlock {
					continue
				}

				blockRes, err := c.cli.BlockResults(&i)
				if err != nil {
					return nil, fmt.Errorf("unable to fetch block results, %w", err)
				}

				consensusParams, err := c.cli.ConsensusParams(&i)
				if err != nil {
					return nil, fmt.Errorf("unable to fetch block info, %w", err)
				}

				blockResult := &BlockResult{
					Number:       i,
					Time:         block.BlockMeta.Header.Time,
					Transactions: block.BlockMeta.Header.NumTxs,
					GasUsed:      0,
					GasLimit:     consensusParams.ConsensusParams.Block.MaxGas,
				}

				for _, tx := range blockRes.Results.DeliverTxs {
					blockResult.GasUsed += tx.GasUsed
				}

				blockResults = append(blockResults, blockResult)
			}

			start = latest + 1
		}
	}

	// Calculate TPS
	firstBlock := blockResults[0]
	lastBlock := blockResults[len(blockResults)-1]

	diff := lastBlock.Time.Sub(firstBlock.Time).Seconds()

	return &RunResult{
		AverageTPS: int(math.Ceil(float64(len(txHashes)) / diff)),
		Blocks:     blockResults,
	}, nil
}
