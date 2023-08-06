package dbclient

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type InvalidExecuteError struct{}

func (e *InvalidExecuteError) Error() string {
	return "oh noes"
}

type ExecutableQuery interface {
	Execute(context.Context, *basestore.Store) (any, error)
}

type DBClient interface {
	DB() database.DB
	Execute(context.Context, ExecutableQuery) (any, error)
}

type dbClient struct {
	*basestore.Store
	db database.DB
}

func (c *dbClient) DB() database.DB {
	return c.db
}

func NewDBClient(db database.DB) DBClient {
	return &dbClient{
		Store: basestore.NewWithHandle(db.Handle()),
		db:    db,
	}
}

func Unmarshal[T any](data any) (t T, err error) {
	return data.(T), nil
}

func (c *dbClient) Execute(ctx context.Context, query ExecutableQuery) (any, error) {
	return query.Execute(ctx, c.Store)
}
