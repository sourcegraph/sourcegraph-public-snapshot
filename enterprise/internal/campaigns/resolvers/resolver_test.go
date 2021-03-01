package resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/campaignutils/overridable"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/search"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/service"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/store"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/testing"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestNullIDResilience(t *testing.T) {
	db := dbtesting.GetDB(t)
	sr := New(store.New(db))

	s, err := graphqlbackend.NewSchema(db, sr, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := backend.WithAuthzBypass(context.Background())

	ids := []graphql.ID{
		marshalCampaignID(0),
		marshalChangesetID(0),
		marshalCampaignSpecRandID(""),
		marshalChangesetSpecRandID(""),
		marshalCampaignsCredentialID(0),
	}

	for _, id := range ids {
		var response struct{ Node struct{ ID string } }

		query := fmt.Sprintf(`query { node(id: %q) { id } }`, id)
		apitest.MustExec(ctx, t, s, nil, &response, query)

		if have, want := response.Node.ID, ""; have != want {
			t.Fatalf("node has wrong ID. have=%q, want=%q", have, want)
		}
	}

	mutations := []string{
		fmt.Sprintf(`mutation { closeCampaign(campaign: %q) { id } }`, marshalCampaignID(0)),
		fmt.Sprintf(`mutation { deleteCampaign(campaign: %q) { alwaysNil } }`, marshalCampaignID(0)),
		fmt.Sprintf(`mutation { syncChangeset(changeset: %q) { alwaysNil } }`, marshalChangesetID(0)),
		fmt.Sprintf(`mutation { reenqueueChangeset(changeset: %q) { id } }`, marshalChangesetID(0)),
		fmt.Sprintf(`mutation { applyCampaign(campaignSpec: %q) { id } }`, marshalCampaignSpecRandID("")),
		fmt.Sprintf(`mutation { createCampaign(campaignSpec: %q) { id } }`, marshalCampaignSpecRandID("")),
		fmt.Sprintf(`mutation { moveCampaign(campaign: %q, newName: "foobar") { id } }`, marshalCampaignID(0)),
		fmt.Sprintf(`mutation { createCampaignsCredential(externalServiceKind: GITHUB, externalServiceURL: "http://test", credential: "123123", user: %q) { id } }`, graphqlbackend.MarshalUserID(0)),
		fmt.Sprintf(`mutation { deleteCampaignsCredential(campaignsCredential: %q) { alwaysNil } }`, marshalCampaignsCredentialID(0)),
	}

	for _, m := range mutations {
		var response struct{}
		errs := apitest.Exec(ctx, t, s, nil, &response, m)
		if len(errs) == 0 {
			t.Fatalf("expected errors but none returned (mutation: %q)", m)
		}
		if have, want := errs[0].Error(), fmt.Sprintf("graphql: %s", ErrIDIsZero{}); have != want {
			t.Fatalf("wrong errors. have=%s, want=%s (mutation: %q)", have, want, m)
		}
	}
}

func TestCreateCampaignSpec(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	db := dbtesting.GetDB(t)

	user := ct.CreateTestUser(t, db, true)
	userID := user.ID

	cstore := store.New(db)
	repoStore := database.ReposWith(cstore)
	esStore := database.ExternalServicesWith(cstore)

	repo := newGitHubTestRepo("github.com/sourcegraph/create-campaign-spec-test", newGitHubExternalService(t, esStore))
	if err := repoStore.Create(ctx, repo); err != nil {
		t.Fatal(err)
	}

	// Create enough changeset specs to hit the licence check.
	changesetSpecs := make([]*campaigns.ChangesetSpec, maxUnlicensedChangesets+1)
	for i := range changesetSpecs {
		changesetSpecs[i] = &campaigns.ChangesetSpec{
			Spec: &campaigns.ChangesetSpecDescription{
				BaseRepository: graphqlbackend.MarshalRepositoryID(repo.ID),
			},
			RepoID: repo.ID,
			UserID: userID,
		}
		if err := cstore.CreateChangesetSpec(ctx, changesetSpecs[i]); err != nil {
			t.Fatal(err)
		}
	}

	r := &Resolver{store: cstore}
	s, err := graphqlbackend.NewSchema(db, r, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	userAPIID := string(graphqlbackend.MarshalUserID(userID))
	rawSpec := ct.TestRawCampaignSpec

	for name, tc := range map[string]struct {
		changesetSpecs []*campaigns.ChangesetSpec
		disableFeature bool
		wantErr        bool
	}{
		"default configuration": {
			changesetSpecs: changesetSpecs,
			disableFeature: false,
			wantErr:        true,
		},
		"no licence, but under the limit": {
			changesetSpecs: changesetSpecs[0:maxUnlicensedChangesets],
			disableFeature: true,
			wantErr:        false,
		},
		"no licence, over the limit": {
			changesetSpecs: changesetSpecs,
			disableFeature: true,
			wantErr:        true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if tc.disableFeature {
				oldMock := licensing.MockCheckFeature
				licensing.MockCheckFeature = func(feature licensing.Feature) error {
					if feature == licensing.FeatureCampaigns {
						return licensing.NewFeatureNotActivatedError("no campaigns for you!")
					}
					return nil
				}

				defer func() {
					licensing.MockCheckFeature = oldMock
				}()
			}

			changesetSpecIDs := make([]graphql.ID, len(tc.changesetSpecs))
			for i, spec := range tc.changesetSpecs {
				changesetSpecIDs[i] = marshalChangesetSpecRandID(spec.RandID)
			}

			input := map[string]interface{}{
				"namespace":      userAPIID,
				"campaignSpec":   rawSpec,
				"changesetSpecs": changesetSpecIDs,
			}

			var response struct{ CreateCampaignSpec apitest.CampaignSpec }

			actorCtx := actor.WithActor(ctx, actor.FromUser(userID))
			errs := apitest.Exec(actorCtx, t, s, input, &response, mutationCreateCampaignSpec)
			if tc.wantErr {
				if errs == nil {
					t.Error("unexpected lack of errors")
				}
			} else {
				if errs != nil {
					t.Errorf("unexpected error(s): %+v", errs)
				}

				var unmarshaled interface{}
				err = json.Unmarshal([]byte(rawSpec), &unmarshaled)
				if err != nil {
					t.Fatal(err)
				}
				have := response.CreateCampaignSpec

				wantNodes := make([]apitest.ChangesetSpec, len(changesetSpecIDs))
				for i, id := range changesetSpecIDs {
					wantNodes[i] = apitest.ChangesetSpec{
						Typename: "VisibleChangesetSpec",
						ID:       string(id),
					}
				}

				want := apitest.CampaignSpec{
					ID:            have.ID,
					CreatedAt:     have.CreatedAt,
					ExpiresAt:     have.ExpiresAt,
					OriginalInput: rawSpec,
					ParsedInput:   graphqlbackend.JSONValue{Value: unmarshaled},
					ApplyURL:      fmt.Sprintf("/users/%s/campaigns/apply/%s", user.Username, have.ID),
					Namespace:     apitest.UserOrg{ID: userAPIID, DatabaseID: userID, SiteAdmin: true},
					Creator:       &apitest.User{ID: userAPIID, DatabaseID: userID, SiteAdmin: true},
					ChangesetSpecs: apitest.ChangesetSpecConnection{
						Nodes: wantNodes,
					},
				}

				if diff := cmp.Diff(want, have); diff != "" {
					t.Fatalf("unexpected response (-want +got):\n%s", diff)
				}
			}
		})
	}
}

const mutationCreateCampaignSpec = `
fragment u on User { id, databaseID, siteAdmin }
fragment o on Org  { id, name }

mutation($namespace: ID!, $campaignSpec: String!, $changesetSpecs: [ID!]!){
  createCampaignSpec(namespace: $namespace, campaignSpec: $campaignSpec, changesetSpecs: $changesetSpecs) {
    id
    originalInput
    parsedInput

    creator  { ...u }
    namespace {
      ... on User { ...u }
      ... on Org  { ...o }
    }

    applyURL

	changesetSpecs {
	  nodes {
		  __typename
		  ... on VisibleChangesetSpec {
			  id
		  }
	  }
	}

    createdAt
    expiresAt
  }
}
`

func TestCreateChangesetSpec(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	db := dbtesting.GetDB(t)

	userID := ct.CreateTestUser(t, db, true).ID

	cstore := store.New(db)
	repoStore := database.ReposWith(cstore)
	esStore := database.ExternalServicesWith(cstore)

	repo := newGitHubTestRepo("github.com/sourcegraph/create-changeset-spec-test", newGitHubExternalService(t, esStore))
	if err := repoStore.Create(ctx, repo); err != nil {
		t.Fatal(err)
	}

	r := &Resolver{store: cstore}
	s, err := graphqlbackend.NewSchema(db, r, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	input := map[string]interface{}{
		"changesetSpec": ct.NewRawChangesetSpecGitBranch(graphqlbackend.MarshalRepositoryID(repo.ID), "d34db33f"),
	}

	var response struct{ CreateChangesetSpec apitest.ChangesetSpec }

	actorCtx := actor.WithActor(ctx, actor.FromUser(userID))
	apitest.MustExec(actorCtx, t, s, input, &response, mutationCreateChangesetSpec)

	have := response.CreateChangesetSpec

	want := apitest.ChangesetSpec{
		Typename:  "VisibleChangesetSpec",
		ID:        have.ID,
		ExpiresAt: have.ExpiresAt,
	}

	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatalf("unexpected response (-want +got):\n%s", diff)
	}

	randID, err := unmarshalChangesetSpecID(graphql.ID(want.ID))
	if err != nil {
		t.Fatal(err)
	}

	cs, err := cstore.GetChangesetSpec(ctx, store.GetChangesetSpecOpts{RandID: randID})
	if err != nil {
		t.Fatal(err)
	}

	if have, want := cs.RepoID, repo.ID; have != want {
		t.Fatalf("wrong RepoID. want=%d, have=%d", want, have)
	}
}

const mutationCreateChangesetSpec = `
mutation($changesetSpec: String!){
  createChangesetSpec(changesetSpec: $changesetSpec) {
	__typename
	... on VisibleChangesetSpec {
		id
		expiresAt
	}
  }
}
`

func TestApplyCampaign(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	db := dbtesting.GetDB(t)

	userID := ct.CreateTestUser(t, db, true).ID

	now := timeutil.Now()
	clock := func() time.Time { return now }
	cstore := store.NewWithClock(db, clock)
	repoStore := database.ReposWith(cstore)
	esStore := database.ExternalServicesWith(cstore)

	repo := newGitHubTestRepo("github.com/sourcegraph/apply-campaign-test", newGitHubExternalService(t, esStore))
	if err := repoStore.Create(ctx, repo); err != nil {
		t.Fatal(err)
	}

	repoAPIID := graphqlbackend.MarshalRepositoryID(repo.ID)

	campaignSpec := &campaigns.CampaignSpec{
		RawSpec: ct.TestRawCampaignSpec,
		Spec: campaigns.CampaignSpecFields{
			Name:        "my-campaign",
			Description: "My description",
			ChangesetTemplate: campaigns.ChangesetTemplate{
				Title:  "Hello there",
				Body:   "This is the body",
				Branch: "my-branch",
				Commit: campaigns.CommitTemplate{
					Message: "Add hello world",
				},
				Published: overridable.FromBoolOrString(false),
			},
		},
		UserID:          userID,
		NamespaceUserID: userID,
	}
	if err := cstore.CreateCampaignSpec(ctx, campaignSpec); err != nil {
		t.Fatal(err)
	}

	changesetSpec := &campaigns.ChangesetSpec{
		CampaignSpecID: campaignSpec.ID,
		Spec: &campaigns.ChangesetSpecDescription{
			BaseRepository: repoAPIID,
		},
		RepoID: repo.ID,
		UserID: userID,
	}
	if err := cstore.CreateChangesetSpec(ctx, changesetSpec); err != nil {
		t.Fatal(err)
	}

	r := &Resolver{store: cstore}
	s, err := graphqlbackend.NewSchema(db, r, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	userAPIID := string(graphqlbackend.MarshalUserID(userID))
	input := map[string]interface{}{
		"campaignSpec": string(marshalCampaignSpecRandID(campaignSpec.RandID)),
	}

	var response struct{ ApplyCampaign apitest.Campaign }
	actorCtx := actor.WithActor(ctx, actor.FromUser(userID))
	apitest.MustExec(actorCtx, t, s, input, &response, mutationApplyCampaign)

	apiUser := &apitest.User{
		ID:         userAPIID,
		DatabaseID: userID,
		SiteAdmin:  true,
	}

	have := response.ApplyCampaign
	want := apitest.Campaign{
		ID:          have.ID,
		Name:        campaignSpec.Spec.Name,
		Description: campaignSpec.Spec.Description,
		Namespace: apitest.UserOrg{
			ID:         userAPIID,
			DatabaseID: userID,
			SiteAdmin:  true,
		},
		InitialApplier: apiUser,
		LastApplier:    apiUser,
		LastAppliedAt:  marshalDateTime(t, now),
		Changesets: apitest.ChangesetConnection{
			Nodes: []apitest.Changeset{
				{Typename: "ExternalChangeset", State: string(campaigns.ChangesetStateProcessing)},
			},
			TotalCount: 1,
		},
	}

	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatalf("unexpected response (-want +got):\n%s", diff)
	}

	// Now we execute it again and make sure we get the same campaign back
	apitest.MustExec(actorCtx, t, s, input, &response, mutationApplyCampaign)
	have2 := response.ApplyCampaign
	if diff := cmp.Diff(want, have2); diff != "" {
		t.Fatalf("unexpected response (-want +got):\n%s", diff)
	}

	// Execute it again with ensureCampaign set to correct campaign's ID
	input["ensureCampaign"] = have2.ID
	apitest.MustExec(actorCtx, t, s, input, &response, mutationApplyCampaign)
	have3 := response.ApplyCampaign
	if diff := cmp.Diff(want, have3); diff != "" {
		t.Fatalf("unexpected response (-want +got):\n%s", diff)
	}

	// Execute it again but ensureCampaign set to wrong campaign ID
	campaignID, err := unmarshalCampaignID(graphql.ID(have3.ID))
	if err != nil {
		t.Fatal(err)
	}
	input["ensureCampaign"] = marshalCampaignID(campaignID + 999)
	errs := apitest.Exec(actorCtx, t, s, input, &response, mutationApplyCampaign)
	if len(errs) == 0 {
		t.Fatalf("expected errors, got none")
	}
}

const mutationApplyCampaign = `
fragment u on User { id, databaseID, siteAdmin }
fragment o on Org  { id, name }

mutation($campaignSpec: ID!, $ensureCampaign: ID){
  applyCampaign(campaignSpec: $campaignSpec, ensureCampaign: $ensureCampaign) {
    id, name, description
    initialApplier    { ...u }
    lastApplier       { ...u }
    lastAppliedAt
    namespace {
        ... on User { ...u }
        ... on Org  { ...o }
    }

    changesets {
      nodes {
        __typename
        state
      }

      totalCount
    }
  }
}
`

func TestCreateCampaign(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	db := dbtesting.GetDB(t)

	userID := ct.CreateTestUser(t, db, true).ID

	cstore := store.New(db)

	campaignSpec := &campaigns.CampaignSpec{
		RawSpec: ct.TestRawCampaignSpec,
		Spec: campaigns.CampaignSpecFields{
			Name:        "my-campaign",
			Description: "My description",
		},
		UserID:          userID,
		NamespaceUserID: userID,
	}
	if err := cstore.CreateCampaignSpec(ctx, campaignSpec); err != nil {
		t.Fatal(err)
	}

	r := &Resolver{store: cstore}
	s, err := graphqlbackend.NewSchema(db, r, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	input := map[string]interface{}{
		"campaignSpec": string(marshalCampaignSpecRandID(campaignSpec.RandID)),
	}

	var response struct{ CreateCampaign apitest.Campaign }
	actorCtx := actor.WithActor(ctx, actor.FromUser(userID))

	// First time it should work, because no campaign exists
	apitest.MustExec(actorCtx, t, s, input, &response, mutationCreateCampaign)

	if response.CreateCampaign.ID == "" {
		t.Fatalf("expected campaign to be created, but was not")
	}

	// Second time it should fail
	errors := apitest.Exec(actorCtx, t, s, input, &response, mutationCreateCampaign)

	if len(errors) != 1 {
		t.Fatalf("expected single errors, but got none")
	}
	if have, want := errors[0].Message, service.ErrMatchingCampaignExists.Error(); have != want {
		t.Fatalf("wrong error. want=%q, have=%q", want, have)
	}
}

const mutationCreateCampaign = `
mutation($campaignSpec: ID!){
  createCampaign(campaignSpec: $campaignSpec) { id }
}
`

func TestMoveCampaign(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	db := dbtesting.GetDB(t)

	user := ct.CreateTestUser(t, db, true)
	userID := user.ID

	orgName := "move-campaign-test"
	orgID := ct.InsertTestOrg(t, db, orgName)

	cstore := store.New(db)

	campaignSpec := &campaigns.CampaignSpec{
		RawSpec:         ct.TestRawCampaignSpec,
		UserID:          userID,
		NamespaceUserID: userID,
	}
	if err := cstore.CreateCampaignSpec(ctx, campaignSpec); err != nil {
		t.Fatal(err)
	}

	campaign := &campaigns.Campaign{
		CampaignSpecID:   campaignSpec.ID,
		Name:             "old-name",
		InitialApplierID: userID,
		LastApplierID:    userID,
		LastAppliedAt:    time.Now(),
		NamespaceUserID:  campaignSpec.UserID,
	}
	if err := cstore.CreateCampaign(ctx, campaign); err != nil {
		t.Fatal(err)
	}

	r := &Resolver{store: cstore}
	s, err := graphqlbackend.NewSchema(db, r, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Move to a new name
	campaignAPIID := string(marshalCampaignID(campaign.ID))
	newCampaignName := "new-name"
	input := map[string]interface{}{
		"campaign": campaignAPIID,
		"newName":  newCampaignName,
	}

	var response struct{ MoveCampaign apitest.Campaign }
	actorCtx := actor.WithActor(ctx, actor.FromUser(userID))
	apitest.MustExec(actorCtx, t, s, input, &response, mutationMoveCampaign)

	haveCampaign := response.MoveCampaign
	if diff := cmp.Diff(input["newName"], haveCampaign.Name); diff != "" {
		t.Fatalf("unexpected name (-want +got):\n%s", diff)
	}

	wantURL := fmt.Sprintf("/users/%s/campaigns/%s", user.Username, newCampaignName)
	if diff := cmp.Diff(wantURL, haveCampaign.URL); diff != "" {
		t.Fatalf("unexpected URL (-want +got):\n%s", diff)
	}

	// Move to a new namespace
	orgAPIID := graphqlbackend.MarshalOrgID(orgID)
	input = map[string]interface{}{
		"campaign":     string(marshalCampaignID(campaign.ID)),
		"newNamespace": orgAPIID,
	}

	apitest.MustExec(actorCtx, t, s, input, &response, mutationMoveCampaign)

	haveCampaign = response.MoveCampaign
	if diff := cmp.Diff(string(orgAPIID), haveCampaign.Namespace.ID); diff != "" {
		t.Fatalf("unexpected namespace (-want +got):\n%s", diff)
	}
	wantURL = fmt.Sprintf("/organizations/%s/campaigns/%s", orgName, newCampaignName)
	if diff := cmp.Diff(wantURL, haveCampaign.URL); diff != "" {
		t.Fatalf("unexpected URL (-want +got):\n%s", diff)
	}
}

const mutationMoveCampaign = `
fragment u on User { id, databaseID, siteAdmin }
fragment o on Org  { id, name }

mutation($campaign: ID!, $newName: String, $newNamespace: ID){
  moveCampaign(campaign: $campaign, newName: $newName, newNamespace: $newNamespace) {
	id, name, description
	initialApplier  { ...u }
	namespace {
		... on User { ...u }
		... on Org  { ...o }
	}
	url
  }
}
`

func TestListChangesetOptsFromArgs(t *testing.T) {
	var wantFirst int32 = 10
	wantPublicationStates := []campaigns.ChangesetPublicationState{
		"PUBLISHED",
		"INVALID",
	}
	wantStates := []campaigns.ChangesetState{"OPEN", "INVALID"}
	wantExternalStates := []campaigns.ChangesetExternalState{"OPEN"}
	wantReviewStates := []campaigns.ChangesetReviewState{"APPROVED", "INVALID"}
	wantCheckStates := []campaigns.ChangesetCheckState{"PENDING", "INVALID"}
	wantOnlyPublishedByThisCampaign := []bool{true}
	wantSearches := []search.TextSearchTerm{{Term: "foo"}, {Term: "bar", Not: true}}
	var campaignID int64 = 1

	tcs := []struct {
		args       *graphqlbackend.ListChangesetsArgs
		wantSafe   bool
		wantErr    string
		wantParsed store.ListChangesetsOpts
	}{
		// No args given.
		{
			args:       nil,
			wantSafe:   true,
			wantParsed: store.ListChangesetsOpts{},
		},
		// First argument is set in opts, and considered safe.
		{
			args: &graphqlbackend.ListChangesetsArgs{
				First: wantFirst,
			},
			wantSafe:   true,
			wantParsed: store.ListChangesetsOpts{LimitOpts: store.LimitOpts{Limit: 10}},
		},
		// Setting state is safe and transferred to opts.
		{
			args: &graphqlbackend.ListChangesetsArgs{
				State: &wantStates[0],
			},
			wantSafe: true,
			wantParsed: store.ListChangesetsOpts{
				ExternalState:    &wantExternalStates[0],
				PublicationState: &wantPublicationStates[0],
				ReconcilerStates: []campaigns.ReconcilerState{campaigns.ReconcilerStateCompleted},
			},
		},
		// Setting invalid state fails.
		{
			args: &graphqlbackend.ListChangesetsArgs{
				State: &wantStates[1],
			},
			wantErr: "changeset state not valid",
		},
		// Setting review state is not safe and transferred to opts.
		{
			args: &graphqlbackend.ListChangesetsArgs{
				ReviewState: &wantReviewStates[0],
			},
			wantSafe:   false,
			wantParsed: store.ListChangesetsOpts{ExternalReviewState: &wantReviewStates[0]},
		},
		// Setting invalid review state fails.
		{
			args: &graphqlbackend.ListChangesetsArgs{
				ReviewState: &wantReviewStates[1],
			},
			wantErr: "changeset review state not valid",
		},
		// Setting check state is not safe and transferred to opts.
		{
			args: &graphqlbackend.ListChangesetsArgs{
				CheckState: &wantCheckStates[0],
			},
			wantSafe:   false,
			wantParsed: store.ListChangesetsOpts{ExternalCheckState: &wantCheckStates[0]},
		},
		// Setting invalid check state fails.
		{
			args: &graphqlbackend.ListChangesetsArgs{
				CheckState: &wantCheckStates[1],
			},
			wantErr: "changeset check state not valid",
		},
		// Setting OnlyPublishedByThisCampaign true.
		{
			args: &graphqlbackend.ListChangesetsArgs{
				OnlyPublishedByThisCampaign: &wantOnlyPublishedByThisCampaign[0],
			},
			wantSafe: true,
			wantParsed: store.ListChangesetsOpts{
				PublicationState:  &wantPublicationStates[0],
				OwnedByCampaignID: campaignID,
			},
		},
		// Setting a positive search.
		{
			args: &graphqlbackend.ListChangesetsArgs{
				Search: stringPtr("foo"),
			},
			wantSafe: false,
			wantParsed: store.ListChangesetsOpts{
				TextSearch: wantSearches[0:1],
			},
		},
		// Setting a negative search.
		{
			args: &graphqlbackend.ListChangesetsArgs{
				Search: stringPtr("-bar"),
			},
			wantSafe: false,
			wantParsed: store.ListChangesetsOpts{
				TextSearch: wantSearches[1:],
			},
		},
	}
	for i, tc := range tcs {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			haveParsed, haveSafe, err := listChangesetOptsFromArgs(tc.args, campaignID)
			if tc.wantErr == "" && err != nil {
				t.Fatal(err)
			}
			haveErr := fmt.Sprintf("%v", err)
			wantErr := tc.wantErr
			if wantErr == "" {
				wantErr = "<nil>"
			}
			if have, want := haveErr, wantErr; have != want {
				t.Errorf("wrong error returned. have=%q want=%q", have, want)
			}
			if diff := cmp.Diff(haveParsed, tc.wantParsed); diff != "" {
				t.Errorf("wrong args returned. diff=%s", diff)
			}
			if have, want := haveSafe, tc.wantSafe; have != want {
				t.Errorf("wrong safe value returned. have=%t want=%t", have, want)
			}
		})
	}
}

func TestCreateCampaignsCredential(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	db := dbtesting.GetDB(t)

	pruneUserCredentials(t, db)

	userID := ct.CreateTestUser(t, db, false).ID

	cstore := store.New(db)

	r := &Resolver{store: cstore}
	s, err := graphqlbackend.NewSchema(db, r, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	input := map[string]interface{}{
		"user":                graphqlbackend.MarshalUserID(userID),
		"externalServiceKind": string(extsvc.KindGitHub),
		"externalServiceURL":  "https://github.com/",
		"credential":          "SOSECRET",
	}

	var response struct{ CreateCampaignsCredential apitest.CampaignsCredential }
	actorCtx := actor.WithActor(ctx, actor.FromUser(userID))

	// First time it should work, because no credential exists
	apitest.MustExec(actorCtx, t, s, input, &response, mutationCreateCredential)

	if response.CreateCampaignsCredential.ID == "" {
		t.Fatalf("expected credential to be created, but was not")
	}

	// Second time it should fail
	errors := apitest.Exec(actorCtx, t, s, input, &response, mutationCreateCredential)

	if len(errors) != 1 {
		t.Fatalf("expected single errors, but got none")
	}
	if have, want := errors[0].Extensions["code"], "ErrDuplicateCredential"; have != want {
		t.Fatalf("wrong error code. want=%q, have=%q", want, have)
	}
}

const mutationCreateCredential = `
mutation($user: ID!, $externalServiceKind: ExternalServiceKind!, $externalServiceURL: String!, $credential: String!) {
  createCampaignsCredential(user: $user, externalServiceKind: $externalServiceKind, externalServiceURL: $externalServiceURL, credential: $credential) { id }
}
`

func TestDeleteCampaignsCredential(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	db := dbtesting.GetDB(t)

	pruneUserCredentials(t, db)

	userID := ct.CreateTestUser(t, db, true).ID

	cstore := store.New(db)

	cred, err := cstore.UserCredentials().Create(ctx, database.UserCredentialScope{
		Domain:              database.UserCredentialDomainCampaigns,
		ExternalServiceType: extsvc.TypeGitHub,
		ExternalServiceID:   "https://github.com/",
		UserID:              userID,
	}, &auth.OAuthBearerToken{Token: "SOSECRET"})
	if err != nil {
		t.Fatal(err)
	}

	r := &Resolver{store: cstore}
	s, err := graphqlbackend.NewSchema(db, r, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	input := map[string]interface{}{
		"campaignsCredential": marshalCampaignsCredentialID(cred.ID),
	}

	var response struct{ DeleteCampaignsCredential apitest.EmptyResponse }
	actorCtx := actor.WithActor(ctx, actor.FromUser(userID))

	// First time it should work, because a credential exists
	apitest.MustExec(actorCtx, t, s, input, &response, mutationDeleteCredential)

	// Second time it should fail
	errors := apitest.Exec(actorCtx, t, s, input, &response, mutationDeleteCredential)

	if len(errors) != 1 {
		t.Fatalf("expected single errors, but got none")
	}
	if have, want := errors[0].Message, fmt.Sprintf("user credential not found: [%d]", cred.ID); have != want {
		t.Fatalf("wrong error code. want=%q, have=%q", want, have)
	}
}

const mutationDeleteCredential = `
mutation($campaignsCredential: ID!) {
  deleteCampaignsCredential(campaignsCredential: $campaignsCredential) { alwaysNil }
}
`

func stringPtr(s string) *string { return &s }
