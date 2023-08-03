// Package chunk provides a utility for sending sets of protobuf messages in
// groups of smaller chunks. This is useful for gRPC, which has limitations around the maximum
// size of a message that you can send.
//
// This code is adapted from the gitaly project, which is licensed
// under the MIT license. A copy of that license text can be found at
// https://mit-license.org/.
//
// The code this file was based off can be found here: https://gitlab.com/gitlab-org/gitaly/-/blob/v16.2.0/internal/helper/chunk/chunker_test.go
package chunk

import (
	"context"
	"io"
	"net"
	"strconv"
	"testing"

	"github.com/dustin/go-humanize"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/interop/grpc_testing"
	"google.golang.org/protobuf/proto"
)

func TestChunker(t *testing.T) {
	s := &server{}
	srv, serverSocketPath := runServer(t, s)
	defer srv.Stop()

	client, conn := newClient(t, serverSocketPath)
	defer conn.Close()
	ctx := context.Background()

	inputPayloadSizeBytes := int(3.5 * maxMessageSize)

	stream, err := client.StreamingOutputCall(ctx, &grpc_testing.StreamingOutputCallRequest{
		Payload: &grpc_testing.Payload{
			Body: []byte(strconv.FormatInt(int64(inputPayloadSizeBytes), 10)),
		},
	})

	require.NoError(t, err)

	messageCount := 0
	var receivedPayload []byte
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		messageCount++
		receivedPayload = append(receivedPayload, resp.GetPayload().GetBody()...)

		require.Less(t, proto.Size(resp), maxMessageSize)
	}

	require.Equal(t, 4, messageCount)

	receivedPayloadSizeBytes := len(receivedPayload)

	if receivedPayloadSizeBytes != inputPayloadSizeBytes {
		t.Fatalf("input payload size is not %d bytes (~ %q), got size: %d (~ %q)",
			inputPayloadSizeBytes, humanize.Bytes(uint64(inputPayloadSizeBytes)),
			receivedPayloadSizeBytes, humanize.Bytes(uint64(receivedPayloadSizeBytes)),
		)
	}
}

type server struct {
	grpc_testing.UnimplementedTestServiceServer
}

func (s *server) StreamingOutputCall(req *grpc_testing.StreamingOutputCallRequest, stream grpc_testing.TestService_StreamingOutputCallServer) error {
	const kilobyte = 1024

	bytesToSend, err := strconv.ParseInt(string(req.GetPayload().GetBody()), 10, 64)
	if err != nil {
		return err
	}

	c := New[*grpc_testing.Payload](func(payloads []*grpc_testing.Payload) error {
		var body []byte
		for _, p := range payloads {
			body = append(body, p.GetBody()...)
		}

		return stream.Send(&grpc_testing.StreamingOutputCallResponse{Payload: &grpc_testing.Payload{Body: body}})
	})
	for numBytes := int64(0); numBytes < bytesToSend; numBytes += kilobyte {
		if err := c.Send(&grpc_testing.Payload{Body: make([]byte, kilobyte)}); err != nil {
			return err
		}
	}

	if err := c.Flush(); err != nil {
		return err
	}
	return nil
}

func runServer(t *testing.T, s *server, opt ...grpc.ServerOption) (*grpc.Server, string) {
	grpcServer := grpc.NewServer(opt...)
	grpc_testing.RegisterTestServiceServer(grpcServer, s)

	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	go func() {
		err := grpcServer.Serve(lis)
		require.NoError(t, err)
	}()

	t.Cleanup(func() {
		grpcServer.Stop()
		lis.Close()
	})

	return grpcServer, lis.Addr().String()
}

func newClient(t *testing.T, serverSocketPath string) (grpc_testing.TestServiceClient, *grpc.ClientConn) {
	connOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.Dial(serverSocketPath, connOpts...)
	if err != nil {
		t.Fatal(err)
	}

	return grpc_testing.NewTestServiceClient(conn), conn
}
