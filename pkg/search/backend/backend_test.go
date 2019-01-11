package backend

import (
	"context"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/search"
	"github.com/sourcegraph/sourcegraph/pkg/vcs"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

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
