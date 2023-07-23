package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultWorkers(t *testing.T) {
	// Should never have 0 workers
	for bufferSize := 0; bufferSize < 100; bufferSize += 1 {
		for workerCount := 0; workerCount < 100; workerCount += 1 {
			assert.NotZero(t, defaultWorkers(bufferSize, workerCount))
		}
	}
	// No workerCount
	assert.Equal(t, 20, defaultWorkers(203, 0))
	// Uses workerCount
	assert.Equal(t, 30, defaultWorkers(10, 30))
}
