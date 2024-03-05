package conf

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/dotcom"
)

func TestAuthPublic(t *testing.T) {
	orig := dotcom.SourcegraphDotComMode()
	dotcom.MockSourcegraphDotComMode(false)
	defer dotcom.MockSourcegraphDotComMode(orig) // reset

	t.Run("Default, self-hosted instance non-public auth", func(t *testing.T) {
		got := AuthPublic()
		want := false
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	dotcom.MockSourcegraphDotComMode(true)

	t.Run("Sourcegraph.com public auth", func(t *testing.T) {
		got := AuthPublic()
		want := true
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}
