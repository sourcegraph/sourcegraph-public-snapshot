package httpapi

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/sourcegraph/internal/frontend"
	proto "github.com/sourcegraph/sourcegraph/internal/frontend/v1"
)

type GRPCService struct {
	Server *frontendServer
	proto.UnimplementedFrontendServiceServer
}

func (s *GRPCService) ExternalServiceConfigs(ctx context.Context, in *proto.ExternalServiceConfigsRequest) (*proto.ExternalServiceConfigsResponse, error) {
	internalReq := frontend.ExternalServiceConfigsRequest{}
	internalReq.FromProto(in)

	configs, err := s.Server.getExternalServiceConfigs(ctx, internalReq)
	if err != nil {
		return nil, err
	}

	jsonConfigs, err := json.Marshal(configs)
	if err != nil {
		return nil, status.Error(codes.Internal, "Error converting to JSON")
	}

	configsString := string(jsonConfigs)

	return &proto.ExternalServiceConfigsResponse{
		Config: configsString,
	}, nil

}
