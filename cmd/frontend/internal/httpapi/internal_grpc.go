package httpapi

import (
	"context"

	proto "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/frontend/v1"
)

type GRPCService struct {
	Server *frontendServer
	proto.UnimplementedFrontendServiceServer
}

func (s *GRPCService) ExternalServiceConfigs(ctx context.Context, in *proto.ExternalServiceConfigsRequest) (*proto.ExternalServiceConfigsResponse, error) {
	panic("TODO")
}
