package sharedresolvers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type preciseIndexResolver struct {
	uploadsSvc       UploadsService
	policySvc        PolicyService
	gitserverClient  gitserver.Client
	siteAdminChecker SiteAdminChecker
	repoStore        database.RepoStore
	locationResolver *CachedLocationResolver
	traceErrs        *observation.ErrCollector
	upload           *types.Upload
	index            *types.Index
}

func NewPreciseIndexResolver(
	ctx context.Context,
	uploadsSvc UploadsService,
	policySvc PolicyService,
	gitserverClient gitserver.Client,
	prefetcher *Prefetcher,
	siteAdminChecker SiteAdminChecker,
	repoStore database.RepoStore,
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

	if upload != nil {
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

	return &preciseIndexResolver{
		uploadsSvc:       uploadsSvc,
		policySvc:        policySvc,
		gitserverClient:  gitserverClient,
		siteAdminChecker: siteAdminChecker,
		repoStore:        repoStore,
		locationResolver: locationResolver,
		traceErrs:        traceErrs,
		upload:           upload,
		index:            index,
	}, nil
}

//
//
//

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

//
//
//
//

func (r *preciseIndexResolver) IsLatestForRepo() bool {
	return r.upload != nil && r.upload.VisibleAtTip
}

func (r *preciseIndexResolver) QueuedAt() *gqlutil.DateTime {
	if r.index != nil {
		return gqlutil.DateTimeOrNil(&r.index.QueuedAt)
	}

	return nil
}

func (r *preciseIndexResolver) UploadedAt() *gqlutil.DateTime {
	if r.upload != nil {
		return gqlutil.DateTimeOrNil(&r.upload.UploadedAt)
	}

	return nil
}

func (r *preciseIndexResolver) IndexingStartedAt() *gqlutil.DateTime {
	if r.index != nil {
		return gqlutil.DateTimeOrNil(r.index.StartedAt)
	}

	return nil
}

func (r *preciseIndexResolver) ProcessingStartedAt() *gqlutil.DateTime {
	if r.upload != nil {
		return gqlutil.DateTimeOrNil(r.upload.StartedAt)
	}

	return nil
}

func (r *preciseIndexResolver) IndexingFinishedAt() *gqlutil.DateTime {
	if r.index != nil {
		return gqlutil.DateTimeOrNil(r.index.FinishedAt)
	}

	return nil
}

func (r *preciseIndexResolver) ProcessingFinishedAt() *gqlutil.DateTime {
	if r.upload != nil {
		return gqlutil.DateTimeOrNil(r.upload.FinishedAt)
	}

	return nil
}

func (r *preciseIndexResolver) Steps() resolverstubs.IndexStepsResolver {
	if r.index != nil {
		return NewIndexStepsResolver(r.siteAdminChecker, *r.index)
	}

	return nil
}

//
//
//
//

func (r *preciseIndexResolver) InputCommit() string {
	if r.upload != nil {
		return r.upload.Commit
	} else if r.index != nil {
		return r.index.Commit
	}

	return ""
}

func (r *preciseIndexResolver) InputRoot() string {
	if r.upload != nil {
		return r.upload.Root
	} else if r.index != nil {
		return r.index.Root
	}

	return ""
}

func (r *preciseIndexResolver) InputIndexer() string {
	if r.upload != nil {
		return r.upload.Indexer
	} else if r.index != nil {
		return r.index.Indexer
	}

	return ""
}

func (r *preciseIndexResolver) Failure() *string {
	if r.upload != nil && r.upload.FailureMessage != nil {
		return r.upload.FailureMessage
	} else if r.index != nil && r.index.FailureMessage != nil {
		return r.index.FailureMessage
	}

	return nil
}

func (r *preciseIndexResolver) PlaceInQueue() *int32 {
	if r.index != nil && r.index.Rank != nil {
		return toInt32(r.index.Rank)
	} else if r.upload != nil && r.upload.Rank != nil {
		return toInt32(r.upload.Rank)
	}

	return nil
}

func (r *preciseIndexResolver) Indexer() resolverstubs.CodeIntelIndexerResolver {
	if r.index != nil {
		// Note: check index as index fields may contain docker shas
		return types.NewCodeIntelIndexerResolver(r.index.Indexer, r.index.Indexer)
	} else if r.upload != nil {
		return types.NewCodeIntelIndexerResolver(r.upload.Indexer, "")
	}

	return nil
}

func (r *preciseIndexResolver) ShouldReindex(ctx context.Context) bool {
	if r.upload != nil {
		// non-nil upload - this and any index record must both be marked
		return r.upload.ShouldReindex && (r.index == nil || r.index.ShouldReindex)
	}

	// nil upload - an index record must be marked
	return r.index != nil && r.index.ShouldReindex
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

//
//
//
//

func (r *preciseIndexResolver) ProjectRoot(ctx context.Context) (_ resolverstubs.GitTreeEntryResolver, err error) {
	// defer r.traceErrs.Collect(&err, log.String("uploadResolver.field", "projectRoot"))

	var (
		repoID api.RepoID
		commit string
		root   string
	)
	if r.upload != nil {
		repoID, commit, root = api.RepoID(r.upload.RepositoryID), r.upload.Commit, r.upload.Root
	} else if r.index != nil {
		repoID, commit, root = api.RepoID(r.index.RepositoryID), r.index.Commit, r.index.Root
	}

	resolver, err := r.locationResolver.Path(ctx, repoID, commit, root, true)
	if err != nil || resolver == nil {
		// Do not return typed nil interface
		return nil, err
	}

	return resolver, nil
}

func (r *preciseIndexResolver) Tags(ctx context.Context) ([]string, error) {
	var (
		repoName api.RepoName
		commit   string
	)
	if r.upload != nil {
		repoName, commit = api.RepoName(r.upload.RepositoryName), r.upload.Commit
	} else if r.index != nil {
		repoName, commit = api.RepoName(r.index.RepositoryName), r.index.Commit
	}

	tags, err := r.gitserverClient.ListTags(ctx, repoName, commit)
	if err != nil {
		if gitdomain.IsRepoNotExist(err) {
			return nil, nil
		}

		return nil, errors.New("unable to return list of tags in the repository.")
	}

	tagNames := make([]string, 0, len(tags))
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}

	return tagNames, nil
}

var DefaultRetentionPolicyMatchesPageSize = 50

func (r *preciseIndexResolver) RetentionPolicyOverview(ctx context.Context, args *resolverstubs.LSIFUploadRetentionPolicyMatchesArgs) (resolverstubs.CodeIntelligenceRetentionPolicyMatchesConnectionResolver, error) {
	if r.upload == nil {
		return nil, nil
	}

	var afterID int64
	if args.After != nil {
		var err error
		afterID, err = resolverstubs.UnmarshalID[int64](graphql.ID(*args.After))
		if err != nil {
			return nil, err
		}
	}

	pageSize := DefaultRetentionPolicyMatchesPageSize
	if args.First != nil {
		pageSize = int(*args.First)
	}

	var term string
	if args.Query != nil {
		term = *args.Query
	}

	matches, totalCount, err := r.policySvc.GetRetentionPolicyOverview(ctx, *r.upload, args.MatchesOnly, pageSize, afterID, term, time.Now())
	if err != nil {
		return nil, err
	}

	resolvers := make([]resolverstubs.CodeIntelligenceRetentionPolicyMatchResolver, 0, len(matches))
	for _, policy := range matches {
		resolvers = append(resolvers, NewRetentionPolicyMatcherResolver(r.repoStore, policy))
	}

	return resolverstubs.NewTotalCountConnectionResolver(resolvers, 0, int32(totalCount)), nil
}

func (r *preciseIndexResolver) AuditLogs(ctx context.Context) (*[]resolverstubs.LSIFUploadsAuditLogsResolver, error) {
	if r.upload == nil {
		return nil, nil
	}

	logs, err := r.uploadsSvc.GetAuditLogsForUpload(ctx, r.upload.ID)
	if err != nil {
		return nil, err
	}

	resolvers := make([]resolverstubs.LSIFUploadsAuditLogsResolver, 0, len(logs))
	for _, uploadLog := range logs {
		resolvers = append(resolvers, NewLSIFUploadsAuditLogsResolver(uploadLog))
	}

	return &resolvers, nil
}
