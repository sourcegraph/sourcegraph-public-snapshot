package graphqlbackend

import (
	"reflect"
	"testing"
)

func TestJSONC(t *testing.T) {
	jsonc := JSONC(`/*a*/{"b":"c"}`)

	t.Run("raw", func(t *testing.T) {
		if got, want := jsonc.Raw(), JSONCString(jsonc); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("formatted", func(t *testing.T) {
		want := "/*a*/ {\n  \"b\": \"c\"\n}"
		if got, err := jsonc.Formatted(); err != nil {
			t.Fatal(err)
		} else if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("parsed", func(t *testing.T) {
		want := map[string]interface{}{"b": "c"}
		if got, err := jsonc.Parsed(); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(got.Value, want) {
			t.Errorf("got %+v, want %+v", got.Value, want)
		}
	})
}
