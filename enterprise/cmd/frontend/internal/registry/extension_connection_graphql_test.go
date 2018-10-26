package registry

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestFilteringExtensionIDs(t *testing.T) {
	t.Run("filterStripLocalExtensionIDs on localhost", func(t *testing.T) {
		conf.Mock(&schema.SiteConfiguration{AppURL: "http://localhost:3080"})
		defer conf.Mock(nil)
		input := []string{"localhost:3080/owner1/name1", "owner2/name2"}
		want := []string{"owner1/name1"}
		got := filterStripLocalExtensionIDs(input)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})
	t.Run("filterStripLocalExtensionIDs on Sourcegraph.com", func(t *testing.T) {
		oldAppURL := globals.AppURL
		globals.AppURL = &url.URL{Scheme: "https", Host: "sourcegraph.com"}
		defer func() { globals.AppURL = oldAppURL }()
		input := []string{"localhost:3080/owner1/name1", "owner2/name2"}
		want := []string{"owner2/name2"}
		got := filterStripLocalExtensionIDs(input)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})
}
