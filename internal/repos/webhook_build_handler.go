package repos

import (
	"context"

	"github.com/sourcegraph/log"

	webhookbuilder "github.com/sourcegraph/sourcegraph/internal/repos/worker"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type webhookBuildHandler struct {
	store Store
}

func (w *webhookBuildHandler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) error {
	wbj, ok := record.(*webhookbuilder.Job)
	if !ok {
		return errors.Newf("expected Job, got %T", record)
	}

	switch wbj.ExtSvcKind {
	case "GITHUB":
		if err := handleCaseGitHub(ctx, logger, w, wbj); err != nil {
			return errors.Wrap(err, "case GitHub")
		}
	}

	return nil
}

func handleCaseGitHub(ctx context.Context, logger log.Logger, w *webhookBuildHandler, wbj *webhookbuilder.Job) error {
	// TODO

	return nil
}
