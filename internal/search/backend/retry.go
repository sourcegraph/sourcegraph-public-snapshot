package backend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	proto "github.com/sourcegraph/zoekt/grpc/protos/zoekt/webserver/v1"
	"google.golang.org/grpc"
)

// automaticRetryClient is a convenience wrapper around a base proto.WebserverServiceClient that automatically retries
// idempotent ("safe") methods in accordance to the policy defined in internal/grpc/defaults.RetryPolicy.
//
// Read the implementation of this type for more details are automatically retried (and why).
//
// Callers are free to override the default retry behavior by proving their own grpc.CallOptions when invoking the RPC.
// (example: providing retry.WithMax(0) will disable retries even when invoking StreamSearch - which is idempotent).
type automaticRetryClient struct {
	base proto.WebserverServiceClient
}

func (a *automaticRetryClient) Search(ctx context.Context, in *proto.SearchRequest, opts ...grpc.CallOption) (*proto.SearchResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return a.base.Search(ctx, in, opts...)
}

func (a *automaticRetryClient) StreamSearch(ctx context.Context, in *proto.StreamSearchRequest, opts ...grpc.CallOption) (proto.WebserverService_StreamSearchClient, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return a.base.StreamSearch(ctx, in, opts...)
}

func (a *automaticRetryClient) List(ctx context.Context, in *proto.ListRequest, opts ...grpc.CallOption) (*proto.ListResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return a.base.List(ctx, in, opts...)
}

var _ proto.WebserverServiceClient = &automaticRetryClient{}
