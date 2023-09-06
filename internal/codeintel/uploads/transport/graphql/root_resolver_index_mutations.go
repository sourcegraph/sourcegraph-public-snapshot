package graphql

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// 🚨 SECURITY: Only site admins may modify code intelligence upload data
func (r *rootResolver) DeletePreciseIndex(ctx context.Context, args *struct{ ID graphql.ID }) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.deletePreciseIndex.With(ctx, &err, observation.Args{})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	uploadID, indexID, err := UnmarshalPreciseIndexGQLID(args.ID)
	if err != nil {
		return nil, err
	}
	if uploadID != 0 {
		if _, err := r.uploadSvc.DeleteUploadByID(ctx, uploadID); err != nil {
			return nil, err
		}
	} else if indexID != 0 {
		if _, err := r.uploadSvc.DeleteIndexByID(ctx, indexID); err != nil {
			return nil, err
		}
	}

	return resolverstubs.Empty, nil
}

// 🚨 SECURITY: Only site admins may modify code intelligence upload data
func (r *rootResolver) DeletePreciseIndexes(ctx context.Context, args *resolverstubs.DeletePreciseIndexesArgs) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.deletePreciseIndexes.With(ctx, &err, observation.Args{})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
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

	var indexerNames []string
	if args.IndexerKey != nil {
		indexerNames = uploadsshared.NamesForKey(*args.IndexerKey)
	}

	repositoryID := 0
	if args.Repository != nil {
		repositoryID, err = resolverstubs.UnmarshalID[int](*args.Repository)
		if err != nil {
			return nil, err
		}
	}
	term := pointers.Deref(args.Query, "")

	visibleAtTip := false
	if args.IsLatestForRepo != nil {
		visibleAtTip = *args.IsLatestForRepo
		skipIndexes = true
	}

	if !skipUploads {
		if err := r.uploadSvc.DeleteUploads(ctx, uploadsshared.DeleteUploadsOptions{
			RepositoryID: repositoryID,
			States:       uploadStates,
			IndexerNames: indexerNames,
			Term:         term,
			VisibleAtTip: visibleAtTip,
		}); err != nil {
			return nil, err
		}
	}
	if !skipIndexes {
		if err := r.uploadSvc.DeleteIndexes(ctx, uploadsshared.DeleteIndexesOptions{
			RepositoryID:  repositoryID,
			States:        indexStates,
			IndexerNames:  indexerNames,
			Term:          term,
			WithoutUpload: true,
		}); err != nil {
			return nil, err
		}
	}

	return resolverstubs.Empty, nil
}

// 🚨 SECURITY: Only site admins may modify code intelligence upload data
func (r *rootResolver) ReindexPreciseIndex(ctx context.Context, args *struct{ ID graphql.ID }) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.reindexPreciseIndex.With(ctx, &err, observation.Args{})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	uploadID, indexID, err := UnmarshalPreciseIndexGQLID(args.ID)
	if err != nil {
		return nil, err
	}
	if uploadID != 0 {
		if err := r.uploadSvc.ReindexUploadByID(ctx, uploadID); err != nil {
			return nil, err
		}
	} else if indexID != 0 {
		if err := r.uploadSvc.ReindexIndexByID(ctx, indexID); err != nil {
			return nil, err
		}
	}

	return resolverstubs.Empty, nil
}

// 🚨 SECURITY: Only site admins may modify code intelligence upload data
func (r *rootResolver) ReindexPreciseIndexes(ctx context.Context, args *resolverstubs.ReindexPreciseIndexesArgs) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.reindexPreciseIndexes.With(ctx, &err, observation.Args{})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
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

	var indexerNames []string
	if args.IndexerKey != nil {
		indexerNames = uploadsshared.NamesForKey(*args.IndexerKey)
	}

	repositoryID := 0
	if args.Repository != nil {
		repositoryID, err = resolverstubs.UnmarshalID[int](*args.Repository)
		if err != nil {
			return nil, err
		}
	}
	term := pointers.Deref(args.Query, "")

	visibleAtTip := false
	if args.IsLatestForRepo != nil {
		visibleAtTip = *args.IsLatestForRepo
		skipIndexes = true
	}

	if !skipUploads {
		if err := r.uploadSvc.ReindexUploads(ctx, uploadsshared.ReindexUploadsOptions{
			States:       uploadStates,
			IndexerNames: indexerNames,
			Term:         term,
			RepositoryID: repositoryID,
			VisibleAtTip: visibleAtTip,
		}); err != nil {
			return nil, err
		}
	}
	if !skipIndexes {
		if err := r.uploadSvc.ReindexIndexes(ctx, uploadsshared.ReindexIndexesOptions{
			States:        indexStates,
			IndexerNames:  indexerNames,
			Term:          term,
			RepositoryID:  repositoryID,
			WithoutUpload: true,
		}); err != nil {
			return nil, err
		}
	}

	return resolverstubs.Empty, nil
}
