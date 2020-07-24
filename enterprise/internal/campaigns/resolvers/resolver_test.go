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

	var campaignSpecs struct{ A, B apitest.CampaignSpec }
	apitest.MustExec(ctx, t, s, map[string]interface{}{
		"admin": users.Admin.ID,
		"org":   orgs.ACME.ID,
		"specA": `{"name":"specA"}`,
		"specB": `{"name":"specB"}`,
	}, &campaignSpecs, `
		fragment s on CampaignSpec {
			id
			namespace {
				id
			}
		}
		mutation($admin: ID!, $org: ID!, $specA: String!, $specB: String!) {
			A: createCampaignSpec(namespace: $admin, campaignSpec: $specA, changesetSpecs: []) { ...s }
			B: createCampaignSpec(namespace: $org, campaignSpec: $specB, changesetSpecs: [])   { ...s }
		}
	`)

	var campaigns struct{ A, B apitest.Campaign }
	apitest.MustExec(ctx, t, s, map[string]interface{}{
		"specA": campaignSpecs.A.ID,
		"specB": campaignSpecs.B.ID,
	}, &campaigns, `
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
		mutation($specA: ID!, $specB: ID!) {
			A: applyCampaign(campaignSpec: $specA) { ...c }
			B: applyCampaign(campaignSpec: $specB) { ...c }
		}
	`)

	if have, want := campaigns.A.Namespace.ID, users.Admin.ID; have != want {
		t.Fatalf("have admin's campaign namespace id %q, want %q", have, want)
	}

	if have, want := campaigns.B.Namespace.ID, orgs.ACME.ID; have != want {
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
	want := []apitest.Campaign{campaigns.A}
	if !reflect.DeepEqual(have, want) {
		t.Errorf("wrong campaigns listed. diff=%s", cmp.Diff(have, want))
	}

	if !listed.First.PageInfo.HasNextPage {
		t.Errorf("wrong page info: %+v", listed.First.PageInfo.HasNextPage)
	}

	have = listed.All.Nodes
	want = []apitest.Campaign{campaigns.A, campaigns.B}
	if !reflect.DeepEqual(have, want) {
		t.Errorf("wrong campaigns listed. diff=%s", cmp.Diff(have, want))
	}

	if listed.All.PageInfo.HasNextPage {
		t.Errorf("wrong page info: %+v", listed.All.PageInfo.HasNextPage)
	}

	// TODO: This updates a campaign
	// campaigns.Admin.Name = "Updated Admin Campaign Name"
	// campaigns.Admin.Description = "Updated Admin Campaign Description"
	// updateInput := map[string]interface{}{
	// 	"input": map[string]interface{}{
	// 		"id":          campaigns.Admin.ID,
	// 		"name":        campaigns.Admin.Name,
	// 		"description": campaigns.Admin.Description,
	// 	},
	// }
	// var updated struct{ UpdateCampaign apitest.Campaign }
	//
	// apitest.MustExec(ctx, t, s, updateInput, &updated, `
	// 	fragment u on User { id, databaseID, siteAdmin }
	// 	fragment o on Org  { id, name }
	// 	fragment c on Campaign {
	// 		id, name, description, createdAt, updatedAt
	// 		author    { ...u }
	// 		namespace {
	// 			... on User { ...u }
	// 			... on Org  { ...o }
	// 		}
	// 	}
	// 	mutation($input: UpdateCampaignInput!){
	// 		updateCampaign(input: $input) { ...c }
	// 	}
	// `)
	//
	// haveUpdated, wantUpdated := updated.UpdateCampaign, campaigns.Admin
	// if !reflect.DeepEqual(haveUpdated, wantUpdated) {
	// 	t.Errorf("wrong campaign updated. diff=%s", cmp.Diff(haveUpdated, wantUpdated))
	// }
	//
	// store := repos.NewDBStore(dbconn.Global, sql.TxOptions{})
	// githubExtSvc := &repos.ExternalService{
	// 	Kind:        extsvc.KindGitHub,
	// 	DisplayName: "GitHub",
	// 	Config: marshalJSON(t, &schema.GitHubConnection{
	// 		Url:   "https://github.com",
	// 		Token: os.Getenv("GITHUB_TOKEN"),
	// 		Repos: []string{"sourcegraph/sourcegraph"},
	// 	}),
	// }
	//
	// bbsURL := os.Getenv("BITBUCKET_SERVER_URL")
	// if bbsURL == "" {
	// 	// The test fixtures and golden files were generated with
	// 	// this config pointed to bitbucket.sgdev.org
	// 	bbsURL = "https://bitbucket.sgdev.org"
	// }
	//
	// bbsExtSvc := &repos.ExternalService{
	// 	Kind:        extsvc.KindBitbucketServer,
	// 	DisplayName: "Bitbucket Server",
	// 	Config: marshalJSON(t, &schema.BitbucketServerConnection{
	// 		Url:   bbsURL,
	// 		Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
	// 		Repos: []string{"SOUR/vegeta"},
	// 	}),
	// }
	//
	// err = store.UpsertExternalServices(ctx, githubExtSvc, bbsExtSvc)
	// if err != nil {
	// 	t.Fatal(t)
	// }
	//
	// githubSrc, err := repos.NewGithubSource(githubExtSvc, cf)
	// if err != nil {
	// 	t.Fatal(t)
	// }
	//
	// githubRepo, err := githubSrc.GetRepo(ctx, "sourcegraph/sourcegraph")
	// if err != nil {
	// 	t.Fatal(t)
	// }
	//
	// bbsSrc, err := repos.NewBitbucketServerSource(bbsExtSvc, cf)
	// if err != nil {
	// 	t.Fatal(t)
	// }
	//
	// bbsRepos := getBitbucketServerRepos(t, ctx, bbsSrc)
	// if len(bbsRepos) != 1 {
	// 	t.Fatalf("wrong number of bitbucket server repos. got=%d", len(bbsRepos))
	// }
	// bbsRepo := bbsRepos[0]
	//
	// err = store.UpsertRepos(ctx, githubRepo, bbsRepo)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	//
	// git.Mocks.ResolveRevision = func(spec string, opt git.ResolveRevisionOptions) (api.CommitID, error) {
	// 	return "mockcommitid", nil
	// }
	// defer func() { git.Mocks.ResolveRevision = nil }()
	//
	// repoupdater.MockRepoLookup = func(args protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
	// 	return &protocol.RepoLookupResult{
	// 		Repo: &protocol.RepoInfo{Name: args.Repo},
	// 	}, nil
	// }
	// defer func() { repoupdater.MockRepoLookup = nil }()
	//
	// var result struct {
	// 	Changesets []apitest.Changeset
	// }
	//
	// graphqlGithubRepoID := string(graphqlbackend.MarshalRepositoryID(githubRepo.ID))
	// graphqlBBSRepoID := string(graphqlbackend.MarshalRepositoryID(bbsRepo.ID))
	//
	// in := fmt.Sprintf(
	// 	`[{repository: %q, externalID: %q}, {repository: %q, externalID: %q}]`,
	// 	graphqlGithubRepoID, "999",
	// 	graphqlBBSRepoID, "2",
	// )
	//
	// apitest.MustExec(ctx, t, s, nil, &result, fmt.Sprintf(`
	// 	fragment gitRef on GitRef {
	// 		name
	// 		abbrevName
	// 		displayName
	// 		prefix
	// 		type
	// 		repository { id }
	// 		url
	// 		target {
	// 			oid
	// 			abbreviatedOID
	// 			type
	// 		}
	// 	}
	// 	fragment cs on ExternalChangeset {
	// 		id
	// 		repository { id }
	// 		createdAt
	// 		updatedAt
	// 		title
	// 		body
	// 		state
	// 		nextSyncAt
	// 		externalURL {
	// 			url
	// 			serviceType
	// 		}
	// 		reviewState
	// 		checkState
	// 		events(first: 100) {
	// 			totalCount
	// 		}
	// 		head { ...gitRef }
	// 		base { ...gitRef }
	// 	}
	// 	mutation() {
	// 		changesets: createChangesets(input: %s) {
	// 			...cs
	// 		}
	// 	}
	// `, in))
	//
	// {
	// 	want := []apitest.Changeset{
	// 		{
	// 			Repository: apitest.Repository{ID: graphqlGithubRepoID},
	// 			CreatedAt:  now.Format(time.RFC3339),
	// 			UpdatedAt:  now.Format(time.RFC3339),
	// 			Title:      "add extension filter to filter bar",
	// 			Body:       "Enables adding extension filters to the filter bar by rendering the extension filter as filter chips inside the filter bar.\r\nWIP for https://github.com/sourcegraph/sourcegraph/issues/962\r\n\r\n> This PR updates the CHANGELOG.md file to describe any user-facing changes.\r\n.\r\n",
	// 			State:      "MERGED",
	// 			ExternalURL: struct{ URL, ServiceType string }{
	// 				URL:         "https://github.com/sourcegraph/sourcegraph/pull/999",
	// 				ServiceType: extsvc.TypeGitHub,
	// 			},
	// 			ReviewState: "APPROVED",
	// 			CheckState:  "PASSED",
	// 			Events: apitest.ChangesetEventConnection{
	// 				TotalCount: 57,
	// 			},
	// 			// Not scheduled, not added to a campaign yet.
	// 			NextSyncAt: "",
	// 			Head: apitest.GitRef{
	// 				Name:        "refs/heads/vo/add-type-issue-filter",
	// 				AbbrevName:  "vo/add-type-issue-filter",
	// 				DisplayName: "vo/add-type-issue-filter",
	// 				Prefix:      "refs/heads/",
	// 				RefType:     "GIT_BRANCH",
	// 				Repository:  struct{ ID string }{ID: "UmVwb3NpdG9yeTox"},
	// 				URL:         "/github.com/sourcegraph/sourcegraph@vo/add-type-issue-filter",
	//
	// 				Target: apitest.GitTarget{
	// 					OID:            "23a5556c7e25aaab1f1529cee4efb90fe6fe3a30",
	// 					AbbreviatedOID: "23a5556",
	// 					TargetType:     "GIT_COMMIT",
	// 				},
	// 			},
	// 			Base: apitest.GitRef{
	// 				Name:        "refs/heads/master",
	// 				AbbrevName:  "master",
	// 				DisplayName: "master",
	// 				Prefix:      "refs/heads/",
	// 				RefType:     "GIT_BRANCH",
	// 				Repository:  struct{ ID string }{ID: "UmVwb3NpdG9yeTox"},
	// 				URL:         "/github.com/sourcegraph/sourcegraph@master",
	// 				Target: apitest.GitTarget{
	// 					OID:            "fa3815ba9ddd49db9111c5e9691e16d27e8f1f60",
	// 					AbbreviatedOID: "fa3815b",
	// 					TargetType:     "GIT_COMMIT",
	// 				},
	// 			},
	// 		},
	// 		{
	// 			Repository: apitest.Repository{ID: graphqlBBSRepoID},
	// 			CreatedAt:  now.Format(time.RFC3339),
	// 			UpdatedAt:  now.Format(time.RFC3339),
	// 			Title:      "Release testing pr",
	// 			Body:       "* Remove dump.go\r\n* make make make",
	// 			State:      "MERGED",
	// 			ExternalURL: struct{ URL, ServiceType string }{
	// 				URL:         "https://bitbucket.sgdev.org/projects/SOUR/repos/vegeta/pull-requests/2",
	// 				ServiceType: "bitbucketServer",
	// 			},
	// 			ReviewState: "PENDING",
	// 			CheckState:  "PENDING",
	// 			Events: apitest.ChangesetEventConnection{
	// 				TotalCount: 10,
	// 			},
	// 			// Not scheduled, not added to a campaign yet.
	// 			NextSyncAt: "",
	// 			Head: apitest.GitRef{
	// 				Name:        "refs/heads/release-testing-pr",
	// 				AbbrevName:  "release-testing-pr",
	// 				DisplayName: "release-testing-pr",
	// 				Prefix:      "refs/heads/",
	// 				RefType:     "GIT_BRANCH",
	// 				Repository:  struct{ ID string }{ID: "UmVwb3NpdG9yeToy"},
	// 				URL:         "/bitbucket.sgdev.org/SOUR/vegeta@release-testing-pr",
	// 				Target: apitest.GitTarget{
	// 					OID:            "be4d84e9c4b0a15e59c5f52900e6d55c7525b8d3",
	// 					AbbreviatedOID: "be4d84e",
	// 					TargetType:     "GIT_COMMIT",
	// 				},
	// 			},
	// 			Base: apitest.GitRef{
	// 				Name:        "refs/heads/master",
	// 				AbbrevName:  "master",
	// 				DisplayName: "master",
	// 				Prefix:      "refs/heads/",
	// 				RefType:     "GIT_BRANCH",
	// 				Repository:  struct{ ID string }{ID: "UmVwb3NpdG9yeToy"},
	// 				URL:         "/bitbucket.sgdev.org/SOUR/vegeta@master",
	// 				Target: apitest.GitTarget{
	// 					OID:            "mockcommitid",
	// 					AbbreviatedOID: "mockcom",
	// 					TargetType:     "GIT_COMMIT",
	// 				},
	// 			},
	// 		},
	// 	}
	//
	// 	have := make([]apitest.Changeset, 0, len(result.Changesets))
	// 	for _, c := range result.Changesets {
	// 		if c.ID == "" {
	// 			t.Fatal("Changeset ID is empty")
	// 		}
	//
	// 		c.ID = ""
	// 		have = append(have, c)
	// 	}
	// 	if diff := cmp.Diff(have, want); diff != "" {
	// 		t.Fatal(diff)
	// 	}
	//
	// 	// Test node resolver has nextSyncAt correctly set.
	// 	for _, c := range result.Changesets {
	// 		var changesetResult struct{ Node apitest.Changeset }
	// 		apitest.MustExec(ctx, t, s, nil, &changesetResult, fmt.Sprintf(`
	// 			query {
	// 				node(id: %q) {
	// 					... on ExternalChangeset {
	// 						nextSyncAt
	// 					}
	// 				}
	// 			}
	// 		`, c.ID))
	// 		if have, want := changesetResult.Node.NextSyncAt, ""; have != want {
	// 			t.Fatalf("incorrect nextSyncAt value, want=%q have=%q", want, have)
	// 		}
	// 	}
	// }
	//
	// var addChangesetsResult struct{ Campaign apitest.Campaign }
	//
	// changesetIDs := make([]string, 0, len(result.Changesets))
	// for _, c := range result.Changesets {
	// 	changesetIDs = append(changesetIDs, c.ID)
	// }
	//
	// // Date when PR #999 from above was created
	// countsFrom := parseJSONTime(t, "2018-11-14T22:07:45Z")
	// // Date when PR #999 from above was merged
	// countsTo := parseJSONTime(t, "2018-12-04T08:10:07Z")
	//
	// apitest.MustExec(ctx, t, s, nil, &addChangesetsResult, fmt.Sprintf(`
	// 	fragment u on User { id, databaseID, siteAdmin }
	// 	fragment o on Org  { id, name }
	//
	// 	fragment cs on ExternalChangeset {
	// 		id
	// 		repository { id }
	// 		createdAt
	// 		updatedAt
	// 		nextSyncAt
	// 		campaigns { nodes { id } }
	// 		title
	// 		body
	// 		state
	// 		externalURL {
	// 			url
	// 			serviceType
	// 		}
	// 		reviewState
	// 	}
	//
	// 	fragment c on Campaign {
	// 		id, name, description, createdAt, updatedAt
	// 		author    { ...u }
	// 		namespace {
	// 			... on User { ...u }
	// 			... on Org  { ...o }
	// 		}
	// 		changesets {
	// 			nodes {
	// 			  ... on ExternalChangeset {
	// 			    ...cs
	// 			  }
	// 			}
	// 			totalCount
	// 			pageInfo { hasNextPage }
	// 		}
	// 		changesetCountsOverTime(from: %s, to: %s) {
	// 		    date
	// 			total
	// 			merged
	// 			closed
	// 			open
	// 			openApproved
	// 			openChangesRequested
	// 			openPending
	// 		}
	// 		diffStat {
	// 			added
	// 			changed
	// 			deleted
	// 		}
	// 	}
	// 	mutation() {
	// 		campaign: addChangesetsToCampaign(campaign: %q, changesets: %s) {
	// 			...c
	// 		}
	// 	}
	// `,
	// 	marshalDateTime(t, countsFrom),
	// 	marshalDateTime(t, countsTo),
	// 	campaigns.Admin.ID,
	// 	marshalJSON(t, changesetIDs),
	// ))
	//
	// {
	// 	have := addChangesetsResult.Campaign.Changesets.TotalCount
	// 	want := len(changesetIDs)
	//
	// 	if have != want {
	// 		t.Fatalf(
	// 			"want campaign changesets totalcount %d, have=%d",
	// 			want, have,
	// 		)
	// 	}
	// }
	//
	// {
	// 	var have []string
	// 	want := changesetIDs
	//
	// 	for _, n := range addChangesetsResult.Campaign.Changesets.Nodes {
	// 		have = append(have, n.ID)
	// 	}
	//
	// 	if !reflect.DeepEqual(have, want) {
	// 		t.Errorf("wrong changesets added to campaign. want=%v, have=%v", want, have)
	// 	}
	// }
	//
	// {
	// 	have := map[string]bool{}
	// 	for _, cs := range addChangesetsResult.Campaign.Changesets.Nodes {
	// 		have[cs.Campaigns.Nodes[0].ID] = true
	// 	}
	//
	// 	if !have[campaigns.Admin.ID] || len(have) != 1 {
	// 		t.Errorf("wrong campaign added to changeset. want=%v, have=%v", campaigns.Admin.ID, have)
	// 	}
	// }
	//
	// {
	// 	counts := addChangesetsResult.Campaign.ChangesetCountsOverTime
	//
	// 	// There's 20 1-day intervals between countsFrom and including countsTo
	// 	if have, want := len(counts), 20; have != want {
	// 		t.Errorf("wrong changeset counts length %d, have=%d", want, have)
	// 	}
	//
	// 	for _, c := range counts {
	// 		if have, want := c.Total, int32(1); have != want {
	// 			t.Errorf("wrong changeset counts total %d, have=%d", want, have)
	// 		}
	// 	}
	// }
	//
	// {
	// 	have := addChangesetsResult.Campaign.DiffStat
	// 	// Expected DiffStat is zeros, because we don't return diffstats for
	// 	// closed changesets
	// 	want := apitest.DiffStat{Added: 0, Changed: 0, Deleted: 0}
	// 	if have != want {
	// 		t.Errorf("wrong campaign combined diffstat. want=%v, have=%v", want, have)
	// 	}
	// }
	//
	// {
	// 	for _, c := range addChangesetsResult.Campaign.Changesets.Nodes {
	// 		if have, want := c.NextSyncAt, now.Add(8*time.Hour).Format(time.RFC3339); have != want {
	// 			t.Fatalf("incorrect nextSyncAt value, want=%q have=%q", want, have)
	// 		}
	// 		var changesetResult struct{ Node apitest.Changeset }
	// 		apitest.MustExec(ctx, t, s, nil, &changesetResult, fmt.Sprintf(`
	// 			query {
	// 				node(id: %q) {
	// 					... on ExternalChangeset {
	// 						nextSyncAt
	// 					}
	// 				}
	// 			}
	// 		`, c.ID))
	// 		if have, want := changesetResult.Node.NextSyncAt, now.Add(8*time.Hour).Format(time.RFC3339); have != want {
	// 			t.Fatalf("incorrect nextSyncAt value, want=%q have=%q", want, have)
	// 		}
	// 	}
	// }
	//
	// deleteInput := map[string]interface{}{"id": campaigns.Admin.ID}
	// apitest.MustExec(ctx, t, s, deleteInput, &struct{}{}, `
	// 	mutation($id: ID!){
	// 		deleteCampaign(campaign: $id) { alwaysNil }
	// 	}
	// `)
	//
	// var campaignsAfterDelete struct {
	// 	Campaigns struct {
	// 		TotalCount int
	// 	}
	// }
	//
	// apitest.MustExec(ctx, t, s, nil, &campaignsAfterDelete, `
	// 	query { campaigns { totalCount } }
	// `)
	//
	// haveCount := campaignsAfterDelete.Campaigns.TotalCount
	// wantCount := listed.All.TotalCount - 1
	// if haveCount != wantCount {
	// 	t.Errorf("wrong campaigns totalcount after delete. want=%d, have=%d", wantCount, haveCount)
	// }
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
		Spec: campaigns.ChangesetSpecDescription{
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

	changesetSpec := &campaigns.ChangesetSpec{
		RawSpec: ct.NewRawChangesetSpecGitBranch(repoApiID, "d34db33f"),
		Spec: campaigns.ChangesetSpecDescription{
			BaseRepository: repoApiID,
		},
		RepoID: repo.ID,
		UserID: userID,
	}
	if err := store.CreateChangesetSpec(ctx, changesetSpec); err != nil {
		t.Fatal(err)
	}

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
		// TODO: Test for CampaignSpec/ChangesetSpecs once they're defined in
		// the schema.
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
