package internal

import (
	"fmt"

	"github.com/gnolang/gno/gnoland"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/std"
)

type signer struct {
	keybase keys.Keybase
}

// newSigner creates a new signer instance
func newSigner(keybase keys.Keybase) *signer {
	return &signer{
		keybase: keybase,
	}
}

// SignTx signs the given transaction by appending the
// signature to it
func (s *signer) SignTx(
	tx *std.Tx,
	account *gnoland.GnoAccount,
	nonce uint64,
	passphrase string,
) error {
	// Fetch existing signers
	signers := tx.GetSigners()
	if tx.Signatures == nil {
		for range signers {
			tx.Signatures = append(tx.Signatures, std.Signature{
				PubKey:    nil, // zero signature
				Signature: nil, // zero signature
			})
		}
	}

	// Generate the signature
	signature, pub, err := s.keybase.Sign(
		account.GetAddress().String(),
		passphrase,
		tx.GetSignBytes("dev", account.AccountNumber, nonce),
	)
	if err != nil {
		return fmt.Errorf("unable to sign transaction, %w", err)
	}

	addr := pub.Address()
	found := false

	// Append the signature to the correct slot
	for i := range tx.Signatures {
		if signers[i] == addr {
			found = true
			tx.Signatures[i] = std.Signature{
				PubKey:    pub,
				Signature: signature,
			}
		}
	}

	if !found {
		return fmt.Errorf("unable to sign transaction with address %s", account.GetAddress())
	}

	return nil
}
