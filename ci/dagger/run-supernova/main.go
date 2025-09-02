package main

import (
	"context"
	"dagger/run-supernova/internal/dagger"
	"fmt"
)

const (
	DEFAULT_CHAINID      = "dev"
	DEFAULT_SUBACCOUNTS  = 1
	DEFAULT_TRANSACTIONS = 10
	MNEMONIC             = "source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast"
)

type Supernova struct{}

// Builds Supernova image from code passed into a *dagger.Directory item
func (s *Supernova) BuildImage(directory *dagger.Directory) *dagger.Container {
	baseBuilder := dag.Container().
		From("golang:1.23-alpine").
		WithDirectory("/src", directory).
		WithWorkdir("/src").
		WithEnvVariable("CGO_ENABLED", "0").
		WithExec([]string{"go", "build", "-o", "supernova", "./cmd"}).
		WithExec([]string{"apk", "add", "--no-cache", "ca-certificates"})

	return dag.Container().
		From("busybox").
		WithFile("/bin/supernova", baseBuilder.File("/src/supernova")).
		WithFile("/etc/ssl/certs/ca-certificates.crt", baseBuilder.File("/etc/ssl/certs/ca-certificates.crt")).
		WithEntrypoint([]string{"/bin/supernova"})
}

// Build image from code or use latest prebuild Docker image
func (s *Supernova) buildOrPull(srcDir *dagger.Directory) *dagger.Container {
	if srcDir == nil {
		return dag.Container().
			From("ghcr.io/gnolang/supernova:latest")
	}
	return s.BuildImage(srcDir)
}

// Runs a simple Supernova task generating transactions
func (s *Supernova) RunStressTest(
	ctx context.Context,
	rpcEndpoint string,
	// +optional
	chainId string,
	// +optional
	subAccounts int,
	// +optional
	transactions int,
	// + optional
	srcDir *dagger.Directory,
) (int, error) {

	if chainId == "" {
		chainId = DEFAULT_CHAINID
	}
	if subAccounts == 0 {
		subAccounts = DEFAULT_SUBACCOUNTS
	}
	if transactions == 0 {
		transactions = DEFAULT_TRANSACTIONS
	}

	return s.buildOrPull(srcDir).
		WithExec([]string{
			"-sub-accounts", fmt.Sprintf("%d", subAccounts),
			"-transactions", fmt.Sprintf("%d", transactions),
			"-chain-id", chainId,
			"-url", rpcEndpoint,
			"-mnemonic", MNEMONIC},
			dagger.ContainerWithExecOpts{
				UseEntrypoint: true,
			}).
		ExitCode(ctx)
}
