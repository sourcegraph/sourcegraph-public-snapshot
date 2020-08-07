package scylladb

import (
	"sync"

	"github.com/gocql/gocql"
)

var session *gocql.Session

func init() {
	cluster := gocql.NewCluster("localhost")
	cluster.Keyspace = "lsif"
	cluster.Consistency = gocql.One

	var err error
	if session, err = cluster.CreateSession(); err != nil {
		panic(err.Error())
	}
}

//
//

type batchWriter struct {
	m   sync.RWMutex
	err error
	wg  sync.WaitGroup
	ch  chan struct {
		query string
		args  []interface{}
	}
}

func newBatchWriter() *batchWriter {
	w := &batchWriter{
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
				if err := session.Query(v.query, v.args...).Exec(); err != nil {
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
	return w.err
}
