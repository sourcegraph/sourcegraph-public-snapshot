package campaigns

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func init() {
	dbtesting.DBNameSuffix = "campaignsenterpriserdb"
}

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.DiscardHandler())
	}
	os.Exit(m.Run())
}

var createTestUser = func() func(*testing.T, bool) *types.User {
	count := 0

	// This function replicates the minium amount of work required by
	// db.Users.Create to create a new user, but it doesn't require passing in
	// a full db.NewUser every time.
	return func(t *testing.T, siteAdmin bool) *types.User {
		t.Helper()

		user := &types.User{
			Username:    fmt.Sprintf("testuser-%d", count),
			DisplayName: "testuser",
		}

		q := sqlf.Sprintf("INSERT INTO users (username, site_admin) VALUES (%s, %t) RETURNING id, site_admin", user.Username, siteAdmin)
		err := dbconn.Global.QueryRow(q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&user.ID, &user.SiteAdmin)
		if err != nil {
			t.Fatal(err)
		}

		if user.SiteAdmin != siteAdmin {
			t.Fatalf("user.SiteAdmin=%t, but expected is %t", user.SiteAdmin, siteAdmin)
		}

		_, err = dbconn.Global.Exec("INSERT INTO names(name, user_id) VALUES($1, $2)", user.Username, user.ID)
		if err != nil {
			t.Fatalf("failed to create name: %s", err)
		}

		count += 1

		return user
	}
}()

func truncateTables(t *testing.T, db *sql.DB, tables ...string) {
	t.Helper()

	_, err := db.Exec("TRUNCATE " + strings.Join(tables, ", ") + " RESTART IDENTITY")
	if err != nil {
		t.Fatal(err)
	}
}

type testSpecOpts struct {
	user         int32
	repo         api.RepoID
	campaignSpec int64

	// If this is non-blank, the changesetSpec will be an import/track spec for
	// the changeset with the given externalID in the given repo.
	externalID string

	// If this is set, the changesetSpec will be a "create commit on this
	// branch" changeset spec.
	headRef string

	// If this is set along with headRef, the changesetSpec will have published
	// set.
	published interface{}

	title             string
	body              string
	commitMessage     string
	commitDiff        string
	commitAuthorEmail string
	commitAuthorName  string
}

var testChangsetSpecDiffStat = &diff.Stat{Added: 10, Changed: 5, Deleted: 2}

func buildChangesetSpec(t *testing.T, opts testSpecOpts) *campaigns.ChangesetSpec {
	t.Helper()

	published := campaigns.PublishedValue{Val: opts.published}
	if opts.published == nil {
		// Set false as the default.
		published.Val = false
	}
	if !published.Valid() {
		t.Fatalf("invalid value for published passed, got %v (%T)", opts.published, opts.published)
	}

	spec := &campaigns.ChangesetSpec{
		UserID:         opts.user,
		RepoID:         opts.repo,
		CampaignSpecID: opts.campaignSpec,
		Spec: &campaigns.ChangesetSpecDescription{
			BaseRepository: graphqlbackend.MarshalRepositoryID(opts.repo),

			ExternalID: opts.externalID,
			HeadRef:    opts.headRef,
			Published:  published,

			Title: opts.title,
			Body:  opts.body,

			Commits: []campaigns.GitCommitDescription{
				{
					Message:     opts.commitMessage,
					Diff:        opts.commitDiff,
					AuthorEmail: opts.commitAuthorEmail,
					AuthorName:  opts.commitAuthorName,
				},
			},
		},
		DiffStatAdded:   testChangsetSpecDiffStat.Added,
		DiffStatChanged: testChangsetSpecDiffStat.Changed,
		DiffStatDeleted: testChangsetSpecDiffStat.Deleted,
	}

	return spec
}

func createChangesetSpec(
	t *testing.T,
	ctx context.Context,
	store *store.Store,
	opts testSpecOpts,
) *campaigns.ChangesetSpec {
	t.Helper()

	spec := buildChangesetSpec(t, opts)

	if err := store.CreateChangesetSpec(ctx, spec); err != nil {
		t.Fatal(err)
	}

	return spec
}

func createCampaignSpec(t *testing.T, ctx context.Context, store *store.Store, name string, userID int32) *campaigns.CampaignSpec {
	t.Helper()

	s := &campaigns.CampaignSpec{
		UserID:          userID,
		NamespaceUserID: userID,
		Spec: campaigns.CampaignSpecFields{
			Name:        name,
			Description: "the description",
			ChangesetTemplate: campaigns.ChangesetTemplate{
				Branch: "branch-name",
			},
		},
	}

	if err := store.CreateCampaignSpec(ctx, s); err != nil {
		t.Fatal(err)
	}

	return s
}

func createCampaign(t *testing.T, ctx context.Context, store *store.Store, name string, userID int32, spec int64) *campaigns.Campaign {
	t.Helper()

	c := &campaigns.Campaign{
		InitialApplierID: userID,
		LastApplierID:    userID,
		LastAppliedAt:    store.Clock()(),
		NamespaceUserID:  userID,
		CampaignSpecID:   spec,
		Name:             name,
		Description:      "campaign description",
	}

	if err := store.CreateCampaign(ctx, c); err != nil {
		t.Fatal(err)
	}

	return c
}
