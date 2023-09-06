package search

import (
	"bytes"
	"context"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/store"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/types"
)

func TestExhaustiveSearch(t *testing.T) {
	// This test exercises the full worker infra from the time a search job is
	// created until it is done.

	require := require.New(t)
	observationCtx := observation.TestContextTB(t)
	logger := observationCtx.Logger
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	s := store.New(db, observation.TestContextTB(t))

	ctx, cancel := context.WithCancel(actor.WithInternalActor(context.Background()))
	defer cancel()

	userID := insertRow(t, s.Store, "users", "username", "alice")
	insertRow(t, s.Store, "repo", "id", 1, "name", "repoa")
	insertRow(t, s.Store, "repo", "id", 2, "name", "repob")

	query := "1@rev1 1@rev2 2@rev3"

	// Create a job
	_, err := s.CreateExhaustiveSearchJob(ctx, types.ExhaustiveSearchJob{
		InitiatorID: userID,
		Query:       query,
	})
	require.NoError(err)

	// Now that the job is created, we start up all the worker routines for
	// exhaustive search and wait until there are no more jobs left.
	searchJob := &searchJob{
		workerDB: db,
		config: config{
			WorkerInterval: 10 * time.Millisecond,
		},
	}

	csvBuf = &concurrentWriter{writer: &bytes.Buffer{}}
	routines, err := searchJob.Routines(ctx, observationCtx)
	require.NoError(err)
	for _, routine := range routines {
		go routine.Start()
		defer routine.Stop()
	}
	require.Eventually(func() bool {
		return !searchJob.hasWork(ctx)
	}, tTimeout(t, 10*time.Second), 10*time.Millisecond)

	require.Equal([][]string{
		{
			"repo,revspec,revision",
			"1,spec,rev1",
		},
		{
			"repo,revspec,revision",
			"1,spec,rev2",
		},
		{
			"repo,revspec,revision",
			"2,spec,rev3",
		},
	}, parseCSV(csvBuf.(*concurrentWriter).String()))
}

func parseCSV(csv string) (o [][]string) {
	rows := strings.Split(csv, "\n")
	for i := 0; i < len(rows)-1; i += 2 {
		o = append(o, []string{rows[i], rows[i+1]})
	}
	sort.Sort(byRow(o))
	return
}

type byRow [][]string

func (b byRow) Len() int {
	return len(b)
}

func (b byRow) Less(i, j int) bool {
	return b[i][1] < b[j][1]
}

func (b byRow) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

// insertRow is a helper for inserting a row into a table. It assumes the
// table has an autogenerated column called id and it will return that value.
func insertRow(t testing.TB, store *basestore.Store, table string, keyValues ...any) int32 {
	var columns, values []*sqlf.Query
	for i, kv := range keyValues {
		if i%2 == 0 {
			columns = append(columns, sqlf.Sprintf(kv.(string)))
		} else {
			values = append(values, sqlf.Sprintf("%v", kv))
		}
	}
	q := sqlf.Sprintf(`INSERT INTO %s(%s) VALUES(%s) RETURNING id`, sqlf.Sprintf(table), sqlf.Join(columns, ", "), sqlf.Join(values, ", "))
	row := store.QueryRow(context.Background(), q)
	var id int32
	if err := row.Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

// tTimeout returns the duration until t's deadline. If there is no deadline
// or the deadline is further away than max, then max is returned.
func tTimeout(t *testing.T, max time.Duration) time.Duration {
	deadline, ok := t.Deadline()
	if !ok {
		return max
	}
	timeout := time.Until(deadline)
	if max < timeout {
		return max
	}
	return timeout
}

type concurrentWriter struct {
	mu     sync.Mutex
	writer *bytes.Buffer
}

func (w *concurrentWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writer.String()
}

func (w *concurrentWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writer.Write(p)
}
