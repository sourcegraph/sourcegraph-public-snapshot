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
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"testing"
	"testing/quick"

	"github.com/dustin/go-humanize"
	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/interop/grpc_testing"
	"google.golang.org/protobuf/proto"
)

func TestChunker_DeliverAllMessages(t *testing.T) {
	runTest := func(inputPayloads [][]byte) error {
		expectedPayloadSizeBytes := 0
		for _, payload := range inputPayloads {
			expectedPayloadSizeBytes += len(payload)
		}

		var receivedPayloads []*grpc_testing.Payload

		// Tell the chunker to just gather all the payloads for later inspection.
		sendFunc := func(payloads []*grpc_testing.Payload) error {
			receivedPayloads = append(receivedPayloads, payloads...)
			return nil
		}

		c := New(sendFunc)

		// send all the payloads
		for _, payload := range inputPayloads {
			if err := c.Send(&grpc_testing.Payload{Body: payload}); err != nil {
				return fmt.Errorf("error sending payload: %s", err)
			}
		}

		if err := c.Flush(); err != nil {
			return fmt.Errorf("error flushing chunker: %s", err)
		}

		// Confirm that we received the same number of payloads as we sent.
		if diff := cmp.Diff(len(inputPayloads), len(receivedPayloads)); diff != "" {
			return fmt.Errorf("unexpected number of payloads (-want +got):\n%s", diff)
		}

		// Confirm that each received payload is the same as the original.
		for i, payload := range receivedPayloads {
			expectedPayload := inputPayloads[i]
			if diff := cmp.Diff(expectedPayload, payload.GetBody()); diff != "" {
				return fmt.Errorf("for payload #%d (-want +got):\n%s", i, diff)
			}
		}

		receivedPayloadSizeBytes := 0
		for _, payload := range receivedPayloads {
			receivedPayloadSizeBytes += len(payload.GetBody())
		}

		// Confirm that the total size of the payloads we received is the same as the total size of the payloads we sent.
		if diff := cmp.Diff(expectedPayloadSizeBytes, receivedPayloadSizeBytes); diff != "" {
			return fmt.Errorf("unexpected payload size (-want +got):\n%s", diff)
		}

		return nil
	}

	t.Run("normal", func(t *testing.T) {
		t.Parallel()

		inputPayloads := [][]byte{
			{1, 2, 3},
			bytes.Repeat([]byte("a"), int(3.5*maxMessageSize)),
			{4, 5, 6},
		}

		if err := runTest(inputPayloads); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("some empty", func(t *testing.T) {
		t.Parallel()

		inputPayloads := [][]byte{
			{},
			[]byte("foo, bar, baz"),
			bytes.Repeat([]byte("a"), int(3.5*maxMessageSize)),
			{},
		}

		if err := runTest(inputPayloads); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("fuzz", func(t *testing.T) {
		t.Parallel()

		var lastErr error

		if err := quick.Check(func(payloads [][]byte) bool {
			lastErr = runTest(payloads)
			if lastErr != nil {
				return false
			}

			return true
		}, nil); err != nil {
			t.Fatal(lastErr)
		}
	})
}

func TestChunkerE2E(t *testing.T) {
	for _, test := range []struct {
		name string

		inputSizeBytes       int
		expectedMessageCount int
	}{
		{
			name: "normal",

			inputSizeBytes:       int(3.5 * maxMessageSize),
			expectedMessageCount: 4,
		},
		{
			name:                 "empty payload",
			inputSizeBytes:       0,
			expectedMessageCount: 1,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			s := &server{}
			srv, serverSocketPath := runServer(t, s)
			t.Cleanup(func() {
				srv.Stop()
			})

			client, conn := newClient(t, serverSocketPath)
			t.Cleanup(func() {
				_ = conn.Close()
			})

			ctx := context.Background()

			stream, err := client.StreamingOutputCall(ctx, &grpc_testing.StreamingOutputCallRequest{
				Payload: &grpc_testing.Payload{
					Body: []byte(strconv.FormatInt(int64(test.inputSizeBytes), 10)),
				},
			})

			require.NoError(t, err)

			messageCount := 0
			var receivedPayload []byte
			for {
				resp, err := stream.Recv()
				if errors.Is(err, io.EOF) {
					break
				}

				if err != nil {
					t.Fatal(err)
				}

				messageCount++
				receivedPayload = append(receivedPayload, resp.GetPayload().GetBody()...)

				require.Less(t, proto.Size(resp), maxMessageSize)
			}

			require.Equal(t, test.expectedMessageCount, messageCount)

			receivedPayloadSizeBytes := len(receivedPayload)

			expectedSizeBytes := test.inputSizeBytes

			if receivedPayloadSizeBytes != expectedSizeBytes {
				t.Fatalf("input payload size is not %d bytes (~ %q), got size: %d (~ %q)",
					expectedSizeBytes, humanize.Bytes(uint64(expectedSizeBytes)),
					receivedPayloadSizeBytes, humanize.Bytes(uint64(receivedPayloadSizeBytes)),
				)
			}

		})
	}
}

type server struct {
	grpc_testing.UnimplementedTestServiceServer
}

func (s *server) StreamingOutputCall(req *grpc_testing.StreamingOutputCallRequest, stream grpc_testing.TestService_StreamingOutputCallServer) error {
	const kilobyte = 1024

	c := New[*grpc_testing.Payload](func(payloads []*grpc_testing.Payload) error {
		var body []byte
		for _, p := range payloads {
			body = append(body, p.GetBody()...)
		}

		return stream.Send(&grpc_testing.StreamingOutputCallResponse{Payload: &grpc_testing.Payload{Body: body}})
	})

	bytesToSend, err := strconv.ParseInt(string(req.GetPayload().GetBody()), 10, 64)
	if err != nil {
		return err
	}

	if bytesToSend == 0 {
		if err := c.Send(&grpc_testing.Payload{}); err != nil {
			return err
		}

		return c.Flush()
	}

	for numBytes := int64(0); numBytes < bytesToSend; numBytes += kilobyte {
		if err := c.Send(&grpc_testing.Payload{Body: make([]byte, kilobyte)}); err != nil {
			return err
		}
	}

	return c.Flush()
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
