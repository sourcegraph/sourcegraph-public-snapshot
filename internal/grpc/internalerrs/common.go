package internalerrs

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"unicode/utf8"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protopath"
	"google.golang.org/protobuf/reflect/protorange"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// requestSavingClientStream is a grpc.ClientStream that saves the initial request sent to the server.
type requestSavingClientStream struct {
	grpc.ClientStream

	initialRequest  atomic.Pointer[proto.Message]
	saveRequestOnce sync.Once
}

func (c *requestSavingClientStream) SendMsg(m any) error {
	c.saveRequestOnce.Do(func() {
		message, ok := m.(proto.Message)
		if !ok {
			return
		}

		c.initialRequest.Store(&message)
	})

	return c.ClientStream.SendMsg(m)
}

// InitialRequest returns the initial request sent by the client on the stream.
func (c *requestSavingClientStream) InitialRequest() *proto.Message {
	return c.initialRequest.Load()
}

var _ grpc.ClientStream = &requestSavingClientStream{}

// requestSavingServerStream is a grpc.ServerStream that saves the initial request sent by the client.
type requestSavingServerStream struct {
	grpc.ServerStream

	initialRequest  atomic.Pointer[proto.Message]
	saveRequestOnce sync.Once
}

func (s *requestSavingServerStream) RecvMsg(m any) error {
	s.saveRequestOnce.Do(func() {
		message, ok := m.(proto.Message)
		if !ok {
			return
		}

		s.initialRequest.Store(&message)
	})

	return s.ServerStream.RecvMsg(m)
}

// InitialRequest returns the initial request sent by the client on the stream.
func (s *requestSavingServerStream) InitialRequest() *proto.Message {
	return s.initialRequest.Load()
}

var _ grpc.ServerStream = &requestSavingServerStream{}

// callBackServerStream is a grpc.ServerStream that calls a function after SendMsg and RecvMsg.
type callBackServerStream struct {
	grpc.ServerStream

	postMessageSend    func(message any, err error)
	postMessageReceive func(message any, err error)
}

func (c *callBackServerStream) SendMsg(m any) error {
	err := c.ServerStream.SendMsg(m)

	if c.postMessageSend != nil {
		c.postMessageSend(m, err)
	}

	return err
}

func (c *callBackServerStream) RecvMsg(m any) error {
	err := c.ServerStream.RecvMsg(m)

	if c.postMessageReceive != nil {
		c.postMessageReceive(m, err)
	}

	return err
}

var _ grpc.ServerStream = &callBackServerStream{}

// probablyInternalGRPCError checks if a gRPC status likely represents an error that comes from
// the go-grpc library.
//
// Note: this is a heuristic and may not be 100% accurate.
// From a cursory glance at the go-grpc source code, it seems most errors are prefixed with "grpc:". This may break in the future, but
// it's better than nothing.
// Some other ad-hoc errors that we traced back to the go-grpc library are also checked for.
func probablyInternalGRPCError(s *status.Status, checkers []internalGRPCErrorChecker) bool {
	if s.Code() == codes.OK {
		return false
	}

	for _, checker := range checkers {
		if checker(s) {
			return true
		}
	}

	return false
}

// internalGRPCErrorChecker is a function that checks if a gRPC status likely represents an error that comes from
// the go-grpc library.
type internalGRPCErrorChecker func(*status.Status) bool

// allCheckers is a list of functions that check if a gRPC status likely represents an
// error that comes from the go-grpc library.
var allCheckers = []internalGRPCErrorChecker{
	gRPCPrefixChecker,
	gRPCResourceExhaustedChecker,
	gRPCUnexpectedContentTypeChecker,
}

// gRPCPrefixChecker checks if a gRPC status likely represents an error that comes from the go-grpc library, by checking if the error message
// is prefixed with "grpc: ".
func gRPCPrefixChecker(s *status.Status) bool {
	return s.Code() != codes.OK && strings.HasPrefix(s.Message(), "grpc: ")
}

// gRPCResourceExhaustedChecker checks if a gRPC status likely represents an error that comes from the go-grpc library, by checking if the error message
// is prefixed with "trying to send message larger than max".
func gRPCResourceExhaustedChecker(s *status.Status) bool {
	// Observed from https://github.com/grpc/grpc-go/blob/756119c7de49e91b6f3b9d693b9850e1598938eb/stream.go#L884
	return s.Code() == codes.ResourceExhausted && strings.HasPrefix(s.Message(), "trying to send message larger than max (")
}

// gRPCUnexpectedContentTypeChecker checks if a gRPC status likely represents an error that comes from the go-grpc library, by checking if the error message
// is prefixed with "transport: received unexpected content-type".
func gRPCUnexpectedContentTypeChecker(s *status.Status) bool {
	// Observed from https://github.com/grpc/grpc-go/blob/2997e84fd8d18ddb000ac6736129b48b3c9773ec/internal/transport/http2_client.go#L1415-L1417
	return s.Code() != codes.OK && strings.Contains(s.Message(), "transport: received unexpected content-type")
}

// findNonUTF8StringFields returns a list of field names that contain invalid UTF-8 strings
// in the given proto message.
//
// Example: ["author", "attachments[1].key_value_attachment.data["key2"]`]
func findNonUTF8StringFields(m proto.Message) ([]string, error) {
	if m == nil {
		return nil, nil
	}

	var fields []string
	err := protorange.Range(m.ProtoReflect(), func(p protopath.Values) error {
		last := p.Index(-1)
		s, ok := last.Value.Interface().(string)
		if ok && !utf8.ValidString(s) {
			fieldName := p.Path[1:].String()
			fields = append(fields, strings.TrimPrefix(fieldName, "."))
		}

		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "iterating over proto message")
	}

	return fields, nil
}

// massageIntoStatusErr converts an error into a status.Status if possible.
func massageIntoStatusErr(err error) (s *status.Status, ok bool) {
	if err == nil {
		return nil, false
	}

	if s, ok := status.FromError(err); ok {
		return s, true
	}

	if errors.Is(err, context.Canceled) {
		return status.New(codes.Canceled, context.Canceled.Error()), true

	}

	if errors.Is(err, context.DeadlineExceeded) {
		return status.New(codes.DeadlineExceeded, context.DeadlineExceeded.Error()), true
	}

	return nil, false
}
