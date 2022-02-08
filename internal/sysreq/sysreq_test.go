package sysreq

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestCheck(t *testing.T) {
	checks = []check{
		{
			Name: "a",
			Check: func(ctx context.Context) (problem, fix string, err error) {
				return "", "", errors.New("foo")
			},
		},
	}
	st := Check(context.Background(), nil)
	if len(st) != 1 {
		t.Fatalf("unexpected number of statuses. want=%d have=%d", 1, len(st))
	}

	want := Status{Name: "a", Err: errors.New("foo")}
	if !st[0].Equals(want) {
		t.Errorf("got %v, want %v", st[0], want)
	}
}

func TestCheck_skip(t *testing.T) {
	checks = []check{
		{
			Name: "a",
			Check: func(ctx context.Context) (problem, fix string, err error) {
				return "", "", errors.New("foo")
			},
		},
	}
	st := Check(context.Background(), []string{"A"})
	if len(st) != 1 {
		t.Fatalf("unexpected number of statuses. want=%d have=%d", 1, len(st))
	}

	want := Status{Name: "a", Skipped: true}
	if !st[0].Equals(want) {
		t.Errorf("got %v, want %v", st[0], want)
	}
}
