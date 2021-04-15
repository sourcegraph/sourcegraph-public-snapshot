package background

import (
	"context"
	"fmt"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// unknownJobTypeErr is returned when a ChangesetJob record is of an unknown type
// and hence cannot be executed.
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
	sourcer sources.Sourcer

	css  sources.ChangesetSource
	repo *types.Repo
	ch   *btypes.Changeset
}

func (b *bulkProcessor) process(ctx context.Context, job *btypes.ChangesetJob) (err error) {
	// Load changeset.
	b.ch, err = b.store.GetChangeset(ctx, store.GetChangesetOpts{ID: job.ChangesetID})
	if err != nil {
		return errors.Wrap(err, "loading changeset")
	}

	// Load repo.
	b.repo, err = b.store.Repos().Get(ctx, b.ch.RepoID)
	if err != nil {
		return errors.Wrap(err, "loading repo")
	}

	// Construct changeset source.
	b.css, err = b.sourcer.ForRepo(ctx, b.store, b.repo)
	if err != nil {
		return errors.Wrap(err, "loading ChangesetSource")
	}
	b.css, err = sources.WithAuthenticatorForUser(ctx, b.store, b.css, job.UserID, b.repo)
	if err != nil {
		return errors.Wrap(err, "authenticating ChangesetSource")
	}

	log15.Info("processing changeset job", "type", job.JobType)

	switch job.JobType {

	case btypes.ChangesetJobTypeComment:
		return b.comment(ctx, job)

	default:
		return &unknownJobTypeErr{jobType: string(job.JobType)}
	}
}

func (b *bulkProcessor) comment(ctx context.Context, job *btypes.ChangesetJob) error {
	typedPayload, ok := job.Payload.(*btypes.ChangesetJobCommentPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for changeset_job, want=%T have=%T", &btypes.ChangesetJobCommentPayload{}, job.Payload)
	}
	cs := &sources.Changeset{
		Changeset: b.ch,
		Repo:      b.repo,
	}
	return b.css.CreateComment(ctx, cs, typedPayload.Message)
}
