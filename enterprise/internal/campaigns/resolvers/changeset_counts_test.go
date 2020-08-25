package resolvers

import (
	"context"
	"database/sql"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/resolvers/apitest"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/testing"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestChangesetCountsOverTimeResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)

	userID := insertTestUser(t, dbconn.Global, "changeset-count-resolver", true)

	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := func() time.Time {
		return now.UTC().Truncate(time.Microsecond)
	}
	store := ee.NewStoreWithClock(dbconn.Global, clock)
	rstore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})

	repo := newGitHubTestRepo("github.com/sourcegraph/sourcegraph", newGitHubExternalService(t, rstore))
	if err := rstore.InsertRepos(ctx, repo); err != nil {
		t.Fatal(err)
	}

	spec := &campaigns.CampaignSpec{
		NamespaceUserID: userID,
		UserID:          userID,
	}
	if err := store.CreateCampaignSpec(ctx, spec); err != nil {
		t.Fatal(err)
	}

	campaign := &campaigns.Campaign{
		Name:             "my-unique-name",
		NamespaceUserID:  userID,
		InitialApplierID: userID,
		LastApplierID:    userID,
		LastAppliedAt:    time.Now(),
		CampaignSpecID:   spec.ID,
	}
	if err := store.CreateCampaign(ctx, campaign); err != nil {
		t.Fatal(err)
	}

	changeset1 := createChangeset(t, ctx, store, testChangesetOpts{
		repo:                repo.ID,
		externalServiceType: "github",
		// Unpublished changesets should not be considered.
		publicationState: campaigns.ChangesetPublicationStateUnpublished,
		ownedByCampaign:  campaign.ID,
		campaign:         campaign.ID,
	})
	changeset2 := createChangeset(t, ctx, store, testChangesetOpts{
		repo:                repo.ID,
		externalServiceType: "github",
		externalState:       campaigns.ChangesetExternalStateOpen,
		publicationState:    campaigns.ChangesetPublicationStatePublished,
		ownedByCampaign:     campaign.ID,
		campaign:            campaign.ID,
		metadata: &github.PullRequest{
			CreatedAt: now.Add(-1 * 24 * time.Hour),
			TimelineItems: []github.TimelineItem{
				{Type: "MergedEvent", Item: &github.LabelEvent{
					CreatedAt: now,
				}},
			},
		},
	})

	addChangeset(t, ctx, store, campaign, changeset1.ID)
	addChangeset(t, ctx, store, campaign, changeset2.ID)

	s, err := graphqlbackend.NewSchema(&Resolver{store: store}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	campaignAPIID := string(campaigns.MarshalCampaignID(campaign.ID))
	input := map[string]interface{}{"campaign": campaignAPIID}
	var response struct{ Node apitest.Campaign }
	apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(userID)), t, s, input, &response, queryChangesetCountsConnection)

	wantCounts := []apitest.ChangesetCounts{
		{Date: marshalDateTime(t, now.Add(-7*24*time.Hour))},
		{Date: marshalDateTime(t, now.Add(-6*24*time.Hour))},
		{Date: marshalDateTime(t, now.Add(-5*24*time.Hour))},
		{Date: marshalDateTime(t, now.Add(-4*24*time.Hour))},
		{Date: marshalDateTime(t, now.Add(-3*24*time.Hour))},
		{Date: marshalDateTime(t, now.Add(-2*24*time.Hour))},
		{Date: marshalDateTime(t, now.Add(-1*24*time.Hour)), Total: 1, Open: 1, OpenPending: 1},
		{Date: marshalDateTime(t, now), Total: 1, Merged: 1},
	}

	if diff := cmp.Diff(wantCounts, response.Node.ChangesetCountsOverTime); diff != "" {
		t.Fatalf("wrong changesets response (-want +got):\n%s", diff)
	}
}

const queryChangesetCountsConnection = `
query($campaign: ID!) {
  node(id: $campaign) {
    ... on Campaign {
      changesetCountsOverTime {
        date
        total
        merged
        closed
        open
        openApproved
        openChangesRequested
        openPending
      }
    }
  }
}
`

func TestChangesetCountsOverTimeIntegration(t *testing.T) {
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

	err = repoStore.InsertRepos(ctx, githubRepo)
	if err != nil {
		t.Fatal(err)
	}

	store := ee.NewStore(dbconn.Global)

	spec := &campaigns.CampaignSpec{
		NamespaceUserID: userID,
		UserID:          userID,
	}
	if err := store.CreateCampaignSpec(ctx, spec); err != nil {
		t.Fatal(err)
	}

	campaign := &campaigns.Campaign{
		Name:             "Test campaign",
		Description:      "Testing changeset counts",
		InitialApplierID: userID,
		NamespaceUserID:  userID,
		LastApplierID:    userID,
		LastAppliedAt:    time.Now(),
		CampaignSpecID:   spec.ID,
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
			PublicationState:    campaigns.ChangesetPublicationStatePublished,
		},
		{
			RepoID:              githubRepo.ID,
			ExternalID:          "5849",
			ExternalServiceType: githubRepo.ExternalRepo.ServiceType,
			CampaignIDs:         []int64{campaign.ID},
			PublicationState:    campaigns.ChangesetPublicationStatePublished,
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
