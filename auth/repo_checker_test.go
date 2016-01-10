package auth

import (
	"errors"
	"testing"

	"golang.org/x/net/context"
)

func TestRepoChecker_propagatesError(t *testing.T) {
	wantErr := errors.New("x")
	ctx := WithRepoChecker(context.Background(), RepoCheckerFunc(func(ctx context.Context, repo string, what PermType) error {
		return wantErr
	}))

	if err := CheckRepo(ctx, "bar", Read); err != wantErr {
		t.Errorf("got err %v, want %v", err, wantErr)
	}
}

func TestRepoChecker_recursionCausesPanic(t *testing.T) {
	panicked := false

	ctx := WithRepoChecker(context.Background(), RepoCheckerFunc(func(ctx context.Context, repo string, what PermType) error {
		defer func() {
			if v := recover(); v != nil {
				panicked = true
			}
		}()
		CheckRepo(ctx, "foo", Read)
		return nil
	}))

	CheckRepo(ctx, "bar", Read)
	if !panicked {
		t.Error("!panicked")
	}
}
