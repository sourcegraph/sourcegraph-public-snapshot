package resolvers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/resolvers/apitest"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/testing"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestChangesetCountsOverTime(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)
	rcache.SetupForTest(t)

	cf, save := httptestutil.NewGitHubRecorderFactory(t, *update, "test-changeset-counts-over-time")
	defer save()

	userID := insertTestUser(t, dbconn.Global, "changeset-counts-over-time", false)

	repoStore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})
	githubExtSvc := &repos.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GitHub",
		Config: marshalJSON(t, &schema.GitHubConnection{
			Url:   "https://github.com",
			Token: os.Getenv("GITHUB_TOKEN"),
			Repos: []string{"sourcegraph/sourcegraph"},
		}),
	}

	err := repoStore.UpsertExternalServices(ctx, githubExtSvc)
	if err != nil {
		t.Fatal(t)
	}

	githubSrc, err := repos.NewGithubSource(githubExtSvc, cf)
	if err != nil {
		t.Fatal(t)
	}

	githubRepo, err := githubSrc.GetRepo(ctx, "sourcegraph/sourcegraph")
	if err != nil {
		t.Fatal(err)
	}

	err = repoStore.UpsertRepos(ctx, githubRepo)
	if err != nil {
		t.Fatal(err)
	}

	store := ee.NewStore(dbconn.Global)

	campaign := &campaigns.Campaign{
		Name:            "Test campaign",
		Description:     "Testing changeset counts",
		AuthorID:        userID,
		NamespaceUserID: userID,
	}

	err = store.CreateCampaign(ctx, campaign)
	if err != nil {
		t.Fatal(err)
	}

	changesets := []*campaigns.Changeset{
		{
			RepoID:              githubRepo.ID,
			ExternalID:          "5834",
			ExternalServiceType: githubRepo.ExternalRepo.ServiceType,
			CampaignIDs:         []int64{campaign.ID},
		},
		{
			RepoID:              githubRepo.ID,
			ExternalID:          "5849",
			ExternalServiceType: githubRepo.ExternalRepo.ServiceType,
			CampaignIDs:         []int64{campaign.ID},
		},
	}

	for _, c := range changesets {
		if err = store.CreateChangeset(ctx, c); err != nil {
			t.Fatal(err)
		}

		campaign.ChangesetIDs = append(campaign.ChangesetIDs, c.ID)
	}

	mockState := ct.MockChangesetSyncState(&protocol.RepoInfo{
		Name: api.RepoName(githubRepo.Name),
		VCS:  protocol.VCSInfo{URL: githubRepo.URI},
	})
	defer mockState.Unmock()

	err = ee.SyncChangesets(ctx, repoStore, store, cf, changesets...)
	if err != nil {
		t.Fatal(err)
	}

	err = store.UpdateCampaign(ctx, campaign)
	if err != nil {
		t.Fatal(err)
	}

	// Date when PR #5834 was created: "2019-10-02T14:49:31Z"
	// We start exactly one day earlier
	// Date when PR #5849 was created: "2019-10-03T15:03:21Z"
	start := parseJSONTime(t, "2019-10-01T14:49:31Z")
	// Date when PR #5834 was merged:  "2019-10-07T13:13:45Z"
	// Date when PR #5849 was merged:  "2019-10-04T08:55:21Z"
	end := parseJSONTime(t, "2019-10-07T13:13:45Z")
	daysBeforeEnd := func(days int) time.Time {
		return end.AddDate(0, 0, -days)
	}

	r := &campaignResolver{store: store, Campaign: campaign}
	rs, err := r.ChangesetCountsOverTime(ctx, &graphqlbackend.ChangesetCountsArgs{
		From: &graphqlbackend.DateTime{Time: start},
		To:   &graphqlbackend.DateTime{Time: end},
	})
	if err != nil {
		t.Fatalf("ChangsetCountsOverTime failed with error: %s", err)
	}

	have := make([]*ee.ChangesetCounts, 0, len(rs))
	for _, cr := range rs {
		r := cr.(*changesetCountsResolver)
		have = append(have, r.counts)
	}

	want := []*ee.ChangesetCounts{
		{Time: daysBeforeEnd(5), Total: 0, Open: 0},
		{Time: daysBeforeEnd(4), Total: 1, Open: 1, OpenPending: 1},
		{Time: daysBeforeEnd(3), Total: 2, Open: 1, OpenPending: 1, Merged: 1},
		{Time: daysBeforeEnd(2), Total: 2, Open: 1, OpenPending: 1, Merged: 1},
		{Time: daysBeforeEnd(1), Total: 2, Open: 1, OpenPending: 1, Merged: 1},
		{Time: end, Total: 2, Merged: 2},
	}

	if !reflect.DeepEqual(have, want) {
		t.Errorf("wrong counts listed. diff=%s", cmp.Diff(have, want))
	}
}

func TestNullIDResilience(t *testing.T) {
	sr := &Resolver{store: ee.NewStore(dbconn.Global)}

	s, err := graphqlbackend.NewSchema(sr, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := backend.WithAuthzBypass(context.Background())

	ids := []graphql.ID{
		campaigns.MarshalCampaignID(0),
		marshalChangesetID(0),
		marshalCampaignSpecRandID(""),
		marshalChangesetSpecRandID(""),
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
		fmt.Sprintf(`mutation { closeCampaign(campaign: %q) { id } }`, campaigns.MarshalCampaignID(0)),
		fmt.Sprintf(`mutation { deleteCampaign(campaign: %q) { alwaysNil } }`, campaigns.MarshalCampaignID(0)),
		fmt.Sprintf(`mutation { syncChangeset(changeset: %q) { alwaysNil } }`, marshalChangesetID(0)),
		fmt.Sprintf(`mutation { applyCampaign(campaignSpec: %q) { id } }`, marshalCampaignSpecRandID("")),
		fmt.Sprintf(`mutation { moveCampaign(campaign: %q, newName: "foobar") { id } }`, campaigns.MarshalCampaignID(0)),
	}

	for _, m := range mutations {
		var response struct{}
		errs := apitest.Exec(ctx, t, s, nil, &response, m)
		if len(errs) == 0 {
			t.Fatalf("expected errors but none returned (mutation: %q)", m)
		}
		if have, want := errs[0].Error(), fmt.Sprintf("graphql: %s", ErrIDIsZero.Error()); have != want {
			t.Fatalf("wrong errors. have=%s, want=%s (mutation: %q)", have, want, m)
		}
	}
}

func getBitbucketServerRepos(t testing.TB, ctx context.Context, src *repos.BitbucketServerSource) []*repos.Repo {
	results := make(chan repos.SourceResult)

	go func() {
		src.ListRepos(ctx, results)
		close(results)
	}()

	var repos []*repos.Repo

	for res := range results {
		if res.Err != nil {
			t.Fatal(res.Err)
		}
		repos = append(repos, res.Repo)
	}

	return repos
}

func TestCreateCampaignSpec(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	dbtesting.SetupGlobalTestDB(t)

	userID := insertTestUser(t, dbconn.Global, "create-campaign-spec", true)

	store := ee.NewStore(dbconn.Global)
	reposStore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})

	repo := newGitHubTestRepo("github.com/sourcegraph/sourcegraph", 1)
	if err := reposStore.UpsertRepos(ctx, repo); err != nil {
		t.Fatal(err)
	}

	changesetSpec := &campaigns.ChangesetSpec{
		Spec: &campaigns.ChangesetSpecDescription{
			BaseRepository: graphqlbackend.MarshalRepositoryID(repo.ID),
		},
		RepoID: repo.ID,
		UserID: userID,
	}
	if err := store.CreateChangesetSpec(ctx, changesetSpec); err != nil {
		t.Fatal(err)
	}

	r := &Resolver{store: store}
	s, err := graphqlbackend.NewSchema(r, nil, nil)
	if err != nil {
		t.Fatal(err)

	}

	userApiID := string(graphqlbackend.MarshalUserID(userID))
	changesetSpecID := marshalChangesetSpecRandID(changesetSpec.RandID)
	rawSpec := ct.TestRawCampaignSpec

	input := map[string]interface{}{
		"namespace":      userApiID,
		"campaignSpec":   rawSpec,
		"changesetSpecs": []graphql.ID{changesetSpecID},
	}

	var response struct{ CreateCampaignSpec apitest.CampaignSpec }

	actorCtx := actor.WithActor(ctx, actor.FromUser(userID))
	apitest.MustExec(actorCtx, t, s, input, &response, mutationCreateCampaignSpec)

	var unmarshaled interface{}
	err = json.Unmarshal([]byte(rawSpec), &unmarshaled)
	if err != nil {
		t.Fatal(err)
	}

	want := apitest.CampaignSpec{
		OriginalInput: rawSpec,
		ParsedInput:   graphqlbackend.JSONValue{Value: unmarshaled},
		PreviewURL:    "/campaigns/new?spec=",
		Namespace:     apitest.UserOrg{ID: userApiID, DatabaseID: userID, SiteAdmin: true},
		Creator:       apitest.User{ID: userApiID, DatabaseID: userID, SiteAdmin: true},
		ChangesetSpecs: apitest.ChangesetSpecConnection{
			Nodes: []apitest.ChangesetSpec{
				{
					Typename: "VisibleChangesetSpec",
					ID:       string(changesetSpecID),
				},
			},
		},
	}
	have := response.CreateCampaignSpec

	want.ID = have.ID
	want.PreviewURL = want.PreviewURL + want.ID
	want.CreatedAt = have.CreatedAt
	want.ExpiresAt = have.ExpiresAt

	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatalf("unexpected response (-want +got):\n%s", diff)
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

    previewURL

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
	dbtesting.SetupGlobalTestDB(t)

	userID := insertTestUser(t, dbconn.Global, "create-changeset-spec", true)

	store := ee.NewStore(dbconn.Global)
	reposStore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})

	repo := newGitHubTestRepo("github.com/sourcegraph/sourcegraph", 1)
	if err := reposStore.UpsertRepos(ctx, repo); err != nil {
		t.Fatal(err)
	}

	r := &Resolver{store: store}
	s, err := graphqlbackend.NewSchema(r, nil, nil)
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

	cs, err := store.GetChangesetSpec(ctx, ee.GetChangesetSpecOpts{RandID: randID})
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
	dbtesting.SetupGlobalTestDB(t)

	userID := insertTestUser(t, dbconn.Global, "apply-campaign", true)

	store := ee.NewStore(dbconn.Global)
	reposStore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})

	repo := newGitHubTestRepo("github.com/sourcegraph/sourcegraph", 1)
	if err := reposStore.UpsertRepos(ctx, repo); err != nil {
		t.Fatal(err)
	}

	repoApiID := graphqlbackend.MarshalRepositoryID(repo.ID)

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
				Published: false,
			},
		},
		UserID:          userID,
		NamespaceUserID: userID,
	}
	if err := store.CreateCampaignSpec(ctx, campaignSpec); err != nil {
		t.Fatal(err)
	}

	changesetSpec := &campaigns.ChangesetSpec{
		CampaignSpecID: campaignSpec.ID,
		Spec: &campaigns.ChangesetSpecDescription{
			BaseRepository: repoApiID,
		},
		RepoID: repo.ID,
		UserID: userID,
	}
	if err := store.CreateChangesetSpec(ctx, changesetSpec); err != nil {
		t.Fatal(err)
	}

	r := &Resolver{store: store}
	s, err := graphqlbackend.NewSchema(r, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	userApiID := string(graphqlbackend.MarshalUserID(userID))
	input := map[string]interface{}{
		"campaignSpec": string(marshalCampaignSpecRandID(campaignSpec.RandID)),
	}

	var response struct{ ApplyCampaign apitest.Campaign }
	actorCtx := actor.WithActor(ctx, actor.FromUser(userID))
	apitest.MustExec(actorCtx, t, s, input, &response, mutationApplyCampaign)

	have := response.ApplyCampaign
	want := apitest.Campaign{
		ID:          have.ID,
		Name:        campaignSpec.Spec.Name,
		Description: campaignSpec.Spec.Description,
		Branch:      campaignSpec.Spec.ChangesetTemplate.Branch,
		Namespace: apitest.UserOrg{
			ID:         userApiID,
			DatabaseID: userID,
			SiteAdmin:  true,
		},
		Author: apitest.User{
			ID:         userApiID,
			DatabaseID: userID,
			SiteAdmin:  true,
		},
		Changesets: apitest.ChangesetConnection{
			Nodes: []apitest.Changeset{
				{Typename: "ExternalChangeset", ReconcilerState: "QUEUED"},
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
	campaignID, err := campaigns.UnmarshalCampaignID(graphql.ID(have3.ID))
	if err != nil {
		t.Fatal(err)
	}
	input["ensureCampaign"] = campaigns.MarshalCampaignID(campaignID + 999)
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
    id, name, description, branch
    author    { ...u }
    namespace {
        ... on User { ...u }
        ... on Org  { ...o }
    }

    changesets {
      nodes {
        __typename

        ... on ExternalChangeset {
          reconcilerState
        }
        ... on HiddenExternalChangeset {
          reconcilerState
        }
      }

      totalCount
    }
  }
}
`

func TestMoveCampaign(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	dbtesting.SetupGlobalTestDB(t)

	userID := insertTestUser(t, dbconn.Global, "move-campaign1", true)

	org, err := db.Orgs.Create(ctx, "org", nil)
	if err != nil {
		t.Fatal(err)
	}

	store := ee.NewStore(dbconn.Global)

	campaignSpec := &campaigns.CampaignSpec{
		RawSpec:         ct.TestRawCampaignSpec,
		UserID:          userID,
		NamespaceUserID: userID,
	}
	if err := store.CreateCampaignSpec(ctx, campaignSpec); err != nil {
		t.Fatal(err)
	}

	campaign := &campaigns.Campaign{
		CampaignSpecID:  campaignSpec.ID,
		Name:            "old-name",
		AuthorID:        userID,
		NamespaceUserID: campaignSpec.UserID,
	}
	if err := store.CreateCampaign(ctx, campaign); err != nil {
		t.Fatal(err)
	}

	r := &Resolver{store: store}
	s, err := graphqlbackend.NewSchema(r, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Move to a new name
	input := map[string]interface{}{
		"campaign": string(campaigns.MarshalCampaignID(campaign.ID)),
		"newName":  "new-name",
	}

	var response struct{ MoveCampaign apitest.Campaign }
	actorCtx := actor.WithActor(ctx, actor.FromUser(userID))
	apitest.MustExec(actorCtx, t, s, input, &response, mutationMoveCampaign)

	haveCampaign := response.MoveCampaign
	if diff := cmp.Diff(input["newName"], haveCampaign.Name); diff != "" {
		t.Fatalf("unexpected name (-want +got):\n%s", diff)
	}

	// Move to a new namespace
	orgApiID := graphqlbackend.MarshalOrgID(org.ID)
	input = map[string]interface{}{
		"campaign":     string(campaigns.MarshalCampaignID(campaign.ID)),
		"newNamespace": orgApiID,
	}

	apitest.MustExec(actorCtx, t, s, input, &response, mutationMoveCampaign)

	haveCampaign = response.MoveCampaign
	if diff := cmp.Diff(string(orgApiID), haveCampaign.Namespace.ID); diff != "" {
		t.Fatalf("unexpected namespace (-want +got):\n%s", diff)
	}
}

const mutationMoveCampaign = `
fragment u on User { id, databaseID, siteAdmin }
fragment o on Org  { id, name }

mutation($campaign: ID!, $newName: String, $newNamespace: ID){
  moveCampaign(campaign: $campaign, newName: $newName, newNamespace: $newNamespace) {
	id, name, description, branch
	author    { ...u }
	namespace {
		... on User { ...u }
		... on Org  { ...o }
	}
  }
}
`

func TestListChangesetOptsFromArgs(t *testing.T) {
	var wantFirst int32 = 10
	wantPublicationStates := []campaigns.ChangesetPublicationState{
		"PUBLISHED",
		"INVALID",
	}
	reconcilerStates := []campaigns.ReconcilerState{
		"PROCESSING",
		campaigns.ReconcilerStateProcessing,
		"INVALID",
	}
	wantExternalStates := []campaigns.ChangesetExternalState{"OPEN", "INVALID"}
	wantReviewStates := []campaigns.ChangesetReviewState{"APPROVED", "INVALID"}
	wantCheckStates := []campaigns.ChangesetCheckState{"PENDING", "INVALID"}

	tcs := []struct {
		args       *graphqlbackend.ListChangesetsArgs
		wantSafe   bool
		wantErr    string
		wantParsed ee.ListChangesetsOpts
	}{
		// No args given.
		{
			args:       nil,
			wantSafe:   true,
			wantParsed: ee.ListChangesetsOpts{},
		},
		// First argument is set in opts, and considered safe.
		{
			args: &graphqlbackend.ListChangesetsArgs{
				First: &wantFirst,
			},
			wantSafe:   true,
			wantParsed: ee.ListChangesetsOpts{Limit: 10},
		},
		// Setting publication state is safe and transferred to opts.
		{
			args: &graphqlbackend.ListChangesetsArgs{
				PublicationState: &wantPublicationStates[0],
			},
			wantSafe: true,
			wantParsed: ee.ListChangesetsOpts{
				PublicationState: &wantPublicationStates[0],
			},
		},
		// Setting invalid publication state fails.
		{
			args: &graphqlbackend.ListChangesetsArgs{
				PublicationState: &wantPublicationStates[1],
			},
			wantErr: "changeset publication state not valid",
		},
		// Setting reconciler state is safe and transferred to opts as lowercase version.
		{
			args: &graphqlbackend.ListChangesetsArgs{
				ReconcilerState: &reconcilerStates[0],
			},
			wantSafe: true,
			wantParsed: ee.ListChangesetsOpts{
				ReconcilerState: &reconcilerStates[1],
			},
		},
		// Setting invalid reconciler state fails.
		{
			args: &graphqlbackend.ListChangesetsArgs{
				ReconcilerState: &reconcilerStates[2],
			},
			wantErr: "changeset reconciler state not valid",
		},
		// Setting external state is safe and transferred to opts.
		{
			args: &graphqlbackend.ListChangesetsArgs{
				ExternalState: &wantExternalStates[0],
			},
			wantSafe:   true,
			wantParsed: ee.ListChangesetsOpts{ExternalState: &wantExternalStates[0]},
		},
		// Setting invalid external state fails.
		{
			args: &graphqlbackend.ListChangesetsArgs{
				ExternalState: &wantExternalStates[1],
			},
			wantErr: "changeset external state not valid",
		},
		// Setting review state is not safe and transferred to opts.
		{
			args: &graphqlbackend.ListChangesetsArgs{
				ReviewState: &wantReviewStates[0],
			},
			wantSafe:   false,
			wantParsed: ee.ListChangesetsOpts{ExternalReviewState: &wantReviewStates[0]},
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
			wantParsed: ee.ListChangesetsOpts{ExternalCheckState: &wantCheckStates[0]},
		},
		// Setting invalid check state fails.
		{
			args: &graphqlbackend.ListChangesetsArgs{
				CheckState: &wantCheckStates[1],
			},
			wantErr: "changeset check state not valid",
		},
	}
	for _, tc := range tcs {
		haveParsed, haveSafe, err := listChangesetOptsFromArgs(tc.args)
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
	}
}

func TestCampaignsListing(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	dbtesting.SetupGlobalTestDB(t)

	userID := insertTestUser(t, dbconn.Global, "campaigns-lsiting", true)
	actorCtx := actor.WithActor(ctx, actor.FromUser(userID))

	org, err := db.Orgs.Create(ctx, "org", nil)
	if err != nil {
		t.Fatal(err)
	}

	store := ee.NewStore(dbconn.Global)

	r := &Resolver{store: store}
	s, err := graphqlbackend.NewSchema(r, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	createCampaign := func(t *testing.T, c *campaigns.Campaign) {
		t.Helper()

		c.Name = "n"
		c.AuthorID = userID
		if err := store.CreateCampaign(ctx, c); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("listing a users campaigns", func(t *testing.T) {
		campaign := &campaigns.Campaign{NamespaceUserID: userID}
		createCampaign(t, campaign)

		userApiID := string(graphqlbackend.MarshalUserID(userID))
		input := map[string]interface{}{"node": userApiID}

		var response struct{ Node apitest.User }
		apitest.MustExec(actorCtx, t, s, input, &response, listNamespacesCampaigns)

		want := apitest.User{
			ID: userApiID,
			Campaigns: apitest.CampaignConnection{
				TotalCount: 1,
				Nodes: []apitest.Campaign{
					{ID: string(campaigns.MarshalCampaignID(campaign.ID))},
				},
			},
		}

		if diff := cmp.Diff(want, response.Node); diff != "" {
			t.Fatalf("wrong campaign response (-want +got):\n%s", diff)
		}
	})

	t.Run("listing an orgs campaigns", func(t *testing.T) {
		campaign := &campaigns.Campaign{NamespaceOrgID: org.ID}
		createCampaign(t, campaign)

		orgApiID := string(graphqlbackend.MarshalOrgID(org.ID))
		input := map[string]interface{}{"node": orgApiID}

		var response struct{ Node apitest.Org }
		apitest.MustExec(actorCtx, t, s, input, &response, listNamespacesCampaigns)

		want := apitest.Org{
			ID: orgApiID,
			Campaigns: apitest.CampaignConnection{
				TotalCount: 1,
				Nodes: []apitest.Campaign{
					{ID: string(campaigns.MarshalCampaignID(campaign.ID))},
				},
			},
		}

		if diff := cmp.Diff(want, response.Node); diff != "" {
			t.Fatalf("wrong campaign response (-want +got):\n%s", diff)
		}
	})
}

const listNamespacesCampaigns = `
query($node: ID!) {
  node(id: $node) {
    ... on User {
      id
      campaigns {
        totalCount
        nodes {
          id
        }
      }
    }

    ... on Org {
      id
      campaigns {
        totalCount
        nodes {
          id
        }
      }
    }
  }
}
`
