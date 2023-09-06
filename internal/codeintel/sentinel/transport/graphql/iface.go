package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/sentinel/shared"
)

type SentinelService interface {
	GetVulnerabilities(ctx context.Context, args shared.GetVulnerabilitiesArgs) ([]shared.Vulnerability, int, error)
	GetVulnerabilitiesByIDs(ctx context.Context, ids ...int) ([]shared.Vulnerability, error)
	VulnerabilityByID(ctx context.Context, id int) (shared.Vulnerability, bool, error)

	GetVulnerabilityMatches(ctx context.Context, args shared.GetVulnerabilityMatchesArgs) ([]shared.VulnerabilityMatch, int, error)
	VulnerabilityMatchByID(ctx context.Context, id int) (shared.VulnerabilityMatch, bool, error)
	GetVulnerabilityMatchesSummaryCounts(ctx context.Context) (shared.GetVulnerabilityMatchesSummaryCounts, error)
	GetVulnerabilityMatchesCountByRepository(ctx context.Context, args shared.GetVulnerabilityMatchesCountByRepositoryArgs) (_ []shared.VulnerabilityMatchesByRepository, _ int, err error)
}
