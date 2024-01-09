package searcher

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	proto "github.com/sourcegraph/sourcegraph/internal/searcher/v1"
	"google.golang.org/grpc"
)

// automaticRetryClient is a convenience wrapper around a base proto.SearcherServiceClient that automatically retries
// idempotent ("safe") methods in accordance to the policy defined in internal/grpc/defaults.RetryPolicy.
//
// Read the implementation of this type for more details on what RPCs automatically retried (and why).
//
// Callers are free to override the default retry behavior by proving their own grpc.CallOptions when invoking the RPC.
// (example: providing retry.WithMax(0) will disable retries even when invoking Search - which is idempotent).
type automaticRetryClient struct {
	base proto.SearcherServiceClient
}

func (a *automaticRetryClient) Search(ctx context.Context, in *proto.SearchRequest, opts ...grpc.CallOption) (proto.SearcherService_SearchClient, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return a.base.Search(ctx, in, opts...)
}

var _ proto.SearcherServiceClient = &automaticRetryClient{}
