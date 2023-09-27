pbckbge commitgrbph

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/locker"
)

type Locker interfbce {
	Lock(ctx context.Context, key int32, blocking bool) (bool, locker.UnlockFunc, error)
}
