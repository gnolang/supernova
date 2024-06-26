package distributor

import (
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type (
	broadcastTransactionDelegate func(*std.Tx) error
	getAccountDelegate           func(string) (*gnoland.GnoAccount, error)
)

type mockClient struct {
	broadcastTransactionFn broadcastTransactionDelegate
	getAccountFn           getAccountDelegate
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
