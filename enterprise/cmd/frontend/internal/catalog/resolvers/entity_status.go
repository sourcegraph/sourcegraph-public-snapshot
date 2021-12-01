package resolvers

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (r *catalogComponentResolver) Status(ctx context.Context) (gql.CatalogEntityStatusResolver, error) {
	var statusContexts []gql.CatalogEntityStatusContextResolver

	{
		// Owners
		owners, err := r.Owners(ctx)
		if err != nil {
			return nil, err
		}

		sc := &catalogEntityStatusContextResolver{
			name:      "owners",
			title:     "Owners",
			targetURL: r.URL() + "/code",
		}
		if owners == nil || len(*owners) == 0 {
			sc.state = "FAILURE"
			sc.description = "No code owners found"
		} else {
			sc.state = "INFO"
		}
		statusContexts = append(statusContexts, sc)
	}

	{
		// Authors
		authors, err := r.Authors(ctx)
		if err != nil {
			return nil, err
		}

		sc := &catalogEntityStatusContextResolver{
			name:      "authors",
			title:     "Authors",
			targetURL: r.URL() + "/code",
		}
		if authors == nil || len(*authors) == 0 {
			sc.state = "FAILURE"
			sc.description = "No authors found"
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
		contexts:  statusContexts,
		component: r,
	}, nil
}

type catalogEntityStatusResolver struct {
	contexts  []gql.CatalogEntityStatusContextResolver
	component *catalogComponentResolver
}

func (r *catalogEntityStatusResolver) ID() graphql.ID {
	return relay.MarshalID("CatalogEntityStatus", r.component.Name())
}

func (r *catalogEntityStatusResolver) Contexts() []gql.CatalogEntityStatusContextResolver {
	return r.contexts
}

type catalogEntityStatusContextResolver struct {
	name, title, description, targetURL string
	state                               gql.CatalogEntityStatusState
}

func (r *catalogEntityStatusContextResolver) ID() graphql.ID {
	b := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s", r.name, r.title, r.description)))
	return relay.MarshalID("CatalogEntityStatus", base64.RawURLEncoding.EncodeToString(b[:16]))
}

func (r *catalogEntityStatusContextResolver) Name() string                        { return r.name }
func (r *catalogEntityStatusContextResolver) State() gql.CatalogEntityStatusState { return r.state }
func (r *catalogEntityStatusContextResolver) Title() string                       { return r.title }
func (r *catalogEntityStatusContextResolver) Description() *string {
	if r.description == "" {
		return nil
	}
	return &r.description
}
func (r *catalogEntityStatusContextResolver) TargetURL() *string {
	if r.targetURL == "" {
		return nil
	}
	return &r.targetURL
}
