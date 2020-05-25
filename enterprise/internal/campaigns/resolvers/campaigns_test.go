package resolvers

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestCampaignsPermissionLevels(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

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

	store := ee.NewStoreWithClock(dbconn.Global, clock)
	sr := &Resolver{store: store}

	s, err := graphqlbackend.NewSchema(sr, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	adminID := insertTestUser(t, dbconn.Global, "perm-level-admin", true)
	userID := insertTestUser(t, dbconn.Global, "perm-level-user", false)

	adminCampaign := &campaigns.Campaign{
		Name:            "Admin",
		AuthorID:        adminID,
		NamespaceUserID: adminID,
	}
	err = store.CreateCampaign(ctx, adminCampaign)
	if err != nil {
		t.Fatal(err)
	}

	userCampaign := &campaigns.Campaign{
		Name:            "User campaign",
		AuthorID:        userID,
		NamespaceUserID: userID,
	}

	err = store.CreateCampaign(ctx, adminCampaign)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name                    string
		currentUser             int32
		campaign                int64
		wantViewerCanAdminister bool
	}{
		{
			name:                    "site-admin viewing own campaign",
			currentUser:             adminID,
			campaign:                adminCampaign.ID,
			wantViewerCanAdminister: true,
		},
		{
			name:                    "non-site-admin viewing other's campaign",
			currentUser:             userID,
			campaign:                adminCampaign.ID,
			wantViewerCanAdminister: false,
		},
		{
			name:                    "site-admin viewing other's campaign",
			currentUser:             adminID,
			campaign:                userCampaign.ID,
			wantViewerCanAdminister: true,
		},
		{
			name:                    "non-site-admin viewing own campaign",
			currentUser:             userID,
			campaign:                userCampaign.ID,
			wantViewerCanAdminister: false,
		},
	}

	for _, tc := range tests {
		graphqlID := string(marshalCampaignID(tc.campaign))
		var queriedCampaign struct{ Node apitest.Campaign }

		input := map[string]interface{}{"campaign": graphqlID}
		queryCampaign := `
		query($campaign: ID!){
			node(id: $campaign) {
				... on Campaign {
					id, viewerCanAdminister
				}
			}
		}
	`

		actorCtx := actor.WithActor(ctx, actor.FromUser(tc.currentUser))
		apitest.MustExec(actorCtx, t, s, input, &queriedCampaign, queryCampaign)

		if have, want := queriedCampaign.Node.ID, graphqlID; have != want {
			t.Fatalf("queried campaign has wrong id %q, want %q", have, want)
		}
		if have, want := queriedCampaign.Node.ViewerCanAdminister, tc.wantViewerCanAdminister; have != want {
			t.Fatalf("queried campaign's ViewerCanAdminister is wrong %t, want %t", have, want)
		}
	}
}

const campaignFragment = `
fragment u on User { id, databaseID, siteAdmin }
fragment c on Campaign {
	id, viewerCanAdminister
	author { ... u }
	namespace { ... on User { ... u } }
}
`

func insertTestUser(t *testing.T, db *sql.DB, name string, isAdmin bool) (userID int32) {
	t.Helper()

	q := sqlf.Sprintf("INSERT INTO users (username, site_admin) VALUES (%s, %t) RETURNING id", name, isAdmin)

	err := db.QueryRow(q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	return userID
}
