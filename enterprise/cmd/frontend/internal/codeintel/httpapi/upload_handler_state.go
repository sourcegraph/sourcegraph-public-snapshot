package httpapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type uploadState struct {
	uploadID         int
	numParts         int
	uploadedParts    []int
	multipart        bool
	suppliedIndex    bool
	index            int
	done             bool
	uncompressedSize *int64
	metadata         codeintelUploadMetadata
}

type codeintelUploadMetadata struct {
	repositoryID      int
	commit            string
	root              string
	indexer           string
	indexerVersion    string
	associatedIndexID int
}

var revhashPattern = lazyregexp.New(`^[a-z0-9]{40}$`)

// constructUploadState reads the query args of the given HTTP request and populates an upload state object.
// This function should be used instead of reading directly from the request as the upload state's fields are
// backfilled/denormalized from the database, depending on the type of request.
func (h *UploadHandler) constructUploadState(ctx context.Context, r *http.Request) (uploadState, int, error) {
	uploadState := uploadState{
		uploadID:      getQueryInt(r, "uploadId"),
		suppliedIndex: hasQuery(r, "index"),
		index:         getQueryInt(r, "index"),
		done:          hasQuery(r, "done"),
	}

	if uploadState.uploadID == 0 {
		return h.hydrateUploadStateFromRequest(ctx, r, uploadState)
	}

	// An upload identifier was supplied; this is a subsequent request of a multi-part
	// upload. Fetch the upload record to ensure that it hasn't since been deleted by
	// the user.
	upload, exists, err := h.dbStore.GetUploadByID(ctx, uploadState.uploadID)
	if err != nil {
		return uploadState, http.StatusInternalServerError, err
	}
	if !exists {
		return uploadState, http.StatusNotFound, errors.Errorf("upload not found")
	}

	return h.hydrateUploadStateFromRecord(ctx, r, uploadState, upload)
}

func (h *UploadHandler) hydrateUploadStateFromRequest(ctx context.Context, r *http.Request, uploadState uploadState) (uploadState, int, error) {
	uncompressedSize := new(int64)
	if size := r.Header.Get("X-Uncompressed-Size"); size != "" {
		parsedSize, err := strconv.ParseInt(size, 10, 64)
		if err != nil {
			return uploadState, http.StatusUnprocessableEntity, errors.New("the header `X-Uncompressed-Size` must be an integer")
		}

		*uncompressedSize = parsedSize
	}

	metadata, statusCode, err := h.metadataFromRequest(ctx, r)
	if err != nil {
		return uploadState, statusCode, err
	}

	uploadState.multipart = hasQuery(r, "multiPart")
	uploadState.numParts = getQueryInt(r, "numParts")
	uploadState.uncompressedSize = uncompressedSize
	uploadState.metadata = metadata

	return uploadState, 0, nil
}

func (h *UploadHandler) hydrateUploadStateFromRecord(_ context.Context, r *http.Request, uploadState uploadState, upload dbstore.Upload) (uploadState, int, error) {
	metadata, statusCode, err := h.metadataFromRecord(upload)
	if err != nil {
		return uploadState, statusCode, err
	}

	// Stash all fields given in the initial request
	uploadState.numParts = upload.NumParts
	uploadState.uploadedParts = upload.UploadedParts
	uploadState.uncompressedSize = upload.UncompressedSize
	uploadState.metadata = metadata

	return uploadState, 0, nil
}

func (h *UploadHandler) metadataFromRequest(ctx context.Context, r *http.Request) (codeintelUploadMetadata, int, error) {
	commit := getQuery(r, "commit")
	if !revhashPattern.Match([]byte(commit)) {
		return codeintelUploadMetadata{}, http.StatusBadRequest, errors.Errorf("commit must be a 40-character revhash")
	}

	// Ensure that the repository and commit given in the request are resolvable.
	repositoryName := getQuery(r, "repository")
	repositoryID, statusCode, err := ensureRepoAndCommitExist(ctx, h.logger, h.db, repositoryName, commit)
	if err != nil {
		return codeintelUploadMetadata{}, statusCode, err
	}

	// Populate state from request
	return codeintelUploadMetadata{
		repositoryID:      repositoryID,
		commit:            commit,
		root:              sanitizeRoot(getQuery(r, "root")),
		indexer:           getQuery(r, "indexerName"),
		indexerVersion:    getQuery(r, "indexerVersion"),
		associatedIndexID: getQueryInt(r, "associatedIndexId"),
	}, 0, nil
}

func (h *UploadHandler) metadataFromRecord(upload dbstore.Upload) (codeintelUploadMetadata, int, error) {
	associatedIndexID := 0
	if upload.AssociatedIndexID != nil {
		associatedIndexID = *upload.AssociatedIndexID
	}

	return codeintelUploadMetadata{
		repositoryID:      upload.RepositoryID,
		commit:            upload.Commit,
		root:              upload.Root,
		indexer:           upload.Indexer,
		indexerVersion:    upload.IndexerVersion,
		associatedIndexID: associatedIndexID,
	}, 0, nil
}

func ensureRepoAndCommitExist(ctx context.Context, logger log.Logger, db database.DB, repoName, commit string) (int, int, error) {
	// 🚨 SECURITY: Bypass authz here; we've already determined that the current request is
	// authorized to view the target repository; they are either a site admin or the code
	// host has explicit listed them with some level of access (depending on the code host).
	ctx = actor.WithInternalActor(ctx)

	//
	// 1. Resolve repository

	repo, err := backend.NewRepos(logger, db).GetByName(ctx, api.RepoName(repoName))
	if err != nil {
		if errcode.IsNotFound(err) {
			return 0, http.StatusNotFound, errors.Errorf("unknown repository %q", repoName)
		}

		return 0, http.StatusInternalServerError, err
	}

	//
	// 2. Resolve commit

	if _, err := backend.NewRepos(logger, db).ResolveRev(ctx, repo, commit); err != nil {
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

func sanitizeRoot(s string) string {
	if s == "" || s == "/" {
		return ""
	}
	if !strings.HasSuffix(s, "/") {
		s += "/"
	}
	return s
}
