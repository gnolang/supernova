package client

import (
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	core_types "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/supernova/internal/common"
)

const (
	simulatePath = ".app/simulate"
	gaspricePath = "auth/gasprice"
)

type Client struct {
	conn *client.RPCClient
}

// NewWSClient creates a new instance of the WS client
func NewWSClient(url string) (*Client, error) {
	cli, err := client.NewWSClient(url)
	if err != nil {
		return nil, fmt.Errorf("unable to create ws client, %w", err)
	}

	return &Client{
		conn: cli,
	}, nil
}

// NewHTTPClient creates a new instance of the HTTP client
func NewHTTPClient(url string) (*Client, error) {
	cli, err := client.NewHTTPClient(url)
	if err != nil {
		return nil, fmt.Errorf("unable to create http client, %w", err)
	}

	return &Client{
		conn: cli,
	}, nil
}

func (h *Client) CreateBatch() common.Batch {
	return &Batch{batch: h.conn.NewBatch()}
}

func (h *Client) ExecuteABCIQuery(path string, data []byte) (*core_types.ResultABCIQuery, error) {
	return h.conn.ABCIQuery(path, data)
}

func (h *Client) GetLatestBlockHeight() (int64, error) {
	status, err := h.conn.Status()
	if err != nil {
		return 0, fmt.Errorf("unable to fetch status, %w", err)
	}

	return status.SyncInfo.LatestBlockHeight, nil
}

func (h *Client) GetBlock(height *int64) (*core_types.ResultBlock, error) {
	return h.conn.Block(height)
}

func (h *Client) GetBlockResults(height *int64) (*core_types.ResultBlockResults, error) {
	return h.conn.BlockResults(height)
}

func (h *Client) GetConsensusParams(height *int64) (*core_types.ResultConsensusParams, error) {
	return h.conn.ConsensusParams(height)
}

func (h *Client) BroadcastTransaction(tx *std.Tx) error {
	marshalledTx, err := amino.Marshal(tx)
	if err != nil {
		return fmt.Errorf("unable to marshal transaction, %w", err)
	}

	res, err := h.conn.BroadcastTxCommit(marshalledTx)
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

func (h *Client) GetAccount(address string) (*gnoland.GnoAccount, error) {
	queryResult, err := h.conn.ABCIQuery(
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

func (h *Client) GetBlockGasUsed(height int64) (int64, error) {
	blockRes, err := h.conn.BlockResults(&height)
	if err != nil {
		return 0, fmt.Errorf("unable to fetch block results, %w", err)
	}

	gasUsed := int64(0)
	for _, tx := range blockRes.Results.DeliverTxs {
		gasUsed += tx.GasUsed
	}

	return gasUsed, nil
}

func (h *Client) GetBlockGasLimit(height int64) (int64, error) {
	consensusParams, err := h.conn.ConsensusParams(&height)
	if err != nil {
		return 0, fmt.Errorf("unable to fetch block info, %w", err)
	}

	return consensusParams.ConsensusParams.Block.MaxGas, nil
}

func (h *Client) EstimateGas(tx *std.Tx) (int64, error) {
	// Prepare the transaction.
	// The transaction needs to be amino-binary encoded
	// in order to be estimated
	encodedTx, err := amino.Marshal(tx)
	if err != nil {
		return 0, fmt.Errorf("unable to marshal tx: %w", err)
	}

	// Perform the simulation query
	resp, err := h.conn.ABCIQuery(simulatePath, encodedTx)
	if err != nil {
		return 0, fmt.Errorf("unable to perform ABCI query: %w", err)
	}

	// Extract the query response
	if err = resp.Response.Error; err != nil {
		return 0, fmt.Errorf("error encountered during ABCI query: %w", err)
	}

	var deliverTx abci.ResponseDeliverTx
	if err = amino.Unmarshal(resp.Response.Value, &deliverTx); err != nil {
		return 0, fmt.Errorf("unable to unmarshal gas estimation response: %w", err)
	}

	if err = deliverTx.Error; err != nil {
		return 0, fmt.Errorf("error encountered during gas estimation: %w", err)
	}

	// Return the actual value returned by the node
	// for executing the transaction
	return deliverTx.GasUsed, nil
}

func (h *Client) FetchGasPrice() (std.GasPrice, error) {
	// Perform auth/gasprice
	gp := std.GasPrice{}

	qres, err := h.conn.ABCIQuery(gaspricePath, []byte{})
	if err != nil {
		return gp, err
	}

	err = amino.UnmarshalJSON(qres.Response.Data, &gp)
	if err != nil {
		return gp, err
	}

	return gp, nil
}
