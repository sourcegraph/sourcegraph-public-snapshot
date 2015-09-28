package sqlwrapper

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/vendored/github.com/resonancelabs/go-pub/instrument"
)

func init() {
	sql.Register("traceguide-wrapper", &driverWrapper{})
}

type driverWrapper struct {
}

// See README.md for supported syntax for name.
func (d *driverWrapper) Open(name string) (driver.Conn, error) {
	idx := strings.Index(name, ":")
	if idx == -1 {
		return nil, fmt.Errorf("Missing ':' in sqlwrapper data source name")
	}
	name, dataSource := name[:idx], name[idx+1:]
	instrument.Log(instrument.Printf("Traceguide SQL driver installed for db = %q, data source = %q",
		name, dataSource))

	db, err := sql.Open(name, dataSource)
	if err != nil {
		return nil, err
	}
	// We could call db.Close() here, since we never use db except to
	// get at the underlying driver.
	dr := db.Driver()
	realConn, err := dr.Open(dataSource)
	if err != nil {
		return realConn, err
	}
	// The underlying realConn *must* satisfy a minimal interface (Prepare,
	// Begin, etc), but – per the golang sql driver docs – it may also
	// optionally implement a Queryer and/or Execer interface.
	//
	// We need to be sure that our returned wrapper driver implements precisely
	// the same combination of those interfaces or the golang sql.DB code might
	// try to Prepare() un-Preparable() statements, like "LOCK TABLES",
	// generating errors like
	//
	//  "This command is not supported in the prepared statement protocol yet".
	//
	// So we test for the two optional interfaces and build our wrapper out of
	// the Conn(Minimal|WithQuery|WithExec) components defined further below.
	_, isQueryer := realConn.(driver.Queryer)
	_, isExecer := realConn.(driver.Execer)
	switch {
	case isQueryer && isExecer:
		return &struct {
			*connMinimal
			*connWithQuery
			*connWithExec
		}{
			newConnMinimal(realConn),
			newConnWithQuery(realConn),
			newConnWithExec(realConn),
		}, nil
	case isQueryer:
		return &struct {
			*connMinimal
			*connWithQuery
		}{
			newConnMinimal(realConn),
			newConnWithQuery(realConn),
		}, nil
	case isExecer:
		return &struct {
			*connMinimal
			*connWithExec
		}{
			newConnMinimal(realConn),
			newConnWithExec(realConn),
		}, nil
	default:
		return newConnMinimal(realConn), nil
	}
}

type connBase struct {
	realConn driver.Conn
}

type connMinimal struct {
	*connBase
}

func newConnMinimal(realConn driver.Conn) *connMinimal {
	return &connMinimal{
		connBase: &connBase{realConn},
	}
}

func runInSpanForSQL(operationPrefix string, query string, args []driver.Value, f func(span instrument.ActiveSpan) error,
) error {
	parsedQuery := parseQuery(query)

	return instrument.RunInSpan(func(span instrument.ActiveSpan) error {
		span.SetOperation(operationNameForQuery(parsedQuery, operationPrefix))
		span.MergeTraceJoinIdsFromStack()
		span.Log(instrument.Printf("Executing SQL %q", parsedQuery.Query).Payload(args))
		span.AddAttribute("query_string", parsedQuery.Query)
		return f(span)
	}, instrument.OnStack)
}

func (c *connMinimal) Prepare(query string) (driver.Stmt, error) {
	realStmt, err := c.realConn.Prepare(query)
	if err != nil {
		return realStmt, err
	}

	// Parse the query
	parsedQuery := parseQuery(query)
	if err != nil {
		// TODO log for now... something better to do?
		instrument.Log(instrument.Printf("Unable to parse SQL statement: %q (%v)", query, err))
	}

	return &stmt{
		realStmt:    realStmt,
		parsedQuery: parsedQuery,
	}, nil
}

func (c *connMinimal) Close() error {
	return c.realConn.Close()
}

func (c *connMinimal) Begin() (driver.Tx, error) {
	span := instrument.StartSpan()
	span.SetOperation("sqlwrapper/conn/transaction")
	realTx, err := c.realConn.Begin()
	if err != nil {
		return realTx, err
	}
	return &tx{
		realTx: realTx,
		span:   span,
	}, nil
}

type connWithExec struct {
	*connBase
}

func newConnWithExec(realConn driver.Conn) *connWithExec {
	return &connWithExec{
		connBase: &connBase{realConn},
	}
}

func (c *connWithExec) Exec(query string, args []driver.Value) (driver.Result, error) {
	execer, ok := c.realConn.(driver.Execer)
	if !ok {
		panic(fmt.Errorf("connWithExec should never wrap a driver.Conn that's not a driver.Execer. realConn type: %v", reflect.TypeOf(c.realConn)))
	}

	var res driver.Result
	// Keep track of ErrSkip separately from other errors: we don't want
	// to log this as a true Error, so we return nil as the result of
	// the Span, but then return ErrSkip as the result of Exec.
	isErrSkip := false
	err := runInSpanForSQL("sqlwrapper/conn/exec", query, args,
		func(span instrument.ActiveSpan) (err error) {
			res, err = execer.Exec(query, args)
			if err == nil {
				rows, err2 := res.RowsAffected()
				if err2 != nil {
					span.Log(instrument.Printf("Unable to get number of rows affected: %v", err2))
				} else {
					span.Log(instrument.Printf("%d rows affected", rows))
				}
			} else if err == driver.ErrSkip {
				span.Log("Go SQL driver does not support connection-level Exec; caller will likely fall back on Statement-level Exec.")
				span.AddAttribute("noop", "true")
				isErrSkip = true
				err = nil
			}
			return err
		})
	if isErrSkip {
		err = driver.ErrSkip
	}
	return res, err
}

type connWithQuery struct {
	*connBase
}

func newConnWithQuery(realConn driver.Conn) *connWithQuery {
	return &connWithQuery{
		connBase: &connBase{realConn},
	}
}

type rows struct {
	realRows driver.Rows
	span     instrument.ActiveSpan
}

func newRows(realRows driver.Rows) *rows {
	span := instrument.StartSpan()
	span.SetOperation("sqlwrapper/stream_rows")
	span.MergeTraceJoinIdsFromStack()
	return &rows{
		realRows: realRows,
		span:     span,
	}
}

func (r *rows) Columns() []string {
	return r.realRows.Columns()
}

func (r *rows) Close() error {
	defer r.span.Finish()
	return r.realRows.Close()
}

func (r *rows) Next(dest []driver.Value) error {
	return r.realRows.Next(dest)
}

func (c *connWithQuery) Query(query string, args []driver.Value) (driver.Rows, error) {
	queryer, ok := c.realConn.(driver.Queryer)
	if !ok {
		panic(fmt.Errorf("connWithQuery should never wrap a driver.Conn that's not a driver.Queryer. realConn type: %v", reflect.TypeOf(c.realConn)))
	}

	var rs driver.Rows
	// See comment at isErrSkip above
	isErrSkip := false
	err := runInSpanForSQL("sqlwrapper/conn/query", query, args,
		func(span instrument.ActiveSpan) (err error) {
			rs, err = queryer.Query(query, args)
			if err == driver.ErrSkip {
				span.Log("Go SQL driver does not support connection-level Query; caller will likely fall back on Statement-level Query.")
				span.AddAttribute("noop", "true")
				isErrSkip = true
				err = nil
			}
			return err
		})
	if isErrSkip {
		err = driver.ErrSkip
	}
	return newRows(rs), err
}

type stmt struct {
	realStmt    driver.Stmt
	parsedQuery *parsedQuery
}

func (s *stmt) Close() error {
	return s.realStmt.Close()
}

func (s *stmt) NumInput() int {
	return s.realStmt.NumInput()
}

func operationNameForQuery(pq *parsedQuery, operationPrefix string) string {
	sqlVerb := "unknown"
	switch pq.QueryType {
	case Select:
		sqlVerb = "select"
	case Insert:
		sqlVerb = "insert"
	case Update:
		sqlVerb = "update"
	case Delete:
		sqlVerb = "delete"
	default:
		// stick with "unknown"
	}
	return fmt.Sprintf("%s/%s", operationPrefix, sqlVerb)
}

func (s *stmt) operationName(operationPrefix string) string {
	return operationNameForQuery(s.parsedQuery, operationPrefix)
}

func (s *stmt) Exec(args []driver.Value) (driver.Result, error) {
	var res driver.Result
	err := instrument.RunInSpan(func(span instrument.ActiveSpan) (err error) {
		span.SetOperation(s.operationName("sqlwrapper/statement/exec"))
		span.MergeTraceJoinIdsFromStack()
		span.Log(instrument.Printf("Executing statement %q", s.parsedQuery.Query).Payload(args))
		span.AddAttribute("query_string", s.parsedQuery.Query)
		res, err = s.realStmt.Exec(args)
		if err == nil {
			rows, err2 := res.RowsAffected()
			if err2 != nil {
				span.Log(instrument.Printf("Unable to get number of rows affected: %v", err2))
			} else {
				span.Log(instrument.Printf("%d rows affected", rows))
			}
		}
		return err
	}, instrument.OnStack)
	return res, err
}

func (s *stmt) Query(args []driver.Value) (driver.Rows, error) {
	var rs driver.Rows
	err := instrument.RunInSpan(func(span instrument.ActiveSpan) (err error) {
		span.SetOperation(s.operationName("sqlwrapper/statement/query"))
		// Ignore the error here: we try to get join ids from parents, but
		// if there are none, there's nothing we can do about it.
		_ = span.MergeTraceJoinIdsFromStack()
		span.Log(instrument.Printf("Executing query %q", s.parsedQuery.Query).Payload(args))
		span.AddAttribute("query_string", s.parsedQuery.Query)
		rs, err = s.realStmt.Query(args)
		return err
	}, instrument.OnStack)
	return newRows(rs), err
}

type tx struct {
	realTx driver.Tx
	span   instrument.ActiveSpan
}

func (t *tx) Commit() error {
	t.span.Log("commiting transaction")
	err := t.realTx.Commit()
	if err != nil {
		t.span.Log(instrument.Print(err).Error())
	}
	t.span.Finish()
	return err
}

func (t *tx) Rollback() error {
	t.span.Log("rolling back transaction")
	err := t.realTx.Rollback()
	if err != nil {
		t.span.Log(instrument.Print(err).Error())
	}
	t.span.Finish()
	return err
}
