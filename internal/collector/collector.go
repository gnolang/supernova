package collector

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/gnolang/gno/pkgs/bft/types"
	"github.com/schollz/progressbar/v3"
)

var (
	errTimeout = errors.New("collector timed out")
)

// Collector is the transaction / block stat
// collector.
// This implementation will heavily change when
// transaction indexing is introduced
type Collector struct {
	cli Client

	requestTimeout time.Duration
}

// NewCollector creates a new instance of the collector
func NewCollector(cli Client) *Collector {
	return &Collector{
		cli:            cli,
		requestTimeout: time.Second * 2,
	}
}

// GetRunResult generates the run result for the passed in transaction hashes and start range
func (c *Collector) GetRunResult(
	txHashes [][]byte,
	startBlock int64,
	startTime time.Time,
) (*RunResult, error) {
	var (
		blockResults = make([]*BlockResult, 0)
		timeout      = time.After(5 * time.Minute)
		start        = startBlock
		txMap        = newTxLookup(txHashes)
		processed    = 0
	)

	fmt.Printf("\n📊 Collecting Results 📊\n\n")

	bar := progressbar.Default(int64(len(txHashes)), "txs collected")

	for {
		// Check if all original transactions
		// were processed
		if processed >= len(txHashes) {
			break
		}

		select {
		case <-timeout:
			return nil, errTimeout
		case <-time.After(c.requestTimeout):
			latest, err := c.cli.GetLatestBlockHeight()
			if err != nil {
				return nil, fmt.Errorf("unable to fetch latest block height, %w", err)
			}

			if latest < start {
				// No need to parse older blocks
				continue
			}

			// Iterate over each block and find relevant transactions
			for blockNum := start; blockNum <= latest; blockNum++ {
				// Fetch the block
				block, err := c.cli.GetBlock(&blockNum)
				if err != nil {
					return nil, fmt.Errorf("unable to fetch block, %w", err)
				}

				// Check if any of the block transactions are the ones
				// sent out in the stress test
				belong := txMap.anyBelong(block.Block.Txs)
				if belong == 0 {
					continue
				}

				processed += belong
				_ = bar.Add(belong)

				// Fetch the total gas used by transactions
				blockGasUsed, err := c.cli.GetBlockGasUsed(blockNum)
				if err != nil {
					return nil, fmt.Errorf("unable to fetch block gas used, %w", err)
				}

				// Fetch the block gas limit
				blockGasLimit, err := c.cli.GetBlockGasLimit(blockNum)
				if err != nil {
					return nil, fmt.Errorf("unable to fetch block gas limit, %w", err)
				}

				blockResults = append(blockResults, &BlockResult{
					Number:       blockNum,
					Time:         block.BlockMeta.Header.Time,
					Transactions: block.BlockMeta.Header.NumTxs,
					GasUsed:      blockGasUsed,
					GasLimit:     blockGasLimit,
				})
			}

			// Update the iteration range
			start = latest + 1
		}
	}

	return &RunResult{
		AverageTPS: calculateTPS(
			startTime,
			blockResults[len(blockResults)-1].Time,
			len(txHashes),
		),
		Blocks: blockResults,
	}, nil
}

// txLookup is a simple lookup map for transaction hashes
type txLookup struct {
	lookup map[string]struct{}
}

// newTxLookup creates a new instance of the tx lookup map
func newTxLookup(txs [][]byte) *txLookup {
	lookup := make(map[string]struct{})

	for _, tx := range txs {
		lookup[string(tx)] = struct{}{}
	}

	return &txLookup{
		lookup: lookup,
	}
}

// anyBelong returns the number of transactions
// that have been found in the lookup map
func (t *txLookup) anyBelong(txs types.Txs) int {
	belong := 0

	for _, tx := range txs {
		txHash := tx.Hash()

		if _, ok := t.lookup[string(txHash)]; ok {
			belong++
		}
	}

	return belong
}

// calculateTPS calculates the TPS for the sequence
func calculateTPS(startBlock, endBlock time.Time, totalTx int) int {
	diff := endBlock.Sub(startBlock).Seconds()

	// ceil(numTxs / commit time)
	return int(math.Ceil(float64(totalTx) / diff))
}
