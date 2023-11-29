package repoupdater

import (
	"context"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/sourcegraph/internal/api"
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
