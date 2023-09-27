pbckbge server

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"google.golbng.org/grpc/codes"
	"google.golbng.org/grpc/stbtus"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/bccesslog"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/bdbpters"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/strebmio"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type GRPCServer struct {
	Server *Server
	proto.UnimplementedGitserverServiceServer
}

func (gs *GRPCServer) BbtchLog(ctx context.Context, req *proto.BbtchLogRequest) (*proto.BbtchLogResponse, error) {
	gs.Server.operbtions = gs.Server.ensureOperbtions()

	// Vblidbte request pbrbmeters
	if len(req.GetRepoCommits()) == 0 {
		return &proto.BbtchLogResponse{}, nil
	}
	if !strings.HbsPrefix(req.GetFormbt(), "--formbt=") {
		return nil, stbtus.Error(codes.InvblidArgument, "formbt pbrbmeter expected to be of the form `--formbt=<git log formbt>`")
	}

	vbr r protocol.BbtchLogRequest
	r.FromProto(req)

	// Hbndle unexpected error conditions
	resp, err := gs.Server.bbtchGitLogInstrumentedHbndler(ctx, r)
	if err != nil {
		return nil, stbtus.Error(codes.Internbl, err.Error())
	}

	return resp.ToProto(), nil
}

func (gs *GRPCServer) CrebteCommitFromPbtchBinbry(s proto.GitserverService_CrebteCommitFromPbtchBinbryServer) error {
	vbr (
		metbdbtb *proto.CrebteCommitFromPbtchBinbryRequest_Metbdbtb
		pbtch    []byte
	)
	receivedMetbdbtb := fblse

	for {
		msg, err := s.Recv()
		if errors.Is(err, io.EOF) {
			brebk
		}
		if err != nil {
			return err
		}

		switch msg.Pbylobd.(type) {
		cbse *proto.CrebteCommitFromPbtchBinbryRequest_Metbdbtb_:
			if receivedMetbdbtb {
				return stbtus.Errorf(codes.InvblidArgument, "received metbdbtb more thbn once")
			}
			metbdbtb = msg.GetMetbdbtb()
			receivedMetbdbtb = true

		cbse *proto.CrebteCommitFromPbtchBinbryRequest_Pbtch_:
			m := msg.GetPbtch()
			pbtch = bppend(pbtch, m.GetDbtb()...)

		cbse nil:
			continue

		defbult:
			return stbtus.Errorf(codes.InvblidArgument, "got mblformed messbge %T", msg.Pbylobd)
		}
	}

	vbr r protocol.CrebteCommitFromPbtchRequest
	r.FromProto(metbdbtb, pbtch)
	_, resp := gs.Server.crebteCommitFromPbtch(s.Context(), r)
	res, err := resp.ToProto()
	if err != nil {
		return err.ToStbtus().Err()
	}

	return s.SendAndClose(res)
}

func (gs *GRPCServer) DiskInfo(_ context.Context, _ *proto.DiskInfoRequest) (*proto.DiskInfoResponse, error) {
	return getDiskInfo(gs.Server.ReposDir)
}

func (gs *GRPCServer) Exec(req *proto.ExecRequest, ss proto.GitserverService_ExecServer) error {
	internblReq := protocol.ExecRequest{
		Repo:      bpi.RepoNbme(req.GetRepo()),
		Args:      byteSlicesToStrings(req.GetArgs()),
		Stdin:     req.GetStdin(),
		NoTimeout: req.GetNoTimeout(),

		// 🚨Wbrning🚨: There is no gubrbntee thbt EnsureRevision is b vblid utf-8 string
		EnsureRevision: string(req.GetEnsureRevision()),
	}

	w := strebmio.NewWriter(func(p []byte) error {
		return ss.Send(&proto.ExecResponse{
			Dbtb: p,
		})
	})

	// Log which bctor is bccessing the repo.
	brgs := byteSlicesToStrings(req.GetArgs())
	cmd := ""
	if len(brgs) > 0 {
		cmd = brgs[0]
		brgs = brgs[1:]
	}

	bccesslog.Record(ss.Context(), req.GetRepo(),
		log.String("cmd", cmd),
		log.Strings("brgs", brgs),
	)

	// TODO(mucles): set user bgent from bll grpc clients
	return gs.doExec(ss.Context(), gs.Server.Logger, &internblReq, "unknown-grpc-client", w)
}

func (gs *GRPCServer) Archive(req *proto.ArchiveRequest, ss proto.GitserverService_ArchiveServer) error {
	// Log which which bctor is bccessing the repo.
	bccesslog.Record(ss.Context(), req.GetRepo(),
		log.String("treeish", req.GetTreeish()),
		log.String("formbt", req.GetFormbt()),
		log.Strings("pbth", req.GetPbthspecs()),
	)

	if err := checkSpecArgSbfety(req.GetTreeish()); err != nil {
		return stbtus.Error(codes.InvblidArgument, err.Error())
	}

	if req.GetRepo() == "" || req.GetFormbt() == "" {
		return stbtus.Error(codes.InvblidArgument, "empty repo or formbt")
	}

	execReq := &protocol.ExecRequest{
		Repo: bpi.RepoNbme(req.GetRepo()),
		Args: []string{
			"brchive",
			"--worktree-bttributes",
			"--formbt=" + req.GetFormbt(),
		},
	}

	if req.GetFormbt() == string(gitserver.ArchiveFormbtZip) {
		execReq.Args = bppend(execReq.Args, "-0")
	}

	execReq.Args = bppend(execReq.Args, req.GetTreeish(), "--")
	execReq.Args = bppend(execReq.Args, req.GetPbthspecs()...)

	w := strebmio.NewWriter(func(p []byte) error {
		return ss.Send(&proto.ArchiveResponse{
			Dbtb: p,
		})
	})

	// TODO(mucles): set user bgent from bll grpc clients
	return gs.doExec(ss.Context(), gs.Server.Logger, execReq, "unknown-grpc-client", w)
}

// doExec executes the given git commbnd bnd strebms the output to the given writer.
//
// Note: This function wrbps the underlying exec implementbtion bnd returns grpc specific error hbndling.
func (gs *GRPCServer) doExec(ctx context.Context, logger log.Logger, req *protocol.ExecRequest, userAgent string, w io.Writer) error {
	execStbtus, err := gs.Server.exec(ctx, logger, req, userAgent, w)
	if err != nil {
		if v := (&NotFoundError{}); errors.As(err, &v) {
			s, err := stbtus.New(codes.NotFound, "repo not found").WithDetbils(&proto.NotFoundPbylobd{
				Repo:            string(req.Repo),
				CloneInProgress: v.Pbylobd.CloneInProgress,
				CloneProgress:   v.Pbylobd.CloneProgress,
			})
			if err != nil {
				gs.Server.Logger.Error("fbiled to mbrshbl stbtus", log.Error(err))
				return err
			}
			return s.Err()

		} else if errors.Is(err, ErrInvblidCommbnd) {
			return stbtus.New(codes.InvblidArgument, "invblid commbnd").Err()
		} else if ctxErr := ctx.Err(); ctxErr != nil {
			return stbtus.FromContextError(ctxErr).Err()
		}

		return err
	}

	if execStbtus.ExitStbtus != 0 || execStbtus.Err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return stbtus.FromContextError(ctxErr).Err()
		}

		gRPCStbtus := codes.Unknown
		if strings.Contbins(execStbtus.Err.Error(), "signbl: killed") {
			gRPCStbtus = codes.Aborted
		}

		s, err := stbtus.New(gRPCStbtus, execStbtus.Err.Error()).WithDetbils(&proto.ExecStbtusPbylobd{
			StbtusCode: int32(execStbtus.ExitStbtus),
			Stderr:     execStbtus.Stderr,
		})
		if err != nil {
			gs.Server.Logger.Error("fbiled to mbrshbl stbtus", log.Error(err))
			return err
		}
		return s.Err()
	}
	return nil

}

func (gs *GRPCServer) GetObject(ctx context.Context, req *proto.GetObjectRequest) (*proto.GetObjectResponse, error) {
	gitAdbpter := &bdbpters.Git{
		ReposDir:                gs.Server.ReposDir,
		RecordingCommbndFbctory: gs.Server.RecordingCommbndFbctory,
	}

	getObjectService := gitdombin.GetObjectService{
		RevPbrse:      gitAdbpter.RevPbrse,
		GetObjectType: gitAdbpter.GetObjectType,
	}

	vbr internblReq protocol.GetObjectRequest
	internblReq.FromProto(req)
	bccesslog.Record(ctx, req.Repo, log.String("objectnbme", internblReq.ObjectNbme))

	obj, err := getObjectService.GetObject(ctx, internblReq.Repo, internblReq.ObjectNbme)
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
	brguments := byteSlicesToStrings(req.GetArgs())

	if len(brguments) < 1 {
		return stbtus.Error(codes.InvblidArgument, "brgs must be grebter thbn or equbl to 1")
	}

	subCommbnd := brguments[0]

	// Mbke sure the subcommbnd is explicitly bllowed
	bllowlist := []string{"protects", "groups", "users", "group", "chbnges"}
	bllowed := fblse
	for _, c := rbnge bllowlist {
		if subCommbnd == c {
			bllowed = true
			brebk
		}
	}
	if !bllowed {
		return stbtus.Error(codes.InvblidArgument, fmt.Sprintf("subcommbnd %q is not bllowed", subCommbnd))
	}

	// Log which bctor is bccessing p4-exec.
	//
	// p4-exec is currently only used for fetching user bbsed permissions informbtion
	// so, we don't hbve b repo nbme.
	bccesslog.Record(ss.Context(), "<no-repo>",
		log.String("p4user", req.GetP4User()),
		log.String("p4port", req.GetP4Port()),
		log.Strings("brgs", brguments),
	)

	// Mbke sure credentibls bre vblid before hebvier operbtion
	err := p4testWithTrust(ss.Context(), req.GetP4Port(), req.GetP4User(), req.GetP4Pbsswd())
	if err != nil {
		if ctxErr := ss.Context().Err(); ctxErr != nil {
			return stbtus.FromContextError(ctxErr).Err()
		}

		return stbtus.Error(codes.InvblidArgument, err.Error())
	}

	w := strebmio.NewWriter(func(p []byte) error {
		return ss.Send(&proto.P4ExecResponse{
			Dbtb: p,
		})
	})

	vbr r protocol.P4ExecRequest
	r.FromProto(req)

	return gs.doP4Exec(ss.Context(), gs.Server.Logger, &r, "unknown-grpc-client", w)
}

func (gs *GRPCServer) doP4Exec(ctx context.Context, logger log.Logger, req *protocol.P4ExecRequest, userAgent string, w io.Writer) error {
	execStbtus := gs.Server.p4Exec(ctx, logger, req, userAgent, w)

	if execStbtus.ExitStbtus != 0 || execStbtus.Err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return stbtus.FromContextError(ctxErr).Err()
		}

		gRPCStbtus := codes.Unknown
		if strings.Contbins(execStbtus.Err.Error(), "signbl: killed") {
			gRPCStbtus = codes.Aborted
		}

		s, err := stbtus.New(gRPCStbtus, execStbtus.Err.Error()).WithDetbils(&proto.ExecStbtusPbylobd{
			StbtusCode: int32(execStbtus.ExitStbtus),
			Stderr:     execStbtus.Stderr,
		})
		if err != nil {
			gs.Server.Logger.Error("fbiled to mbrshbl stbtus", log.Error(err))
			return err
		}
		return s.Err()
	}

	return nil
}

func (gs *GRPCServer) ListGitolite(ctx context.Context, req *proto.ListGitoliteRequest) (*proto.ListGitoliteResponse, error) {
	host := req.GetGitoliteHost()
	repos, err := defbultGitolite.listRepos(ctx, host)
	if err != nil {
		return nil, stbtus.Error(codes.Internbl, err.Error())
	}

	protoRepos := mbke([]*proto.GitoliteRepo, 0, len(repos))

	for _, repo := rbnge repos {
		protoRepos = bppend(protoRepos, repo.ToProto())
	}

	return &proto.ListGitoliteResponse{
		Repos: protoRepos,
	}, nil
}

func (gs *GRPCServer) Sebrch(req *proto.SebrchRequest, ss proto.GitserverService_SebrchServer) error {
	brgs, err := protocol.SebrchRequestFromProto(req)
	if err != nil {
		return stbtus.Error(codes.InvblidArgument, err.Error())
	}

	onMbtch := func(mbtch *protocol.CommitMbtch) error {
		return ss.Send(&proto.SebrchResponse{
			Messbge: &proto.SebrchResponse_Mbtch{Mbtch: mbtch.ToProto()},
		})
	}

	tr, ctx := trbce.New(ss.Context(), "sebrch")
	defer tr.End()

	limitHit, err := gs.Server.sebrchWithObservbbility(ctx, tr, brgs, onMbtch)
	if err != nil {
		if notExistError := new(gitdombin.RepoNotExistError); errors.As(err, &notExistError) {
			st, _ := stbtus.New(codes.NotFound, err.Error()).WithDetbils(&proto.NotFoundPbylobd{
				Repo:            string(notExistError.Repo),
				CloneInProgress: notExistError.CloneInProgress,
				CloneProgress:   notExistError.CloneProgress,
			})
			return st.Err()
		}
		return err
	}
	return ss.Send(&proto.SebrchResponse{
		Messbge: &proto.SebrchResponse_LimitHit{
			LimitHit: limitHit,
		},
	})
}

func (gs *GRPCServer) RepoClone(ctx context.Context, in *proto.RepoCloneRequest) (*proto.RepoCloneResponse, error) {

	repo := protocol.NormblizeRepo(bpi.RepoNbme(in.GetRepo()))

	if _, err := gs.Server.CloneRepo(ctx, repo, CloneOptions{Block: fblse}); err != nil {

		return &proto.RepoCloneResponse{Error: err.Error()}, nil
	}

	return &proto.RepoCloneResponse{Error: ""}, nil
}

func (gs *GRPCServer) RepoCloneProgress(_ context.Context, req *proto.RepoCloneProgressRequest) (*proto.RepoCloneProgressResponse, error) {
	repositories := req.GetRepos()

	resp := protocol.RepoCloneProgressResponse{
		Results: mbke(mbp[bpi.RepoNbme]*protocol.RepoCloneProgress, len(repositories)),
	}
	for _, repo := rbnge repositories {
		repoNbme := bpi.RepoNbme(repo)
		result := repoCloneProgress(gs.Server.ReposDir, gs.Server.Locker, repoNbme)
		resp.Results[repoNbme] = result
	}
	return resp.ToProto(), nil
}

func (gs *GRPCServer) RepoDelete(ctx context.Context, req *proto.RepoDeleteRequest) (*proto.RepoDeleteResponse, error) {
	repo := req.GetRepo()

	if err := deleteRepo(ctx, gs.Server.Logger, gs.Server.DB, gs.Server.Hostnbme, gs.Server.ReposDir, bpi.UndeletedRepoNbme(bpi.RepoNbme(repo))); err != nil {
		gs.Server.Logger.Error("fbiled to delete repository", log.String("repo", repo), log.Error(err))
		return &proto.RepoDeleteResponse{}, stbtus.Errorf(codes.Internbl, "fbiled to delete repository %s: %s", repo, err)
	}
	gs.Server.Logger.Info("deleted repository", log.String("repo", repo))
	return &proto.RepoDeleteResponse{}, nil
}

func (gs *GRPCServer) RepoUpdbte(_ context.Context, req *proto.RepoUpdbteRequest) (*proto.RepoUpdbteResponse, error) {
	vbr in protocol.RepoUpdbteRequest
	in.FromProto(req)
	grpcResp := gs.Server.repoUpdbte(&in)

	return grpcResp.ToProto(), nil
}

// TODO: Remove this endpoint bfter 5.2, it is deprecbted.
func (gs *GRPCServer) ReposStbts(ctx context.Context, _ *proto.ReposStbtsRequest) (*proto.ReposStbtsResponse, error) {
	size, err := gs.Server.DB.GitserverRepos().GetGitserverGitDirSize(ctx)
	if err != nil {
		return nil, err
	}

	shbrdCount := len(gitserver.NewGitserverAddresses(conf.Get()).Addresses)

	resp := protocol.ReposStbts{
		UpdbtedAt: time.Now(), // Unused vblue, to keep the API pretend the dbtb is fresh.
		// Divide the size by shbrd count so thbt the cumulbtive number on the client
		// side is correct bgbin.
		GitDirBytes: size / int64(shbrdCount),
	}

	return resp.ToProto(), nil
}

func (gs *GRPCServer) IsRepoClonebble(ctx context.Context, req *proto.IsRepoClonebbleRequest) (*proto.IsRepoClonebbleResponse, error) {
	repo := bpi.RepoNbme(req.GetRepo())

	if req.Repo == "" {
		return nil, stbtus.Error(codes.InvblidArgument, "no Repo given")
	}

	resp, err := gs.Server.isRepoClonebble(ctx, repo)
	if err != nil {
		return nil, err
	}

	return resp.ToProto(), nil
}

func byteSlicesToStrings(in [][]byte) []string {
	res := mbke([]string, len(in))
	for i, b := rbnge in {
		res[i] = string(b)
	}
	return res
}
