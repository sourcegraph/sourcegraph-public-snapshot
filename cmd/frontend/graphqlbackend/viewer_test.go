package graphqlbackend

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestViewerCanChangeLibraryItemVisibilityToPublic(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		got, err := (&schemaResolver{db: db}).ViewerCanChangeLibraryItemVisibilityToPublic(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if want := false; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("authenticated as site admin", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		got, err := (&schemaResolver{db: db}).ViewerCanChangeLibraryItemVisibilityToPublic(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if want := true; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(nil, database.ErrNoCurrentUser)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)

		ctx := actor.WithActor(context.Background(), &actor.Actor{})
		got, err := (&schemaResolver{db: db}).ViewerCanChangeLibraryItemVisibilityToPublic(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if want := false; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}
