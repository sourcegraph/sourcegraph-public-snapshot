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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/batches"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	proto "github.com/sourcegraph/sourcegraph/internal/repoupdater/v1"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
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
	ChangesetSyncRegistry batches.ChangesetSyncRegistry
	RateLimitSyncer       interface {
		// SyncRateLimiters should be called when an external service changes so that
		// our internal rate limiters are kept in sync
		SyncRateLimiters(ctx context.Context, ids ...int64) error
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
	mux.HandleFunc("/external-service-namespaces", trace.WithRouteName("external-service-namespaces", s.handleExternalServiceNamespaces))
	mux.HandleFunc("/external-service-repositories", trace.WithRouteName("external-service-repositories", s.handleExternalServiceRepositories))
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
	tr, ctx := trace.New(ctx, "enqueueRepoUpdate", req.String())
	defer func() {
		s.Logger.Debug("enqueueRepoUpdate", log.Object("http", log.Int("status", httpStatus), log.String("resp", fmt.Sprint(resp)), log.Error(err)))
		if resp != nil {
			tr.SetAttributes(
				attribute.Int("resp.id", int(resp.ID)),
				attribute.String("resp.name", resp.Name),
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
	logger := s.Logger.With(log.Int64("ExternalServiceID", req.ExternalServiceID))

	externalServiceID := req.ExternalServiceID

	es, err := s.ExternalServiceStore().GetByID(ctx, externalServiceID)
	if err != nil {
		if errcode.IsNotFound(err) {
			s.respond(w, http.StatusNotFound, err)
		} else {
			s.respond(w, http.StatusInternalServerError, err)
		}
		return
	}

	genericSourcer := s.NewGenericSourcer(logger)
	genericSrc, err := genericSourcer(ctx, es)
	if err != nil {
		logger.Error("server.external-service-sync", log.Error(err))
		return
	}

	// sync the rate limit first, because externalServiceValidate potentially
	// makes a call to the code host, which might be rate limited
	if s.RateLimitSyncer != nil {
		err = s.RateLimitSyncer.SyncRateLimiters(ctx, req.ExternalServiceID)
		if err != nil {
			logger.Warn("Handling rate limiter sync", log.Error(err))
		}
	}

	statusCode, resp := handleExternalServiceValidate(ctx, logger, es, genericSrc)
	if statusCode > 0 {
		s.respond(w, statusCode, resp)
		return
	}
	if statusCode == 0 {
		// client is gone
		return
	}

	if err := s.Syncer.TriggerExternalServiceSync(ctx, req.ExternalServiceID); err != nil {
		logger.Warn("Enqueueing external service sync job", log.Error(err))
	}

	logger.Info("server.external-service-sync", log.Bool("synced", true))
	s.respond(w, http.StatusOK, &protocol.ExternalServiceSyncResult{})
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

func handleExternalServiceValidate(ctx context.Context, logger log.Logger, es *types.ExternalService, src repos.Source) (int, any) {
	err := externalServiceValidate(ctx, es, src)
	if err == github.ErrIncompleteResults {
		logger.Info("server.external-service-sync", log.Error(err))
		syncResult := &protocol.ExternalServiceSyncResult{
			Error: err.Error(),
		}
		return http.StatusOK, syncResult
	}
	if ctx.Err() != nil {
		// client is gone
		return 0, nil
	}
	if err != nil {
		logger.Error("server.external-service-sync", log.Error(err))
		if errcode.IsUnauthorized(err) {
			return http.StatusUnauthorized, err
		}
		if errcode.IsForbidden(err) {
			return http.StatusForbidden, err
		}
		return http.StatusInternalServerError, err
	}
	return -1, nil
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

func (s *Server) handleExternalServiceNamespaces(w http.ResponseWriter, r *http.Request) {
	var req protocol.ExternalServiceNamespacesArgs
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger := s.Logger.With(log.String("ExternalServiceKind", req.Kind))

	result, err := s.externalServiceNamespaces(r.Context(), logger, req.ToProto())
	if err != nil {
		logger.Error("server.query-external-service-namespaces", log.Error(err))
		httpCode := grpcErrToStatus(err)
		s.respond(w, httpCode, &protocol.ExternalServiceNamespacesResult{Error: err.Error()})
		return
	}
	s.respond(w, http.StatusOK, protocol.ExternalServiceNamespacesResultFromProto(result))
}

func (s *Server) externalServiceNamespaces(ctx context.Context, logger log.Logger, req *proto.ExternalServiceNamespacesRequest) (*proto.ExternalServiceNamespacesResponse, error) {
	var externalSvc *types.ExternalService
	if req.ExternalServiceId != nil {
		var err error
		externalSvc, err = s.ExternalServiceStore().GetByID(ctx, *req.ExternalServiceId)
		if err != nil {
			if errcode.IsNotFound(err) {
				return nil, status.Error(codes.NotFound, err.Error())
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		externalSvc = &types.ExternalService{
			Kind:   req.Kind,
			Config: extsvc.NewUnencryptedConfig(req.Config),
		}
	}

	genericSourcer := s.NewGenericSourcer(logger)
	genericSrc, err := genericSourcer(ctx, externalSvc)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err = genericSrc.CheckConnection(ctx); err != nil {
		if errcode.IsUnauthorized(err) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Unavailable, err.Error())
	}

	discoverableSrc, ok := genericSrc.(repos.DiscoverableSource)
	if !ok {
		return nil, status.Error(codes.Unimplemented, repos.UnimplementedDiscoverySource)
	}

	results := make(chan repos.SourceNamespaceResult)
	go func() {
		discoverableSrc.ListNamespaces(ctx, results)
		close(results)
	}()

	var sourceErrs error
	namespaces := make([]*proto.ExternalServiceNamespace, 0)

	for res := range results {
		if res.Err != nil {
			sourceErrs = errors.Append(sourceErrs, &repos.SourceError{Err: res.Err, ExtSvc: externalSvc})
			continue
		}
		namespaces = append(namespaces, &proto.ExternalServiceNamespace{
			Id:         int64(res.Namespace.ID),
			Name:       res.Namespace.Name,
			ExternalId: res.Namespace.ExternalID,
		})
	}

	return &proto.ExternalServiceNamespacesResponse{Namespaces: namespaces}, sourceErrs
}

func (s *Server) handleExternalServiceRepositories(w http.ResponseWriter, r *http.Request) {
	var req protocol.ExternalServiceRepositoriesArgs
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger := s.Logger.With(log.String("ExternalServiceKind", req.Kind))

	result, err := s.externalServiceRepositories(r.Context(), logger, req.ToProto())
	if err != nil {
		logger.Error("server.query-external-service-repositories", log.Error(err))
		httpCode := grpcErrToStatus(err)
		s.respond(w, httpCode, &protocol.ExternalServiceRepositoriesResult{Error: err.Error()})
		return
	}
	s.respond(w, http.StatusOK, protocol.ExternalServiceRepositoriesResultFromProto(result))
}

func (s *Server) externalServiceRepositories(ctx context.Context, logger log.Logger, req *proto.ExternalServiceRepositoriesRequest) (*proto.ExternalServiceRepositoriesResponse, error) {
	var externalSvc *types.ExternalService
	if req.ExternalServiceId != nil {
		var err error
		externalSvc, err = s.ExternalServiceStore().GetByID(ctx, *req.ExternalServiceId)
		if err != nil {
			if errcode.IsNotFound(err) {
				return nil, status.Error(codes.NotFound, err.Error())
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		externalSvc = &types.ExternalService{
			Kind:   req.Kind,
			Config: extsvc.NewUnencryptedConfig(req.Config),
		}
	}

	genericSourcer := s.NewGenericSourcer(logger)
	genericSrc, err := genericSourcer(ctx, externalSvc)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err = genericSrc.CheckConnection(ctx); err != nil {
		if errcode.IsUnauthorized(err) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Unavailable, err.Error())
	}

	discoverableSrc, ok := genericSrc.(repos.DiscoverableSource)
	if !ok {
		return nil, status.Error(codes.Unimplemented, repos.UnimplementedDiscoverySource)
	}

	results := make(chan repos.SourceResult)

	first := int(req.First)
	if first > 100 {
		first = 100
	}

	go func() {
		discoverableSrc.SearchRepositories(ctx, req.Query, first, req.GetExcludeRepos(), results)
		close(results)
	}()

	var sourceErrs error
	repositories := make([]*proto.ExternalServiceRepository, 0)

	for res := range results {
		if res.Err != nil {
			sourceErrs = errors.Append(sourceErrs, &repos.SourceError{Err: res.Err, ExtSvc: externalSvc})
			continue
		}
		repositories = append(repositories, &proto.ExternalServiceRepository{
			Id:         int32(res.Repo.ID),
			Name:       string(res.Repo.Name),
			ExternalId: res.Repo.ExternalRepo.ID,
		})
	}

	return &proto.ExternalServiceRepositoriesResponse{Repos: repositories}, sourceErrs
}

// grpcErrToStatus translates the grpc status codes used in this package to http status codes.
func grpcErrToStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}

	s, ok := status.FromError(err)
	if !ok {
		// we deliberately make context.Canceled and context.DeadlineExceeded return 500
		return http.StatusInternalServerError
	}

	switch s.Code() {
	case codes.NotFound:
		return http.StatusNotFound
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.PermissionDenied:
		return http.StatusUnauthorized
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.Unimplemented:
		return http.StatusNotImplemented
	default:
		return http.StatusInternalServerError
	}
}

var mockNewGenericSourcer func() repos.Sourcer

func (s *Server) NewGenericSourcer(logger log.Logger) repos.Sourcer {
	if mockNewGenericSourcer != nil {
		return mockNewGenericSourcer()
	}

	// We use the generic sourcer that doesn't have observability attached to it here because the way externalServiceValidate is set up,
	// using the regular sourcer will cause a large dump of errors to be logged when it exits ListRepos prematurely.
	sourcerLogger := logger.Scoped("repos.Sourcer", "repositories source")
	db := database.NewDBWith(sourcerLogger.Scoped("db", "sourcer database"), s)
	dependenciesService := dependencies.NewService(s.ObservationCtx, db)
	cf := httpcli.NewExternalClientFactory(httpcli.NewLoggingMiddleware(sourcerLogger))
	return repos.NewSourcer(sourcerLogger, db, cf, repos.WithDependenciesService(dependenciesService))
}
