package internal

import (
	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
)

type createCommitFromPatchBinaryCallbackServer struct {
	recvCallback func(req *proto.CreateCommitFromPatchBinaryRequest, err error)

	proto.GitserverService_CreateCommitFromPatchBinaryServer
}

func (s *createCommitFromPatchBinaryCallbackServer) Recv() (req *proto.CreateCommitFromPatchBinaryRequest, err error) {
	defer func() {
		if s.recvCallback != nil {
			s.recvCallback(req, err)
		}
	}()

	return s.GitserverService_CreateCommitFromPatchBinaryServer.Recv()
}

func newCreateCommitFromPatchBinaryCallbackServer(s proto.GitserverService_CreateCommitFromPatchBinaryServer, recvCallback func(req *proto.CreateCommitFromPatchBinaryRequest, err error)) proto.GitserverService_CreateCommitFromPatchBinaryServer {
	return &createCommitFromPatchBinaryCallbackServer{
		recvCallback: recvCallback,
		GitserverService_CreateCommitFromPatchBinaryServer: s,
	}
}

var _ proto.GitserverService_CreateCommitFromPatchBinaryServer = &createCommitFromPatchBinaryCallbackServer{}
