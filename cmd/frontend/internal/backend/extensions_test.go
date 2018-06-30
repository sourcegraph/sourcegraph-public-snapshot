package backend

import (
	"context"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/pkg/registry"
)

func TestGetExtensionByExtensionID(t *testing.T) {
	ctx := context.Background()

	t.Run("root", func(t *testing.T) {
		mockLocalRegistryExtensionIDPrefix = &strnilptr
		defer func() { mockLocalRegistryExtensionIDPrefix = nil }()

		t.Run("2-part", func(t *testing.T) {
			db.Mocks.RegistryExtensions.GetByExtensionID = func(extensionID string) (*db.RegistryExtension, error) {
				return &db.RegistryExtension{ID: 1, Name: "b"}, nil
			}
			defer func() { db.Mocks = db.MockStores{} }()
			local, remote, err := GetExtensionByExtensionID(ctx, "a/b")
			if err != nil {
				t.Fatal(err)
			}
			if want := (&db.RegistryExtension{ID: 1, Name: "b"}); !reflect.DeepEqual(local, want) {
				t.Errorf("got %+v, want %+v", local, want)
			}
			if remote != nil {
				t.Error()
			}
		})

		t.Run("3-part", func(t *testing.T) {
			if _, _, err := GetExtensionByExtensionID(ctx, "a/b/c"); err == nil {
				t.Fatal()
			}
		})
	})

	t.Run("non-root", func(t *testing.T) {
		mockLocalRegistryExtensionIDPrefix = strptrptr("x")
		defer func() { mockLocalRegistryExtensionIDPrefix = nil }()

		t.Run("2-part", func(t *testing.T) {
			mockGetRemoteRegistryExtension = func(field, value string) (*registry.Extension, error) {
				if want := "extensionID"; field != want {
					t.Errorf("got field %q, want %q", field, want)
				}
				if want := "a/b"; value != want {
					t.Errorf("got value %q, want %q", value, want)
				}
				return &registry.Extension{UUID: "u", ExtensionID: "a/b"}, nil
			}
			defer func() { mockGetRemoteRegistryExtension = nil }()
			local, remote, err := GetExtensionByExtensionID(ctx, "a/b")
			if err != nil {
				t.Fatal(err)
			}
			if local != nil {
				t.Error()
			}
			if want := (&registry.Extension{UUID: "u", ExtensionID: "a/b"}); !reflect.DeepEqual(remote, want) {
				t.Errorf("got %+v, want %+v", remote, want)
			}
		})

		t.Run("3-part", func(t *testing.T) {
			db.Mocks.RegistryExtensions.GetByExtensionID = func(extensionID string) (*db.RegistryExtension, error) {
				return &db.RegistryExtension{ID: 1, Name: "b", NonCanonicalExtensionID: "b/c"}, nil
			}
			defer func() { db.Mocks = db.MockStores{} }()
			local, remote, err := GetExtensionByExtensionID(ctx, "x/b/c")
			if err != nil {
				t.Fatal(err)
			}
			if want := (&db.RegistryExtension{ID: 1, Name: "b", NonCanonicalExtensionID: "x/b/c", NonCanonicalRegistry: "x"}); !reflect.DeepEqual(local, want) {
				t.Errorf("got %+v, want %+v", local, want)
			}
			if remote != nil {
				t.Error()
			}
		})
	})

	t.Run("invalid extension ID", func(t *testing.T) {
		if _, _, err := GetExtensionByExtensionID(ctx, "a/b/c/d"); err == nil {
			t.Fatal()
		}
	})
}

var strnilptr *string

func strptrptr(s string) **string {
	tmp := &s
	return &tmp
}
