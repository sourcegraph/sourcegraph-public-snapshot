package graphql

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/opentracing/opentracing-go/log"

	autoindexingshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	uploadsshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const DefaultPageSize = 50

func (r *rootResolver) PreciseIndexes(ctx context.Context, args *resolverstubs.PreciseIndexesQueryArgs) (_ resolverstubs.PreciseIndexConnectionResolver, err error) {
	ctx, errTracer, endObservation := r.operations.preciseIndexes.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		// log.String("uploadID", string(id)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	pageSize := DefaultPageSize
	if args.First != nil {
		pageSize = int(*args.First)
	}
	uploadOffset := 0
	indexOffset := 0
	if args.After != nil {
		parts := strings.Split(*args.After, ":")
		if len(parts) != 2 {
			return nil, errors.New("invalid cursor")
		}

		if parts[0] != "" {
			v, err := strconv.Atoi(parts[0])
			if err != nil {
				return nil, errors.New("invalid cursor")
			}

			uploadOffset = v
		}
		if parts[1] != "" {
			v, err := strconv.Atoi(parts[1])
			if err != nil {
				return nil, errors.New("invalid cursor")
			}

			indexOffset = v
		}
	}

	var uploadStates, indexStates []string
	if args.States != nil {
		uploadStates, indexStates, err = bifurcateStates(*args.States)
		if err != nil {
			return nil, err
		}
	}
	skipUploads := len(uploadStates) == 0 && len(indexStates) != 0
	skipIndexes := len(uploadStates) != 0 && len(indexStates) == 0

	var dependencyOf int
	if args.DependencyOf != nil {
		v, v2, err := unmarshalPreciseIndexGQLID(graphql.ID(*args.DependencyOf))
		if err != nil {
			return nil, err
		}
		if v == 0 {
			return nil, errors.Newf("requested dependency of precise index record without data (indexid=%d)", v2)
		}

		dependencyOf = v
		skipIndexes = true
	}
	var dependentOf int
	if args.DependentOf != nil {
		v, v2, err := unmarshalPreciseIndexGQLID(graphql.ID(*args.DependentOf))
		if err != nil {
			return nil, err
		}
		if v == 0 {
			return nil, errors.Newf("requested dependent of precise index record without data (indexid=%d)", v2)
		}

		dependentOf = v
		skipIndexes = true
	}

	var repositoryID int
	if args.Repo != nil {
		v, err := UnmarshalRepositoryID(*args.Repo)
		if err != nil {
			return nil, err
		}

		repositoryID = int(v)
	}

	term := ""
	if args.Query != nil {
		term = *args.Query
	}

	var uploads []types.Upload
	totalUploadCount := 0
	if !skipUploads {
		if uploads, totalUploadCount, err = r.uploadSvc.GetUploads(ctx, uploadsshared.GetUploadsOptions{
			RepositoryID: repositoryID,
			States:       uploadStates,
			Term:         term,
			DependencyOf: dependencyOf,
			DependentOf:  dependentOf,
			Limit:        pageSize,
			Offset:       uploadOffset,
		}); err != nil {
			return nil, err
		}
	}

	var indexes []types.Index
	totalIndexCount := 0
	if !skipIndexes {
		if indexes, totalIndexCount, err = r.autoindexSvc.GetIndexes(ctx, autoindexingshared.GetIndexesOptions{
			RepositoryID:  repositoryID,
			States:        indexStates,
			Term:          term,
			WithoutUpload: true,
			Limit:         pageSize,
			Offset:        indexOffset,
		}); err != nil {
			return nil, err
		}
	}

	type pair struct {
		upload *types.Upload
		index  *types.Index
	}
	pairs := make([]pair, 0, pageSize)
	addUpload := func(upload types.Upload) { pairs = append(pairs, pair{&upload, nil}) }
	addIndex := func(index types.Index) { pairs = append(pairs, pair{nil, &index}) }

	uIdx := 0
	iIdx := 0
	for uIdx < len(uploads) && iIdx < len(indexes) && (uIdx+iIdx) < pageSize {
		if uploads[uIdx].UploadedAt.After(indexes[iIdx].QueuedAt) {
			addUpload(uploads[uIdx])
			uIdx++
		} else {
			addIndex(indexes[iIdx])
			iIdx++
		}
	}
	for uIdx < len(uploads) && (uIdx+iIdx) < pageSize {
		addUpload(uploads[uIdx])
		uIdx++
	}
	for iIdx < len(indexes) && (uIdx+iIdx) < pageSize {
		addIndex(indexes[iIdx])
		iIdx++
	}

	cursor := ""
	if newUploadOffset := uploadOffset + uIdx; newUploadOffset < totalUploadCount {
		cursor += strconv.Itoa(newUploadOffset)
	}
	cursor += ":"
	if newIndexOffset := indexOffset + iIdx; newIndexOffset < totalIndexCount {
		cursor += strconv.Itoa(newIndexOffset)
	}
	if cursor == ":" {
		cursor = ""
	}

	// Create a new prefetcher here as we only want to cache upload and index records in
	// the same graphQL request, not across different request.
	prefetcher := sharedresolvers.NewPrefetcher(r.autoindexSvc, r.uploadSvc)
	db := r.autoindexSvc.GetUnsafeDB()
	locationResolver := sharedresolvers.NewCachedLocationResolver(db, gitserver.NewClient())

	for _, pair := range pairs {
		if pair.upload != nil && pair.upload.AssociatedIndexID != nil {
			prefetcher.MarkIndex(*pair.upload.AssociatedIndexID)
		}
	}

	resolvers := make([]resolverstubs.PreciseIndexResolver, 0, len(pairs))
	for _, pair := range pairs {
		resolver, err := NewPreciseIndexResolver(ctx, r.autoindexSvc, r.uploadSvc, r.policySvc, prefetcher, locationResolver, errTracer, pair.upload, pair.index)
		if err != nil {
			return nil, err
		}

		resolvers = append(resolvers, resolver)
	}

	return NewPreciseIndexConnectionResolver(resolvers, totalUploadCount+totalIndexCount, cursor), nil
}

func (r *rootResolver) PreciseIndexByID(ctx context.Context, id graphql.ID) (_ resolverstubs.PreciseIndexResolver, err error) {
	ctx, errTracer, endObservation := r.operations.preciseIndexes.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("id", string(id)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	uploadID, indexID, err := unmarshalPreciseIndexGQLID(id)
	if err != nil {
		return nil, err
	}

	// Create a new prefetcher here as we only want to cache upload and index records in
	// the same graphQL request, not across different request.
	prefetcher := sharedresolvers.NewPrefetcher(r.autoindexSvc, r.uploadSvc)
	db := r.autoindexSvc.GetUnsafeDB()
	locationResolver := sharedresolvers.NewCachedLocationResolver(db, gitserver.NewClient())

	if uploadID != 0 {
		upload, ok, err := r.uploadSvc.GetUploadByID(ctx, uploadID)
		if err != nil || !ok {
			return nil, err
		}

		return NewPreciseIndexResolver(ctx, r.autoindexSvc, r.uploadSvc, r.policySvc, prefetcher, locationResolver, errTracer, &upload, nil)
	}
	if indexID != 0 {
		index, ok, err := r.autoindexSvc.GetIndexByID(ctx, indexID)
		if err != nil || !ok {
			return nil, err
		}

		return NewPreciseIndexResolver(ctx, r.autoindexSvc, r.uploadSvc, r.policySvc, prefetcher, locationResolver, errTracer, nil, &index)
	}

	return nil, errors.New("invalid identifier")
}

// 🚨 SECURITY: Only site admins may modify code intelligence upload data
func (r *rootResolver) DeletePreciseIndex(ctx context.Context, args *struct{ ID graphql.ID }) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.deletePreciseIndex.With(ctx, &err, observation.Args{})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.autoindexSvc.GetUnsafeDB()); err != nil {
		return nil, err
	}

	uploadID, indexID, err := unmarshalPreciseIndexGQLID(args.ID)
	if err != nil {
		return nil, err
	}
	if uploadID != 0 {
		if _, err := r.uploadSvc.DeleteUploadByID(ctx, uploadID); err != nil {
			return nil, err
		}
	} else if indexID != 0 {
		if _, err := r.autoindexSvc.DeleteIndexByID(ctx, indexID); err != nil {
			return nil, err
		}
	}

	return &resolverstubs.EmptyResponse{}, nil
}

// 🚨 SECURITY: Only site admins may modify code intelligence upload data
func (r *rootResolver) DeletePreciseIndexes(ctx context.Context, args *resolverstubs.DeletePreciseIndexesArgs) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.deletePreciseIndexes.With(ctx, &err, observation.Args{})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.autoindexSvc.GetUnsafeDB()); err != nil {
		return nil, err
	}

	var uploadStates, indexStates []string
	if args.States != nil {
		uploadStates, indexStates, err = bifurcateStates(*args.States)
		if err != nil {
			return nil, err
		}
	}
	skipUploads := len(uploadStates) == 0 && len(indexStates) != 0
	skipIndexes := len(uploadStates) != 0 && len(indexStates) == 0

	repositoryID := 0
	if args.Repository != nil {
		repositoryID, err = resolveRepositoryID(*args.Repository)
		if err != nil {
			return nil, err
		}
	}
	term := derefString(args.Query, "")

	visibleAtTip := false
	if args.IsLatestForRepo != nil {
		visibleAtTip = *args.IsLatestForRepo
		skipIndexes = true
	}

	if !skipUploads {
		if err := r.uploadSvc.DeleteUploads(ctx, uploadsshared.DeleteUploadsOptions{
			RepositoryID: repositoryID,
			States:       uploadStates,
			Term:         term,
			VisibleAtTip: visibleAtTip,
		}); err != nil {
			return nil, err
		}
	}
	if !skipIndexes {
		if err := r.autoindexSvc.DeleteIndexes(ctx, autoindexingshared.DeleteIndexesOptions{
			RepositoryID:  repositoryID,
			States:        indexStates,
			Term:          term,
			WithoutUpload: true,
		}); err != nil {
			return nil, err
		}
	}

	return &resolverstubs.EmptyResponse{}, nil
}

// 🚨 SECURITY: Only site admins may modify code intelligence upload data
func (r *rootResolver) ReindexPreciseIndex(ctx context.Context, args *struct{ ID graphql.ID }) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.reindexPreciseIndex.With(ctx, &err, observation.Args{})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.autoindexSvc.GetUnsafeDB()); err != nil {
		return nil, err
	}

	uploadID, indexID, err := unmarshalPreciseIndexGQLID(args.ID)
	if err != nil {
		return nil, err
	}
	if uploadID != 0 {
		if err := r.uploadSvc.ReindexUploadByID(ctx, uploadID); err != nil {
			return nil, err
		}
	} else if indexID != 0 {
		if err := r.autoindexSvc.ReindexIndexByID(ctx, indexID); err != nil {
			return nil, err
		}
	}

	return &resolverstubs.EmptyResponse{}, nil
}

// 🚨 SECURITY: Only site admins may modify code intelligence upload data
func (r *rootResolver) ReindexPreciseIndexes(ctx context.Context, args *resolverstubs.ReindexPreciseIndexesArgs) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.reindexPreciseIndexes.With(ctx, &err, observation.Args{})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.autoindexSvc.GetUnsafeDB()); err != nil {
		return nil, err
	}

	var uploadStates, indexStates []string
	if args.States != nil {
		uploadStates, indexStates, err = bifurcateStates(*args.States)
		if err != nil {
			return nil, err
		}
	}
	skipUploads := len(uploadStates) == 0 && len(indexStates) != 0
	skipIndexes := len(uploadStates) != 0 && len(indexStates) == 0

	repositoryID := 0
	if args.Repository != nil {
		repositoryID, err = resolveRepositoryID(*args.Repository)
		if err != nil {
			return nil, err
		}
	}
	term := derefString(args.Query, "")

	visibleAtTip := false
	if args.IsLatestForRepo != nil {
		visibleAtTip = *args.IsLatestForRepo
		skipIndexes = true
	}

	if !skipUploads {
		if err := r.uploadSvc.ReindexUploads(ctx, uploadsshared.ReindexUploadsOptions{
			States:       uploadStates,
			Term:         term,
			RepositoryID: repositoryID,
			VisibleAtTip: visibleAtTip,
		}); err != nil {
			return nil, err
		}
	}
	if !skipIndexes {
		if err := r.autoindexSvc.ReindexIndexes(ctx, autoindexingshared.ReindexIndexesOptions{
			States:        indexStates,
			Term:          term,
			RepositoryID:  repositoryID,
			WithoutUpload: true,
		}); err != nil {
			return nil, err
		}
	}

	return &resolverstubs.EmptyResponse{}, nil
}

type preciseIndexConnectionResolver struct {
	nodes      []resolverstubs.PreciseIndexResolver
	totalCount int
	cursor     string
}

func NewPreciseIndexConnectionResolver(
	nodes []resolverstubs.PreciseIndexResolver,
	totalCount int,
	cursor string,
) resolverstubs.PreciseIndexConnectionResolver {
	return &preciseIndexConnectionResolver{
		nodes:      nodes,
		totalCount: totalCount,
		cursor:     cursor,
	}
}

func (r *preciseIndexConnectionResolver) Nodes(ctx context.Context) ([]resolverstubs.PreciseIndexResolver, error) {
	return r.nodes, nil
}

func (r *preciseIndexConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	count := int32(r.totalCount)
	return &count, nil
}

func (r *preciseIndexConnectionResolver) PageInfo(ctx context.Context) (resolverstubs.PageInfo, error) {
	if r.cursor != "" {
		return &pageInfo{hasNextPage: true, endCursor: &r.cursor}, nil
	}

	return &pageInfo{hasNextPage: false}, nil
}

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
	prefetcher *sharedresolvers.Prefetcher,
	locationResolver *sharedresolvers.CachedLocationResolver,
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
		uploadResolver = sharedresolvers.NewUploadResolver(uploadsSvc, autoindexingSvc, policySvc, *upload, prefetcher, locationResolver, traceErrs)

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
		indexResolver = sharedresolvers.NewIndexResolver(autoindexingSvc, uploadsSvc, policySvc, *index, prefetcher, locationResolver, traceErrs)
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

func unmarshalPreciseIndexGQLID(id graphql.ID) (uploadID, indexID int, err error) {
	uploadID, indexID, err = unmarshalRawPreciseIndexGQLID(id)
	if err == nil && uploadID == 0 && indexID == 0 {
		err = errors.New("no payload")
	}

	return uploadID, indexID, errors.Wrap(err, "unexpected precise index ID")
}

var errExpectedPairs = errors.New("expected pairs of `U:<id>`, `I:<id>`")

func unmarshalRawPreciseIndexGQLID(id graphql.ID) (uploadID, indexID int, err error) {
	var rawPayload string
	if err := relay.UnmarshalSpec(id, &rawPayload); err != nil {
		return 0, 0, errors.Wrap(err, "unexpected precise index ID")
	}

	parts := strings.Split(rawPayload, ":")
	if len(parts)%2 != 0 {
		return 0, 0, errExpectedPairs
	}
	for i := 0; i < len(parts)-1; i += 2 {
		id, err := strconv.Atoi(parts[i+1])
		if err != nil {
			return 0, 0, errExpectedPairs
		}

		switch parts[i] {
		case "U":
			uploadID = id
		case "I":
			indexID = id
		default:
			return 0, 0, errExpectedPairs
		}
	}

	return uploadID, indexID, nil
}

type pageInfo struct {
	endCursor   *string
	hasNextPage bool
}

func (r *pageInfo) EndCursor() *string { return r.endCursor }
func (r *pageInfo) HasNextPage() bool  { return r.hasNextPage }

func bifurcateStates(states []string) (uploadStates, indexStates []string, _ error) {
	for _, state := range states {
		switch strings.ToUpper(state) {
		case "QUEUED_FOR_INDEXING":
			indexStates = append(indexStates, "queued")
		case "INDEXING":
			indexStates = append(indexStates, "processing")
		case "INDEXING_ERRORED":
			indexStates = append(indexStates, "errored")

		case "UPLOADING_INDEX":
			uploadStates = append(uploadStates, "uploading")
		case "QUEUED_FOR_PROCESSING":
			uploadStates = append(uploadStates, "queued")
		case "PROCESSING":
			uploadStates = append(uploadStates, "processing")
		case "PROCESSING_ERRORED":
			uploadStates = append(uploadStates, "errored")
		case "COMPLETED":
			uploadStates = append(uploadStates, "completed")
		case "DELETING":
			uploadStates = append(uploadStates, "deleting")
		case "DELETED":
			uploadStates = append(uploadStates, "deleted")

		default:
			return nil, nil, errors.Newf("filtering by state %q is unsupported", state)
		}
	}

	return uploadStates, indexStates, nil
}
