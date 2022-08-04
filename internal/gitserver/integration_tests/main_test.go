package inttests

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/sync/semaphore"

	sglog "github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

var root string

// This is a default gitserver test client currently used for RequestRepoUpdate
// gitserver calls during invocation of MakeGitRepository function
var testGitserverClient gitserver.Client
var gitserverAddresses []string

func TestMain(m *testing.M) {
	flag.Parse()

	if !testing.Verbose() {
		logtest.InitWithLevel(m, sglog.LevelNone)
	}

	code := m.Run()

	_ = os.RemoveAll(root)

	os.Exit(code)
}

// done in init since the go vet analysis "ctrlflow" is tripped up if this is
// done as part of TestMain.
func init() {
	// Ignore users configuration in tests
	os.Setenv("GIT_CONFIG_NOSYSTEM", "true")
	os.Setenv("HOME", "/dev/null")

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("listen failed: %s", err)
	}

	root, err = os.MkdirTemp("", "test")
	if err != nil {
		log.Fatal(err)
	}

	db := database.NewMockDB()
	gr := database.NewMockGitserverRepoStore()
	db.GitserverReposFunc.SetDefaultReturn(gr)

	srv := &http.Server{
		Handler: (&server.Server{
			Logger:   sglog.Scoped("server", "the gitserver service"),
			ReposDir: filepath.Join(root, "repos"),
			GetRemoteURLFunc: func(ctx context.Context, name api.RepoName) (string, error) {
				return filepath.Join(root, "remotes", string(name)), nil
			},
			GetVCSSyncer: func(ctx context.Context, name api.RepoName) (server.VCSSyncer, error) {
				return &server.GitRepoSyncer{}, nil
			},
			GlobalBatchLogSemaphore: semaphore.NewWeighted(32),
			DB:                      db,
		}).Handler(),
	}
	go func() {
		if err := srv.Serve(l); err != nil {
			log.Fatal(err)
		}
	}()

	serverAddress := l.Addr().String()
	testGitserverClient = gitserver.NewTestClient(httpcli.InternalDoer, db, []string{serverAddress})
	gitserverAddresses = []string{serverAddress}
}

var Times = []string{
	AppleTime("2006-01-02T15:04:05Z"),
	AppleTime("2014-05-06T19:20:21Z"),
}

// InitGitRepository initializes a new Git repository and runs cmds in a new
// temporary directory (returned as dir).
func InitGitRepository(t testing.TB, cmds ...string) string {
	t.Helper()
	remotes := filepath.Join(root, "remotes")
	if err := os.MkdirAll(remotes, 0700); err != nil {
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

func AppleTime(t string) string {
	ti, _ := time.Parse(time.RFC3339, t)
	return ti.Local().Format("200601021504.05")
}
