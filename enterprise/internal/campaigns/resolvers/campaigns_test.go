package resolvers

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestCampaignsPermissionLevels(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()

	dbtesting.SetupGlobalTestDB(t)
	rcache.SetupForTest(t)

	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := func() time.Time {
		return now.UTC().Truncate(time.Microsecond)
	}

	// We need to enable read access so that non-site-admin users can access
	// the API and we can check for their admin rights.
	readAccessEnabled := true
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{
		CampaignsReadAccessEnabled: &readAccessEnabled,
	}})
	defer conf.Mock(nil)

	adminID := insertTestUser(t, dbconn.Global, "perm-level-admin", true)
	userID := insertTestUser(t, dbconn.Global, "perm-level-user", false)

	createCampaign := func(t *testing.T, s *ee.Store, name string, userID int32) int64 {
		t.Helper()
		c := &campaigns.Campaign{Name: name, AuthorID: userID, NamespaceUserID: userID}
		err := s.CreateCampaign(ctx, c)
		if err != nil {
			t.Fatal(err)
		}

		job := &campaigns.ChangesetJob{CampaignID: c.ID, PatchID: 999, Error: "This is an error"}
		if err := s.CreateChangesetJob(ctx, job); err != nil {
			t.Fatal(err)
		}

		return c.ID
	}

	t.Run("queries", func(t *testing.T) {
		// Wrap everything in the store in a transaction, so that the foreign-key
		// constraints are deferred
		tx := dbtest.NewTx(t, dbconn.Global)
		store := ee.NewStoreWithClock(tx, clock)
		sr := &Resolver{store: store}
		s, err := graphqlbackend.NewSchema(sr, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		adminCampaign := createCampaign(t, store, "admin", adminID)
		userCampaign := createCampaign(t, store, "user", userID)

		tests := []struct {
			name                    string
			currentUser             int32
			campaign                int64
			wantViewerCanAdminister bool
			wantErrors              []string
		}{
			{
				name:                    "site-admin viewing own campaign",
				currentUser:             adminID,
				campaign:                adminCampaign,
				wantViewerCanAdminister: true,
				wantErrors:              []string{"This is an error"},
			},
			{
				name:                    "non-site-admin viewing other's campaign",
				currentUser:             userID,
				campaign:                adminCampaign,
				wantViewerCanAdminister: false,
				wantErrors:              []string{},
			},
			{
				name:                    "site-admin viewing other's campaign",
				currentUser:             adminID,
				campaign:                userCampaign,
				wantViewerCanAdminister: true,
				wantErrors:              []string{"This is an error"},
			},
			{
				name:                    "non-site-admin viewing own campaign",
				currentUser:             userID,
				campaign:                userCampaign,
				wantViewerCanAdminister: true,
				wantErrors:              []string{"This is an error"},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				graphqlID := string(campaigns.MarshalCampaignID(tc.campaign))

				var queriedCampaign struct{ Node apitest.Campaign }

				input := map[string]interface{}{"campaign": graphqlID}
				queryCampaign := `query($campaign: ID!) {
				node(id: $campaign) { ... on Campaign { id, viewerCanAdminister, status { errors } } }
			}`

				actorCtx := actor.WithActor(ctx, actor.FromUser(tc.currentUser))
				apitest.MustExec(actorCtx, t, s, input, &queriedCampaign, queryCampaign)

				if have, want := queriedCampaign.Node.ID, graphqlID; have != want {
					t.Fatalf("queried campaign has wrong id %q, want %q", have, want)
				}
				if have, want := queriedCampaign.Node.ViewerCanAdminister, tc.wantViewerCanAdminister; have != want {
					t.Fatalf("queried campaign's ViewerCanAdminister is wrong %t, want %t", have, want)
				}
				if diff := cmp.Diff(queriedCampaign.Node.Status.Errors, tc.wantErrors); diff != "" {
					t.Fatalf("queried campaign's Errors is wrong: %s", diff)
				}
			})
		}
	})

	t.Run("mutations", func(t *testing.T) {
		mutations := []struct {
			name string
			tmpl string
		}{
			{"closeCampaign", `mutation { closeCampaign(campaign: %q, closeChangesets: false) { id } }`},
			{"deleteCampaign", `mutation { deleteCampaign(campaign: %q, closeChangesets: false) { alwaysNil } }`},
			{"retryCampaign", `mutation { retryCampaign(campaign: %q) { id } }`},
			{"publishCampaign", `mutation { publishCampaign(campaign: %q) { id } }`},
			{"updateCampaign", `mutation { updateCampaign(input: {id: %q, name: "new name"}) { id } }`},
			// TODO: publishChangeset
			// TODO: addChangesetsToCampaign
			// TODO: syncChangeset
		}

		for _, m := range mutations {
			t.Run(m.name, func(t *testing.T) {
				tests := []struct {
					currentUser    int32
					campaignAuthor int32
					wantAuthErr    bool
				}{
					{currentUser: userID, campaignAuthor: adminID, wantAuthErr: true},
					{currentUser: userID, campaignAuthor: userID, wantAuthErr: false},
				}

				for _, tc := range tests {
					tx := dbtest.NewTx(t, dbconn.Global)
					store := ee.NewStoreWithClock(tx, clock)
					sr := &Resolver{store: store}
					s, err := graphqlbackend.NewSchema(sr, nil, nil)
					if err != nil {
						t.Fatal(err)
					}

					c := createCampaign(t, store, "test-campaign", tc.campaignAuthor)

					mutation := fmt.Sprintf(m.tmpl, string(campaigns.MarshalCampaignID(c)))
					actorCtx := actor.WithActor(ctx, actor.FromUser(tc.currentUser))

					var response struct{}
					errs := apitest.Exec(actorCtx, t, s, nil, &response, mutation)

					if tc.wantAuthErr {
						if len(errs) != 1 {
							t.Fatalf("expected 1 error, but got %d: %s", len(errs), errs)
						}
						if !strings.Contains(errs[0].Error(), "must be authenticated") {
							t.Fatalf("wrong error: %s %T", errs[0], errs[0])
						}
					} else {
						// We don't care about other errors, we only want to
						// check that we didn't get an auth error.
						for _, e := range errs {
							if strings.Contains(e.Error(), "must be authenticated") {
								t.Fatalf("auth error wrongly returned: %s %T", errs[0], errs[0])
							}
						}
					}

				}
			})
		}
	})
}

func insertTestUser(t *testing.T, db *sql.DB, name string, isAdmin bool) (userID int32) {
	t.Helper()

	q := sqlf.Sprintf("INSERT INTO users (username, site_admin) VALUES (%s, %t) RETURNING id", name, isAdmin)

	err := db.QueryRow(q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	return userID
}
