package dbconn

import (
	"context"

	"github.com/qustavo/sqlhooks/v2"
)

type hookCollection struct {
	beforeHooks []sqlhooks.Hook
	afterHooks  []sqlhooks.Hook
	errorHooks  []sqlhooks.ErrorHook
}

var _ sqlhooks.Hooks = &hookCollection{}
var _ sqlhooks.OnErrorer = &hookCollection{}

func combineHooks(hooks ...sqlhooks.Hooks) sqlhooks.Hooks {
	beforeHooks := make([]sqlhooks.Hook, 0, len(hooks))
	afterHooks := make([]sqlhooks.Hook, 0, len(hooks))
	errorHooks := make([]sqlhooks.ErrorHook, 0, len(hooks))

	for _, hook := range hooks {
		beforeHooks = append(beforeHooks, hook.Before)
		afterHooks = append(afterHooks, hook.After)

		if errorHook, ok := hook.(sqlhooks.OnErrorer); ok {
			errorHooks = append(errorHooks, errorHook.OnError)
		}
	}

	return &hookCollection{
		beforeHooks: beforeHooks,
		afterHooks:  afterHooks,
		errorHooks:  errorHooks,
	}
}

func (h *hookCollection) Before(ctx context.Context, query string, args ...any) (_ context.Context, err error) {
	return runHooks(ctx, h.beforeHooks, query, args...)
}

func (h *hookCollection) After(ctx context.Context, query string, args ...any) (_ context.Context, err error) {
	return runHooks(ctx, h.afterHooks, query, args...)
}

func (h *hookCollection) OnError(ctx context.Context, err error, query string, args ...any) error {
	return runErrorHooks(ctx, h.errorHooks, err, query, args...)
}

func runHooks(ctx context.Context, hooks []sqlhooks.Hook, query string, args ...any) (_ context.Context, err error) {
	for _, hook := range hooks {
		ctx, err = hook(ctx, query, args...)
		if err != nil {
			break
		}
	}

	return ctx, err
}

func runErrorHooks(ctx context.Context, hooks []sqlhooks.ErrorHook, err error, query string, args ...any) error {
	for _, hook := range hooks {
		err = hook(ctx, err, query, args...)
	}

	return err
}
