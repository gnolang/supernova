package collector

import core_types "github.com/gnolang/gno/pkgs/bft/rpc/core/types"

type (
	getBlockDelegate             func(height *int64) (*core_types.ResultBlock, error)
	getBlockGasUsedDelegate      func(height int64) (int64, error)
	getBlockGasLimitDelegate     func(height int64) (int64, error)
	getLatestBlockHeightDelegate func() (int64, error)
)

type mockClient struct {
	getBlockFn             getBlockDelegate
	getBlockGasUsedFn      getBlockGasUsedDelegate
	getBlockGasLimitFn     getBlockGasLimitDelegate
	getLatestBlockHeightFn getLatestBlockHeightDelegate
}

func (m *mockClient) GetBlock(height *int64) (*core_types.ResultBlock, error) {
	if m.getBlockFn != nil {
		return m.getBlockFn(height)
	}

	return nil, nil
}

func (m *mockClient) GetBlockGasUsed(height int64) (int64, error) {
	if m.getBlockGasUsedFn != nil {
		return m.getBlockGasUsedFn(height)
	}

	return 0, nil
}

func (m *mockClient) GetBlockGasLimit(height int64) (int64, error) {
	if m.getBlockGasLimitFn != nil {
		return m.getBlockGasLimitFn(height)
	}

	return 0, nil
}

func (m *mockClient) GetLatestBlockHeight() (int64, error) {
	if m.getLatestBlockHeightFn != nil {
		return m.getLatestBlockHeightFn()
	}

	return 0, nil
}
