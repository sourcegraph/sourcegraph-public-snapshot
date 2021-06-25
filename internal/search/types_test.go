package search

import (
	"context"
	"reflect"
	"testing"
)

func TestPromiseGet(t *testing.T) {
	in := []*RepositoryRevisions{nil}
	p := (&RepoPromise{}).Resolve(in)
	out, err := p.Get(context.Background())
	if err != nil {
		t.Fatal("error should have been nil, because we supplied a context.Background()")
	}
	if ok := reflect.DeepEqual(in, out); !ok {
		t.Fatalf("got %+v, expected %+v", out, in)
	}
}

func TestPromiseGetWithCancel(t *testing.T) {
	rp := RepoPromise{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := rp.Get(ctx)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestPromiseGetConcurrent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in := []*RepositoryRevisions{nil}
	p := RepoPromise{}
	go func() {
		p.Resolve(in)
	}()
	out, err := p.Get(ctx)
	if err != nil {
		t.Fatal("error should have been nil, because we didn't cancel the context")
	}
	if ok := reflect.DeepEqual(in, out); !ok {
		t.Fatalf("got %+v, expected %+v", out, in)
	}

	cancel()

	out, err = p.Get(ctx)
	if err != context.Canceled {
		t.Fatalf("got %s, but error should have been \"context canceled\"", err)
	}
	if out != nil {
		t.Fatalf("got %+v, expected nil", out)
	}
}
