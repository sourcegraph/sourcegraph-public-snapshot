package db

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestSavedSearchesIsEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := dbtesting.TestContext(t)
	isEmpty, err := SavedSearches.IsEmpty(ctx)
	if err != nil {
		t.Fatal()
	}
	want := true
	if want != isEmpty {
		t.Errorf("want %v, got %v", want, isEmpty)
	}

	_, err = Users.Create(ctx, NewUser{DisplayName: "test", Email: "test@test.com", Username: "test", Password: "test", EmailVerificationCode: "c2"})
	if err != nil {
		t.Fatal("can't create user", err)
	}
	userID := int32(1)
	fake := &types.SavedSearch{
		Query:       "test",
		Description: "test",
		Notify:      true,
		NotifySlack: true,
		OwnerKind:   "user",
		UserID:      &userID,
		OrgID:       nil,
	}
	_, err = SavedSearches.Create(ctx, fake)
	if err != nil {
		t.Fatal(err)
	}

	isEmpty, err = SavedSearches.IsEmpty(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want = false
	if want != isEmpty {
		t.Errorf("want %v, got %v", want, isEmpty)
	}

}

func TestSavedSearchesCreate(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := dbtesting.TestContext(t)
	_, err := Users.Create(ctx, NewUser{DisplayName: "test", Email: "test@test.com", Username: "test", Password: "test", EmailVerificationCode: "c2"})
	if err != nil {
		t.Fatal("can't create user", err)
	}
	userID := int32(1)
	fake := &types.SavedSearch{
		Query:       "test",
		Description: "test",
		Notify:      true,
		NotifySlack: true,
		OwnerKind:   "user",
		UserID:      &userID,
		OrgID:       nil,
	}
	ss, err := SavedSearches.Create(ctx, fake)
	if err != nil {
		t.Fatal(err)
	}
	if ss == nil {
		t.Fatalf("no saved search returned, create failed")
	}

	want := &types.SavedSearch{
		ID:          1,
		Query:       "test",
		Description: "test",
		Notify:      true,
		NotifySlack: true,
		OwnerKind:   "user",
		UserID:      &userID,
		OrgID:       nil,
	}
	if !reflect.DeepEqual(ss, want) {
		t.Errorf("query is '%v', want '%v'", ss, want)
	}
}

func TestSavedSearchesUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := dbtesting.TestContext(t)
	_, err := Users.Create(ctx, NewUser{DisplayName: "test", Email: "test@test.com", Username: "test", Password: "test", EmailVerificationCode: "c2"})
	if err != nil {
		t.Fatal("can't create user", err)
	}
	userID := int32(1)
	fake := &types.SavedSearch{
		Query:       "test",
		Description: "test",
		Notify:      true,
		NotifySlack: true,
		OwnerKind:   "user",
		UserID:      &userID,
		OrgID:       nil,
	}
	_, err = SavedSearches.Create(ctx, fake)
	if err != nil {
		t.Fatal(err)
	}

	updated := &types.SavedSearch{
		ID:          1,
		Query:       "test2",
		Description: "test2",
		Notify:      true,
		NotifySlack: true,
		OwnerKind:   "user",
		UserID:      &userID,
		OrgID:       nil,
	}

	updatedSearch, err := SavedSearches.Update(ctx, updated)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(updatedSearch, updated) {
		t.Errorf("updatedSearch is %v, want %v", updatedSearch, updated)
	}
}

func TestSavedSearchesDelete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := dbtesting.TestContext(t)
	_, err := Users.Create(ctx, NewUser{DisplayName: "test", Email: "test@test.com", Username: "test", Password: "test", EmailVerificationCode: "c2"})
	if err != nil {
		t.Fatal("can't create user", err)
	}
	userID := int32(1)
	fake := &types.SavedSearch{
		Query:       "test",
		Description: "test",
		Notify:      true,
		NotifySlack: true,
		OwnerKind:   "user",
		UserID:      &userID,
		OrgID:       nil,
	}
	_, err = SavedSearches.Create(ctx, fake)
	if err != nil {
		t.Fatal(err)
	}

	err = SavedSearches.Delete(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}

	allQueries, err := SavedSearches.ListAll(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(allQueries) > 0 {
		t.Error("expected no queries in saved_searches table")
	}
}

func TestSavedSearchesGetByUserID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := dbtesting.TestContext(t)
	_, err := Users.Create(ctx, NewUser{DisplayName: "test", Email: "test@test.com", Username: "test", Password: "test", EmailVerificationCode: "c2"})
	if err != nil {
		t.Fatal("can't create user", err)
	}
	userID := int32(1)
	fake := &types.SavedSearch{
		Query:       "test",
		Description: "test",
		Notify:      true,
		NotifySlack: true,
		OwnerKind:   "user",
		UserID:      &userID,
		OrgID:       nil,
	}
	ss, err := SavedSearches.Create(ctx, fake)
	if err != nil {
		t.Fatal(err)
	}

	if ss == nil {
		t.Fatalf("no saved search returned, create failed")
	}
	savedSearch, err := SavedSearches.ListSavedSearchesByUserID(ctx, 1)
	want := []*types.SavedSearch{{
		ID:          1,
		Query:       "test",
		Description: "test",
		Notify:      true,
		NotifySlack: true,
		OwnerKind:   "user",
		UserID:      &userID,
		OrgID:       nil,
	}}
	if !reflect.DeepEqual(savedSearch, want) {
		t.Errorf("query is '%v+', want '%v+'", savedSearch, want)
	}
}
