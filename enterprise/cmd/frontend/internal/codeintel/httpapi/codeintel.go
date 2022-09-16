package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/sourcegraph/log"
	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/httpapi/auth"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewUploadHandler(
	db database.DB,
	dbStore DBStore,
	uploadStore uploadstore.Store,
	internal bool,
	authValidators auth.AuthValidatorMap,
	operations *Operations,
) http.Handler {
	handler := &UploadHandler{
		logger: sglog.Scoped("UploadHandler", "").With(
			sglog.Bool("internal", internal),
		),
		db:          db,
		dbStore:     dbStore,
		uploadStore: uploadStore,
		operations:  operations,
	}

	if internal {
		return http.HandlerFunc(handler.handleEnqueue)
	}

	// 🚨 SECURITY: Non-internal installations of this handler will require a user/repo
	// visibility check with the remote code host (if enabled via site configuration).
	return auth.AuthMiddleware(http.HandlerFunc(handler.handleEnqueue), db, authValidators, operations.authMiddleware)
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
