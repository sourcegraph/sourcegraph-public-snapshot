package internal

import (
	"context"
	"errors"
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
		name           string
		mockServer     *mockCreateCommitFromPatchBinaryServer
		expectedReq    *proto.CreateCommitFromPatchBinaryRequest
		expectedErr    error
		callbackInvoke bool
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
			expectedErr:    nil,
			callbackInvoke: true,
		},
		{
			name: "receive error",
			mockServer: &mockCreateCommitFromPatchBinaryServer{
				recvErr: errors.New("receive error"),
			},
			expectedErr:    errors.New("receive error"),
			callbackInvoke: true,
		},
		{
			name:           "no callback set",
			mockServer:     &mockCreateCommitFromPatchBinaryServer{},
			expectedErr:    nil,
			callbackInvoke: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var callbackInvoked bool
			callback := func(req *proto.CreateCommitFromPatchBinaryRequest, err error) {
				callbackInvoked = true
				if req != test.expectedReq {
					t.Errorf("Expected request %v, but got %v", test.expectedReq, req)
				}
				if !errors.Is(err, test.expectedErr) {
					t.Errorf("Expected error %v, but got %v", test.expectedErr, err)
				}
			}
			server := &createCommitFromPatchBinaryCallbackServer{
				stream:   test.mockServer,
				callback: callback,
			}

			req, err := server.Recv()
			if err != test.expectedErr {
				t.Errorf("Expected error %v, but got %v", test.expectedErr, err)
			}
			if req != test.expectedReq {
				t.Errorf("Expected request %v, but got %v", test.expectedReq, req)
			}
			if callbackInvoked != test.callbackInvoke {
				t.Errorf("Expected callback invoked %v, but got %v", test.callbackInvoke, callbackInvoked)
			}
		})
	}
}
