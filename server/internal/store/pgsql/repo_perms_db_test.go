// +build pgsqltest

package pgsql

import (
	"reflect"
	"testing"

	"src.sourcegraph.com/sourcegraph/store"

	"golang.org/x/net/context"
)

func (s *repoPerms) mustAdd(ctx context.Context, t *testing.T, uid int32, repo string) {
	if err := s.Add(ctx, uid, repo); err != nil {
		t.Fatal(err)
	}
}

func TestRepoPerms_user(t *testing.T) {
	t.Parallel()

	var s repoPerms
	ctx, done := testContext()
	defer done()

	uid := int32(123)
	uid2 := int32(456)
	repos := []string{"r1", "r2", "r3"}

	// check that user permission is empty
	dbRepos, err := s.ListUserRepos(ctx, uid)
	if err != nil {
		t.Fatal(err)
	}
	if len(dbRepos) != 0 {
		t.Errorf("got %+v, want empty slice", dbRepos)
	}

	// check user permission is added
	s.mustAdd(ctx, t, uid, repos[0])
	s.mustAdd(ctx, t, uid, repos[1])
	s.mustAdd(ctx, t, uid2, repos[1])
	s.mustAdd(ctx, t, uid2, repos[2])

	// check user permission is not added again
	err = s.Add(ctx, uid, repos[0])
	if err == nil {
		t.Errorf("expected store.ErrRepoPermissionExists, got nil")
	} else if err != store.ErrRepoPermissionExists {
		t.Errorf("expected store.ErrRepoPermissionExists, got %v", err)
	}

	// get user repo permissions
	valid, err := s.Get(ctx, uid, repos[0])
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Errorf("expected uid %d to have permission, got false", uid)
	}
	valid, err = s.Get(ctx, uid2, repos[0])
	if err != nil {
		t.Fatal(err)
	}
	if valid {
		t.Errorf("expected uid %d to not have permission, got true", uid2)
	}

	// list user repos
	dbRepos, err = s.ListUserRepos(ctx, uid)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dbRepos, repos[:2]) {
		t.Errorf("got %+v, want %+v", dbRepos, repos[:2])
	}

	// check user permissions are updated
	err = s.Update(ctx, uid, repos[1:])
	if err != nil {
		t.Fatal(err)
	}
	dbRepos, err = s.ListUserRepos(ctx, uid)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dbRepos, repos[1:]) {
		t.Errorf("got %+v, want %+v", dbRepos, repos[1:])
	}

	// check user permissions are deleted
	err = s.Delete(ctx, uid, repos[1])
	if err != nil {
		t.Fatal(err)
	}
	err = s.DeleteRepo(ctx, repos[2])
	if err != nil {
		t.Fatal(err)
	}
	dbRepos, err = s.ListUserRepos(ctx, uid)
	if err != nil {
		t.Fatal(err)
	}
	if len(dbRepos) != 0 {
		t.Errorf("got %+v, want empty slice", dbRepos)
	}

	// check that other user's permissions are unaffected
	dbRepos, err = s.ListUserRepos(ctx, uid2)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dbRepos, repos[1:2]) {
		t.Errorf("got %+v, want %+v", dbRepos, repos[1:2])
	}
}
