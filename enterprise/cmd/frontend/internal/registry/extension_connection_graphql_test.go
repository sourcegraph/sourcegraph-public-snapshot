package registry

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestFilteringExtensionIDs(t *testing.T) {
	t.Run("filterStripLocalExtensionIDs on localhost", func(t *testing.T) {
		conf.Mock(&conf.Unified{Critical: schema.CriticalConfiguration{ExternalURL: "http://localhost:3080"}})
		defer conf.Mock(nil)
		input := []string{"localhost:3080/owner1/name1", "owner2/name2"}
		want := []string{"owner1/name1"}
		got := filterStripLocalExtensionIDs(input)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})
	t.Run("filterStripLocalExtensionIDs on Sourcegraph.com", func(t *testing.T) {
		oldExternalURL := globals.ExternalURL
		globals.ExternalURL = &url.URL{Scheme: "https", Host: "sourcegraph.com"}
		defer func() { globals.ExternalURL = oldExternalURL }()
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
		want dbExtensionsListOptions
	}{
		"empty": {
			args: graphqlbackend.RegistryExtensionConnectionArgs{},
			want: dbExtensionsListOptions{ExcludeWIP: true},
		},
		"Query simple": {
			args: graphqlbackend.RegistryExtensionConnectionArgs{Query: strptr("q")},
			want: dbExtensionsListOptions{Query: "q", ExcludeWIP: true},
		},
		"PrioritizeExensionIDs": {
			args: graphqlbackend.RegistryExtensionConnectionArgs{PrioritizeExtensionIDs: strarrayptr([]string{"a", "b"})},
			want: dbExtensionsListOptions{PrioritizeExtensionIDs: []string{"a", "b"}, ExcludeWIP: true},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := toDBExtensionsListOptions(test.args)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("got %+v, want %+v", got, test.want)
			}
		})
	}
}

func strarrayptr(values []string) *[]string {
	return &values
}
