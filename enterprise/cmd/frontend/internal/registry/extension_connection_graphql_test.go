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
