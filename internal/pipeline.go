package internal

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/gnolang/gno/pkgs/bft/rpc/client"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"go.uber.org/zap"
)

const (
	encryptPassword = "encrypt"
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
	urlRegex = regexp.MustCompile(`(https?://.*)(:(\d*)\/?(.*))?`)
)

type Pipeline struct {
	ctx    context.Context
	logger *zap.Logger

	cfg     *Config
	keybase keys.Keybase
	cli     client.Client
}

// NewPipeline creates a new pipeline instance
func NewPipeline(ctx context.Context, logger *zap.Logger, cfg *Config) *Pipeline {
	return &Pipeline{
		ctx:     ctx,
		logger:  logger.Named("pipeline"),
		cfg:     cfg,
		keybase: keys.NewInMemory(),
		cli:     client.NewHTTP(cfg.URL, ""),
	}
}

func (p *Pipeline) Execute() error {
	_, err := newDistributor(p.logger, p.cli, p.keybase).distribute(
		p.cfg.SubAccounts,
		p.cfg.Transactions,
		p.cfg.Mnemonic,
	)
	if err != nil {
		return fmt.Errorf("unable to distribute funds, %w", err)
	}

	return nil
}
