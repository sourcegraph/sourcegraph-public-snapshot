package graphqlbackend

import (
	"context"
	"reflect"
	"testing"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestSavedSearches(t *testing.T) {
	ctx := context.Background()
	db := new(dbtesting.MockDB)
	defer resetMocks()

	key := int32(1)

	database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true, ID: key}, nil
	}

	database.Mocks.SavedSearches.ListSavedSearchesByUserID = func(ctx context.Context, userID int32) ([]*types.SavedSearch, error) {
		return []*types.SavedSearch{{ID: key, Description: "test query", Query: "test type:diff patternType:regexp", Notify: true, NotifySlack: false, UserID: &userID, OrgID: nil}}, nil
	}

	savedSearches, err := (&schemaResolver{db: db}).SavedSearches(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := []*savedSearchResolver{{db, types.SavedSearch{
		ID:          key,
		Description: "test query",
		Query:       "test type:diff patternType:regexp",
		Notify:      true,
		NotifySlack: false,
		UserID:      &key,
		OrgID:       nil,
	}}}
	if !reflect.DeepEqual(savedSearches, want) {
		t.Errorf("got %v+, want %v+", savedSearches[0], want[0])
	}
}

func TestCreateSavedSearch(t *testing.T) {
	ctx := context.Background()
	defer resetMocks()
	db := new(dbtesting.MockDB)

	key := int32(1)
	createSavedSearchCalled := false

	database.Mocks.SavedSearches.Create = func(ctx context.Context,
		newSavedSearch *types.SavedSearch,
	) (*types.SavedSearch, error) {
		createSavedSearchCalled = true
		return &types.SavedSearch{ID: key, Description: newSavedSearch.Description, Query: newSavedSearch.Query, Notify: newSavedSearch.Notify, NotifySlack: newSavedSearch.NotifySlack, UserID: newSavedSearch.UserID, OrgID: newSavedSearch.OrgID}, nil
	}
	database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true, ID: key}, nil
	}
	userID := MarshalUserID(key)
	savedSearches, err := (&schemaResolver{db: db}).CreateSavedSearch(ctx, &struct {
		Description string
		Query       string
		NotifyOwner bool
		NotifySlack bool
		OrgID       *graphql.ID
		UserID      *graphql.ID
	}{Description: "test query", Query: "test type:diff patternType:regexp", NotifyOwner: true, NotifySlack: false, OrgID: nil, UserID: &userID})
	if err != nil {
		t.Fatal(err)
	}
	want := &savedSearchResolver{db, types.SavedSearch{
		ID:          key,
		Description: "test query",
		Query:       "test type:diff patternType:regexp",
		Notify:      true,
		NotifySlack: false,
		OrgID:       nil,
		UserID:      &key,
	}}

	if !createSavedSearchCalled {
		t.Errorf("Database method database.SavedSearches.Create not called")
	}

	if !reflect.DeepEqual(savedSearches, want) {
		t.Errorf("got %v+, want %v+", savedSearches, want)
	}

	// Ensure create saved search errors when patternType is not provided in the query.
	_, err = (&schemaResolver{db: db}).CreateSavedSearch(ctx, &struct {
		Description string
		Query       string
		NotifyOwner bool
		NotifySlack bool
		OrgID       *graphql.ID
		UserID      *graphql.ID
	}{Description: "test query", Query: "test type:diff", NotifyOwner: true, NotifySlack: false, OrgID: nil, UserID: &userID})
	if err == nil {
		t.Error("Expected error for createSavedSearch when query does not provide a patternType: field.")
	}
}

func TestUpdateSavedSearch(t *testing.T) {
	ctx := context.Background()
	defer resetMocks()
	db := new(dbtesting.MockDB)

	key := int32(1)
	database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true, ID: key}, nil
	}
	updateSavedSearchCalled := false

	database.Mocks.SavedSearches.Update = func(ctx context.Context, savedSearch *types.SavedSearch) (*types.SavedSearch, error) {
		updateSavedSearchCalled = true
		return &types.SavedSearch{ID: key, Description: savedSearch.Description, Query: savedSearch.Query, Notify: savedSearch.Notify, NotifySlack: savedSearch.NotifySlack, UserID: savedSearch.UserID, OrgID: savedSearch.OrgID}, nil
	}
	userID := MarshalUserID(key)
	savedSearches, err := (&schemaResolver{db: db}).UpdateSavedSearch(ctx, &struct {
		ID          graphql.ID
		Description string
		Query       string
		NotifyOwner bool
		NotifySlack bool
		OrgID       *graphql.ID
		UserID      *graphql.ID
	}{ID: marshalSavedSearchID(key), Description: "updated query description", Query: "test type:diff patternType:regexp", NotifyOwner: true, NotifySlack: false, OrgID: nil, UserID: &userID})
	if err != nil {
		t.Fatal(err)
	}

	want := &savedSearchResolver{db, types.SavedSearch{
		ID:          key,
		Description: "updated query description",
		Query:       "test type:diff patternType:regexp",
		Notify:      true,
		NotifySlack: false,
		OrgID:       nil,
		UserID:      &key,
	}}

	if !updateSavedSearchCalled {
		t.Errorf("Database method database.SavedSearches.Update not called")
	}

	if !reflect.DeepEqual(savedSearches, want) {
		t.Errorf("got %v+, want %v+", savedSearches, want)
	}

	// Ensure update saved search errors when patternType is not provided in the query.
	_, err = (&schemaResolver{db: db}).UpdateSavedSearch(ctx, &struct {
		ID          graphql.ID
		Description string
		Query       string
		NotifyOwner bool
		NotifySlack bool
		OrgID       *graphql.ID
		UserID      *graphql.ID
	}{ID: marshalSavedSearchID(key), Description: "updated query description", Query: "test type:diff", NotifyOwner: true, NotifySlack: false, OrgID: nil, UserID: &userID})
	if err == nil {
		t.Error("Expected error for updateSavedSearch when query does not provide a patternType: field.")
	}
}

func TestDeleteSavedSearch(t *testing.T) {
	ctx := context.Background()
	db := new(dbtesting.MockDB)
	defer resetMocks()

	key := int32(1)
	database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true, ID: key}, nil
	}
	database.Mocks.SavedSearches.GetByID = func(ctx context.Context, id int32) (*api.SavedQuerySpecAndConfig, error) {
		return &api.SavedQuerySpecAndConfig{Spec: api.SavedQueryIDSpec{Subject: api.SettingsSubject{User: &key}, Key: "1"}, Config: api.ConfigSavedQuery{Key: "1", Description: "test query", Query: "test type:diff", Notify: true, NotifySlack: false, UserID: &key, OrgID: nil}}, nil
	}

	deleteSavedSearchCalled := false

	database.Mocks.SavedSearches.Delete = func(ctx context.Context, id int32) error {
		deleteSavedSearchCalled = true
		return nil
	}

	firstSavedSearchGraphqlID := graphql.ID("U2F2ZWRTZWFyY2g6NTI=")
	_, err := (&schemaResolver{db: db}).DeleteSavedSearch(ctx, &struct {
		ID graphql.ID
	}{ID: firstSavedSearchGraphqlID})
	if err != nil {
		t.Fatal(err)
	}

	if !deleteSavedSearchCalled {
		t.Errorf("Database method database.SavedSearches.Delete not called")
	}
}
