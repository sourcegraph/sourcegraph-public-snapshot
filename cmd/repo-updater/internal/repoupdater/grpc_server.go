package repoupdater

import (
	"context"
	"net/http"

	"github.com/sourcegraph/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	proto "github.com/sourcegraph/sourcegraph/internal/repoupdater/v1"
)

type RepoUpdaterServiceServer struct {
	Server *Server
	proto.UnimplementedRepoUpdaterServiceServer
}

func (s *RepoUpdaterServiceServer) RepoUpdateSchedulerInfo(_ context.Context, req *proto.RepoUpdateSchedulerInfoRequest) (*proto.RepoUpdateSchedulerInfoResponse, error) {
	res := s.Server.Scheduler.ScheduleInfo(api.RepoID(req.GetId()))
	return res.ToProto(), nil
}

func (s *RepoUpdaterServiceServer) RepoLookup(ctx context.Context, req *proto.RepoLookupRequest) (*proto.RepoLookupResponse, error) {
	res, err := s.Server.repoLookup(ctx, api.RepoName(req.Repo))
	if err != nil {
		return nil, err
	}
	return res.ToProto(), nil
}

func (s *RepoUpdaterServiceServer) EnqueueRepoUpdate(ctx context.Context, req *proto.EnqueueRepoUpdateRequest) (*proto.EnqueueRepoUpdateResponse, error) {
	args := &protocol.RepoUpdateRequest{
		Repo: api.RepoName(req.GetRepo()),
	}
	res, httpStatus, err := s.Server.enqueueRepoUpdate(ctx, args)
	if err != nil {
		if httpStatus == http.StatusNotFound {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, err
	}
	return &proto.EnqueueRepoUpdateResponse{
		Id:   int32(res.ID),
		Name: res.Name,
	}, nil
}

func (s *RepoUpdaterServiceServer) EnqueueChangesetSync(ctx context.Context, req *proto.EnqueueChangesetSyncRequest) (*proto.EnqueueChangesetSyncResponse, error) {
	if s.Server.ChangesetSyncRegistry == nil {
		s.Server.Logger.Warn("ChangesetSyncer is nil")
		return nil, status.Error(codes.Internal, "changeset syncer is not configured")
	}

	if len(req.Ids) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no ids provided")
	}

	return &proto.EnqueueChangesetSyncResponse{}, s.Server.ChangesetSyncRegistry.EnqueueChangesetSyncs(ctx, req.Ids)
}

func (s *RepoUpdaterServiceServer) SyncExternalService(ctx context.Context, req *proto.SyncExternalServiceRequest) (*proto.SyncExternalServiceResponse, error) {
	logger := s.Server.Logger.With(log.Int64("ExternalServiceID", req.ExternalServiceId))

	// We use the generic sourcer that doesn't have observability attached to it here because the way externalServiceValidate is set up,
	// using the regular sourcer will cause a large dump of errors to be logged when it exits ListRepos prematurely.
	var genericSourcer repos.Sourcer
	sourcerLogger := logger.Scoped("repos.Sourcer", "repositories source")
	db := database.NewDBWith(sourcerLogger.Scoped("db", "sourcer database"), s.Server)
	dependenciesService := dependencies.NewService(s.Server.ObservationCtx, db)
	cf := httpcli.NewExternalClientFactory(httpcli.NewLoggingMiddleware(sourcerLogger))
	genericSourcer = repos.NewSourcer(sourcerLogger, db, cf, repos.WithDependenciesService(dependenciesService))

	externalServiceID := req.ExternalServiceId

	es, err := s.Server.ExternalServiceStore().GetByID(ctx, externalServiceID)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	genericSrc, err := genericSourcer(ctx, es)
	if err != nil {
		logger.Error("server.external-service-sync", log.Error(err))
		return &proto.SyncExternalServiceResponse{}, nil
	}

	err = externalServiceValidate(ctx, es, genericSrc)
	if err == github.ErrIncompleteResults {
		logger.Info("server.external-service-sync", log.Error(err))
		return nil, status.Error(codes.Unknown, err.Error())
	} else if err != nil {
		logger.Error("server.external-service-sync", log.Error(err))
		if errcode.IsUnauthorized(err) {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		if errcode.IsForbidden(err) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := s.Server.Syncer.TriggerExternalServiceSync(ctx, req.ExternalServiceId); err != nil {
		logger.Warn("Enqueueing external service sync job", log.Error(err))
	}

	logger.Info("server.external-service-sync", log.Bool("synced", true))
	return &proto.SyncExternalServiceResponse{}, nil
}

func (s *RepoUpdaterServiceServer) ExternalServiceNamespaces(ctx context.Context, req *proto.ExternalServiceNamespacesRequest) (*proto.ExternalServiceNamespacesResponse, error) {
	logger := s.Server.Logger.With(log.String("ExternalServiceKind", req.Kind))
	return s.Server.externalServiceNamespaces(ctx, logger, req)
}

// ExternalServiceRepositories retrieves a list of repositories sourced by the given external service configuration
func (s *RepoUpdaterServiceServer) ExternalServiceRepositories(ctx context.Context, req *proto.ExternalServiceRepositoriesRequest) (*proto.ExternalServiceRepositoriesResponse, error) {
	logger := s.Server.Logger.With(log.String("ExternalServiceKind", req.Kind))
	return s.Server.externalServiceRepositories(ctx, logger, req)
}
