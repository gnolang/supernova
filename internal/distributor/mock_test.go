package distributor

import (
	"github.com/gnolang/gno/gnoland"
	"github.com/gnolang/gno/pkgs/std"
)

type broadcastTxWithCommitDelegate func(*std.Tx) error

type mockTxBroadcaster struct {
	broadcastTxWithCommitFn broadcastTxWithCommitDelegate
}

func (m *mockTxBroadcaster) BroadcastTxWithCommit(tx *std.Tx) error {
	if m.broadcastTxWithCommitFn != nil {
		return m.broadcastTxWithCommitFn(tx)
	}

	return nil
}

type getAccountDelegate func(string) (*gnoland.GnoAccount, error)

type mockAccountStore struct {
	getAccountFn getAccountDelegate
}

func (m *mockAccountStore) GetAccount(address string) (*gnoland.GnoAccount, error) {
	if m.getAccountFn != nil {
		return m.getAccountFn(address)
	}

	return nil, nil
}
