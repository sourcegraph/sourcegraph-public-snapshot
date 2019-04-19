package graphqlbackend

import (
	"context"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func TestSavedSearches(t *testing.T) {
	ctx := context.Background()
	defer resetMocks()
	key := int32(1)
	db.Mocks.SavedSearches.ListAll = func(ctx context.Context) ([]api.SavedQuerySpecAndConfig, error) {
		return []api.SavedQuerySpecAndConfig{{Spec: api.SavedQueryIDSpec{Subject: api.SettingsSubject{User: &key}, Key: "1"}, Config: api.ConfigSavedQuery{Key: "1", Description: "test query", Query: "test type:diff", Notify: true, NotifySlack: false, OwnerKind: "user", UserID: &key, OrgID: nil}}}, nil
	}

	savedQueries, err := (&schemaResolver{}).SavedQueries(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := []*savedQueryResolver{{
		key:         "1",
		description: "test query",
		query:       "test type:diff",
		notify:      true,
		notifySlack: false,
		ownerKind:   "user",
		userID:      &key,
		orgID:       nil,
	}}
	if !reflect.DeepEqual(savedQueries, want) {
		t.Errorf("got %v+, want %v+", savedQueries, want)
	}
}

func TestCreateSavedSearch(t *testing.T) {
	ctx := context.Background()
	defer resetMocks()

	key := int32(1)
	createSavedSearchCalled := false

	db.Mocks.SavedSearches.Create = func(ctx context.Context,
		description,
		query string,
		notifyOwner,
		notifySlack bool,
		ownerKind string,
		userID,
		orgID *int32,
	) (*api.ConfigSavedQuery, error) {
		createSavedSearchCalled = true
		return &api.ConfigSavedQuery{Key: "1", Description: description, Query: query, Notify: notifyOwner, NotifySlack: notifySlack, OwnerKind: ownerKind, UserID: userID, OrgID: orgID}, nil
	}

	savedQueries, err := (&schemaResolver{}).CreateSavedSearch(ctx, &struct {
		Description string
		Query       string
		NotifyOwner bool
		NotifySlack bool
		OwnerKind   string
		OrgID       *int32
		UserID      *int32
	}{Description: "test query", Query: "test type:diff", NotifyOwner: true, NotifySlack: false, OwnerKind: "user", OrgID: nil, UserID: &key})

	if err != nil {
		t.Fatal(err)
	}
	want := &savedQueryResolver{
		key:         "1",
		description: "test query",
		query:       "test type:diff",
		notify:      true,
		notifySlack: false,
		ownerKind:   "user",
		orgID:       nil,
		userID:      &key,
	}

	if !createSavedSearchCalled {
		t.Errorf("Database method db.SavedSearches.Create not called")
	}

	if !reflect.DeepEqual(savedQueries, want) {
		t.Errorf("got %v+, want %v+", savedQueries, want)
	}
}

func TestUpdateSavedSearch(t *testing.T) {
	ctx := context.Background()
	defer resetMocks()

	key := int32(1)
	updateSavedSearchCalled := false

	db.Mocks.SavedSearches.Update = func(ctx context.Context, id, description,
		query string,
		notifyOwner,
		notifySlack bool,
		ownerKind string,
		userID,
		orgID *int32) (*api.ConfigSavedQuery, error) {
		updateSavedSearchCalled = true
		return &api.ConfigSavedQuery{Key: "1", Description: description, Query: query, Notify: notifyOwner, NotifySlack: notifySlack, OwnerKind: ownerKind, UserID: userID, OrgID: orgID}, nil
	}

	savedQueries, err := (&schemaResolver{}).UpdateSavedSearch(ctx, &struct {
		ID          string
		Description string
		Query       string
		NotifyOwner bool
		NotifySlack bool
		OwnerKind   string
		OrgID       *int32
		UserID      *int32
	}{ID: "1", Description: "updated query description", Query: "test type:diff", NotifyOwner: true, NotifySlack: false, OwnerKind: "user", OrgID: nil, UserID: &key})
	if err != nil {
		t.Fatal(err)
	}

	want := &savedQueryResolver{
		key:         "1",
		description: "updated query description",
		query:       "test type:diff",
		notify:      true,
		notifySlack: false,
		ownerKind:   "user",
		orgID:       nil,
		userID:      &key,
	}

	if !updateSavedSearchCalled {
		t.Errorf("Database method db.SavedSearches.Update not called")
	}

	if !reflect.DeepEqual(savedQueries, want) {
		t.Errorf("got %v+, want %v+", savedQueries, want)
	}
}

func TestDeleteSavedSearch(t *testing.T) {
	ctx := context.Background()
	defer resetMocks()

	deleteSavedSearchCalled := false

	db.Mocks.SavedSearches.Delete = func(ctx context.Context, id string) error {
		deleteSavedSearchCalled = true
		return nil
	}

	_, err := (&schemaResolver{}).DeleteSavedSearch(ctx, &struct {
		ID string
	}{ID: "1"})
	if err != nil {
		t.Fatal(err)
	}

	if !deleteSavedSearchCalled {
		t.Errorf("Database method db.SavedSearches.Delete not called")
	}
}
