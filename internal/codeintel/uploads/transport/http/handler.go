package http

import (
	"context"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/uploadhandler"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var revhashPattern = lazyregexp.New(`^[a-z0-9]{40}$`)

func newHandler(
	repoStore RepoStore,
	uploadStore uploadstore.Store,
	dbStore uploadhandler.DBStore[uploads.UploadMetadata],
	operations *uploadhandler.Operations,
) http.Handler {
	logger := log.Scoped("UploadHandler")

	metadataFromRequest := func(ctx context.Context, r *http.Request) (uploads.UploadMetadata, int, error) {
		commit := getQuery(r, "commit")
		if !revhashPattern.Match([]byte(commit)) {
			return uploads.UploadMetadata{}, http.StatusBadRequest, errors.Errorf("commit must be a 40-character revhash")
		}

		// Ensure that the repository and commit given in the request are resolvable.
		repositoryName := getQuery(r, "repository")
		repositoryID, statusCode, err := ensureRepoAndCommitExist(ctx, repoStore, repositoryName, commit, logger)
		if err != nil {
			return uploads.UploadMetadata{}, statusCode, err
		}

		contentType := r.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/x-ndjson+lsif"
		}

		// Populate state from request
		return uploads.UploadMetadata{
			RepositoryID:      repositoryID,
			Commit:            commit,
			Root:              sanitizeRoot(getQuery(r, "root")),
			Indexer:           getQuery(r, "indexerName"),
			IndexerVersion:    getQuery(r, "indexerVersion"),
			AssociatedIndexID: getQueryInt(r, "associatedIndexId"),
			ContentType:       contentType,
		}, 0, nil
	}

	handler := uploadhandler.NewUploadHandler(
		logger,
		dbStore,
		uploadStore,
		operations,
		metadataFromRequest,
	)

	return handler
}

func ensureRepoAndCommitExist(ctx context.Context, repoStore RepoStore, repoName, commit string, logger log.Logger) (int, int, error) {
	// 🚨 SECURITY: Bypass authz here; we've already determined that the current request is
	// authorized to view the target repository; they are either a site admin or the code
	// host has explicit listed them with some level of access (depending on the code host).
	ctx = actor.WithInternalActor(ctx)

	//
	// 1. Resolve repository

	repo, err := repoStore.GetByName(ctx, api.RepoName(repoName))
	if err != nil {
		if errcode.IsNotFound(err) {
			return 0, http.StatusNotFound, errors.Errorf("unknown repository %q", repoName)
		}

		return 0, http.StatusInternalServerError, err
	}

	//
	// 2. Resolve commit

	if _, err := repoStore.ResolveRev(ctx, repo, commit); err != nil {
		var reason string
		if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
			reason = "commit not found"
		} else if gitdomain.IsCloneInProgress(err) {
			reason = "repository still cloning"
		} else {
			return 0, http.StatusInternalServerError, err
		}

		logger.Warn("Accepting LSIF upload with unresolvable commit", log.String("reason", reason))
	}

	return int(repo.ID), 0, nil
}
