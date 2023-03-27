package signer

import (
	"github.com/gnolang/gno/gnoland"
	"github.com/gnolang/gno/pkgs/std"
)

type Signer interface {
	SignTx(tx *std.Tx, account *gnoland.GnoAccount, nonce uint64, passphrase string) error
}
