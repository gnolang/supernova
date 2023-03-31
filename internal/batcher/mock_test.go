package batcher

import (
	"github.com/gnolang/supernova/internal/common"
)

type (
	createBatchDelegate          func() common.Batch
	getLatestBlockHeightDelegate func() (int64, error)
)

type mockClient struct {
	createBatchFn          createBatchDelegate
	getLatestBlockHeightFn getLatestBlockHeightDelegate
}

func (m *mockClient) CreateBatch() common.Batch {
	if m.createBatchFn != nil {
		return m.createBatchFn()
	}

	return nil
}

func (m *mockClient) GetLatestBlockHeight() (int64, error) {
	if m.getLatestBlockHeightFn != nil {
		return m.getLatestBlockHeightFn()
	}

	return 0, nil
}

type (
	addTxBroadcastDelegate func(tx []byte) error
	executeDelegate        func() ([]interface{}, error)
)

type mockBatch struct {
	addTxBroadcastFn addTxBroadcastDelegate
	executeFn        executeDelegate
}

func (m *mockBatch) AddTxBroadcast(tx []byte) error {
	if m.addTxBroadcastFn != nil {
		return m.addTxBroadcastFn(tx)
	}

	return nil
}

func (m *mockBatch) Execute() ([]interface{}, error) {
	if m.executeFn != nil {
		return m.executeFn()
	}

	return nil, nil
}
