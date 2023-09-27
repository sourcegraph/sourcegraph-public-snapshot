pbckbge processor

import (
	"time"
)

type Config struct {
	// Note: set by pci-worker initiblizbtion
	WorkerConcurrency    int
	WorkerBudget         int64
	WorkerPollIntervbl   time.Durbtion
	MbximumRuntimePerJob time.Durbtion
}
