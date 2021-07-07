package resolvers

import (
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type Resolver struct {
}

func NewResolver(db dbutil.DB, clock func() time.Time) graphqlbackend.CodeGraphResolver {
	return &Resolver{}
}
