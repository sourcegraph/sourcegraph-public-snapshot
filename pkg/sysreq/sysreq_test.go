package sysreq

import (
	"errors"
	"reflect"
	"testing"

	"context"
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
	want := []Status{{Name: "a", Err: errors.New("foo")}}
	if !reflect.DeepEqual(st, want) {
		t.Errorf("got %v, want %v", st, want)
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
	want := []Status{{Name: "a", Skipped: true}}
	if !reflect.DeepEqual(st, want) {
		t.Errorf("got %v, want %v", st, want)
	}
}
