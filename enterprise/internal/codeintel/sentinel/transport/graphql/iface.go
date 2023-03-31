package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel"
)

type SentinelService interface {
	GetVulnerabilities(ctx context.Context, args sentinel.GetVulnerabilitiesArgs) ([]sentinel.Vulnerability, int, error)
	GetVulnerabilitiesByIDs(ctx context.Context, ids ...int) ([]sentinel.Vulnerability, error)
	VulnerabilityByID(ctx context.Context, id int) (sentinel.Vulnerability, bool, error)

	GetVulnerabilityMatches(ctx context.Context, args sentinel.GetVulnerabilityMatchesArgs) ([]sentinel.VulnerabilityMatch, int, error)
	VulnerabilityMatchByID(ctx context.Context, id int) (sentinel.VulnerabilityMatch, bool, error)
	GetVulnerabilityMatchesSummaryCounts(ctx context.Context) (sentinel.GetVulnerabilityMatchesSummaryCounts, error)
	GetVulnerabilityMatchesCountByRepository(ctx context.Context, args sentinel.GetVulnerabilityMatchesCountByRepositoryArgs) (_ []sentinel.VulnerabilityMatchesByRepository, _ int, err error)
}
