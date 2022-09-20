package graphql

import (
	"context"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	uploads "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Resolver interface {
	// Uploads
	GetUploadsByIDs(ctx context.Context, ids ...int) (_ []types.Upload, err error)
	GetUploadDocumentsForPath(ctx context.Context, bundleID int, pathPattern string) ([]string, int, error)
	GetCommitsVisibleToUpload(ctx context.Context, uploadID, limit int, token *string) (_ []string, nextToken *string, err error)
	GetRecentUploadsSummary(ctx context.Context, repositoryID int) (upload []shared.UploadsWithRepositoryNamespace, err error)
	GetLastUploadRetentionScanForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error)
	DeleteUploadByID(ctx context.Context, id int) (_ bool, err error)

	// Audit Logs
	GetAuditLogsForUpload(ctx context.Context, uploadID int) (_ []types.UploadLog, err error)
}
type resolver struct {
	svc            *uploads.Service
	autoindexer    AutoIndexingService
	policyResolver PolicyResolver
	operations     *operations
}

func New(svc *uploads.Service, autoindexer AutoIndexingService, policyResolver PolicyResolver, observationContext *observation.Context) Resolver {
	return &resolver{
		svc:            svc,
		autoindexer:    autoindexer,
		policyResolver: policyResolver,
		operations:     newOperations(observationContext),
	}
}

func (r *resolver) GetUploadsByIDs(ctx context.Context, ids ...int) (_ []types.Upload, err error) {
	ctx, _, endObservation := r.operations.getIndexByID.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.String("ids", fmt.Sprintf("%v", ids))},
	})
	defer endObservation(1, observation.Args{})

	return r.svc.GetUploadsByIDs(ctx, ids...)
}

func (r *resolver) GetUploadDocumentsForPath(ctx context.Context, bundleID int, pathPattern string) (_ []string, _ int, err error) {
	ctx, _, endObservation := r.operations.getUploadDocumentsForPath.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("bundleID", bundleID), log.String("pathPattern", pathPattern)},
	})
	defer endObservation(1, observation.Args{})

	return r.svc.GetUploadDocumentsForPath(ctx, bundleID, pathPattern)
}

func (r *resolver) GetCommitsVisibleToUpload(ctx context.Context, uploadID, limit int, token *string) (_ []string, _ *string, err error) {
	ctx, _, endObservation := r.operations.getCommitsVisibleToUpload.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("uploadID", uploadID), log.Int("limit", limit), log.String("token", fmt.Sprintf("%v", token))},
	})
	defer endObservation(1, observation.Args{})

	return r.svc.GetCommitsVisibleToUpload(ctx, uploadID, limit, token)
}

func (r *resolver) GetRecentUploadsSummary(ctx context.Context, repositoryID int) (upload []shared.UploadsWithRepositoryNamespace, err error) {
	ctx, _, endObservation := r.operations.getRecentUploadsSummary.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("repositoryID", repositoryID)},
	})
	defer endObservation(1, observation.Args{})

	return r.svc.GetRecentUploadsSummary(ctx, repositoryID)
}

func (r *resolver) GetLastUploadRetentionScanForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error) {
	ctx, _, endObservation := r.operations.getLastUploadRetentionScanForRepository.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("repositoryID", repositoryID)},
	})
	defer endObservation(1, observation.Args{})

	return r.svc.GetLastUploadRetentionScanForRepository(ctx, repositoryID)
}

func (r *resolver) DeleteUploadByID(ctx context.Context, id int) (_ bool, err error) {
	ctx, _, endObservation := r.operations.getIndexByID.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("id", id)},
	})
	defer endObservation(1, observation.Args{})

	return r.svc.DeleteUploadByID(ctx, id)
}

func (r *resolver) GetAuditLogsForUpload(ctx context.Context, uploadID int) (_ []types.UploadLog, err error) {
	ctx, _, endObservation := r.operations.getAuditLogsForUpload.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("uploadID", uploadID)},
	})
	defer endObservation(1, observation.Args{})

	return r.svc.GetAuditLogsForUpload(ctx, uploadID)
}
