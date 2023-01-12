package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type SiteConfigurationChangeConnectionResolver struct {
	graphqlutil.ConnectionResolverArgs
	db database.DB
}

type SiteConfigurationChangeConnectionResolverArgs struct {
}

func (r SiteConfigurationChangeConnectionResolver) ID() graphql.ID {
	return graphql.ID("FIXME")
}

// FIXME: Do we need more args here?
type SiteConfigurationChangeConnectionStore struct {
	db database.DB
}

func (s *SiteConfigurationChangeConnectionStore) ComputeTotal(ctx context.Context) (*int32, error) {
	count, err := s.db.Conf().GetSiteConfigCount(ctx)
	c := int32(count)
	return &c, err
}

// FIXME: Pagination offset is not implemented yet.
func (s *SiteConfigurationChangeConnectionStore) ComputeNodes(ctx context.Context, args *database.PaginationArgs) ([]*SiteConfigurationChangeResolver, error) {
	opt := database.SiteConfigListOptions{}
	if args.First != nil {
		opt.LimitOffset = &database.LimitOffset{Limit: *args.First}
		opt.OrderByDirection = database.AscendingOrderByDirection
	} else if args.Last != nil {
		opt.LimitOffset = &database.LimitOffset{Limit: *args.Last}
		opt.OrderByDirection = database.DescendingOrderByDirection
	}

	history, err := s.db.Conf().ListSiteConfig(ctx, opt)
	if err != nil {
		return nil, err
	}

	resolvers := []*SiteConfigurationChangeResolver{}
	var previousSiteConfig *database.SiteConfig
	for _, config := range history {
		resolvers = append(resolvers, &SiteConfigurationChangeResolver{
			db:                 s.db,
			siteConfig:         config,
			previousSiteConfig: previousSiteConfig,
		})

		previousSiteConfig = config
	}

	return resolvers, nil
}

// FIXME: Implement when paginating.
func (s *SiteConfigurationChangeConnectionStore) MarshalCursor(node *SiteConfigurationChangeResolver) (*string, error) {
	return nil, nil
}

// FIXME: Implement when paginating.
func (s *SiteConfigurationChangeConnectionStore) UnmarshalCursor(cursor string) (*int, error) {
	return nil, nil
}
