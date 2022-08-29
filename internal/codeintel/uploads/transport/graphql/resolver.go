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

func (r *resolver) DeleteUploadByID(ctx context.Context, id int) (_ bool, err error) {
	ctx, _, endObservation := r.operations.getIndexByID.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("id", id)},
	})
	defer endObservation(1, observation.Args{})

	return r.svc.DeleteUploadByID(ctx, id)
}

func (r *resolver) UploadsConnectionResolverFromFactory(opts shared.GetUploadsOptions) *UploadsResolver {
	return NewUploadsResolver(r.svc, opts)
}
