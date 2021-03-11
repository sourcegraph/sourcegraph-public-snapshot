package gitlab

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"sort"
	"strconv"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func init() {
	spew.Config.DisablePointerAddresses = true
	spew.Config.SortKeys = true
	spew.Config.SpewKeys = true
}

// mockGitLab is a mock for the GitLab client that can be used by tests. Instantiating a mockGitLab
// instance itself does nothing, but its methods can be used to replace the mock functions (e.g.,
// MockListProjects).
//
// We prefer to do it this way, instead of defining an interface for the GitLab client, because this
// preserves the ability to jump-to-def around the actual implementation.
type mockGitLab struct {
	t *testing.T

	// projs is a map of all projects on the instance, keyed by project ID
	projs map[int]*gitlab.Project

	// users is a list of all users
	users []*gitlab.User

	// privateGuest is a map from GitLab user ID to list of metadata-accessible private project IDs on GitLab
	privateGuest map[int32][]int

	// privateRepo is a map from GitLab user ID to list of repo-content-accessible private project IDs on GitLab.
	// Projects in each list are also metadata-accessible.
	privateRepo map[int32][]int

	// oauthToks is a map from OAuth token to GitLab user account ID
	oauthToks map[string]int32

	// sudoTok is the sudo token, if there is one
	sudoTok string

	// madeGetProject records what GetProject calls have been made. It's a map from oauth token -> GetProjectOp -> count.
	madeGetProject map[string]map[gitlab.GetProjectOp]int

	// madeListProjects records what ListProjects calls have been made. It's a map from oauth token -> string (urlStr) -> count.
	madeListProjects map[string]map[string]int

	// madeListTree records what ListTree calls have been made. It's a map from oauth token -> ListTreeOp -> count.
	madeListTree map[string]map[gitlab.ListTreeOp]int

	// madeUsers records what ListUsers calls have been made. It's a map from oauth token -> URL string -> count
	madeUsers map[string]map[string]int
}

type mockGitLabOp struct {
	t *testing.T

	// users is a list of users on the GitLab instance
	users []*gitlab.User

	// publicProjs is the list of public project IDs
	publicProjs []int

	// internalProjs is the list of internal project IDs
	internalProjs []int

	// privateProjs is a map from { privateProjectID -> [ guestUserIDs, contentUserIDs ] } It
	// determines the structure of private project permissions. A "guest" user can access private
	// project metadata, but not project repository contents. A "content" user can access both.
	privateProjs map[int][2][]int32

	// oauthToks is a map from OAuth tokens to the corresponding GitLab user ID
	oauthToks map[string]int32

	// sudoTok, if non-empty, is the personal access token accepted with sudo permissions on this
	// instance. The mock implementation only supports having one such token value.
	sudoTok string
}

// newMockGitLab returns a new mockGitLab instance
func newMockGitLab(op mockGitLabOp) mockGitLab {
	projs := make(map[int]*gitlab.Project)
	privateGuest := make(map[int32][]int)
	privateRepo := make(map[int32][]int)
	for _, p := range op.publicProjs {
		projs[p] = &gitlab.Project{Visibility: gitlab.Public, ProjectCommon: gitlab.ProjectCommon{ID: p}}
	}
	for _, p := range op.internalProjs {
		projs[p] = &gitlab.Project{Visibility: gitlab.Internal, ProjectCommon: gitlab.ProjectCommon{ID: p}}
	}
	for p, userAccess := range op.privateProjs {
		projs[p] = &gitlab.Project{Visibility: gitlab.Private, ProjectCommon: gitlab.ProjectCommon{ID: p}}

		guestUsers, contentUsers := userAccess[0], userAccess[1]
		for _, u := range guestUsers {
			privateGuest[u] = append(privateGuest[u], p)
		}
		for _, u := range contentUsers {
			privateRepo[u] = append(privateRepo[u], p)
		}
	}
	return mockGitLab{
		t:                op.t,
		projs:            projs,
		users:            op.users,
		privateGuest:     privateGuest,
		privateRepo:      privateRepo,
		oauthToks:        op.oauthToks,
		sudoTok:          op.sudoTok,
		madeGetProject:   map[string]map[gitlab.GetProjectOp]int{},
		madeListProjects: map[string]map[string]int{},
		madeListTree:     map[string]map[gitlab.ListTreeOp]int{},
		madeUsers:        map[string]map[string]int{},
	}
}

func (m *mockGitLab) GetProject(c *gitlab.Client, ctx context.Context, op gitlab.GetProjectOp) (*gitlab.Project, error) {
	if _, ok := m.madeGetProject[c.Auth.Hash()]; !ok {
		m.madeGetProject[c.Auth.Hash()] = map[gitlab.GetProjectOp]int{}
	}
	m.madeGetProject[c.Auth.Hash()][op]++

	proj, ok := m.projs[op.ID]
	if !ok {
		return nil, gitlab.ErrProjectNotFound
	}
	if proj.Visibility == gitlab.Public {
		return proj, nil
	}
	if proj.Visibility == gitlab.Internal && m.isClientAuthenticated(c) {
		return proj, nil
	}

	acctID := m.getAcctID(c)
	for _, accessibleProjID := range append(m.privateGuest[acctID], m.privateRepo[acctID]...) {
		if accessibleProjID == op.ID {
			return proj, nil
		}
	}

	return nil, gitlab.ErrProjectNotFound
}

func (m *mockGitLab) ListProjects(c *gitlab.Client, ctx context.Context, urlStr string) (projs []*gitlab.Project, nextPageURL *string, err error) {
	if _, ok := m.madeListProjects[c.Auth.Hash()]; !ok {
		m.madeListProjects[c.Auth.Hash()] = map[string]int{}
	}
	m.madeListProjects[c.Auth.Hash()][urlStr]++

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, nil, err
	}
	query := u.Query()
	if query.Get("pagination") == "keyset" {
		return nil, nil, errors.New("This mock does not support keyset pagination")
	}
	perPage, err := strconv.Atoi(query.Get("per_page"))
	if err != nil {
		return nil, nil, err
	}
	page := 1
	if p := query.Get("page"); p != "" {
		page, err = strconv.Atoi(p)
		if err != nil {
			return nil, nil, err
		}
	}

	acctID := m.getAcctID(c)
	for _, proj := range m.projs {
		if proj.Visibility == gitlab.Public || (proj.Visibility == gitlab.Internal && acctID != 0) {
			projs = append(projs, proj)
		}
	}
	for _, pid := range m.privateGuest[acctID] {
		projs = append(projs, m.projs[pid])
	}
	for _, pid := range m.privateRepo[acctID] {
		projs = append(projs, m.projs[pid])
	}

	sort.Sort(projSort(projs))
	if (page-1)*perPage >= len(projs) {
		return nil, nil, nil
	}
	if page*perPage < len(projs) {
		nextURL, _ := url.Parse(urlStr)
		q := nextURL.Query()
		q.Set("page", strconv.Itoa(page+1))
		nextURL.RawQuery = q.Encode()
		nextURLStr := nextURL.String()
		return projs[(page-1)*perPage : page*perPage], &nextURLStr, nil
	}
	return projs[(page-1)*perPage:], nil, nil
}
func (m *mockGitLab) ListTree(c *gitlab.Client, ctx context.Context, op gitlab.ListTreeOp) ([]*gitlab.Tree, error) {
	if _, ok := m.madeListTree[c.Auth.Hash()]; !ok {
		m.madeListTree[c.Auth.Hash()] = map[gitlab.ListTreeOp]int{}
	}
	m.madeListTree[c.Auth.Hash()][op]++

	ret := []*gitlab.Tree{
		{
			ID:   "123",
			Name: "file.txt",
			Type: "blob",
			Path: "dir/file.txt",
			Mode: "100644",
		},
	}

	proj, ok := m.projs[op.ProjID]
	if !ok {
		return nil, gitlab.ErrProjectNotFound
	}
	if proj.Visibility == gitlab.Public {
		return ret, nil
	}
	if proj.Visibility == gitlab.Internal && m.isClientAuthenticated(c) {
		return ret, nil
	}

	acctID := m.getAcctID(c)
	for _, accessibleProjID := range m.privateRepo[acctID] {
		if accessibleProjID == op.ProjID {
			return ret, nil
		}
	}

	return nil, gitlab.ErrProjectNotFound
}

// isClientAuthenticated returns true if the client is authenticated. User is authenticated if OAuth
// token is non-empty (note: this mock impl doesn't verify validity of the OAuth token) or if the
// personal access token is non-empty (note: this mock impl requires that the PAT be equivalent to
// the mock GitLab sudo token).
func (m *mockGitLab) isClientAuthenticated(c *gitlab.Client) bool {
	return c.Auth.Hash() != "" || (m.sudoTok != "" && c.Auth.(*gitlab.SudoableToken).Token == m.sudoTok)
}

func (m *mockGitLab) getAcctID(c *gitlab.Client) int32 {
	if a, ok := c.Auth.(*auth.OAuthBearerToken); ok {
		return m.oauthToks[a.Hash()]
	}

	pat := c.Auth.(*gitlab.SudoableToken)
	if m.sudoTok != "" && m.sudoTok == pat.Token && pat.Sudo != "" {
		sudo, err := strconv.Atoi(pat.Sudo)
		if err != nil {
			m.t.Fatalf("mockGitLab requires all Sudo params to be numerical: %s", err)
		}
		return int32(sudo)
	}
	return 0
}

func (m *mockGitLab) ListUsers(c *gitlab.Client, ctx context.Context, urlStr string) (users []*gitlab.User, nextPageURL *string, err error) {
	key := ""
	if c.Auth != nil {
		key = c.Auth.Hash()
	}

	if _, ok := m.madeUsers[key]; !ok {
		m.madeUsers[key] = map[string]int{}
	}
	m.madeUsers[key][urlStr]++

	u, err := url.Parse(urlStr)
	if err != nil {
		m.t.Fatalf("could not parse ListUsers urlStr %q: %s", urlStr, err)
	}

	var matchingUsers []*gitlab.User
	for _, user := range m.users {
		userMatches := true
		if qExternUID := u.Query().Get("extern_uid"); qExternUID != "" {
			qProvider := u.Query().Get("provider")

			match := false
			for _, identity := range user.Identities {
				if identity.ExternUID == qExternUID && identity.Provider == qProvider {
					match = true
					break
				}
			}
			if !match {
				userMatches = false
			}
		}
		if qUsername := u.Query().Get("username"); qUsername != "" {
			if user.Username != qUsername {
				userMatches = false
			}
		}
		if userMatches {
			matchingUsers = append(matchingUsers, user)
		}
	}

	// pagination
	perPage, err := getIntOrDefault(u.Query().Get("per_page"), 10)
	if err != nil {
		return nil, nil, err
	}
	page, err := getIntOrDefault(u.Query().Get("page"), 1)
	if err != nil {
		return nil, nil, err
	}
	p := page - 1

	var pagedUsers []*gitlab.User

	if perPage*p > len(matchingUsers)-1 {
		pagedUsers = nil
	} else if perPage*(p+1) > len(matchingUsers)-1 {
		pagedUsers = matchingUsers[perPage*p:]
	} else {
		pagedUsers = matchingUsers[perPage*p : perPage*(p+1)]
		if perPage*(p+1) <= len(matchingUsers)-1 {
			newU := *u
			q := u.Query()
			q.Set("page", strconv.Itoa(page+1))
			newU.RawQuery = q.Encode()
			s := newU.String()
			nextPageURL = &s
		}
	}
	return pagedUsers, nextPageURL, nil
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

type mockAuthnProvider struct {
	configID  providers.ConfigID
	serviceID string
}

func (m mockAuthnProvider) ConfigID() providers.ConfigID {
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

func (m mockAuthnProvider) CachedInfo() *providers.Info {
	return &providers.Info{ServiceID: m.serviceID}
}

func (m mockAuthnProvider) Refresh(ctx context.Context) error {
	panic("should not be called")
}

func acct(t *testing.T, userID int32, serviceType, serviceID, accountID, oauthTok string) *extsvc.Account {
	var data extsvc.AccountData

	var authData *oauth2.Token
	if oauthTok != "" {
		authData = &oauth2.Token{AccessToken: oauthTok}
	}

	if serviceType == extsvc.TypeGitLab {
		gitlabAcctID, err := strconv.Atoi(accountID)
		if err != nil {
			t.Fatalf("Could not convert accountID to number: %s", err)
		}

		gitlab.SetExternalAccountData(&data, &gitlab.User{
			ID: int32(gitlabAcctID),
		}, authData)
	}

	return &extsvc.Account{
		UserID: userID,
		AccountSpec: extsvc.AccountSpec{
			ServiceType: serviceType,
			ServiceID:   serviceID,
			AccountID:   accountID,
		},
		AccountData: data,
	}
}

func repo(uri, serviceType, serviceID, id string) *types.Repo {
	return &types.Repo{
		Name: api.RepoName(uri),
		ExternalRepo: api.ExternalRepoSpec{
			ID:          id,
			ServiceType: serviceType,
			ServiceID:   serviceID,
		},
	}
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

func getIntOrDefault(str string, def int) (int, error) {
	if str == "" {
		return def, nil
	}
	return strconv.Atoi(str)
}

// projSort sorts Projects in order of ID
type projSort []*gitlab.Project

func (p projSort) Len() int           { return len(p) }
func (p projSort) Less(i, j int) bool { return p[i].ID < p[j].ID }
func (p projSort) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
