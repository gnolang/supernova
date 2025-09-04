package distributor

import (
	"context"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type (
	broadcastTransactionDelegate func(context.Context, *std.Tx) error
	getAccountDelegate           func(context.Context, string) (*gnoland.GnoAccount, error)
	estimateGasDelegate          func(context.Context, *std.Tx) (int64, error)
	fetchGasPriceDelegate        func(context.Context) (std.GasPrice, error)
)

type mockClient struct {
	broadcastTransactionFn broadcastTransactionDelegate
	getAccountFn           getAccountDelegate
	estimateGasFn          estimateGasDelegate
	fetchGasPriceFn        fetchGasPriceDelegate
}

func (m *mockClient) BroadcastTransaction(ctx context.Context, tx *std.Tx) error {
	if m.broadcastTransactionFn != nil {
		return m.broadcastTransactionFn(ctx, tx)
	}

	return nil
}

func (m *mockClient) GetAccount(ctx context.Context, address string) (*gnoland.GnoAccount, error) {
	if m.getAccountFn != nil {
		return m.getAccountFn(ctx, address)
	}

	return nil, nil
}

func (m *mockClient) EstimateGas(ctx context.Context, tx *std.Tx) (int64, error) {
	if m.estimateGasFn != nil {
		return m.estimateGasFn(ctx, tx)
	}

	return 0, nil
}

func (m *mockClient) FetchGasPrice(ctx context.Context) (std.GasPrice, error) {
	if m.fetchGasPriceFn != nil {
		return m.fetchGasPriceFn(ctx)
	}

	return std.GasPrice{}, nil
}
