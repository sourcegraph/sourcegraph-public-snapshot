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

	"github.com/sourcegraph/log/logtest"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestNextAndFinish(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))
	store := basestore.NewWithHandle(insightsDB.Handle())
	clock := glock.NewMockClock()
	clock.SetCurrent(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))

	repos := []int32{1, 6, 10, 22, 55}

	itr, err := NewWithClock(ctx, store, clock, repos)
	if err != nil {
		t.Fatal(err)
	}

	var seen []api.RepoID
	doAThing := func(ctx context.Context, repoId api.RepoID, finish finishFunc) (err error) {
		seen = append(seen, repoId)
		defer func() { err = finish(ctx, store, err) }()
		clock.Advance(time.Second * 1)
		return nil
	}

	var more = true
	var repoId api.RepoID
	var finish finishFunc
	for more {
		repoId, more, finish = itr.NextWithFinish()
		if !more {
			break
		}
		err := doAThing(ctx, repoId, finish)
		if err != nil {
			t.Fatal(err)
		}
	}

	err = itr.MarkComplete(ctx, store)
	if err != nil {
		t.Fatal(err)
	}

	reloaded, err := LoadWithClock(ctx, store, itr.id, clock)
	require.Equal(t, itr, reloaded)
	require.Equal(t, fmt.Sprintf("%v", repos), fmt.Sprintf("%v", seen))

	jsonify, _ := json.Marshal(reloaded)
	autogold.Want("iterate 5 times with no errors", `{"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"2021-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:05Z","RuntimeDuration":5000000000,"PercentComplete":1,"TotalCount":5,"SuccessCount":5,"Cursor":5}`).Equal(t, string(jsonify))
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

	load, err := Load(ctx, store, itr.id)
	if err != nil {
		return
	}
	require.Equal(t, itr, load)
}
