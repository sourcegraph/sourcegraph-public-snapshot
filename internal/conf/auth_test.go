package conf

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
)

func TestAuthPublic(t *testing.T) {
	orig := envvar.SourcegraphDotComMode()
	envvar.MockSourcegraphDotComMode(false)
	defer envvar.MockSourcegraphDotComMode(orig) // reset

	t.Run("Default, self-hosted instance non-public auth", func(t *testing.T) {
		got := AuthPublic()
		want := false
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	envvar.MockSourcegraphDotComMode(true)

	t.Run("Sourcegraph.com public auth", func(t *testing.T) {
		got := AuthPublic()
		want := true
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}
