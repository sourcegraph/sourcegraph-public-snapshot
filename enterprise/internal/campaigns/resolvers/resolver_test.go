package resolvers

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/google/go-cmp/cmp"
	graphql "github.com/graph-gophers/graphql-go"
	gqlerrors "github.com/graph-gophers/graphql-go/errors"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/schema"
)

func init() {
	dbtesting.DBNameSuffix = "campaignsresolversdb"
}

var update = flag.Bool("update", false, "update testdata")

func TestCampaigns(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)
	rcache.SetupForTest(t)

	cf, save := newGithubClientFactory(t, "test-campaigns")
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

	type User struct {
		ID         string
		DatabaseID int32
		SiteAdmin  bool
	}

	var users struct {
		Admin, User struct {
			User `json:"user"`
		}
	}

	mustExec(ctx, t, s, nil, &users, `
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

	type Org struct {
		ID   string
		Name string
	}

	var orgs struct {
		ACME Org
	}

	ctx = actor.WithActor(ctx, actor.FromUser(users.Admin.DatabaseID))
	mustExec(ctx, t, s, nil, &orgs, `
		fragment o on Org { id, name }
		mutation {
			acme: createOrganization(name: "ACME") { ...o }
		}
	`)

	type UserOrg struct {
		ID         string
		DatabaseID int32
		SiteAdmin  bool
		Name       string
	}

	type Campaign struct {
		ID          string
		Name        string
		Description string
		Author      User
		CreatedAt   string
		UpdatedAt   string
		PublishedAt string
		Namespace   UserOrg
	}

	var campaigns struct{ Admin, Org Campaign }

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

	mustExec(ctx, t, s, input, &campaigns, `
		fragment u on User { id, databaseID, siteAdmin }
		fragment o on Org  { id, name }
		fragment c on Campaign {
			id, name, description, createdAt, updatedAt, publishedAt
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

	type CampaignConnection struct {
		Nodes      []Campaign
		TotalCount int
		PageInfo   struct {
			HasNextPage bool
		}
	}

	var listed struct {
		First, All CampaignConnection
	}

	mustExec(ctx, t, s, nil, &listed, `
		fragment u on User { id, databaseID, siteAdmin }
		fragment o on Org  { id, name }
		fragment c on Campaign {
			id, name, description, createdAt, updatedAt, publishedAt
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
	want := []Campaign{campaigns.Admin}
	if !reflect.DeepEqual(have, want) {
		t.Errorf("wrong campaigns listed. diff=%s", cmp.Diff(have, want))
	}

	if !listed.First.PageInfo.HasNextPage {
		t.Errorf("wrong page info: %+v", listed.First.PageInfo.HasNextPage)
	}

	have = listed.All.Nodes
	want = []Campaign{campaigns.Admin, campaigns.Org}
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
	var updated struct {
		UpdateCampaign Campaign
	}

	mustExec(ctx, t, s, updateInput, &updated, `
		fragment u on User { id, databaseID, siteAdmin }
		fragment o on Org  { id, name }
		fragment c on Campaign {
			id, name, description, createdAt, updatedAt, publishedAt
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
		Kind:        "GITHUB",
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
		Kind:        "BITBUCKETSERVER",
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

	type ChangesetEventConnection struct {
		TotalCount int
	}

	git.Mocks.ResolveRevision = func(spec string, opt *git.ResolveRevisionOptions) (api.CommitID, error) {
		return "mockcommitid", nil
	}
	defer func() { git.Mocks.ResolveRevision = nil }()

	repoupdater.MockRepoLookup = func(args protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
		return &protocol.RepoLookupResult{
			Repo: &protocol.RepoInfo{Name: args.Repo},
		}, nil
	}
	defer func() { repoupdater.MockRepoLookup = nil }()

	type GitTarget struct {
		OID            string
		AbbreviatedOID string
		TargetType     string `json:"type"`
	}

	type GitRef struct {
		Name        string
		AbbrevName  string
		DisplayName string
		Prefix      string
		RefType     string `json:"type"`
		Repository  struct{ ID string }
		URL         string
		Target      GitTarget
	}

	type Changeset struct {
		ID          string
		Repository  struct{ ID string }
		Campaigns   CampaignConnection
		CreatedAt   string
		UpdatedAt   string
		Title       string
		Body        string
		State       string
		ExternalURL struct {
			URL         string
			ServiceType string
		}
		ReviewState string
		CheckState  string
		Events      ChangesetEventConnection
		Head        GitRef
		Base        GitRef
	}

	var result struct {
		Changesets []Changeset
	}

	graphqlGithubRepoID := string(graphqlbackend.MarshalRepositoryID(api.RepoID(githubRepo.ID)))
	graphqlBBSRepoID := string(graphqlbackend.MarshalRepositoryID(api.RepoID(bbsRepo.ID)))

	in := fmt.Sprintf(
		`[{repository: %q, externalID: %q}, {repository: %q, externalID: %q}]`,
		graphqlGithubRepoID, "999",
		graphqlBBSRepoID, "2",
	)

	mustExec(ctx, t, s, nil, &result, fmt.Sprintf(`
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
		want := []Changeset{
			{
				Repository: struct{ ID string }{ID: graphqlGithubRepoID},
				CreatedAt:  now.Format(time.RFC3339),
				UpdatedAt:  now.Format(time.RFC3339),
				Title:      "add extension filter to filter bar",
				Body:       "Enables adding extension filters to the filter bar by rendering the extension filter as filter chips inside the filter bar.\r\nWIP for https://github.com/sourcegraph/sourcegraph/issues/962\r\n\r\n> This PR updates the CHANGELOG.md file to describe any user-facing changes.\r\n.\r\n",
				State:      "MERGED",
				ExternalURL: struct{ URL, ServiceType string }{
					URL:         "https://github.com/sourcegraph/sourcegraph/pull/999",
					ServiceType: "github",
				},
				ReviewState: "APPROVED",
				CheckState:  "PASSED",
				Events: ChangesetEventConnection{
					TotalCount: 57,
				},
				Head: GitRef{
					Name:        "refs/heads/vo/add-type-issue-filter",
					AbbrevName:  "vo/add-type-issue-filter",
					DisplayName: "vo/add-type-issue-filter",
					Prefix:      "refs/heads/",
					RefType:     "GIT_BRANCH",
					Repository:  struct{ ID string }{ID: "UmVwb3NpdG9yeTox"},
					URL:         "/github.com/sourcegraph/sourcegraph@vo/add-type-issue-filter",

					Target: GitTarget{
						OID:            "7db302f07955e41d50e656d5faebefb4d87bce8a",
						AbbreviatedOID: "7db302f",
						TargetType:     "GIT_COMMIT",
					},
				},
				Base: GitRef{
					Name:        "refs/heads/master",
					AbbrevName:  "master",
					DisplayName: "master",
					Prefix:      "refs/heads/",
					RefType:     "GIT_BRANCH",
					Repository:  struct{ ID string }{ID: "UmVwb3NpdG9yeTox"},
					URL:         "/github.com/sourcegraph/sourcegraph@master",
					Target: GitTarget{
						OID:            "fa3815ba9ddd49db9111c5e9691e16d27e8f1f60",
						AbbreviatedOID: "fa3815b",
						TargetType:     "GIT_COMMIT",
					},
				},
			},
			{
				Repository: struct{ ID string }{ID: graphqlBBSRepoID},
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
				Events: ChangesetEventConnection{
					TotalCount: 9,
				},
				Head: GitRef{
					Name:        "refs/heads/release-testing-pr",
					AbbrevName:  "release-testing-pr",
					DisplayName: "release-testing-pr",
					Prefix:      "refs/heads/",
					RefType:     "GIT_BRANCH",
					Repository:  struct{ ID string }{ID: "UmVwb3NpdG9yeToy"},
					URL:         "/bitbucket.sgdev.org/SOUR/vegeta@release-testing-pr",
					Target: GitTarget{
						OID:            "mockcommitid",
						AbbreviatedOID: "mockcom",
						TargetType:     "GIT_COMMIT",
					},
				},
				Base: GitRef{
					Name:        "refs/heads/master",
					AbbrevName:  "master",
					DisplayName: "master",
					Prefix:      "refs/heads/",
					RefType:     "GIT_BRANCH",
					Repository:  struct{ ID string }{ID: "UmVwb3NpdG9yeToy"},
					URL:         "/bitbucket.sgdev.org/SOUR/vegeta@master",
					Target: GitTarget{
						OID:            "mockcommitid",
						AbbreviatedOID: "mockcom",
						TargetType:     "GIT_COMMIT",
					},
				},
			},
		}

		have := make([]Changeset, 0, len(result.Changesets))
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
	}

	type ChangesetConnection struct {
		Nodes      []Changeset
		TotalCount int
		PageInfo   struct {
			HasNextPage bool
		}
	}

	type ChangesetCounts struct {
		Date                 graphqlbackend.DateTime
		Total                int32
		Merged               int32
		Closed               int32
		Open                 int32
		OpenApproved         int32
		OpenChangesRequested int32
		OpenPending          int32
	}

	type CampaignWithChangesets struct {
		ID                      string
		Name                    string
		Description             string
		Author                  User
		CreatedAt               string
		UpdatedAt               string
		Namespace               UserOrg
		Changesets              ChangesetConnection
		ChangesetCountsOverTime []ChangesetCounts
	}

	var addChangesetsResult struct{ Campaign CampaignWithChangesets }

	changesetIDs := make([]string, 0, len(result.Changesets))
	for _, c := range result.Changesets {
		changesetIDs = append(changesetIDs, c.ID)
	}

	// Date when PR #999 from above was created
	countsFrom := parseJSONTime(t, "2018-11-14T22:07:45Z")
	// Date when PR #999 from above was merged
	countsTo := parseJSONTime(t, "2018-12-04T08:10:07Z")

	mustExec(ctx, t, s, nil, &addChangesetsResult, fmt.Sprintf(`
		fragment u on User { id, databaseID, siteAdmin }
		fragment o on Org  { id, name }

		fragment cs on ExternalChangeset {
			id
			repository { id }
			createdAt
			updatedAt
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
				nodes { ...cs }
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

	deleteInput := map[string]interface{}{"id": campaigns.Admin.ID}
	mustExec(ctx, t, s, deleteInput, &struct{}{}, `
		mutation($id: ID!){
			deleteCampaign(campaign: $id) { alwaysNil }
		}
	`)

	var campaignsAfterDelete struct {
		Campaigns struct {
			TotalCount int
		}
	}

	mustExec(ctx, t, s, nil, &campaignsAfterDelete, `
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

	cf, save := newGithubClientFactory(t, "test-changeset-counts-over-time")
	defer save()

	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := func() time.Time {
		return now.UTC().Truncate(time.Microsecond)
	}

	u, err := db.Users.Create(ctx, db.NewUser{
		Email:                 "thorsten@sourcegraph.com",
		Username:              "thorsten",
		DisplayName:           "thorsten",
		Password:              "1234",
		EmailVerificationCode: "foobar",
	})
	if err != nil {
		t.Fatal(err)
	}

	repoStore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})
	githubExtSvc := &repos.ExternalService{
		Kind:        "GITHUB",
		DisplayName: "GitHub",
		Config: marshalJSON(t, &schema.GitHubConnection{
			Url:   "https://github.com",
			Token: os.Getenv("GITHUB_TOKEN"),
			Repos: []string{"sourcegraph/sourcegraph"},
		}),
	}

	err = repoStore.UpsertExternalServices(ctx, githubExtSvc)
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

	store := ee.NewStoreWithClock(dbconn.Global, clock)

	campaign := &campaigns.Campaign{
		Name:            "Test campaign",
		Description:     "Testing changeset counts",
		AuthorID:        u.ID,
		NamespaceUserID: u.ID,
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

	syncer := ee.ChangesetSyncer{
		ReposStore:  repoStore,
		Store:       store,
		HTTPFactory: cf,
	}
	err = syncer.SyncChangesets(ctx, changesets...)
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

const testDiff = `diff README.md README.md
index 671e50a..851b23a 100644
--- README.md
+++ README.md
@@ -1,2 +1,2 @@
 # README
-This file is hosted at example.com and is a test file.
+This file is hosted at sourcegraph.com and is a test file.
diff --git urls.txt urls.txt
index 6f8b5d9..17400bc 100644
--- urls.txt
+++ urls.txt
@@ -1,3 +1,3 @@
 another-url.com
-example.com
+sourcegraph.com
 never-touch-the-mouse.com
`

// wantFileDiffs is the parsed representation of testDiff.
var wantFileDiffs = FileDiffs{
	RawDiff:  testDiff,
	DiffStat: DiffStat{Changed: 2},
	Nodes: []FileDiff{
		{
			OldPath: "README.md",
			NewPath: "README.md",
			OldFile: File{Name: "README.md"},
			Hunks: []FileDiffHunk{
				{
					Body:     " # README\n-This file is hosted at example.com and is a test file.\n+This file is hosted at sourcegraph.com and is a test file.\n",
					OldRange: DiffRange{StartLine: 1, Lines: 2},
					NewRange: DiffRange{StartLine: 1, Lines: 2},
				},
			},
			Stat: DiffStat{Changed: 1},
		},
		{
			OldPath: "urls.txt",
			NewPath: "urls.txt",
			OldFile: File{Name: "urls.txt"},
			Hunks: []FileDiffHunk{
				{
					Body:     " another-url.com\n-example.com\n+sourcegraph.com\n never-touch-the-mouse.com\n",
					OldRange: DiffRange{StartLine: 1, Lines: 3},
					NewRange: DiffRange{StartLine: 1, Lines: 3},
				},
			},
			Stat: DiffStat{Changed: 1},
		},
	},
}

type DiffRange struct{ StartLine, Lines int }

type FileDiffHunk struct {
	Body, Section      string
	OldNoNewlineAt     bool
	OldRange, NewRange DiffRange
}

type DiffStat struct{ Added, Deleted, Changed int }

type File struct {
	Name string
	// Ignoring other fields of File2, since that would require gitserver
}

type FileDiff struct {
	OldPath, NewPath string
	Hunks            []FileDiffHunk
	Stat             DiffStat
	OldFile          File
}

type FileDiffs struct {
	RawDiff  string
	DiffStat DiffStat
	Nodes    []FileDiff
}

type ChangesetPlan struct {
	Repository struct{ Name, URL string }
	Diff       struct {
		FileDiffs FileDiffs
	}
}

type CampaignPlan struct {
	ID         string
	Changesets struct {
		Nodes []ChangesetPlan
	}
	PreviewURL string
}

func TestCreateCampaignPlanFromPatchesResolver(t *testing.T) {
	ctx := backend.WithAuthzBypass(context.Background())

	dbtesting.SetupGlobalTestDB(t)

	user := createTestUser(ctx, t)
	act := actor.FromUser(user.ID)
	ctx = actor.WithActor(ctx, act)

	t.Run("invalid patch", func(t *testing.T) {
		args := graphqlbackend.CreateCampaignPlanFromPatchesArgs{
			Patches: []graphqlbackend.CampaignPlanPatch{
				{
					Repository:   graphqlbackend.MarshalRepositoryID(1),
					BaseRevision: "f00b4r",
					BaseRef:      "master",
					Patch:        "!!! this is not a valid unified diff !!!\n--- x\n+++ y\n@@ 1,1 2,2\na",
				},
			},
		}

		_, err := (&Resolver{}).CreateCampaignPlanFromPatches(ctx, args)
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

		backend.Mocks.Repos.ResolveRev = func(_ context.Context, _ *types.Repo, _ string) (api.CommitID, error) {
			return testingRev, nil
		}
		defer func() { backend.Mocks.Repos.ResolveRev = nil }()

		backend.Mocks.Repos.GetCommit = func(_ context.Context, _ *types.Repo, _ api.CommitID) (*git.Commit, error) {
			return &git.Commit{ID: testingRev}, nil
		}
		defer func() { backend.Mocks.Repos.GetCommit = nil }()

		reposStore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})
		repo := &repos.Repo{
			Name: "github.com/sourcegraph/sourcegraph",
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "external-id",
				ServiceType: "github",
				ServiceID:   "https://github.com/",
			},
			Sources: map[string]*repos.SourceInfo{
				"extsvc:github:4": {
					ID:       "extsvc:github:4",
					CloneURL: "https://secrettoken@github.com/sourcegraph/sourcegraph",
				},
			},
		}
		if err := reposStore.UpsertRepos(ctx, repo); err != nil {
			t.Fatal(err)
		}

		store := ee.NewStoreWithClock(dbconn.Global, clock)

		sr := &Resolver{store: store}
		s, err := graphqlbackend.NewSchema(sr, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		var response struct{ CreateCampaignPlanFromPatches CampaignPlan }

		mustExec(ctx, t, s, nil, &response, fmt.Sprintf(`
      mutation {
		createCampaignPlanFromPatches(patches: [{repository: %q, baseRevision: "f00b4r", baseRef: "master", patch: %q}]) {
          ... on CampaignPlan {
            id
            changesets(first: %d) {
              nodes {
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
            previewURL
          }
        }
      }
	`, graphqlbackend.MarshalRepositoryID(api.RepoID(repo.ID)), testDiff, 1))

		result := response.CreateCampaignPlanFromPatches

		wantChangesets := []ChangesetPlan{
			{
				Repository: struct{ Name, URL string }{Name: repo.Name},
				Diff:       struct{ FileDiffs FileDiffs }{FileDiffs: wantFileDiffs},
			},
		}
		if !cmp.Equal(result.Changesets.Nodes, wantChangesets) {
			t.Error("wrong changesets", cmp.Diff(result.Changesets.Nodes, wantChangesets))
		}

		if have, want := result.PreviewURL, "http://example.com/campaigns/new?plan=Q2FtcGFpZ25QbGFuOjE%3D"; have != want {
			t.Fatalf("have PreviewURL %q, want %q", have, want)
		}
	})
}

func TestCampaignPlanResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)
	rcache.SetupForTest(t)

	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := func() time.Time {
		return now.UTC().Truncate(time.Microsecond)
	}

	// For testing purposes they all share the same rev, across repos
	testingRev := api.CommitID("24f7ca7c1190835519e261d7eefa09df55ceea4f")

	backend.Mocks.Repos.ResolveRev = func(_ context.Context, _ *types.Repo, _ string) (api.CommitID, error) {
		return testingRev, nil
	}
	defer func() { backend.Mocks.Repos.ResolveRev = nil }()

	backend.Mocks.Repos.GetCommit = func(_ context.Context, _ *types.Repo, _ api.CommitID) (*git.Commit, error) {
		return &git.Commit{ID: testingRev}, nil
	}
	defer func() { backend.Mocks.Repos.GetCommit = nil }()

	repoupdater.MockRepoLookup = func(args protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
		return &protocol.RepoLookupResult{
			Repo: &protocol.RepoInfo{Name: args.Repo},
		}, nil
	}
	defer func() { repoupdater.MockRepoLookup = nil }()

	reposStore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})

	var rs []*repos.Repo
	for i := 0; i < 3; i++ {
		repo := &repos.Repo{
			Name:        fmt.Sprintf("github.com/sourcegraph/sourcegraph-%d", i),
			URI:         fmt.Sprintf("github.com/sourcegraph/sourcegraph-%d", i),
			Description: "Code search and navigation tool",
			ExternalRepo: api.ExternalRepoSpec{
				ID:          fmt.Sprintf("external-id-%d", i),
				ServiceType: "github",
				ServiceID:   "https://github.com/",
			},
			Sources: map[string]*repos.SourceInfo{
				"extsvc:github:4": {
					ID:       "extsvc:github:4",
					CloneURL: "https://secrettoken@github.com/sourcegraph/sourcegraph",
				},
			},
		}
		err := reposStore.UpsertRepos(ctx, repo)
		if err != nil {
			t.Fatal(err)
		}
		rs = append(rs, repo)
	}

	store := ee.NewStoreWithClock(dbconn.Global, clock)

	user := createTestUser(ctx, t)
	plan := &campaigns.CampaignPlan{UserID: user.ID}
	err := store.CreateCampaignPlan(ctx, plan)
	if err != nil {
		t.Fatal(err)
	}

	var jobs []*campaigns.CampaignJob
	for _, repo := range rs {
		job := &campaigns.CampaignJob{
			CampaignPlanID: plan.ID,
			RepoID:         repo.ID,
			Rev:            testingRev,
			BaseRef:        "master",
			Diff:           testDiff,
		}

		err := store.CreateCampaignJob(ctx, job)
		if err != nil {
			t.Fatal(err)
		}
		jobs = append(jobs, job)
	}

	type Response struct {
		Node CampaignPlan
	}

	sr := &Resolver{store: store}
	s, err := graphqlbackend.NewSchema(sr, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	var response Response

	mustExec(ctx, t, s, nil, &response, fmt.Sprintf(`
      query {
        node(id: %q) {
          ... on CampaignPlan {
            id
            changesets(first: %d) {
              nodes {
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
        }
      }
	`, marshalCampaignPlanID(plan.ID), len(jobs)))

	if have, want := len(response.Node.Changesets.Nodes), len(jobs); have != want {
		t.Fatalf("have %d changeset plans, want %d", have, want)
	}

	for i, changesetPlan := range response.Node.Changesets.Nodes {
		if have, want := changesetPlan.Repository.Name, rs[i].Name; have != want {
			t.Fatalf("wrong Repository Name %q. want=%q", have, want)
		}

		if have, want := changesetPlan.Diff.FileDiffs.RawDiff, testDiff; have != want {
			t.Fatalf("wrong RawDiff. diff=%s", cmp.Diff(have, want))
		}

		if have, want := changesetPlan.Diff.FileDiffs.DiffStat.Changed, 2; have != want {
			t.Fatalf("wrong DiffStat.Changed %d, want=%d", have, want)
		}

		haveFileDiffs := changesetPlan.Diff.FileDiffs
		if !reflect.DeepEqual(haveFileDiffs, wantFileDiffs) {
			t.Fatal(cmp.Diff(haveFileDiffs, wantFileDiffs))
		}
	}
}

func mustExec(
	ctx context.Context,
	t testing.TB,
	s *graphql.Schema,
	in map[string]interface{},
	out interface{},
	query string,
) {
	t.Helper()
	if errs := exec(ctx, t, s, in, out, query); len(errs) > 0 {
		t.Fatalf("unexpected graphql query errors: %v", errs)
	}
}

func exec(
	ctx context.Context,
	t testing.TB,
	s *graphql.Schema,
	in map[string]interface{},
	out interface{},
	query string,
) []*gqlerrors.QueryError {
	t.Helper()

	query = strings.Replace(query, "\t", "  ", -1)

	r := s.Exec(ctx, query, "", in)
	if len(r.Errors) != 0 {
		return r.Errors
	}

	if testing.Verbose() {
		t.Logf("\n---- GraphQL Query ----\n%s\n\nVars: %s\n---- GraphQL Result ----\n%s\n -----------", query, toJSON(t, in), r.Data)
	}

	if err := json.Unmarshal(r.Data, out); err != nil {
		t.Fatalf("failed to unmarshal graphql data: %v", err)
	}

	return nil
}

func toJSON(t testing.TB, v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	formatted, err := jsonc.Format(string(data), nil)
	if err != nil {
		t.Fatal(err)
	}

	return formatted
}

func newGithubClientFactory(t testing.TB, name string) (*httpcli.Factory, func()) {
	t.Helper()

	cassete := filepath.Join("testdata/vcr/", strings.Replace(name, " ", "-", -1))

	rec, err := httptestutil.NewRecorder(cassete, *update, func(i *cassette.Interaction) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	mw := httpcli.NewMiddleware(githubProxyRedirectMiddleware)

	hc := httpcli.NewFactory(mw, httptestutil.NewRecorderOpt(rec))

	return hc, func() {
		if err := rec.Stop(); err != nil {
			t.Errorf("failed to update test data: %s", err)
		}
	}
}

func githubProxyRedirectMiddleware(cli httpcli.Doer) httpcli.Doer {
	return httpcli.DoerFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Hostname() == "github-proxy" {
			req.URL.Host = "api.github.com"
			req.URL.Scheme = "https"
		}
		return cli.Do(req)
	})
}

func marshalJSON(t testing.TB, v interface{}) string {
	t.Helper()

	bs, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	return string(bs)
}

func marshalDateTime(t testing.TB, ts time.Time) string {
	t.Helper()

	dt := graphqlbackend.DateTime{Time: ts}

	bs, err := dt.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	return string(bs)
}

func parseJSONTime(t testing.TB, ts string) time.Time {
	t.Helper()

	timestamp, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		t.Fatal(err)
	}

	return timestamp
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

var testUser = db.NewUser{
	Email:                "test@sourcegraph.com",
	Username:             "test",
	DisplayName:          "Test",
	Password:             "test",
	EmailIsVerified:      true,
	FailIfNotInitialUser: false,
}

func createTestUser(ctx context.Context, t *testing.T) *types.User {
	t.Helper()
	user, err := db.Users.Create(ctx, testUser)
	if err != nil {
		t.Fatal(err)
	}
	return user
}
