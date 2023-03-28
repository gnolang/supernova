package signer

import (
	"github.com/gnolang/gno/gnoland"
	"github.com/gnolang/gno/pkgs/std"
)

type signTxDelegate func(*std.Tx, *gnoland.GnoAccount, uint64, string) error

type MockSigner struct {
	SignTxFn signTxDelegate
}

func (m *MockSigner) SignTx(tx *std.Tx, account *gnoland.GnoAccount, nonce uint64, passphrase string) error {
	if m.SignTxFn != nil {
		return m.SignTxFn(tx, account, nonce, passphrase)
	}

	return nil
}
