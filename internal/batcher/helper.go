package batcher

import "math"

// generateBatches generates data batches based on passed in params
func generateBatches[T []byte | string](items []T, batchSize int) [][]T {
	numBatches := int(math.Ceil(float64(len(items)) / float64(batchSize)))
	if numBatches == 0 {
		numBatches = 1
	}

	batches := make([][]T, numBatches)
	for i := 0; i < numBatches; i++ {
		batches[i] = make([]T, 0)
	}

	currentBatch := 0
	for _, item := range items {
		batches[currentBatch] = append(batches[currentBatch], item)

		if len(batches[currentBatch])%batchSize == 0 {
			currentBatch++
		}
	}

	return batches
}
