package grpcutil

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestSplitMethodName(t *testing.T) {
	testCases := []struct {
		name string

		fullMethod  string
		wantService string
		wantMethod  string
	}{
		{
			name: "full method with service and method",

			fullMethod:  "/package.service/method",
			wantService: "package.service",
			wantMethod:  "method",
		},
		{
			name: "method without leading slash",

			fullMethod:  "package.service/method",
			wantService: "package.service",
			wantMethod:  "method",
		},
		{
			name: "service without method",

			fullMethod:  "/package.service/",
			wantService: "package.service",
			wantMethod:  "",
		},
		{
			name: "empty input",

			fullMethod:  "",
			wantService: "unknown",
			wantMethod:  "unknown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service, method := SplitMethodName(tc.fullMethod)
			if diff := cmp.Diff(service, tc.wantService); diff != "" {
				t.Errorf("splitMethodName(%q) service (-want +got):\n%s", tc.fullMethod, diff)
			}

			if diff := cmp.Diff(method, tc.wantMethod); diff != "" {
				t.Errorf("splitMethodName(%q) method (-want +got):\n%s", tc.fullMethod, diff)
			}
		})
	}
}

func TestCallBackClientStream(t *testing.T) {
	t.Run("SendMsg calls postMessageSend with message and error", func(t *testing.T) {
		sentinelMessage := struct{}{}
		sentinelErr := errors.New("send error")

		var called bool
		stream := callBackClientStream{
			ClientStream: &mockClientStream{
				sendErr: sentinelErr,
			},
			postMessageSend: func(message any, err error) {
				called = true

				if diff := cmp.Diff(message, sentinelMessage); diff != "" {
					t.Errorf("postMessageSend called with unexpected message (-want +got):\n%s", diff)
				}
				if !errors.Is(err, sentinelErr) {
					t.Errorf("got %v, want %v", err, sentinelErr)
				}
			},
		}

		sendErr := stream.SendMsg(sentinelMessage)
		if !called {
			t.Error("postMessageSend not called")
		}

		if !errors.Is(sendErr, sentinelErr) {
			t.Errorf("got %v, want %v", sendErr, sentinelErr)
		}
	})

	t.Run("RecvMsg calls postMessageReceive with message and error", func(t *testing.T) {
		sentinelMessage := struct{}{}
		sentinelErr := errors.New("receive error")

		var called bool
		stream := callBackClientStream{
			ClientStream: &mockClientStream{
				recvErr: sentinelErr,
			},
			postMessageReceive: func(message any, err error) {
				called = true

				if diff := cmp.Diff(message, sentinelMessage); diff != "" {
					t.Errorf("postMessageReceive called with unexpected message (-want +got):\n%s", diff)
				}
				if !errors.Is(err, sentinelErr) {
					t.Errorf("got %v, want %v", err, sentinelErr)
				}
			},
		}

		receiveErr := stream.RecvMsg(sentinelMessage)
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

func (s *mockClientStream) SendMsg(any) error {
	return s.sendErr
}

func (s *mockClientStream) RecvMsg(any) error {
	return s.recvErr
}
