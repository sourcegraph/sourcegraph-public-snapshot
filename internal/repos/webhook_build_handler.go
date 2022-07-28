package repos

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type webhookBuildHandler struct {
	store Store
}

func newWebHookBuildHandler(store Store) *webhookBuildHandler {
	return &webhookBuildHandler{store: store}
}

func (w *webhookBuildHandler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) error {
	// TODO
	return nil
}
