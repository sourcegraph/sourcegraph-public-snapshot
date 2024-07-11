package internal

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies"
	policiesshared "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type PolicyType string

const (
	SyntacticIndexing PolicyType = "SYNTACTIC_INDEXING"
	PreciseIndexing   PolicyType = "PRECISE_INDEXING"
)

// This iterator abstracts away the pagination logic for retrieving policies batches,
// propagating any errors
type PolicyIterator interface {
	// Iterate over all matching policies in batches. The `handle` function is NEVER
	// invoked with an empty policies list
	ForEachPoliciesBatch(ctx context.Context, handle func([]policiesshared.ConfigurationPolicy) error) error
}

type policyIterator struct {
	Service      policies.Service
	RepositoryID int
	PolicyType   PolicyType
	BatchSize    int
}

func (p policyIterator) ForEachPoliciesBatch(ctx context.Context, handle func([]policiesshared.ConfigurationPolicy) error) error {
	forSyntacticIndexing := pointers.Ptr(p.PolicyType == SyntacticIndexing)
	forPreciseIndexing := pointers.Ptr(p.PolicyType == PreciseIndexing)

	options := policiesshared.GetConfigurationPoliciesOptions{
		RepositoryID:         p.RepositoryID,
		ForSyntacticIndexing: forSyntacticIndexing,
		ForPreciseIndexing:   forPreciseIndexing,
		Limit:                p.BatchSize,
	}

	for offset := 0; ; {
		options.Offset = 0
		policiesBatch, totalCount, err := p.Service.GetConfigurationPolicies(ctx, options)
		if err != nil {
			return err
		}
		if len(policiesBatch) == 0 {
			break
		}
		if handlerError := handle(policiesBatch); handlerError != nil {
			return handlerError
		}
		if offset = offset + len(policiesBatch); offset >= totalCount {
			break
		}
	}

	return nil
}

var _ PolicyIterator = policyIterator{}

func NewPolicyIterator(service policies.Service, repositoryId int, policyType PolicyType, batchSize int) PolicyIterator {
	return policyIterator{
		Service:      service,
		RepositoryID: repositoryId,
		PolicyType:   policyType,
		BatchSize:    batchSize,
	}
}
