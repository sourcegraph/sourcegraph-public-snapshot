package inttests

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/semaphore"

	sglog "github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server"
	srp "github.com/sourcegraph/sourcegraph/enterprise/internal/authz/subrepoperms"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/schema"
)

var root string

var gitserverAddresses []string

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
		}).Handler(),
	}
	go func() {
		if err := srv.Serve(l); err != nil {
			log.Fatal(err)
		}
	}()

	serverAddress := l.Addr().String()
	gitserverAddresses = []string{serverAddress}
}

func TestReadDir_SubRepoFiltering(t *testing.T) {
	ctx := actor.WithActor(context.Background(), &actor.Actor{
		UID: 1,
	})
	gitCommands := []string{
		"touch file1",
		"git add file1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"mkdir app",
		"touch app/file2",
		"git add app",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit2 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	repo := gitserver.MakeGitRepository(t, gitCommands...)
	commitID := api.CommitID("b1c725720de2bbd0518731b4a61959797ff345f3")
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				SubRepoPermissions: &schema.SubRepoPermissions{
					Enabled: true,
				},
			},
		},
	})
	defer conf.Mock(nil)
	srpGetter := database.NewMockSubRepoPermsStore()
	testSubRepoPerms := map[api.RepoName]authz.SubRepoPermissions{
		repo: {
			Paths: []string{"/**", "-/app/**"},
		},
	}
	srpGetter.GetByUserFunc.SetDefaultReturn(testSubRepoPerms, nil)
	checker, err := srp.NewSubRepoPermsClient(srpGetter)
	if err != nil {
		t.Fatalf("unexpected error creating sub-repo perms client: %s", err)
	}

	db := database.NewMockDB()
	gr := database.NewMockGitserverRepoStore()
	db.GitserverReposFunc.SetDefaultReturn(gr)
	client := gitserver.NewTestClient(http.DefaultClient, db, gitserverAddresses)
	files, err := client.ReadDir(ctx, checker, repo, commitID, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// Because we have a wildcard matcher we still allow directory visibility
	assert.Len(t, files, 1)
	assert.Equal(t, "file1", files[0].Name())
	assert.False(t, files[0].IsDir())
}
