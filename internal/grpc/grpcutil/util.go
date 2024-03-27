package grpcutil

import (
	"strings"

	"google.golang.org/grpc"
)

// SplitMethodName splits a full gRPC method name (e.g. "/package.service/method") in to its individual components (service, method)
//
// Copied from github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/reporter.go
func SplitMethodName(fullMethod string) (string, string) {
	fullMethod = strings.TrimPrefix(fullMethod, "/") // remove leading slash
	if i := strings.Index(fullMethod, "/"); i >= 0 {
		return fullMethod[:i], fullMethod[i+1:]
	}
	return "unknown", "unknown"
}

// NewCallBackClientStream is a grpc.ClientStream that calls the provided functions after SendMsg and RecvMsg.
func NewCallBackClientStream(clientStream grpc.ClientStream, postMessageSend func(message any, err error), postMessageReceive func(message any, err error)) grpc.ClientStream {
	return &callBackClientStream{
		ClientStream:       clientStream,
		postMessageSend:    postMessageSend,
		postMessageReceive: postMessageReceive,
	}
}

type callBackClientStream struct {
	grpc.ClientStream

	postMessageSend    func(message any, err error)
	postMessageReceive func(message any, err error)
}

func (c *callBackClientStream) SendMsg(m any) error {
	err := c.ClientStream.SendMsg(m)
	if c.postMessageSend != nil {
		c.postMessageSend(m, err)
	}

	return err
}

func (c *callBackClientStream) RecvMsg(m any) error {
	err := c.ClientStream.RecvMsg(m)
	if c.postMessageReceive != nil {
		c.postMessageReceive(m, err)
	}

	return err
}

var _ grpc.ClientStream = &callBackClientStream{}
