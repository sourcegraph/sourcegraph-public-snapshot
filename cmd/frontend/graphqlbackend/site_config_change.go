package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

const siteConfigurationChangeKind = "SiteConfigurationChange"

type SiteConfigurationChangeResolver struct {
	db                 database.DB
	siteConfig         *database.SiteConfig
	previousSiteConfig *database.SiteConfig
}

func (r SiteConfigurationChangeResolver) ID() graphql.ID {
	return marshalSiteConfigurationChangeID(r.siteConfig.ID)
}

// One line wrapper to be able to use in tests as well.
func marshalSiteConfigurationChangeID(id int32) graphql.ID {
	return relay.MarshalID(siteConfigurationChangeKind, &id)
}

func (r SiteConfigurationChangeResolver) Author(ctx context.Context) (*UserResolver, error) {
	if r.siteConfig.AuthorUserID == 0 {
		return nil, nil
	}

	user, err := UserByIDInt32(ctx, r.db, r.siteConfig.AuthorUserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// TODO: Implement this.
func (r SiteConfigurationChangeResolver) Diff() string {
	// TODO: We will do something like this, but for now return an empty string to not leak secrets
	// until we have implemented redaction.
	//
	// if r.previousSiteConfig == nil {
	// 	return ""
	// }

	// return cmp.Diff(r.siteConfig.Contents, r.previousSiteConfig.Contents)
	return ""
}

func (r SiteConfigurationChangeResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.siteConfig.CreatedAt}
}

func (r SiteConfigurationChangeResolver) UpdatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.siteConfig.UpdatedAt}
}
