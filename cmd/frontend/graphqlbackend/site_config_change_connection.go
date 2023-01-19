package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type SiteConfigurationChangeConnectionStore struct {
	db database.DB
}

func (s *SiteConfigurationChangeConnectionStore) ComputeTotal(ctx context.Context) (*int32, error) {
	count, err := s.db.Conf().GetSiteConfigCount(ctx)
	c := int32(count)
	return &c, err
}

func copyIntPtr(n *int) *int {
	if n == nil {
		return nil
	}

	c := *n
	return &c
}

func (s *SiteConfigurationChangeConnectionStore) ComputeNodes(ctx context.Context, args *database.PaginationArgs) ([]*SiteConfigurationChangeResolver, error) {
	isModifiedPaginationArgs := false

	var paginationArgs *database.PaginationArgs
	// var wanted int
	if args != nil {
		// NOTE: Do not modify "args" in-place because it is used by the caller of ComputeNodes to
		// determine next / previous page. Instead, dereference the values from args first (if
		// they're non-nil) and then assign them address of the new variables.
		paginationArgs = &database.PaginationArgs{
			First:  copyIntPtr(args.First),
			Last:   copyIntPtr(args.Last),
			After:  copyIntPtr(args.After),
			Before: copyIntPtr(args.Before),
		}

		isModifiedPaginationArgs = modifyArgs(paginationArgs)
	}

	history, err := s.db.Conf().ListSiteConfigs(ctx, paginationArgs)
	if err != nil {
		return []*SiteConfigurationChangeResolver{}, err
	}

	resolvers := []*SiteConfigurationChangeResolver{}
	for i := 0; i < len(history); i++ {

		resolvers = append(resolvers, &SiteConfigurationChangeResolver{
			db:         s.db,
			siteConfig: history[i],
			// previousSiteConfig: previousSiteConfig,
		})
	}

	if isModifiedPaginationArgs {
		if paginationArgs.Last != nil {
			resolvers = resolvers[1:]
		} else if paginationArgs.First != nil && len(history) == *paginationArgs.First {
			resolvers = resolvers[:len(resolvers)-1]
		}
	}

	return resolvers, nil
}

// modifyArgs will fetch one more than the originally requested number of items because we need one
// older item to get the diff of the oldes item in the list.
//
// A separate function so that this can be tested in isolation.
func modifyArgs(args *database.PaginationArgs) bool {
	var modified bool
	if args.First != nil {
		*args.First += 1
		modified = true
	} else if args.Last != nil && args.Before != nil {
		if *args.Before > 0 {
			modified = true
			*args.Last += 1
			*args.Before -= 1
		}
	}

	return modified
}

// FIXME: Implement when paginating.
func (s *SiteConfigurationChangeConnectionStore) MarshalCursor(node *SiteConfigurationChangeResolver) (*string, error) {
	return nil, nil
}

// FIXME: Implement when paginating.
func (s *SiteConfigurationChangeConnectionStore) UnmarshalCursor(cursor string) (*int, error) {
	return nil, nil
}
