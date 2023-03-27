package internal

import (
	"github.com/gnolang/gno/pkgs/crypto/bip39"
	"github.com/gnolang/supernova/internal/runtime"
)

type Config struct {
	URL      string
	ChainID  string
	Mnemonic string
	Mode     string
	Output   string

	SubAccounts  uint64
	Transactions uint64
	BatchSize    uint64
}

// Validate validates the stress-test configuration
func (cfg *Config) Validate() error {
	// Make sure the URL is valid
	if !urlRegex.MatchString(cfg.URL) {
		return errInvalidURL
	}

	// Make sure the mnemonic is valid
	if !bip39.IsMnemonicValid(cfg.Mnemonic) {
		return errInvalidMnemonic
	}

	// Make sure the mode is valid
	if !runtime.IsRuntime(runtime.Type(cfg.Mode)) {
		return errInvalidMode
	}

	// Make sure the number of subaccounts is valid
	if cfg.SubAccounts < 1 {
		return errInvalidSubaccounts
	}

	// Make sure the number of transactions is valid
	if cfg.Transactions < 1 {
		return errInvalidTransactions
	}

	// Make sure the batch size is valid
	if cfg.BatchSize < 1 {
		return errInvalidBatchSize
	}

	return nil
}
