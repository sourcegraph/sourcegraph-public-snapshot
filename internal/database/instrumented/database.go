package instrumented

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

//go:embed frontend
var frontend embed.FS

func NewInstrumentedDB(inner dbutil.DB) dbutil.DB {
	db := instrumentedDB{
		inner:  inner,
		logger: log.Scoped("DB", "queries"),
	}
	db.queryCh = db.listen()

	err := db.dashboard()
	if err != nil {
		db.logger.Warn("cannot start dashboard server", log.Error(err))
	}

	if _, ok := inner.(dbutil.TxBeginner); ok {
		return &transactableInstrumentedDB{db}
	}
	return &db
}

type query struct {
	query   string
	args    []any
	err     error
	time    time.Duration
	explain []string
}

type instrumentedDB struct {
	inner   dbutil.DB
	logger  log.Logger
	queryCh chan<- query

	lock    sync.RWMutex
	queries []query
}

type transactableInstrumentedDB struct {
	instrumentedDB
}

var (
	_ dbutil.DB         = &instrumentedDB{}
	_ dbutil.DB         = &transactableInstrumentedDB{}
	_ dbutil.TxBeginner = &transactableInstrumentedDB{}
)

func (db *instrumentedDB) QueryContext(ctx context.Context, q string, args ...any) (*sql.Rows, error) {
	before := time.Now()
	rows, err := db.inner.QueryContext(ctx, q, args...)
	time := time.Now().Sub(before)

	go func() {
		db.queryCh <- query{
			query: q,
			args:  args,
			err:   err,
			time:  time,
		}
	}()

	return rows, err
}

func (db *instrumentedDB) ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error) {
	before := time.Now()
	result, err := db.inner.ExecContext(ctx, q, args...)
	time := time.Now().Sub(before)

	go func() {
		db.queryCh <- query{
			query: q,
			args:  args,
			err:   err,
			time:  time,
		}
	}()

	return result, err
}

func (db *instrumentedDB) QueryRowContext(ctx context.Context, q string, args ...any) *sql.Row {
	before := time.Now()
	row := db.inner.QueryRowContext(ctx, q, args...)
	time := time.Now().Sub(before)

	go func() {
		db.queryCh <- query{
			query: q,
			args:  args,
			err:   nil,
			time:  time,
		}
	}()

	return row
}

func (db *transactableInstrumentedDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return db.inner.(dbutil.TxBeginner).BeginTx(ctx, opts)
}

func (db *instrumentedDB) listen() chan<- query {
	c := make(chan query)
	go func() {
		eligibilityRegex := regexp.MustCompile(`(?i)\bSELECT\b`)

		for q := range c {
			go func(q query) {
				ctx := context.Background()

				if eligibilityRegex.MatchString(q.query) {
					rows, err := db.inner.QueryContext(ctx, "EXPLAIN ANALYZE "+q.query, q.args...)
					if err != nil {
						db.logger.Warn(
							"cannot explain query",
							log.Error(err),
							log.String("query", q.query),
							log.String("args", fmt.Sprintf("%+v", q.args)),
							log.NamedError("original error", q.err),
						)
						return
					}

					for rows.Next() {
						var s string
						if err := rows.Scan(&s); err != nil {
							db.logger.Warn(
								"error getting row from explain resultset",
								log.Error(err),
							)
							return
						}

						q.explain = append(q.explain, s)
					}
				}

				db.lock.Lock()
				defer db.lock.Unlock()
				db.queries = append(db.queries, q)
			}(q)
		}
	}()

	return c
}

func (db *instrumentedDB) dashboard() error {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return err
	}

	mux := http.NewServeMux()

	writeErrResponse := func(w http.ResponseWriter, code int, msg string) {
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msg))
	}

	writeErr := func(w http.ResponseWriter, err error) {
		writeErrResponse(w, http.StatusInternalServerError, err.Error())
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		data, err := frontend.ReadFile("frontend" + path)
		if perr, ok := err.(*fs.PathError); ok {
			if perr.Err == fs.ErrNotExist {
				writeErrResponse(w, http.StatusNotFound, "not found")
			} else {
				writeErr(w, err)
			}
			return
		} else if err != nil {
			writeErr(w, err)
			return
		}

		contentType := "text/plain; charset=utf-8"
		switch filepath.Ext(path) {
		case ".css":
			contentType = "text/css; charset=utf-8"
		case ".html", ".htm":
			contentType = "text/html; charset=utf-8"
		case ".js":
			contentType = "text/javascript; charset=utf-8"
		}

		w.Header().Add("Content-Type", contentType)
		w.Write(data)
	})

	mux.HandleFunc("/queries", func(w http.ResponseWriter, r *http.Request) {
		count := 100
		if countQuery := r.URL.Query().Get("count"); countQuery != "" {
			if c, err := strconv.Atoi(countQuery); err == nil {
				count = c
			}
		}

		db.lock.RLock()
		once := sync.Once{}
		unlock := func() { once.Do(func() { db.lock.RUnlock() }) }
		defer unlock()

		type querySummary struct {
			ID       int    `json:"id"`
			Query    string `json:"query"`
			TimeNS   int64  `json:"time_ns"`
			HasError bool   `json:"has_error"`
			HasPlan  bool   `json:"has_plan"`
		}
		summaries := make([]querySummary, 0, count)
		for i := len(db.queries) - 1; i >= 0; i-- {
			summaries = append(summaries, querySummary{
				ID:       i,
				Query:    db.queries[i].query,
				TimeNS:   db.queries[i].time.Nanoseconds(),
				HasError: db.queries[i].err != nil,
				HasPlan:  db.queries[i].explain != nil,
			})
		}

		unlock()

		data, err := json.Marshal(&summaries)
		if err != nil {
			writeErr(w, err)
			return
		}

		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		w.Write(data)
	})

	mux.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		db.lock.RLock()
		defer db.lock.RUnlock()

		id := 0
		if idQuery := r.URL.Query().Get("id"); idQuery != "" {
			i, err := strconv.Atoi(idQuery)
			if err != nil {
				writeErrResponse(w, http.StatusBadRequest, "cannot parse ID")
				return
			}

			if i < 0 || i >= len(db.queries) {
				writeErrResponse(w, http.StatusNotFound, "ID not found")
				return
			}

			id = i
		}

		q := &db.queries[id]
		args := make([]string, len(q.args))
		for i := range q.args {
			args[i] = fmt.Sprintf("%+v", q.args[i])
		}

		var qerr *string
		if q.err != nil {
			err := q.err.Error()
			qerr = &err
		}

		data, err := json.Marshal(&struct {
			Query  string   `json:"query"`
			Args   []string `json:"args"`
			Err    *string  `json:"error,omitempty"`
			TimeNS int64    `json:"time"`
			Plan   []string `json:"plan"`
		}{
			Query:  q.query,
			Args:   args,
			Err:    qerr,
			TimeNS: q.time.Nanoseconds(),
			Plan:   q.explain,
		})
		if err != nil {
			writeErr(w, err)
			return
		}

		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		w.Write(data)
	})

	go func(listener net.Listener, mux *http.ServeMux) {
		db.logger.Info(fmt.Sprintf("started DB dashboard at http://localhost:%d/", listener.Addr().(*net.TCPAddr).Port))
		if err := http.Serve(listener, mux); err != nil {
			db.logger.Warn("error from DB dashboard server", log.Error(err))
		}
	}(listener, mux)

	return nil
}
