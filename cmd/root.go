package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/supernova/internal"
	"github.com/gnolang/supernova/internal/runtime"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func main() {
	var (
		cfg = &internal.Config{}
		fs  = flag.NewFlagSet("pipeline", flag.ExitOnError)
	)

	// Register the flags
	registerFlags(fs, cfg)

	cmd := &ffcli.Command{
		ShortUsage: "[flags] [<arg>...]",
		LongHelp:   "Starts the stress testing suite against a Gno TM2 cluster",
		FlagSet:    fs,
		Exec: func(_ context.Context, _ []string) error {
			return execMain(cfg)
		},
	}

	if err := cmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v", err)

		os.Exit(1)
	}
}

// registerFlags registers the main configuration flags
func registerFlags(fs *flag.FlagSet, c *internal.Config) {
	fs.StringVar(
		&c.URL,
		"url",
		"",
		"the JSON-RPC URL of the cluster",
	)

	fs.StringVar(
		&c.ChainID,
		"chain-id",
		"dev",
		"the chain ID of the Gno blockchain",
	)

	fs.StringVar(
		&c.Mnemonic,
		"mnemonic",
		"",
		"the mnemonic used to generate sub-accounts",
	)

	fs.StringVar(
		&c.Mode,
		"mode",
		runtime.RealmDeployment.String(),
		fmt.Sprintf(
			"the mode for the stress test. Possible modes: [%s, %s, %s]",
			runtime.RealmDeployment.String(), runtime.PackageDeployment.String(), runtime.RealmCall.String(),
		),
	)

	fs.StringVar(
		&c.Output,
		"output",
		"",
		"the output path for the results JSON",
	)

	fs.Uint64Var(
		&c.SubAccounts,
		"sub-accounts",
		10,
		"the number of sub-accounts that will send out transactions",
	)

	fs.Uint64Var(
		&c.Transactions,
		"transactions",
		100,
		"the total number of transactions to be emitted",
	)

	fs.Uint64Var(
		&c.BatchSize,
		"batch",
		20,
		"the batch size of JSON-RPC transactions",
	)
}

// execMain starts the stress test workflow (runs the pipeline)
func execMain(cfg *internal.Config) error {
	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration, %w", err)
	}

	// Create and run the pipeline
	return internal.NewPipeline(cfg).Execute()
}
