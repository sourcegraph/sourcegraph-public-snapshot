package graphqlbackend

import (
	"context"
	"reflect"
	"testing"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func TestSavedSearches(t *testing.T) {
	ctx := context.Background()
	defer resetMocks()

	key := int32(1)

	db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true, ID: key}, nil
	}

	db.Mocks.SavedSearches.ListSavedSearchesByUserID = func(ctx context.Context, userID int32) ([]*types.SavedSearch, error) {
		return []*types.SavedSearch{{ID: key, Description: "test query", Query: "test type:diff", Notify: true, NotifySlack: false, OwnerKind: "user", UserID: &userID, OrgID: nil}}, nil
	}

	savedSearches, err := (&schemaResolver{}).SavedSearches(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := []*savedSearchResolver{{
		id:          key,
		description: "test query",
		query:       "test type:diff",
		notify:      true,
		notifySlack: false,
		ownerKind:   "user",
		userID:      &key,
		orgID:       nil,
	}}
	if !reflect.DeepEqual(savedSearches, want) {
		t.Errorf("got %v+, want %v+", savedSearches[0], want[0])
	}
}

func TestCreateSavedSearch(t *testing.T) {
	ctx := context.Background()
	defer resetMocks()

	key := int32(1)
	createSavedSearchCalled := false

	db.Mocks.SavedSearches.Create = func(ctx context.Context,
		newSavedSearch *types.SavedSearch,
	) (*types.SavedSearch, error) {
		createSavedSearchCalled = true
		return &types.SavedSearch{ID: key, Description: newSavedSearch.Description, Query: newSavedSearch.Query, Notify: newSavedSearch.Notify, NotifySlack: newSavedSearch.NotifySlack, OwnerKind: newSavedSearch.OwnerKind, UserID: newSavedSearch.UserID, OrgID: newSavedSearch.OrgID}, nil
	}
	db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true, ID: key}, nil
	}

	savedSearches, err := (&schemaResolver{}).CreateSavedSearch(ctx, &struct {
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
	want := &savedSearchResolver{
		id:          key,
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

	if !reflect.DeepEqual(savedSearches, want) {
		t.Errorf("got %v+, want %v+", savedSearches, want)
	}
}

func TestUpdateSavedSearch(t *testing.T) {
	ctx := context.Background()
	defer resetMocks()

	key := int32(1)
	db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true, ID: key}, nil
	}
	updateSavedSearchCalled := false

	db.Mocks.SavedSearches.Update = func(ctx context.Context, savedSearch *types.SavedSearch) (*types.SavedSearch, error) {
		updateSavedSearchCalled = true
		return &types.SavedSearch{ID: key, Description: savedSearch.Description, Query: savedSearch.Query, Notify: savedSearch.Notify, NotifySlack: savedSearch.NotifySlack, OwnerKind: savedSearch.OwnerKind, UserID: savedSearch.UserID, OrgID: savedSearch.OrgID}, nil
	}

	savedSearches, err := (&schemaResolver{}).UpdateSavedSearch(ctx, &struct {
		ID          graphql.ID
		Description string
		Query       string
		NotifyOwner bool
		NotifySlack bool
		OwnerKind   string
		OrgID       *int32
		UserID      *int32
	}{ID: marshalSavedSearchID(key), Description: "updated query description", Query: "test type:diff", NotifyOwner: true, NotifySlack: false, OwnerKind: "user", OrgID: nil, UserID: &key})
	if err != nil {
		t.Fatal(err)
	}

	want := &savedSearchResolver{
		id:          key,
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

	if !reflect.DeepEqual(savedSearches, want) {
		t.Errorf("got %v+, want %v+", savedSearches, want)
	}
}

func TestDeleteSavedSearch(t *testing.T) {
	ctx := context.Background()
	defer resetMocks()

	key := int32(1)
	db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true, ID: key}, nil
	}
	db.Mocks.SavedSearches.GetSavedSearchByID = func(ctx context.Context, id int32) (*api.SavedQuerySpecAndConfig, error) {
		return &api.SavedQuerySpecAndConfig{Spec: api.SavedQueryIDSpec{Subject: api.SettingsSubject{User: &key}, Key: "1"}, Config: api.ConfigSavedQuery{Key: "1", Description: "test query", Query: "test type:diff", Notify: true, NotifySlack: false, OwnerKind: "user", UserID: &key, OrgID: nil}}, nil
	}

	deleteSavedSearchCalled := false

	db.Mocks.SavedSearches.Delete = func(ctx context.Context, id int32) error {
		deleteSavedSearchCalled = true
		return nil
	}

	firstSavedSearchGraphqlID := graphql.ID("U2F2ZWRTZWFyY2g6NTI=")
	_, err := (&schemaResolver{}).DeleteSavedSearch(ctx, &struct {
		ID graphql.ID
	}{ID: firstSavedSearchGraphqlID})
	if err != nil {
		t.Fatal(err)
	}

	if !deleteSavedSearchCalled {
		t.Errorf("Database method db.SavedSearches.Delete not called")
	}
}
