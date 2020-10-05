package graphs

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/graphs"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

func Init(ctx context.Context, enterpriseServices *enterprise.Services) error {
	graphsStore := graphs.NewStoreWithClock(dbconn.Global, msResolutionClock)
	enterpriseServices.GraphsResolver = graphs.NewResolver(graphsStore)
	return nil
}

var msResolutionClock = func() time.Time { return time.Now().UTC().Truncate(time.Microsecond) }
