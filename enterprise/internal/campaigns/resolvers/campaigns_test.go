package resolvers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/resolvers/apitest"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/testing"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestCampaignResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)

	username := "campaign-resolver-username"
	userID := insertTestUser(t, dbconn.Global, username, true)

	store := ee.NewStore(dbconn.Global)

	now := time.Now().UTC().Truncate(time.Microsecond)

	campaignSpec := &campaigns.CampaignSpec{
		RawSpec:         ct.TestRawCampaignSpec,
		UserID:          userID,
		NamespaceUserID: userID,
	}
	if err := store.CreateCampaignSpec(ctx, campaignSpec); err != nil {
		t.Fatal(err)
	}

	campaign := &campaigns.Campaign{
		Name:             "my-unique-name",
		Description:      "The campaign description",
		NamespaceUserID:  userID,
		InitialApplierID: userID,
		LastApplierID:    userID,
		LastAppliedAt:    now,
		CampaignSpecID:   campaignSpec.ID,
	}
	if err := store.CreateCampaign(ctx, campaign); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("campaign=%+v\n", campaign)

	s, err := graphqlbackend.NewSchema(&Resolver{store: store}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	campaignApiID := string(campaigns.MarshalCampaignID(campaign.ID))

	input := map[string]interface{}{"campaign": campaignApiID}
	var response struct{ Node apitest.Campaign }
	apitest.MustExec(ctx, t, s, input, &response, queryCampaign)

	wantCampaign := apitest.Campaign{
		ID:             campaignApiID,
		Name:           campaign.Name,
		Description:    campaign.Description,
		Namespace:      apitest.UserOrg{DatabaseID: userID, SiteAdmin: true},
		InitialApplier: apitest.User{DatabaseID: userID, SiteAdmin: true},
		LastApplier:    apitest.User{DatabaseID: userID, SiteAdmin: true},
		LastAppliedAt:  marshalDateTime(t, now),
		URL:            fmt.Sprintf("/users/%s/campaigns/%s", username, campaignApiID),
	}
	if diff := cmp.Diff(wantCampaign, response.Node); diff != "" {
		t.Fatalf("wrong campaign response (-want +got):\n%s", diff)
	}
}

const queryCampaign = `
fragment u on User { databaseID, siteAdmin }
fragment o on Org  { name }

query($campaign: ID!){
  node(id: $campaign) {
    ... on Campaign {
      id, name, description
      initialApplier { ...u }
      lastApplier    { ...u }
      lastAppliedAt
      namespace {
        ... on User { ...u }
        ... on Org  { ...o }
      }
      url
    }
  }
}
`
