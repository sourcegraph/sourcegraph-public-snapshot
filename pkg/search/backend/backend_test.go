package backend

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/search"
	"github.com/sourcegraph/sourcegraph/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/pkg/vcs"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

func TestShardedSearch_error(t *testing.T) {
	// Test we properly drain and cleanup when encountering an error on a
	// shard
	shards := make(chan shard, 2)
	// First shard just fails
	shards <- shard{Searcher: &Mock{Error: errors.New("intentional failure")}}
	// Second shard waits to be cancelled
	shards <- shard{Searcher: searchFunc(func(ctx context.Context, q query.Q, opts *search.Options) (*search.Result, error) {
		select {
		case <-ctx.Done():
		case <-time.After(5 * time.Second):
			t.Fatal("2nd shard did not get cancelled")
		}
		return nil, ctx.Err()
	})}
	close(shards)

	_, err := shardedSearch(context.Background(), shards)
	if got, want := err.Error(), "intentional failure"; got != want {
		t.Fatalf("got error:  %s\nwant error: %s", got, want)
	}
}

func TestHandleError(t *testing.T) {
	cases := []struct {
		Error  error
		Status search.RepositoryStatusType
	}{{
		Status: search.RepositoryStatusSearched,
	}, {
		Error:  &vcs.RepoNotExistError{},
		Status: search.RepositoryStatusMissing,
	}, {
		Error:  notFound{},
		Status: search.RepositoryStatusMissing,
	}, {
		Error:  &vcs.RepoNotExistError{CloneInProgress: true},
		Status: search.RepositoryStatusCloning,
	}, {
		Error:  &git.RevisionNotFoundError{},
		Status: search.RepositoryStatusCommitMissing,
	}, {
		Error:  context.DeadlineExceeded,
		Status: search.RepositoryStatusTimedOut,
	}, {
		Error: context.Canceled,
		// Does not get a status
	}}

	for _, c := range cases {
		want := &search.RepositoryStatus{
			Repository: search.Repository{Name: "test"},
			Source:     search.Source("testsource"),
			Status:     c.Status,
		}
		got, err := handleError(want.Source, want.Repository, c.Error)
		if err != nil {
			if c.Status != "" {
				t.Errorf("handleError(%v) got error %v", c.Error, err)
			}
			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("handleError(%v) got %v want %v", c.Error, got, want)
		}
	}
}

type notFound struct{}

func (notFound) Error() string  { return "" }
func (notFound) NotFound() bool { return true }

type searchFunc func(ctx context.Context, q query.Q, opts *search.Options) (*search.Result, error)

func (fn searchFunc) Search(ctx context.Context, q query.Q, opts *search.Options) (*search.Result, error) {
	return fn(ctx, q, opts)
}

func (searchFunc) Close() {}

func (searchFunc) String() string { return "searchFunc" }
