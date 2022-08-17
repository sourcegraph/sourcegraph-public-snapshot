package backfill

import "context"

type UploadService interface {
	BackfillCommittedAtBatch(ctx context.Context, batchSize int) error
}
