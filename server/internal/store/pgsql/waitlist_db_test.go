// +build pgsqltest

package pgsql

import (
	"reflect"
	"testing"
	"time"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/store"
)

func (s *waitlist) mustAddUser(ctx context.Context, t *testing.T, uid int32) {
	if err := s.AddUser(ctx, uid); err != nil {
		t.Fatal(err)
	}
}

func TestWaitlist_AddUser(t *testing.T) {
	t.Parallel()

	var s waitlist
	ctx, done := testContext()
	defer done()

	uid := int32(123)

	// check user is added
	s.mustAddUser(ctx, t, uid)

	// check user is not added again
	err := s.AddUser(ctx, uid)
	if err == nil {
		t.Fatalf("expected store.ErrWaitlistedUserExists, got nil")
	} else if err != store.ErrWaitlistedUserExists {
		t.Fatalf("expected store.ErrWaitlistedUserExists, got %v", err)
	}
}

func TestWaitlist_GetUser(t *testing.T) {
	t.Parallel()

	var s waitlist
	ctx, done := testContext()
	defer done()

	now := time.Now()
	uid := int32(123)

	// check user not found error
	wUser, err := s.GetUser(ctx, uid)
	if err == nil {
		t.Fatalf("expected *store.WaitlistedUserNotFoundError, got nil")
	} else if _, ok := err.(*store.WaitlistedUserNotFoundError); !ok {
		t.Fatalf("expected *store.WaitlistedUserNotFoundError, got %v", err)
	}

	s.mustAddUser(ctx, t, uid)

	// check user is found
	wUser, err = s.GetUser(ctx, uid)
	if err != nil {
		t.Fatal(err)
	}
	if wUser.UID != uid {
		t.Errorf("got uid %v, want %+v", wUser.UID, uid)
	}
	if wUser.GrantedAt != nil {
		t.Errorf("got granted_at %v, want nil", wUser.GrantedAt)
	}
	if now.After(wUser.AddedAt.Time()) {
		t.Errorf("got added_at %v, want after %v", wUser.AddedAt.Time(), now)
	}
}

func TestWaitlist_GrantUser(t *testing.T) {
	t.Parallel()

	var s waitlist
	ctx, done := testContext()
	defer done()

	uid := int32(123)

	// check user not found error
	err := s.GrantUser(ctx, uid)
	if err == nil {
		t.Fatalf("expected *store.WaitlistedUserNotFoundError, got nil")
	} else if _, ok := err.(*store.WaitlistedUserNotFoundError); !ok {
		t.Fatalf("expected *store.WaitlistedUserNotFoundError, got %v", err)
	}

	s.mustAddUser(ctx, t, uid)

	now := time.Now()

	// check user is granted access
	err = s.GrantUser(ctx, uid)
	if err != nil {
		t.Fatal(err)
	}
	wUser, err := s.GetUser(ctx, uid)
	if err != nil {
		t.Fatal(err)
	}
	if wUser.UID != uid {
		t.Errorf("got uid %v, want %+v", wUser.UID, uid)
	}
	if wUser.GrantedAt == nil {
		t.Errorf("got granted_at nil, want non nil")
	} else if now.After(wUser.GrantedAt.Time()) {
		t.Errorf("got granted_at %v, want after %v", wUser.AddedAt.Time(), now)
	}

	if now.Before(wUser.AddedAt.Time()) {
		t.Errorf("got added_at %v, want before %v", wUser.AddedAt.Time(), now)
	}
}

func TestWaitlist_ListUsers(t *testing.T) {
	t.Parallel()

	var s waitlist
	ctx, done := testContext()
	defer done()

	uids := []int32{123, 456, 789}

	s.mustAddUser(ctx, t, uids[0])
	s.mustAddUser(ctx, t, uids[1])
	s.mustAddUser(ctx, t, uids[2])

	err := s.GrantUser(ctx, uids[2])
	if err != nil {
		t.Fatal(err)
	}

	// check all users are returned
	wUsers, err := s.ListUsers(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	wUIDs := make([]int32, len(wUsers))
	for i := range wUsers {
		wUIDs[i] = wUsers[i].UID
	}
	if !reflect.DeepEqual(wUIDs, uids) {
		t.Errorf("got %+v, want %+v", wUIDs, uids)
	}

	// check only waitlisted users are returned
	wUsers, err = s.ListUsers(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	wUIDs = make([]int32, len(wUsers))
	for i := range wUsers {
		wUIDs[i] = wUsers[i].UID
	}
	if !reflect.DeepEqual(wUIDs, uids[:2]) {
		t.Errorf("got %+v, want %+v", wUIDs, uids[:2])
	}
}

func (s *waitlist) mustAddOrg(ctx context.Context, t *testing.T, orgName string) {
	if err := s.AddOrg(ctx, orgName); err != nil {
		t.Fatal(err)
	}
}

func TestWaitlist_AddOrg(t *testing.T) {
	t.Parallel()

	var s waitlist
	ctx, done := testContext()
	defer done()

	orgName := "org"

	// check org is added
	s.mustAddOrg(ctx, t, orgName)

	// check org is not added again
	err := s.AddOrg(ctx, orgName)
	if err == nil {
		t.Fatalf("expected store.ErrWaitlistedOrgExists, got nil")
	} else if err != store.ErrWaitlistedOrgExists {
		t.Fatalf("expected store.ErrWaitlistedOrgExists, got %v", err)
	}
}

func TestWaitlist_GetOrg(t *testing.T) {
	t.Parallel()

	var s waitlist
	ctx, done := testContext()
	defer done()

	now := time.Now()
	orgName := "org"

	// check org not found error
	wOrg, err := s.GetOrg(ctx, orgName)
	if err == nil {
		t.Fatalf("expected *store.WaitlistedOrgNotFoundError, got nil")
	} else if _, ok := err.(*store.WaitlistedOrgNotFoundError); !ok {
		t.Fatalf("expected *store.WaitlistedOrgNotFoundError, got %v", err)
	}

	s.mustAddOrg(ctx, t, orgName)

	// check user is found
	wOrg, err = s.GetOrg(ctx, orgName)
	if err != nil {
		t.Fatal(err)
	}
	if wOrg.Name != orgName {
		t.Errorf("got org %v, want %+v", wOrg.Name, orgName)
	}
	if wOrg.GrantedAt != nil {
		t.Errorf("got granted_at %v, want nil", wOrg.GrantedAt)
	}
	if now.After(wOrg.AddedAt.Time()) {
		t.Errorf("got added_at %v, want after %v", wOrg.AddedAt.Time(), now)
	}
}

func TestWaitlist_GrantOrg(t *testing.T) {
	t.Parallel()

	var s waitlist
	ctx, done := testContext()
	defer done()

	orgName := "org"

	// check org not found error
	err := s.GrantOrg(ctx, orgName)
	if err == nil {
		t.Fatalf("expected *store.WaitlistedOrgNotFoundError, got nil")
	} else if _, ok := err.(*store.WaitlistedOrgNotFoundError); !ok {
		t.Fatalf("expected *store.WaitlistedOrgNotFoundError, got %v", err)
	}

	s.mustAddOrg(ctx, t, orgName)

	now := time.Now()

	// check org is granted access
	err = s.GrantOrg(ctx, orgName)
	if err != nil {
		t.Fatal(err)
	}
	wOrg, err := s.GetOrg(ctx, orgName)
	if err != nil {
		t.Fatal(err)
	}
	if wOrg.Name != orgName {
		t.Errorf("got org %v, want %+v", wOrg.Name, orgName)
	}
	if wOrg.GrantedAt == nil {
		t.Errorf("got granted_at nil, want non nil")
	} else if now.After(wOrg.GrantedAt.Time()) {
		t.Errorf("got granted_at %v, want after %v", wOrg.AddedAt.Time(), now)
	}

	if now.Before(wOrg.AddedAt.Time()) {
		t.Errorf("got added_at %v, want before %v", wOrg.AddedAt.Time(), now)
	}
}

func TestWaitlist_ListOrgs(t *testing.T) {
	t.Parallel()

	var s waitlist
	ctx, done := testContext()
	defer done()

	orgNames := []string{"o1", "o2", "o3"}

	s.mustAddOrg(ctx, t, orgNames[0])
	s.mustAddOrg(ctx, t, orgNames[1])
	s.mustAddOrg(ctx, t, orgNames[2])

	err := s.GrantOrg(ctx, orgNames[2])
	if err != nil {
		t.Fatal(err)
	}

	// check all orgs are returned
	wOrgs, err := s.ListOrgs(ctx, false, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	wNames := make([]string, len(wOrgs))
	for i := range wOrgs {
		wNames[i] = wOrgs[i].Name
	}
	if !reflect.DeepEqual(wNames, orgNames) {
		t.Errorf("got %+v, want %+v", wNames, orgNames)
	}

	// check only waitlisted orgs are returned
	wOrgs, err = s.ListOrgs(ctx, true, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	wNames = make([]string, len(wOrgs))
	for i := range wOrgs {
		wNames[i] = wOrgs[i].Name
	}
	if !reflect.DeepEqual(wNames, orgNames[:2]) {
		t.Errorf("got %+v, want %+v", wNames, orgNames[:2])
	}

	// check only granted orgs are returned
	wOrgs, err = s.ListOrgs(ctx, false, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	wNames = make([]string, len(wOrgs))
	for i := range wOrgs {
		wNames[i] = wOrgs[i].Name
	}
	if !reflect.DeepEqual(wNames, orgNames[2:]) {
		t.Errorf("got %+v, want %+v", wNames, orgNames[2:])
	}

	// check only filtered orgs are returned
	filterNames := []string{"o1", "o3"}
	wOrgs, err = s.ListOrgs(ctx, false, false, filterNames)
	if err != nil {
		t.Fatal(err)
	}
	wNames = make([]string, len(wOrgs))
	for i := range wOrgs {
		wNames[i] = wOrgs[i].Name
	}
	if !reflect.DeepEqual(wNames, filterNames) {
		t.Errorf("got %+v, want %+v", wNames, filterNames)
	}
}
