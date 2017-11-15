package dbutil2

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"github.com/keegancsmith/sqlhooks"
	"github.com/lib/pq"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

var registerOnce sync.Once

// Open creates a new DB handle with the given schema by connecting to
// the database identified by dataSource (e.g., "dbname=mypgdb" or
// blank to use the PG* env vars).
//
// Open assumes that the database already exists.
func Open(dataSource string) (*sql.DB, error) {
	registerOnce.Do(func() {
		sql.Register("postgres-proxy", sqlhooks.Wrap(&pq.Driver{}, &hook{}))
	})
	db, err := sql.Open("postgres-proxy", dataSource)
	if err != nil {
		return nil, fmt.Errorf("%s (datasource=%q)", err, dataSource)
	}

	// Ensure we're in UTC.
	var tz string
	if err := db.QueryRow("SELECT current_setting('TIMEZONE')").Scan(&tz); err != nil {
		return nil, fmt.Errorf("getting DB timezone: %s", err)
	}
	if tz != "UTC" {
		return nil, fmt.Errorf("PostgresQL timezone must be UTC, but it is set to %q. (Set it by specifying `timezone = 'UTC'` in postgresql.conf and then restart PostgreSQL.)", tz)
	}
	return db, nil
}

// IsAlreadyExistsError returns true if err is a PostgreSQL error that
// something "already exists" (such as a table).
func IsAlreadyExistsError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "already exists")
}

type hook struct{}

// Before implements sqlhooks.Hooks
func (h *hook) Before(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	parent := opentracing.SpanFromContext(ctx)
	if parent == nil {
		return ctx, nil
	}
	span := opentracing.StartSpan("sql",
		opentracing.ChildOf(parent.Context()),
		ext.SpanKindRPCClient)
	ext.DBStatement.Set(span, query)
	ext.DBType.Set(span, "sql")
	span.LogFields(
		otlog.Object("args", args),
	)

	return opentracing.ContextWithSpan(ctx, span), nil
}

// After implements sqlhooks.Hooks
func (h *hook) After(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		span.Finish()
	}
	return ctx, nil
}
