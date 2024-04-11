package batches

import (
	"context"
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/api"
	bstore "github.com/sourcegraph/sourcegraph/internal/batches/store"
	bt "github.com/sourcegraph/sourcegraph/internal/batches/testing"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"

	bbcs "github.com/sourcegraph/sourcegraph/internal/batches/sources/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestExternalForkNameMigrator(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	s := bstore.New(db, observation.TestContextTB(t), nil)

	migrator := NewExternalForkNameMigrator(s.Store, 100)
	progress, err := migrator.Progress(ctx, false)
	assert.NoError(t, err)

	if have, want := progress, 1.0; have != want {
		t.Fatalf("got invalid progress with no DB entries, want=%f have=%f", want, have)
	}

	rs := database.ReposWith(logger, s)
	es := database.ExternalServicesWith(logger, s)
	ghrepo := bt.TestRepo(t, es, extsvc.KindGitHub)
	glrepo := bt.TestRepo(t, es, extsvc.KindGitLab)
	bbsrepo := bt.TestRepo(t, es, extsvc.KindBitbucketServer)
	bbcrepo := bt.TestRepo(t, es, extsvc.KindBitbucketCloud)

	if err := rs.Create(ctx, ghrepo, glrepo, bbsrepo, bbcrepo); err != nil {
		t.Fatal(err)
	}

	testData := []struct {
		extID            string
		extSvcType       string
		repoID           api.RepoID
		extForkNamespace string
		extForkName      string
		extDeleted       bool
		metadata         any
		wantExtForkName  string
	}{
		// Changesets on GitHub/GitLab should not be migrated.
		{
			extID:            "gh1",
			extSvcType:       extsvc.TypeGitHub,
			repoID:           ghrepo.ID,
			extForkNamespace: "user",
			extForkName:      "",
			extDeleted:       false,
			metadata:         nil,
			wantExtForkName:  "",
		},
		{
			extID:            "gl1",
			extSvcType:       extsvc.TypeGitLab,
			repoID:           glrepo.ID,
			extForkNamespace: "user",
			extForkName:      "",
			extDeleted:       false,
			metadata:         nil,
			wantExtForkName:  "",
		},
		// A changeset on Bitbucket Server/Cloud that is not on a fork should not be migrated.
		{
			extID:            "bbs1",
			extSvcType:       extsvc.TypeBitbucketServer,
			repoID:           bbsrepo.ID,
			extForkNamespace: "",
			extForkName:      "",
			extDeleted:       true,
			metadata:         nil,
			wantExtForkName:  "",
		},
		{
			extID:            "bbc1",
			extSvcType:       extsvc.TypeBitbucketCloud,
			repoID:           bbcrepo.ID,
			extForkNamespace: "",
			extForkName:      "",
			extDeleted:       true,
			metadata:         nil,
			wantExtForkName:  "",
		},
		// A changeset on Bitbucket Server/Cloud that already has a fork name should not be migrated.
		{
			extID:            "bbs2",
			extSvcType:       extsvc.TypeBitbucketServer,
			repoID:           bbsrepo.ID,
			extForkNamespace: "user",
			extForkName:      "my-fork-name",
			extDeleted:       false,
			metadata:         nil,
			wantExtForkName:  "my-fork-name",
		},
		{
			extID:            "bbc2",
			extSvcType:       extsvc.TypeBitbucketCloud,
			repoID:           bbcrepo.ID,
			extForkNamespace: "user",
			extForkName:      "my-fork-name",
			extDeleted:       false,
			metadata:         nil,
			wantExtForkName:  "my-fork-name",
		},
		// A changeset on Bitbucket Server/Cloud that was deleted on the code host should not be migrated.
		{
			extID:            "bbs3",
			extSvcType:       extsvc.TypeBitbucketServer,
			repoID:           bbsrepo.ID,
			extForkNamespace: "user",
			extForkName:      "",
			extDeleted:       true,
			metadata:         nil,
			wantExtForkName:  "",
		},
		{
			extID:            "bbc3",
			extSvcType:       extsvc.TypeBitbucketCloud,
			repoID:           bbcrepo.ID,
			extForkNamespace: "user",
			extForkName:      "",
			extDeleted:       true,
			metadata:         nil,
			wantExtForkName:  "",
		},
		// A changeset on Bitbucket Server/Cloud that has a fork namespace and no fork name should be migrated.
		{
			extID:            "bbs4",
			extSvcType:       extsvc.TypeBitbucketServer,
			repoID:           bbsrepo.ID,
			extForkNamespace: "user",
			extForkName:      "",
			extDeleted:       false,
			metadata: &bitbucketserver.PullRequest{
				FromRef: bitbucketserver.Ref{Repository: bitbucketserver.RefRepository{Slug: "my-bbs-fork-name"}},
			},
			wantExtForkName: "my-bbs-fork-name",
		},
		{
			extID:            "bbc4",
			extSvcType:       extsvc.TypeBitbucketCloud,
			repoID:           bbcrepo.ID,
			extForkNamespace: "user",
			extForkName:      "",
			extDeleted:       false,
			metadata: &bbcs.AnnotatedPullRequest{
				PullRequest: &bitbucketcloud.PullRequest{
					Source: bitbucketcloud.PullRequestEndpoint{Repo: bitbucketcloud.Repo{Name: "my-bbc-fork-name"}},
				},
			},
			wantExtForkName: "my-bbc-fork-name",
		},
	}

	for _, tc := range testData {
		cs := bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
			ExternalServiceType:   tc.extSvcType,
			ExternalID:            tc.extID,
			Repo:                  tc.repoID,
			ExternalForkNamespace: tc.extForkNamespace,
			ExternalForkName:      tc.extForkName,
			Metadata:              tc.metadata,
		})

		if tc.extDeleted {
			bt.DeleteChangeset(t, ctx, s, cs)
		}
	}

	count, _, err := basestore.ScanFirstInt(s.Query(ctx, sqlf.Sprintf("SELECT count(*) FROM changesets")))
	if err != nil {
		t.Fatal(err)
	}
	if count != 10 {
		t.Fatalf("got %d changesets, want %d", count, 10)
	}

	progress, err = migrator.Progress(ctx, false)
	assert.NoError(t, err)

	// We expect to start with progress at 50% because 2 of the 4 changesets on forks on
	// Bitbucket Server/Cloud already have a fork name set.
	if have, want := progress, 0.5; have != want {
		t.Fatalf("got invalid progress with unmigrated entries, want=%f have=%f", want, have)
	}

	if err := migrator.Up(ctx); err != nil {
		t.Fatal(err)
	}

	progress, err = migrator.Progress(ctx, false)
	assert.NoError(t, err)

	if have, want := progress, 1.0; have != want {
		t.Fatalf("got invalid progress after up migration, want=%f have=%f", want, have)
	}

	for _, tc := range testData {
		// Check that we can find the empty spec with its new ID.
		cs, err := s.GetChangeset(ctx, bstore.GetChangesetOpts{ExternalID: tc.extID, ExternalServiceType: tc.extSvcType})

		if err != nil {
			t.Fatalf("could not find changeset with external ID %s after migration", tc.extID)
		}
		if tc.wantExtForkName != "" && cs.ExternalForkName != tc.wantExtForkName {
			t.Fatalf("changeset with external id %s has wrong fork name. got %q, want %q", tc.extID, cs.ExternalForkName, tc.wantExtForkName)
		} else if tc.wantExtForkName == "" && cs.ExternalForkName != "" {
			t.Fatalf("changeset with external id %s has wrong fork name. got %q, want empty string", tc.extID, cs.ExternalForkName)
		}
	}
}
