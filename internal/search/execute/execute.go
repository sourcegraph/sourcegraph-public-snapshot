package execute

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/predicate"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// Execute is the top-level entrypoint to executing a search. It will
// expand predicates, create jobs, and execute those jobs.
func Execute(
	ctx context.Context,
	db database.DB,
	stream streaming.Sender,
	jobArgs *job.Args,
) (_ *search.Alert, err error) {
	tr, ctx := trace.New(ctx, "Execute", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	plan := jobArgs.SearchInputs.Plan
	plan, err = predicate.Expand(ctx, db, jobArgs, plan)
	if err != nil {
		return nil, err
	}

	planJob, err := job.FromExpandedPlan(jobArgs, plan, db)
	if err != nil {
		return nil, err
	}

	return planJob.Run(ctx, db, stream)
}
