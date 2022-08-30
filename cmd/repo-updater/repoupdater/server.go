// Package repoupdater implements the repo-updater service HTTP handler.
package repoupdater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/batches"
	livedependencies "github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/live"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Server is a repoupdater server.
type Server struct {
	repos.Store
	*repos.Syncer
	Logger                log.Logger
	SourcegraphDotComMode bool
	Scheduler             interface {
		UpdateOnce(id api.RepoID, name api.RepoName)
		ScheduleInfo(id api.RepoID) *protocol.RepoUpdateSchedulerInfoResult
	}
	ChangesetSyncRegistry batches.ChangesetSyncRegistry
	RateLimitSyncer       interface {
		// SyncRateLimiters should be called when an external service changes so that
		// our internal rate limiters are kept in sync
		SyncRateLimiters(ctx context.Context, ids ...int64) error
	}
	PermsSyncer interface {
		// ScheduleUsers schedules new permissions syncing requests for given users.
		ScheduleUsers(ctx context.Context, opts authz.FetchPermsOptions, userIDs ...int32)
		// ScheduleRepos schedules new permissions syncing requests for given repositories.
		ScheduleRepos(ctx context.Context, repoIDs ...api.RepoID)
	}
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
	mux.HandleFunc("/sync-external-service", trace.WithRouteName("sync-external-service", s.handleExternalServiceSync))
	mux.HandleFunc("/enqueue-changeset-sync", trace.WithRouteName("enqueue-changeset-sync", s.handleEnqueueChangesetSync))
	mux.HandleFunc("/schedule-perms-sync", trace.WithRouteName("schedule-perms-sync", s.handleSchedulePermsSync))
	return mux
}

// TODO(tsenart): Reuse this function in all handlers.
func (s *Server) respond(w http.ResponseWriter, code int, v any) {
	switch val := v.(type) {
	case error:
		if val != nil {
			s.Logger.Error("response value error", log.String("err", val.Error()))
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

func (s *Server) handleRepoUpdateSchedulerInfo(w http.ResponseWriter, r *http.Request) {
	var args protocol.RepoUpdateSchedulerInfoArgs
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := s.Scheduler.ScheduleInfo(args.ID)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleRepoLookup(w http.ResponseWriter, r *http.Request) {
	var args protocol.RepoLookupArgs
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := s.repoLookup(r.Context(), args)
	if err != nil {
		if r.Context().Err() != nil {
			http.Error(w, "request canceled", http.StatusGatewayTimeout)
			return
		}
		s.Logger.Error("repoLookup failed",
			log.Object("repo",
				log.String("name", string(args.Repo)),
				log.Bool("update", args.Update),
			),
			log.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleEnqueueRepoUpdate(w http.ResponseWriter, r *http.Request) {
	var req protocol.RepoUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respond(w, http.StatusBadRequest, err)
		return
	}
	result, status, err := s.enqueueRepoUpdate(r.Context(), &req)
	if err != nil {
		s.Logger.Error("enqueueRepoUpdate failed", log.String("req", fmt.Sprint(req)), log.Error(err))
		s.respond(w, status, err)
		return
	}
	s.respond(w, status, result)
}

func (s *Server) enqueueRepoUpdate(ctx context.Context, req *protocol.RepoUpdateRequest) (resp *protocol.RepoUpdateResponse, httpStatus int, err error) {
	tr, ctx := trace.New(ctx, "enqueueRepoUpdate", req.String())
	defer func() {
		s.Logger.Debug("enqueueRepoUpdate", log.Object("http", log.Int("status", httpStatus), log.String("resp", fmt.Sprint(resp)), log.Error(err)))
		if resp != nil {
			tr.LogFields(
				otlog.Int32("resp.id", int32(resp.ID)),
				otlog.String("resp.name", resp.Name),
			)
		}
		tr.SetError(err)
		tr.Finish()
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

func (s *Server) handleExternalServiceSync(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var req protocol.ExternalServiceSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logger := s.Logger.With(log.Object("ExternalService",
		log.Int64("id", req.ExternalServiceID)),
	)

	var sourcer repos.Sourcer
	if sourcer = s.Sourcer; sourcer == nil {
		db := database.NewDBWith(logger, s)
		depsSvc := livedependencies.GetService(db)
		sourcer = repos.NewSourcer(s.Logger.Scoped("repos.Sourcer", ""), db, httpcli.ExternalClientFactory, repos.WithDependenciesService(depsSvc))
	}

	externalServiceID := req.ExternalServiceID

	// We want to get soft-deleted external services as well, since we do a final
	// sync when an external service gets deleted.
	es, err := s.ExternalServiceStore().List(ctx, database.ExternalServicesListOptions{
		IDs:            []int64{externalServiceID},
		IncludeDeleted: true,
	})
	if err != nil {
		s.respond(w, http.StatusInternalServerError, err)
		return
	}

	if len(es) != 1 {
		s.respond(w, http.StatusNotFound, errors.Newf("external service %d not found", externalServiceID))
		return
	}

	src, err := sourcer(ctx, es[0])

	if err != nil {
		logger.Error("server.external-service-sync", log.Error(err))
		return
	}

	err = externalServiceValidate(ctx, es[0], src)
	if err == github.ErrIncompleteResults {
		logger.Info("server.external-service-sync", log.Error(err))
		syncResult := &protocol.ExternalServiceSyncResult{
			Error: err.Error(),
		}
		s.respond(w, http.StatusOK, syncResult)
		return
	} else if ctx.Err() != nil {
		// client is gone
		return
	} else if err != nil {
		logger.Error("server.external-service-sync", log.Error(err))
		if errcode.IsUnauthorized(err) {
			s.respond(w, http.StatusUnauthorized, err)
			return
		}
		if errcode.IsForbidden(err) {
			s.respond(w, http.StatusForbidden, err)
			return
		}
		s.respond(w, http.StatusInternalServerError, err)
		return
	}

	if s.RateLimitSyncer != nil {
		err = s.RateLimitSyncer.SyncRateLimiters(ctx, req.ExternalServiceID)
		if err != nil {
			logger.Warn("Handling rate limiter sync", log.Error(err))
		}
	}

	if err := s.Syncer.TriggerExternalServiceSync(ctx, req.ExternalServiceID); err != nil {
		logger.Warn("Enqueueing external service sync job", log.Error(err))
	}

	logger.Info("server.external-service-sync", log.Bool("synced", true))
	s.respond(w, http.StatusOK, &protocol.ExternalServiceSyncResult{})
}

func externalServiceValidate(ctx context.Context, es *types.ExternalService, src repos.Source) error {
	if !es.DeletedAt.IsZero() {
		// We don't need to check deleted services.
		return nil
	}

	if v, ok := src.(repos.UserSource); ok {
		return v.ValidateAuthenticator(ctx)
	}

	ctx, cancel := context.WithCancel(ctx)
	results := make(chan repos.SourceResult)

	defer func() {
		cancel()

		// We need to drain the rest of the results to not leak a blocked goroutine.
		for range results {
		}
	}()

	go func() {
		src.ListRepos(ctx, results)
		close(results)
	}()

	select {
	case res := <-results:
		// As soon as we get the first result back, we've got what we need to validate the external service.
		return res.Err
	case <-ctx.Done():
		return ctx.Err()
	}
}

var mockRepoLookup func(protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error)

func (s *Server) repoLookup(ctx context.Context, args protocol.RepoLookupArgs) (result *protocol.RepoLookupResult, err error) {
	// Sourcegraph.com: this is on the user path, do not block forever if codehost is
	// being bad. Ideally block before cloudflare 504s the request (1min). Other: we
	// only speak to our database, so response should be in a few ms.
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tr, ctx := trace.New(ctx, "repoLookup", args.String())
	defer func() {
		s.Logger.Debug("repoLookup", log.String("result", fmt.Sprint(result)), log.Error(err))
		if result != nil {
			tr.LazyPrintf("result: %s", result)
		}
		tr.SetError(err)
		tr.Finish()
	}()

	if args.Repo == "" {
		return nil, errors.New("Repo must be set (is blank)")
	}

	if mockRepoLookup != nil {
		return mockRepoLookup(args)
	}

	repo, err := s.Syncer.SyncRepo(ctx, args.Repo, true)

	switch {
	case err == nil:
		break
	case errcode.IsNotFound(err):
		return &protocol.RepoLookupResult{ErrorNotFound: true}, nil
	case errcode.IsUnauthorized(err) || errcode.IsForbidden(err):
		return &protocol.RepoLookupResult{ErrorUnauthorized: true}, nil
	case errcode.IsTemporary(err):
		return &protocol.RepoLookupResult{ErrorTemporarilyUnavailable: true}, nil
	default:
		return nil, err
	}

	if s.Scheduler != nil && args.Update {
		// Enqueue a high priority update for this repo.
		s.Scheduler.UpdateOnce(repo.ID, repo.Name)
	}

	repoInfo := protocol.NewRepoInfo(repo)

	return &protocol.RepoLookupResult{Repo: repoInfo}, nil
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

func (s *Server) handleSchedulePermsSync(w http.ResponseWriter, r *http.Request) {
	if s.PermsSyncer == nil {
		s.respond(w, http.StatusForbidden, nil)
		return
	}

	var req protocol.PermsSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respond(w, http.StatusBadRequest, err)
		return
	}
	if len(req.UserIDs) == 0 && len(req.RepoIDs) == 0 {
		s.respond(w, http.StatusBadRequest, errors.New("neither user IDs nor repo IDs was provided in request (must provide at least one)"))
		return
	}

	s.PermsSyncer.ScheduleUsers(r.Context(), req.Options, req.UserIDs...)
	s.PermsSyncer.ScheduleRepos(r.Context(), req.RepoIDs...)

	s.respond(w, http.StatusOK, nil)
}
