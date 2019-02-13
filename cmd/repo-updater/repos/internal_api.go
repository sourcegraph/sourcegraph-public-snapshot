package repos

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// InternalAPI captures the internal API methods needed for syncing external services' repos.
type InternalAPI interface {
	ExternalServicesList(context.Context, api.ExternalServicesListRequest) ([]*api.ExternalService, error)
	ReposCreateIfNotExists(context.Context, api.RepoCreateOrUpdateRequest) (*api.Repo, error)
	ReposUpdateMetadata(ctx context.Context, repo api.RepoName, description string, fork, archived bool) error
}

// NewInternalAPI returns a new internal API client with the given timeout for outgoing calls.
func NewInternalAPI(timeout time.Duration) InternalAPI {
	return &internalAPI{timeout: timeout}
}

// internalAPI wraps api.InternalClient with timeouts.
type internalAPI struct{ timeout time.Duration }

// ExternalServicesList lists external services of the given Kind.
func (a *internalAPI) ExternalServicesList(ctx context.Context, opts api.ExternalServicesListRequest) ([]*api.ExternalService, error) {
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()
	return api.InternalClient.ExternalServicesList(ctx, opts)
}

// RepoCreateIfNotExists creates the given repo if it doesn't exist.
func (a *internalAPI) ReposCreateIfNotExists(ctx context.Context, req api.RepoCreateOrUpdateRequest) (*api.Repo, error) {
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()
	return api.InternalClient.ReposCreateIfNotExists(ctx, req)
}

// ReposUpdateMetdata updates the metadata of repo with the given name.
func (a *internalAPI) ReposUpdateMetadata(ctx context.Context, repoName api.RepoName, description string, fork, archived bool) error {
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()
	return api.InternalClient.ReposUpdateMetadata(ctx, repoName, description, fork, archived)
}

//
// Test utilities
//

// FakeInternalAPI implements the InternalAPI interface with the given in memory data.
// It's safe for concurrent use.
type FakeInternalAPI struct {
	mu     sync.RWMutex
	svcs   map[string][]*api.ExternalService
	repos  map[api.RepoName]*api.Repo
	repoID api.RepoID
}

// NewFakeInternalAPI returns a new FakeInternalAPI initialised with the given data.
func NewFakeInternalAPI(svcs []*api.ExternalService, repos []*api.Repo) *FakeInternalAPI {
	fa := FakeInternalAPI{
		svcs:  map[string][]*api.ExternalService{},
		repos: map[api.RepoName]*api.Repo{},
	}

	for _, svc := range svcs {
		fa.svcs[svc.Kind] = append(fa.svcs[svc.Kind], svc)
	}

	for _, repo := range repos {
		fa.repos[repo.Name] = repo
	}

	return &fa
}

// ExternalServicesList lists external services of the given Kind. A non-existent kind
// will result in an error being returned.
func (a *FakeInternalAPI) ExternalServicesList(
	_ context.Context,
	req api.ExternalServicesListRequest,
) ([]*api.ExternalService, error) {

	a.mu.RLock()
	defer a.mu.RUnlock()

	var svcs []*api.ExternalService
	for _, kind := range req.Kinds {
		svcs = append(svcs, a.svcs[kind]...)
	}

	sort.Slice(svcs, func(i, j int) bool { return svcs[i].ID < svcs[j].ID })

	return svcs, nil
}

// ReposList returns the list of all repos in the API.
func (a *FakeInternalAPI) ReposList() []*api.Repo {
	a.mu.RLock()
	defer a.mu.RUnlock()

	repos := make([]*api.Repo, 0, len(a.repos))
	for _, repo := range a.repos {
		repos = append(repos, repo)
	}

	return repos
}

// RepoCreateIfNotExists creates the given repo if it doesn't exist. Repos with
// with an empty name are invalid and will result in an error to be returned.
func (a *FakeInternalAPI) ReposCreateIfNotExists(
	ctx context.Context,
	req api.RepoCreateOrUpdateRequest,
) (*api.Repo, error) {

	if req.RepoName == "" {
		return nil, errors.New("invalid empty repo name")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	repo, ok := a.repos[req.RepoName]
	if !ok {
		a.repoID++
		repo = &api.Repo{
			ID:           a.repoID,
			Name:         req.RepoName,
			Enabled:      req.Enabled,
			ExternalRepo: req.ExternalRepo,
		}
		a.repos[req.RepoName] = repo
	}

	return repo, nil
}

// ReposUpdateMetdata updates the metadata of repo with the given name.
// Non-existent repos return an error.
func (a *FakeInternalAPI) ReposUpdateMetadata(
	ctx context.Context,
	repoName api.RepoName,
	description string,
	fork, archived bool,
) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	_, ok := a.repos[repoName]
	if !ok {
		return fmt.Errorf("repo %q not found", repoName)
	}

	// Fake success, no updates needed because the returned types (api.Repo)
	// don't include these fields.

	return nil
}
