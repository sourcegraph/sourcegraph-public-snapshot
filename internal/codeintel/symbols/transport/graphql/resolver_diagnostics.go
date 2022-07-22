package graphql

import (
	"context"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const slowDiagnosticsRequestThreshold = time.Second

// Diagnostics returns the diagnostics for documents with the given path prefix.
func (r *resolver) Diagnostics(ctx context.Context, args shared.RequestArgs) (diagnosticsAtUploads []shared.DiagnosticAtUpload, _ int, err error) {
	ctx, trace, endObservation := observeResolver(ctx, &err, r.operations.diagnostics, slowDiagnosticsRequestThreshold, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", args.RepositoryID),
			log.String("commit", args.Commit),
			log.String("path", args.Path),
			log.Int("numUploads", len(r.dataLoader.uploads)),
			log.String("uploads", uploadIDsToString(r.dataLoader.uploads)),
			log.Int("limit", args.Limit),
		},
	})
	defer endObservation()

	visibleUploads, err := r.getUploadPaths(ctx, args.Path)
	if err != nil {
		return nil, 0, err
	}

	totalCount := 0

	checkerEnabled := authz.SubRepoEnabled(r.authChecker)
	var a *actor.Actor
	if checkerEnabled {
		a = actor.FromContext(ctx)
	}
	for i := range visibleUploads {
		trace.Log(log.Int("uploadID", visibleUploads[i].Upload.ID))

		diagnostics, count, err := r.svc.GetDiagnostics(
			ctx,
			visibleUploads[i].Upload.ID,
			visibleUploads[i].TargetPathWithoutRoot,
			args.Limit-len(diagnosticsAtUploads),
			0,
		)
		if err != nil {
			return nil, 0, errors.Wrap(err, "lsifStore.Diagnostics")
		}

		for _, diagnostic := range diagnostics {
			adjustedDiagnostic, err := r.getRequestedCommitDiagnostic(ctx, visibleUploads[i], diagnostic)
			if err != nil {
				return nil, 0, err
			}

			if !checkerEnabled {
				diagnosticsAtUploads = append(diagnosticsAtUploads, adjustedDiagnostic)
				continue
			}

			// sub-repo checker is enabled, proceeding with check
			if include, err := authz.FilterActorPath(ctx, r.authChecker, a, api.RepoName(adjustedDiagnostic.Dump.RepositoryName), adjustedDiagnostic.Path); err != nil {
				return nil, 0, err
			} else if include {
				diagnosticsAtUploads = append(diagnosticsAtUploads, adjustedDiagnostic)
			}
		}

		totalCount += count
	}

	if len(diagnosticsAtUploads) > args.Limit {
		diagnosticsAtUploads = diagnosticsAtUploads[:args.Limit]
	}
	trace.Log(
		log.Int("totalCount", totalCount),
		log.Int("numDiagnostics", len(diagnosticsAtUploads)),
	)

	return diagnosticsAtUploads, totalCount, nil
}

// getUploadPaths adjusts the current target path for each upload visible from the current target
// commit. If an upload cannot be adjusted, it will be omitted from the returned slice.
func (r *resolver) getUploadPaths(ctx context.Context, path string) ([]visibleUpload, error) {
	visibleUploads := make([]visibleUpload, 0, len(r.dataLoader.uploads))
	for i := range r.dataLoader.uploads {
		targetPath, ok, err := r.GitTreeTranslator.GetTargetCommitPathFromSourcePath(ctx, r.dataLoader.uploads[i].Commit, path, false) //.AdjustPath(ctx, r.inMemoryUploads[i].Commit, r.path, false)
		if err != nil {
			return nil, errors.Wrap(err, "positionAdjuster.AdjustPath")
		}
		if !ok {
			continue
		}

		visibleUploads = append(visibleUploads, visibleUpload{
			Upload:                r.dataLoader.uploads[i],
			TargetPath:            targetPath,
			TargetPathWithoutRoot: strings.TrimPrefix(targetPath, r.dataLoader.uploads[i].Root),
		})
	}

	return visibleUploads, nil
}

// adjustDiagnostic translates a diagnostic (relative to the indexed commit) into an equivalent diagnostic
// in the requested commit.
func (r *resolver) getRequestedCommitDiagnostic(ctx context.Context, adjustedUpload visibleUpload, diagnostic shared.Diagnostic) (shared.DiagnosticAtUpload, error) {
	rn := shared.Range{
		Start: shared.Position{
			Line:      diagnostic.StartLine,
			Character: diagnostic.StartCharacter,
		},
		End: shared.Position{
			Line:      diagnostic.EndLine,
			Character: diagnostic.EndCharacter,
		},
	}

	// Adjust path in diagnostic before reading it. This value is used in the adjustRange
	// call below, and is also reflected in the embedded diagnostic value in the return.
	diagnostic.Path = adjustedUpload.Upload.Root + diagnostic.Path

	adjustedCommit, adjustedRange, _, err := r.getSourceRange(
		ctx,
		adjustedUpload.Upload.RepositoryID,
		adjustedUpload.Upload.Commit,
		diagnostic.Path,
		rn,
	)
	if err != nil {
		return shared.DiagnosticAtUpload{}, err
	}

	return shared.DiagnosticAtUpload{
		Diagnostic:     diagnostic,
		Dump:           adjustedUpload.Upload,
		AdjustedCommit: adjustedCommit,
		AdjustedRange:  adjustedRange,
	}, nil
}
