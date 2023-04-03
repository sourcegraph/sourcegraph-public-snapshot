package graphql

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	uploadsshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/api"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
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
		v, err := resolverstubs.UnmarshalID[api.RepoID](*args.Repo)
		if err != nil {
			return nil, err
		}

		repositoryID = int(v)
	}

	term := ""
	if args.Query != nil {
		term = *args.Query
	}

	var indexerNames []string
	if args.IndexerKey != nil {
		indexerNames = types.NamesForKey(*args.IndexerKey)
	}

	var uploads []types.Upload
	totalUploadCount := 0
	if !skipUploads {
		if uploads, totalUploadCount, err = r.uploadSvc.GetUploads(ctx, uploadsshared.GetUploadsOptions{
			RepositoryID:       repositoryID,
			States:             uploadStates,
			Term:               term,
			DependencyOf:       dependencyOf,
			DependentOf:        dependentOf,
			AllowDeletedUpload: args.IncludeDeleted != nil && *args.IncludeDeleted,
			IndexerNames:       indexerNames,
			Limit:              pageSize,
			Offset:             uploadOffset,
		}); err != nil {
			return nil, err
		}
	}

	var indexes []types.Index
	totalIndexCount := 0
	if !skipIndexes {
		if indexes, totalIndexCount, err = r.uploadSvc.GetIndexes(ctx, uploadsshared.GetIndexesOptions{
			RepositoryID:  repositoryID,
			States:        indexStates,
			Term:          term,
			IndexerNames:  indexerNames,
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

	prefetcher := r.prefetcherFactory.Create()

	for _, pair := range pairs {
		if pair.upload != nil && pair.upload.AssociatedIndexID != nil {
			prefetcher.MarkIndex(*pair.upload.AssociatedIndexID)
		}
	}

	resolvers := make([]resolverstubs.PreciseIndexResolver, 0, len(pairs))
	for _, pair := range pairs {
		resolver, err := NewPreciseIndexResolver(ctx, r.uploadSvc, r.policySvc, r.gitserverClient, prefetcher, r.siteAdminChecker, r.repoStore, r.locationResolverFactory.Create(), errTracer, pair.upload, pair.index)
		if err != nil {
			return nil, err
		}

		resolvers = append(resolvers, resolver)
	}

	return resolverstubs.NewCursorWithTotalCountConnectionResolver(resolvers, cursor, int32(totalUploadCount+totalIndexCount)), nil
}

func (r *rootResolver) PreciseIndexByID(ctx context.Context, id graphql.ID) (_ resolverstubs.PreciseIndexResolver, err error) {
	ctx, errTracer, endObservation := r.operations.preciseIndexByID.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("id", string(id)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	uploadID, indexID, err := unmarshalPreciseIndexGQLID(id)
	if err != nil {
		return nil, err
	}

	if uploadID != 0 {
		upload, ok, err := r.uploadSvc.GetUploadByID(ctx, uploadID)
		if err != nil || !ok {
			return nil, err
		}

		return NewPreciseIndexResolver(ctx, r.uploadSvc, r.policySvc, r.gitserverClient, r.prefetcherFactory.Create(), r.siteAdminChecker, r.repoStore, r.locationResolverFactory.Create(), errTracer, &upload, nil)
	}
	if indexID != 0 {
		index, ok, err := r.uploadSvc.GetIndexByID(ctx, indexID)
		if err != nil || !ok {
			return nil, err
		}

		return NewPreciseIndexResolver(ctx, r.uploadSvc, r.policySvc, r.gitserverClient, r.prefetcherFactory.Create(), r.siteAdminChecker, r.repoStore, r.locationResolverFactory.Create(), errTracer, nil, &index)
	}

	return nil, errors.New("invalid identifier")
}

func (r *rootResolver) IndexerKeys(ctx context.Context, args *resolverstubs.IndexerKeyQueryArgs) ([]string, error) {
	var repositoryID int
	if args.Repo != nil {
		v, err := resolverstubs.UnmarshalID[api.RepoID](*args.Repo)
		if err != nil {
			return nil, err
		}

		repositoryID = int(v)
	}

	indexers, err := r.uploadSvc.GetIndexers(ctx, uploadsshared.GetIndexersOptions{
		RepositoryID: repositoryID,
	})
	if err != nil {
		return nil, err
	}

	keyMap := map[string]struct{}{}
	for _, indexer := range indexers {
		keyMap[types.NewCodeIntelIndexerResolver(indexer, "").Key()] = struct{}{}
	}

	var keys []string
	for key := range keyMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return keys, nil
}
