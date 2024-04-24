package internal

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	"google.golang.org/grpc/metadata"
)

type mockCreateCommitFromPatchBinaryServer struct {
	mockSendAndClose func(*proto.CreateCommitFromPatchBinaryResponse) error
	mockSetHeader    func(metadata.MD) error
	mockSendHeader   func(metadata.MD) error
	mockContext      func() context.Context
	mockSendMsg      func(any) error
	mockRecvMsg      func(any) error
	mockRecv         func() (*proto.CreateCommitFromPatchBinaryRequest, error)
	mockSetTrailer   func(metadata.MD)
}

func (s *mockCreateCommitFromPatchBinaryServer) SendAndClose(*proto.CreateCommitFromPatchBinaryResponse) error {
	if s.mockSendAndClose != nil {
		return s.mockSendAndClose(nil)
	}

	panic("mockSendAndClose is not set")
}

func (s *mockCreateCommitFromPatchBinaryServer) SetHeader(metadata.MD) error {
	if s.mockSetHeader != nil {
		return s.mockSetHeader(nil)
	}

	panic("mockSetHeader is not set")
}

func (s *mockCreateCommitFromPatchBinaryServer) SendHeader(metadata.MD) error {
	if s.mockSendHeader != nil {
		return s.mockSendHeader(nil)
	}

	panic("mockSendHeader is not set")
}

func (s *mockCreateCommitFromPatchBinaryServer) Context() context.Context {
	if s.mockContext != nil {
		return s.mockContext()
	}

	panic("mockContext is not set")
}

func (s *mockCreateCommitFromPatchBinaryServer) SendMsg(any) error {
	if s.mockSendMsg != nil {
		return s.mockSendMsg(nil)
	}

	panic("mockSendMsg is not set")
}

func (s *mockCreateCommitFromPatchBinaryServer) RecvMsg(m any) error {
	if s.mockRecvMsg != nil {
		return s.mockRecvMsg(nil)
	}

	panic("mockRecvMsg is not set")
}

func (s *mockCreateCommitFromPatchBinaryServer) Recv() (*proto.CreateCommitFromPatchBinaryRequest, error) {
	if s.mockRecv != nil {
		return s.mockRecv()
	}

	panic("mockRecv is not set")
}

func (s *mockCreateCommitFromPatchBinaryServer) SetTrailer(metadata.MD) {
	if s.mockSetTrailer != nil {
		s.mockSetTrailer(nil)
		return
	}

	panic("mockSetTrailer is not set")
}

func TestNewCreateCommitFromPatchBinaryCallbackServer_MocksCalled(t *testing.T) {
	// These tests are to ensure that correct mocks are called when the server methods are called, ensuring that
	// we are correctly wrapping the server and forwarding the calls to the underlying server.

	t.Run("SendAndClose", func(t *testing.T) {
		called := false
		expectedError := errors.New("expected error")
		mockServer := &mockCreateCommitFromPatchBinaryServer{
			mockSendAndClose: func(*proto.CreateCommitFromPatchBinaryResponse) error {
				called = true
				return expectedError
			},
		}

		server := newCreateCommitFromPatchBinaryCallbackServer(mockServer, nil)
		err := server.SendAndClose(nil)

		if !called {
			t.Error("mockSendAndClose was not called")
		}

		if !errors.Is(err, expectedError) {
			t.Errorf("expected error %v, but got %v", expectedError, err)
		}
	})

	t.Run("SetHeader", func(t *testing.T) {
		called := false
		expectedError := errors.New("expected error")
		mockServer := &mockCreateCommitFromPatchBinaryServer{
			mockSetHeader: func(metadata.MD) error {
				called = true
				return expectedError
			},
		}

		server := newCreateCommitFromPatchBinaryCallbackServer(mockServer, nil)
		err := server.SetHeader(nil)

		if !called {
			t.Error("mockSetHeader was not called")
		}

		if !errors.Is(err, expectedError) {
			t.Errorf("expected error %v, but got %v", expectedError, err)
		}
	})

	t.Run("SendHeader", func(t *testing.T) {
		called := false
		expectedError := errors.New("expected error")
		mockServer := &mockCreateCommitFromPatchBinaryServer{
			mockSendHeader: func(metadata.MD) error {
				called = true
				return expectedError
			},
		}

		server := newCreateCommitFromPatchBinaryCallbackServer(mockServer, nil)
		err := server.SendHeader(nil)

		if !called {
			t.Error("mockSendHeader was not called")
		}

		if !errors.Is(err, expectedError) {
			t.Errorf("expected error %v, but got %v", expectedError, err)
		}

	})

	t.Run("Context", func(t *testing.T) {
		called := false
		mockServer := &mockCreateCommitFromPatchBinaryServer{
			mockContext: func() context.Context {
				called = true
				return context.Background()
			},
		}

		server := newCreateCommitFromPatchBinaryCallbackServer(mockServer, nil)
		server.Context()

		if !called {
			t.Error("mockContext was not called")
		}
	})

	t.Run("SendMsg", func(t *testing.T) {
		called := false
		expectedError := errors.New("expected error")
		mockServer := &mockCreateCommitFromPatchBinaryServer{
			mockSendMsg: func(any) error {
				called = true
				return expectedError
			},
		}

		server := newCreateCommitFromPatchBinaryCallbackServer(mockServer, nil)
		err := server.SendMsg(nil)

		if !called {
			t.Error("mockSendMsg was not called")
		}

		if !errors.Is(err, expectedError) {
			t.Errorf("expected error %v, but got %v", expectedError, err)
		}

	})

	t.Run("RecvMsg", func(t *testing.T) {
		called := false
		expectedError := errors.New("expected error")
		mockServer := &mockCreateCommitFromPatchBinaryServer{
			mockRecvMsg: func(any) error {
				called = true
				return expectedError
			},
		}

		server := newCreateCommitFromPatchBinaryCallbackServer(mockServer, nil)
		err := server.RecvMsg(nil)

		if !called {
			t.Error("mockRecvMsg was not called")
		}

		if !errors.Is(err, expectedError) {
			t.Errorf("expected error %v, but got %v", expectedError, err)
		}

	})

	t.Run("Recv", func(t *testing.T) {
		called := false

		expectedRequest := &proto.CreateCommitFromPatchBinaryRequest{
			Payload: &proto.CreateCommitFromPatchBinaryRequest_Metadata_{
				Metadata: &proto.CreateCommitFromPatchBinaryRequest_Metadata{
					Repo: "test-repo",
				},
			},
		}
		expectedError := errors.New("expected error")
		mockServer := &mockCreateCommitFromPatchBinaryServer{
			mockRecv: func() (*proto.CreateCommitFromPatchBinaryRequest, error) {
				called = true
				return expectedRequest, expectedError
			},
		}

		server := newCreateCommitFromPatchBinaryCallbackServer(mockServer, nil)
		req, err := server.Recv()

		if !called {
			t.Error("mockRecv was not called")
		}

		if diff := cmp.Diff(req, expectedRequest, protocmp.Transform()); diff != "" {
			t.Errorf("Unexpected request (-want +got):\n%s", diff)
		}

		if !errors.Is(err, expectedError) {
			t.Errorf("expected error %v, but got %v", expectedError, err)
		}
	})

	t.Run("SetTrailer", func(t *testing.T) {
		called := false
		mockServer := &mockCreateCommitFromPatchBinaryServer{
			mockSetTrailer: func(metadata.MD) {
				called = true
			},
		}

		server := newCreateCommitFromPatchBinaryCallbackServer(mockServer, nil)
		server.SetTrailer(nil)

		if !called {
			t.Error("mockSetTrailer was not called")
		}
	})
}

func TestCreateCommitFromPatchBinaryCallbackServer_Recv(t *testing.T) {
	tests := []struct {
		name string

		expectedReq *proto.CreateCommitFromPatchBinaryRequest
		expectedErr error

		shouldSetCallback       bool
		callbackShouldBeInvoked bool
	}{
		{
			name: "successful receive",
			expectedReq: &proto.CreateCommitFromPatchBinaryRequest{
				Payload: &proto.CreateCommitFromPatchBinaryRequest_Metadata_{
					Metadata: &proto.CreateCommitFromPatchBinaryRequest_Metadata{
						Repo: "test-repo",
					},
				},
			},
			shouldSetCallback:       true,
			expectedErr:             nil,
			callbackShouldBeInvoked: true,
		},
		{
			name: "receive error",

			shouldSetCallback:       true,
			expectedErr:             errors.New("receive error"),
			callbackShouldBeInvoked: true,
		},
		{
			name:                    "no recvCallback set",
			shouldSetCallback:       false,
			expectedErr:             nil,
			callbackShouldBeInvoked: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var callbackInvoked bool
			callback := func(req *proto.CreateCommitFromPatchBinaryRequest, err error) {
				callbackInvoked = true

				// Ensure that the callback receives the same request and error as the server
				// should return.
				if diff := cmp.Diff(test.expectedReq, req, protocmp.Transform()); diff != "" {
					t.Errorf("Unexpected request (-want +got):\n%s", diff)
				}

				if diff := cmp.Diff(fmt.Sprintf("%s", test.expectedErr), fmt.Sprintf("%s", err)); diff != "" {
					t.Errorf("Unexpected error (-want +got):\n%s", diff)
				}
			}

			if !test.shouldSetCallback {
				callback = nil
			}

			mockServer := &mockCreateCommitFromPatchBinaryServer{
				mockRecv: func() (*proto.CreateCommitFromPatchBinaryRequest, error) {
					return test.expectedReq, test.expectedErr
				},
				mockRecvMsg: func(a any) error {
					req, ok := a.(*proto.CreateCommitFromPatchBinaryRequest)
					if !ok {
						return errors.New("failed to cast to CreateCommitFromPatchBinaryRequest")
					}
					*req = *test.expectedReq
					return test.expectedErr
				},
			}

			server := newCreateCommitFromPatchBinaryCallbackServer(
				mockServer,
				callback,
			)

			req, err := server.Recv()
			if diff := cmp.Diff(req, test.expectedReq, protocmp.Transform()); diff != "" {
				t.Errorf("Unexpected request (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(fmt.Sprintf("%s", test.expectedErr), fmt.Sprintf("%s", err)); diff != "" {
				t.Errorf("Unexpected error (-want +got):\n%s", diff)
			}

			if callbackInvoked != test.callbackShouldBeInvoked {
				t.Errorf("Expected recvCallback invoked %v, but got %v", test.callbackShouldBeInvoked, callbackInvoked)
			}
		})
	}
}
