package collector

import (
	"context"

	core_types "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
)

type (
	getBlockDelegate             func(ctx context.Context, height *int64) (*core_types.ResultBlock, error)
	getBlockGasUsedDelegate      func(ctx context.Context, height int64) (int64, error)
	getBlockGasLimitDelegate     func(ctx context.Context, height int64) (int64, error)
	getLatestBlockHeightDelegate func(ctx context.Context) (int64, error)
)

type mockClient struct {
	getBlockFn             getBlockDelegate
	getBlockGasUsedFn      getBlockGasUsedDelegate
	getBlockGasLimitFn     getBlockGasLimitDelegate
	getLatestBlockHeightFn getLatestBlockHeightDelegate
}

func (m *mockClient) GetBlock(ctx context.Context, height *int64) (*core_types.ResultBlock, error) {
	if m.getBlockFn != nil {
		return m.getBlockFn(ctx, height)
	}

	return nil, nil
}

func (m *mockClient) GetBlockGasUsed(ctx context.Context, height int64) (int64, error) {
	if m.getBlockGasUsedFn != nil {
		return m.getBlockGasUsedFn(ctx, height)
	}

	return 0, nil
}

func (m *mockClient) GetBlockGasLimit(ctx context.Context, height int64) (int64, error) {
	if m.getBlockGasLimitFn != nil {
		return m.getBlockGasLimitFn(ctx, height)
	}

	return 0, nil
}

func (m *mockClient) GetLatestBlockHeight(ctx context.Context) (int64, error) {
	if m.getLatestBlockHeightFn != nil {
		return m.getLatestBlockHeightFn(ctx)
	}

	return 0, nil
}
