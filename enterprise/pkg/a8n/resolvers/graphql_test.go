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
	"github.com/graph-gophers/graphql-go/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ee "github.com/sourcegraph/sourcegraph/enterprise/pkg/a8n"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/schema"
)

func init() {
	dbtesting.DBNameSuffix = "a8nresolversdb"
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

	s, err := graphqlbackend.NewSchema(sr)
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
		Events      ChangesetEventConnection
	}

	var result struct {
		Changesets []Changeset
	}

	graphqlGithubRepoID := string(marshalRepositoryID(api.RepoID(githubRepo.ID)))
	graphqlBBSRepoID := string(marshalRepositoryID(api.RepoID(bbsRepo.ID)))

	in := fmt.Sprintf(
		`[{repository: %q, externalID: %q}, {repository: %q, externalID: %q}]`,
		graphqlGithubRepoID, "999",
		graphqlBBSRepoID, "3",
	)

	mustExec(ctx, t, s, nil, &result, fmt.Sprintf(`
		fragment cs on Changeset {
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
			events(first: 100) {
				totalCount
			}
		}
		mutation() {
			changesets: createChangesets(input: %s) {
				...cs
			}
		}
	`, string(in)))

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
				Events: ChangesetEventConnection{
					TotalCount: 26,
				},
			},
			{
				Repository: struct{ ID string }{ID: graphqlBBSRepoID},
				CreatedAt:  now.Format(time.RFC3339),
				UpdatedAt:  now.Format(time.RFC3339),
				Title:      "attacker: Simplify timeout configuration",
				Body:       "This commit simplifies the timeout configuration internally to make use\r\nof the `http.Client.Timeout` field which didn't exist when this code was\r\noriginally written.\r\n\r\nIt also fixes #396: a test flake of `TestTimeout`.",
				State:      "OPEN",
				ExternalURL: struct{ URL, ServiceType string }{
					URL:         "https://bitbucket.sgdev.org/projects/SOUR/repos/vegeta/pull-requests/3",
					ServiceType: "bitbucketServer",
				},
				ReviewState: "CHANGES_REQUESTED",
				Events: ChangesetEventConnection{
					TotalCount: 0,
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

		if !reflect.DeepEqual(have, want) {
			t.Fatal(cmp.Diff(have, want))
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

		fragment cs on Changeset {
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
) []*errors.QueryError {
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

	dt := graphqlbackend.DateTime{ts}

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
