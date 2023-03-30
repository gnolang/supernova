package common

// Batch is a common transaction batch
type Batch interface {
	// AddTxBroadcast adds the transaction broadcast to the batch
	AddTxBroadcast(tx []byte) error

	// Execute executes the batch send
	Execute() ([]interface{}, error)
}
