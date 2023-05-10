package internalerrs

import (
	"errors"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCallBackClientStream(t *testing.T) {
	t.Run("SendMsg calls postMessageSend with error", func(t *testing.T) {
		sentinelErr := errors.New("send error")

		var called bool
		stream := callBackClientStream{
			ClientStream: &mockClientStream{
				sendErr: sentinelErr,
			},
			postMessageSend: func(err error) {
				called = true

				if !errors.Is(err, sentinelErr) {
					t.Errorf("got %v, want %v", err, sentinelErr)
				}
			},
		}

		sendErr := stream.SendMsg(nil)
		if !called {
			t.Error("postMessageSend not called")
		}

		if !errors.Is(sendErr, sentinelErr) {
			t.Errorf("got %v, want %v", sendErr, sentinelErr)
		}
	})

	t.Run("RecvMsg calls postMessageReceive with error", func(t *testing.T) {
		sentinelErr := errors.New("receive error")

		var called bool
		stream := callBackClientStream{
			ClientStream: &mockClientStream{
				recvErr: sentinelErr,
			},
			postMessageReceive: func(err error) {
				called = true

				if !errors.Is(err, sentinelErr) {
					t.Errorf("got %v, want %v", err, sentinelErr)
				}
			},
		}

		receiveErr := stream.RecvMsg(nil)
		if !called {
			t.Error("postMessageReceive not called")
		}

		if !errors.Is(receiveErr, sentinelErr) {
			t.Errorf("got %v, want %v", receiveErr, sentinelErr)
		}
	})
}

// mockClientStream is a grpc.ClientStream that returns a given error on SendMsg and RecvMsg.
type mockClientStream struct {
	grpc.ClientStream
	sendErr error
	recvErr error
}

func (s *mockClientStream) SendMsg(m interface{}) error {
	return s.sendErr
}

func (s *mockClientStream) RecvMsg(m interface{}) error {
	return s.recvErr
}

func TestProbablyInternalGRPCError(t *testing.T) {
	tests := []struct {
		name   string
		status *status.Status
		want   bool
	}{
		{
			name: "should return false for OK status",

			status: status.New(codes.OK, "grpc: ok"),
			want:   false,
		},
		{
			name: "should return false if message does not start with grpc:",

			status: status.New(codes.Unavailable, "server unavailable: grpc: hmm"),
			want:   false,
		},
		{
			name: "should return false if message doesn't contain 'grpc:' at all",

			status: status.New(codes.Unavailable, "something broke"),
			want:   false,
		},
		{
			name: "should return true if status is non-OK and message starts with grpc:",

			status: status.New(codes.Internal, "grpc: internal error"),
			want:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := probablyInternalGRPCError(test.status)
			if got != test.want {
				t.Errorf("probablyInternalGRPCError(%v) = %v, want %v", test.status, got, test.want)
			}
		})
	}
}
