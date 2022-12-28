package registry

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/registry/stores"
)

func TestFilteringExtensionIDs(t *testing.T) {
	t.Run("filterStripLocalExtensionIDs on localhost", func(t *testing.T) {
		before := globals.ExternalURL()
		globals.SetExternalURL(&url.URL{Scheme: "http", Host: "localhost:3080"})
		defer globals.SetExternalURL(before)

		input := []string{"localhost:3080/owner1/name1", "owner2/name2"}
		want := []string{"owner1/name1"}
		got := filterStripLocalExtensionIDs(input)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})
	t.Run("filterStripLocalExtensionIDs on Sourcegraph.com", func(t *testing.T) {
		orig := envvar.SourcegraphDotComMode()
		envvar.MockSourcegraphDotComMode(true)
		defer envvar.MockSourcegraphDotComMode(orig) // reset

		input := []string{"localhost:3080/owner1/name1", "owner2/name2"}
		want := []string{"owner2/name2"}
		got := filterStripLocalExtensionIDs(input)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})
}

func TestToDBExtensionsListOptions(t *testing.T) {
	tests := map[string]struct {
		args graphqlbackend.RegistryExtensionConnectionArgs
		want stores.ExtensionsListOptions
	}{
		"empty": {
			args: graphqlbackend.RegistryExtensionConnectionArgs{},
			want: stores.ExtensionsListOptions{},
		},
		"ExensionIDs": {
			args: graphqlbackend.RegistryExtensionConnectionArgs{ExtensionIDs: strarrayptr([]string{"a", "b"})},
			want: stores.ExtensionsListOptions{ExtensionIDs: []string{"a", "b"}},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := toDBExtensionsListOptions(test.args)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("got %+v, want %+v", got, test.want)
			}
		})
	}
}

func strarrayptr(values []string) *[]string {
	return &values
}
