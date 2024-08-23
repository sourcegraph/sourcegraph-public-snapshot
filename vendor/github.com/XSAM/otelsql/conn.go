// Copyright Sam Xie
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package otelsql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"

	"go.opentelemetry.io/otel/trace"
)

var (
	_ driver.Pinger             = (*otConn)(nil)
	_ driver.Execer             = (*otConn)(nil) // nolint
	_ driver.ExecerContext      = (*otConn)(nil)
	_ driver.Queryer            = (*otConn)(nil) // nolint
	_ driver.QueryerContext     = (*otConn)(nil)
	_ driver.Conn               = (*otConn)(nil)
	_ driver.ConnPrepareContext = (*otConn)(nil)
	_ driver.ConnBeginTx        = (*otConn)(nil)
	_ driver.SessionResetter    = (*otConn)(nil)
	_ driver.NamedValueChecker  = (*otConn)(nil)
)

type otConn struct {
	driver.Conn
	cfg config
}

func newConn(conn driver.Conn, cfg config) *otConn {
	return &otConn{
		Conn: conn,
		cfg:  cfg,
	}
}

func (c *otConn) Ping(ctx context.Context) (err error) {
	pinger, ok := c.Conn.(driver.Pinger)
	if !ok {
		// Driver doesn't implement, nothing to do
		return nil
	}

	method := MethodConnPing
	onDefer := recordMetric(ctx, c.cfg.Instruments, c.cfg.Attributes, method)
	defer func() {
		onDefer(err)
	}()

	if c.cfg.SpanOptions.Ping {
		if filterSpan(ctx, c.cfg.SpanOptions, method, "", nil) {
			var span trace.Span
			ctx, span = createSpan(ctx, c.cfg, method, false, "", nil)
			defer func() {
				if err != nil {
					recordSpanError(span, c.cfg.SpanOptions, err)
				}
				span.End()
			}()
		}
	}

	err = pinger.Ping(ctx)
	return err
}

func (c *otConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	execer, ok := c.Conn.(driver.Execer) // nolint
	if !ok {
		return nil, driver.ErrSkip
	}
	return execer.Exec(query, args)
}

func (c *otConn) ExecContext(
	ctx context.Context, query string, args []driver.NamedValue,
) (res driver.Result, err error) {
	execer, ok := c.Conn.(driver.ExecerContext)
	if !ok {
		return nil, driver.ErrSkip
	}

	method := MethodConnExec
	onDefer := recordMetric(ctx, c.cfg.Instruments, c.cfg.Attributes, method)
	defer func() {
		onDefer(err)
	}()

	var span trace.Span
	if filterSpan(ctx, c.cfg.SpanOptions, method, query, args) {
		ctx, span = createSpan(ctx, c.cfg, method, true, query, args)
		defer span.End()
	}

	res, err = execer.ExecContext(ctx, c.cfg.SQLCommenter.withComment(ctx, query), args)
	if err != nil {
		recordSpanError(span, c.cfg.SpanOptions, err)
		return nil, err
	}
	return res, nil
}

func (c *otConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	queryer, ok := c.Conn.(driver.Queryer) // nolint
	if !ok {
		return nil, driver.ErrSkip
	}
	return queryer.Query(query, args)
}

func (c *otConn) QueryContext(
	ctx context.Context, query string, args []driver.NamedValue,
) (rows driver.Rows, err error) {
	queryer, ok := c.Conn.(driver.QueryerContext)
	if !ok {
		return nil, driver.ErrSkip
	}

	method := MethodConnQuery
	onDefer := recordMetric(ctx, c.cfg.Instruments, c.cfg.Attributes, method)
	defer func() {
		onDefer(err)
	}()

	var span trace.Span
	queryCtx := ctx
	if !c.cfg.SpanOptions.OmitConnQuery && filterSpan(ctx, c.cfg.SpanOptions, method, query, args) {
		queryCtx, span = createSpan(ctx, c.cfg, method, true, query, args)
		defer span.End()
	}

	rows, err = queryer.QueryContext(queryCtx, c.cfg.SQLCommenter.withComment(queryCtx, query), args)
	if err != nil {
		recordSpanError(span, c.cfg.SpanOptions, err)
		return nil, err
	}
	return newRows(ctx, rows, c.cfg), nil
}

func (c *otConn) PrepareContext(ctx context.Context, query string) (stmt driver.Stmt, err error) {
	method := MethodConnPrepare
	onDefer := recordMetric(ctx, c.cfg.Instruments, c.cfg.Attributes, method)
	defer func() {
		onDefer(err)
	}()

	var span trace.Span
	if !c.cfg.SpanOptions.OmitConnPrepare && filterSpan(ctx, c.cfg.SpanOptions, method, query, nil) {
		ctx, span = createSpan(ctx, c.cfg, method, true, query, nil)
		defer span.End()
		defer recordSpanErrorDeferred(span, c.cfg.SpanOptions, &err)
	}

	commentedQuery := c.cfg.SQLCommenter.withComment(ctx, query)

	if preparer, ok := c.Conn.(driver.ConnPrepareContext); ok {
		if stmt, err = preparer.PrepareContext(ctx, commentedQuery); err != nil {
			return nil, err
		}
	} else {
		if stmt, err = c.Conn.Prepare(commentedQuery); err != nil {
			return nil, err
		}

		select {
		default:
		case <-ctx.Done():
			stmt.Close()
			return nil, ctx.Err()
		}
	}

	return newStmt(stmt, c.cfg, query), nil
}

func (c *otConn) BeginTx(ctx context.Context, opts driver.TxOptions) (tx driver.Tx, err error) {
	method := MethodConnBeginTx
	onDefer := recordMetric(ctx, c.cfg.Instruments, c.cfg.Attributes, method)
	defer func() {
		onDefer(err)
	}()

	var beginTxCtx context.Context
	if filterSpan(ctx, c.cfg.SpanOptions, method, "", nil) {
		var span trace.Span
		beginTxCtx, span = createSpan(ctx, c.cfg, method, false, "", nil)
		defer span.End()
		defer recordSpanErrorDeferred(span, c.cfg.SpanOptions, &err)
	} else {
		beginTxCtx = ctx
	}

	if connBeginTx, ok := c.Conn.(driver.ConnBeginTx); ok {
		if tx, err = connBeginTx.BeginTx(beginTxCtx, opts); err != nil {
			return nil, err
		}
	} else {
		// Code borrowed from ctxutil.go in the go standard library.
		// Check the transaction level. If the transaction level is non-default
		// then return an error here as the BeginTx driver value is not supported.
		if opts.Isolation != driver.IsolationLevel(sql.LevelDefault) {
			return nil, errors.New("sql: driver does not support non-default isolation level")
		}

		// If a read-only transaction is requested return an error as the
		// BeginTx driver value is not supported.
		if opts.ReadOnly {
			return nil, errors.New("sql: driver does not support read-only transactions")
		}

		if tx, err = c.Conn.Begin(); err != nil { //nolint:staticcheck
			return nil, err
		}

		if ctx.Done() != nil {
			select {
			default:
			case <-ctx.Done():
				_ = tx.Rollback()
				return nil, ctx.Err()
			}
		}
	}
	return newTx(ctx, tx, c.cfg), nil
}

func (c *otConn) ResetSession(ctx context.Context) (err error) {
	sessionResetter, ok := c.Conn.(driver.SessionResetter)
	if !ok {
		// Driver does not implement, there is nothing to do.
		return nil
	}

	method := MethodConnResetSession
	onDefer := recordMetric(ctx, c.cfg.Instruments, c.cfg.Attributes, method)
	defer func() {
		onDefer(err)
	}()

	var span trace.Span
	if !c.cfg.SpanOptions.OmitConnResetSession && filterSpan(ctx, c.cfg.SpanOptions, method, "", nil) {
		ctx, span = createSpan(ctx, c.cfg, method, false, "", nil)
		defer span.End()
	}

	err = sessionResetter.ResetSession(ctx)
	if err != nil {
		recordSpanError(span, c.cfg.SpanOptions, err)
		return err
	}
	return nil
}

func (c *otConn) CheckNamedValue(namedValue *driver.NamedValue) error {
	namedValueChecker, ok := c.Conn.(driver.NamedValueChecker)
	if !ok {
		return driver.ErrSkip
	}

	return namedValueChecker.CheckNamedValue(namedValue)
}

// Raw returns the underlying driver connection
// Issue: https://github.com/XSAM/otelsql/issues/98
func (c *otConn) Raw() driver.Conn {
	return c.Conn
}
