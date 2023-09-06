package iterator

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/log/logtest"

	edb "github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

type testFunc func(context.Context, api.RepoID, FinishFunc) bool
type testPageFunc func(context.Context, []api.RepoID, FinishNFunc) bool

func testForNextAndFinish(t *testing.T, store *basestore.Store, itr *PersistentRepoIterator, config IterationConfig, seen []api.RepoID, do testFunc) (*PersistentRepoIterator, []api.RepoID) {
	ctx := context.Background()

	for true {
		repoId, more, finish := itr.NextWithFinish(config)
		if !more {
			break
		}
		shouldNext := do(ctx, repoId, finish)
		if !shouldNext {
			return itr, seen
		}
		seen = append(seen, repoId)
	}

	err := itr.MarkComplete(ctx, store)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, fmt.Sprintf("%v", itr.repos), fmt.Sprintf("%v", seen))
	return itr, seen
}

func testForNextNAndFinish(t *testing.T, store *basestore.Store, itr *PersistentRepoIterator, config IterationConfig, pageSize int, seen []api.RepoID, do testPageFunc) (*PersistentRepoIterator, []api.RepoID) {
	ctx := context.Background()

	for true {
		repoIds, more, finish := itr.NextPageWithFinish(pageSize, config)
		if !more {
			break
		}
		shouldNext := do(ctx, repoIds, finish)
		if !shouldNext {
			return itr, seen
		}
		seen = append(seen, repoIds...)
	}

	err := itr.MarkComplete(ctx, store)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, fmt.Sprintf("%v", itr.repos), fmt.Sprintf("%v", seen))
	return itr, seen
}

func TestForNextAndFinish(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	store := basestore.NewWithHandle(insightsDB.Handle())

	ctx := context.Background()

	t.Run("iterate with no errors and no interruptions", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		var seen []api.RepoID

		got, _ := testForNextAndFinish(t, store, itr, IterationConfig{}, seen, func(ctx context.Context, id api.RepoID, fn FinishFunc) bool {
			clock.Advance(time.Second * 1)
			err := fn(ctx, store, nil)
			if err != nil {
				t.Fatal(err)
			}
			return true
		})
		jsonify, _ := json.Marshal(got)
		autogold.Expect(`{"Id":1,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"2021-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:05Z","RuntimeDuration":5000000000,"PercentComplete":1,"TotalCount":5,"SuccessCount":5,"Cursor":5}`).Equal(t, string(jsonify))
	})

	t.Run("iterate with one error and no interruptions", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		var seen []api.RepoID

		got, _ := testForNextAndFinish(t, store, itr, IterationConfig{}, seen, func(ctx context.Context, id api.RepoID, fn FinishFunc) bool {
			clock.Advance(time.Second * 1)
			var executionErr error
			if id == 6 {
				executionErr = errors.New("this repo errored")
			}
			err := fn(ctx, store, executionErr)
			if err != nil {
				t.Fatal(err)
			}
			return true
		})
		jsonify, _ := json.Marshal(got)
		autogold.Expect(`{"Id":2,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"2021-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:05Z","RuntimeDuration":5000000000,"PercentComplete":1,"TotalCount":5,"SuccessCount":4,"Cursor":5}`).Equal(t, string(jsonify))
		autogold.Expect(errorMap{6: &IterationError{
			id:            1,
			RepoId:        6,
			FailureCount:  1,
			ErrorMessages: []string{"this repo errored"},
		}}).Equal(t, got.errors)

		autogold.Expect(false).Equal(t, got.HasMore())
		autogold.Expect(true).Equal(t, got.HasErrors())
	})

	t.Run("iterate with one error no interruptions and MaxFailures configured", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		var seen []api.RepoID

		got, _ := testForNextAndFinish(t, store, itr, IterationConfig{MaxFailures: 3}, seen, func(ctx context.Context, id api.RepoID, fn FinishFunc) bool {
			clock.Advance(time.Second * 1)
			var executionErr error
			if id == 6 {
				executionErr = errors.New("this repo errored")
			}
			err := fn(ctx, store, executionErr)
			if err != nil {
				t.Fatal(err)
			}
			return true
		})
		jsonify, _ := json.Marshal(got)
		autogold.Expect(`{"Id":3,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"2021-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:05Z","RuntimeDuration":5000000000,"PercentComplete":1,"TotalCount":5,"SuccessCount":4,"Cursor":5}`).Equal(t, string(jsonify))
		autogold.Expect(errorMap{6: &IterationError{
			id:            2,
			RepoId:        6,
			FailureCount:  1,
			ErrorMessages: []string{"this repo errored"},
		}}).Equal(t, got.errors)
		autogold.Expect(false).Equal(t, got.HasMore())
		autogold.Expect(true).Equal(t, got.HasErrors())
	})

	t.Run("iterate with one error no interruptions finished if MaxFailures reached", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		var seen []api.RepoID

		got, _ := testForNextAndFinish(t, store, itr, IterationConfig{MaxFailures: 1}, seen, func(ctx context.Context, id api.RepoID, fn FinishFunc) bool {
			clock.Advance(time.Second * 1)
			var executionErr error
			if id == 6 {
				executionErr = errors.New("this repo errored")
			}
			err := fn(ctx, store, executionErr)
			if err != nil {
				t.Fatal(err)
			}
			return true
		})
		jsonify, _ := json.Marshal(got)
		autogold.Expect(`{"Id":4,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"2021-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:05Z","RuntimeDuration":5000000000,"PercentComplete":1,"TotalCount":5,"SuccessCount":4,"Cursor":5}`).Equal(t, string(jsonify))
		autogold.Expect(errorMap{6: &IterationError{
			id:            3,
			RepoId:        6,
			FailureCount:  1,
			ErrorMessages: []string{"this repo errored"},
		}}).Equal(t, got.errors)
		autogold.Expect(false).Equal(t, got.HasMore())
		autogold.Expect(true).Equal(t, got.HasErrors())
	})

	t.Run("iterate with no errors and one interruption", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		var seen []api.RepoID

		hasStopped := false
		do := func(ctx context.Context, id api.RepoID, fn FinishFunc) bool {
			clock.Advance(time.Second * 1)
			var executionErr error
			if id == 6 && !hasStopped {
				hasStopped = true
				return false
			}
			err := fn(ctx, store, executionErr)
			if err != nil {
				t.Fatal(err)
			}
			return true
		}

		got, seen := testForNextAndFinish(t, store, itr, IterationConfig{}, seen, do)

		require.Equal(t, got.Cursor, 2)
		reloaded, _ := LoadWithClock(ctx, store, got.Id, clock)
		require.Equal(t, reloaded.Cursor, got.Cursor)

		// now iterate from the starting position _after_ reloading from the db
		secondItr, _ := testForNextAndFinish(t, store, reloaded, IterationConfig{}, seen, do)
		jsonify, _ := json.Marshal(secondItr)
		autogold.Expect(`{"Id":5,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"2021-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:06Z","RuntimeDuration":5000000000,"PercentComplete":1,"TotalCount":5,"SuccessCount":5,"Cursor":5}`).Equal(t, string(jsonify))
	})

	t.Run("iterate twice and verify progress updates", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)

		var finishFunc FinishFunc

		// iterate once
		_, _, finishFunc = itr.NextWithFinish(IterationConfig{})
		err := finishFunc(ctx, store, nil)
		if err != nil {
			t.Fatal(err)
		}

		// then twice
		_, _, finishFunc = itr.NextWithFinish(IterationConfig{})
		err = finishFunc(ctx, store, nil)
		if err != nil {
			t.Fatal(err)
		}

		// we should see 40% progress
		reloaded, err := Load(ctx, store, itr.Id)
		if err != nil {
			t.Fatal(err)
		}
		jsonify, _ := json.Marshal(reloaded)
		autogold.Expect(`{"Id":6,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"2021-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDuration":0,"PercentComplete":0.4,"TotalCount":5,"SuccessCount":2,"Cursor":2}`).Equal(t, string(jsonify))
	})

	//test paging
	t.Run("iterate page with no errors and no interruptions", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		var seen []api.RepoID

		got, _ := testForNextNAndFinish(t, store, itr, IterationConfig{}, 2, seen, func(ctx context.Context, ids []api.RepoID, fn FinishNFunc) bool {
			clock.Advance(time.Second * 1)
			err := fn(ctx, store, nil)
			if err != nil {
				t.Fatal(err)
			}
			return true
		})
		jsonify, _ := json.Marshal(got)
		autogold.Expect(`{"Id":7,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"2021-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:03Z","RuntimeDuration":3000000000,"PercentComplete":1,"TotalCount":5,"SuccessCount":5,"Cursor":5}`).Equal(t, string(jsonify))
	})

	t.Run("iterate page with one error and no interruptions", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		var seen []api.RepoID

		got, _ := testForNextNAndFinish(t, store, itr, IterationConfig{}, 2, seen, func(ctx context.Context, ids []api.RepoID, fn FinishNFunc) bool {
			clock.Advance(time.Second * 1)
			executionErrs := map[int32]error{}
			for _, id := range ids {
				if id == 6 {
					executionErrs[int32(id)] = errors.New("this repo errored")
				}
			}

			err := fn(ctx, store, executionErrs)
			if err != nil {
				t.Fatal(err)
			}
			return true
		})
		jsonify, _ := json.Marshal(got)
		autogold.Expect(`{"Id":8,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"2021-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:03Z","RuntimeDuration":3000000000,"PercentComplete":1,"TotalCount":5,"SuccessCount":4,"Cursor":5}`).Equal(t, string(jsonify))
		autogold.Expect(errorMap{6: &IterationError{
			id:            4,
			RepoId:        6,
			FailureCount:  1,
			ErrorMessages: []string{"this repo errored"},
		}}).Equal(t, got.errors)
		autogold.Expect(false).Equal(t, got.HasMore())
		autogold.Expect(true).Equal(t, got.HasErrors())
	})

	t.Run("iterate page with no errors and one interruption", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		var seen []api.RepoID

		hasStopped := false
		do := func(ctx context.Context, ids []api.RepoID, fn FinishNFunc) bool {
			clock.Advance(time.Second * 1)
			executionErrs := map[int32]error{}
			for _, id := range ids {
				if id == 6 && !hasStopped {
					hasStopped = true
					return false
				}
			}

			err := fn(ctx, store, executionErrs)
			if err != nil {
				t.Fatal(err)
			}
			return true
		}

		got, seen := testForNextNAndFinish(t, store, itr, IterationConfig{}, 2, seen, do)

		require.Equal(t, got.Cursor, 2)
		reloaded, _ := LoadWithClock(ctx, store, got.Id, clock)
		require.Equal(t, reloaded.Cursor, got.Cursor)

		// now iterate from the starting position _after_ reloading from the db
		secondItr, _ := testForNextNAndFinish(t, store, reloaded, IterationConfig{}, 2, seen, do)
		jsonify, _ := json.Marshal(secondItr)
		autogold.Expect(`{"Id":9,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"2021-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:04Z","RuntimeDuration":3000000000,"PercentComplete":1,"TotalCount":5,"SuccessCount":5,"Cursor":5}`).Equal(t, string(jsonify))
	})

	t.Run("iterate two pages and verify progress updates", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)

		var finishFunc FinishNFunc

		// iterate once
		_, _, finishFunc = itr.NextPageWithFinish(2, IterationConfig{})
		err := finishFunc(ctx, store, nil)
		if err != nil {
			t.Fatal(err)
		}

		// then twice
		_, _, finishFunc = itr.NextPageWithFinish(2, IterationConfig{})
		err = finishFunc(ctx, store, nil)
		if err != nil {
			t.Fatal(err)
		}

		// we should see 80% progress
		reloaded, err := Load(ctx, store, itr.Id)
		if err != nil {
			t.Fatal(err)
		}
		jsonify, _ := json.Marshal(reloaded)
		autogold.Expect(`{"Id":10,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"2021-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDuration":0,"PercentComplete":0.8,"TotalCount":5,"SuccessCount":4,"Cursor":4}`).Equal(t, string(jsonify))
	})

	t.Run("iterate pages with one error no interruptions and MaxFailures configured", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		var seen []api.RepoID

		got, _ := testForNextNAndFinish(t, store, itr, IterationConfig{MaxFailures: 3}, 2, seen, func(ctx context.Context, ids []api.RepoID, fn FinishNFunc) bool {
			clock.Advance(time.Second * 1)
			executionErr := map[int32]error{}
			for _, id := range ids {
				if id == 6 {
					executionErr[int32(id)] = errors.New("this repo errored")
				}
			}

			err := fn(ctx, store, executionErr)
			if err != nil {
				t.Fatal(err)
			}
			return true
		})
		jsonify, _ := json.Marshal(got)
		autogold.Expect(`{"Id":11,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"2021-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:03Z","RuntimeDuration":3000000000,"PercentComplete":1,"TotalCount":5,"SuccessCount":4,"Cursor":5}`).Equal(t, string(jsonify))
		autogold.Expect(errorMap{6: &IterationError{
			id:            5,
			RepoId:        6,
			FailureCount:  1,
			ErrorMessages: []string{"this repo errored"},
		}}).Equal(t, got.errors)
		autogold.Expect(false).Equal(t, got.HasMore())
		autogold.Expect(true).Equal(t, got.HasErrors())
	})

	t.Run("iterate page with one error no interruptions finished if MaxFailures reached", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		var seen []api.RepoID

		got, _ := testForNextNAndFinish(t, store, itr, IterationConfig{MaxFailures: 1}, 2, seen, func(ctx context.Context, ids []api.RepoID, fn FinishNFunc) bool {
			clock.Advance(time.Second * 1)
			executionErrs := map[int32]error{}
			for _, id := range ids {
				if id == 6 {
					executionErrs[int32(id)] = errors.New("this repo errored")
				}
			}

			err := fn(ctx, store, executionErrs)
			if err != nil {
				t.Fatal(err)
			}
			return true
		})
		jsonify, _ := json.Marshal(got)
		autogold.Expect(`{"Id":12,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"2021-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:03Z","RuntimeDuration":3000000000,"PercentComplete":1,"TotalCount":5,"SuccessCount":4,"Cursor":5}`).Equal(t, string(jsonify))
		autogold.Expect(errorMap{6: &IterationError{
			id:            6,
			RepoId:        6,
			FailureCount:  1,
			ErrorMessages: []string{"this repo errored"},
		}}).Equal(t, got.errors)
		autogold.Expect(false).Equal(t, got.HasMore())
		autogold.Expect(true).Equal(t, got.HasErrors())
	})
}

func TestNew(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	store := basestore.NewWithHandle(insightsDB.Handle())

	repos := []int32{1, 6, 10, 22, 55}

	itr, err := New(ctx, store, repos)
	if err != nil {
		t.Fatal(err)
	}

	load, err := Load(ctx, store, itr.Id)
	if err != nil {
		return
	}
	require.Equal(t, itr, load)
}

func TestForNextRetryAndFinish(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	store := basestore.NewWithHandle(insightsDB.Handle())

	ctx := context.Background()

	t.Run("iterate retry with one error", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		var seen []api.RepoID

		addError(ctx, itr, store, t)
		require.Equal(t, 1, itr.Cursor)
		require.Equal(t, 1, len(itr.errors))
		require.Equal(t, float64(0), itr.PercentComplete)

		got, _ := testForNextRetryAndFinish(t, itr, seen, func(ctx context.Context, id api.RepoID, fn FinishFunc) bool {
			clock.Advance(time.Second * 1)
			err := fn(ctx, store, nil)
			if err != nil {
				t.Fatal(err)
			}
			return true
		}, IterationConfig{})
		jsonify, _ := json.Marshal(got)
		autogold.Expect(`{"Id":1,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"2021-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDuration":1000000000,"PercentComplete":0.2,"TotalCount":5,"SuccessCount":1,"Cursor":1}`).Equal(t, string(jsonify))
	})

	t.Run("ensure retries are reloaded", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		var seen []api.RepoID

		addError(ctx, itr, store, t)
		addError(ctx, itr, store, t)
		require.Equal(t, 2, itr.Cursor)
		require.Equal(t, 2, len(itr.errors))
		require.Equal(t, float64(0), itr.PercentComplete)

		testForNextRetryAndFinish(t, itr, seen, func(ctx context.Context, id api.RepoID, fn FinishFunc) bool {
			clock.Advance(time.Second * 1)
			if id == 1 {
				// we will not retry repo 1 (implying it was successfully retried)
				fn(ctx, store, nil)
				return true
			}
			err := fn(ctx, store, errors.New("fake err"))
			if err != nil {
				t.Fatal(err)
			}
			return true
		}, IterationConfig{})
		require.Equal(t, 1, len(itr.errors))
		require.Equal(t, 2, itr.Cursor)
		require.Equal(t, 0.5, itr.PercentComplete)

		reloaded, err := Load(ctx, store, itr.Id)
		if err != nil {
			t.Fatal(err)
		}

		require.Equal(t, 1, len(reloaded.errors))
		require.Equal(t, 2, reloaded.Cursor)
		require.Equal(t, 0.5, reloaded.PercentComplete)

		var currentErrors []IterationError
		for _, val := range reloaded.errors {
			v := val
			currentErrors = append(currentErrors, *v)
		}
		require.Equal(t, 1, len(currentErrors))
		require.Equal(t, int32(5), currentErrors[0].RepoId)
		require.Equal(t, 2, currentErrors[0].FailureCount)

		jsonify, _ := json.Marshal(reloaded)
		autogold.Expect(`{"Id":2,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"2021-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDuration":2000000000,"PercentComplete":0.5,"TotalCount":2,"SuccessCount":1,"Cursor":2}`).Equal(t, string(jsonify))
	})
	t.Run("ensure retries complete", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		var seen []api.RepoID

		addError(ctx, itr, store, t)
		addError(ctx, itr, store, t)
		require.Equal(t, 2, itr.Cursor)
		require.Equal(t, 2, len(itr.errors))
		require.Equal(t, float64(0), itr.PercentComplete)

		testForNextRetryAndFinish(t, itr, seen, func(ctx context.Context, id api.RepoID, fn FinishFunc) bool {
			clock.Advance(time.Second * 1)

			err := fn(ctx, store, nil)
			if err != nil {
				t.Fatal(err)
			}
			return true
		}, IterationConfig{})
		require.Equal(t, 0, len(itr.errors))
		require.Equal(t, 2, itr.Cursor)
		require.Equal(t, float64(1), itr.PercentComplete)
		require.Equal(t, 0, len(itr.errors))

		jsonify, _ := json.Marshal(itr)
		autogold.Expect(`{"Id":3,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"2021-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDuration":2000000000,"PercentComplete":1,"TotalCount":2,"SuccessCount":2,"Cursor":2}`).Equal(t, string(jsonify))
	})
	t.Run("ensure retry that exceeds max attempts calls back", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		var seen []api.RepoID

		addError(ctx, itr, store, t)
		addError(ctx, itr, store, t)
		require.Equal(t, 2, itr.Cursor)
		require.Equal(t, 2, len(itr.errors))
		require.Equal(t, float64(0), itr.PercentComplete)

		terminalCount := 0
		testForNextRetryAndFinish(t, itr, seen, func(ctx context.Context, id api.RepoID, fn FinishFunc) bool {
			clock.Advance(time.Second * 1)

			err := fn(ctx, store, errors.New("second err"))
			if err != nil {
				t.Fatal(err)
			}
			return true
		}, IterationConfig{MaxFailures: 2, OnTerminal: func(ctx context.Context, store *basestore.Store, repoId int32, terminalErr error) error {
			terminalCount += 1
			return nil
		}})

		require.Equal(t, 0, len(itr.errors))
		require.Equal(t, 2, len(itr.terminalErrors))
		require.Equal(t, 2, itr.Cursor)
		require.Equal(t, float64(0), itr.PercentComplete)
		require.Equal(t, 2, terminalCount)

		jsonify, _ := json.Marshal(itr)
		autogold.Expect(`{"Id":4,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"2021-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDuration":2000000000,"PercentComplete":0,"TotalCount":2,"SuccessCount":0,"Cursor":2}`).Equal(t, string(jsonify))
	})

	t.Run("ensure retry with all terminal errors has no errors to continue", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		var seen []api.RepoID

		addError(ctx, itr, store, t)
		addError(ctx, itr, store, t)
		require.Equal(t, 2, itr.Cursor)
		require.Equal(t, 2, len(itr.errors))
		require.Equal(t, float64(0), itr.PercentComplete)
		itr.retryRepos = nil

		testForNextRetryAndFinish(t, itr, seen, func(ctx context.Context, id api.RepoID, fn FinishFunc) bool {
			clock.Advance(time.Second * 1)

			err := fn(ctx, store, errors.New("second err"))
			if err != nil {
				t.Fatal(err)
			}
			return true
		}, IterationConfig{MaxFailures: 1, OnTerminal: func(ctx context.Context, store *basestore.Store, repoId int32, terminalErr error) error {
			return nil
		}})

		require.False(t, itr.HasErrors())
		require.Equal(t, 2, len(itr.terminalErrors))
		require.Equal(t, 2, itr.Cursor)
		require.Equal(t, float64(0), itr.PercentComplete)

		jsonify, _ := json.Marshal(itr)
		autogold.Expect(`{"Id":5,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"2021-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDuration":0,"PercentComplete":0,"TotalCount":2,"SuccessCount":0,"Cursor":2}`).Equal(t, string(jsonify))
	})
}

func TestReset(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	store := basestore.NewWithHandle(insightsDB.Handle())

	ctx := context.Background()

	clock := glock.NewMockClock()
	clock.SetCurrent(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
	repos := []int32{1, 5, 6, 14, 17}
	itr, _ := NewWithClock(ctx, store, clock, repos)
	var seen []api.RepoID

	addError(ctx, itr, store, t)
	require.Equal(t, 1, itr.Cursor)
	require.Equal(t, 1, len(itr.errors))
	require.Equal(t, float64(0), itr.PercentComplete)

	itrAfterStep, _ := testForNextRetryAndFinish(t, itr, seen, func(ctx context.Context, id api.RepoID, fn FinishFunc) bool {
		clock.Advance(time.Second * 1)
		err := fn(ctx, store, nil)
		if err != nil {
			t.Fatal(err)
		}
		return true
	}, IterationConfig{})
	jsonify, _ := json.Marshal(itrAfterStep)
	autogold.Expect(`{"Id":1,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"2021-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDuration":1000000000,"PercentComplete":0.2,"TotalCount":5,"SuccessCount":1,"Cursor":1}`).Equal(t, string(jsonify))

	err := itrAfterStep.Restart(ctx, store)
	require.NoError(t, err, "restart should not error")
	reloaded, err := LoadWithClock(ctx, store, itrAfterStep.Id, clock)
	require.NoError(t, err, "load should not error")
	resetJson, _ := json.Marshal(reloaded)
	autogold.Expect(`{"Id":1,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"0001-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDuration":0,"PercentComplete":0,"TotalCount":5,"SuccessCount":0,"Cursor":0}`).Equal(t, string(resetJson))

}

func TestEmptyIterator(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	store := basestore.NewWithHandle(insightsDB.Handle())

	ctx := context.Background()

	clock := glock.NewMockClock()
	clock.SetCurrent(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
	repos := []int32{}
	itr, _ := NewWithClock(ctx, store, clock, repos)
	var seen []api.RepoID

	got, _ := testForNextAndFinish(t, store, itr, IterationConfig{}, seen, func(ctx context.Context, id api.RepoID, fn FinishFunc) bool {
		clock.Advance(time.Second * 1)
		err := fn(ctx, store, nil)
		if err != nil {
			t.Fatal(err)
		}
		return true
	})
	jsonify, _ := json.Marshal(got)
	autogold.Expect(`{"Id":1,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"0001-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:00Z","RuntimeDuration":0,"PercentComplete":1,"TotalCount":0,"SuccessCount":0,"Cursor":0}`).Equal(t, string(jsonify))

}

func addError(ctx context.Context, itr *PersistentRepoIterator, store *basestore.Store, t *testing.T) {
	// create an error
	_, _, finish := itr.NextWithFinish(IterationConfig{})
	err := finish(ctx, store, errors.New("fake err"))
	if err != nil {
		t.Fatal(err)
	}
}

func testForNextRetryAndFinish(t *testing.T, itr *PersistentRepoIterator, seen []api.RepoID, do testFunc, config IterationConfig) (*PersistentRepoIterator, []api.RepoID) {
	ctx := context.Background()

	for true {
		repoId, more, finish := itr.NextRetryWithFinish(config)
		if !more {
			break
		}
		shouldNext := do(ctx, repoId, finish)
		if !shouldNext {
			return itr, seen
		}
		seen = append(seen, repoId)
	}

	require.Equal(t, fmt.Sprintf("%v", itr.retryRepos), fmt.Sprintf("%v", seen))
	return itr, seen
}
