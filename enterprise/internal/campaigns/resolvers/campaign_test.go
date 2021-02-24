package resolvers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/store"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/testing"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestCampaignResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	db := dbtesting.GetDB(t)

	userID := ct.CreateTestUser(t, db, true).ID
	orgName := "test-campaign-resolver-org"
	orgID := ct.InsertTestOrg(t, db, orgName)

	now := timeutil.Now()
	clock := func() time.Time { return now }
	cstore := store.NewWithClock(db, clock)

	campaignSpec := &campaigns.CampaignSpec{
		RawSpec:        ct.TestRawCampaignSpec,
		UserID:         userID,
		NamespaceOrgID: orgID,
	}
	if err := cstore.CreateCampaignSpec(ctx, campaignSpec); err != nil {
		t.Fatal(err)
	}

	campaign := &campaigns.Campaign{
		Name:             "my-unique-name",
		Description:      "The campaign description",
		NamespaceOrgID:   orgID,
		InitialApplierID: userID,
		LastApplierID:    userID,
		LastAppliedAt:    now,
		CampaignSpecID:   campaignSpec.ID,
	}
	if err := cstore.CreateCampaign(ctx, campaign); err != nil {
		t.Fatal(err)
	}

	s, err := graphqlbackend.NewSchema(db, &Resolver{store: cstore}, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	campaignAPIID := string(marshalCampaignID(campaign.ID))
	namespaceAPIID := string(graphqlbackend.MarshalOrgID(orgID))
	apiUser := &apitest.User{DatabaseID: userID, SiteAdmin: true}
	wantCampaign := apitest.Campaign{
		ID:             campaignAPIID,
		Name:           campaign.Name,
		Description:    campaign.Description,
		Namespace:      apitest.UserOrg{ID: namespaceAPIID, Name: orgName},
		InitialApplier: apiUser,
		LastApplier:    apiUser,
		SpecCreator:    apiUser,
		LastAppliedAt:  marshalDateTime(t, now),
		URL:            fmt.Sprintf("/organizations/%s/campaigns/%s", orgName, campaign.Name),
		CreatedAt:      marshalDateTime(t, now),
		UpdatedAt:      marshalDateTime(t, now),
		// Not closed.
		ClosedAt: "",
	}

	input := map[string]interface{}{"campaign": campaignAPIID}
	{
		var response struct{ Node apitest.Campaign }
		apitest.MustExec(ctx, t, s, input, &response, queryCampaign)

		if diff := cmp.Diff(wantCampaign, response.Node); diff != "" {
			t.Fatalf("wrong campaign response (-want +got):\n%s", diff)
		}
	}
	// Test resolver by namespace and name
	byNameInput := map[string]interface{}{"name": campaign.Name, "namespace": namespaceAPIID}
	{
		var response struct{ Campaign apitest.Campaign }
		apitest.MustExec(ctx, t, s, byNameInput, &response, queryCampaignByName)

		if diff := cmp.Diff(wantCampaign, response.Campaign); diff != "" {
			t.Fatalf("wrong campaign response (-want +got):\n%s", diff)
		}
	}

	// Now soft-delete the user and check we can still access the campaign in the org namespace.
	err = database.UsersWith(cstore).Delete(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}

	wantCampaign.InitialApplier = nil
	wantCampaign.LastApplier = nil
	wantCampaign.SpecCreator = nil

	{
		var response struct{ Node apitest.Campaign }
		apitest.MustExec(ctx, t, s, input, &response, queryCampaign)

		if diff := cmp.Diff(wantCampaign, response.Node); diff != "" {
			t.Fatalf("wrong campaign response (-want +got):\n%s", diff)
		}
	}

	// Now hard-delete the user and check we can still access the campaign in the org namespace.
	err = database.UsersWith(cstore).HardDelete(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	{
		var response struct{ Node apitest.Campaign }
		apitest.MustExec(ctx, t, s, input, &response, queryCampaign)

		if diff := cmp.Diff(wantCampaign, response.Node); diff != "" {
			t.Fatalf("wrong campaign response (-want +got):\n%s", diff)
		}
	}
}

const queryCampaign = `
fragment u on User { databaseID, siteAdmin }
fragment o on Org  { id, name }

query($campaign: ID!){
  node(id: $campaign) {
    ... on Campaign {
      id, name, description
      initialApplier { ...u }
      lastApplier    { ...u }
      specCreator    { ...u }
      lastAppliedAt
      createdAt
      updatedAt
      closedAt
      namespace {
        ... on User { ...u }
        ... on Org  { ...o }
      }
      url
    }
  }
}
`
const queryCampaignByName = `
fragment u on User { databaseID, siteAdmin }
fragment o on Org  { id, name }

query($namespace: ID!, $name: String!){
  campaign(namespace: $namespace, name: $name) {
    id, name, description
    initialApplier { ...u }
    lastApplier    { ...u }
    specCreator    { ...u }
    lastAppliedAt
    createdAt
    updatedAt
    closedAt
    namespace {
      ... on User { ...u }
      ... on Org  { ...o }
    }
    url
  }
}
`
