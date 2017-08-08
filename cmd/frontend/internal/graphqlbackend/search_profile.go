package graphqlbackend

import (
	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type searchProfile struct {
	name        string
	description string
	repos       []*sourcegraph.Repo
}

func (p *searchProfile) Name() string {
	return p.name
}

func (p *searchProfile) Description() *string {
	if p.description == "" {
		return nil
	}
	return &p.description
}

func (p *searchProfile) Repositories() []*repositoryResolver {
	resolvers := make([]*repositoryResolver, 0, len(p.repos))
	for _, repo := range p.repos {
		resolvers = append(resolvers, &repositoryResolver{repo: repo})
	}
	return resolvers
}

func (*rootResolver) SearchProfiles(ctx context.Context) ([]*searchProfile, error) {
	active, inactive, err := listActiveAndInactive(ctx)
	if err != nil {
		return nil, err
	}
	profiles := []*searchProfile{}
	if len(active) > 0 {
		profiles = append(profiles, &searchProfile{
			name:        "Active",
			description: "Repositories that are active.",
			repos:       active,
		})
	}
	if len(inactive) > 0 {
		profiles = append(profiles, &searchProfile{
			name:        "Inactive",
			description: "Repositories that are marked inactive.",
			repos:       inactive,
		})
	}
	return profiles, nil
}
