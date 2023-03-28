package graphql

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// DefaultDiagnosticsPageSize is the diagnostic result page size when no limit is supplied.
const DefaultDiagnosticsPageSize = 100

// Diagnostics returns the diagnostics for documents with the given path prefix.
func (r *gitBlobLSIFDataResolver) Diagnostics(ctx context.Context, args *resolverstubs.LSIFDiagnosticsArgs) (_ resolverstubs.DiagnosticConnectionResolver, err error) {
	limit := int(resolverstubs.Deref(args.First, DefaultDiagnosticsPageSize))
	if limit <= 0 {
		return nil, ErrIllegalLimit
	}

	requestArgs := shared.RequestArgs{RepositoryID: r.requestState.RepositoryID, Commit: r.requestState.Commit, Path: r.requestState.Path, Limit: limit}
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.diagnostics, time.Second, getObservationArgs(requestArgs))
	defer endObservation()

	diagnostics, totalCount, err := r.codeNavSvc.GetDiagnostics(ctx, requestArgs, r.requestState)
	if err != nil {
		return nil, errors.Wrap(err, "codeNavSvc.GetDiagnostics")
	}

	resolvers := make([]resolverstubs.DiagnosticResolver, 0, len(diagnostics))
	for i := range diagnostics {
		resolvers = append(resolvers, newDiagnosticResolver(diagnostics[i], r.locationResolver))
	}

	return resolverstubs.NewTotalCountConnectionResolver(resolvers, 0, int32(totalCount)), nil
}

//
//

type diagnosticResolver struct {
	diagnostic       shared.DiagnosticAtUpload
	locationResolver *sharedresolvers.CachedLocationResolver
}

func newDiagnosticResolver(diagnostic shared.DiagnosticAtUpload, locationResolver *sharedresolvers.CachedLocationResolver) resolverstubs.DiagnosticResolver {
	return &diagnosticResolver{
		diagnostic:       diagnostic,
		locationResolver: locationResolver,
	}
}

func (r *diagnosticResolver) Severity() (*string, error) { return toSeverity(r.diagnostic.Severity) }
func (r *diagnosticResolver) Code() (*string, error) {
	return resolverstubs.NonZeroPtr(r.diagnostic.Code), nil
}
func (r *diagnosticResolver) Source() (*string, error) {
	return resolverstubs.NonZeroPtr(r.diagnostic.Source), nil
}
func (r *diagnosticResolver) Message() (*string, error) {
	return resolverstubs.NonZeroPtr(r.diagnostic.Message), nil
}

func (r *diagnosticResolver) Location(ctx context.Context) (resolverstubs.LocationResolver, error) {
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
