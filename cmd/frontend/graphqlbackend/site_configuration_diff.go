package graphqlbackend

import (
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

const siteConfigurationKind = "SiteConfiguration"

type SiteConfigurationChangeResolver struct {
	db         database.DB
	siteConfig *database.SiteConfig
}

func (r SiteConfigurationChangeResolver) ID() graphql.ID {
	return relay.MarshalID(siteConfigurationKind, r.siteConfig.ID)
}

func (r SiteConfigurationChangeResolver) PreviousID() *graphql.ID {
	// FIXME: Needs to return previous ID
	return nil
}

func (r SiteConfigurationChangeResolver) Author() *UserResolver {
	// FIXME
	return NewUserResolver(r.db, nil)
}

func (r SiteConfigurationChangeResolver) Diff() string {
	// FIXME
	return "do do diff"
}

func (r SiteConfigurationChangeResolver) CreatedAt() gqlutil.DateTime {
	// FIXME
	return gqlutil.DateTime{Time: time.Now()}
}

func (r SiteConfigurationChangeResolver) UpdatedAt() gqlutil.DateTime {
	// FIXME
	return gqlutil.DateTime{Time: time.Now()}
}
