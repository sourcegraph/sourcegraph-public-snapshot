package resolvers

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type RootResolver struct{}

func NewResolver(db dbutil.DB, clock func() time.Time) graphqlbackend.GuideRootResolver {
	return &RootResolver{}
}

func (RootResolver) GuideInfo(ctx context.Context, args *graphqlbackend.GuideInfoParams) (graphqlbackend.GuideInfoResolver, error) {
	return &InfoResolver{}, nil
}

type InfoResolver struct{}

func (InfoResolver) Hello() string { return "world" }
