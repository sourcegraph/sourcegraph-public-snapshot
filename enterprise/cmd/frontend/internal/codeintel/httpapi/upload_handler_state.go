package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type uploadState struct {
	repositoryName    string
	repositoryID      int
	uploadID          int
	commit            string
	root              string
	indexer           string
	indexerVersion    string
	associatedIndexID int
	numParts          int
	uploadedParts     []int
	multipart         bool
	suppliedIndex     bool
	index             int
	done              bool
}

var revhashPattern = lazyregexp.New(`^[a-z0-9]{40}$`)

// constructUploadState reads the query args of the given HTTP request and populates an upload state object.
// This function should be used instead of reading directly from the request as the upload state's fields are
// backfilled/denormalized from the database, depending on the type of request.
func (h *UploadHandler) constructUploadState(ctx context.Context, r *http.Request) (uploadState, int, error) {
	uploadState := uploadState{
		repositoryName:    getQuery(r, "repository"),
		uploadID:          getQueryInt(r, "uploadId"),
		commit:            getQuery(r, "commit"),
		root:              sanitizeRoot(getQuery(r, "root")),
		indexer:           getQuery(r, "indexerName"),
		indexerVersion:    getQuery(r, "indexerVersion"),
		associatedIndexID: getQueryInt(r, "associatedIndexId"),
		numParts:          getQueryInt(r, "numParts"),
		multipart:         hasQuery(r, "multiPart"),
		suppliedIndex:     hasQuery(r, "index"),
		index:             getQueryInt(r, "index"),
		done:              hasQuery(r, "done"),
	}

	if uploadState.commit != "" && !revhashPattern.Match([]byte(uploadState.commit)) {
		return uploadState, http.StatusBadRequest, errors.Errorf("commit must be a 40-character revhash")
	}

	if uploadState.uploadID == 0 {
		// No upload identifier supplied; this is a single payload upload or the start
		// of a multi-part upload. Ensure that the repository and commit given in the
		// request are resolvable. Subsequent multi-part requests will use the new
		// upload identifier returned in this response.
		repositoryID, statusCode, err := ensureRepoAndCommitExist(ctx, h.db, uploadState.repositoryName, uploadState.commit)
		if err != nil {
			return uploadState, statusCode, err
		}

		// Stash repository id (user only gives us the name)
		uploadState.repositoryID = repositoryID
	} else {
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

		// Stash all fields given in the initial request
		uploadState.repositoryID = upload.RepositoryID
		uploadState.commit = upload.Commit
		uploadState.root = upload.Root
		uploadState.numParts = upload.NumParts
		uploadState.uploadedParts = upload.UploadedParts
	}

	return uploadState, 0, nil
}

func ensureRepoAndCommitExist(ctx context.Context, db database.DB, repoName, commit string) (int, int, error) {
	// 🚨 SECURITY: Bypass authz here; we've already determined that the current request is
	// authorized to view the target repository; they are either a site admin or the code
	// host has explicit listed them with some level of access (depending on the code host).
	ctx = actor.WithInternalActor(ctx)

	//
	// 1. Resolve repository

	repo, err := backend.NewRepos(db).GetByName(ctx, api.RepoName(repoName))
	if err != nil {
		if errcode.IsNotFound(err) {
			return 0, http.StatusNotFound, errors.Errorf("unknown repository %q", repoName)
		}

		return 0, http.StatusInternalServerError, err
	}

	//
	// 2. Resolve commit

	if _, err := backend.NewRepos(db).ResolveRev(ctx, repo, commit); err != nil {
		var reason string
		if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
			reason = "commit not found"
		} else if gitdomain.IsCloneInProgress(err) {
			reason = "repository still cloning"
		} else {
			return 0, http.StatusInternalServerError, err
		}

		log15.Warn("Accepting LSIF upload with unresolvable commit", "reason", reason)
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
