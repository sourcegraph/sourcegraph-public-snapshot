package sysreq

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"golang.org/x/net/context"
)

func TestCheck(t *testing.T) {
	checks = []check{
		{
			name: "a",
			check: func(ctx context.Context) (*status, error) {
				return nil, errors.New("foo")
			},
		},
	}
	st := Check(context.Background(), nil)
	want := []Status{{Name: "a", Err: errors.New("foo")}}
	if !reflect.DeepEqual(st, want) {
		t.Errorf("got %v, want %v", st, want)
	}
}

func TestCheck_skip(t *testing.T) {
	checks = []check{
		{
			name: "a",
			check: func(ctx context.Context) (*status, error) {
				return nil, errors.New("foo")
			},
		},
	}
	st := Check(context.Background(), []string{"A"})
	want := []Status{{Name: "a", Skipped: true}}
	if !reflect.DeepEqual(st, want) {
		t.Errorf("got %v, want %v", st, want)
	}
}

func TestCheck_timeout(t *testing.T) {
	checks = []check{
		{
			name: "a",
			check: func(ctx context.Context) (*status, error) {
				time.Sleep(150 * time.Millisecond)
				return nil, nil
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	st := Check(ctx, nil)
	want := []Status{{Name: "a", Err: context.DeadlineExceeded}}
	if !reflect.DeepEqual(st, want) {
		t.Errorf("got %v, want %v", st, want)
	}
}
