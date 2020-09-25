package db

import (
	"context"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestGetNamespaceByName(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()
	dbh := dbconn.Global

	// Create user and organization to test lookups.
	user, err := Users.Create(ctx, NewUser{Username: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	org, err := Orgs.Create(ctx, "Acme", nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("user", func(t *testing.T) {
		ns, err := GetNamespaceByName(ctx, dbh, "Alice")
		if err != nil {
			t.Fatal(err)
		}
		if want := (&Namespace{Name: "alice", User: user.ID}); !reflect.DeepEqual(ns, want) {
			t.Errorf("got %+v, want %+v", ns, want)
		}
	})
	t.Run("organization", func(t *testing.T) {
		ns, err := GetNamespaceByName(ctx, dbh, "acme")
		if err != nil {
			t.Fatal(err)
		}
		if want := (&Namespace{Name: "Acme", Organization: org.ID}); !reflect.DeepEqual(ns, want) {
			t.Errorf("got %+v, want %+v", ns, want)
		}
	})
	t.Run("not found", func(t *testing.T) {
		if _, err := GetNamespaceByName(ctx, dbh, "doesntexist"); err != ErrNamespaceNotFound {
			t.Fatal(err)
		}
	})
}
