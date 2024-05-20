package internal

import (
	"context"
	"encoding/base64"

	"github.com/sourcegraph/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/gitserverfs"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GRPCRepositoryServiceConfig struct {
	ExhaustiveRequestLoggingEnabled bool
}

func NewRepositoryServiceServer(server *Server, config *GRPCRepositoryServiceConfig) proto.GitserverRepositoryServiceServer {
	var srv proto.GitserverRepositoryServiceServer = &repositoryServiceServer{
		logger:   server.logger,
		db:       server.db,
		hostname: server.hostname,
		svc:      server,
		fs:       server.fs,
	}

	if config.ExhaustiveRequestLoggingEnabled {
		logger := server.logger.Scoped("gRPCRequestLogger")

		srv = &loggingRepositoryServiceServer{
			base:   srv,
			logger: logger,
		}
	}

	return srv
}

type repositoryServiceServer struct {
	logger   log.Logger
	db       database.DB
	hostname string
	fs       gitserverfs.FS
	svc      service

	proto.UnimplementedGitserverRepositoryServiceServer
}

var _ proto.GitserverRepositoryServiceServer = &repositoryServiceServer{}

func (s *repositoryServiceServer) DeleteRepository(ctx context.Context, req *proto.DeleteRepositoryRequest) (*proto.DeleteRepositoryResponse, error) {
	if req.GetRepoName() == "" {
		return nil, status.New(codes.InvalidArgument, "repo_name must be specified").Err()
	}

	repoName := api.RepoName(req.GetRepoName())

	cloned, err := s.fs.RepoCloned(repoName)
	if err != nil {
		return nil, status.New(codes.Internal, "failed to determine clone status").Err()
	}

	if !cloned {
		return nil, newRepoNotFoundError(repoName, false, "")
	}

	err = s.fs.RemoveRepo(repoName)
	if err != nil {
		err = errors.Wrap(err, "removing repo directory")
		s.logger.Error("failed to delete repository", log.String("repo", string(repoName)), log.Error(err))
		return &proto.DeleteRepositoryResponse{}, status.Errorf(codes.Internal, "failed to delete repository %s: %s", repoName, err)
	}

	err = s.db.GitserverRepos().SetCloneStatus(ctx, repoName, types.CloneStatusNotCloned, s.hostname)
	if err != nil {
		err = errors.Wrap(err, "setting clone status after delete")
		return &proto.DeleteRepositoryResponse{}, status.Errorf(codes.Internal, "failed to delete repository %s: %s", repoName, err)
	}

	s.logger.Info("repository deleted", log.String("repo", string(repoName)))

	return &proto.DeleteRepositoryResponse{}, nil
}

func (s *repositoryServiceServer) FetchRepository(ctx context.Context, req *proto.FetchRepositoryRequest) (*proto.FetchRepositoryResponse, error) {
	if req.GetRepoName() == "" {
		return nil, status.New(codes.InvalidArgument, "repo_name must be specified").Err()
	}

	repoName := api.RepoName(req.GetRepoName())

	lastFetched, lastChanged, err := s.svc.FetchRepository(ctx, repoName)
	if err != nil {
		return nil, status.New(codes.Internal, errors.Wrap(err, "failed to fetch repository").Error()).Err()
	}

	return &proto.FetchRepositoryResponse{
		LastFetched: timestamppb.New(lastFetched),
		LastChanged: timestamppb.New(lastChanged),
	}, nil
}

func (s *repositoryServiceServer) ListRepositories(ctx context.Context, req *proto.ListRepositoriesRequest) (*proto.ListRepositoriesResponse, error) {
	if req.GetPageSize() == 0 {
		return nil, status.New(codes.InvalidArgument, "page_size must be > 0").Err()
	}

	token, err := base64.StdEncoding.DecodeString(req.GetPageToken())
	if err != nil {
		return nil, status.New(codes.InvalidArgument, "invalid page_token").Err()
	}

	var nextCursor string
	atStart := false

	repos := make([]*proto.ListRepositoriesResponse_GitRepository, 0, req.GetPageSize())
	err = s.fs.ForEachRepo(func(rn api.RepoName, gd common.GitDir) (done bool) {
		if err := ctx.Err(); err != nil {
			return true
		}

		if !atStart && string(token) != "" && string(gd) != string(token) {
			return false
		}

		atStart = true

		if uint32(len(repos)) > req.GetPageSize()-1 {
			nextCursor = base64.StdEncoding.EncodeToString([]byte(gd))
			return true
		}

		repos = append(repos, &proto.ListRepositoriesResponse_GitRepository{
			Path: []byte(s.fs.CanonicalPath(gd)),
			Name: string(rn),
		})

		return false
	})
	if err != nil {
		return nil, err
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return &proto.ListRepositoriesResponse{
		Repositories:  repos,
		NextPageToken: nextCursor,
	}, nil
}
