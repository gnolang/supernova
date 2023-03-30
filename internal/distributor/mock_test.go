package distributor

import (
	"github.com/gnolang/gno/gnoland"
	"github.com/gnolang/gno/pkgs/std"
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

type signTxDelegate func(*std.Tx, *gnoland.GnoAccount, uint64, string) error

type mockSigner struct {
	signTxFn signTxDelegate
}

func (m *mockSigner) SignTx(tx *std.Tx, account *gnoland.GnoAccount, nonce uint64, passphrase string) error {
	if m.signTxFn != nil {
		return m.signTxFn(tx, account, nonce, passphrase)
	}

	return nil
}
