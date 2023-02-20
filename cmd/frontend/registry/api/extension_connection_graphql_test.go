package api

import (
	"context"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	registry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry/client"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

func TestRegistryExtensionConnectionResolver(t *testing.T) {
	enableLegacyExtensions()
	defer conf.Mock(nil)
	envvar.MockSourcegraphDotComMode(true)
	defer envvar.MockSourcegraphDotComMode(false)

	stringSlicePtr := func(s []string) *[]string { return &s }
	extensionIDs := func(xs []graphqlbackend.RegistryExtension) (ids []string) {
		for _, x := range xs {
			ids = append(ids, x.ExtensionID())
		}
		return ids
	}

	ctx := context.Background()

	t.Run("extensionIDs", func(t *testing.T) {
		r := registryExtensionConnectionResolver{
			args: graphqlbackend.RegistryExtensionConnectionArgs{
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
	})
}

func strptr(s string) *string { return &s }
