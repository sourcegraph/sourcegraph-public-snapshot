// Copyright 2022-2023 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package grpcreflect

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"

	"connectrpc.com/connect"
	reflectionv1 "connectrpc.com/grpcreflect/internal/gen/go/connectext/grpc/reflection/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// Client is a Connect client for the server reflection service.
type Client struct {
	clientV1        *reflectClient
	clientV1Alpha   *reflectClient
	v1unimplemented atomic.Bool
}

// NewClient returns a client for interacting with the gRPC server reflection service.
// The given HTTP client, base URL, and options are used to connect to the service.
//
// This client will try "v1" of the service first (grpc.reflection.v1.ServerReflection).
// If this results in a "Not Implemented" error, the client will fall back to "v1alpha"
// of the service (grpc.reflection.v1alpha.ServerReflection).
func NewClient(httpClient connect.HTTPClient, baseURL string, options ...connect.ClientOption) *Client {
	clientV1 := connect.NewClient[reflectionv1.ServerReflectionRequest, reflectionv1.ServerReflectionResponse](
		httpClient,
		baseURL+serviceURLPathV1+methodName,
		options...,
	)
	clientV1Alpha := connect.NewClient[reflectionv1.ServerReflectionRequest, reflectionv1.ServerReflectionResponse](
		httpClient,
		baseURL+serviceURLPathV1Alpha+methodName,
		options...,
	)
	return &Client{clientV1: clientV1, clientV1Alpha: clientV1Alpha}
}

// NewStream creates a new stream that is used to download reflection information from
// the server. This is a bidirectional stream, so can only be successfully used over
// HTTP/2. The [ClientStream.Close] method should be called when the caller is done
// with the stream.
//
// If any operation returns an error for which [IsReflectionStreamBroken] returns true,
// the stream is broken and all subsequent operations will fail. If the error is not
// a permanent error, the caller should create another stream and try again.
func (c *Client) NewStream(ctx context.Context, options ...ClientStreamOption) *ClientStream {
	clientStream := &ClientStream{
		ctx:    ctx,
		client: c,
	}
	for _, option := range options {
		option.apply(&clientStream.clientStreamOptions)
	}
	// warm-up the stream
	clientStream.getStream()
	return clientStream
}

// ClientStreamOption is an option that can be provided when calling [Client.NewStream].
type ClientStreamOption interface {
	apply(*clientStreamOptions)
}

// WithRequestHeaders is an option that allows the caller to provide the request headers
// that will be sent when a stream is created.
func WithRequestHeaders(headers http.Header) ClientStreamOption {
	return &withRequestHeaders{headers: headers}
}

// WithReflectionHost is an option that allows the caller to provide the hostname that
// will be included with all requests on the stream. This may be used by the server
// when deciding what source of reflection information to use (which could include
// forwarding the request message to a different host).
func WithReflectionHost(host string) ClientStreamOption {
	return &withReflectionHost{host: host}
}

// ClientStream represents a bidirectional stream for downloading reflection information.
// The reflection protocol resembles a sequence of unary RPCs: multiple requests sent on the
// stream, each getting back a corresponding response. However, all such requests and responses
// and sent on a single stream to a single server, to ensure consistency in the information
// downloaded (since different servers could potentially have different versions of reflection
// information).
type ClientStream struct {
	ctx    context.Context //nolint:containedctx
	client *Client
	clientStreamOptions

	mu     sync.Mutex
	stream *reflectStream
	isV1   bool
}

// Spec returns the specification for the reflection RPC.
func (cs *ClientStream) Spec() connect.Spec {
	return cs.getStream().Spec()
}

// Peer describes the server for the RPC.
func (cs *ClientStream) Peer() connect.Peer {
	return cs.getStream().Peer()
}

// ResponseHeader returns the headers received from the server. It blocks until
// the response headers have been sent by the server.
//
// It is possible that the server implementation won't send back response headers
// until after it receives the first request message, sending back headers along
// with the first response message. So it is safest to either call this method
// from a different goroutine than the one that invokes other stream operations
// or to not call this until after the first such operation has completed.
//
// The operations that send a message on the stream are [ListServices], [FileByFilename],
// [FileContainingSymbol], [FileContainingExtension], and [AllExtensionNumbers].
func (cs *ClientStream) ResponseHeader() http.Header {
	return cs.getStream().ResponseHeader()
}

// ListServices retrieves the fully-qualified names for services exposed the server.
//
// This may return a [*connect.Error] indicating the reason the list of services could
// not be retrieved. But if [IsReflectionStreamBroken] returns true for the returned error,
// the stream is broken and cannot be used for further operations.
//
// This operation sends a request message on the stream and waits for the corresponding
// response.
func (cs *ClientStream) ListServices() ([]protoreflect.FullName, error) {
	resp, err := cs.send(&reflectionv1.ServerReflectionRequest{
		MessageRequest: &reflectionv1.ServerReflectionRequest_ListServices{
			ListServices: "",
		},
	})
	if err != nil {
		return nil, err
	}
	respNames := resp.GetListServicesResponse()
	if respNames == nil {
		return nil, errWrongResponseType(resp, "list_services")
	}
	names := make([]protoreflect.FullName, len(respNames.Service))
	for i, svc := range respNames.Service {
		names[i] = protoreflect.FullName(svc.Name)
	}
	return names, nil
}

// FileByFilename retrieves the descriptor for the file with the given path and name.
// The server may respond with multiple files, which represent the request file plus
// its imports or full transitive dependency graph. If the server has already sent back
// some of those files on this stream, they may be omitted from the response.
//
// This may return a [*connect.Error] indicating the reason the file could not be retrieved
// (such as "Not Found" if the given path is not known). But if [IsReflectionStreamBroken]
// returns true for the returned error, the stream is broken and cannot be used for further
// operations.
//
// This operation sends a request message on the stream and waits for the corresponding
// response.
func (cs *ClientStream) FileByFilename(filename string) ([]*descriptorpb.FileDescriptorProto, error) {
	return cs.getDescriptors("file_by_filename", &reflectionv1.ServerReflectionRequest{
		MessageRequest: &reflectionv1.ServerReflectionRequest_FileByFilename{
			FileByFilename: filename,
		},
	})
}

// FileContainingSymbol retrieves the descriptor for the file that defines the element
// with the given fully-qualified name. The server may respond with multiple files, which
// represent not only the file containing the requested symbol but also its imports or
// full transitive dependency graph. If the server has already sent back some of those
// files on this stream, they may be omitted from the response.
//
// This may return a [*connect.Error] indicating the reason the file could not be retrieved
// (such as "Not Found" if the given element is not known). But if [IsReflectionStreamBroken]
// returns true for the returned error, the stream is broken and cannot be used for further
// operations.
//
// This operation sends a request message on the stream and waits for the corresponding
// response.
func (cs *ClientStream) FileContainingSymbol(name protoreflect.FullName) ([]*descriptorpb.FileDescriptorProto, error) {
	return cs.getDescriptors("file_containing_symbol", &reflectionv1.ServerReflectionRequest{
		MessageRequest: &reflectionv1.ServerReflectionRequest_FileContainingSymbol{
			FileContainingSymbol: string(name),
		},
	})
}

// FileContainingExtension retrieves the descriptor for the file that defines the extension
// with the given tag number and for the given extendable message. The server may respond with
// multiple files, which represent not only the file containing the requested symbol but also
// its imports or full transitive dependency graph. If the server has already sent back some
// of those files on this stream, they may be omitted from the response.
//
// This may return a [*connect.Error] indicating the reason the file could not be retrieved
// (such as "Not Found" if the given extension is not known). But if [IsReflectionStreamBroken]
// returns true for the returned error, the stream is broken and cannot be used for further
// operations.
//
// This operation sends a request message on the stream and waits for the corresponding
// response.
func (cs *ClientStream) FileContainingExtension(messageName protoreflect.FullName, extensionNumber protoreflect.FieldNumber) ([]*descriptorpb.FileDescriptorProto, error) {
	return cs.getDescriptors("file_containing_extension", &reflectionv1.ServerReflectionRequest{
		MessageRequest: &reflectionv1.ServerReflectionRequest_FileContainingExtension{
			FileContainingExtension: &reflectionv1.ExtensionRequest{
				ContainingType:  string(messageName),
				ExtensionNumber: int32(extensionNumber),
			},
		},
	})
}

// AllExtensionNumbers retrieves the tag numbers for all extensions of the given message that
// are known to the server.
//
// This may return a [*connect.Error] indicating the reason the list of extension numbers
// could not be retrieved (such as "Not Found" if the given message is not known). But if
// [IsReflectionStreamBroken] returns true for the returned error, the stream is broken and
// cannot be used for further operations.
//
// This operation sends a request message on the stream and waits for the corresponding
// response.
func (cs *ClientStream) AllExtensionNumbers(messageName protoreflect.FullName) ([]protoreflect.FieldNumber, error) {
	resp, err := cs.send(&reflectionv1.ServerReflectionRequest{
		MessageRequest: &reflectionv1.ServerReflectionRequest_AllExtensionNumbersOfType{
			AllExtensionNumbersOfType: string(messageName),
		},
	})
	if err != nil {
		return nil, err
	}
	respExtNumbers := resp.GetAllExtensionNumbersResponse()
	if respExtNumbers == nil {
		return nil, errWrongResponseType(resp, "all_extension_numbers")
	}
	extNumbers := make([]protoreflect.FieldNumber, len(respExtNumbers.ExtensionNumber))
	for i, num := range respExtNumbers.ExtensionNumber {
		extNumbers[i] = protoreflect.FieldNumber(num)
	}
	return extNumbers, nil
}

// Close closes the stream and returns any trailers sent by the server.
func (cs *ClientStream) Close() (http.Header, error) {
	stream := cs.getStream()

	// half-close
	_ = stream.CloseRequest()
	// await final EOF from server (which is also when we get trailers)
	msg, err := stream.Receive()
	if err == nil {
		err = fmt.Errorf("protocol error: server sent unexpected response message (%s)", respType(msg))
	} else if errors.Is(err, io.EOF) {
		err = nil
	}
	// now we can close the stream and retrieve the trailers
	closeErr := stream.CloseResponse()
	if err == nil && closeErr != nil {
		err = closeErr
	}
	return stream.ResponseTrailer(), err
}

func (cs *ClientStream) getStream() *reflectStream {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.getStreamLocked()
}

func (cs *ClientStream) getStreamLocked() *reflectStream {
	if cs.stream != nil {
		return cs.stream
	}
	var connectClient *reflectClient
	if cs.client.v1unimplemented.Load() {
		connectClient = cs.client.clientV1Alpha
		cs.isV1 = false
	} else {
		connectClient = cs.client.clientV1
		cs.isV1 = true
	}
	stream := connectClient.CallBidiStream(cs.ctx)
	for k, v := range cs.headers {
		stream.RequestHeader()[k] = v
	}
	// we can eagerly send request headers; we can ignore return
	// value because caller will see any errors when calling any
	// other method on returned stream
	_ = stream.Send(nil)
	cs.stream = stream
	return cs.stream
}

func (cs *ClientStream) getDescriptors(operation string, req *reflectionv1.ServerReflectionRequest) ([]*descriptorpb.FileDescriptorProto, error) {
	resp, err := cs.send(req)
	if err != nil {
		return nil, err
	}
	respDescriptors := resp.GetFileDescriptorResponse()
	if respDescriptors == nil {
		return nil, errWrongResponseType(resp, operation)
	}
	descriptors := make([]*descriptorpb.FileDescriptorProto, len(respDescriptors.FileDescriptorProto))
	for i, data := range respDescriptors.FileDescriptorProto {
		fileDescriptor := &descriptorpb.FileDescriptorProto{}
		if err := proto.Unmarshal(data, fileDescriptor); err != nil {
			return nil, fmt.Errorf("reply to %s contained invalid descriptor proto: %w", operation, err)
		}
		descriptors[i] = fileDescriptor
	}
	return descriptors, nil
}

func (cs *ClientStream) send(req *reflectionv1.ServerReflectionRequest) (*reflectionv1.ServerReflectionResponse, error) {
	req.Host = cs.host
	// Sending on a bidi stream is usually thread-safe. But the replies are in the same order
	// as the requests. So to prevent concurrent use from interleaving replies (which would
	// require much more logic here to properly correlate replies with requests), we send and
	// receive while holding the mutex. This means that this API does not support pipelining
	// reflection requests. In theory, pipelining could reduce latency, but only if the client
	// knows all of their requests up-front, which is rarely the case since subsequent calls
	// often depend on the data in prior responses.
	cs.mu.Lock()
	defer cs.mu.Unlock()
	for {
		stream := cs.getStreamLocked()
		if err := stream.Send(req); err != nil {
			if errors.Is(err, io.EOF) {
				// need to call Receive to get actual error code
				_, recvErr := stream.Receive()
				if recvErr != nil {
					err = recvErr
				}
			}
			if cs.shouldRetryLocked(err) {
				continue
			}
			return nil, &streamError{err: err}
		}
		resp, err := stream.Receive()
		if err != nil {
			if cs.shouldRetryLocked(err) {
				continue
			}
			return nil, &streamError{err: err}
		}
		if errResp := resp.GetErrorResponse(); errResp != nil {
			return nil, connect.NewWireError(connect.Code(errResp.ErrorCode), errors.New(errResp.ErrorMessage))
		}
		return resp, nil
	}
}

func (cs *ClientStream) shouldRetryLocked(err error) bool {
	if connect.CodeOf(err) == connect.CodeUnimplemented && cs.isV1 {
		// retry w/ v1alpha
		cs.stream = nil
		cs.client.v1unimplemented.Store(true)
		return true
	}
	return false
}

type reflectClient = connect.Client[reflectionv1.ServerReflectionRequest, reflectionv1.ServerReflectionResponse]
type reflectStream = connect.BidiStreamForClient[reflectionv1.ServerReflectionRequest, reflectionv1.ServerReflectionResponse]

type clientStreamOptions struct {
	host    string
	headers http.Header
}

type withRequestHeaders struct {
	headers http.Header
}

func (w *withRequestHeaders) apply(options *clientStreamOptions) {
	options.headers = w.headers
}

type withReflectionHost struct {
	host string
}

func (w *withReflectionHost) apply(options *clientStreamOptions) {
	options.host = w.host
}

type streamError struct {
	err error
}

func (e *streamError) Error() string {
	return e.err.Error()
}

func (e *streamError) Unwrap() error {
	return e.err
}

// IsReflectionStreamBroken returns true if the given error was the result of a [ClientStream]
// failing. If the stream returns an error for which this function returns false, only the
// one operation failed; the stream is still intact and may be used for subsequent operations.
func IsReflectionStreamBroken(err error) bool {
	var streamErr *streamError
	return errors.As(err, &streamErr)
}

func errWrongResponseType(resp *reflectionv1.ServerReflectionResponse, operation string) error {
	return fmt.Errorf("protocol error: wrong response type %T in reply to %s", resp.MessageResponse, operation)
}

func respType(msg *reflectionv1.ServerReflectionResponse) string {
	switch resp := msg.MessageResponse.(type) {
	case *reflectionv1.ServerReflectionResponse_FileDescriptorResponse:
		return "file_descriptor_response"
	case *reflectionv1.ServerReflectionResponse_AllExtensionNumbersResponse:
		return "all_extension_numbers_response"
	case *reflectionv1.ServerReflectionResponse_ListServicesResponse:
		return "list_services_response"
	case *reflectionv1.ServerReflectionResponse_ErrorResponse:
		return fmt.Sprintf("error_response: %v", connect.Code(resp.ErrorResponse.ErrorCode))
	case nil:
		return "empty?"
	default:
		return fmt.Sprintf("unknown: %T", resp)
	}
}
