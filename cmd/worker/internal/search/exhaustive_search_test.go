package search

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/object/mocks"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/service"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/store"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/types"
	"github.com/sourcegraph/sourcegraph/lib/iterator"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestExhaustiveSearch(t *testing.T) {
	// This test exercises the full worker infra from the time a search job is
	// created until it is done.

	enabled := true
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			ExperimentalFeatures: &schema.ExperimentalFeatures{SearchJobs: &enabled}}})
	defer conf.Mock(nil)

	require := require.New(t)
	observationCtx := observation.TestContextTB(t)
	logger := observationCtx.Logger

	mockUploadStore, bucket := newMockUploadStore(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	s := store.New(db, observation.TestContextTB(t))
	svc := service.New(observationCtx, s, mockUploadStore, service.NewSearcherFake())

	userID := insertRow(t, s.Store, "users", "username", "alice")
	userBadID := insertRow(t, s.Store, "users", "username", "mallory")
	insertRow(t, s.Store, "repo", "id", 1, "name", "repoa")
	insertRow(t, s.Store, "repo", "id", 2, "name", "repob")

	workerCtx, cancel1 := context.WithCancel(actor.WithInternalActor(context.Background()))
	defer cancel1()
	userCtx, cancel2 := context.WithCancel(actor.WithActor(context.Background(), actor.FromUser(userID)))
	defer cancel2()

	query := "1@rev1 1@rev2 2@rev3"

	// Create a job
	job, err := svc.CreateSearchJob(userCtx, query)
	require.NoError(err)

	// Do some assertions on the job before it runs
	{
		require.Equal(userID, job.InitiatorID)
		require.Equal(query, job.Query)
		require.Equal(types.JobStateQueued, job.State)
		require.NotZero(job.CreatedAt)
		require.NotZero(job.UpdatedAt)
		job2, err := svc.GetSearchJob(userCtx, job.ID)
		require.NoError(err)
		require.Equal(job, job2)
	}

	// TODO these sort of tests need to live somewhere that makes more sense.
	// But for now we have a fully functioning setup here lets test List.
	{
		jobs, err := svc.ListSearchJobs(userCtx, store.ListArgs{})
		require.NoError(err)

		require.Equal([]*types.ExhaustiveSearchJob{job}, jobs)
	}

	// Now that the job is created, we start up all the worker routines for
	// exhaustive search and wait until there are no more jobs left.
	searchJob := &searchJob{
		workerDB: db,
		config: config{
			WorkerInterval: 10 * time.Millisecond,
		},
	}

	newSearcherFactory := func(_ *observation.Context, _ database.DB) service.NewSearcher {
		return service.NewSearcherFake()
	}

	routines, err := searchJob.newSearchJobRoutines(workerCtx, observationCtx, mockUploadStore, newSearcherFactory)
	require.NoError(err)
	for _, routine := range routines {
		go routine.Start()
		defer func() {
			err := routine.Stop(context.Background())
			require.NoError(err)
		}()
	}
	require.Eventually(func() bool {
		return !searchJob.hasWork(workerCtx)
	}, tTimeout(t, 10*time.Second), 10*time.Millisecond)

	// Assert that we ended up writing the expected results. This validates
	// that somehow the work happened (but doesn't dive into the guts of how
	// we co-ordinate our workers)
	{
		var vals []string
		for _, v := range bucket {
			vals = append(vals, v)
		}
		sort.Strings(vals)
		require.Equal([]string{`{"type":"path","path":"path/to/file.go","repositoryID":1,"repository":"repo1","commit":"rev1","language":"Go"}
`, `{"type":"path","path":"path/to/file.go","repositoryID":1,"repository":"repo1","commit":"rev2","language":"Go"}
`, `{"type":"path","path":"path/to/file.go","repositoryID":2,"repository":"repo2","commit":"rev3","language":"Go"}
`}, vals)
	}

	// Minor assertion that the job is regarded as finished.
	{
		job2, err := svc.GetSearchJob(userCtx, job.ID)
		require.NoError(err)
		// Only the WorkerJob fields should change. And in that case we will
		// only assert on State since the rest are non-deterministic.
		require.Equal(types.JobStateCompleted, job2.State)
		job2.WorkerJob = job.WorkerJob
		// ignore AggState. We fetched the job at different stages of its lifecycle so
		// the states differ.
		job2.AggState = job.AggState
		require.Equal(job, job2)
	}

	{
		stats, err := svc.GetAggregateRepoRevState(userCtx, job.ID)
		require.NoError(err)
		require.Equal(&types.RepoRevJobStats{
			Total:      6,
			Completed:  6, // 1 search job + 2 repo jobs + 3 repo rev jobs
			Failed:     0,
			InProgress: 0,
		}, stats)
	}

	// Assert that we can write the job logs to a writer and that the number of
	// lines and columns matches our expectation.
	{
		service.JobLogsIterLimit = 2
		writerTo, err := svc.GetSearchJobLogsWriterTo(userCtx, job.ID)
		require.NoError(err)
		var buf bytes.Buffer
		_, err = writerTo.WriteTo(&buf)
		require.NoError(err)
		lines := strings.Split(buf.String(), "\n")
		// 1 header + 3 rows + 1 newline
		require.Equal(5, len(lines), fmt.Sprintf("got %q", buf))
		require.Equal("repository,revision,started_at,finished_at,status,failure_message", lines[0])
		// We should use the CSV reader to parse this but since we know none of the
		// columns have a "," in the context of this test, this is fine.
		require.Equal(6, len(strings.Split(lines[1], ",")))
	}

	// Assert that we fail without writing anything if the user is not allowed
	// to view the logs
	{
		userBadCtx := actor.WithActor(context.Background(), actor.FromUser(userBadID))
		_, err = svc.GetSearchJobLogsWriterTo(userBadCtx, job.ID)
		require.ErrorIs(err, auth.ErrMustBeSiteAdminOrSameUser)
	}

	// Assert that cancellation affects the number of rows we expect. This is a bit
	// counterintuitive at this point because we have already completed the job.
	// However, cancellation affects the rows independently of the job state.
	{
		wantCount := 6
		gotCount, err := s.CancelSearchJob(userCtx, job.ID)
		require.NoError(err)
		require.Equal(wantCount, gotCount)
	}

	// Delete should remove the job from the database and the uploadstore.
	{
		require.Equal(3, len(bucket))
		err = svc.DeleteSearchJob(userCtx, job.ID)
		require.NoError(err)
		require.Equal(0, len(bucket))
		_, err = svc.GetSearchJob(userCtx, job.ID)
		require.Error(err)
	}
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

func newMockUploadStore(t *testing.T) (*mocks.MockStorage, map[string]string) {
	t.Helper()

	// Each entry in bucket corresponds to one 1 uploaded csv file.
	mu := sync.Mutex{}
	bucket := make(map[string]string)

	mockStore := mocks.NewMockStorage()
	mockStore.UploadFunc.SetDefaultHook(func(ctx context.Context, key string, r io.Reader) (int64, error) {
		b, err := io.ReadAll(r)
		if err != nil {
			return 0, err
		}

		mu.Lock()
		bucket[key] = string(b)
		mu.Unlock()

		return int64(len(b)), nil
	})

	mockStore.DeleteFunc.SetDefaultHook(func(ctx context.Context, key string) error {
		mu.Lock()
		delete(bucket, key)
		mu.Unlock()

		return nil
	})

	mockStore.ListFunc.SetDefaultHook(func(ctx context.Context, prefix string) (*iterator.Iterator[string], error) {
		var keys []string
		mu.Lock()
		for k := range bucket {
			keys = append(keys, k)
		}
		mu.Unlock()
		return iterator.From(keys), nil
	})

	return mockStore, bucket
}
