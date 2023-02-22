package sharedresolvers

import (
	"context"
	"fmt"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type preciseIndexResolver struct {
	upload         *types.Upload
	index          *types.Index
	uploadResolver resolverstubs.LSIFUploadResolver
	indexResolver  resolverstubs.LSIFIndexResolver
}

func NewPreciseIndexResolver(
	ctx context.Context,
	autoindexingSvc AutoIndexingService,
	uploadsSvc UploadsService,
	policySvc PolicyService,
	prefetcher *Prefetcher,
	locationResolver *CachedLocationResolver,
	traceErrs *observation.ErrCollector,
	upload *types.Upload,
	index *types.Index,
) (resolverstubs.PreciseIndexResolver, error) {
	if index != nil && index.AssociatedUploadID != nil && upload == nil {
		v, ok, err := prefetcher.GetUploadByID(ctx, *index.AssociatedUploadID)
		if err != nil {
			return nil, err
		}
		if ok {
			upload = &v
		}
	}

	var uploadResolver resolverstubs.LSIFUploadResolver
	if upload != nil {
		uploadResolver = NewUploadResolver(uploadsSvc, autoindexingSvc, policySvc, *upload, prefetcher, locationResolver, traceErrs)

		if upload.AssociatedIndexID != nil {
			v, ok, err := prefetcher.GetIndexByID(ctx, *upload.AssociatedIndexID)
			if err != nil {
				return nil, err
			}
			if ok {
				index = &v
			}
		}
	}

	var indexResolver resolverstubs.LSIFIndexResolver
	if index != nil {
		indexResolver = NewIndexResolver(autoindexingSvc, uploadsSvc, policySvc, *index, prefetcher, locationResolver, traceErrs)
	}

	return &preciseIndexResolver{
		upload:         upload,
		index:          index,
		uploadResolver: uploadResolver,
		indexResolver:  indexResolver,
	}, nil
}

func (r *preciseIndexResolver) ID() graphql.ID {
	var parts []string
	if r.upload != nil {
		parts = append(parts, fmt.Sprintf("U:%d", r.upload.ID))
	}
	if r.index != nil {
		parts = append(parts, fmt.Sprintf("I:%d", r.index.ID))
	}

	return relay.MarshalID("PreciseIndex", strings.Join(parts, ":"))
}

func (r *preciseIndexResolver) ProjectRoot(ctx context.Context) (resolverstubs.GitTreeEntryResolver, error) {
	if r.uploadResolver != nil {
		return r.uploadResolver.ProjectRoot(ctx)
	}

	return r.indexResolver.ProjectRoot(ctx)
}

func (r *preciseIndexResolver) InputCommit() string {
	if r.uploadResolver != nil {
		return r.uploadResolver.InputCommit()
	}

	return r.indexResolver.InputCommit()
}

func (r *preciseIndexResolver) Tags(ctx context.Context) ([]string, error) {
	if r.uploadResolver != nil {
		return r.uploadResolver.Tags(ctx)
	}

	return r.indexResolver.Tags(ctx)
}

func (r *preciseIndexResolver) InputRoot() string {
	if r.uploadResolver != nil {
		return r.uploadResolver.InputRoot()
	}

	return r.indexResolver.InputRoot()
}

func (r *preciseIndexResolver) InputIndexer() string {
	if r.uploadResolver != nil {
		return r.uploadResolver.InputIndexer()
	}

	return r.indexResolver.InputIndexer()
}

func (r *preciseIndexResolver) Indexer() resolverstubs.CodeIntelIndexerResolver {
	if r.uploadResolver != nil {
		return r.uploadResolver.Indexer()
	}

	return r.indexResolver.Indexer()
}

func (r *preciseIndexResolver) State() string {
	if r.upload != nil {
		switch strings.ToUpper(r.upload.State) {
		case "UPLOADING":
			return "UPLOADING_INDEX"

		case "QUEUED":
			return "QUEUED_FOR_PROCESSING"

		case "PROCESSING":
			return "PROCESSING"

		case "FAILED":
			fallthrough
		case "ERRORED":
			return "PROCESSING_ERRORED"

		case "COMPLETED":
			return "COMPLETED"

		case "DELETING":
			return "DELETING"

		case "DELETED":
			return "DELETED"

		default:
			panic(fmt.Sprintf("unrecognized upload state %q", r.upload.State))
		}
	}

	switch strings.ToUpper(r.index.State) {
	case "QUEUED":
		return "QUEUED_FOR_INDEXING"

	case "PROCESSING":
		return "INDEXING"

	case "FAILED":
		fallthrough
	case "ERRORED":
		return "INDEXING_ERRORED"

	case "COMPLETED":
		// Should not actually occur in practice (where did upload go?)
		return "INDEXING_COMPLETED"

	default:
		panic(fmt.Sprintf("unrecognized index state %q", r.index.State))
	}
}

func (r *preciseIndexResolver) QueuedAt() *gqlutil.DateTime {
	if r.indexResolver != nil {
		t := r.indexResolver.QueuedAt()
		return &t
	}

	return nil
}

func (r *preciseIndexResolver) UploadedAt() *gqlutil.DateTime {
	if r.uploadResolver != nil {
		t := r.uploadResolver.UploadedAt()
		return &t
	}

	return nil
}

func (r *preciseIndexResolver) IndexingStartedAt() *gqlutil.DateTime {
	if r.indexResolver != nil {
		return r.indexResolver.StartedAt()
	}

	return nil
}

func (r *preciseIndexResolver) ProcessingStartedAt() *gqlutil.DateTime {
	if r.uploadResolver != nil {
		return r.uploadResolver.StartedAt()
	}

	return nil
}

func (r *preciseIndexResolver) IndexingFinishedAt() *gqlutil.DateTime {
	if r.indexResolver != nil {
		return r.indexResolver.FinishedAt()
	}

	return nil
}

func (r *preciseIndexResolver) ProcessingFinishedAt() *gqlutil.DateTime {
	if r.uploadResolver != nil {
		return r.uploadResolver.FinishedAt()
	}

	return nil
}

func (r *preciseIndexResolver) Steps() resolverstubs.IndexStepsResolver {
	if r.indexResolver == nil {
		return nil
	}

	return r.indexResolver.Steps()
}

func (r *preciseIndexResolver) Failure() *string {
	if r.upload != nil && r.upload.FailureMessage != nil {
		return r.upload.FailureMessage
	} else if r.index != nil {
		return r.index.FailureMessage
	}

	return nil
}

func (r *preciseIndexResolver) PlaceInQueue() *int32 {
	if r.index != nil && r.index.Rank != nil {
		return toInt32(r.index.Rank)
	}

	if r.upload != nil && r.upload.Rank != nil {
		return toInt32(r.upload.Rank)
	}

	return nil
}

func (r *preciseIndexResolver) ShouldReindex(ctx context.Context) bool {
	if r.index != nil {
		if r.upload != nil {
			return r.upload.ShouldReindex && r.index.ShouldReindex
		}

		return r.index.ShouldReindex
	}
	if r.upload != nil {
		return r.upload.ShouldReindex
	}
	return false
}

func (r *preciseIndexResolver) IsLatestForRepo() bool {
	if r.upload == nil {
		return false
	}

	return r.upload.VisibleAtTip
}

func (r *preciseIndexResolver) RetentionPolicyOverview(ctx context.Context, args *resolverstubs.LSIFUploadRetentionPolicyMatchesArgs) (resolverstubs.CodeIntelligenceRetentionPolicyMatchesConnectionResolver, error) {
	if r.uploadResolver == nil {
		return nil, nil
	}

	return r.uploadResolver.RetentionPolicyOverview(ctx, args)
}

func (r *preciseIndexResolver) AuditLogs(ctx context.Context) (*[]resolverstubs.LSIFUploadsAuditLogsResolver, error) {
	if r.uploadResolver == nil {
		return nil, nil
	}

	return r.uploadResolver.AuditLogs(ctx)
}
