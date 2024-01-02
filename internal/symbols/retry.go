package symbols

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	proto "github.com/sourcegraph/sourcegraph/internal/symbols/v1"
	"google.golang.org/grpc"
)

// automaticRetryClient is a convenience wrapper around a base proto.SymbolsServiceClient that automatically retries
// idempotent ("safe") methods in accordance to the policy defined in internal/grpc/defaults.RetryPolicy.
//
// Read the implementation of this type for more details on what methods automatically retried (and why).
//
// Callers are free to override the default retry behavior by proving their own grpc.CallOptions when invoking the RPC.
// (example: providing retry.WithMax(0) will disable retries even when invoking Search - which is idempotent).
type automaticRetryClient struct {
	base proto.SymbolsServiceClient
}

func (a *automaticRetryClient) Search(ctx context.Context, in *proto.SearchRequest, opts ...grpc.CallOption) (*proto.SearchResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return a.base.Search(ctx, in, opts...)
}

func (a *automaticRetryClient) LocalCodeIntel(ctx context.Context, in *proto.LocalCodeIntelRequest, opts ...grpc.CallOption) (proto.SymbolsService_LocalCodeIntelClient, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return a.base.LocalCodeIntel(ctx, in, opts...)
}

func (a *automaticRetryClient) SymbolInfo(ctx context.Context, in *proto.SymbolInfoRequest, opts ...grpc.CallOption) (*proto.SymbolInfoResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return a.base.SymbolInfo(ctx, in, opts...)
}

func (a *automaticRetryClient) Healthz(ctx context.Context, in *proto.HealthzRequest, opts ...grpc.CallOption) (*proto.HealthzResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return a.base.Healthz(ctx, in, opts...)
}

var _ proto.SymbolsServiceClient = &automaticRetryClient{}
