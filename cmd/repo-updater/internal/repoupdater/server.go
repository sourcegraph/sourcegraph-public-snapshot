package repoupdater

import (
	"context"
	"fmt"

	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/batches/syncer"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	proto "github.com/sourcegraph/sourcegraph/internal/repoupdater/v1"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Scheduler interface {
	UpdateOnce(id api.RepoID, name api.RepoName)
	ScheduleInfo(id api.RepoID) *protocol.RepoUpdateSchedulerInfoResult
}

// Server is a repoupdater server.
type Server struct {
	proto.UnimplementedRepoUpdaterServiceServer

	Store                 repos.Store
	Syncer                *repos.Syncer
	Logger                log.Logger
	Scheduler             Scheduler
	ChangesetSyncRegistry syncer.ChangesetSyncRegistry
}

func (s *Server) RepoUpdateSchedulerInfo(_ context.Context, req *proto.RepoUpdateSchedulerInfoRequest) (*proto.RepoUpdateSchedulerInfoResponse, error) {
	res := s.Scheduler.ScheduleInfo(api.RepoID(req.GetId()))
	return res.ToProto(), nil
}

func (s *Server) EnqueueRepoUpdate(ctx context.Context, req *proto.EnqueueRepoUpdateRequest) (resp *proto.EnqueueRepoUpdateResponse, err error) {
	// NOTE: Internal actor is required to have full visibility of the repo table
	// 	(i.e. bypass repository authorization).
	ctx = actor.WithInternalActor(ctx)

	tr, ctx := trace.New(ctx, "enqueueRepoUpdate", attribute.Stringer("req", req))
	defer func() {
		s.Logger.Debug("enqueueRepoUpdate", log.Object("http", log.String("resp", fmt.Sprint(resp)), log.Error(err)))
		if resp != nil {
			tr.SetAttributes(
				attribute.Int("resp.id", int(resp.Id)),
				attribute.String("resp.name", resp.Name),
			)
		}
		tr.SetError(err)
		tr.End()
	}()

	rs, err := s.Store.RepoStore().List(ctx, database.ReposListOptions{Names: []string{string(req.Repo)}})
	if err != nil {
		return nil, errors.Wrap(err, "store.list-repos")
	}

	if len(rs) != 1 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("repo %q not found in store", req.Repo))
	}

	repo := rs[0]

	s.Scheduler.UpdateOnce(repo.ID, repo.Name)

	return &proto.EnqueueRepoUpdateResponse{
		Id:   int32(repo.ID),
		Name: string(repo.Name),
	}, nil
}

func (s *Server) RecloneRepository(ctx context.Context, req *proto.RecloneRepositoryRequest) (*proto.RecloneRepositoryResponse, error) {
	// NOTE: Internal actor is required to have full visibility of the repo table
	// 	(i.e. bypass repository authorization).
	ctx = actor.WithInternalActor(ctx)

	repoName := api.RepoName(req.GetRepoName())
	if repoName == "" {
		return nil, status.Error(codes.InvalidArgument, "repo_name must be specified")
	}

	rs, err := s.Store.RepoStore().List(ctx, database.ReposListOptions{Names: []string{string(repoName)}})
	if err != nil {
		return nil, errors.Wrap(err, "store.list-repos")
	}

	if len(rs) != 1 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("repo %q not found in store", repoName))
	}

	repo := rs[0]

	svc := gitserver.NewRepositoryServiceClient("repoupdater.reclone")

	if err := svc.DeleteRepository(ctx, repoName); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete repository %q: %s", repoName, err))
	}

	// Enqueue a reclone through scheduler.
	s.Scheduler.UpdateOnce(repo.ID, repo.Name)

	return &proto.RecloneRepositoryResponse{}, nil
}

func (s *Server) EnqueueChangesetSync(ctx context.Context, req *proto.EnqueueChangesetSyncRequest) (*proto.EnqueueChangesetSyncResponse, error) {
	// NOTE: Internal actor is required to have full visibility of the repo table
	// 	(i.e. bypass repository authorization).
	ctx = actor.WithInternalActor(ctx)

	if s.ChangesetSyncRegistry == nil {
		s.Logger.Warn("ChangesetSyncer is nil")
		return nil, status.Error(codes.Internal, "changeset syncer is not configured")
	}

	if len(req.Ids) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no ids provided")
	}

	return &proto.EnqueueChangesetSyncResponse{}, s.ChangesetSyncRegistry.EnqueueChangesetSyncs(ctx, req.Ids)
}
