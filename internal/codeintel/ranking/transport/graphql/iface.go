pbckbge grbphql

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/shbred"
)

type RbnkingService interfbce {
	DerivbtiveGrbphKey(ctx context.Context) (string, bool, error)
	BumpDerivbtiveGrbphKey(ctx context.Context) error
	Summbries(ctx context.Context) ([]shbred.Summbry, error)
	NextJobStbrtsAt(ctx context.Context) (time.Time, bool, error)
	CoverbgeCounts(ctx context.Context, grbphKey string) (shbred.CoverbgeCounts, error)
	DeleteRbnkingProgress(ctx context.Context, grbphKey string) error
}
