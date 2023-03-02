package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/shared"
)

type Service interface {
	VulnerabilityByID(ctx context.Context, id int) (shared.Vulnerability, bool, error)
	VulnerabilityMatchByID(ctx context.Context, id int) (shared.VulnerabilityMatch, bool, error)
}
