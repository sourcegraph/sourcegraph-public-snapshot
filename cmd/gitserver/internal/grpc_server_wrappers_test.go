package internal

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	"testing"

	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	"google.golang.org/grpc/metadata"
)

type mockCreateCommitFromPatchBinaryServer struct {
	recvResp *proto.CreateCommitFromPatchBinaryRequest
	recvErr  error
}

func (s *mockCreateCommitFromPatchBinaryServer) SendAndClose(*proto.CreateCommitFromPatchBinaryResponse) error {
	return nil
}

func (s *mockCreateCommitFromPatchBinaryServer) SetHeader(metadata.MD) error {
	return nil
}

func (s *mockCreateCommitFromPatchBinaryServer) SendHeader(metadata.MD) error {
	return nil
}

func (s *mockCreateCommitFromPatchBinaryServer) Context() context.Context {
	return context.Background()
}

func (s *mockCreateCommitFromPatchBinaryServer) SendMsg(any) error {
	return nil
}

func (s *mockCreateCommitFromPatchBinaryServer) RecvMsg(m any) error {
	req, ok := m.(*proto.CreateCommitFromPatchBinaryRequest)
	if !ok {
		return errors.New("failed to cast to CreateCommitFromPatchBinaryRequest")
	}
	*req = *s.recvResp
	return s.recvErr
}

func (s *mockCreateCommitFromPatchBinaryServer) Send(*proto.CreateCommitFromPatchBinaryResponse) error {
	return nil
}

func (s *mockCreateCommitFromPatchBinaryServer) Recv() (*proto.CreateCommitFromPatchBinaryRequest, error) {
	return s.recvResp, s.recvErr
}

func (s *mockCreateCommitFromPatchBinaryServer) SetTrailer(metadata.MD) {
}

func TestCreateCommitFromPatchBinaryCallbackServer_Recv(t *testing.T) {
	tests := []struct {
		name string

		mockServer        *mockCreateCommitFromPatchBinaryServer
		shouldSetCallback bool

		expectedReq             *proto.CreateCommitFromPatchBinaryRequest
		expectedErr             error
		callbackShouldBeInvoked bool
	}{
		{
			name: "successful receive",
			mockServer: &mockCreateCommitFromPatchBinaryServer{
				recvResp: &proto.CreateCommitFromPatchBinaryRequest{
					Payload: &proto.CreateCommitFromPatchBinaryRequest_Metadata_{
						Metadata: &proto.CreateCommitFromPatchBinaryRequest_Metadata{
							Repo: "test-repo",
						},
					},
				},
			},
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
			mockServer: &mockCreateCommitFromPatchBinaryServer{
				recvErr: errors.New("receive error"),
			},
			shouldSetCallback:       true,
			expectedErr:             errors.New("receive error"),
			callbackShouldBeInvoked: true,
		},
		{
			name:                    "no recvCallback set",
			mockServer:              &mockCreateCommitFromPatchBinaryServer{},
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

			server := newCreateCommitFromPatchBinaryCallbackServer(
				test.mockServer,
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
