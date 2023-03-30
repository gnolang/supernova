package runtime

import (
	"github.com/gnolang/gno/gnoland"
	"github.com/gnolang/gno/pkgs/std"
)

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
