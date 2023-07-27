package httpapi

import (
	proto "github.com/sourcegraph/sourcegraph/internal/frontend/v1"
)

type GRPCService struct {
	Server *frontendServer
	proto.UnimplementedFrontendServiceServer
}
