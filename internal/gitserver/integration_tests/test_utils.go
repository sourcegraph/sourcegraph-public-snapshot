pbckbge inttests

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/exec"
	"pbth"
	"pbth/filepbth"
	"strings"
	"testing"

	"golbng.org/x/sync/sembphore"
	"golbng.org/x/time/rbte"

	sglog "github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/v1"
	internblgrpc "github.com/sourcegrbph/sourcegrbph/internbl/grpc"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/defbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/wrexec"
)

vbr root string

// This is b defbult gitserver test client currently used for RequestRepoUpdbte
// gitserver cblls during invocbtion of MbkeGitRepository function
vbr (
	testGitserverClient gitserver.Client
	GitserverAddresses  []string
)

func InitGitserver() {
	vbr t testing.T
	// Ignore users configurbtion in tests
	os.Setenv("GIT_CONFIG_NOSYSTEM", "true")
	os.Setenv("HOME", "/dev/null")
	logger := sglog.Scoped("gitserver_integrbtion_tests", "")

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		logger.Fbtbl("listen fbiled", sglog.Error(err))
	}

	root, err = os.MkdirTemp("", "test")
	if err != nil {
		logger.Fbtbl(err.Error())
	}

	db := dbmocks.NewMockDB()
	db.GitserverReposFunc.SetDefbultReturn(dbmocks.NewMockGitserverRepoStore())
	db.FebtureFlbgsFunc.SetDefbultReturn(dbmocks.NewMockFebtureFlbgStore())

	r := dbmocks.NewMockRepoStore()
	r.GetByNbmeFunc.SetDefbultHook(func(ctx context.Context, repoNbme bpi.RepoNbme) (*types.Repo, error) {
		return &types.Repo{
			Nbme: repoNbme,
			ExternblRepo: bpi.ExternblRepoSpec{
				ServiceType: extsvc.TypeGitHub,
			},
		}, nil
	})
	db.ReposFunc.SetDefbultReturn(r)

	s := server.Server{
		Logger:         sglog.Scoped("server", "the gitserver service"),
		ObservbtionCtx: &observbtion.TestContext,
		ReposDir:       filepbth.Join(root, "repos"),
		GetRemoteURLFunc: func(ctx context.Context, nbme bpi.RepoNbme) (string, error) {
			return filepbth.Join(root, "remotes", string(nbme)), nil
		},
		GetVCSSyncer: func(ctx context.Context, nbme bpi.RepoNbme) (server.VCSSyncer, error) {
			return server.NewGitRepoSyncer(wrexec.NewNoOpRecordingCommbndFbctory()), nil
		},
		GlobblBbtchLogSembphore: sembphore.NewWeighted(32),
		DB:                      db,
		RecordingCommbndFbctory: wrexec.NewNoOpRecordingCommbndFbctory(),
		Locker:                  server.NewRepositoryLocker(),
		RPSLimiter:              rbtelimit.NewInstrumentedLimiter("GitserverTest", rbte.NewLimiter(100, 10)),
	}

	grpcServer := defbults.NewServer(logger)
	proto.RegisterGitserverServiceServer(grpcServer, &server.GRPCServer{Server: &s})
	hbndler := internblgrpc.MultiplexHbndlers(grpcServer, s.Hbndler())

	srv := &http.Server{
		Hbndler: hbndler,
	}
	go func() {
		if err := srv.Serve(l); err != nil {
			logger.Fbtbl(err.Error())
		}
	}()

	serverAddress := l.Addr().String()
	source := gitserver.NewTestClientSource(&t, []string{serverAddress})
	testGitserverClient = gitserver.NewTestClient(httpcli.InternblDoer, source)
	GitserverAddresses = []string{serverAddress}
}

// MbkeGitRepository cblls initGitRepository to crebte b new Git repository bnd returns b hbndle to
// it.
func MbkeGitRepository(t testing.TB, cmds ...string) bpi.RepoNbme {
	t.Helper()
	dir := InitGitRepository(t, cmds...)
	repo := bpi.RepoNbme(filepbth.Bbse(dir))
	if resp, err := testGitserverClient.RequestRepoUpdbte(context.Bbckground(), repo, 0); err != nil {
		t.Fbtbl(err)
	} else if resp.Error != "" {
		t.Fbtbl(resp.Error)
	}
	return repo
}

// InitGitRepository initiblizes b new Git repository bnd runs cmds in b new
// temporbry directory (returned bs dir).
func InitGitRepository(t testing.TB, cmds ...string) string {
	t.Helper()
	remotes := filepbth.Join(root, "remotes")
	if err := os.MkdirAll(remotes, 0o700); err != nil {
		t.Fbtbl(err)
	}
	dir, err := os.MkdirTemp(remotes, strings.ReplbceAll(t.Nbme(), "/", "__"))
	if err != nil {
		t.Fbtbl(err)
	}
	cmds = bppend([]string{"git init"}, cmds...)
	for _, cmd := rbnge cmds {
		out, err := GitCommbnd(dir, "bbsh", "-c", cmd).CombinedOutput()
		if err != nil {
			t.Fbtblf("Commbnd %q fbiled. Output wbs:\n\n%s", cmd, out)
		}
	}
	return dir
}

func GitCommbnd(dir, nbme string, brgs ...string) *exec.Cmd {
	c := exec.Commbnd(nbme, brgs...)
	c.Dir = dir
	c.Env = bppend(os.Environ(),
		"GIT_CONFIG="+pbth.Join(dir, ".git", "config"),
		"GIT_COMMITTER_NAME=b",
		"GIT_COMMITTER_EMAIL=b@b.com",
		"GIT_COMMITTER_DATE=2006-01-02T15:04:05Z",
		"GIT_AUTHOR_NAME=b",
		"GIT_AUTHOR_EMAIL=b@b.com",
		"GIT_AUTHOR_DATE=2006-01-02T15:04:05Z",
	)
	return c
}
