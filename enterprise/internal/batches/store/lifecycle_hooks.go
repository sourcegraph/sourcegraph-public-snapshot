package store

import (
	"context"

	"github.com/keegancsmith/sqlf"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type CountLifecycleHooksOpts struct {
	IncludeExpired bool
}

func (s *Store) CountLifecycleHooks(ctx context.Context, opts CountLifecycleHooksOpts) (count int, err error) {
	ctx, endObservation := s.operations.countLifecycleHooks.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := countLifecycleHooksQuery(&opts)
	return s.queryCount(ctx, q)
}

func (s *Store) CreateLifecycleHook(ctx context.Context, hook *btypes.LifecycleHook) (err error) {
	ctx, endObservation := s.operations.createLifecycleHook.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	if hook.CreatedAt.IsZero() {
		hook.CreatedAt = s.now()
	}
	if hook.UpdatedAt.IsZero() {
		hook.UpdatedAt = s.now()
	}

	q := createLifecycleHookQuery(hook)
	return s.query(ctx, q, func(sc dbutil.Scanner) error {
		return scanLifecycleHook(hook, sc)
	})
}

func (s *Store) DeleteLifecycleHook(ctx context.Context, id int64) (err error) {
	ctx, endObservation := s.operations.deleteLifecycleHook.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	res, err := s.ExecResult(ctx, deleteLifecycleHookQuery(id))
	if err != nil {
		return err
	}

	if rows, err := res.RowsAffected(); err != nil {
		return err
	} else if rows == 0 {
		return ErrNoResults
	}
	return nil
}

func (s *Store) GetLifecycleHook(ctx context.Context, id int64) (hook *btypes.LifecycleHook, err error) {
	ctx, endObservation := s.operations.getLifecycleHook.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	hook = &btypes.LifecycleHook{}
	q := getLifecycleHookQuery(id)
	if err := s.query(ctx, q, func(sc dbutil.Scanner) error {
		return scanLifecycleHook(hook, sc)
	}); err != nil {
		return nil, err
	}

	if hook.ID == 0 {
		return nil, ErrNoResults
	}

	return hook, err
}

type ListLifecycleHookOpts struct {
	LimitOpts
	Cursor         int64
	IncludeExpired bool
}

func (s *Store) ListLifecycleHooks(ctx context.Context, opts ListLifecycleHookOpts) (hooks []*btypes.LifecycleHook, next int64, err error) {
	ctx, endObservation := s.operations.listLifecycleHooks.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	hooks = make([]*btypes.LifecycleHook, 0, opts.DBLimit())
	q := listLifecycleHooksQuery(&opts)
	if err := s.query(ctx, q, func(sc dbutil.Scanner) error {
		hook := btypes.LifecycleHook{}
		if err := scanLifecycleHook(&hook, sc); err != nil {
			return err
		}

		hooks = append(hooks, &hook)
		return nil
	}); err != nil {
		return nil, 0, err
	}

	if opts.Limit != 0 && len(hooks) == opts.DBLimit() {
		next = hooks[len(hooks)-1].ID
		hooks = hooks[:len(hooks)-1]
	}

	return hooks, next, err
}

const countLifecycleHooksQueryFmtstr = `
-- source: enterprise/internal/batches/store/lifecycle_hooks.go:CountLifecycleHook

SELECT
	COUNT(id)
FROM
	batch_changes_lifecycle_hooks
WHERE
	%s
`

func countLifecycleHooksQuery(opts *CountLifecycleHooksOpts) *sqlf.Query {
	preds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if !opts.IncludeExpired {
		preds = append(preds, sqlf.Sprintf("expires_at IS NULL"))
	}

	return sqlf.Sprintf(
		countLifecycleHooksQueryFmtstr,
		sqlf.Join(preds, " AND "),
	)
}

const createLifecycleHookQueryFmtstr = `
-- source: enterprise/internal/batches/store/lifecycle_hooks.go:CreateLifecycleHook

INSERT INTO batch_changes_lifecycle_hooks (
	created_at,
	updated_at,
	expires_at,
	url,
	secret
)
VALUES
	(%s, %s, %s, %s, %s)
RETURNING
	%s
`

func createLifecycleHookQuery(hook *btypes.LifecycleHook) *sqlf.Query {
	return sqlf.Sprintf(
		createLifecycleHookQueryFmtstr,
		hook.CreatedAt,
		hook.UpdatedAt,
		nullTimeColumn(hook.ExpiresAt),
		hook.URL,
		hook.Secret,
		sqlf.Join(lifecycleHookColumns, ", "),
	)
}

const deleteLifecycleHookQueryFmtstr = `
-- source: enterprise/internal/batches/store/lifecycle_hooks.go:DeleteLifecycleHook

DELETE FROM
	batch_changes_lifecycle_hooks
WHERE
	id = %s
`

func deleteLifecycleHookQuery(id int64) *sqlf.Query {
	return sqlf.Sprintf(
		deleteLifecycleHookQueryFmtstr,
		id,
	)
}

const getLifecycleHookQueryFmtstr = `
-- source: enterprise/internal/batches/store/lifecycle_hooks.go:GetLifecycleHook

SELECT
	%s
FROM
	batch_changes_lifecycle_hooks
WHERE
	id = %s
`

func getLifecycleHookQuery(id int64) *sqlf.Query {
	return sqlf.Sprintf(
		getLifecycleHookQueryFmtstr,
		sqlf.Join(lifecycleHookColumns, ", "),
		id,
	)
}

const listLifecycleHooksQueryFmtstr = `
-- source: enterprise/internal/batches/store/lifecycle_hooks.go:ListLifecycleHooks

SELECT
	%s
FROM
	batch_changes_lifecycle_hooks
WHERE
	%s
ORDER BY
	id ASC
`

func listLifecycleHooksQuery(opts *ListLifecycleHookOpts) *sqlf.Query {
	preds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if opts.Cursor != 0 {
		preds = append(preds, sqlf.Sprintf("id >= %s", opts.Cursor))
	}
	if !opts.IncludeExpired {
		preds = append(preds, sqlf.Sprintf("expires_at IS NULL"))
	}

	return sqlf.Sprintf(
		listLifecycleHooksQueryFmtstr+opts.ToDB(),
		sqlf.Join(lifecycleHookColumns, ", "),
		sqlf.Join(preds, " AND "),
	)
}

var lifecycleHookColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("updated_at"),
	sqlf.Sprintf("expires_at"),
	sqlf.Sprintf("url"),
	sqlf.Sprintf("secret"),
}

func scanLifecycleHook(hook *btypes.LifecycleHook, sc dbutil.Scanner) error {
	return sc.Scan(
		&hook.ID,
		&hook.CreatedAt,
		&hook.UpdatedAt,
		&dbutil.NullTime{Time: &hook.ExpiresAt},
		&hook.URL,
		&hook.Secret,
	)
}
