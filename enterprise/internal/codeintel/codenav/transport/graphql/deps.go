package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/api"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

// TODO - the good things
func (r *rootResolver) Dependencies(ctx context.Context, args *resolverstubs.DependenciesArgs) (_ []resolverstubs.DependencyDescriptionResolver, err error) {
	repositoryID, err := graphqlbackend.UnmarshalRepositoryID(args.Repository)
	if err != nil {
		return nil, err
	}

	deps, err := r.svc.GetDependencies(ctx, int(repositoryID))
	if err != nil {
		return nil, err
	}

	var resolvers []resolverstubs.DependencyDescriptionResolver
	for _, dep := range deps {
		resolvers = append(resolvers, NewDependencyResolver(dep.Manager, dep.Name, dep.Version))
	}

	return resolvers, err
}

func (r *rootResolver) DependencyOccurrences(ctx context.Context, args *resolverstubs.DependencyOccurrencesArgs) (_ []resolverstubs.LocationResolver, err error) {
	repositoryID, err := graphqlbackend.UnmarshalRepositoryID(args.Repository)
	if err != nil {
		return nil, err
	}

	occurrences, err := r.svc.GetDependencyOccurrences(ctx, int(repositoryID), args.Manager, args.Name, args.Version)
	if err != nil {
		return nil, err
	}

	locationResolver := sharedresolvers.NewCachedLocationResolver(r.svc.GetUnsafeDB(), gitserver.NewClient())

	var locations []resolverstubs.LocationResolver
	for _, location := range occurrences {
		treeResolver, err := locationResolver.Path(ctx, api.RepoID(location.RepositoryID), location.Commit, location.Path)
		if err != nil || treeResolver == nil {
			return nil, err
		}

		lspRange := convertRange(location.Range)
		locations = append(locations, NewLocationResolver(treeResolver, &lspRange))
	}

	return locations, nil
}

func (r *rootResolver) Vulnerabilities(ctx context.Context, args *resolverstubs.VulnerabilitiesArgs) (_ []resolverstubs.VulnerabilityResolver, err error) {
	repositoryID, err := graphqlbackend.UnmarshalRepositoryID(args.Repository)
	if err != nil {
		return nil, err
	}

	deps, err := r.svc.GetDependencies(ctx, int(repositoryID))
	if err != nil {
		return nil, err
	}

	var resolvers []resolverstubs.VulnerabilityResolver
	for _, dep := range deps {
		if dep.Manager != "npm" {
			continue
		}
		vulnerabilities, err := jsDepToVuln(apiKey, dep.Name, dep.Version)
		if err != nil {
			return nil, err
		}
		if len(vulnerabilities) == 0 {
			continue
		}

		occurrences, err := r.svc.GetDependencyOccurrences(ctx, int(repositoryID), dep.Manager, dep.Name, dep.Version)
		if err != nil {
			return nil, err
		}

		locationResolver := sharedresolvers.NewCachedLocationResolver(r.svc.GetUnsafeDB(), gitserver.NewClient())

		var locations []resolverstubs.LocationResolver
		for _, location := range occurrences {
			treeResolver, err := locationResolver.Path(ctx, api.RepoID(location.RepositoryID), location.Commit, location.Path)
			if err != nil || treeResolver == nil {
				return nil, err
			}

			lspRange := convertRange(location.Range)
			locations = append(locations, NewLocationResolver(treeResolver, &lspRange))
		}

		for _, vulnerability := range vulnerabilities {
			resolvers = append(resolvers, NewVulnerabilityResolver(vulnerability, locations))
		}
	}

	return resolvers, err
}

type dependencyDescriptionResolver struct {
	manager string
	name    string
	version string
}

func NewDependencyResolver(manager, name, version string) resolverstubs.DependencyDescriptionResolver {
	return &dependencyDescriptionResolver{
		manager: manager,
		name:    name,
		version: version,
	}
}

func (r *dependencyDescriptionResolver) Manager() string { return r.manager }
func (r *dependencyDescriptionResolver) Name() string    { return r.name }
func (r *dependencyDescriptionResolver) Version() string { return r.version }

type vulnerabilityResolver struct {
	vulnerability Vulnerability
	locations     []resolverstubs.LocationResolver
}

func NewVulnerabilityResolver(vulnerability Vulnerability, locations []resolverstubs.LocationResolver) resolverstubs.VulnerabilityResolver {
	return &vulnerabilityResolver{
		vulnerability: vulnerability,
		locations:     locations,
	}
}

func (r *vulnerabilityResolver) CVE() string             { return r.vulnerability.CVE }
func (r *vulnerabilityResolver) Description() string     { return r.vulnerability.Description }
func (r *vulnerabilityResolver) Dependency() string      { return r.vulnerability.Dependency }
func (r *vulnerabilityResolver) PackageManager() string  { return r.vulnerability.PackageManager }
func (r *vulnerabilityResolver) PublishedDate() string   { return r.vulnerability.PublishedDate }
func (r *vulnerabilityResolver) LastUpdate() string      { return r.vulnerability.LastUpdate }
func (r *vulnerabilityResolver) AffectedVersion() string { return r.vulnerability.AffectedVersion }
func (r *vulnerabilityResolver) CurrentVersion() string  { return r.vulnerability.CurrentVersion }
func (r *vulnerabilityResolver) SeverityScore() string   { return r.vulnerability.SeverityScore }
func (r *vulnerabilityResolver) SeverityString() string  { return r.vulnerability.SeverityString }

func (r *vulnerabilityResolver) Locations() []resolverstubs.LocationResolver {
	return r.locations
}
