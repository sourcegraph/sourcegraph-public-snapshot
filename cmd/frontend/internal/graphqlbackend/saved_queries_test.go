package graphqlbackend

import (
	"context"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"

	graphql "github.com/neelance/graphql-go"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

func TestSavedQueries(t *testing.T) {
	ctx := context.Background()

	uid := int32(1)

	defer resetMocks()
	store.Mocks.Users.MockGetByAuthID_Return(t, &sourcegraph.User{ID: uid}, nil)
	store.Mocks.Settings.GetLatest = func(ctx context.Context, subject sourcegraph.ConfigurationSubject) (*sourcegraph.Settings, error) {
		return &sourcegraph.Settings{Contents: `{"search.savedQueries":[{"key":"a","description":"d","query":"q"}]}`}, nil
	}

	mockConfigurationCascadeSubjects = func() ([]*configurationSubject, error) {
		return []*configurationSubject{{user: &userResolver{user: &sourcegraph.User{ID: uid}}}}, nil
	}
	defer func() { mockConfigurationCascadeSubjects = nil }()

	savedQueries, err := (&schemaResolver{}).SavedQueries(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := []*savedQueryResolver{
		{
			key:         "a",
			subject:     &configurationSubject{user: &userResolver{user: &sourcegraph.User{ID: uid}}},
			index:       0,
			description: "d",
			query:       searchQuery{query: "q"},
		},
	}
	if !reflect.DeepEqual(savedQueries, want) {
		t.Errorf("got %+v, want %+v", savedQueries, want)
	}
}

func TestCreateSavedQuery(t *testing.T) {
	ctx := context.Background()

	uid := int32(1)
	ctx = actor.WithActor(ctx, &actor.Actor{UID: "uid"})
	lastID := int32(5)
	subject := &configurationSubject{user: &userResolver{user: &sourcegraph.User{ID: uid, AuthID: "uid"}}}

	defer resetMocks()
	store.Mocks.Users.MockGetByAuthID_Return(t, &sourcegraph.User{ID: uid, AuthID: "uid"}, nil)
	store.Mocks.Users.MockGetByID_Return(t, &sourcegraph.User{ID: uid, AuthID: "uid"}, nil)
	calledSettingsCreateIfUpToDate := false
	store.Mocks.Settings.GetLatest = func(ctx context.Context, subject sourcegraph.ConfigurationSubject) (*sourcegraph.Settings, error) {
		return &sourcegraph.Settings{ID: lastID, Contents: `{"search.savedQueries":[{"key":"a","description":"d","query":"q"}]}`}, nil
	}
	store.Mocks.Settings.CreateIfUpToDate = func(ctx context.Context, subject sourcegraph.ConfigurationSubject, lastKnownSettingsID *int32, authorAuth0ID, contents string) (latestSetting *sourcegraph.Settings, err error) {
		calledSettingsCreateIfUpToDate = true
		return &sourcegraph.Settings{ID: lastID + 1, Contents: `not used`}, nil
	}

	mockConfigurationCascadeSubjects = func() ([]*configurationSubject, error) {
		return []*configurationSubject{subject}, nil
	}
	defer func() { mockConfigurationCascadeSubjects = nil }()

	mutation, err := (&schemaResolver{}).ConfigurationMutation(ctx, &struct {
		Input *configurationMutationGroupInput
	}{Input: &configurationMutationGroupInput{LastID: &lastID, Subject: subject.ID()}})
	if err != nil {
		t.Fatal(err)
	}
	created, err := mutation.CreateSavedQuery(ctx, &struct {
		Description string
		Query       string
		ScopeQuery  string
	}{
		Description: "d2",
		Query:       "q2",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.key == "" {
		t.Error("created.key is empty")
	}
	created.key = "" // randomly generated, can't check against want
	want := &savedQueryResolver{
		subject:     subject,
		index:       1,
		description: "d2",
		query:       searchQuery{query: "q2"},
	}
	if !reflect.DeepEqual(created, want) {
		t.Errorf("got %+v, want %+v", created, want)
	}

	if !calledSettingsCreateIfUpToDate {
		t.Error("!calledSettingsCreateIfUpToDate")
	}
}

func TestUpdateSavedQuery(t *testing.T) {
	ctx := context.Background()

	uid := int32(1)
	ctx = actor.WithActor(ctx, &actor.Actor{UID: "uid"})
	lastID := int32(5)
	subject := &configurationSubject{user: &userResolver{user: &sourcegraph.User{ID: uid, AuthID: "uid"}}}
	newDescription := "d2"

	defer resetMocks()
	store.Mocks.Users.MockGetByAuthID_Return(t, &sourcegraph.User{ID: uid, AuthID: "uid"}, nil)
	store.Mocks.Users.MockGetByID_Return(t, &sourcegraph.User{ID: uid, AuthID: "uid"}, nil)
	calledSettingsGetLatest := false
	calledSettingsCreateIfUpToDate := false
	store.Mocks.Settings.GetLatest = func(ctx context.Context, subject sourcegraph.ConfigurationSubject) (*sourcegraph.Settings, error) {
		calledSettingsGetLatest = true
		if calledSettingsCreateIfUpToDate {
			return &sourcegraph.Settings{ID: lastID + 1, Contents: `{"search.savedQueries":[{"key":"a","description":"d2","query":"q"}]}`}, nil
		}
		return &sourcegraph.Settings{ID: lastID, Contents: `{"search.savedQueries":[{"key":"a","description":"d","query":"q"}]}`}, nil
	}
	store.Mocks.Settings.CreateIfUpToDate = func(ctx context.Context, subject sourcegraph.ConfigurationSubject, lastKnownSettingsID *int32, authorAuth0ID, contents string) (latestSetting *sourcegraph.Settings, err error) {
		calledSettingsCreateIfUpToDate = true
		return &sourcegraph.Settings{ID: lastID + 1, Contents: `not used`}, nil
	}

	mockConfigurationCascadeSubjects = func() ([]*configurationSubject, error) {
		return []*configurationSubject{subject}, nil
	}
	defer func() { mockConfigurationCascadeSubjects = nil }()

	mutation, err := (&schemaResolver{}).ConfigurationMutation(ctx, &struct {
		Input *configurationMutationGroupInput
	}{Input: &configurationMutationGroupInput{LastID: &lastID, Subject: subject.ID()}})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := mutation.UpdateSavedQuery(ctx, &struct {
		ID          graphql.ID
		Description *string
		Query       *string
		ScopeQuery  *string
	}{
		ID:          marshalSavedQueryID(savedQueryIDSpec{Subject: subject.toSubject(), Key: "a"}),
		Description: &newDescription,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := &savedQueryResolver{
		key:         "a",
		subject:     subject,
		index:       0,
		description: "d2",
		query:       searchQuery{query: "q"},
	}
	if !reflect.DeepEqual(updated, want) {
		t.Errorf("got %+v, want %+v", updated, want)
	}

	if !calledSettingsGetLatest {
		t.Error("!calledSettingsGetLatest")
	}
	if !calledSettingsCreateIfUpToDate {
		t.Error("!calledSettingsCreateIfUpToDate")
	}
}
