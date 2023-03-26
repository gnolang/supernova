package internal

import (
	"fmt"

	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/bft/rpc/client"
	"github.com/gnolang/gno/pkgs/std"
)

type broadcaster struct {
	cli client.Client
}

// newBroadcaster creates a new instance of the broadcaster
func newBroadcaster(cli client.Client) *broadcaster {
	return &broadcaster{
		cli: cli,
	}
}

// BroadcastTxWithCommit broadcasts the transaction
// and waits for it to be committed
func (b *broadcaster) BroadcastTxWithCommit(tx *std.Tx) error {
	marshalledTx, err := amino.Marshal(tx)
	if err != nil {
		return fmt.Errorf("unable to marshal transaction, %w", err)
	}

	res, err := b.cli.BroadcastTxCommit(marshalledTx)
	if err != nil {
		return fmt.Errorf("unable to broadcast transaction, %w", err)
	}

	if res.CheckTx.IsErr() {
		return fmt.Errorf("broadcast transaction check failed, %w", res.CheckTx.Error)
	}

	if res.DeliverTx.IsErr() {
		return fmt.Errorf("broadcast transaction delivery failed, %w", res.DeliverTx.Error)
	}

	return nil
}
