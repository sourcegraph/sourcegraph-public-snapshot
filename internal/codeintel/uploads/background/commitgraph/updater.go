package commitgraph

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type updater struct {
	uploadSvc  UploadService
	operations *operations
}

var (
	_ goroutine.Handler      = &updater{}
	_ goroutine.ErrorHandler = &updater{}
)

func (u *updater) Handle(ctx context.Context) error {
	if err := u.HandleUpdateDirtyRepositories(ctx); err != nil {
		return err
	}

	return nil
}

func (u *updater) HandleError(err error) {}
