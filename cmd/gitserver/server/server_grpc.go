package server

import (
	"context"
	"encoding/json"
	"io"
	"path/filepath"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server/accesslog"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	"github.com/sourcegraph/sourcegraph/internal/grpc/streamio"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
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

	// Log which actor is accessing the repo.
	args := req.GetArgs()
	cmd := ""
	if len(args) > 0 {
		cmd = args[0]
		args = args[1:]
	}

	accesslog.Record(ss.Context(), req.GetRepo(),
		log.String("cmd", cmd),
		log.Strings("args", args),
	)

	// TODO(mucles): set user agent from all grpc clients
	return gs.doExec(ss.Context(), gs.Server.Logger, &internalReq, "unknown-grpc-client", w)
}

func (gs *GRPCServer) Archive(req *proto.ArchiveRequest, ss proto.GitserverService_ArchiveServer) error {
	// Log which which actor is accessing the repo.
	accesslog.Record(ss.Context(), req.Repo,
		log.String("treeish", req.Treeish),
		log.String("format", req.Format),
		log.Strings("path", req.Pathspecs),
	)

	if err := checkSpecArgSafety(req.GetTreeish()); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	if req.GetRepo() == "" || req.GetFormat() == "" {
		return status.Error(codes.InvalidArgument, "empty repo or format")
	}

	execReq := &protocol.ExecRequest{
		Repo: api.RepoName(req.GetRepo()),
		Args: []string{
			"archive",
			"--worktree-attributes",
			"--format=" + req.Format,
		},
	}

	if req.GetFormat() == string(gitserver.ArchiveFormatZip) {
		execReq.Args = append(execReq.Args, "-0")
	}

	execReq.Args = append(execReq.Args, req.GetTreeish(), "--")
	execReq.Args = append(execReq.Args, req.GetPathspecs()...)

	w := streamio.NewWriter(func(p []byte) error {
		return ss.Send(&proto.ArchiveResponse{
			Data: p,
		})
	})

	// TODO(mucles): set user agent from all grpc clients
	return gs.doExec(ss.Context(), gs.Server.Logger, execReq, "unknown-grpc-client", w)
}

// doExec executes the given git command and streams the output to the given writer.
//
// Note: This function wraps the underlying exec implementation and returns grpc specific error handling.
func (gs *GRPCServer) doExec(ctx context.Context, logger log.Logger, req *protocol.ExecRequest, userAgent string, w io.Writer) error {
	execStatus, err := gs.Server.exec(ctx, logger, req, userAgent, w)
	if err != nil {
		if v := (&NotFoundError{}); errors.As(err, &v) {
			s, err := status.New(codes.NotFound, "repo not found").WithDetails(&proto.NotFoundPayload{
				Repo:            string(req.Repo),
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

func (gs *GRPCServer) ReposStats(ctx context.Context, req *proto.ReposStatsRequest) (*proto.ReposStatsResponse, error) {
	b, err := gs.Server.readReposStatsFile(filepath.Join(gs.Server.ReposDir, reposStatsName))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read %s: %v", reposStatsName, err.Error())
	}

	var stats *protocol.ReposStats
	if err := json.Unmarshal(b, &stats); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unmarshal %s: %v", reposStatsName, err.Error())
	}

	return stats.ToProto(), nil
}

func (gs *GRPCServer) IsRepoCloneable(ctx context.Context, req *proto.IsRepoCloneableRequest) (*proto.IsRepoCloneableResponse, error) {
	if req.Repo == "" {
		return nil, errors.New("no Repo given")
	}

	repo := api.RepoName(req.Repo)

	var syncer VCSSyncer
	// We use an internal actor here as the repo may be private. It is safe since all
	// we return is a bool indicating whether the repo is cloneable or not. Perhaps
	// the only things that could leak here is whether a private repo exists although
	// the endpoint is only available internally so it's low risk.
	remoteURL, err := gs.Server.getRemoteURL(actor.WithInternalActor(ctx), repo)
	if err != nil {
		// We use this endpoint to verify if a repo exists without consuming
		// API rate limit, since many users visit private or bogus repos,
		// so we deduce the unauthenticated clone URL from the repo name.
		remoteURL, _ = vcs.ParseURL("https://" + string(req.Repo) + ".git")

		// At this point we are assuming it's a git repo
		syncer = &GitRepoSyncer{}
	} else {
		syncer, err = gs.Server.GetVCSSyncer(ctx, repo)
		if err != nil {
			return nil, err
		}
	}

	resp := &proto.IsRepoCloneableResponse{
		Cloned: repoCloned(gs.Server.dir(repo)),
	}
	if err := syncer.IsCloneable(ctx, remoteURL); err == nil {
		resp.Cloneable = true
	} else {
		resp.Reason = err.Error()
	}

	return resp, nil
}
