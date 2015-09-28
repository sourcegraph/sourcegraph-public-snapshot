package authzchecked

import (
	"fmt"
	"reflect"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/auth"
)

type mockRepoChecker struct {
	calledCheckRepo bool
	repoArgs        map[string]struct{}
	returns         error
}

func (m *mockRepoChecker) CheckRepo(ctx context.Context, repo string, perm auth.PermType) error {
	if m.repoArgs == nil {
		m.repoArgs = map[string]struct{}{}
	}
	m.repoArgs[repo] = struct{}{}
	m.calledCheckRepo = true
	return m.returns
}

func (m *mockRepoChecker) calledWithRepoArgs(want ...string) error {
	wantMap := make(map[string]struct{}, len(want))
	for _, repo := range want {
		wantMap[repo] = struct{}{}
	}
	if !reflect.DeepEqual(m.repoArgs, wantMap) {
		return fmt.Errorf("got repo args %v, want %v", m.repoArgs, wantMap)
	}
	return nil
}

func mockRepoCheckerContext() (context.Context, *mockRepoChecker) {
	var rc mockRepoChecker
	ctx := auth.WithRepoChecker(context.Background(), &rc)
	return ctx, &rc
}
