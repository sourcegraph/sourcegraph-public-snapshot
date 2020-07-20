package resolvers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-diff/diff"
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
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestCampaigns(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)
	rcache.SetupForTest(t)

	cf, save := httptestutil.NewGitHubRecorderFactory(t, *update, "test-campaigns")
	defer save()

	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := func() time.Time {
		return now.UTC().Truncate(time.Microsecond)
	}

	sr := &Resolver{
		store:       ee.NewStoreWithClock(dbconn.Global, clock),
		httpFactory: cf,
	}

	s, err := graphqlbackend.NewSchema(sr, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	var users struct {
		Admin, User struct {
			apitest.User `json:"user"`
		}
	}

	apitest.MustExec(ctx, t, s, nil, &users, `
		fragment u on User { id, databaseID, siteAdmin }
		mutation {
			admin: createUser(username: "admin") {
				user { ...u }
			}
			user: createUser(username: "user") {
				user { ...u }
			}
		}
	`)

	if !users.Admin.SiteAdmin {
		t.Fatal("admin must be a site-admin, since it was the first user created")
	}

	var orgs struct {
		ACME apitest.Org
	}

	ctx = actor.WithActor(ctx, actor.FromUser(users.Admin.DatabaseID))
	apitest.MustExec(ctx, t, s, nil, &orgs, `
		fragment o on Org { id, name }
		mutation {
			acme: createOrganization(name: "ACME") { ...o }
		}
	`)

	var campaigns struct{ Admin, Org apitest.Campaign }

	input := map[string]interface{}{
		"admin": map[string]interface{}{
			"namespace":   users.Admin.ID,
			"name":        "Admin Campaign",
			"description": "It's an admin's campaign",
		},
		"org": map[string]interface{}{
			"namespace":   orgs.ACME.ID,
			"name":        "ACME's Campaign",
			"description": "It's an ACME's campaign",
		},
	}

	apitest.MustExec(ctx, t, s, input, &campaigns, `
		fragment u on User { id, databaseID, siteAdmin }
		fragment o on Org  { id, name }
		fragment c on Campaign {
			id, name, description, createdAt, updatedAt
			author    { ...u }
			namespace {
				... on User { ...u }
				... on Org  { ...o }
			}
		}
		mutation($admin: CreateCampaignInput!, $org: CreateCampaignInput!){
			admin: createCampaign(input: $admin) { ...c }
			org: createCampaign(input: $org)     { ...c }
		}
	`)

	if have, want := campaigns.Admin.Namespace.ID, users.Admin.ID; have != want {
		t.Fatalf("have admin's campaign namespace id %q, want %q", have, want)
	}

	if have, want := campaigns.Org.Namespace.ID, orgs.ACME.ID; have != want {
		t.Fatalf("have orgs's campaign namespace id %q, want %q", have, want)
	}

	var listed struct{ First, All apitest.CampaignConnection }
	apitest.MustExec(ctx, t, s, nil, &listed, `
		fragment u on User { id, databaseID, siteAdmin }
		fragment o on Org  { id, name }
		fragment c on Campaign {
			id, name, description, createdAt, updatedAt
			author    { ...u }
			namespace {
				... on User { ...u }
				... on Org  { ...o }
			}
		}
		fragment n on CampaignConnection {
			nodes { ...c }
			totalCount
			pageInfo { hasNextPage }
		}
		query {
			first: campaigns(first: 1) { ...n }
			all: campaigns() { ...n }
		}
	`)

	have := listed.First.Nodes
	want := []apitest.Campaign{campaigns.Admin}
	if !reflect.DeepEqual(have, want) {
		t.Errorf("wrong campaigns listed. diff=%s", cmp.Diff(have, want))
	}

	if !listed.First.PageInfo.HasNextPage {
		t.Errorf("wrong page info: %+v", listed.First.PageInfo.HasNextPage)
	}

	have = listed.All.Nodes
	want = []apitest.Campaign{campaigns.Admin, campaigns.Org}
	if !reflect.DeepEqual(have, want) {
		t.Errorf("wrong campaigns listed. diff=%s", cmp.Diff(have, want))
	}

	if listed.All.PageInfo.HasNextPage {
		t.Errorf("wrong page info: %+v", listed.All.PageInfo.HasNextPage)
	}

	campaigns.Admin.Name = "Updated Admin Campaign Name"
	campaigns.Admin.Description = "Updated Admin Campaign Description"
	updateInput := map[string]interface{}{
		"input": map[string]interface{}{
			"id":          campaigns.Admin.ID,
			"name":        campaigns.Admin.Name,
			"description": campaigns.Admin.Description,
		},
	}
	var updated struct{ UpdateCampaign apitest.Campaign }

	apitest.MustExec(ctx, t, s, updateInput, &updated, `
		fragment u on User { id, databaseID, siteAdmin }
		fragment o on Org  { id, name }
		fragment c on Campaign {
			id, name, description, createdAt, updatedAt
			author    { ...u }
			namespace {
				... on User { ...u }
				... on Org  { ...o }
			}
		}
		mutation($input: UpdateCampaignInput!){
			updateCampaign(input: $input) { ...c }
		}
	`)

	haveUpdated, wantUpdated := updated.UpdateCampaign, campaigns.Admin
	if !reflect.DeepEqual(haveUpdated, wantUpdated) {
		t.Errorf("wrong campaign updated. diff=%s", cmp.Diff(haveUpdated, wantUpdated))
	}

	store := repos.NewDBStore(dbconn.Global, sql.TxOptions{})
	githubExtSvc := &repos.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GitHub",
		Config: marshalJSON(t, &schema.GitHubConnection{
			Url:   "https://github.com",
			Token: os.Getenv("GITHUB_TOKEN"),
			Repos: []string{"sourcegraph/sourcegraph"},
		}),
	}

	bbsURL := os.Getenv("BITBUCKET_SERVER_URL")
	if bbsURL == "" {
		// The test fixtures and golden files were generated with
		// this config pointed to bitbucket.sgdev.org
		bbsURL = "https://bitbucket.sgdev.org"
	}

	bbsExtSvc := &repos.ExternalService{
		Kind:        extsvc.KindBitbucketServer,
		DisplayName: "Bitbucket Server",
		Config: marshalJSON(t, &schema.BitbucketServerConnection{
			Url:   bbsURL,
			Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
			Repos: []string{"SOUR/vegeta"},
		}),
	}

	err = store.UpsertExternalServices(ctx, githubExtSvc, bbsExtSvc)
	if err != nil {
		t.Fatal(t)
	}

	githubSrc, err := repos.NewGithubSource(githubExtSvc, cf)
	if err != nil {
		t.Fatal(t)
	}

	githubRepo, err := githubSrc.GetRepo(ctx, "sourcegraph/sourcegraph")
	if err != nil {
		t.Fatal(t)
	}

	bbsSrc, err := repos.NewBitbucketServerSource(bbsExtSvc, cf)
	if err != nil {
		t.Fatal(t)
	}

	bbsRepos := getBitbucketServerRepos(t, ctx, bbsSrc)
	if len(bbsRepos) != 1 {
		t.Fatalf("wrong number of bitbucket server repos. got=%d", len(bbsRepos))
	}
	bbsRepo := bbsRepos[0]

	err = store.UpsertRepos(ctx, githubRepo, bbsRepo)
	if err != nil {
		t.Fatal(err)
	}

	git.Mocks.ResolveRevision = func(spec string, opt git.ResolveRevisionOptions) (api.CommitID, error) {
		return "mockcommitid", nil
	}
	defer func() { git.Mocks.ResolveRevision = nil }()

	var result struct {
		Changesets []apitest.Changeset
	}

	graphqlGithubRepoID := string(graphqlbackend.MarshalRepositoryID(githubRepo.ID))
	graphqlBBSRepoID := string(graphqlbackend.MarshalRepositoryID(bbsRepo.ID))

	in := fmt.Sprintf(
		`[{repository: %q, externalID: %q}, {repository: %q, externalID: %q}]`,
		graphqlGithubRepoID, "999",
		graphqlBBSRepoID, "2",
	)

	state := ct.MockChangesetSyncState(&protocol.RepoInfo{
		Name: api.RepoName(githubRepo.Name),
		VCS:  protocol.VCSInfo{URL: githubRepo.URI},
	})
	defer state.Unmock()

	apitest.MustExec(ctx, t, s, nil, &result, fmt.Sprintf(`
		fragment gitRef on GitRef {
			name
			abbrevName
			displayName
			prefix
			type
			repository { id }
			url
			target {
				oid
				abbreviatedOID
				type
			}
		}
		fragment cs on ExternalChangeset {
			id
			repository { id }
			createdAt
			updatedAt
			title
			body
			state
			nextSyncAt
			externalURL {
				url
				serviceType
			}
			reviewState
			checkState
			events(first: 100) {
				totalCount
			}
			head { ...gitRef }
			base { ...gitRef }
		}
		mutation() {
			changesets: createChangesets(input: %s) {
				...cs
			}
		}
	`, in))

	{
		want := []apitest.Changeset{
			{
				Repository: apitest.Repository{ID: graphqlGithubRepoID},
				CreatedAt:  now.Format(time.RFC3339),
				UpdatedAt:  now.Format(time.RFC3339),
				Title:      "add extension filter to filter bar",
				Body:       "Enables adding extension filters to the filter bar by rendering the extension filter as filter chips inside the filter bar.\r\nWIP for https://github.com/sourcegraph/sourcegraph/issues/962\r\n\r\n> This PR updates the CHANGELOG.md file to describe any user-facing changes.\r\n.\r\n",
				State:      "MERGED",
				ExternalURL: struct{ URL, ServiceType string }{
					URL:         "https://github.com/sourcegraph/sourcegraph/pull/999",
					ServiceType: extsvc.TypeGitHub,
				},
				ReviewState: "APPROVED",
				CheckState:  "PASSED",
				Events: apitest.ChangesetEventConnection{
					TotalCount: 57,
				},
				// Not scheduled, not added to a campaign yet.
				NextSyncAt: "",
				Head: apitest.GitRef{
					Name:        "refs/heads/vo/add-type-issue-filter",
					AbbrevName:  "vo/add-type-issue-filter",
					DisplayName: "vo/add-type-issue-filter",
					Prefix:      "refs/heads/",
					RefType:     "GIT_BRANCH",
					Repository:  struct{ ID string }{ID: "UmVwb3NpdG9yeTox"},
					URL:         "/github.com/sourcegraph/sourcegraph@vo/add-type-issue-filter",

					Target: apitest.GitTarget{
						OID:            "23a5556c7e25aaab1f1529cee4efb90fe6fe3a30",
						AbbreviatedOID: "23a5556",
						TargetType:     "GIT_COMMIT",
					},
				},
				Base: apitest.GitRef{
					Name:        "refs/heads/master",
					AbbrevName:  "master",
					DisplayName: "master",
					Prefix:      "refs/heads/",
					RefType:     "GIT_BRANCH",
					Repository:  struct{ ID string }{ID: "UmVwb3NpdG9yeTox"},
					URL:         "/github.com/sourcegraph/sourcegraph@master",
					Target: apitest.GitTarget{
						OID:            "fa3815ba9ddd49db9111c5e9691e16d27e8f1f60",
						AbbreviatedOID: "fa3815b",
						TargetType:     "GIT_COMMIT",
					},
				},
			},
			{
				Repository: apitest.Repository{ID: graphqlBBSRepoID},
				CreatedAt:  now.Format(time.RFC3339),
				UpdatedAt:  now.Format(time.RFC3339),
				Title:      "Release testing pr",
				Body:       "* Remove dump.go\r\n* make make make",
				State:      "MERGED",
				ExternalURL: struct{ URL, ServiceType string }{
					URL:         "https://bitbucket.sgdev.org/projects/SOUR/repos/vegeta/pull-requests/2",
					ServiceType: "bitbucketServer",
				},
				ReviewState: "PENDING",
				CheckState:  "PENDING",
				Events: apitest.ChangesetEventConnection{
					TotalCount: 10,
				},
				// Not scheduled, not added to a campaign yet.
				NextSyncAt: "",
				Head: apitest.GitRef{
					Name:        "refs/heads/release-testing-pr",
					AbbrevName:  "release-testing-pr",
					DisplayName: "release-testing-pr",
					Prefix:      "refs/heads/",
					RefType:     "GIT_BRANCH",
					Repository:  struct{ ID string }{ID: "UmVwb3NpdG9yeToy"},
					URL:         "/bitbucket.sgdev.org/SOUR/vegeta@release-testing-pr",
					Target: apitest.GitTarget{
						OID:            "be4d84e9c4b0a15e59c5f52900e6d55c7525b8d3",
						AbbreviatedOID: "be4d84e",
						TargetType:     "GIT_COMMIT",
					},
				},
				Base: apitest.GitRef{
					Name:        "refs/heads/master",
					AbbrevName:  "master",
					DisplayName: "master",
					Prefix:      "refs/heads/",
					RefType:     "GIT_BRANCH",
					Repository:  struct{ ID string }{ID: "UmVwb3NpdG9yeToy"},
					URL:         "/bitbucket.sgdev.org/SOUR/vegeta@master",
					Target: apitest.GitTarget{
						OID:            "mockcommitid",
						AbbreviatedOID: "mockcom",
						TargetType:     "GIT_COMMIT",
					},
				},
			},
		}

		have := make([]apitest.Changeset, 0, len(result.Changesets))
		for _, c := range result.Changesets {
			if c.ID == "" {
				t.Fatal("Changeset ID is empty")
			}

			c.ID = ""
			have = append(have, c)
		}
		if diff := cmp.Diff(have, want); diff != "" {
			t.Fatal(diff)
		}

		// Test node resolver has nextSyncAt correctly set.
		for _, c := range result.Changesets {
			var changesetResult struct{ Node apitest.Changeset }
			apitest.MustExec(ctx, t, s, nil, &changesetResult, fmt.Sprintf(`
				query {
					node(id: %q) {
						... on ExternalChangeset {
							nextSyncAt
						}
					}
				}
			`, c.ID))
			if have, want := changesetResult.Node.NextSyncAt, ""; have != want {
				t.Fatalf("incorrect nextSyncAt value, want=%q have=%q", want, have)
			}
		}
	}

	var addChangesetsResult struct{ Campaign apitest.Campaign }

	changesetIDs := make([]string, 0, len(result.Changesets))
	for _, c := range result.Changesets {
		changesetIDs = append(changesetIDs, c.ID)
	}

	// Date when PR #999 from above was created
	countsFrom := parseJSONTime(t, "2018-11-14T22:07:45Z")
	// Date when PR #999 from above was merged
	countsTo := parseJSONTime(t, "2018-12-04T08:10:07Z")

	apitest.MustExec(ctx, t, s, nil, &addChangesetsResult, fmt.Sprintf(`
		fragment u on User { id, databaseID, siteAdmin }
		fragment o on Org  { id, name }

		fragment cs on ExternalChangeset {
			id
			repository { id }
			createdAt
			updatedAt
			nextSyncAt
			campaigns { nodes { id } }
			title
			body
			state
			externalURL {
				url
				serviceType
			}
			reviewState
		}

		fragment c on Campaign {
			id, name, description, createdAt, updatedAt
			author    { ...u }
			namespace {
				... on User { ...u }
				... on Org  { ...o }
			}
			changesets {
				nodes {
				  ... on ExternalChangeset {
				    ...cs
				  }
				}
				totalCount
				pageInfo { hasNextPage }
			}
			changesetCountsOverTime(from: %s, to: %s) {
			    date
				total
				merged
				closed
				open
				openApproved
				openChangesRequested
				openPending
			}
			diffStat {
				added
				changed
				deleted
			}
		}
		mutation() {
			campaign: addChangesetsToCampaign(campaign: %q, changesets: %s) {
				...c
			}
		}
	`,
		marshalDateTime(t, countsFrom),
		marshalDateTime(t, countsTo),
		campaigns.Admin.ID,
		marshalJSON(t, changesetIDs),
	))

	{
		have := addChangesetsResult.Campaign.Changesets.TotalCount
		want := len(changesetIDs)

		if have != want {
			t.Fatalf(
				"want campaign changesets totalcount %d, have=%d",
				want, have,
			)
		}
	}

	{
		var have []string
		want := changesetIDs

		for _, n := range addChangesetsResult.Campaign.Changesets.Nodes {
			have = append(have, n.ID)
		}

		if !reflect.DeepEqual(have, want) {
			t.Errorf("wrong changesets added to campaign. want=%v, have=%v", want, have)
		}
	}

	{
		have := map[string]bool{}
		for _, cs := range addChangesetsResult.Campaign.Changesets.Nodes {
			have[cs.Campaigns.Nodes[0].ID] = true
		}

		if !have[campaigns.Admin.ID] || len(have) != 1 {
			t.Errorf("wrong campaign added to changeset. want=%v, have=%v", campaigns.Admin.ID, have)
		}
	}

	{
		counts := addChangesetsResult.Campaign.ChangesetCountsOverTime

		// There's 20 1-day intervals between countsFrom and including countsTo
		if have, want := len(counts), 20; have != want {
			t.Errorf("wrong changeset counts length %d, have=%d", want, have)
		}

		for _, c := range counts {
			if have, want := c.Total, int32(1); have != want {
				t.Errorf("wrong changeset counts total %d, have=%d", want, have)
			}
		}
	}

	{
		have := addChangesetsResult.Campaign.DiffStat
		want := apitest.DiffStat{Added: 2, Changed: 2, Deleted: 6}
		if have != want {
			t.Errorf("wrong campaign combined diffstat. want=%v, have=%v", want, have)
		}
	}

	{
		for _, c := range addChangesetsResult.Campaign.Changesets.Nodes {
			if have, want := c.NextSyncAt, now.Add(8*time.Hour).Format(time.RFC3339); have != want {
				t.Fatalf("incorrect nextSyncAt value, want=%q have=%q", want, have)
			}
			var changesetResult struct{ Node apitest.Changeset }
			apitest.MustExec(ctx, t, s, nil, &changesetResult, fmt.Sprintf(`
				query {
					node(id: %q) {
						... on ExternalChangeset {
							nextSyncAt
						}
					}
				}
			`, c.ID))
			if have, want := changesetResult.Node.NextSyncAt, now.Add(8*time.Hour).Format(time.RFC3339); have != want {
				t.Fatalf("incorrect nextSyncAt value, want=%q have=%q", want, have)
			}
		}
	}

	deleteInput := map[string]interface{}{"id": campaigns.Admin.ID}
	apitest.MustExec(ctx, t, s, deleteInput, &struct{}{}, `
		mutation($id: ID!){
			deleteCampaign(campaign: $id) { alwaysNil }
		}
	`)

	var campaignsAfterDelete struct {
		Campaigns struct {
			TotalCount int
		}
	}

	apitest.MustExec(ctx, t, s, nil, &campaignsAfterDelete, `
		query { campaigns { totalCount } }
	`)

	haveCount := campaignsAfterDelete.Campaigns.TotalCount
	wantCount := listed.All.TotalCount - 1
	if haveCount != wantCount {
		t.Errorf("wrong campaigns totalcount after delete. want=%d, have=%d", wantCount, haveCount)
	}
}

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

	err = store.CreateChangesets(ctx, changesets...)
	if err != nil {
		t.Fatal(err)
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

	for _, c := range changesets {
		campaign.ChangesetIDs = append(campaign.ChangesetIDs, c.ID)
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
		marshalPatchSetID(0),
		marshalPatchID(0),
		campaigns.MarshalCampaignID(0),
		marshalExternalChangesetID(0),
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
		fmt.Sprintf(`mutation { retryCampaignChangesets(campaign: %q) { id } }`, campaigns.MarshalCampaignID(0)),
		fmt.Sprintf(`mutation { closeCampaign(campaign: %q) { id } }`, campaigns.MarshalCampaignID(0)),
		fmt.Sprintf(`mutation { deleteCampaign(campaign: %q) { alwaysNil } }`, campaigns.MarshalCampaignID(0)),
		fmt.Sprintf(`mutation { publishChangeset(patch: %q) { alwaysNil } }`, marshalPatchID(0)),
		fmt.Sprintf(`mutation { syncChangeset(changeset: %q) { alwaysNil } }`, marshalExternalChangesetID(0)),
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

func TestCreatePatchSetFromPatchesResolver(t *testing.T) {
	ctx := backend.WithAuthzBypass(context.Background())

	dbtesting.SetupGlobalTestDB(t)

	userID := insertTestUser(t, dbconn.Global, "create-patch-set", false)
	act := actor.FromUser(userID)
	ctx = actor.WithActor(ctx, act)

	t.Run("invalid patch", func(t *testing.T) {
		args := graphqlbackend.CreatePatchSetFromPatchesArgs{
			Patches: []graphqlbackend.PatchInput{
				{
					Repository:   graphqlbackend.MarshalRepositoryID(1),
					BaseRevision: "f00b4r",
					BaseRef:      "master",
					Patch:        "!!! this is not a valid unified diff !!!\n--- x\n+++ y\n@@ 1,1 2,2\na",
				},
			},
		}

		_, err := (&Resolver{}).CreatePatchSetFromPatches(ctx, args)
		if err == nil {
			t.Fatal("want error")
		}
		if _, ok := errors.Cause(err).(*diff.ParseError); !ok {
			t.Fatalf("got error %q (%T), want a diff ParseError", err, errors.Cause(err))
		}
	})

	t.Run("integration", func(t *testing.T) {
		if testing.Short() {
			t.Skip()
		}

		rcache.SetupForTest(t)

		now := time.Now().UTC().Truncate(time.Microsecond)
		clock := func() time.Time {
			return now.UTC().Truncate(time.Microsecond)
		}

		// For testing purposes they all share the same rev, across repos
		testingRev := api.CommitID("24f7ca7c1190835519e261d7eefa09df55ceea4f")
		mockBackendCommits(t, testingRev)

		reposStore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})
		repo := newGitHubTestRepo("github.com/sourcegraph/sourcegraph", 1)
		if err := reposStore.UpsertRepos(ctx, repo); err != nil {
			t.Fatal(err)
		}

		store := ee.NewStoreWithClock(dbconn.Global, clock)

		sr := &Resolver{store: store}
		s, err := graphqlbackend.NewSchema(sr, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		var response struct{ CreatePatchSetFromPatches apitest.PatchSet }

		apitest.MustExec(ctx, t, s, nil, &response, fmt.Sprintf(`
      mutation {
		createPatchSetFromPatches(patches: [{repository: %q, baseRevision: %q, baseRef: "master", patch: %q}]) {
          ... on PatchSet {
            id
            patches(first: %d) {
              nodes {
			    ... on Patch {
                  repository {
                    name
                  }
				  diff {
                    fileDiffs {
                      rawDiff
                      diffStat {
                        added
                        deleted
                        changed
                      }
                      nodes {
                        oldPath
                        newPath
                        hunks {
                          body
                          section
                          newRange { startLine, lines }
                          oldRange { startLine, lines }
                          oldNoNewlineAt
                        }
                        stat {
                          added
                          deleted
                          changed
                        }
                        oldFile {
                          name
                          externalURLs {
                            serviceType
                            url
                          }
                        }
                      }
                    }
                  }
				}
			  }
            }
            previewURL
            diffStat {
              added
              deleted
              changed
            }
          }
        }
      }
	`, graphqlbackend.MarshalRepositoryID(repo.ID), testingRev, testDiff, 1))

		result := response.CreatePatchSetFromPatches

		wantPatches := []apitest.Patch{
			{
				Repository: struct{ Name, URL string }{Name: repo.Name},
				Diff:       struct{ FileDiffs apitest.FileDiffs }{FileDiffs: testDiffGraphQL},
			},
		}
		if !cmp.Equal(result.Patches.Nodes, wantPatches) {
			t.Error("wrong patches", cmp.Diff(result.Patches.Nodes, wantPatches))
		}

		if have, want := result.PreviewURL, "http://example.com/campaigns/new?patchSet=UGF0Y2hTZXQ6MQ%3D%3D"; have != want {
			t.Fatalf("have PreviewURL %q, want %q", have, want)
		}

		if have, want := result.DiffStat, (apitest.DiffStat{Changed: 2}); have != want {
			t.Fatalf("wrong PatchSet.DiffStat.Changed %d, want=%d", have, want)
		}
	})
}

func TestCreateCampaignWithPatchSet(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	dbtesting.SetupGlobalTestDB(t)
	rcache.SetupForTest(t)

	ctx := backend.WithAuthzBypass(context.Background())
	userID := insertTestUser(t, dbconn.Global, "create-patch-set", true)
	act := actor.FromUser(userID)
	ctx = actor.WithActor(ctx, act)

	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := func() time.Time {
		return now.UTC().Truncate(time.Microsecond)
	}

	testBaseRevision := api.CommitID("24f7ca7c1190835519e261d7eefa09df55ceea4f")
	testBaseRef := "refs/heads/master"
	testHeadRef := "refs/heads/my-cool-branch"

	// gitserver Mocks
	mockBackendCommits(t, testBaseRevision)

	git.Mocks.MergeBase = func(repo gitserver.Repo, a, b api.CommitID) (api.CommitID, error) {
		if string(a) != testBaseRef || string(b) != testHeadRef {
			t.Fatalf("gitserver.MergeBase received wrong args: %s %s", a, b)
		}
		return testBaseRevision, nil
	}
	t.Cleanup(func() { git.Mocks.MergeBase = nil })

	// repo & external service setup
	reposStore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})
	ext := &repos.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GitHub",
		Config: marshalJSON(t, &schema.GitHubConnection{
			Url:   "https://github.com",
			Token: "SECRETTOKEN",
		}),
	}

	if err := reposStore.UpsertExternalServices(ctx, ext); err != nil {
		t.Fatal(err)
	}

	repo := newGitHubTestRepo("github.com/sourcegraph/sourcegraph", 1)
	repo.Sources = map[string]*repos.SourceInfo{ext.URN(): {ID: ext.URN()}}

	if err := reposStore.UpsertRepos(ctx, repo); err != nil {
		t.Fatal(err)
	}

	// Setup schema resolver
	store := ee.NewStoreWithClock(dbconn.Global, clock)
	sr := &Resolver{store: store}
	s, err := graphqlbackend.NewSchema(sr, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Start test
	var createPatchSetResponse struct{ CreatePatchSetFromPatches apitest.PatchSet }
	apitest.MustExec(ctx, t, s, nil, &createPatchSetResponse, fmt.Sprintf(`
		mutation {
			createPatchSetFromPatches(patches: [{repository: %q, baseRevision: %q, baseRef: %q, patch: %q}]) {
				... on PatchSet {
					id
					previewURL
				}
			}
		}
	`, graphqlbackend.MarshalRepositoryID(repo.ID), testBaseRevision, testBaseRef, testDiff))

	patchSetID := createPatchSetResponse.CreatePatchSetFromPatches.ID

	var createCampaignResponse struct{ CreateCampaign apitest.Campaign }

	input := map[string]interface{}{
		"input": map[string]interface{}{
			"namespace":   string(graphqlbackend.MarshalUserID(userID)),
			"name":        "Campaign with PatchSet",
			"description": "This campaign has a patchset",
			"patchSet":    patchSetID,
			"branch":      "my-cool-branch",
		},
	}

	apitest.MustExec(ctx, t, s, input, &createCampaignResponse, `
    fragment c on Campaign {
      id
      branch
      status { state }
      hasUnpublishedPatches
      patches {
        nodes {
          ... on HiddenPatch {
            id
          }
          ... on Patch {
            id
            publicationEnqueued
            repository {
              name
            }
            diff {
              fileDiffs {
                rawDiff
                diffStat {
                  added
                  deleted
                  changed
                }
                nodes {
                  oldPath
                  newPath
                  hunks {
                    body
                    section
                    newRange { startLine, lines }
                    oldRange { startLine, lines }
                    oldNoNewlineAt
                  }
                  stat {
                    added
                    deleted
                    changed
                  }
                  oldFile {
                    name
                    externalURLs {
                      serviceType
                      url
                    }
                  }
                }
              }
            }
		  }
        }
      }
      diffStat {
        added
        deleted
        changed
      }
    }

    mutation($input: CreateCampaignInput!) {
      createCampaign(input: $input) { ...c }
    }
	`)

	campaign := createCampaignResponse.CreateCampaign
	if campaign.ID == "" {
		log.Fatalf("Campaign does not have ID!")
	}

	if have, want := len(campaign.Patches.Nodes), 1; have != want {
		log.Fatalf("wrong length of patches. want=%d, have=%d", want, have)
	}

	if campaign.DiffStat.Changed != 2 {
		t.Fatalf("diffstat is wrong: %+v", campaign.DiffStat)
	}

	if !campaign.HasUnpublishedPatches {
		t.Errorf("campaign HasUnpublishedPatches is false, want true")
	}

	patch := campaign.Patches.Nodes[0]
	if have, want := campaign.DiffStat, patch.Diff.FileDiffs.DiffStat; have != want {
		t.Errorf("wrong campaign combined diffstat. want=%v, have=%v", want, have)
	}

	if patch.PublicationEnqueued {
		t.Errorf("patch PublicationEnqueued is true, want false")
	}

	// Publish the changesets in the campaign
	for _, p := range campaign.Patches.Nodes {
		var res struct{}
		input := map[string]interface{}{"patch": p.ID}
		q := `mutation($patch: ID!) { publishChangeset(patch: $patch) { alwaysNil } }`
		apitest.MustExec(ctx, t, s, input, &res, q)
	}

	// Now we need to run the created ChangsetJob
	changesetJobs, _, err := store.ListChangesetJobs(ctx, ee.ListChangesetJobsOpts{})
	if err != nil {
		t.Fatal(err)
	}

	if len(changesetJobs) != 1 {
		t.Fatalf("wrong number of changeset jobs created: %d", len(changesetJobs))
	}

	headRef := "refs/heads/" + campaign.Branch

	fakePR := &github.PullRequest{
		ID:          "FOOBARID",
		Title:       campaign.Name,
		Body:        campaign.Description,
		BaseRefName: git.AbbreviateRef(testBaseRef),
		HeadRefName: git.AbbreviateRef(headRef),
		Number:      12345,
		State:       "OPEN",
		TimelineItems: []github.TimelineItem{
			{Type: "PullRequestCommit", Item: &github.PullRequestCommit{
				Commit: github.Commit{
					OID:           "new-f00bar",
					PushedDate:    now,
					CommittedDate: now,
				},
			}},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	gitClient := &ct.FakeGitserverClient{Response: headRef, ResponseErr: nil}

	sourcer := repos.NewFakeSourcer(nil, &ct.FakeChangesetSource{
		Svc:          ext,
		WantHeadRef:  headRef,
		WantBaseRef:  testBaseRef,
		FakeMetadata: fakePR,
	})

	state := ct.MockChangesetSyncState(&protocol.RepoInfo{
		Name: api.RepoName(repo.Name),
		VCS:  protocol.VCSInfo{URL: repo.URI},
	})
	defer state.Unmock()

	job := changesetJobs[0]

	c, err := store.GetCampaign(ctx, ee.GetCampaignOpts{ID: job.CampaignID})
	if err != nil {
		t.Fatal(err)
	}

	err = ee.ExecChangesetJob(ctx, c, job, ee.ExecChangesetJobOpts{
		Clock: clock, Store: store, GitClient: gitClient, Sourcer: sourcer, ExternalURL: "http://localhost",
	})
	if err != nil {
		t.Fatal(err)
	}

	updatedJob, err := store.GetChangesetJob(ctx, ee.GetChangesetJobOpts{ID: job.ID})
	if err != nil {
		t.Fatal(err)
	}
	if updatedJob.ChangesetID == 0 {
		t.Fatal("ChangesetJob.ChangesetID has not been updated")
	}

	cs, err := store.GetChangeset(ctx, ee.GetChangesetOpts{
		ID: job.ChangesetID,
	})
	if err != nil {
		t.Fatal(err)
	}
	cs.SetDiffStat(&diff.Stat{Added: 1, Changed: 1, Deleted: 3})
	if err := store.UpdateChangesets(ctx, cs); err != nil {
		t.Fatal(err)
	}

	// We need to setup these mocks because the GraphQL now needs to talk to
	// gitserver to calculate the diff for a changeset.
	git.Mocks.GetCommit = func(api.CommitID) (*git.Commit, error) {
		return &git.Commit{ID: testBaseRevision}, nil
	}
	defer func() { git.Mocks.GetCommit = nil }()

	var queryCampaignResponse struct{ Node apitest.Campaign }

	apitest.MustExec(ctx, t, s, nil, &queryCampaignResponse, fmt.Sprintf(`
	    fragment c on Campaign {
	      id
	      status { state }
	      hasUnpublishedPatches
	      branch
	      patches {
	        totalCount
	      }
	      changesets {
	        nodes {
			  ... on ExternalChangeset {
				state
	            diff {
	              fileDiffs {
	                diffStat {
	                  added
	                  deleted
	                  changed
	                }
	              }
	            }
			  }
	        }
	        totalCount
	      }
	      openChangesets {
	        totalCount
	      }
	      diffStat {
	        added
	        deleted
	        changed
	      }
	    }

	    query {
	      node(id: %q) { ...c }
	    }
	`, campaign.ID))

	campaign = queryCampaignResponse.Node
	if campaign.Status.State != "COMPLETED" {
		t.Fatalf("campaign is not in state 'COMPLETED': %q", campaign.Status.State)
	}

	if campaign.HasUnpublishedPatches {
		t.Errorf("campaign HasUnpublishedPatches is true, want false")
	}

	if campaign.Patches.TotalCount != 0 {
		t.Fatalf("campaign.Patches.TotalCount is not zero: %d", campaign.Patches.TotalCount)
	}

	if campaign.OpenChangesets.TotalCount != 1 {
		t.Fatalf("campaign.OpenChangesets.TotalCount is not 1: %d", campaign.OpenChangesets.TotalCount)
	}
	if campaign.Changesets.TotalCount != 1 {
		t.Fatalf("campaign.Changesets.TotalCount is not 1: %d", campaign.Changesets.TotalCount)
	}

	if campaign.DiffStat.Changed != 1 {
		t.Fatalf("diffstat is wrong: %+v", campaign.DiffStat)
	}

	changeset := campaign.Changesets.Nodes[0]
	if have, want := campaign.DiffStat, changeset.Diff.FileDiffs.DiffStat; have != want {
		t.Errorf("wrong campaign combined diffstat. want=%v, have=%v", want, have)
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
