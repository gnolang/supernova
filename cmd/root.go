package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
)

// mainCfg is the main runtime configuration
type mainCfg struct {
}

// registerFlags registers the main configuration flags
func (c *mainCfg) registerFlags(fs *flag.FlagSet) {
	// TODO register flags
}

func main() {
	cfg := &mainCfg{}
	fs := flag.NewFlagSet("", flag.ExitOnError)

	cfg.registerFlags(fs)

	cmd := &ffcli.Command{
		ShortUsage: "[flags] [<arg>...]",
		LongHelp:   "Starts the stress testing suite against a TM2 cluster",
		FlagSet:    fs,
		Exec: func(ctx context.Context, _ []string) error {
			return execMain(ctx, cfg)
		},
	}

	if err := cmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v", err)

		os.Exit(1)
	}
}

// execMain starts the stress test workflow (runs the pipeline)
func execMain(ctx context.Context, cfg *mainCfg) error {
	// TODO start the test workflow
	return nil
}
