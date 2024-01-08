package internalapi

import (
	"context"

	proto "github.com/sourcegraph/sourcegraph/internal/api/internalapi/v1"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"google.golang.org/grpc"
)

// automaticRetryClient is a convenience wrapper around a base proto.ConfigServiceClient that automatically retries
// idempotent ("safe") methods in accordance to the policy defined in internal/grpc/defaults.RetryPolicy.
//
// Read the implementation of this type for more details on what RPCs are automatically retried (and why).
//
// Callers are free to override the default retry behavior by proving their own grpc.CallOptions when invoking the RPC.
// (example: providing retry.WithMax(0) will disable retries even when invoking GetConfig - which is idempotent).
type automaticRetryClient struct {
	base proto.ConfigServiceClient
}

func (a *automaticRetryClient) GetConfig(ctx context.Context, in *proto.GetConfigRequest, opts ...grpc.CallOption) (*proto.GetConfigResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return a.base.GetConfig(ctx, in, opts...)
}

var _ proto.ConfigServiceClient = &automaticRetryClient{}
