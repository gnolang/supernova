package internal

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/gnolang/supernova/internal/collector"
)

// displayResults displays the runtime result in the terminal
func displayResults(result *collector.RunResult) {
	w := tabwriter.NewWriter(os.Stdout, 10, 20, 2, ' ', 0)

	// TPS //
	_, _ = fmt.Fprintln(w, fmt.Sprintf("\nTPS: %d", result.AverageTPS))

	// Block info //
	_, _ = fmt.Fprintln(w, "\nBlock #\tGas Used\tGas Limit\tTransactions\tUtilization")
	for _, block := range result.Blocks {
		_, _ = fmt.Fprintln(
			w,
			fmt.Sprintf(
				"Block #%d\t%d\t%d\t%d\t%.2f%%",
				block.Number,
				block.GasUsed,
				block.GasLimit,
				block.Transactions,
				(float64(block.GasUsed)/float64(block.GasLimit))*100,
			),
		)
	}

	_, _ = fmt.Fprintln(w, "")

	_ = w.Flush()
}
