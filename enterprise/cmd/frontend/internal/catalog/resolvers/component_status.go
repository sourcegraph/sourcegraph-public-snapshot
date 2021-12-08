package resolvers

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (r *catalogComponentResolver) Status(ctx context.Context) (gql.CatalogEntityStatusResolver, error) {
	var statusContexts []gql.CatalogEntityStatusContextResolver

	{
		// Owner
		owner, err := r.Owner(ctx)
		if err != nil {
			return nil, err
		}

		sc := &catalogEntityStatusContextResolver{
			name:  "owner",
			title: "Owner",
		}
		if owner == nil {
			sc.state = "FAILURE"
			sc.description = "No owner specified"
		} else {
			sc.state = "INFO"
		}
		statusContexts = append(statusContexts, sc)
	}

	{
		// Code owners
		codeOwners, err := r.CodeOwners(ctx)
		if err != nil {
			// return nil, err
		}

		if err == nil {
			sc := &catalogEntityStatusContextResolver{
				name:      "codeOwners",
				title:     "Code owners",
				targetURL: r.URL() + "/code",
			}
			if codeOwners == nil || len(*codeOwners) == 0 {
				sc.state = "FAILURE"
				sc.description = "No code owners found"
			} else {
				sc.state = "INFO"
			}
			statusContexts = append(statusContexts, sc)
		}
	}

	{
		// Authors
		authors, err := r.Authors(ctx)
		if err != nil {
			return nil, err
		}

		sc := &catalogEntityStatusContextResolver{
			name:      "contributors",
			title:     "Contributors",
			targetURL: r.URL() + "/code",
		}
		if authors == nil || len(*authors) == 0 {
			sc.state = "FAILURE"
			sc.description = "No contributors found"
		} else {
			sc.state = "INFO"
		}
		statusContexts = append(statusContexts, sc)
	}

	{
		// Usage
		usage, err := r.Usage(ctx, &gql.CatalogComponentUsageArgs{})
		if err != nil {
			return nil, err
		}

		if usage != nil {
			usagePeople, err := usage.People(ctx)
			if err != nil {
				return nil, err
			}

			sc := &catalogEntityStatusContextResolver{
				name:      "usage",
				title:     "Usage",
				targetURL: r.URL() + "/usage",
			}
			if usagePeople == nil || len(usagePeople) == 0 {
				sc.state = "FAILURE"
				sc.description = "No users found"
			} else {
				sc.state = "INFO"
			}
			statusContexts = append(statusContexts, sc)
		}
	}

	statusContexts = append(statusContexts,
		&catalogEntityStatusContextResolver{
			name:        "deploy",
			state:       "SUCCESS",
			title:       "Deploy",
			description: "Deployed `f38ca7d` to Sourcegraph.com 4 min ago ([monitor](#TODO))",
		},
		&catalogEntityStatusContextResolver{
			name:        "ci",
			state:       "SUCCESS",
			title:       "CI",
			description: "Build `f38ca7d` passed 7 min ago",
			targetURL:   "https://example.com",
		},
	)

	return &catalogEntityStatusResolver{
		contexts: statusContexts,
		entityID: r.ID(),
	}, nil
}
