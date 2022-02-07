package graphqlbackend

import (
	"context"
	"reflect"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestSavedSearches(t *testing.T) {
	key := int32(1)

	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true, ID: key}, nil)

	ss := database.NewMockSavedSearchStore()
	ss.ListSavedSearchesByUserIDFunc.SetDefaultHook(func(_ context.Context, userID int32) ([]*types.SavedSearch, error) {
		return []*types.SavedSearch{{ID: key, Description: "test query", Query: "test type:diff patternType:regexp", UserID: &userID, OrgID: nil}}, nil
	})

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.SavedSearchesFunc.SetDefaultReturn(ss)

	savedSearches, err := (&schemaResolver{db: db}).SavedSearches(actor.WithActor(context.Background(), actor.FromUser(key)))
	if err != nil {
		t.Fatal(err)
	}
	want := []*savedSearchResolver{{db, types.SavedSearch{
		ID:          key,
		Description: "test query",
		Query:       "test type:diff patternType:regexp",
		UserID:      &key,
		OrgID:       nil,
	}}}
	if !reflect.DeepEqual(savedSearches, want) {
		t.Errorf("got %v+, want %v+", savedSearches[0], want[0])
	}
}

func TestSavedSearchByIDOwner(t *testing.T) {
	ctx := context.Background()

	userID := int32(1)
	ssID := marshalSavedSearchID(1)

	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: false, ID: userID}, nil)

	ss := database.NewMockSavedSearchStore()
	ss.GetByIDFunc.SetDefaultReturn(
		&api.SavedQuerySpecAndConfig{
			Spec: api.SavedQueryIDSpec{},
			Config: api.ConfigSavedQuery{
				UserID:      &userID,
				Description: "test query",
				Query:       "test type:diff patternType:regexp",
				OrgID:       nil,
			},
		},
		nil,
	)

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.SavedSearchesFunc.SetDefaultReturn(ss)

	ctx = actor.WithActor(ctx, &actor.Actor{
		UID: userID,
	})

	savedSearch, err := (&schemaResolver{db: db}).savedSearchByID(ctx, ssID)
	if err != nil {
		t.Fatal(err)
	}
	want := &savedSearchResolver{
		db: db,
		s: types.SavedSearch{
			ID:          userID,
			Description: "test query",
			Query:       "test type:diff patternType:regexp",
			UserID:      &userID,
			OrgID:       nil,
		},
	}

	if !reflect.DeepEqual(savedSearch, want) {
		t.Errorf("got %v+, want %v+", savedSearch, want)
	}
}

func TestSavedSearchByIDNonOwner(t *testing.T) {
	// Non owners, including site admins cannot view a user's saved searches
	userID := int32(1)
	adminID := int32(2)
	ssID := marshalSavedSearchID(1)

	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true, ID: adminID}, nil)

	ss := database.NewMockSavedSearchStore()
	ss.GetByIDFunc.SetDefaultReturn(
		&api.SavedQuerySpecAndConfig{
			Spec: api.SavedQueryIDSpec{},
			Config: api.ConfigSavedQuery{
				UserID:      &userID,
				Description: "test query",
				Query:       "test type:diff patternType:regexp",
				OrgID:       nil,
			},
		},
		nil,
	)

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.SavedSearchesFunc.SetDefaultReturn(ss)

	ctx := actor.WithActor(context.Background(), &actor.Actor{
		UID: adminID,
	})

	_, err := (&schemaResolver{db: db}).savedSearchByID(ctx, ssID)
	t.Log(err)
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestCreateSavedSearch(t *testing.T) {
	ctx := context.Background()
	key := int32(1)

	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true, ID: key}, nil)

	ss := database.NewMockSavedSearchStore()
	ss.CreateFunc.SetDefaultHook(func(_ context.Context, newSavedSearch *types.SavedSearch) (*types.SavedSearch, error) {
		return &types.SavedSearch{
			ID:          key,
			Description: newSavedSearch.Description,
			Query:       newSavedSearch.Query,
			Notify:      newSavedSearch.Notify,
			NotifySlack: newSavedSearch.NotifySlack,
			UserID:      newSavedSearch.UserID,
			OrgID:       newSavedSearch.OrgID,
		}, nil
	})

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.SavedSearchesFunc.SetDefaultReturn(ss)

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

	mockrequire.Called(t, ss.CreateFunc)

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

	key := int32(1)
	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true, ID: key}, nil)

	ss := database.NewMockSavedSearchStore()
	ss.UpdateFunc.SetDefaultHook(func(ctx context.Context, savedSearch *types.SavedSearch) (*types.SavedSearch, error) {
		return &types.SavedSearch{
			ID:          key,
			Description: savedSearch.Description,
			Query:       savedSearch.Query,
			Notify:      savedSearch.Notify,
			NotifySlack: savedSearch.NotifySlack,
			UserID:      savedSearch.UserID,
			OrgID:       savedSearch.OrgID,
		}, nil
	})

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.SavedSearchesFunc.SetDefaultReturn(ss)

	userID := MarshalUserID(key)
	savedSearches, err := (&schemaResolver{db: db}).UpdateSavedSearch(ctx, &struct {
		ID          graphql.ID
		Description string
		Query       string
		NotifyOwner bool
		NotifySlack bool
		OrgID       *graphql.ID
		UserID      *graphql.ID
	}{ID: marshalSavedSearchID(key), Description: "updated query description", Query: "test type:diff patternType:regexp", OrgID: nil, UserID: &userID})
	if err != nil {
		t.Fatal(err)
	}

	want := &savedSearchResolver{db, types.SavedSearch{
		ID:          key,
		Description: "updated query description",
		Query:       "test type:diff patternType:regexp",
		OrgID:       nil,
		UserID:      &key,
	}}

	mockrequire.Called(t, ss.UpdateFunc)

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

	key := int32(1)
	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true, ID: key}, nil)

	ss := database.NewMockSavedSearchStore()
	ss.GetByIDFunc.SetDefaultReturn(&api.SavedQuerySpecAndConfig{
		Spec: api.SavedQueryIDSpec{
			Subject: api.SettingsSubject{User: &key},
			Key:     "1",
		},
		Config: api.ConfigSavedQuery{
			Key:         "1",
			Description: "test query",
			Query:       "test type:diff",
			UserID:      &key,
			OrgID:       nil,
		},
	}, nil)

	ss.DeleteFunc.SetDefaultReturn(nil)

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.SavedSearchesFunc.SetDefaultReturn(ss)

	firstSavedSearchGraphqlID := graphql.ID("U2F2ZWRTZWFyY2g6NTI=")
	_, err := (&schemaResolver{db: db}).DeleteSavedSearch(ctx, &struct {
		ID graphql.ID
	}{ID: firstSavedSearchGraphqlID})
	if err != nil {
		t.Fatal(err)
	}

	mockrequire.Called(t, ss.DeleteFunc)
}
