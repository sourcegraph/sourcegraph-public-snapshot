package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
	if args == nil {
		return []*SiteConfigurationChangeResolver{}, errors.New("pagintation args cannot be nil")
	}

	// NOTE: Do not modify "args" in-place because it is used by the caller of ComputeNodes to
	// determine next / previous page. Instead, dereference the values from args first (if
	// they're non-nil) and then assign them address of the new variables.
	paginationArgs := &database.PaginationArgs{
		First:  copyIntPtr(args.First),
		Last:   copyIntPtr(args.Last),
		After:  copyIntPtr(args.After),
		Before: copyIntPtr(args.Before),
	}

	isModifiedPaginationArgs := modifyArgs(paginationArgs)

	history, err := s.db.Conf().ListSiteConfigs(ctx, paginationArgs)
	if err != nil {
		return []*SiteConfigurationChangeResolver{}, err
	}

	totalFetched := len(history)
	if totalFetched == 0 {
		return []*SiteConfigurationChangeResolver{}, nil
	}

	resolvers := []*SiteConfigurationChangeResolver{}
	for i := 0; i < totalFetched; i++ {
		var previousSiteConfig *database.SiteConfig
		// If First is used then "history" is in descending order: 5, 4, 3, 2, 1. So look ahead for
		// the "previousSiteConfig", but also only if we're not at the end of the slice yet.
		//
		// "previousSiteConfig" for the last item in "history" will be nil and that is okay, because
		// we will truncate it from the end result being returned. The user did not request this.
		// _We_ fetched an extra item to determine the "previousSiteConfig" of all the items.
		//
		// Similarly, if Last is used then history is in ascending order: 1, 2, 3, 4, 5. So look
		// behind for the "previousSiteConfig", but also only if we're not at the start of the
		// slice.
		//
		// "previousSiteConfig" will be nil for the first item in history in this case and that is
		// okay, because we will truncate it from the end result being returned. The user did not
		// request this. _We_ fetched an extra item to determine the "previousSiteConfig" of all the
		// items.
		if paginationArgs.First != nil && i != totalFetched-1 {
			previousSiteConfig = history[i+1]
		} else if paginationArgs.Last != nil && i > 0 {
			previousSiteConfig = history[i-1]
		}

		resolvers = append(resolvers, &SiteConfigurationChangeResolver{
			db:                 s.db,
			siteConfig:         history[i],
			previousSiteConfig: previousSiteConfig,
		})
	}

	if isModifiedPaginationArgs {
		if paginationArgs.Last != nil {
			resolvers = resolvers[1:]
		} else if paginationArgs.First != nil && totalFetched == *paginationArgs.First {
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

func (s *SiteConfigurationChangeConnectionStore) MarshalCursor(node *SiteConfigurationChangeResolver) (*string, error) {
	cursor := string(node.ID())
	return &cursor, nil
}

func (s *SiteConfigurationChangeConnectionStore) UnmarshalCursor(cursor string) (*int, error) {
	var id int
	err := relay.UnmarshalSpec(graphql.ID(cursor), &id)
	if err != nil {
		return nil, err
	}

	return &id, err
}
