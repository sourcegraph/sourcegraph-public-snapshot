package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type SiteConfigurationChangeConnectionResolver struct {
	graphqlutil.ConnectionResolverArgs
	db database.DB
}

type SiteConfigurationChangeConnectionResolverArgs struct {
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
// FIXME: Last does not work yet
func (s *SiteConfigurationChangeConnectionStore) ComputeNodes(ctx context.Context, args *database.PaginationArgs) ([]*SiteConfigurationChangeResolver, error) {

	// Ascending order by default.
	opt := database.SiteConfigListOptions{
		OrderByDirection: database.AscendingOrderByDirection,
	}
	if args != nil {
		if args.First != nil {
			opt.LimitOffset = &database.LimitOffset{Limit: *args.First}
		} else if args.Last != nil {
			opt.LimitOffset = &database.LimitOffset{Limit: *args.Last}
			opt.OrderByDirection = database.DescendingOrderByDirection
		}
	}

	// Fetch 100 entries by default.
	if opt.LimitOffset == nil {
		opt.LimitOffset = &database.LimitOffset{Limit: 100}
	}

	// Get one more than needed to read the previousSiteConfig of the last item.
	opt.LimitOffset.Limit += 1

	history, err := s.db.Conf().ListSiteConfigs(ctx, opt)
	if err != nil {
		return nil, err
	}

	// Iterate from the end if the ORDER BY DESC was used so that we can retrieve the previousSiteConfig easily.

	// // We fetched one more than requested, so stop before we reach the end.
	// for i := len(history); i >= 1; i-- {
	// 	resolvers = append(resolvers, &SiteConfigurationChangeResolver{
	// 		db:                 s.db,
	// 		siteConfig:         history[i],
	// 		previousSiteConfig: history[i-1],
	// 	})
	// }

	// Now flip it back

	// Reset the state.
	opt.LimitOffset.Limit -= 1

	// Only truncate the results if we fetched more than the total requested in the API call.
	// But if we fetched less or equal to the total, keep the entire result set.
	total := len(history)
	if total > opt.LimitOffset.Limit {
		total = len(history) - 1
	}

	resolvers := []*SiteConfigurationChangeResolver{}
	for i := 0; i < total; i++ {
		var previousSiteConfig *database.SiteConfig

		if opt.OrderByDirection == database.DescendingOrderByDirection {
			//
			// if i < (limit-1) && limit < len(history) {
			// 	previousSiteConfig = history[i+1]
			// }

			// The last element?
			// if i < (total-1) && total < len(history) {
			// 	previousSiteConfig = history[i+1]
			// }

			if opt.LimitOffset.Limit >= len(history) {
				if i == (total - 1) {
					previousSiteConfig = nil
				} else {
					previousSiteConfig = history[i+1]
				}
			} else {
				previousSiteConfig = history[i+1]
			}

		} else {
			if i > 0 {
				previousSiteConfig = history[i-1]
			}
		}
		resolvers = append(resolvers, &SiteConfigurationChangeResolver{
			db:                 s.db,
			siteConfig:         history[i],
			previousSiteConfig: previousSiteConfig,
		})
	}

	// } else {
	// We fetched one more than requested, so stop before we reach the end.
	// 	for _, config := range history[:len(history)-1] {
	// 		resolvers = append(resolvers, &SiteConfigurationChangeResolver{
	// 			db:                 s.db,
	// 			siteConfig:         config,
	// 			previousSiteConfig: previousSiteConfig,
	// 		})

	// 		previousSiteConfig = config
	// 	}
	// }

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
