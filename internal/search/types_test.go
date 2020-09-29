package search

import (
	"context"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

func repoRev(revSpec string) *RepositoryRevisions {
	return &RepositoryRevisions{
		Repo: &types.Repo{ID: api.RepoID(0), Name: "test/repo"},
		Revs: []RevisionSpecifier{
			{RevSpec: revSpec},
		},
	}
}

func TestRepoPromise(t *testing.T) {
	in := []*RepositoryRevisions{repoRev("HEAD")}
	rp := (&Promise{}).Resolve(in)
	out, err := rp.Get(context.Background())
	if err != nil {
		t.Fatal("error should have been nil, because we supplied a context.Background()")
	}
	if ok := reflect.DeepEqual(in, out); !ok {
		t.Fatalf("got %+v, expected %+v", out, in)
	}
}

func TestRepoPromiseWithCancel(t *testing.T) {
	rp := Promise{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := rp.Get(ctx)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestRepoPromiseConcurrent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in := []*RepositoryRevisions{repoRev("HEAD")}
	rp := Promise{}
	go func() {
		rp.Resolve(in)
	}()
	out, err := rp.Get(ctx)
	if err != nil {
		t.Fatal("error should have been nil, because we didn't cancel the context")
	}
	if ok := reflect.DeepEqual(in, out); !ok {
		t.Fatalf("got %+v, expected %+v", out, in)
	}

	cancel()

	out, err = rp.Get(ctx)
	if err != nil {
		t.Fatal("error should have been nil, because we canceled the context after the first call to get")
	}
	if ok := reflect.DeepEqual(in, out); !ok {
		t.Fatalf("got %+v, expected %+v", out, in)
	}
}
