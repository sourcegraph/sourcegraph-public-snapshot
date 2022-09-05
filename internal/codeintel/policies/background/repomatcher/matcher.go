package repomatcher

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type matcher struct {
	dbStore DBStore
	metrics *metrics
}

var (
	_ goroutine.Handler      = &matcher{}
	_ goroutine.ErrorHandler = &matcher{}
)

func (m *matcher) Handle(ctx context.Context) error {
	if err := m.HandleRepositoryPatternMatcher(ctx); err != nil {
		return err
	}

	return nil
}

func (m *matcher) HandleError(err error) {}
