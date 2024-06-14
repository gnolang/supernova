package client

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
)

type Batch struct {
	batch *client.RPCBatch
}

func (b *Batch) AddTxBroadcast(tx []byte) error {
	if err := b.batch.BroadcastTxSync(tx); err != nil {
		return fmt.Errorf("unable to prepare transaction, %w", err)
	}

	return nil
}

func (b *Batch) Execute() ([]interface{}, error) {
	return b.batch.Send(context.Background())
}
