package gitserver

import (
	"context"

	"google.golang.org/grpc"

	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
)

type MockGRPCClient struct {
	MockBatchLog                    func(ctx context.Context, in *proto.BatchLogRequest, opts ...grpc.CallOption) (*proto.BatchLogResponse, error)
	MockCreateCommitFromPatchBinary func(ctx context.Context, opts ...grpc.CallOption) (proto.GitserverService_CreateCommitFromPatchBinaryClient, error)
	MockDiskInfo                    func(ctx context.Context, in *proto.DiskInfoRequest, opts ...grpc.CallOption) (*proto.DiskInfoResponse, error)
	MockExec                        func(ctx context.Context, in *proto.ExecRequest, opts ...grpc.CallOption) (proto.GitserverService_ExecClient, error)
	MockGetObject                   func(ctx context.Context, in *proto.GetObjectRequest, opts ...grpc.CallOption) (*proto.GetObjectResponse, error)
	MockIsRepoCloneable             func(ctx context.Context, in *proto.IsRepoCloneableRequest, opts ...grpc.CallOption) (*proto.IsRepoCloneableResponse, error)
	MockListGitolite                func(ctx context.Context, in *proto.ListGitoliteRequest, opts ...grpc.CallOption) (*proto.ListGitoliteResponse, error)
	MockRepoClone                   func(ctx context.Context, in *proto.RepoCloneRequest, opts ...grpc.CallOption) (*proto.RepoCloneResponse, error)
	MockRepoCloneProgress           func(ctx context.Context, in *proto.RepoCloneProgressRequest, opts ...grpc.CallOption) (*proto.RepoCloneProgressResponse, error)
	MockRepoDelete                  func(ctx context.Context, in *proto.RepoDeleteRequest, opts ...grpc.CallOption) (*proto.RepoDeleteResponse, error)
	MockRepoUpdate                  func(ctx context.Context, in *proto.RepoUpdateRequest, opts ...grpc.CallOption) (*proto.RepoUpdateResponse, error)
	MockArchive                     func(ctx context.Context, in *proto.ArchiveRequest, opts ...grpc.CallOption) (proto.GitserverService_ArchiveClient, error)
	MockSearch                      func(ctx context.Context, in *proto.SearchRequest, opts ...grpc.CallOption) (proto.GitserverService_SearchClient, error)
	MockP4Exec                      func(ctx context.Context, in *proto.P4ExecRequest, opts ...grpc.CallOption) (proto.GitserverService_P4ExecClient, error)
	MockIsPerforcePathCloneable     func(ctx context.Context, in *proto.IsPerforcePathCloneableRequest, opts ...grpc.CallOption) (*proto.IsPerforcePathCloneableResponse, error)
	MockCheckPerforceCredentials    func(ctx context.Context, in *proto.CheckPerforceCredentialsRequest, opts ...grpc.CallOption) (*proto.CheckPerforceCredentialsResponse, error)
}

// BatchLog implements v1.GitserverServiceClient.
func (mc *MockGRPCClient) BatchLog(ctx context.Context, in *proto.BatchLogRequest, opts ...grpc.CallOption) (*proto.BatchLogResponse, error) {
	return mc.MockBatchLog(ctx, in, opts...)
}

// DiskInfo implements v1.GitserverServiceClient.
func (mc *MockGRPCClient) DiskInfo(ctx context.Context, in *proto.DiskInfoRequest, opts ...grpc.CallOption) (*proto.DiskInfoResponse, error) {
	return mc.MockDiskInfo(ctx, in, opts...)
}

// GetObject implements v1.GitserverServiceClient.
func (mc *MockGRPCClient) GetObject(ctx context.Context, in *proto.GetObjectRequest, opts ...grpc.CallOption) (*proto.GetObjectResponse, error) {
	return mc.MockGetObject(ctx, in, opts...)
}

// ListGitolite implements v1.GitserverServiceClient.
func (mc *MockGRPCClient) ListGitolite(ctx context.Context, in *proto.ListGitoliteRequest, opts ...grpc.CallOption) (*proto.ListGitoliteResponse, error) {
	return mc.MockListGitolite(ctx, in, opts...)
}

// P4Exec implements v1.GitserverServiceClient.
func (mc *MockGRPCClient) P4Exec(ctx context.Context, in *proto.P4ExecRequest, opts ...grpc.CallOption) (proto.GitserverService_P4ExecClient, error) {
	return mc.MockP4Exec(ctx, in, opts...)
}

// CreateCommitFromPatchBinary implements v1.GitserverServiceClient.
func (mc *MockGRPCClient) CreateCommitFromPatchBinary(ctx context.Context, opts ...grpc.CallOption) (proto.GitserverService_CreateCommitFromPatchBinaryClient, error) {
	return mc.MockCreateCommitFromPatchBinary(ctx, opts...)
}

// RepoUpdate implements v1.GitserverServiceClient
func (mc *MockGRPCClient) RepoUpdate(ctx context.Context, in *proto.RepoUpdateRequest, opts ...grpc.CallOption) (*proto.RepoUpdateResponse, error) {
	return mc.MockRepoUpdate(ctx, in, opts...)
}

// RepoDelete implements v1.GitserverServiceClient
func (mc *MockGRPCClient) RepoDelete(ctx context.Context, in *proto.RepoDeleteRequest, opts ...grpc.CallOption) (*proto.RepoDeleteResponse, error) {
	return mc.MockRepoDelete(ctx, in, opts...)
}

// RepoCloneProgress implements v1.GitserverServiceClient
func (mc *MockGRPCClient) RepoCloneProgress(ctx context.Context, in *proto.RepoCloneProgressRequest, opts ...grpc.CallOption) (*proto.RepoCloneProgressResponse, error) {
	return mc.MockRepoCloneProgress(ctx, in, opts...)
}

// Exec implements v1.GitserverServiceClient
func (mc *MockGRPCClient) Exec(ctx context.Context, in *proto.ExecRequest, opts ...grpc.CallOption) (proto.GitserverService_ExecClient, error) {
	return mc.MockExec(ctx, in, opts...)
}

// RepoClone implements v1.GitserverServiceClient
func (mc *MockGRPCClient) RepoClone(ctx context.Context, in *proto.RepoCloneRequest, opts ...grpc.CallOption) (*proto.RepoCloneResponse, error) {
	return mc.MockRepoClone(ctx, in, opts...)
}

func (ms *MockGRPCClient) IsRepoCloneable(ctx context.Context, in *proto.IsRepoCloneableRequest, opts ...grpc.CallOption) (*proto.IsRepoCloneableResponse, error) {
	return ms.MockIsRepoCloneable(ctx, in, opts...)
}

// Search implements v1.GitserverServiceClient
func (ms *MockGRPCClient) Search(ctx context.Context, in *proto.SearchRequest, opts ...grpc.CallOption) (proto.GitserverService_SearchClient, error) {
	return ms.MockSearch(ctx, in, opts...)
}

func (mc *MockGRPCClient) Archive(ctx context.Context, in *proto.ArchiveRequest, opts ...grpc.CallOption) (proto.GitserverService_ArchiveClient, error) {
	return mc.MockArchive(ctx, in, opts...)
}

func (mc *MockGRPCClient) IsPerforcePathCloneable(ctx context.Context, in *proto.IsPerforcePathCloneableRequest, opts ...grpc.CallOption) (*proto.IsPerforcePathCloneableResponse, error) {
	return mc.MockIsPerforcePathCloneable(ctx, in, opts...)
}

func (mc *MockGRPCClient) CheckPerforceCredentials(ctx context.Context, in *proto.CheckPerforceCredentialsRequest, opts ...grpc.CallOption) (*proto.CheckPerforceCredentialsResponse, error) {
	return mc.MockCheckPerforceCredentials(ctx, in, opts...)
}

var _ proto.GitserverServiceClient = &MockGRPCClient{}
