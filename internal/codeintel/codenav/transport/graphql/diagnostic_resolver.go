package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	sharedresolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type DiagnosticResolver interface {
	Severity() (*string, error)
	Code() (*string, error)
	Source() (*string, error)
	Message() (*string, error)
	Location(ctx context.Context) (LocationResolver, error)
}

type diagnosticResolver struct {
	diagnostic       shared.DiagnosticAtUpload
	locationResolver *sharedresolvers.CachedLocationResolver
}

func NewDiagnosticResolver(diagnostic shared.DiagnosticAtUpload, locationResolver *sharedresolvers.CachedLocationResolver) DiagnosticResolver {
	return &diagnosticResolver{
		diagnostic:       diagnostic,
		locationResolver: locationResolver,
	}
}

func (r *diagnosticResolver) Severity() (*string, error) { return toSeverity(r.diagnostic.Severity) }
func (r *diagnosticResolver) Code() (*string, error)     { return strPtr(r.diagnostic.Code), nil }
func (r *diagnosticResolver) Source() (*string, error)   { return strPtr(r.diagnostic.Source), nil }
func (r *diagnosticResolver) Message() (*string, error)  { return strPtr(r.diagnostic.Message), nil }

func (r *diagnosticResolver) Location(ctx context.Context) (LocationResolver, error) {
	return resolveLocation(
		ctx,
		r.locationResolver,
		types.UploadLocation{
			Dump:         r.diagnostic.Dump,
			Path:         r.diagnostic.Path,
			TargetCommit: r.diagnostic.AdjustedCommit,
			TargetRange:  r.diagnostic.AdjustedRange,
		},
	)
}

var severities = map[int]string{
	1: "ERROR",
	2: "WARNING",
	3: "INFORMATION",
	4: "HINT",
}

func toSeverity(val int) (*string, error) {
	severity, ok := severities[val]
	if !ok {
		return nil, errors.Errorf("unknown diagnostic severity %d", val)
	}

	return &severity, nil
}
