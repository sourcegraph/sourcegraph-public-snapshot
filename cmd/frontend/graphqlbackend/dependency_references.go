package graphqlbackend

import "github.com/sourcegraph/sourcegraph/pkg/api"

type dependencyReferencesResolver struct {
	dependencyReferenceData *dependencyReferencesDataResolver
	repoData                *repoDataMapResolver
}

type dependencyReferencesDataResolver struct {
	references []*dependencyReferenceResolver
	location   *dependencyLocationResolver
}

type dependencyReferenceResolver struct {
	dependencyData string
	repo           api.RepoID
	hints          string
}

type dependencyLocationResolver struct {
	location string
	symbol   string
}

type repoDataMapResolver struct {
	repos   []*repositoryResolver
	repoIDs []api.RepoID
}

func (r *repoDataMapResolver) Repos() []*repositoryResolver {
	return r.repos
}

func (r *repoDataMapResolver) RepoIDs() []int32 {
	return repoIDsToInt32s(r.repoIDs)
}

func (r *dependencyReferencesResolver) DependencyReferenceData() *dependencyReferencesDataResolver {
	return r.dependencyReferenceData
}

func (r *dependencyReferencesResolver) RepoData() *repoDataMapResolver {
	return r.repoData
}

func (r *dependencyReferencesDataResolver) References() []*dependencyReferenceResolver {
	return r.references
}

func (r *dependencyReferencesDataResolver) Location() *dependencyLocationResolver {
	return r.location
}

func (r *dependencyReferenceResolver) DependencyData() string {
	return r.dependencyData
}

func (r *dependencyReferenceResolver) RepoID() int32 {
	return int32(r.repo)
}

func (r *dependencyReferenceResolver) Hints() string {
	return r.hints
}

func (r *dependencyLocationResolver) Location() string {
	return r.location
}

func (r *dependencyLocationResolver) Symbol() string {
	return r.symbol
}
