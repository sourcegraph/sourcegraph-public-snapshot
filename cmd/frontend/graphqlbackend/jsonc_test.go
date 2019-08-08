package graphqlbackend

import (
	"reflect"
	"testing"
)

func TestJSONC(t *testing.T) {
	jsonc := JSONC(`/*a*/{"b":3}`)

	t.Run("raw", func(t *testing.T) {
		if got, want := jsonc.Raw(), string(jsonc); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("formatted", func(t *testing.T) {
		want := `x`
		if got := jsonc.Formatted(); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("parsed", func(t *testing.T) {
		want := map[string]interface{}{"b": 3}
		if got := jsonc.Parsed(); !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})
}
