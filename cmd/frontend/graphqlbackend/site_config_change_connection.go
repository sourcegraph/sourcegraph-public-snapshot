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

	// Reset the state since we fetched one more than asked for.
	opt.LimitOffset.Limit -= 1

	// Limit the results if we fetched more than the upperLimit requested in the API call.
	// But if we fetched less or equal to the upperLimit, keep the entire result set.
	upperLimit := len(history)
	if upperLimit > opt.LimitOffset.Limit {
		upperLimit = len(history) - 1
	}

	resolvers := []*SiteConfigurationChangeResolver{}
	for i := 0; i < upperLimit; i++ {
		var previousSiteConfig *database.SiteConfig

		// FOR ORDER BY DESC, we need to look ahead for the previousSiteConfig.
		if opt.OrderByDirection == database.DescendingOrderByDirection {
			// But, if the total number of items fetched from the DB is <= the original limit
			// requested in the API call, this would mean the DB only has that many change entries
			// and we need to handle the edge case where the very first entry will have no
			// previousSiteConfig.
			if len(history) <= opt.LimitOffset.Limit {
				// Only look ahead if we are not already at the last index to avoid out of bounds
				// error. For the last index previousSiteConfig is already nil.
				if i < (upperLimit - 1) {
					previousSiteConfig = history[i+1]
				}
			} else {
				// There's more in the DB and we fetched one more than the user requested, so safely
				// look ahead to get the previousSiteConfig. This is safe from out of bounds error
				// because we only iterate upto the last but one element if we fetched one more than
				// the original limit in the API call.
				previousSiteConfig = history[i+1]
			}
		} else if i > 0 {
			// For ORDDER BY ASC is far simpler.
			//
			// We only need to look behind for the previousSiteConfig, but make sure to not do that
			// if we're at the 0th index to avoid out of bounds errors.
			//
			// TODO: This will change as soon as offset comes into play. 😞
			previousSiteConfig = history[i-1]
		}

		resolvers = append(resolvers, &SiteConfigurationChangeResolver{
			db:                 s.db,
			siteConfig:         history[i],
			previousSiteConfig: previousSiteConfig,
		})
	}

	return resolvers, nil
}

func (s *SiteConfigurationChangeConnectionStore) ComputeNodes_Reincarnated(ctx context.Context, args *database.PaginationArgs) ([]*SiteConfigurationChangeConnectionResolver, error) {

}

// FIXME: Implement when paginating.
func (s *SiteConfigurationChangeConnectionStore) MarshalCursor(node *SiteConfigurationChangeResolver) (*string, error) {
	return nil, nil
}

// FIXME: Implement when paginating.
func (s *SiteConfigurationChangeConnectionStore) UnmarshalCursor(cursor string) (*int, error) {
	return nil, nil
}
