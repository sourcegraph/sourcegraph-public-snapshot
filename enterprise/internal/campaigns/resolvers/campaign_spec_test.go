package resolvers

import (
	"context"
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

func TestCampaignSpecResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)

	userID := insertTestUser(t, dbconn.Global, "campaign-spec-by-id", false)

	spec := &campaigns.CampaignSpec{
		RawSpec: ct.TestRawCampaignSpec,
		Spec: campaigns.CampaignSpecFields{
			Name:        "Foobar",
			Description: "My description",
			ChangesetTemplate: campaigns.ChangesetTemplate{
				Title:  "Hello there",
				Body:   "This is the body",
				Branch: "my-branch",
				Commit: campaigns.CommitTemplate{
					Message: "Add hello world",
				},
				Published: false,
			},
		},
		UserID:          userID,
		NamespaceUserID: userID,
	}

	store := ee.NewStore(dbconn.Global)
	if err := store.CreateCampaignSpec(ctx, spec); err != nil {
		t.Fatal(err)
	}

	s, err := graphqlbackend.NewSchema(&Resolver{store: store}, nil, nil)
	if err != nil {
		t.Fatal(err)

	}

	apiID := string(marshalCampaignSpecRandID(spec.RandID))
	userApiID := string(graphqlbackend.MarshalUserID(userID))

	want := apitest.CampaignSpec{
		Typename:      "CampaignSpec",
		ID:            apiID,
		OriginalInput: spec.RawSpec,
		ParsedInput: apitest.CampaignSpecParsedInput{
			Name:        spec.Spec.Name,
			Description: spec.Spec.Description,
			ChangesetTemplate: apitest.ChangesetTemplate{
				Title:  spec.Spec.ChangesetTemplate.Title,
				Body:   spec.Spec.ChangesetTemplate.Body,
				Branch: spec.Spec.ChangesetTemplate.Branch,
				Commit: apitest.CommitTemplate{
					Message: spec.Spec.ChangesetTemplate.Commit.Message,
				},
				Published: spec.Spec.ChangesetTemplate.Published,
			},
		},
		PreviewURL: "/campaigns/new?spec=" + apiID,
		Namespace:  apitest.UserOrg{ID: userApiID, DatabaseID: userID},
		Creator:    apitest.User{ID: userApiID, DatabaseID: userID},
		CreatedAt:  &graphqlbackend.DateTime{Time: spec.CreatedAt.Truncate(time.Second)},
		ExpiresAt:  &graphqlbackend.DateTime{Time: spec.CreatedAt.Truncate(time.Second).Add(2 * time.Hour)},
	}

	input := map[string]interface{}{"campaignSpec": apiID}
	var response struct{ Node apitest.CampaignSpec }
	apitest.MustExec(ctx, t, s, input, &response, queryCampaignSpecNode)

	if diff := cmp.Diff(want, response.Node); diff != "" {
		t.Fatalf("unexpected response (-want +got):\n%s", diff)
	}
}

const queryCampaignSpecNode = `
fragment u on User { id, databaseID, siteAdmin }
fragment o on Org  { id, name }

query($campaignSpec: ID!) {
  node(id: $campaignSpec) {
    __typename

    ... on CampaignSpec {
      id
      originalInput
      parsedInput

      creator  { ...u }
      namespace {
        ... on User { ...u }
        ... on Org  { ...o }
      }

      previewURL

      createdAt
      expiresAt
    }
  }
}
`
