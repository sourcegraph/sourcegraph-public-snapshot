package internalerrs

import (
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// callBackClientStream is a grpc.ClientStream that calls a function after SendMsg and RecvMsg.
type callBackClientStream struct {
	grpc.ClientStream

	postMessageSend    func(error)
	postMessageReceive func(error)
}

func (c *callBackClientStream) SendMsg(m interface{}) error {
	err := c.ClientStream.SendMsg(m)
	c.postMessageSend(err)

	return err
}

func (c *callBackClientStream) RecvMsg(m interface{}) error {
	err := c.ClientStream.RecvMsg(m)
	c.postMessageReceive(err)

	return err
}

var _ grpc.ClientStream = &callBackClientStream{}

// probablyInternalGRPCError checks if a gRPC status likely represents an error that comes from
// the go-grpc library.
//
// Note: this is a heuristic and may not be 100% accurate. From a cursory glance at the go-grpc
// source code, it seems most errors are prefixed with "grpc:". This may break in the future, but
// it's better than nothing.
func probablyInternalGRPCError(s *status.Status) bool {
	return s.Code() != codes.OK && strings.HasPrefix(s.Message(), "grpc:")
}

// splitMethodName splits a full gRPC method name in to its components (service, method)
//
// Copied from github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/reporter.go
func splitMethodName(fullMethod string) (string, string) {
	fullMethod = strings.TrimPrefix(fullMethod, "/") // remove leading slash
	if i := strings.Index(fullMethod, "/"); i >= 0 {
		return fullMethod[:i], fullMethod[i+1:]
	}
	return "unknown", "unknown"
}
