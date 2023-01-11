package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type SiteConfigurationChangeConnectionResolver struct {
	graphqlutil.ConnectionResolverArgs
	db database.DB
}

type SiteConfigurationChangeConnectionResolverArgs struct {
}

func (r SiteConfigurationChangeConnectionResolver) compute(ctx context.Context) {
	// FIXME
}

func (r SiteConfigurationChangeConnectionResolver) Nodes(ctx context.Context) ([]*SiteConfigurationChangeResolver, error) {
	if r.First != nil && r.Last != nil {
		return nil, errors.New("Cannot use both first and last at the same time")
	}

	// FIXME
	return nil, nil
}

func (r SiteConfigurationChangeConnectionResolver) ID() graphql.ID {
	return graphql.ID("FIXME")
}

// func (*SiteConfigChangeConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
// 	// FIXME
// 	return nil, nil
// }

// func (*SiteConfigChangeConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.BidirectionalPageInfo, error) {
// 	// FIXME
// 	return nil, nil
// }

// FIXME
type SiteConfigurationChangeConnectionStore struct {
	db database.DB
}

// FIXME
func (s *SiteConfigurationChangeConnectionStore) ComputeTotal(ctx context.Context) (*int32, error) {
	return nil, nil
}

// FIXME
func (s *SiteConfigurationChangeConnectionStore) ComputeNodes(ctx context.Context, p *database.PaginationArgs) ([]*SiteConfigurationChangeResolver, error) {
	return nil, nil
}

// FIXME
func (s *SiteConfigurationChangeConnectionStore) MarshalCursor(node *SiteConfigurationChangeResolver) (*string, error) {
	return nil, nil
}

// FIXME
func (s *SiteConfigurationChangeConnectionStore) UnmarshalCursor(cursor string) (*int, error) {
	return nil, nil
}
