package repos

import (
	"context"
	"time"

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
