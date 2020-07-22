package authz

import (
	"context"
	"database/sql"
	"flag"
	"net/url"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/api"
	authzGitHub "github.com/sourcegraph/sourcegraph/internal/authz/github"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	extsvcGitHub "github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
)

var updateRegex = flag.String("update", "", "Update testdata of tests matching the given regex")

func update(name string) bool {
	if updateRegex == nil || *updateRegex == "" {
		return false
	}
	return regexp.MustCompile(*updateRegex).MatchString(name)
}

var dsn = flag.String("dsn", "", "Database connection string to use in integration tests")

// This integration test performs a repository-centric permissions syncing against
// https://github.com, then check if permissions are correctly granted for the test
// user "sourcegraph-vcr-bob", who is a outside collaborator of the repository
// "sourcegraph-vcr-repos/private-org-repo-1".
//
// NOTE: To update VCR for this test, please use the token of "sourcegraph-vcr"
// for GITHUB_TOKEN, which can be found in 1Password.
func TestIntegration_GitHubPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	const name = "Integration_GitHubPermissions"
	cf, save := httptestutil.NewGitHubRecorderFactory(t, update(name), name)
	defer save()

	uri, err := url.Parse("https://github.com")
	if err != nil {
		t.Fatal(err)
	}

	doer, err := cf.Doer()
	if err != nil {
		t.Fatal(err)
	}

	token := os.Getenv("GITHUB_TOKEN")
	cli := extsvcGitHub.NewClient(uri, token, doer)

	provider := authzGitHub.NewProvider("extsvc:github:1", uri, token, cli)

	authz.SetProviders(false, []authz.Provider{provider})
	defer authz.SetProviders(true, nil)

	testDB := dbtest.NewDB(t, *dsn)
	ctx := context.Background()

	// Set up repository, user, and user external account
	q := sqlf.Sprintf(`
INSERT INTO repo (id, name, description, language, fork, private,
					uri, external_service_type, external_service_id, sources)
	VALUES (1, 'github.com/sourcegraph-vcr-repos/private-org-repo-1', '', '', FALSE, TRUE,
			'github.com/sourcegraph-vcr-repos/private-org-repo-1', 'github', 'https://github.com/', '{"extsvc:github:1": {}}'::jsonb)`)
	_, err = testDB.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		t.Fatal(err)
	}

	dbconn.Global = testDB
	newUser := db.NewUser{
		Email:           "sourcegraph-vcr-bob@sourcegraph.com",
		Username:        "sourcegraph-vcr-bob",
		EmailIsVerified: true,
	}
	spec := extsvc.AccountSpec{
		ServiceType: extsvc.TypeGitHub,
		ServiceID:   "https://github.com/",
		AccountID:   "66464926",
	}
	userID, err := db.ExternalAccounts.CreateUserAndSave(ctx, newUser, spec, extsvc.AccountData{})
	if err != nil {
		t.Fatal(err)
	}

	clock := func() time.Time {
		return time.Now().UTC().Truncate(time.Microsecond)
	}
	reposStore := repos.NewDBStore(testDB, sql.TxOptions{})
	permsStore := edb.NewPermsStore(testDB, clock)
	syncer := NewPermsSyncer(reposStore, permsStore, clock, nil)

	err = syncer.syncRepoPerms(ctx, api.RepoID(1), true)
	if err != nil {
		t.Fatal(err)
	}

	p := &authz.UserPermissions{
		UserID: userID,
		Perm:   authz.Read,
		Type:   authz.PermRepos,
	}
	err = permsStore.LoadUserPermissions(ctx, p)
	if err != nil {
		t.Fatal(err)
	}

	wantIDs := []uint32{1}
	if diff := cmp.Diff(wantIDs, p.IDs.ToArray()); diff != "" {
		t.Fatalf("IDs mismatch (-want +got):\n%s", diff)
	}
}
