package repos

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type webhookBuildHandler struct {
	store Store
}

func newWebhookBuildHandler(store Store) *webhookBuildHandler {
	return &webhookBuildHandler{store: store}
}

func (w *webhookBuildHandler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) error {
	return errors.New("TODO")
}
