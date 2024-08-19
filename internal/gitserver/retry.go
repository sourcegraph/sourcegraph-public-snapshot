package gitserver

import (
	"context"

	"google.golang.org/grpc"

	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
)

// automaticRetryClient is a convenience wrapper around a base proto.GitserverServiceClient that automatically retries
// idempotent ("safe") methods in accordance to the policy defined in internal/grpc/defaults.RetryPolicy.
//
// Read the implementation of this type for more details are automatically retried (and why).
//
// Callers are free to override the default retry behavior by proving their own grpc.CallOptions when invoking the RPC.
// (example: providing retry.WithMax(0) will disable retries even when invoking DiskInfo - which is idempotent).
type automaticRetryClient struct {
	base proto.GitserverServiceClient
}

// Non-idempotent methods.

func (r *automaticRetryClient) CreateCommitFromPatchBinary(ctx context.Context, opts ...grpc.CallOption) (proto.GitserverService_CreateCommitFromPatchBinaryClient, error) {
	// CreateCommitFromPatchBinary isn't idempotent. It also is a client-streaming method, which is currently unsupported by our automatic retry logic.
	// The caller is responsible for implementing their own retry semantics for this method.
	return r.base.CreateCommitFromPatchBinary(ctx, opts...)
}

// Idempotent methods.

func (r *automaticRetryClient) DiskInfo(ctx context.Context, in *proto.DiskInfoRequest, opts ...grpc.CallOption) (*proto.DiskInfoResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.DiskInfo(ctx, in, opts...)
}

func (r *automaticRetryClient) GetObject(ctx context.Context, in *proto.GetObjectRequest, opts ...grpc.CallOption) (*proto.GetObjectResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.GetObject(ctx, in, opts...)
}

func (r *automaticRetryClient) IsRepoCloneable(ctx context.Context, in *proto.IsRepoCloneableRequest, opts ...grpc.CallOption) (*proto.IsRepoCloneableResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.IsRepoCloneable(ctx, in, opts...)
}

func (r *automaticRetryClient) ListGitolite(ctx context.Context, in *proto.ListGitoliteRequest, opts ...grpc.CallOption) (*proto.ListGitoliteResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.ListGitolite(ctx, in, opts...)
}

func (r *automaticRetryClient) Search(ctx context.Context, in *proto.SearchRequest, opts ...grpc.CallOption) (proto.GitserverService_SearchClient, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.Search(ctx, in, opts...)
}

func (r *automaticRetryClient) Archive(ctx context.Context, in *proto.ArchiveRequest, opts ...grpc.CallOption) (proto.GitserverService_ArchiveClient, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.Archive(ctx, in, opts...)
}

func (r *automaticRetryClient) RepoCloneProgress(ctx context.Context, in *proto.RepoCloneProgressRequest, opts ...grpc.CallOption) (*proto.RepoCloneProgressResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.RepoCloneProgress(ctx, in, opts...)
}

func (r *automaticRetryClient) IsPerforcePathCloneable(ctx context.Context, in *proto.IsPerforcePathCloneableRequest, opts ...grpc.CallOption) (*proto.IsPerforcePathCloneableResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.IsPerforcePathCloneable(ctx, in, opts...)
}

func (r *automaticRetryClient) CheckPerforceCredentials(ctx context.Context, in *proto.CheckPerforceCredentialsRequest, opts ...grpc.CallOption) (*proto.CheckPerforceCredentialsResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.CheckPerforceCredentials(ctx, in, opts...)
}

func (r *automaticRetryClient) PerforceUsers(ctx context.Context, in *proto.PerforceUsersRequest, opts ...grpc.CallOption) (*proto.PerforceUsersResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.PerforceUsers(ctx, in, opts...)
}

func (r *automaticRetryClient) PerforceProtectsForUser(ctx context.Context, in *proto.PerforceProtectsForUserRequest, opts ...grpc.CallOption) (*proto.PerforceProtectsForUserResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.PerforceProtectsForUser(ctx, in, opts...)
}

func (r *automaticRetryClient) PerforceProtectsForDepot(ctx context.Context, in *proto.PerforceProtectsForDepotRequest, opts ...grpc.CallOption) (*proto.PerforceProtectsForDepotResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.PerforceProtectsForDepot(ctx, in, opts...)
}

func (r *automaticRetryClient) PerforceGroupMembers(ctx context.Context, in *proto.PerforceGroupMembersRequest, opts ...grpc.CallOption) (*proto.PerforceGroupMembersResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.PerforceGroupMembers(ctx, in, opts...)
}

func (r *automaticRetryClient) IsPerforceSuperUser(ctx context.Context, in *proto.IsPerforceSuperUserRequest, opts ...grpc.CallOption) (*proto.IsPerforceSuperUserResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.IsPerforceSuperUser(ctx, in, opts...)
}

func (r *automaticRetryClient) PerforceGetChangelist(ctx context.Context, in *proto.PerforceGetChangelistRequest, opts ...grpc.CallOption) (*proto.PerforceGetChangelistResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.PerforceGetChangelist(ctx, in, opts...)
}

func (r *automaticRetryClient) MergeBase(ctx context.Context, in *proto.MergeBaseRequest, opts ...grpc.CallOption) (*proto.MergeBaseResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.MergeBase(ctx, in, opts...)
}

func (r *automaticRetryClient) Blame(ctx context.Context, in *proto.BlameRequest, opts ...grpc.CallOption) (proto.GitserverService_BlameClient, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.Blame(ctx, in, opts...)
}

func (r *automaticRetryClient) DefaultBranch(ctx context.Context, in *proto.DefaultBranchRequest, opts ...grpc.CallOption) (*proto.DefaultBranchResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.DefaultBranch(ctx, in, opts...)
}

func (r *automaticRetryClient) ReadFile(ctx context.Context, in *proto.ReadFileRequest, opts ...grpc.CallOption) (proto.GitserverService_ReadFileClient, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.ReadFile(ctx, in, opts...)
}

func (r *automaticRetryClient) ListRefs(ctx context.Context, in *proto.ListRefsRequest, opts ...grpc.CallOption) (proto.GitserverService_ListRefsClient, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.ListRefs(ctx, in, opts...)
}

func (r *automaticRetryClient) GetCommit(ctx context.Context, in *proto.GetCommitRequest, opts ...grpc.CallOption) (*proto.GetCommitResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.GetCommit(ctx, in, opts...)
}

func (r *automaticRetryClient) ResolveRevision(ctx context.Context, in *proto.ResolveRevisionRequest, opts ...grpc.CallOption) (*proto.ResolveRevisionResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.ResolveRevision(ctx, in, opts...)
}

func (r *automaticRetryClient) RevAtTime(ctx context.Context, in *proto.RevAtTimeRequest, opts ...grpc.CallOption) (*proto.RevAtTimeResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.RevAtTime(ctx, in, opts...)
}

func (r *automaticRetryClient) RawDiff(ctx context.Context, in *proto.RawDiffRequest, opts ...grpc.CallOption) (proto.GitserverService_RawDiffClient, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.RawDiff(ctx, in, opts...)
}

func (r *automaticRetryClient) ContributorCounts(ctx context.Context, in *proto.ContributorCountsRequest, opts ...grpc.CallOption) (*proto.ContributorCountsResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.ContributorCounts(ctx, in, opts...)
}

func (r *automaticRetryClient) FirstEverCommit(ctx context.Context, in *proto.FirstEverCommitRequest, opts ...grpc.CallOption) (*proto.FirstEverCommitResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.FirstEverCommit(ctx, in, opts...)
}

func (r *automaticRetryClient) BehindAhead(ctx context.Context, in *proto.BehindAheadRequest, opts ...grpc.CallOption) (*proto.BehindAheadResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.BehindAhead(ctx, in, opts...)
}

func (r *automaticRetryClient) ChangedFiles(ctx context.Context, in *proto.ChangedFilesRequest, opts ...grpc.CallOption) (proto.GitserverService_ChangedFilesClient, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.ChangedFiles(ctx, in, opts...)
}

func (r *automaticRetryClient) Stat(ctx context.Context, in *proto.StatRequest, opts ...grpc.CallOption) (*proto.StatResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.Stat(ctx, in, opts...)
}

func (r *automaticRetryClient) ReadDir(ctx context.Context, in *proto.ReadDirRequest, opts ...grpc.CallOption) (proto.GitserverService_ReadDirClient, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.ReadDir(ctx, in, opts...)
}

func (r *automaticRetryClient) CommitLog(ctx context.Context, in *proto.CommitLogRequest, opts ...grpc.CallOption) (proto.GitserverService_CommitLogClient, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.CommitLog(ctx, in, opts...)
}

func (r *automaticRetryClient) MergeBaseOctopus(ctx context.Context, in *proto.MergeBaseOctopusRequest, opts ...grpc.CallOption) (*proto.MergeBaseOctopusResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return r.base.MergeBaseOctopus(ctx, in, opts...)
}

var _ proto.GitserverServiceClient = &automaticRetryClient{}
