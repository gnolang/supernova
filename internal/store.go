package internal

import (
	"fmt"

	"github.com/gnolang/gno/gnoland"
	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/bft/rpc/client"
)

type store struct {
	cli client.Client
}

func newStore(cli client.Client) *store {
	return &store{
		cli: cli,
	}
}

func (s *store) GetAccount(address string) (*gnoland.GnoAccount, error) {
	queryResult, err := s.cli.ABCIQuery(
		fmt.Sprintf("auth/accounts/%s", address),
		[]byte{},
	)

	if err != nil {
		return nil, fmt.Errorf("unable to fetch account %s, %w", address, err)
	}

	if queryResult.Response.IsErr() {
		return nil, fmt.Errorf("invalid account query result, %w", queryResult.Response.Error)
	}

	var acc gnoland.GnoAccount
	if err := amino.UnmarshalJSON(queryResult.Response.Data, &acc); err != nil {
		return nil, fmt.Errorf("unable to unmarshal query response, %w", err)
	}

	return &acc, nil
}
