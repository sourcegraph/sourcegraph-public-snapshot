package commitgraph

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database/locker"
)

type Locker interface {
	Lock(ctx context.Context, key int32, blocking bool) (bool, locker.UnlockFunc, error)
}
