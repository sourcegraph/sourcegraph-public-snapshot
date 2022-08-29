package graphql

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go/log"

	uploads "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Resolver interface {
	// Uploads
	GetUploadsByIDs(ctx context.Context, ids ...int) (_ []shared.Upload, err error)
	GetUploadDocumentsForPath(ctx context.Context, bundleID int, pathPattern string) ([]string, int, error)
	GetCommitsVisibleToUpload(ctx context.Context, uploadID, limit int, token *string) (_ []string, nextToken *string, err error)
	DeleteUploadByID(ctx context.Context, id int) (_ bool, err error)

	// Uploads Connection Factory
	UploadsConnectionResolverFromFactory(opts shared.GetUploadsOptions) *UploadsResolver
}
type resolver struct {
	svc        *uploads.Service
	operations *operations
}

func New(svc *uploads.Service, observationContext *observation.Context) Resolver {
	return &resolver{
		svc:        svc,
		operations: newOperations(observationContext),
	}
}

func (r *resolver) GetUploadsByIDs(ctx context.Context, ids ...int) (_ []shared.Upload, err error) {
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

func (r *resolver) DeleteUploadByID(ctx context.Context, id int) (_ bool, err error) {
	ctx, _, endObservation := r.operations.getIndexByID.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("id", id)},
	})
	defer endObservation(1, observation.Args{})

	return r.svc.DeleteUploadByID(ctx, id)
}

// func (r *resolver) GetCommitGraphMetadata(ctx context.Context, repositoryID int) (stale bool, updatedAt *time.Time, err error) {
// 	ctx, _, endObservation := r.operations.getCommitGraphMetadata.With(ctx, &err, observation.Args{
// 		LogFields: []log.Field{log.Int("repositoryID", repositoryID)},
// 	})
// 	defer endObservation(1, observation.Args{})

// 	return r.svc.GetCommitGraphMetadata(ctx, repositoryID)
// }

func (r *resolver) UploadsConnectionResolverFromFactory(opts shared.GetUploadsOptions) *UploadsResolver {
	return NewUploadsResolver(r.svc, opts)
}

func (r *resolver) CommitGraphResolverFromFactory(ctx context.Context, repositoryID int) *CommitGraphResolver {
	stale, updatedAt, err := r.svc.GetCommitGraphMetadata(ctx, repositoryID)
	if err != nil {
		return nil
	}

	return NewCommitGraphResolver(stale, updatedAt)
}
