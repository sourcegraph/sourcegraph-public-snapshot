package executorqueue

import (
	"github.com/inconshreveable/log15"

	executortypes "github.com/sourcegraph/sourcegraph/internal/executor/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type QueueAllocation struct {
	PercentageAWS float64
	PercentageGCP float64
}

var validCloudProviderNames = []string{"aws", "gcp"}

func normalizeAllocations(m map[string]map[string]float64, awsConfigured, gcpConfigured bool) (map[string]QueueAllocation, error) {
	for queueName := range m {
		if !contains(executortypes.ValidQueueNames, queueName) {
			return nil, errors.Errorf("invalid queue '%s'", queueName)
		}
	}

	allocations := make(map[string]QueueAllocation, len(executortypes.ValidQueueNames))
	for _, queueName := range executortypes.ValidQueueNames {
		queueAllocation, err := normalizeQueueAllocation(queueName, m[queueName], awsConfigured, gcpConfigured)
		if err != nil {
			return nil, err
		}

		allocations[queueName] = queueAllocation
	}

	return allocations, nil
}

func normalizeQueueAllocation(queueName string, queueAllocation map[string]float64, awsConfigured, gcpConfigured bool) (QueueAllocation, error) {
	if len(queueAllocation) == 0 {
		if awsConfigured {
			if gcpConfigured {
				log15.Warn("Sending 100% of executor queue metrics to BOTH AWS and GCP")
				return QueueAllocation{PercentageAWS: 1, PercentageGCP: 1}, nil
			}

			return QueueAllocation{PercentageAWS: 1}, nil
		} else if gcpConfigured {
			return QueueAllocation{PercentageGCP: 1}, nil
		}

		return QueueAllocation{}, nil
	}

	for cloudProvider, allocation := range queueAllocation {
		if !contains(validCloudProviderNames, cloudProvider) {
			return QueueAllocation{}, errors.Errorf("invalid cloud provider '%s', expected 'aws' or 'gcp'", cloudProvider)
		}

		if allocation < 0 || allocation > 1 {
			return QueueAllocation{}, errors.Errorf("invalid cloud provider allocation '%.2f'", allocation)
		}
	}

	if !awsConfigured && queueAllocation["aws"] > 0 {
		log15.Warn("AWS executor queue metrics not configured - setting allocation to zero", "queueName", queueName)
		queueAllocation["aws"] = 0
	}
	if !gcpConfigured && queueAllocation["gcp"] > 0 {
		log15.Warn("GCP executor queue metrics not configured - setting allocation to zero", "queueName", queueName)
		queueAllocation["gcp"] = 0
	}

	if totalAllocation := queueAllocation["aws"] + queueAllocation["gcp"]; totalAllocation < 1 {
		log15.Warn("Not configured to send full executor queue metrics", "queueName", queueName, "totalAllocation", totalAllocation)
	}

	return QueueAllocation{
		PercentageAWS: queueAllocation["aws"],
		PercentageGCP: queueAllocation["gcp"],
	}, nil
}

func contains(slice []string, value string) bool {
	for _, v := range slice {
		if value == v {
			return true
		}
	}

	return false
}
