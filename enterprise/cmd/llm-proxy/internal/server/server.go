package server

import (
	"context"

	proto "github.com/sourcegraph/sourcegraph/internal/llmproxy/v1"
)

// Server implements the gRPC LLMProxyServiceServer.
type Server struct {
	// UnimplementedLLMProxyServiceServer must be embedded for forward-compatibility.
	proto.UnimplementedLLMProxyServiceServer
}

var _ proto.LLMProxyServiceServer = (*Server)(nil)

func (s *Server) Complete(context.Context, *proto.CompleteRequest) (*proto.CompleteResponse, error) {
	return &proto.CompleteResponse{}, nil
}
