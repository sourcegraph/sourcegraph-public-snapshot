package server

import (
	"github.com/sourcegraph/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	"github.com/sourcegraph/sourcegraph/internal/grpc/streamio"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GRPCServer struct {
	Server *Server
	v1.UnimplementedGitserverServiceServer
}

func (gs *GRPCServer) Exec(req *v1.ExecRequest, ss v1.GitserverService_ExecServer) error {
	internalReq := protocol.ExecRequest{
		Repo:           api.RepoName(req.GetRepo()),
		EnsureRevision: req.GetEnsureRevision(),
		Args:           req.GetArgs(),
		Stdin:          req.GetStdin(),
		NoTimeout:      req.GetNoTimeout(),
	}

	w := streamio.NewWriter(func(p []byte) error {
		return ss.Send(&v1.ExecResponse{
			Data: p,
		})
	})

	// TODO(camdencheek): set user agent from all grpc clients
	execStatus, err := gs.Server.exec(ss.Context(), gs.Server.Logger, &internalReq, "unknown-grpc-client", w)
	if err != nil {
		if v := (&NotFoundError{}); errors.As(err, &v) {
			s, err := status.New(codes.NotFound, "repo not found").WithDetails(&v1.NotFoundPayload{
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
		s, err := status.New(codes.Unknown, execStatus.Err.Error()).WithDetails(&v1.ExecStatusPayload{
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
