package background

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type unknownJobTypeErr struct {
	jobType string
}

func (e unknownJobTypeErr) Error() string {
	return fmt.Sprintf("invalid job type %q", e.jobType)
}

func (e unknownJobTypeErr) NonRetryable() bool {
	return true
}

type bulkProcessor struct {
	store   *store.Store
	sourcer repos.Sourcer
}

func (b *bulkProcessor) HandlerFunc() dbworker.HandlerFunc {
	return func(ctx context.Context, tx dbworkerstore.Store, record workerutil.Record) error {
		return b.process(ctx, b.store.With(tx), record.(*store.ChangesetJob))
	}
}

func (b *bulkProcessor) process(ctx context.Context, tx *store.Store, job *store.ChangesetJob) error {
	switch job.JobType {
	case store.ChangesetJobTypeComment:
		return b.comment(ctx, tx, job)
	default:
		return &unknownJobTypeErr{jobType: string(job.JobType)}
	}
}

func (b *bulkProcessor) comment(ctx context.Context, tx *store.Store, job *store.ChangesetJob) error {

	return nil
}
