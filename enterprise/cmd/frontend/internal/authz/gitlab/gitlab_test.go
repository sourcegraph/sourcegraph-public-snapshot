package gitlab

import (
	"context"
	"encoding/json"
	"net/url"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/schema"
	"golang.org/x/oauth2"
)

func Test_GitLab_RepoPerms(t *testing.T) {
	gitlabMock := newMockGitLab(t,
		[]int{ // Repos
			11, // gitlab.mine/bl/repo-1
			12, // gitlab.mine/bl/repo-2
			13, // gitlab.mine/bl/repo-3
			21, // gitlab.mine/kl/repo-1
			22, // gitlab.mine/kl/repo-2
			23, // gitlab.mine/kl/repo-3
			31, // gitlab.mine/org/repo-1
			32, // gitlab.mine/org/repo-2
			33, // gitlab.mine/org/repo-3
			41, // gitlab.mine/public/repo-1
		},
		map[string][]int{ // GitLab user IDs to repo IDs
			"101":    {11, 12, 13, 31, 32, 33, 41},
			"201":    {21, 22, 23, 31, 32, 33, 41},
			"PUBLIC": {41},
		},
		map[string]string{ // GitLab OAuth tokens to GitLab user IDs
			"oauth101": "101",
			"oauth201": "201",
		},
		1)
	gitlab.MockListProjects = gitlabMock.ListProjects

	tests := []GitLab_RepoPerms_Test{
		{
			description: "standard config",
			op: GitLabOAuthAuthzProviderOp{
				BaseURL: mustURL(t, "https://gitlab.mine"),
			},
			calls: []GitLab_RepoPerms_call{
				{
					description: "bl user has expected perms",
					account:     acct(1, "gitlab", "https://gitlab.mine/", "101", "oauth101"),
					repos: map[authz.Repo]struct{}{
						repo("bl/repo-1", gitlab.ServiceType, "https://gitlab.mine/", "11"):                 {},
						repo("bl/repo-2", gitlab.ServiceType, "other", "12"):                                {},
						repo("gitlab.mine/bl/repo-3", gitlab.ServiceType, "https://gitlab.mine/", "999"):    {},
						repo("kl/repo-1", gitlab.ServiceType, "https://gitlab.mine/", "21"):                 {},
						repo("kl/repo-2", gitlab.ServiceType, "other", "22"):                                {},
						repo("gitlab.mine/kl/repo-3", gitlab.ServiceType, "https://gitlab.mine/", "998"):    {},
						repo("org/repo-1", gitlab.ServiceType, "https://gitlab.mine/", "31"):                {},
						repo("org/repo-2", gitlab.ServiceType, "other", "32"):                               {},
						repo("gitlab.mine/org/repo-3", gitlab.ServiceType, "https://gitlab.mine/", "997"):   {},
						repo("gitlab.mine/public/repo-1", gitlab.ServiceType, "https://gitlab.mine/", "41"): {},
					},
					expPerms: map[api.RepoName]map[authz.Perm]bool{
						"bl/repo-1":                 {authz.Read: true},
						"gitlab.mine/bl/repo-3":     {},
						"kl/repo-1":                 {},
						"gitlab.mine/kl/repo-3":     {},
						"org/repo-1":                {authz.Read: true},
						"gitlab.mine/org/repo-3":    {},
						"gitlab.mine/public/repo-1": {authz.Read: true},
					},
				},
				{
					description: "kl user has expected perms",
					account:     acct(2, "gitlab", "https://gitlab.mine/", "201", "oauth201"),
					repos: map[authz.Repo]struct{}{
						repo("bl/repo-1", gitlab.ServiceType, "https://gitlab.mine/", "11"):                 {},
						repo("bl/repo-2", gitlab.ServiceType, "other", "12"):                                {},
						repo("gitlab.mine/bl/repo-3", gitlab.ServiceType, "https://gitlab.mine/", "999"):    {},
						repo("kl/repo-1", gitlab.ServiceType, "https://gitlab.mine/", "21"):                 {},
						repo("kl/repo-2", gitlab.ServiceType, "other", "22"):                                {},
						repo("gitlab.mine/kl/repo-3", gitlab.ServiceType, "https://gitlab.mine/", "998"):    {},
						repo("org/repo-1", gitlab.ServiceType, "https://gitlab.mine/", "31"):                {},
						repo("org/repo-2", gitlab.ServiceType, "other", "32"):                               {},
						repo("gitlab.mine/org/repo-3", gitlab.ServiceType, "https://gitlab.mine/", "997"):   {},
						repo("gitlab.mine/public/repo-1", gitlab.ServiceType, "https://gitlab.mine/", "41"): {},
					},
					expPerms: map[api.RepoName]map[authz.Perm]bool{
						"bl/repo-1":                 {},
						"gitlab.mine/bl/repo-3":     {},
						"kl/repo-1":                 {authz.Read: true},
						"gitlab.mine/kl/repo-3":     {},
						"org/repo-1":                {authz.Read: true},
						"gitlab.mine/org/repo-3":    {},
						"gitlab.mine/public/repo-1": {authz.Read: true},
					},
				},
				{
					description: "unknown user has access to public only",
					account:     acct(3, "gitlab", "https://gitlab.mine/", "999", "oauth999"),
					repos: map[authz.Repo]struct{}{
						repo("bl/repo-1", gitlab.ServiceType, "https://gitlab.mine/", "11"):                 {},
						repo("bl/repo-2", gitlab.ServiceType, "other", "12"):                                {},
						repo("gitlab.mine/bl/repo-3", gitlab.ServiceType, "https://gitlab.mine/", "999"):    {},
						repo("kl/repo-1", gitlab.ServiceType, "https://gitlab.mine/", "21"):                 {},
						repo("kl/repo-2", gitlab.ServiceType, "other", "22"):                                {},
						repo("gitlab.mine/kl/repo-3", gitlab.ServiceType, "https://gitlab.mine/", "998"):    {},
						repo("org/repo-1", gitlab.ServiceType, "https://gitlab.mine/", "31"):                {},
						repo("org/repo-2", gitlab.ServiceType, "other", "32"):                               {},
						repo("gitlab.mine/org/repo-3", gitlab.ServiceType, "https://gitlab.mine/", "997"):   {},
						repo("gitlab.mine/public/repo-1", gitlab.ServiceType, "https://gitlab.mine/", "41"): {},
					},
					expPerms: map[api.RepoName]map[authz.Perm]bool{
						"bl/repo-1":                 {},
						"gitlab.mine/bl/repo-3":     {},
						"kl/repo-1":                 {},
						"gitlab.mine/kl/repo-3":     {},
						"org/repo-1":                {},
						"gitlab.mine/org/repo-3":    {},
						"gitlab.mine/public/repo-1": {authz.Read: true},
					},
				},
				{
					description: "unauthenticated user has access to public only",
					account:     nil,
					repos: map[authz.Repo]struct{}{
						repo("bl/repo-1", gitlab.ServiceType, "https://gitlab.mine/", "11"):                 {},
						repo("bl/repo-2", gitlab.ServiceType, "other", "12"):                                {},
						repo("gitlab.mine/bl/repo-3", gitlab.ServiceType, "https://gitlab.mine/", "999"):    {},
						repo("kl/repo-1", gitlab.ServiceType, "https://gitlab.mine/", "21"):                 {},
						repo("kl/repo-2", gitlab.ServiceType, "other", "22"):                                {},
						repo("gitlab.mine/kl/repo-3", gitlab.ServiceType, "https://gitlab.mine/", "998"):    {},
						repo("org/repo-1", gitlab.ServiceType, "https://gitlab.mine/", "31"):                {},
						repo("org/repo-2", gitlab.ServiceType, "other", "32"):                               {},
						repo("gitlab.mine/org/repo-3", gitlab.ServiceType, "https://gitlab.mine/", "997"):   {},
						repo("gitlab.mine/public/repo-1", gitlab.ServiceType, "https://gitlab.mine/", "41"): {},
					},
					expPerms: map[api.RepoName]map[authz.Perm]bool{
						"bl/repo-1":                 {},
						"gitlab.mine/bl/repo-3":     {},
						"kl/repo-1":                 {},
						"gitlab.mine/kl/repo-3":     {},
						"org/repo-1":                {},
						"gitlab.mine/org/repo-3":    {},
						"gitlab.mine/public/repo-1": {authz.Read: true},
					},
				},
				{
					description: "user with no oauth token has access to public only",
					account:     acct(2, "gitlab", "https://gitlab.mine/", "201", ""),
					repos: map[authz.Repo]struct{}{
						repo("bl/repo-1", gitlab.ServiceType, "https://gitlab.mine/", "11"):                 {},
						repo("bl/repo-2", gitlab.ServiceType, "other", "12"):                                {},
						repo("gitlab.mine/bl/repo-3", gitlab.ServiceType, "https://gitlab.mine/", "999"):    {},
						repo("kl/repo-1", gitlab.ServiceType, "https://gitlab.mine/", "21"):                 {},
						repo("kl/repo-2", gitlab.ServiceType, "other", "22"):                                {},
						repo("gitlab.mine/kl/repo-3", gitlab.ServiceType, "https://gitlab.mine/", "998"):    {},
						repo("org/repo-1", gitlab.ServiceType, "https://gitlab.mine/", "31"):                {},
						repo("org/repo-2", gitlab.ServiceType, "other", "32"):                               {},
						repo("gitlab.mine/org/repo-3", gitlab.ServiceType, "https://gitlab.mine/", "997"):   {},
						repo("gitlab.mine/public/repo-1", gitlab.ServiceType, "https://gitlab.mine/", "41"): {},
					},
					expPerms: map[api.RepoName]map[authz.Perm]bool{
						"bl/repo-1":                 {},
						"gitlab.mine/bl/repo-3":     {},
						"kl/repo-1":                 {},
						"gitlab.mine/kl/repo-3":     {},
						"org/repo-1":                {},
						"gitlab.mine/org/repo-3":    {},
						"gitlab.mine/public/repo-1": {authz.Read: true},
					},
				},
			},
		},
	}
	for _, test := range tests {
		test.run(t)
	}
}

type GitLab_RepoPerms_Test struct {
	description string

	op GitLabOAuthAuthzProviderOp

	calls []GitLab_RepoPerms_call
}

type GitLab_RepoPerms_call struct {
	description string
	account     *extsvc.ExternalAccount
	repos       map[authz.Repo]struct{}
	expPerms    map[api.RepoName]map[authz.Perm]bool
}

func (g GitLab_RepoPerms_Test) run(t *testing.T) {
	t.Logf("Test case %q", g.description)

	for _, c := range g.calls {
		t.Logf("Call %q", c.description)

		// Recreate the authz provider cache every time, before running twice (once uncached, once cached)
		ctx := context.Background()
		op := g.op
		op.MockCache = make(mockCache)
		authzProvider := NewProvider(op)

		for i := 0; i < 2; i++ {
			t.Logf("iter %d", i)
			perms, err := authzProvider.RepoPerms(ctx, c.account, c.repos)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				continue
			}
			if !reflect.DeepEqual(perms, c.expPerms) {
				t.Errorf("expected %s, but got %s", asJSON(t, c.expPerms), asJSON(t, perms))
			}
		}
	}
}

func Test_GitLab_RepoPerms_cache(t *testing.T) {
	gitlabMock := newMockGitLab(t, []int{}, map[string][]int{}, map[string]string{}, 1)
	gitlab.MockListProjects = gitlabMock.ListProjects

	ctx := context.Background()
	authzProvider := NewProvider(GitLabOAuthAuthzProviderOp{
		BaseURL:   mustURL(t, "https://gitlab.mine"),
		MockCache: make(mockCache),
		CacheTTL:  3 * time.Hour,
	})
	if _, err := authzProvider.RepoPerms(ctx, acct(1, gitlab.ServiceType, "https://gitlab.mine/", "bl", "oauth_bl"), nil); err != nil {
		t.Fatal(err)
	}
	if exp := map[string]map[string]int{
		"projects?per_page=100": {"oauth_bl": 1},
	}; !reflect.DeepEqual(gitlabMock.madeProjectReqs, exp) {
		t.Errorf("Unexpected cache behavior. Expected underying requests to be %v, but got %v", exp, gitlabMock.madeProjectReqs)
	}

	if _, err := authzProvider.RepoPerms(ctx, acct(1, gitlab.ServiceType, "https://gitlab.mine/", "bl", "oauth_bl"), nil); err != nil {
		t.Fatal(err)
	}
	if exp := map[string]map[string]int{
		"projects?per_page=100": {"oauth_bl": 1},
	}; !reflect.DeepEqual(gitlabMock.madeProjectReqs, exp) {
		t.Errorf("Unexpected cache behavior. Expected underying requests to be %v, but got %v", exp, gitlabMock.madeProjectReqs)
	}

	if _, err := authzProvider.RepoPerms(ctx, acct(1, gitlab.ServiceType, "https://gitlab.mine/", "kl", "oauth_kl"), nil); err != nil {
		t.Fatal(err)
	}
	if exp := map[string]map[string]int{
		"projects?per_page=100": {"oauth_bl": 1, "oauth_kl": 1},
	}; !reflect.DeepEqual(gitlabMock.madeProjectReqs, exp) {
		t.Errorf("Unexpected cache behavior. Expected underying requests to be %v, but got %v", exp, gitlabMock.madeProjectReqs)
	}

	if _, err := authzProvider.RepoPerms(ctx, acct(1, gitlab.ServiceType, "https://gitlab.mine/", "kl", "oauth_kl"), nil); err != nil {
		t.Fatal(err)
	}
	if exp := map[string]map[string]int{
		"projects?per_page=100": {"oauth_bl": 1, "oauth_kl": 1},
	}; !reflect.DeepEqual(gitlabMock.madeProjectReqs, exp) {
		t.Errorf("Unexpected cache behavior. Expected underying requests to be %v, but got %v", exp, gitlabMock.madeProjectReqs)
	}
}

// Test_GitLab_RepoPerms_cache_ttl tests the behavior of overwriting cache entries when the TTL changes
func Test_GitLab_RepoPerms_cache_ttl(t *testing.T) {
	gitlabMock := newMockGitLab(t,
		[]int{
			11, // gitlab.mine/bl/repo-1
		},
		map[string][]int{
			"101": {11},
		},
		map[string]string{
			"oauth101": "101",
		}, 1)
	gitlab.MockListProjects = gitlabMock.ListProjects

	cache := make(mockCache)
	ctx := context.Background()
	authzProvider := NewProvider(GitLabOAuthAuthzProviderOp{
		BaseURL:   mustURL(t, "https://gitlab.mine"),
		MockCache: cache,
	})
	if expCache := mockCache(map[string]string{}); !reflect.DeepEqual(cache, expCache) {
		t.Errorf("expected cache to be %+v, but was %+v", expCache, cache)
	}

	if _, err := authzProvider.RepoPerms(ctx, acct(1, gitlab.ServiceType, "https://gitlab.mine/", "101", "oauth101"), nil); err != nil {
		t.Fatal(err)
	}
	if expCache := mockCache(map[string]string{"101": `{"repos":{"11":{}},"ttl":0}`}); !reflect.DeepEqual(cache, expCache) {
		t.Errorf("expected cache to be %+v, but was %+v", expCache, cache)
	}

	if _, err := authzProvider.RepoPerms(ctx, acct(1, gitlab.ServiceType, "https://gitlab.mine/", "101", "oauth101"), nil); err != nil {
		t.Fatal(err)
	}
	if expCache := mockCache(map[string]string{"101": `{"repos":{"11":{}},"ttl":0}`}); !reflect.DeepEqual(cache, expCache) {
		t.Errorf("expected cache to be %+v, but was %+v", expCache, cache)
	}

	authzProvider.cacheTTL = time.Hour * 5

	if _, err := authzProvider.RepoPerms(ctx, acct(1, gitlab.ServiceType, "https://gitlab.mine/", "101", "oauth101"), nil); err != nil {
		t.Fatal(err)
	}
	if expCache := mockCache(map[string]string{"101": `{"repos":{"11":{}},"ttl":18000000000000}`}); !reflect.DeepEqual(cache, expCache) {
		t.Errorf("expected cache to be %+v, but was %+v", expCache, cache)
	}

	authzProvider.cacheTTL = time.Second * 5

	// Use lower TTL
	if _, err := authzProvider.RepoPerms(ctx, acct(1, gitlab.ServiceType, "https://gitlab.mine/", "101", "oauth101"), nil); err != nil {
		t.Fatal(err)
	}
	if expCache := mockCache(map[string]string{"101": `{"repos":{"11":{}},"ttl":5000000000}`}); !reflect.DeepEqual(cache, expCache) {
		t.Errorf("expected cache to be %+v, but was %+v", expCache, cache)
	}

	authzProvider.cacheTTL = time.Second * 60

	// Increase in TTL doesn't overwrite cache entry
	if _, err := authzProvider.RepoPerms(ctx, acct(1, gitlab.ServiceType, "https://gitlab.mine/", "101", "oauth101"), nil); err != nil {
		t.Fatal(err)
	}
	if expCache := mockCache(map[string]string{"101": `{"repos":{"11":{}},"ttl":5000000000}`}); !reflect.DeepEqual(cache, expCache) {
		t.Errorf("expected cache to be %+v, but was %+v", expCache, cache)
	}
}

func Test_GitLab_Repos(t *testing.T) {
	tests := []GitLab_Repos_Test{
		{
			description: "standard config",
			op: GitLabOAuthAuthzProviderOp{
				BaseURL: mustURL(t, "https://gitlab.mine"),
			},
			calls: []GitLab_Repos_call{
				{
					repos: map[authz.Repo]struct{}{
						repo("gitlab.mine/bl/repo-1", "", "", ""):                   {},
						repo("gitlab.mine/kl/repo-1", "", "", ""):                   {},
						repo("another.host/bl/repo-1", "", "", ""):                  {},
						repo("a", gitlab.ServiceType, "https://gitlab.mine/", "23"): {},
						repo("b", gitlab.ServiceType, "https://not-mine/", "34"):    {},
						repo("c", "not-gitlab", "https://gitlab.mine/", "45"):       {},
					},
					expMine: map[authz.Repo]struct{}{
						repo("a", gitlab.ServiceType, "https://gitlab.mine/", "23"): {},
					},
					expOthers: map[authz.Repo]struct{}{
						repo("gitlab.mine/bl/repo-1", "", "", ""):                {},
						repo("gitlab.mine/kl/repo-1", "", "", ""):                {},
						repo("another.host/bl/repo-1", "", "", ""):               {},
						repo("b", gitlab.ServiceType, "https://not-mine/", "34"): {},
						repo("c", "not-gitlab", "https://gitlab.mine/", "45"):    {},
					},
				},
			},
		},
	}
	for _, test := range tests {
		test.run(t)
	}
}

type GitLab_Repos_Test struct {
	description string
	op          GitLabOAuthAuthzProviderOp
	calls       []GitLab_Repos_call
}

type GitLab_Repos_call struct {
	repos     map[authz.Repo]struct{}
	expMine   map[authz.Repo]struct{}
	expOthers map[authz.Repo]struct{}
}

func (g GitLab_Repos_Test) run(t *testing.T) {
	t.Logf("Test case %q", g.description)
	for _, c := range g.calls {
		ctx := context.Background()
		op := g.op
		op.MockCache = make(mockCache)
		authzProvider := NewProvider(op)

		mine, others := authzProvider.Repos(ctx, c.repos)
		if !reflect.DeepEqual(mine, c.expMine) {
			t.Errorf("For input %v, expected mine to be %v, but got %v", c.repos, c.expMine, mine)
		}
		if !reflect.DeepEqual(others, c.expOthers) {
			t.Errorf("For input %v, expected others to be %v, but got %v", c.repos, c.expOthers, others)
		}
	}
}

// mockGitLab is a mock for the GitLab client that can be used by tests. Instantiating a mockGitLab
// instance itself does nothing, but its methods can be used to replace the mock functions (e.g.,
// MockListProjects).
//
// We prefer to do it this way, instead of defining an interface for the GitLab client, because this
// preserves the ability to jump-to-def around the actual implementation.
type mockGitLab struct {
	t *testing.T

	// acls is a map from GitLab user ID to list of accessible project IDs on GitLab
	// Publicly accessible projects are keyed by the special "PUBLIC" key
	acls map[string][]int

	// projs is a map of all projects on the instance, keyed by project ID
	projs map[int]*gitlab.Project

	// oauthToks is a map from OAuth token to GitLab user account ID
	oauthToks map[string]string

	// maxPerPage returns the max per_page value for the instance
	maxPerPage int

	// madeProjectReqs records how many ListProjects requests were made by url string and oauth
	// token
	madeProjectReqs map[string]map[string]int
}

// newMockGitLab returns a new mockGitLab instance with the specified projects (projIDs), ACLs (acls
// maps from user account ID to list of project IDs), and OAuth tokens (oauthToks maps OAuth token
// to user ID).
func newMockGitLab(t *testing.T, projIDs []int, acls map[string][]int, oauthToks map[string]string, maxPerPage int) mockGitLab {
	projs := make(map[int]*gitlab.Project)
	for _, p := range projIDs {
		projs[p] = &gitlab.Project{ProjectCommon: gitlab.ProjectCommon{ID: p}}
	}
	return mockGitLab{
		t:          t,
		projs:      projs,
		acls:       acls,
		oauthToks:  oauthToks,
		maxPerPage: maxPerPage,
	}
}

func (m *mockGitLab) ListProjects(c *gitlab.Client, ctx context.Context, urlStr string) (proj []*gitlab.Project, nextPageURL *string, err error) {
	if m.madeProjectReqs == nil {
		m.madeProjectReqs = make(map[string]map[string]int)
	}
	if m.madeProjectReqs[urlStr] == nil {
		m.madeProjectReqs[urlStr] = make(map[string]int)
	}
	m.madeProjectReqs[urlStr][c.OAuthToken]++

	u, err := url.Parse(urlStr)
	if err != nil {
		m.t.Fatalf("could not parse ListProjects urlStr %q: %s", urlStr, err)
	}
	acceptedQ := map[string]struct{}{"page": {}, "per_page": {}}
	for k := range u.Query() {
		if _, ok := acceptedQ[k]; !ok {
			m.t.Fatalf("mockGitLab unable to handle urlStr %q", urlStr)
		}
	}

	acctID := m.oauthToks[c.OAuthToken]
	var repoIDs []int
	if acctID == "" {
		repoIDs = m.acls["PUBLIC"]
	} else {
		repoIDs = m.acls[acctID]
	}

	allProjs := make([]*gitlab.Project, len(repoIDs))
	for i, repoID := range repoIDs {
		proj, ok := m.projs[repoID]
		if !ok {
			m.t.Fatalf("Dangling project reference in mockGitLab: %d", repoID)
		}
		allProjs[i] = proj
	}

	// pagination
	perPage, err := getIntOrDefault(u.Query().Get("per_page"), m.maxPerPage)
	if err != nil {
		return nil, nil, err
	}
	if perPage > m.maxPerPage {
		perPage = m.maxPerPage
	}
	page, err := getIntOrDefault(u.Query().Get("page"), 1)
	if err != nil {
		return nil, nil, err
	}
	p := page - 1
	var (
		pagedProjs []*gitlab.Project
	)
	if perPage*p > len(allProjs)-1 {
		pagedProjs = nil
	} else if perPage*(p+1) > len(allProjs)-1 {
		pagedProjs = allProjs[perPage*p:]
	} else {
		pagedProjs = allProjs[perPage*p : perPage*(p+1)]
		if perPage*(p+1) <= len(allProjs)-1 {
			newU := *u
			q := u.Query()
			q.Set("page", strconv.Itoa(page+1))
			newU.RawQuery = q.Encode()
			s := newU.String()
			nextPageURL = &s
		}
	}
	return pagedProjs, nextPageURL, nil
}

type mockCache map[string]string

func (m mockCache) Get(key string) ([]byte, bool) {
	v, ok := m[key]
	return []byte(v), ok
}
func (m mockCache) Set(key string, b []byte) {
	m[key] = string(b)
}
func (m mockCache) Delete(key string) {
	delete(m, key)
}

func getIntOrDefault(str string, def int) (int, error) {
	if str == "" {
		return def, nil
	}
	return strconv.Atoi(str)
}

func acct(userID int32, serviceType, serviceID, accountID, oauthTok string) *extsvc.ExternalAccount {
	var data extsvc.ExternalAccountData
	gitlab.SetExternalAccountData(&data, &gitlab.User{
		ID: userID,
	}, &oauth2.Token{
		AccessToken: oauthTok,
	})
	return &extsvc.ExternalAccount{
		UserID: userID,
		ExternalAccountSpec: extsvc.ExternalAccountSpec{
			ServiceType: serviceType,
			ServiceID:   serviceID,
			AccountID:   accountID,
		},
		ExternalAccountData: data,
	}
}

func repo(uri, serviceType, serviceID, id string) authz.Repo {
	return authz.Repo{
		RepoName: api.RepoName(uri),
		ExternalRepoSpec: api.ExternalRepoSpec{
			ID:          id,
			ServiceType: serviceType,
			ServiceID:   serviceID,
		},
	}
}

type mockAuthnProvider struct {
	configID  auth.ProviderConfigID
	serviceID string
}

func (m mockAuthnProvider) ConfigID() auth.ProviderConfigID {
	return m.configID
}

func (m mockAuthnProvider) Config() schema.AuthProviders {
	return schema.AuthProviders{
		Gitlab: &schema.GitLabAuthProvider{
			Type: m.configID.Type,
			Url:  m.configID.ID,
		},
	}
}

func (m mockAuthnProvider) CachedInfo() *auth.ProviderInfo {
	return &auth.ProviderInfo{ServiceID: m.serviceID}
}

func (m mockAuthnProvider) Refresh(ctx context.Context) error {
	panic("should not be called")
}

func mustURL(t *testing.T, u string) *url.URL {
	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

func asJSON(t *testing.T, v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
