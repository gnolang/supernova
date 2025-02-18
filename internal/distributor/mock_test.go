package distributor

import (
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type (
	broadcastTransactionDelegate func(*std.Tx) error
	getAccountDelegate           func(string) (*gnoland.GnoAccount, error)
	estimateGasDelegate          func(*std.Tx) (int64, error)
)

type mockClient struct {
	broadcastTransactionFn broadcastTransactionDelegate
	getAccountFn           getAccountDelegate
	estimateGasFn          estimateGasDelegate
}

func (m *mockClient) BroadcastTransaction(tx *std.Tx) error {
	if m.broadcastTransactionFn != nil {
		return m.broadcastTransactionFn(tx)
	}

	return nil
}

func (m *mockClient) GetAccount(address string) (*gnoland.GnoAccount, error) {
	if m.getAccountFn != nil {
		return m.getAccountFn(address)
	}

	return nil, nil
}

func (m *mockClient) EstimateGas(tx *std.Tx) (int64, error) {
	if m.estimateGasFn != nil {
		return m.estimateGasFn(tx)
	}

	return 0, nil
}
