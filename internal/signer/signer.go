package signer

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/hd"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// SignCfg specifies the sign configuration
type SignCfg struct {
	ChainID       string // the ID of the chain
	AccountNumber uint64 // the account number of the signer
	Sequence      uint64 // the Sequence of the signer
}

// SignTx signs the specified transaction using
// the provided key and config
func SignTx(tx *std.Tx, key crypto.PrivKey, cfg SignCfg) error {
	// Get the sign bytes
	signBytes, err := tx.GetSignBytes(
		cfg.ChainID,
		cfg.AccountNumber,
		cfg.Sequence,
	)
	if err != nil {
		return fmt.Errorf("unable to get tx signature payload, %w", err)
	}

	// Sign the transaction
	signature, err := key.Sign(signBytes)
	if err != nil {
		return fmt.Errorf("unable to sign transaction, %w", err)
	}

	// Save the signature
	tx.Signatures = append(tx.Signatures, std.Signature{
		PubKey:    key.PubKey(),
		Signature: signature,
	})

	return nil
}

// GenerateKeyFromSeed generates a private key from
// the provided seed and index
func GenerateKeyFromSeed(seed []byte, index uint32) crypto.PrivKey {
	pathParams := hd.NewFundraiserParams(0, crypto.CoinType, index)

	masterPriv, ch := hd.ComputeMastersFromSeed(seed)

	//nolint:errcheck // This derivation can never error out, since the path params
	// are always going to be valid
	derivedPriv, _ := hd.DerivePrivateKeyForPath(masterPriv, ch, pathParams.String())

	return secp256k1.PrivKeySecp256k1(derivedPriv)
}
