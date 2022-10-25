package api

import (
	"context"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	registry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry/client"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func TestRegistryExtensionConnectionResolver(t *testing.T) {
	enableLegacyExtensions()
	defer conf.Mock(nil)

	int32Ptr := func(v int32) *int32 { return &v }
	stringSlicePtr := func(s []string) *[]string { return &s }
	extensionIDs := func(xs []graphqlbackend.RegistryExtension) (ids []string) {
		for _, x := range xs {
			ids = append(ids, x.ExtensionID())
		}
		return ids
	}

	ListLocalRegistryExtensions = func(context.Context, database.DB, graphqlbackend.RegistryExtensionConnectionArgs) ([]graphqlbackend.RegistryExtension, error) {
		return nil, nil
	}
	defer func() {
		ListLocalRegistryExtensions = nil
	}()

	ctx := context.Background()

	t.Run("first, prioritizeExtensionIDs", func(t *testing.T) {
		r := registryExtensionConnectionResolver{
			args: graphqlbackend.RegistryExtensionConnectionArgs{
				ConnectionArgs:         graphqlutil.ConnectionArgs{First: int32Ptr(3)},
				Remote:                 true,
				Local:                  true,
				PrioritizeExtensionIDs: stringSlicePtr([]string{"a/z1"}),
			},
			listRemoteRegistryExtensions: func(_ context.Context, query string) ([]*registry.Extension, error) {
				return []*registry.Extension{
					{ExtensionID: "a/wip0", Manifest: strptr(`{"wip": true}`)},
					{ExtensionID: "a/z0", Manifest: strptr(`{}`)},
					{ExtensionID: "a/z1", Manifest: strptr(`{}`)},
					{ExtensionID: "z/z", Manifest: strptr(`{"wip": true}`)},
				}, nil
			},
		}

		nodes, err := r.Nodes(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if ids, want := extensionIDs(nodes), []string{"a/z1", "a/z0", "a/wip0"}; !reflect.DeepEqual(ids, want) {
			t.Errorf("got ids %v, want %v", ids, want)
		}

		totalCount, err := r.TotalCount(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if want := int32(4); totalCount != want {
			t.Errorf("got totalCount %d, want %d", totalCount, want)
		}
	})

	t.Run("extensionIDs", func(t *testing.T) {
		r := registryExtensionConnectionResolver{
			args: graphqlbackend.RegistryExtensionConnectionArgs{
				Remote:       true,
				Local:        true,
				ExtensionIDs: stringSlicePtr([]string{"a/a"}),
			},
			listRemoteRegistryExtensions: func(_ context.Context, query string) ([]*registry.Extension, error) {
				return []*registry.Extension{
					{ExtensionID: "a/b", Manifest: strptr(`{}`)},
					{ExtensionID: "a/a", Manifest: strptr(`{}`)},
				}, nil
			},
		}

		nodes, err := r.Nodes(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if ids, want := extensionIDs(nodes), []string{"a/a"}; !reflect.DeepEqual(ids, want) {
			t.Errorf("got ids %v, want %v", ids, want)
		}

		totalCount, err := r.TotalCount(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if want := int32(1); totalCount != want {
			t.Errorf("got totalCount %d, want %d", totalCount, want)
		}
	})
}
