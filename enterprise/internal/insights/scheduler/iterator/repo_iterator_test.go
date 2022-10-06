package iterator

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"

	"github.com/sourcegraph/log/logtest"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestIterator(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	// clock := timeutil.Now
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))
	store := basestore.NewWithHandle(insightsDB.Handle())

	repos := []int{1, 6, 10, 22, 55}

	itr, err := New(ctx, store, repos)
	if err != nil {
		t.Fatal(err)
	}

	var more = true
	var repoId api.RepoID
	var finish finishFunc
	for more {
		repoId, more, finish = itr.NextWithFinish()
		if !more {
			break
		}
		err := doAThing(ctx, store, repoId, finish)
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Log(fmt.Sprintf("%v", *itr))
	t.Log(fmt.Sprintf("%v", itr.errors.String()))

	err = itr.MarkComplete(ctx, store)
	if err != nil {
		t.Fatal(err)
	}
}

func doAThing(ctx context.Context, store *basestore.Store, repoId api.RepoID, finish finishFunc) (err error) {
	fmt.Println(fmt.Sprintf("repo_id: %d", repoId))
	defer func() { err = finish(ctx, store, err) }()

	if repoId == api.RepoID(22) {
		return errors.New("testing error")
	}
	return nil
}
