package testutils

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/supernova/internal/signer"
)

// GenerateMnemonic generates a new BIP39 mnemonic
func GenerateMnemonic(t *testing.T) string {
	t.Helper()

	// Generate the entropy seed
	entropySeed, err := bip39.NewEntropy(256)
	if err != nil {
		t.Fatalf("unable to generate entropy seed, %v", err)
	}

	// Generate the actual mnemonic
	mnemonic, err := bip39.NewMnemonic(entropySeed[:])
	if err != nil {
		t.Fatalf("unable to generate mnemonic, %v", err)
	}

	return mnemonic
}

// GenerateAccounts generates mock keybase accounts
func GenerateAccounts(t *testing.T, count int) []crypto.PrivKey {
	t.Helper()

	var (
		accounts = make([]crypto.PrivKey, count)
		mnemonic = GenerateMnemonic(t)
		seed     = bip39.NewSeed(mnemonic, "")
	)

	for i := 0; i < count; i++ {
		accounts[i] = signer.GenerateKeyFromSeed(seed, uint32(i))
	}

	return accounts
}
