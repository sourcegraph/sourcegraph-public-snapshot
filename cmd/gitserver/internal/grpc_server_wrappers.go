package internal

import (
	"context"
	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	"google.golang.org/grpc/metadata"
)

type createCommitFromPatchBinaryCallbackServer struct {
	stream   proto.GitserverService_CreateCommitFromPatchBinaryServer
	callback func(req *proto.CreateCommitFromPatchBinaryRequest, err error)
}

func (s *createCommitFromPatchBinaryCallbackServer) SendAndClose(response *proto.CreateCommitFromPatchBinaryResponse) error {
	return s.stream.SendAndClose(response)
}

func (s *createCommitFromPatchBinaryCallbackServer) SetHeader(md metadata.MD) error {
	return s.stream.SetHeader(md)
}

func (s *createCommitFromPatchBinaryCallbackServer) SendHeader(md metadata.MD) error {
	return s.stream.SendHeader(md)
}

func (s *createCommitFromPatchBinaryCallbackServer) Context() context.Context {
	return s.stream.Context()
}

func (s *createCommitFromPatchBinaryCallbackServer) SendMsg(m any) error {
	return s.stream.SendMsg(m)
}

func (s *createCommitFromPatchBinaryCallbackServer) RecvMsg(m any) error {
	return s.stream.RecvMsg(m)
}

func (s *createCommitFromPatchBinaryCallbackServer) Send(resp *proto.CreateCommitFromPatchBinaryResponse) error {
	return s.stream.SendMsg(resp)
}

func (s *createCommitFromPatchBinaryCallbackServer) Recv() (req *proto.CreateCommitFromPatchBinaryRequest, err error) {
	defer func() {
		if s.callback != nil {
			s.callback(req, err)
		}
	}()

	return s.stream.Recv()
}

func (s *createCommitFromPatchBinaryCallbackServer) SetTrailer(md metadata.MD) {
	s.stream.SetTrailer(md)
}

var _ proto.GitserverService_CreateCommitFromPatchBinaryServer = &createCommitFromPatchBinaryCallbackServer{}
