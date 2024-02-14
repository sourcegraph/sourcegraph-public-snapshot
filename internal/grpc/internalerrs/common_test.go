package internalerrs

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	newspb "github.com/sourcegraph/sourcegraph/internal/grpc/testprotos/news/v1"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRequestSavingClientStream_InitialRequest(t *testing.T) {
	// Setup: create a mock ClientStream that returns a sentinel error on SendMsg
	sentinelErr := errors.New("send error")
	mockClientStream := &mockClientStream{
		sendErr: sentinelErr,
	}

	// Setup: create a requestSavingClientStream with the mock ClientStream
	stream := &requestSavingClientStream{
		ClientStream: mockClientStream,
	}

	// Setup: create a sample proto.Message for the request
	request := &newspb.BinaryAttachment{
		Name: "sample_request",
		Data: []byte("sample data"),
	}

	// Test: call SendMsg with the request
	err := stream.SendMsg(request)

	// Check: assert SendMsg propagates the error
	if !errors.Is(err, sentinelErr) {
		t.Errorf("got %v, want %v", err, sentinelErr)
	}

	// Check: assert InitialRequest returns the request
	if diff := cmp.Diff(request, *stream.InitialRequest(), cmpopts.IgnoreUnexported(newspb.BinaryAttachment{})); diff != "" {
		t.Fatalf("InitialRequest() (-want +got):\n%s", diff)
	}
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

func TestCallBackServerStream(t *testing.T) {
	t.Run("SendMsg calls postMessageSend with message and error", func(t *testing.T) {
		sentinelMessage := struct{}{}
		sentinelErr := errors.New("send error")

		var called bool
		stream := callBackServerStream{
			ServerStream: &mockServerStream{
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
		stream := callBackServerStream{
			ServerStream: &mockServerStream{
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

func TestRequestSavingServerStream_InitialRequest(t *testing.T) {
	// Setup: create a mock ServerStream that returns a sentinel error on SendMsg
	sentinelErr := errors.New("receive error")
	mockServerStream := &mockServerStream{
		recvErr: sentinelErr,
	}

	// Setup: create a requestSavingServerStream with the mock ServerStream
	stream := &requestSavingServerStream{
		ServerStream: mockServerStream,
	}

	// Setup: create a sample proto.Message for the request
	request := &newspb.BinaryAttachment{
		Name: "sample_request",
		Data: []byte("sample data"),
	}

	// Test: call RecvMsg with the request
	err := stream.RecvMsg(request)

	// Check: assert RecvMsg propagates the error
	if !errors.Is(err, sentinelErr) {
		t.Errorf("got %v, want %v", err, sentinelErr)
	}

	// Check: assert InitialRequest returns the request
	if diff := cmp.Diff(request, *stream.InitialRequest(), cmpopts.IgnoreUnexported(newspb.BinaryAttachment{})); diff != "" {
		t.Fatalf("InitialRequest() (-want +got):\n%s", diff)
	}
}

// mockServerStream is a grpc.ServerStream that returns a given error on SendMsg and RecvMsg.
type mockServerStream struct {
	grpc.ServerStream
	sendErr error
	recvErr error
}

func (s *mockServerStream) SendMsg(any) error {
	return s.sendErr
}

func (s *mockServerStream) RecvMsg(any) error {
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

func TestGRPCUnexpectedContentTypeChecker(t *testing.T) {
	tests := []struct {
		name   string
		status *status.Status
		want   bool
	}{
		{
			name:   "gRPC error with OK status",
			status: status.New(codes.OK, "transport: received unexpected content-type"),
			want:   false,
		},
		{
			name:   "gRPC error without unexpected content-type message",
			status: status.New(codes.Internal, "some random error"),
			want:   false,
		},
		{
			name:   "gRPC error with unexpected content-type message",
			status: status.Newf(codes.Internal, "transport: received unexpected content-type %q", "application/octet-stream"),
			want:   true,
		},
		{
			name:   "gRPC error with unexpected content-type message as part of chain",
			status: status.Newf(codes.Unknown, "transport: malformed grpc-status %q; transport: received unexpected content-type %q", "random-status", "application/octet-stream"),
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := gRPCUnexpectedContentTypeChecker(tt.status); got != tt.want {
				t.Errorf("gRPCUnexpectedContentTypeChecker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindNonUTF8StringFields(t *testing.T) {
	// Create instances of the BinaryAttachment and KeyValueAttachment messages
	invalidBinaryAttachment := &newspb.BinaryAttachment{
		Name: "inval\x80id_binary",
		Data: []byte("sample data"),
	}

	invalidKeyValueAttachment := &newspb.KeyValueAttachment{
		Name: "inval\x80id_key_value",
		Data: map[string]string{
			"key1": "value1",
			"key2": "inval\x80id_value",
		},
	}

	// Create a sample Article message with invalid UTF-8 strings
	article := &newspb.Article{
		Author:  "inval\x80id_author",
		Date:    &timestamppb.Timestamp{Seconds: 1234567890},
		Title:   "valid_title",
		Content: "valid_content",
		Status:  newspb.Article_STATUS_PUBLISHED,
		Attachments: []*newspb.Attachment{
			{Contents: &newspb.Attachment_BinaryAttachment{BinaryAttachment: invalidBinaryAttachment}},
			{Contents: &newspb.Attachment_KeyValueAttachment{KeyValueAttachment: invalidKeyValueAttachment}},
		},
	}

	tests := []struct {
		name          string
		message       proto.Message
		expectedPaths []string
	}{
		{
			name:    "Article with invalid UTF-8 strings",
			message: article,
			expectedPaths: []string{
				"author",
				"attachments[0].binary_attachment.name",
				"attachments[1].key_value_attachment.name",
				`attachments[1].key_value_attachment.data["key2"]`,
			},
		},
		{
			name:          "nil message",
			message:       nil,
			expectedPaths: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invalidFields, err := findNonUTF8StringFields(tt.message)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			sort.Strings(invalidFields)
			sort.Strings(tt.expectedPaths)

			if diff := cmp.Diff(tt.expectedPaths, invalidFields, cmpopts.EquateEmpty()); diff != "" {
				t.Fatalf("unexpected invalid fields (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMassageIntoStatusErr(t *testing.T) {
	testCases := []struct {
		description string
		input       error
		expected    *status.Status
		expectedOk  bool
	}{
		{
			description: "nil error",
			input:       nil,
			expected:    nil,
			expectedOk:  false,
		},
		{
			description: "status error",
			input:       status.Errorf(codes.InvalidArgument, "invalid argument"),
			expected:    status.New(codes.InvalidArgument, "invalid argument"),
			expectedOk:  true,
		},
		{
			description: "context.Canceled error",
			input:       context.Canceled,
			expected:    status.New(codes.Canceled, "context canceled"),
			expectedOk:  true,
		},
		{
			description: "context.DeadlineExceeded error",
			input:       context.DeadlineExceeded,
			expected:    status.New(codes.DeadlineExceeded, "context deadline exceeded"),
			expectedOk:  true,
		},
		{
			description: "non-status error",
			input:       errors.New("non-status error"),
			expected:    nil,
			expectedOk:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, ok := massageIntoStatusErr(tc.input)
			if ok != tc.expectedOk {
				t.Errorf("Expected ok to be %v, but got %v", tc.expectedOk, ok)
			}

			expectedStatusString := fmt.Sprintf("%s", tc.expected)
			actualStatusString := fmt.Sprintf("%s", result)

			if diff := cmp.Diff(expectedStatusString, actualStatusString); diff != "" {
				t.Fatalf("Unexpected status string (-want +got):\n%s", diff)
			}
		})
	}
}
