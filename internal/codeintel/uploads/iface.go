package uploads

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/store"
)

type Store interface {
	List(ctx context.Context, opts store.ListOpts) ([]Upload, error)
	DeleteUploadsWithoutRepository(ctx context.Context, now time.Time) (_ map[int]int, err error)
	DeleteIndexesWithoutRepository(ctx context.Context, now time.Time) (_ map[int]int, err error)
}
