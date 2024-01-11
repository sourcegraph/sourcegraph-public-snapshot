// Package repoupdater implements the repo-updater service HTTP handler.
package repoupdater

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/batches/syncer"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Server is a repoupdater server.
type Server struct {
	repos.Store
	*repos.Syncer
	Logger                log.Logger
	ObservationCtx        *observation.Context
	SourcegraphDotComMode bool
	Scheduler             interface {
		UpdateOnce(id api.RepoID, name api.RepoName)
		ScheduleInfo(id api.RepoID) *protocol.RepoUpdateSchedulerInfoResult
	}
	ChangesetSyncRegistry syncer.ChangesetSyncRegistry
}

// Handler returns the http.Handler that should be used to serve requests.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", trace.WithRouteName("healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	return mux
}

func (s *Server) enqueueRepoUpdate(ctx context.Context, req *protocol.RepoUpdateRequest) (resp *protocol.RepoUpdateResponse, httpStatus int, err error) {
	tr, ctx := trace.New(ctx, "enqueueRepoUpdate", attribute.Stringer("req", req))
	defer func() {
		s.Logger.Debug("enqueueRepoUpdate", log.Object("http", log.Int("status", httpStatus), log.String("resp", fmt.Sprint(resp)), log.Error(err)))
		if resp != nil {
			tr.SetAttributes(
				attribute.Int("resp.id", int(resp.ID)),
				attribute.String("resp.name", resp.Name),
			)
		}
		tr.SetError(err)
		tr.End()
	}()

	rs, err := s.Store.RepoStore().List(ctx, database.ReposListOptions{Names: []string{string(req.Repo)}})
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "store.list-repos")
	}

	if len(rs) != 1 {
		return nil, http.StatusNotFound, errors.Errorf("repo %q not found in store", req.Repo)
	}

	repo := rs[0]

	s.Scheduler.UpdateOnce(repo.ID, repo.Name)

	return &protocol.RepoUpdateResponse{
		ID:   repo.ID,
		Name: string(repo.Name),
	}, http.StatusOK, nil
}

var mockRepoLookup func(api.RepoName) (*protocol.RepoLookupResult, error)

func (s *Server) repoLookup(ctx context.Context, repoName api.RepoName) (result *protocol.RepoLookupResult, err error) {
	// Sourcegraph.com: this is on the user path, do not block forever if codehost is
	// being bad. Ideally block before cloudflare 504s the request (1min). Other: we
	// only speak to our database, so response should be in a few ms.
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tr, ctx := trace.New(ctx, "repoLookup", attribute.String("repo", string(repoName)))
	defer func() {
		s.Logger.Debug("repoLookup", log.String("result", fmt.Sprint(result)), log.Error(err))
		tr.SetError(err)
		tr.End()
	}()

	if repoName == "" {
		return nil, errors.New("Repo must be set (is blank)")
	}

	if mockRepoLookup != nil {
		return mockRepoLookup(repoName)
	}

	repo, err := s.Syncer.SyncRepo(ctx, repoName, true)
	if err != nil {
		if errcode.IsNotFound(err) {
			return &protocol.RepoLookupResult{ErrorNotFound: true}, nil
		}
		if errcode.IsUnauthorized(err) || errcode.IsForbidden(err) {
			return &protocol.RepoLookupResult{ErrorUnauthorized: true}, nil
		}
		if errcode.IsTemporary(err) {
			return &protocol.RepoLookupResult{ErrorTemporarilyUnavailable: true}, nil
		}
		if errcode.IsRepoDenied(err) {
			return &protocol.RepoLookupResult{ErrorRepoDenied: err.Error()}, nil
		}
		return nil, err
	}

	return &protocol.RepoLookupResult{Repo: protocol.NewRepoInfo(repo)}, nil
}
