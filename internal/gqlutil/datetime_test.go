package gqlutil

import (
	"testing"
	"time"
)

func TestDateTime(t *testing.T) {
	t0 := time.Unix(123456789, 0).UTC()
	t.Run("marshal", func(t *testing.T) {
		v := DateTime{Time: t0}
		if got, err := v.MarshalJSON(); err != nil {
			t.Fatal(err)
		} else if want := `"1973-11-29T21:33:09Z"`; string(got) != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
	t.Run("unmarshal", func(t *testing.T) {
		var got DateTime
		if err := got.UnmarshalGraphQL("1973-11-29T21:33:09Z"); err != nil {
			t.Fatal(err)
		}
		if want := (DateTime{Time: t0}); !got.Time.Equal(want.Time) {
			t.Errorf("got %v, want %v", got.Time, want.Time)
		}
	})
}
