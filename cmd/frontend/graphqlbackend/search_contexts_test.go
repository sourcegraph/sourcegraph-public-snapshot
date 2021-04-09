package graphqlbackend

import (
	"context"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/search/searchcontexts"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestAutoDefinedSearchContexts(t *testing.T) {
	key := int32(1)
	username := "alice"
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: key})
	db := new(dbtesting.MockDB)

	database.Mocks.Users.GetByID = func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{Username: username}, nil
	}
	defer resetMocks()

	searchContexts, err := (&schemaResolver{db: db}).AutoDefinedSearchContexts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := []*searchContextResolver{
		{sc: searchcontexts.GetGlobalSearchContext(), db: db},
		{sc: searchcontexts.GetUserSearchContext(username, key), db: db},
	}
	if !reflect.DeepEqual(searchContexts, want) {
		t.Fatalf("got %+v, want %+v", searchContexts, want)
	}

	for _, resolver := range searchContexts {
		repositories, err := resolver.Repositories(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(repositories) != 0 {
			t.Fatal("auto-defined search contexts should not return repositories")
		}
	}
}
