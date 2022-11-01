package iterator

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/hexops/autogold"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/log/logtest"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

type testFunc func(context.Context, api.RepoID, FinishFunc) bool

func testForNextAndFinish(t *testing.T, store *basestore.Store, itr *PersistentRepoIterator, seen []api.RepoID, do testFunc) (*PersistentRepoIterator, []api.RepoID) {
	ctx := context.Background()

	for true {
		repoId, more, finish := itr.NextWithFinish()
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

func TestForNextAndFinish(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))
	store := basestore.NewWithHandle(insightsDB.Handle())

	ctx := context.Background()

	t.Run("iterate with no errors and no interruptions", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		var seen []api.RepoID

		got, _ := testForNextAndFinish(t, store, itr, seen, func(ctx context.Context, id api.RepoID, fn FinishFunc) bool {
			clock.Advance(time.Second * 1)
			err := fn(ctx, store, nil)
			if err != nil {
				t.Fatal(err)
			}
			return true
		})
		jsonify, _ := json.Marshal(got)
		autogold.Want("iterate with no errors and no interruptions", `{"Id":1,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"2021-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:05Z","RuntimeDuration":5000000000,"PercentComplete":1,"TotalCount":5,"SuccessCount":5,"Cursor":5}`).Equal(t, string(jsonify))
	})

	t.Run("iterate with one error and no interruptions", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		var seen []api.RepoID

		got, _ := testForNextAndFinish(t, store, itr, seen, func(ctx context.Context, id api.RepoID, fn FinishFunc) bool {
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
		autogold.Want("iterate with 1 error and no interruptions", `{"Id":2,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"2021-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:05Z","RuntimeDuration":5000000000,"PercentComplete":1,"TotalCount":5,"SuccessCount":4,"Cursor":5}`).Equal(t, string(jsonify))
		autogold.Want("iterate with 1 error and no interruptions error check", errorMap{6: &IterationError{
			id:            1,
			RepoId:        6,
			FailureCount:  1,
			ErrorMessages: []string{"this repo errored"},
		}}).Equal(t, got.errors)
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

		got, seen := testForNextAndFinish(t, store, itr, seen, do)

		require.Equal(t, got.Cursor, 2)
		reloaded, _ := LoadWithClock(ctx, store, got.Id, clock)
		require.Equal(t, reloaded.Cursor, got.Cursor)

		// now iterate from the starting position _after_ reloading from the db
		secondItr, _ := testForNextAndFinish(t, store, reloaded, seen, do)
		jsonify, _ := json.Marshal(secondItr)
		autogold.Want("iterate with no error and 1 interruptions", `{"Id":3,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"2021-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:06Z","RuntimeDuration":5000000000,"PercentComplete":1,"TotalCount":5,"SuccessCount":5,"Cursor":5}`).Equal(t, string(jsonify))
	})

	t.Run("iterate twice and verify progress updates", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)

		var finishFunc FinishFunc

		// iterate once
		_, _, finishFunc = itr.NextWithFinish()
		err := finishFunc(ctx, store, nil)
		if err != nil {
			t.Fatal(err)
		}

		// then twice
		_, _, finishFunc = itr.NextWithFinish()
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
		autogold.Want("iterate twice and verify progress", `{"Id":4,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"2021-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDuration":0,"PercentComplete":0.4,"TotalCount":5,"SuccessCount":2,"Cursor":2}`).Equal(t, string(jsonify))
	})
}

func TestNew(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))
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
