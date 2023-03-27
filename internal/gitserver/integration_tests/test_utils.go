package inttests

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sync/semaphore"

	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

var root string

// This is a default gitserver test client currently used for RequestRepoUpdate
// gitserver calls during invocation of MakeGitRepository function
var (
	testGitserverClient gitserver.Client
	GitserverAddresses  []string
)

func InitGitserver() {
	// Ignore users configuration in tests
	os.Setenv("GIT_CONFIG_NOSYSTEM", "true")
	os.Setenv("HOME", "/dev/null")
	logger := sglog.Scoped("gitserver_integration_tests", "")

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		logger.Fatal("listen failed", sglog.Error(err))
	}

	root, err = os.MkdirTemp("", "test")
	if err != nil {
		logger.Fatal(err.Error())
	}

	db := database.NewMockDB()
	db.GitserverReposFunc.SetDefaultReturn(database.NewMockGitserverRepoStore())
	db.FeatureFlagsFunc.SetDefaultReturn(database.NewMockFeatureFlagStore())

	s := server.Server{
		Logger:         sglog.Scoped("server", "the gitserver service"),
		ObservationCtx: &observation.TestContext,
		ReposDir:       filepath.Join(root, "repos"),
		GetRemoteURLFunc: func(ctx context.Context, name api.RepoName) (string, error) {
			return filepath.Join(root, "remotes", string(name)), nil
		},
		GetVCSSyncer: func(ctx context.Context, name api.RepoName) (server.VCSSyncer, error) {
			return &server.GitRepoSyncer{}, nil
		},
		GlobalBatchLogSemaphore: semaphore.NewWeighted(32),
		DB:                      db,
	}

	grpcServer := defaults.NewServer(logger)
	proto.RegisterGitserverServiceServer(grpcServer, &server.GRPCServer{Server: &s})
	handler := internalgrpc.MultiplexHandlers(grpcServer, s.Handler())

	srv := &http.Server{
		Handler: handler,
	}
	go func() {
		if err := srv.Serve(l); err != nil {
			logger.Fatal(err.Error())
		}
	}()

	serverAddress := l.Addr().String()
	testGitserverClient = gitserver.NewTestClient(httpcli.InternalDoer, []string{serverAddress})
	GitserverAddresses = []string{serverAddress}
}

// MakeGitRepository calls initGitRepository to create a new Git repository and returns a handle to
// it.
func MakeGitRepository(t testing.TB, cmds ...string) api.RepoName {
	t.Helper()
	dir := InitGitRepository(t, cmds...)
	repo := api.RepoName(filepath.Base(dir))
	if resp, err := testGitserverClient.RequestRepoUpdate(context.Background(), repo, 0); err != nil {
		t.Fatal(err)
	} else if resp.Error != "" {
		t.Fatal(resp.Error)
	}
	return repo
}

// InitGitRepository initializes a new Git repository and runs cmds in a new
// temporary directory (returned as dir).
func InitGitRepository(t testing.TB, cmds ...string) string {
	t.Helper()
	remotes := filepath.Join(root, "remotes")
	if err := os.MkdirAll(remotes, 0o700); err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp(remotes, strings.ReplaceAll(t.Name(), "/", "__"))
	if err != nil {
		t.Fatal(err)
	}
	cmds = append([]string{"git init"}, cmds...)
	for _, cmd := range cmds {
		out, err := GitCommand(dir, "bash", "-c", cmd).CombinedOutput()
		if err != nil {
			t.Fatalf("Command %q failed. Output was:\n\n%s", cmd, out)
		}
	}
	return dir
}

func GitCommand(dir, name string, args ...string) *exec.Cmd {
	c := exec.Command(name, args...)
	c.Dir = dir
	c.Env = append(os.Environ(), "GIT_CONFIG="+path.Join(dir, ".git", "config"))
	return c
}
