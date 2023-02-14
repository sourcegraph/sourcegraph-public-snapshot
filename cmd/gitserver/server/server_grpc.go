package server

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	"github.com/sourcegraph/sourcegraph/internal/grpc/streamio"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GRPCServer struct {
	Server *Server
	proto.UnimplementedGitserverServiceServer
}

func (gs *GRPCServer) Exec(req *proto.ExecRequest, ss proto.GitserverService_ExecServer) error {
	internalReq := protocol.ExecRequest{
		Repo:           api.RepoName(req.GetRepo()),
		EnsureRevision: req.GetEnsureRevision(),
		Args:           req.GetArgs(),
		Stdin:          req.GetStdin(),
		NoTimeout:      req.GetNoTimeout(),
	}

	w := streamio.NewWriter(func(p []byte) error {
		return ss.Send(&proto.ExecResponse{
			Data: p,
		})
	})

	// TODO(camdencheek): set user agent from all grpc clients
	execStatus, err := gs.Server.exec(ss.Context(), gs.Server.Logger, &internalReq, "unknown-grpc-client", w)
	if err != nil {
		if v := (&NotFoundError{}); errors.As(err, &v) {
			s, err := status.New(codes.NotFound, "repo not found").WithDetails(&proto.NotFoundPayload{
				Repo:            req.GetRepo(),
				CloneInProgress: v.Payload.CloneInProgress,
				CloneProgress:   v.Payload.CloneProgress,
			})
			if err != nil {
				gs.Server.Logger.Error("failed to marshal status", log.Error(err))
				return err
			}
			return s.Err()

		} else if errors.Is(err, ErrInvalidCommand) {
			return status.New(codes.InvalidArgument, "invalid command").Err()
		}

		return err
	}

	if execStatus.ExitStatus != 0 || execStatus.Err != nil {
		s, err := status.New(codes.Unknown, execStatus.Err.Error()).WithDetails(&proto.ExecStatusPayload{
			StatusCode: int32(execStatus.ExitStatus),
			Stderr:     execStatus.Stderr,
		})
		if err != nil {
			gs.Server.Logger.Error("failed to marshal status", log.Error(err))
			return err
		}
		return s.Err()
	}

	return nil
}

func (gs *GRPCServer) Search(req *proto.SearchRequest, ss proto.GitserverService_SearchServer) error {
	args, err := protocol.SearchRequestFromProto(req)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	onMatch := func(match *protocol.CommitMatch) error {
		return ss.Send(&proto.SearchResponse{
			Message: &proto.SearchResponse_Match{Match: match.ToProto()},
		})
	}

	limitHit, err := gs.Server.search(ss.Context(), args, onMatch)
	if err != nil {
		if notExistError := new(gitdomain.RepoNotExistError); errors.As(err, &notExistError) {
			st, _ := status.New(codes.NotFound, err.Error()).WithDetails(&proto.NotFoundPayload{
				Repo:            string(notExistError.Repo),
				CloneInProgress: notExistError.CloneInProgress,
				CloneProgress:   notExistError.CloneProgress,
			})
			return st.Err()
		}
		return err
	}
	return ss.Send(&proto.SearchResponse{
		Message: &proto.SearchResponse_LimitHit{
			LimitHit: limitHit,
		},
	})
}
