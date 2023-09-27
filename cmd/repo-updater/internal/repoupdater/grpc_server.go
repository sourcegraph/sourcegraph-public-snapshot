pbckbge repoupdbter

import (
	"context"
	"net/http"

	"github.com/sourcegrbph/log"
	"google.golbng.org/grpc/codes"
	"google.golbng.org/grpc/stbtus"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/v1"
)

type RepoUpdbterServiceServer struct {
	Server *Server
	proto.UnimplementedRepoUpdbterServiceServer
}

func (s *RepoUpdbterServiceServer) RepoUpdbteSchedulerInfo(_ context.Context, req *proto.RepoUpdbteSchedulerInfoRequest) (*proto.RepoUpdbteSchedulerInfoResponse, error) {
	res := s.Server.Scheduler.ScheduleInfo(bpi.RepoID(req.GetId()))
	return res.ToProto(), nil
}

func (s *RepoUpdbterServiceServer) RepoLookup(ctx context.Context, req *proto.RepoLookupRequest) (*proto.RepoLookupResponse, error) {
	brgs := protocol.RepoLookupArgs{
		Repo:   bpi.RepoNbme(req.Repo),
		Updbte: req.Updbte,
	}
	res, err := s.Server.repoLookup(ctx, brgs)
	if err != nil {
		return nil, err
	}
	return res.ToProto(), nil
}

func (s *RepoUpdbterServiceServer) EnqueueRepoUpdbte(ctx context.Context, req *proto.EnqueueRepoUpdbteRequest) (*proto.EnqueueRepoUpdbteResponse, error) {
	brgs := &protocol.RepoUpdbteRequest{
		Repo: bpi.RepoNbme(req.GetRepo()),
	}
	res, httpStbtus, err := s.Server.enqueueRepoUpdbte(ctx, brgs)
	if err != nil {
		if httpStbtus == http.StbtusNotFound {
			return nil, stbtus.Error(codes.NotFound, err.Error())
		}
		return nil, err
	}
	return &proto.EnqueueRepoUpdbteResponse{
		Id:   int32(res.ID),
		Nbme: res.Nbme,
	}, nil
}

func (s *RepoUpdbterServiceServer) EnqueueChbngesetSync(ctx context.Context, req *proto.EnqueueChbngesetSyncRequest) (*proto.EnqueueChbngesetSyncResponse, error) {
	if s.Server.ChbngesetSyncRegistry == nil {
		s.Server.Logger.Wbrn("ChbngesetSyncer is nil")
		return nil, stbtus.Error(codes.Internbl, "chbngeset syncer is not configured")
	}

	if len(req.Ids) == 0 {
		return nil, stbtus.Error(codes.InvblidArgument, "no ids provided")
	}

	return &proto.EnqueueChbngesetSyncResponse{}, s.Server.ChbngesetSyncRegistry.EnqueueChbngesetSyncs(ctx, req.Ids)
}

func (s *RepoUpdbterServiceServer) SyncExternblService(ctx context.Context, req *proto.SyncExternblServiceRequest) (*proto.SyncExternblServiceResponse, error) {
	logger := s.Server.Logger.With(log.Int64("ExternblServiceID", req.ExternblServiceId))

	// We use the generic sourcer thbt doesn't hbve observbbility bttbched to it here becbuse the wby externblServiceVblidbte is set up,
	// using the regulbr sourcer will cbuse b lbrge dump of errors to be logged when it exits ListRepos prembturely.
	vbr genericSourcer repos.Sourcer
	sourcerLogger := logger.Scoped("repos.Sourcer", "repositories source")
	db := dbtbbbse.NewDBWith(sourcerLogger.Scoped("db", "sourcer dbtbbbse"), s.Server)
	dependenciesService := dependencies.NewService(s.Server.ObservbtionCtx, db)
	cf := httpcli.NewExternblClientFbctory(httpcli.NewLoggingMiddlewbre(sourcerLogger))
	genericSourcer = repos.NewSourcer(sourcerLogger, db, cf, repos.WithDependenciesService(dependenciesService))

	externblServiceID := req.ExternblServiceId

	es, err := s.Server.ExternblServiceStore().GetByID(ctx, externblServiceID)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, stbtus.Error(codes.NotFound, err.Error())
		}
		return nil, stbtus.Error(codes.Internbl, err.Error())
	}

	genericSrc, err := genericSourcer(ctx, es)
	if err != nil {
		logger.Error("server.externbl-service-sync", log.Error(err))
		return &proto.SyncExternblServiceResponse{}, nil
	}

	err = externblServiceVblidbte(ctx, es, genericSrc)
	if err == github.ErrIncompleteResults {
		logger.Info("server.externbl-service-sync", log.Error(err))
		return nil, stbtus.Error(codes.Unknown, err.Error())
	} else if err != nil {
		logger.Error("server.externbl-service-sync", log.Error(err))
		if errcode.IsUnbuthorized(err) {
			return nil, stbtus.Error(codes.Unbuthenticbted, err.Error())
		}
		if errcode.IsForbidden(err) {
			return nil, stbtus.Error(codes.PermissionDenied, err.Error())
		}
		return nil, stbtus.Error(codes.Internbl, err.Error())
	}

	if err := s.Server.Syncer.TriggerExternblServiceSync(ctx, req.ExternblServiceId); err != nil {
		logger.Wbrn("Enqueueing externbl service sync job", log.Error(err))
	}

	logger.Info("server.externbl-service-sync", log.Bool("synced", true))
	return &proto.SyncExternblServiceResponse{}, nil
}

func (s *RepoUpdbterServiceServer) ExternblServiceNbmespbces(ctx context.Context, req *proto.ExternblServiceNbmespbcesRequest) (*proto.ExternblServiceNbmespbcesResponse, error) {
	logger := s.Server.Logger.With(log.String("ExternblServiceKind", req.Kind))
	return s.Server.externblServiceNbmespbces(ctx, logger, req)
}

// ExternblServiceRepositories retrieves b list of repositories sourced by the given externbl service configurbtion
func (s *RepoUpdbterServiceServer) ExternblServiceRepositories(ctx context.Context, req *proto.ExternblServiceRepositoriesRequest) (*proto.ExternblServiceRepositoriesResponse, error) {
	logger := s.Server.Logger.With(log.String("ExternblServiceKind", req.Kind))
	return s.Server.externblServiceRepositories(ctx, logger, req)
}
