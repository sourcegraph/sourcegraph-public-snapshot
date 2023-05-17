package internalerrs

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
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
	checker := func(s *status.Status) bool {
		return strings.HasPrefix(s.Message(), "custom error")
	}

	testCases := []struct {
		status     *status.Status
		checkers   []internalGRPCErrorChecker
		wantResult bool
	}{
		{
			status:     status.New(codes.OK, ""),
			checkers:   []internalGRPCErrorChecker{func(*status.Status) bool { return true }},
			wantResult: false,
		},
		{
			status:     status.New(codes.Internal, "custom error message"),
			checkers:   []internalGRPCErrorChecker{checker},
			wantResult: true,
		},
		{
			status:     status.New(codes.Internal, "some other error"),
			checkers:   []internalGRPCErrorChecker{checker},
			wantResult: false,
		},
	}

	for _, tc := range testCases {
		gotResult := probablyInternalGRPCError(tc.status, tc.checkers)
		if gotResult != tc.wantResult {
			t.Errorf("probablyInternalGRPCError(%v, %v) = %v, want %v", tc.status, tc.checkers, gotResult, tc.wantResult)
		}
	}
}

func TestGRPCResourceExhaustedChecker(t *testing.T) {
	testCases := []struct {
		status     *status.Status
		expectPass bool
	}{
		{
			status:     status.New(codes.ResourceExhausted, "trying to send message larger than max (1024 vs 2)"),
			expectPass: true,
		},
		{
			status:     status.New(codes.ResourceExhausted, "some other error"),
			expectPass: false,
		},
		{
			status:     status.New(codes.OK, "trying to send message larger than max (1024 vs 5)"),
			expectPass: false,
		},
	}

	for _, tc := range testCases {
		actual := gRPCResourceExhaustedChecker(tc.status)
		if actual != tc.expectPass {
			t.Errorf("gRPCResourceExhaustedChecker(%v) got %t, want %t", tc.status, actual, tc.expectPass)
		}
	}
}

func TestGRPCPrefixChecker(t *testing.T) {
	tests := []struct {
		status *status.Status
		want   bool
	}{
		{
			status: status.New(codes.OK, "not a grpc error"),
			want:   false,
		},
		{
			status: status.New(codes.Internal, "grpc: internal server error"),
			want:   true,
		},
		{
			status: status.New(codes.Unavailable, "some other error"),
			want:   false,
		},
	}
	for _, test := range tests {
		got := gRPCPrefixChecker(test.status)
		if got != test.want {
			t.Errorf("gRPCPrefixChecker(%v) = %v, want %v", test.status, got, test.want)
		}
	}
}

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
			service, method := splitMethodName(tc.fullMethod)
			if diff := cmp.Diff(service, tc.wantService); diff != "" {
				t.Errorf("splitMethodName(%q) service (-want +got):\n%s", tc.fullMethod, diff)
			}

			if diff := cmp.Diff(method, tc.wantMethod); diff != "" {
				t.Errorf("splitMethodName(%q) method (-want +got):\n%s", tc.fullMethod, diff)
			}
		})
	}
}
