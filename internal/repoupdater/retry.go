package repoupdater

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	proto "github.com/sourcegraph/sourcegraph/internal/repoupdater/v1"
	"google.golang.org/grpc"
)

// automaticRetryClient is a convenience wrapper around a base proto.RepoUpdaterServiceClient that automatically retries
// idempotent ("safe") methods in accordance to the policy defined in internal/grpc/defaults.RetryPolicy.
//
// Read the implementation of this type for more details are automatically retried (and why).
//
// Callers are free to override the default retry behavior by proving their own grpc.CallOptions when invoking the RPC.
// (example: providing retry.WithMax(0) will disable retries even when invoking RepoLookup - which is idempotent).
type automaticRetryClient struct {
	base proto.RepoUpdaterServiceClient
}

func (a *automaticRetryClient) RepoUpdateSchedulerInfo(ctx context.Context, in *proto.RepoUpdateSchedulerInfoRequest, opts ...grpc.CallOption) (*proto.RepoUpdateSchedulerInfoResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return a.base.RepoUpdateSchedulerInfo(ctx, in, opts...)
}

func (a *automaticRetryClient) RepoLookup(ctx context.Context, in *proto.RepoLookupRequest, opts ...grpc.CallOption) (*proto.RepoLookupResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return a.base.RepoLookup(ctx, in, opts...)
}

func (a *automaticRetryClient) EnqueueRepoUpdate(ctx context.Context, in *proto.EnqueueRepoUpdateRequest, opts ...grpc.CallOption) (*proto.EnqueueRepoUpdateResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return a.base.EnqueueRepoUpdate(ctx, in, opts...)
}

func (a *automaticRetryClient) EnqueueChangesetSync(ctx context.Context, in *proto.EnqueueChangesetSyncRequest, opts ...grpc.CallOption) (*proto.EnqueueChangesetSyncResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return a.base.EnqueueChangesetSync(ctx, in, opts...)
}

var _ proto.RepoUpdaterServiceClient = &automaticRetryClient{}
