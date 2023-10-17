// Package repoupdater implements the repo-updater service HTTP handler.
package repoupdater

import (
	"context"
	"encoding/json"
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
	mux.HandleFunc("/repo-update-scheduler-info", trace.WithRouteName("repo-update-scheduler-info", s.handleRepoUpdateSchedulerInfo))
	mux.HandleFunc("/repo-lookup", trace.WithRouteName("repo-lookup", s.handleRepoLookup))
	mux.HandleFunc("/enqueue-repo-update", trace.WithRouteName("enqueue-repo-update", s.handleEnqueueRepoUpdate))
	mux.HandleFunc("/enqueue-changeset-sync", trace.WithRouteName("enqueue-changeset-sync", s.handleEnqueueChangesetSync))
	return mux
}

func (s *Server) handleRepoUpdateSchedulerInfo(w http.ResponseWriter, r *http.Request) {
	var args protocol.RepoUpdateSchedulerInfoArgs
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		s.respond(w, http.StatusBadRequest, err)
		return
	}

	result := s.Scheduler.ScheduleInfo(args.ID)
	s.respond(w, http.StatusOK, result)
}

func (s *Server) handleRepoLookup(w http.ResponseWriter, r *http.Request) {
	var args protocol.RepoLookupArgs
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := s.repoLookup(r.Context(), args.Repo)
	if err != nil {
		if r.Context().Err() != nil {
			http.Error(w, "request canceled", http.StatusGatewayTimeout)
			return
		}
		s.Logger.Error("repoLookup failed", log.String("name", string(args.Repo)), log.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respond(w, http.StatusOK, result)
}

func (s *Server) handleEnqueueRepoUpdate(w http.ResponseWriter, r *http.Request) {
	var req protocol.RepoUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respond(w, http.StatusBadRequest, err)
		return
	}
	result, status, err := s.enqueueRepoUpdate(r.Context(), &req)
	if err != nil {
		s.Logger.Warn("enqueueRepoUpdate failed", log.String("req", fmt.Sprint(req)), log.Error(err))
		s.respond(w, status, err)
		return
	}
	s.respond(w, status, result)
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

func (s *Server) respond(w http.ResponseWriter, code int, v any) {
	switch val := v.(type) {
	case error:
		if val != nil {
			s.Logger.Error("response value error", log.Error(val))
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(code)
			fmt.Fprintf(w, "%v", val)
		}
	default:
		w.Header().Set("Content-Type", "application/json")
		bs, err := json.Marshal(v)
		if err != nil {
			s.respond(w, http.StatusInternalServerError, err)
			return
		}

		w.WriteHeader(code)
		if _, err = w.Write(bs); err != nil {
			s.Logger.Error("failed to write response", log.Error(err))
		}
	}
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
		return nil, err
	}

	return &protocol.RepoLookupResult{Repo: protocol.NewRepoInfo(repo)}, nil
}

func (s *Server) handleEnqueueChangesetSync(w http.ResponseWriter, r *http.Request) {
	if s.ChangesetSyncRegistry == nil {
		s.Logger.Warn("ChangesetSyncer is nil")
		s.respond(w, http.StatusForbidden, nil)
		return
	}

	var req protocol.ChangesetSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respond(w, http.StatusBadRequest, err)
		return
	}
	if len(req.IDs) == 0 {
		s.respond(w, http.StatusBadRequest, errors.New("no ids provided"))
		return
	}
	err := s.ChangesetSyncRegistry.EnqueueChangesetSyncs(r.Context(), req.IDs)
	if err != nil {
		resp := protocol.ChangesetSyncResponse{Error: err.Error()}
		s.respond(w, http.StatusInternalServerError, resp)
		return
	}
	s.respond(w, http.StatusOK, nil)
}
