package internal

import (
	"errors"
	"regexp"

	"github.com/gnolang/gno/pkgs/crypto/bip39"
	"github.com/gnolang/supernova/internal/runtime"
)

var (
	errInvalidURL          = errors.New("invalid node URL specified")
	errInvalidMnemonic     = errors.New("invalid Mnemonic specified")
	errInvalidMode         = errors.New("invalid mode specified")
	errInvalidSubaccounts  = errors.New("invalid number of subaccounts specified")
	errInvalidTransactions = errors.New("invalid number of transactions specified")
	errInvalidBatchSize    = errors.New("invalid batch size specified")
)

var (
	// urlRegex is used for verifying the cluster's JSON-RPC endpoint
	urlRegex = regexp.MustCompile(`(https?://.*)(:(\d*)\/?(.*))?`)
)

// Config is the central pipeline configuration
type Config struct {
	URL      string // the URL of the cluster
	ChainID  string // the chain ID of the cluster
	Mnemonic string // the mnemonic for the keyring
	Mode     string // the stress test mode
	Output   string // output path for results JSON, if any

	SubAccounts  uint64 // the number of sub-accounts in the run
	Transactions uint64 // the total number of transactions
	BatchSize    uint64 // the maximum size of the batch
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
