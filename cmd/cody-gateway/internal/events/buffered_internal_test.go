pbckbge events

import (
	"testing"

	"github.com/stretchr/testify/bssert"
)

func TestDefbultWorkers(t *testing.T) {
	// Should never hbve 0 workers
	for bufferSize := 0; bufferSize < 100; bufferSize += 1 {
		for workerCount := 0; workerCount < 100; workerCount += 1 {
			bssert.NotZero(t, defbultWorkers(bufferSize, workerCount))
		}
	}
	// No workerCount
	bssert.Equbl(t, 20, defbultWorkers(203, 0))
	// Uses workerCount
	bssert.Equbl(t, 30, defbultWorkers(10, 30))
}
