package graphqlbackend

import (
	"context"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestSearchContexts(t *testing.T) {
	key := int32(1)
	username := "alice"
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: key})

	orig := envvar.SourcegraphDotComMode()
	envvar.MockSourcegraphDotComMode(true)
	defer envvar.MockSourcegraphDotComMode(orig)

	database.Mocks.Users.GetByID = func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{Username: username}, nil
	}
	defer resetMocks()

	searchContexts, err := (&schemaResolver{}).SearchContexts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := []*searchContextResolver{
		{sc: types.SearchContext{Name: "global"}},
		{sc: types.SearchContext{Name: username, UserID: &key}},
	}
	if !reflect.DeepEqual(searchContexts, want) {
		t.Errorf("got %v+, want %v+", searchContexts, want)
	}
}
