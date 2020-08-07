package postgres

import (
	"database/sql"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

type batchWriter struct {
	m   sync.RWMutex
	err error
	wg  sync.WaitGroup
	ch  chan struct {
		query string
		args  []interface{}
	}
	tx *sql.Tx
}

func newBatchWriter() *batchWriter {
	tx, err := dbconn.Global.Begin()
	if err != nil {
		panic(err.Error())
	}

	w := &batchWriter{
		tx: tx,
		ch: make(chan struct {
			query string
			args  []interface{}
		}),
	}

	for i := 0; i < 100; i++ {
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()

			for v := range w.ch {
				if _, err := w.tx.Exec(v.query, v.args...); err != nil {
					w.m.Lock()
					w.err = err
					w.m.Unlock()
				}
			}
		}()
	}

	return w
}

func (w *batchWriter) Write(query string, args ...interface{}) {
	w.ch <- struct {
		query string
		args  []interface{}
	}{query, args}
}

func (w *batchWriter) Flush() error {
	close(w.ch)
	w.wg.Wait()
	if w.err != nil {
		_ = w.tx.Rollback()
		return w.err
	}

	if err := w.tx.Commit(); err != nil {
		return err
	}

	return nil
}
