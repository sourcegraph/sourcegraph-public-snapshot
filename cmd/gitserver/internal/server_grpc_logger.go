package internal

import (
	"context"
	"github.com/sourcegraph/log"
	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"time"
)

type loggingGRPCServer struct {
	base   proto.GitserverServiceServer
	logger log.Logger
}

func (l *loggingGRPCServer) doLog(message string, fields ...log.Field) {
	l.logger.Info(message, fields...)
}

func (l *loggingGRPCServer) CreateCommitFromPatchBinary(server proto.GitserverService_CreateCommitFromPatchBinaryServer) error {
	//TODO implement m
	panic("implement me")
}

func (l *loggingGRPCServer) DiskInfo(ctx context.Context, request *proto.DiskInfoRequest) (response *proto.DiskInfoResponse, err error) {
	start := time.Now()

	defer func() {
		elapsed := time.Since(start)
		fields := []log.Field{
			log.String("method", "DiskInfo"),
			log.String("request", "<empty>"),
			log.String("status", status.Code(err).String()),
			log.Duration("duration", elapsed),
		}

		l.doLog("Received DiskInfo request", fields...)
	}()

	return l.base.DiskInfo(ctx, request)
}

func (l *loggingGRPCServer) Exec(request *proto.ExecRequest, server proto.GitserverService_ExecServer) error {
	//TODO implement me
	panic("implement me")
}

func (l *loggingGRPCServer) GetObject(ctx context.Context, request *proto.GetObjectRequest) (response *proto.GetObjectResponse, err error) {
	start := time.Now()

	defer func() {
		elapsed := time.Since(start)
		fields := []log.Field{
			log.String("method", "DiskInfo"),
			log.String("status", status.Code(err).String()),
			log.String("request", protojson.Format(request)),
			log.Duration("duration", elapsed),
		}

		l.doLog("Received DiskInfo request", fields...)
	}()

	return l.base.GetObject(ctx, request)
}

func (l *loggingGRPCServer) IsRepoCloneable(ctx context.Context, request *proto.IsRepoCloneableRequest) (*proto.IsRepoCloneableResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (l *loggingGRPCServer) ListGitolite(ctx context.Context, request *proto.ListGitoliteRequest) (*proto.ListGitoliteResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (l *loggingGRPCServer) Search(request *proto.SearchRequest, server proto.GitserverService_SearchServer) error {
	//TODO implement me
	panic("implement me")
}

func (l *loggingGRPCServer) Archive(request *proto.ArchiveRequest, server proto.GitserverService_ArchiveServer) error {

}

func (l *loggingGRPCServer) RepoClone(ctx context.Context, request *proto.RepoCloneRequest) (*proto.RepoCloneResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (l *loggingGRPCServer) RepoCloneProgress(ctx context.Context, request *proto.RepoCloneProgressRequest) (*proto.RepoCloneProgressResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (l *loggingGRPCServer) RepoDelete(ctx context.Context, request *proto.RepoDeleteRequest) (*proto.RepoDeleteResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (l *loggingGRPCServer) RepoUpdate(ctx context.Context, request *proto.RepoUpdateRequest) (*proto.RepoUpdateResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (l *loggingGRPCServer) IsPerforcePathCloneable(ctx context.Context, request *proto.IsPerforcePathCloneableRequest) (*proto.IsPerforcePathCloneableResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (l *loggingGRPCServer) CheckPerforceCredentials(ctx context.Context, request *proto.CheckPerforceCredentialsRequest) (*proto.CheckPerforceCredentialsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (l *loggingGRPCServer) PerforceUsers(ctx context.Context, request *proto.PerforceUsersRequest) (*proto.PerforceUsersResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (l *loggingGRPCServer) PerforceProtectsForUser(ctx context.Context, request *proto.PerforceProtectsForUserRequest) (*proto.PerforceProtectsForUserResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (l *loggingGRPCServer) PerforceProtectsForDepot(ctx context.Context, request *proto.PerforceProtectsForDepotRequest) (*proto.PerforceProtectsForDepotResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (l *loggingGRPCServer) PerforceGroupMembers(ctx context.Context, request *proto.PerforceGroupMembersRequest) (*proto.PerforceGroupMembersResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (l *loggingGRPCServer) IsPerforceSuperUser(ctx context.Context, request *proto.IsPerforceSuperUserRequest) (*proto.IsPerforceSuperUserResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (l *loggingGRPCServer) PerforceGetChangelist(ctx context.Context, request *proto.PerforceGetChangelistRequest) (*proto.PerforceGetChangelistResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (l *loggingGRPCServer) MergeBase(ctx context.Context, request *proto.MergeBaseRequest) (*proto.MergeBaseResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (l *loggingGRPCServer) Blame(request *proto.BlameRequest, server proto.GitserverService_BlameServer) error {
	//TODO implement me
	panic("implement me")
}

func (l *loggingGRPCServer) DefaultBranch(ctx context.Context, request *proto.DefaultBranchRequest) (*proto.DefaultBranchResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (l *loggingGRPCServer) ReadFile(request *proto.ReadFileRequest, server proto.GitserverService_ReadFileServer) error {
	//TODO implement me
	panic("implement me")
}

func (l *loggingGRPCServer) GetCommit(ctx context.Context, request *proto.GetCommitRequest) (*proto.GetCommitResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (l *loggingGRPCServer) ResolveRevision(ctx context.Context, request *proto.ResolveRevisionRequest) (*proto.ResolveRevisionResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (l *loggingGRPCServer) mustEmbedUnimplementedGitserverServiceServer() {
	//TODO implement me
	panic("implement me")
}

var _ proto.GitserverServiceServer = &loggingGRPCServer{}
