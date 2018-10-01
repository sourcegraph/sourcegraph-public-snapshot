package graphqlbackend

import (
	"context"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/api"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func TestSavedQueries(t *testing.T) {
	ctx := context.Background()

	uid := int32(1)
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})

	defer resetMocks()
	db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{ID: uid}, nil
	}
	db.Mocks.Settings.GetLatest = func(ctx context.Context, subject api.ConfigurationSubject) (*api.Settings, error) {
		return &api.Settings{Contents: `{"search.savedQueries":[{"key":"a","description":"d","query":"q"}]}`}, nil
	}

	mockConfigurationCascadeSubjects = func() ([]*configurationSubject, error) {
		return []*configurationSubject{{user: &UserResolver{user: &types.User{ID: uid}}}}, nil
	}
	defer func() { mockConfigurationCascadeSubjects = nil }()

	savedQueries, err := (&schemaResolver{}).SavedQueries(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := []*savedQueryResolver{
		{
			key:            "a",
			subject:        &configurationSubject{user: &UserResolver{user: &types.User{ID: uid}}},
			index:          0,
			description:    "d",
			query:          "q",
			showOnHomepage: false,
		},
	}
	if !reflect.DeepEqual(savedQueries, want) {
		t.Errorf("got %+v, want %+v", savedQueries, want)
	}
}

func TestCreateSavedQuery(t *testing.T) {
	ctx := context.Background()

	uid := int32(1)
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
	lastID := int32(5)
	subject := &configurationSubject{user: &UserResolver{user: &types.User{ID: uid}}}

	defer resetMocks()
	db.Mocks.Users.MockGetByID_Return(t, &types.User{ID: uid}, nil)
	calledSettingsCreateIfUpToDate := false
	db.Mocks.Settings.GetLatest = func(ctx context.Context, subject api.ConfigurationSubject) (*api.Settings, error) {
		return &api.Settings{ID: lastID, Contents: `{"search.savedQueries":[{"key":"a","description":"d","query":"q"}]}`}, nil
	}
	db.Mocks.Settings.CreateIfUpToDate = func(ctx context.Context, subject api.ConfigurationSubject, _ *int32, authorUserID int32, contents string) (latestSetting *api.Settings, err error) {
		calledSettingsCreateIfUpToDate = true
		return &api.Settings{ID: lastID + 1, Contents: `not used`}, nil
	}

	mockConfigurationCascadeSubjects = func() ([]*configurationSubject, error) {
		return []*configurationSubject{subject}, nil
	}
	defer func() { mockConfigurationCascadeSubjects = nil }()

	mutation, err := (&schemaResolver{}).ConfigurationMutation(ctx, &struct {
		Input *configurationMutationGroupInput
	}{Input: &configurationMutationGroupInput{LastID: &lastID, Subject: subject.user.ID()}})
	if err != nil {
		t.Fatal(err)
	}
	created, err := mutation.CreateSavedQuery(ctx, &struct {
		Description                      string
		Query                            string
		ShowOnHomepage                   bool
		Notify                           bool
		NotifySlack                      bool
		DisableSubscriptionNotifications bool
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
		query:       "q2",
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
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
	lastID := int32(5)
	subject := &configurationSubject{user: &UserResolver{user: &types.User{ID: uid}}}
	newDescription := "d2"

	defer resetMocks()
	db.Mocks.Users.MockGetByID_Return(t, &types.User{ID: uid}, nil)
	calledSettingsGetLatest := false
	calledSettingsCreateIfUpToDate := false
	db.Mocks.Settings.GetLatest = func(ctx context.Context, subject api.ConfigurationSubject) (*api.Settings, error) {
		calledSettingsGetLatest = true
		if calledSettingsCreateIfUpToDate {
			return &api.Settings{ID: lastID + 1, Contents: `{"search.savedQueries":[{"key":"a","description":"d2","query":"q"}]}`}, nil
		}
		return &api.Settings{ID: lastID, Contents: `{"search.savedQueries":[{"key":"a","description":"d","query":"q"}]}`}, nil
	}
	db.Mocks.Settings.CreateIfUpToDate = func(ctx context.Context, subject api.ConfigurationSubject, _ *int32, authorUserID int32, contents string) (latestSetting *api.Settings, err error) {
		calledSettingsCreateIfUpToDate = true
		return &api.Settings{ID: lastID + 1, Contents: `not used`}, nil
	}

	mockConfigurationCascadeSubjects = func() ([]*configurationSubject, error) {
		return []*configurationSubject{subject}, nil
	}
	defer func() { mockConfigurationCascadeSubjects = nil }()

	mutation, err := (&schemaResolver{}).ConfigurationMutation(ctx, &struct {
		Input *configurationMutationGroupInput
	}{Input: &configurationMutationGroupInput{LastID: &lastID, Subject: subject.user.ID()}})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := mutation.UpdateSavedQuery(ctx, &struct {
		ID             graphql.ID
		Description    *string
		Query          *string
		ShowOnHomepage bool
		Notify         bool
		NotifySlack    bool
	}{
		ID:          marshalSavedQueryID(api.SavedQueryIDSpec{Subject: subject.toSubject(), Key: "a"}),
		Description: &newDescription,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := &savedQueryResolver{
		key:            "a",
		subject:        subject,
		index:          0,
		description:    "d2",
		query:          "q",
		showOnHomepage: false,
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
