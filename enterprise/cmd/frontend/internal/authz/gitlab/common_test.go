package gitlab

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/schema"
	"golang.org/x/oauth2"
)

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

	// privateGuest is a map from GitLab user ID to list of metadata-accessible private project IDs on GitLab
	privateGuest map[string][]int

	// privateRepo is a map from GitLab user ID to list of repo-content-accessible private project IDs on GitLab.
	// Projects in each list are also metadata-accessible.
	privateRepo map[string][]int

	// oauthToks is a map from OAuth token to GitLab user account ID
	oauthToks map[string]string

	// madeGetProject records what GetProject calls have been made. It's a map from oauth token -> GetProjectOp -> count.
	madeGetProject map[string]map[gitlab.GetProjectOp]int

	// madeListTree recores what ListTree calls have been made. It's a map from oauth token -> ListTreeOp -> count.
	madeListTree map[string]map[gitlab.ListTreeOp]int
}

// newMockGitLab returns a new mockGitLab instance
func newMockGitLab(
	t *testing.T, publicProjs []int, internalProjs []int, privateProjs map[int][2][]string, // privateProjs maps from { projID -> [ guestUserIDs, contentUserIDs ] }
	oauthToks map[string]string,
) mockGitLab {
	projs := make(map[int]*gitlab.Project)
	privateGuest := make(map[string][]int)
	privateRepo := make(map[string][]int)
	for _, p := range publicProjs {
		projs[p] = &gitlab.Project{Visibility: gitlab.Public, ProjectCommon: gitlab.ProjectCommon{ID: p}}
	}
	for _, p := range internalProjs {
		projs[p] = &gitlab.Project{Visibility: gitlab.Internal, ProjectCommon: gitlab.ProjectCommon{ID: p}}
	}
	for p, userAccess := range privateProjs {
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
		t:              t,
		projs:          projs,
		privateGuest:   privateGuest,
		privateRepo:    privateRepo,
		oauthToks:      oauthToks,
		madeGetProject: map[string]map[gitlab.GetProjectOp]int{},
		madeListTree:   map[string]map[gitlab.ListTreeOp]int{},
	}
}

func (m *mockGitLab) GetProject(c *gitlab.Client, ctx context.Context, op gitlab.GetProjectOp) (*gitlab.Project, error) {
	if _, ok := m.madeGetProject[c.OAuthToken]; !ok {
		m.madeGetProject[c.OAuthToken] = map[gitlab.GetProjectOp]int{}
	}
	m.madeGetProject[c.OAuthToken][op]++

	proj, ok := m.projs[op.ID]
	if !ok {
		return nil, gitlab.ErrNotFound
	}
	if proj.Visibility == gitlab.Public {
		return proj, nil
	}
	if c.OAuthToken != "" && proj.Visibility == gitlab.Internal {
		return proj, nil
	}

	acctID := m.oauthToks[c.OAuthToken]
	for _, accessibleProjID := range append(m.privateGuest[acctID], m.privateRepo[acctID]...) {
		if accessibleProjID == op.ID {
			return proj, nil
		}
	}

	return nil, gitlab.ErrNotFound
}

func (m *mockGitLab) ListTree(c *gitlab.Client, ctx context.Context, op gitlab.ListTreeOp) ([]*gitlab.Tree, error) {
	if _, ok := m.madeListTree[c.OAuthToken]; !ok {
		m.madeListTree[c.OAuthToken] = map[gitlab.ListTreeOp]int{}
	}
	m.madeListTree[c.OAuthToken][op]++

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
		return nil, gitlab.ErrNotFound
	}
	if proj.Visibility == gitlab.Public {
		return ret, nil
	}
	if c.OAuthToken != "" && proj.Visibility == gitlab.Internal {
		return ret, nil
	}

	acctID := m.oauthToks[c.OAuthToken]
	for _, accessibleProjID := range m.privateRepo[acctID] {
		if accessibleProjID == op.ProjID {
			return ret, nil
		}
	}

	return nil, gitlab.ErrNotFound
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
