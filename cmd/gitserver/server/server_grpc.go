package server

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server/accesslog"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/adapters"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	"github.com/sourcegraph/sourcegraph/internal/grpc/streamio"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GRPCServer struct {
	Server *Server
	proto.UnimplementedGitserverServiceServer
}

func (gs *GRPCServer) BatchLog(ctx context.Context, req *proto.BatchLogRequest) (*proto.BatchLogResponse, error) {
	gs.Server.operations = gs.Server.ensureOperations()

	// Validate request parameters
	if len(req.GetRepoCommits()) == 0 {
		return &proto.BatchLogResponse{}, nil
	}
	if !strings.HasPrefix(req.GetFormat(), "--format=") {
		return nil, status.Error(codes.InvalidArgument, "format parameter expected to be of the form `--format=<git log format>`")
	}

	var r protocol.BatchLogRequest
	r.FromProto(req)

	// Handle unexpected error conditions
	resp, err := gs.Server.batchGitLogInstrumentedHandler(ctx, r)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp.ToProto(), nil
}

func (gs *GRPCServer) CreateCommitFromPatchBinary(s proto.GitserverService_CreateCommitFromPatchBinaryServer) error {
	var (
		metadata *proto.CreateCommitFromPatchBinaryRequest_Metadata
		patch    []byte
	)
	receivedMetadata := false

	for {
		msg, err := s.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		switch msg.Payload.(type) {
		case *proto.CreateCommitFromPatchBinaryRequest_Metadata_:
			if receivedMetadata {
				return status.Errorf(codes.InvalidArgument, "received metadata more than once")
			}
			metadata = msg.GetMetadata()
			receivedMetadata = true

		case *proto.CreateCommitFromPatchBinaryRequest_Patch_:
			m := msg.GetPatch()
			patch = append(patch, m.GetData()...)

		case nil:
			continue

		default:
			return status.Errorf(codes.InvalidArgument, "got malformed message %T", msg.Payload)
		}
	}

	var r protocol.CreateCommitFromPatchRequest
	r.FromProto(metadata, patch)
	_, resp := gs.Server.createCommitFromPatch(s.Context(), r)
	res, err := resp.ToProto()
	if err != nil {
		return err.ToStatus().Err()
	}

	return s.SendAndClose(res)
}

func (gs *GRPCServer) DiskInfo(_ context.Context, _ *proto.DiskInfoRequest) (*proto.DiskInfoResponse, error) {
	return getDiskInfo(gs.Server.ReposDir)
}

func (gs *GRPCServer) Exec(req *proto.ExecRequest, ss proto.GitserverService_ExecServer) error {
	internalReq := protocol.ExecRequest{
		Repo:      api.RepoName(req.GetRepo()),
		Args:      byteSlicesToStrings(req.GetArgs()),
		Stdin:     req.GetStdin(),
		NoTimeout: req.GetNoTimeout(),

		// 🚨Warning🚨: There is no guarantee that EnsureRevision is a valid utf-8 string
		EnsureRevision: string(req.GetEnsureRevision()),
	}

	w := streamio.NewWriter(func(p []byte) error {
		return ss.Send(&proto.ExecResponse{
			Data: p,
		})
	})

	// Log which actor is accessing the repo.
	args := byteSlicesToStrings(req.GetArgs())
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
	accesslog.Record(ss.Context(), req.GetRepo(),
		log.String("treeish", req.GetTreeish()),
		log.String("format", req.GetFormat()),
		log.Strings("path", req.GetPathspecs()),
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
			"--format=" + req.GetFormat(),
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
		} else if ctxErr := ctx.Err(); ctxErr != nil {
			return status.FromContextError(ctxErr).Err()
		}

		return err
	}

	if execStatus.ExitStatus != 0 || execStatus.Err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return status.FromContextError(ctxErr).Err()
		}

		gRPCStatus := codes.Unknown
		if strings.Contains(execStatus.Err.Error(), "signal: killed") {
			gRPCStatus = codes.Aborted
		}

		s, err := status.New(gRPCStatus, execStatus.Err.Error()).WithDetails(&proto.ExecStatusPayload{
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

func (gs *GRPCServer) GetObject(ctx context.Context, req *proto.GetObjectRequest) (*proto.GetObjectResponse, error) {
	gitAdapter := &adapters.Git{
		ReposDir:                gs.Server.ReposDir,
		RecordingCommandFactory: gs.Server.RecordingCommandFactory,
	}

	getObjectService := gitdomain.GetObjectService{
		RevParse:      gitAdapter.RevParse,
		GetObjectType: gitAdapter.GetObjectType,
	}

	var internalReq protocol.GetObjectRequest
	internalReq.FromProto(req)
	accesslog.Record(ctx, req.Repo, log.String("objectname", internalReq.ObjectName))

	obj, err := getObjectService.GetObject(ctx, internalReq.Repo, internalReq.ObjectName)
	if err != nil {
		gs.Server.Logger.Error("getting object", log.Error(err))
		return nil, err
	}

	resp := protocol.GetObjectResponse{
		Object: *obj,
	}

	return resp.ToProto(), nil
}

func (gs *GRPCServer) P4Exec(req *proto.P4ExecRequest, ss proto.GitserverService_P4ExecServer) error {
	arguments := byteSlicesToStrings(req.GetArgs())

	if len(arguments) < 1 {
		return status.Error(codes.InvalidArgument, "args must be greater than or equal to 1")
	}

	subCommand := arguments[0]

	// Make sure the subcommand is explicitly allowed
	allowlist := []string{"protects", "groups", "users", "group", "changes"}
	allowed := false
	for _, c := range allowlist {
		if subCommand == c {
			allowed = true
			break
		}
	}
	if !allowed {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("subcommand %q is not allowed", subCommand))
	}

	// Log which actor is accessing p4-exec.
	//
	// p4-exec is currently only used for fetching user based permissions information
	// so, we don't have a repo name.
	accesslog.Record(ss.Context(), "<no-repo>",
		log.String("p4user", req.GetP4User()),
		log.String("p4port", req.GetP4Port()),
		log.Strings("args", arguments),
	)

	// Make sure credentials are valid before heavier operation
	err := p4testWithTrust(ss.Context(), req.GetP4Port(), req.GetP4User(), req.GetP4Passwd(), filepath.Join(gs.Server.ReposDir, P4HomeName))
	if err != nil {
		if ctxErr := ss.Context().Err(); ctxErr != nil {
			return status.FromContextError(ctxErr).Err()
		}

		return status.Error(codes.InvalidArgument, err.Error())
	}

	w := streamio.NewWriter(func(p []byte) error {
		return ss.Send(&proto.P4ExecResponse{
			Data: p,
		})
	})

	var r protocol.P4ExecRequest
	r.FromProto(req)

	return gs.doP4Exec(ss.Context(), gs.Server.Logger, &r, "unknown-grpc-client", w)
}

func (gs *GRPCServer) doP4Exec(ctx context.Context, logger log.Logger, req *protocol.P4ExecRequest, userAgent string, w io.Writer) error {
	execStatus := gs.Server.p4Exec(ctx, logger, req, userAgent, w)

	if execStatus.ExitStatus != 0 || execStatus.Err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return status.FromContextError(ctxErr).Err()
		}

		gRPCStatus := codes.Unknown
		if strings.Contains(execStatus.Err.Error(), "signal: killed") {
			gRPCStatus = codes.Aborted
		}

		s, err := status.New(gRPCStatus, execStatus.Err.Error()).WithDetails(&proto.ExecStatusPayload{
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

func (gs *GRPCServer) ListGitolite(ctx context.Context, req *proto.ListGitoliteRequest) (*proto.ListGitoliteResponse, error) {
	host := req.GetGitoliteHost()
	repos, err := defaultGitolite.listRepos(ctx, host)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoRepos := make([]*proto.GitoliteRepo, 0, len(repos))

	for _, repo := range repos {
		protoRepos = append(protoRepos, repo.ToProto())
	}

	return &proto.ListGitoliteResponse{
		Repos: protoRepos,
	}, nil
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

	tr, ctx := trace.New(ss.Context(), "search")
	defer tr.End()

	limitHit, err := gs.Server.searchWithObservability(ctx, tr, args, onMatch)
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

func (gs *GRPCServer) RepoClone(ctx context.Context, in *proto.RepoCloneRequest) (*proto.RepoCloneResponse, error) {

	repo := protocol.NormalizeRepo(api.RepoName(in.GetRepo()))

	if _, err := gs.Server.CloneRepo(ctx, repo, CloneOptions{Block: false}); err != nil {

		return &proto.RepoCloneResponse{Error: err.Error()}, nil
	}

	return &proto.RepoCloneResponse{Error: ""}, nil
}

func (gs *GRPCServer) RepoCloneProgress(_ context.Context, req *proto.RepoCloneProgressRequest) (*proto.RepoCloneProgressResponse, error) {
	repositories := req.GetRepos()

	resp := protocol.RepoCloneProgressResponse{
		Results: make(map[api.RepoName]*protocol.RepoCloneProgress, len(repositories)),
	}
	for _, repo := range repositories {
		repoName := api.RepoName(repo)
		result := repoCloneProgress(gs.Server.ReposDir, gs.Server.Locker, repoName)
		resp.Results[repoName] = result
	}
	return resp.ToProto(), nil
}

func (gs *GRPCServer) RepoDelete(ctx context.Context, req *proto.RepoDeleteRequest) (*proto.RepoDeleteResponse, error) {
	repo := req.GetRepo()

	if err := deleteRepo(ctx, gs.Server.Logger, gs.Server.DB, gs.Server.Hostname, gs.Server.ReposDir, api.UndeletedRepoName(api.RepoName(repo))); err != nil {
		gs.Server.Logger.Error("failed to delete repository", log.String("repo", repo), log.Error(err))
		return &proto.RepoDeleteResponse{}, status.Errorf(codes.Internal, "failed to delete repository %s: %s", repo, err)
	}
	gs.Server.Logger.Info("deleted repository", log.String("repo", repo))
	return &proto.RepoDeleteResponse{}, nil
}

func (gs *GRPCServer) RepoUpdate(_ context.Context, req *proto.RepoUpdateRequest) (*proto.RepoUpdateResponse, error) {
	var in protocol.RepoUpdateRequest
	in.FromProto(req)
	grpcResp := gs.Server.repoUpdate(&in)

	return grpcResp.ToProto(), nil
}

// TODO: Remove this endpoint after 5.2, it is deprecated.
func (gs *GRPCServer) ReposStats(ctx context.Context, _ *proto.ReposStatsRequest) (*proto.ReposStatsResponse, error) {
	size, err := gs.Server.DB.GitserverRepos().GetGitserverGitDirSize(ctx)
	if err != nil {
		return nil, err
	}

	shardCount := len(gitserver.NewGitserverAddresses(conf.Get()).Addresses)

	resp := protocol.ReposStats{
		UpdatedAt: time.Now(), // Unused value, to keep the API pretend the data is fresh.
		// Divide the size by shard count so that the cumulative number on the client
		// side is correct again.
		GitDirBytes: size / int64(shardCount),
	}

	return resp.ToProto(), nil
}

func (gs *GRPCServer) IsRepoCloneable(ctx context.Context, req *proto.IsRepoCloneableRequest) (*proto.IsRepoCloneableResponse, error) {
	repo := api.RepoName(req.GetRepo())

	if req.Repo == "" {
		return nil, status.Error(codes.InvalidArgument, "no Repo given")
	}

	resp, err := gs.Server.isRepoCloneable(ctx, repo)
	if err != nil {
		return nil, err
	}

	return resp.ToProto(), nil
}

func byteSlicesToStrings(in [][]byte) []string {
	res := make([]string, len(in))
	for i, b := range in {
		res[i] = string(b)
	}
	return res
}
