package resolvers

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

func (r *componentResolver) Status(ctx context.Context) (gql.ComponentStatusResolver, error) {
	url, err := r.URL(ctx)
	if err != nil {
		return nil, err
	}

	var statusContexts []gql.ComponentStatusContextResolver

	{
		// Owner
		owner, err := r.Owner(ctx)
		if err != nil {
			return nil, err
		}

		sc := &componentStatusContextResolver{
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
		codeOwnerConnection, err := r.CodeOwners(ctx, &graphqlutil.ConnectionArgs{})
		if err != nil {
			// return nil, err
		}
		var codeOwners []gql.ComponentCodeOwnerEdgeResolver
		if codeOwnerConnection != nil {
			codeOwners = codeOwnerConnection.Edges()
		}

		if err == nil {
			sc := &componentStatusContextResolver{
				name:      "codeOwners",
				title:     "Code owners",
				targetURL: url + "/code",
			}
			if codeOwners == nil || len(codeOwners) == 0 {
				sc.state = "FAILURE"
				sc.description = "No code owners found"
			} else {
				sc.state = "INFO"
			}
			statusContexts = append(statusContexts, sc)
		}
	}

	{
		// Contributors
		contributorConnection, err := r.Contributors(ctx, &graphqlutil.ConnectionArgs{})
		if err != nil {
			return nil, err
		}
		var contributors []gql.ComponentAuthorEdgeResolver
		if contributorConnection != nil {
			contributors = contributorConnection.Edges()
		}

		sc := &componentStatusContextResolver{
			name:      "contributors",
			title:     "Contributors",
			targetURL: url + "/code",
		}
		if contributors == nil || len(contributors) == 0 {
			sc.state = "FAILURE"
			sc.description = "No contributors found"
		} else {
			sc.state = "INFO"
		}
		statusContexts = append(statusContexts, sc)
	}

	{
		// Usage
		usage, err := r.Usage(ctx)
		if err != nil {
			return nil, err
		}

		if usage != nil {
			usagePeople, err := usage.People(ctx)
			if err != nil {
				return nil, err
			}

			sc := &componentStatusContextResolver{
				name:      "usage",
				title:     "Usage",
				targetURL: url + "/usage",
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
		&componentStatusContextResolver{
			name:        "deploy",
			state:       "SUCCESS",
			title:       "Deploy",
			description: "Deployed `f38ca7d` to Sourcegraph.com 4 min ago ([monitor](#TODO))",
		},
		&componentStatusContextResolver{
			name:        "ci",
			state:       "SUCCESS",
			title:       "CI",
			description: "Build `f38ca7d` passed 7 min ago",
			targetURL:   "https://example.com",
		},
	)

	return &componentStatusResolver{
		contexts: statusContexts,
		entityID: r.ID(),
	}, nil
}
