package types

import "github.com/sourcegraph/sourcegraph/internal/authz"

type ProviderInitResult struct {
	UserPermissionsFetchers []authz.UserPermissionsFetcher
	RepoPermissionsFetchers []authz.RepoPermissionsFetcher
	Problems                []string
	Warnings                []string
	InvalidConnections      []string
}

func (r *ProviderInitResult) Append(res *ProviderInitResult) {
	r.UserPermissionsFetchers = append(r.UserPermissionsFetchers, res.UserPermissionsFetchers...)
	r.RepoPermissionsFetchers = append(r.RepoPermissionsFetchers, res.RepoPermissionsFetchers...)
	r.Problems = append(r.Problems, res.Problems...)
	r.Warnings = append(r.Warnings, res.Warnings...)
	r.InvalidConnections = append(r.InvalidConnections, res.InvalidConnections...)
}
