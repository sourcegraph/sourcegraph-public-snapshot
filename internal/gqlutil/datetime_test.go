package gqlutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var t0 = time.Unix(123456789, 0).UTC()

func TestDateTime(t *testing.T) {
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

func TestDateTimeOrNil(t *testing.T) {
	tests := map[string]struct {
		timePtr *time.Time
		want    *DateTime
	}{
		"Nil time pointer input": {
			timePtr: nil,
			want:    nil,
		},
		"Non-nil time pointer input": {
			timePtr: &t0,
			want:    &DateTime{Time: t0},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := DateTimeOrNil(test.timePtr)
			require.Equal(t, test.want, got)
		})
	}
}

func TestFromTime(t *testing.T) {
	var zeroTime time.Time
	tests := map[string]struct {
		inputTime time.Time
		want      *DateTime
	}{
		"Zero time input": {
			inputTime: zeroTime,
			want:      nil,
		},
		"Non-zero time input": {
			inputTime: t0,
			want:      &DateTime{Time: t0},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := FromTime(test.inputTime)
			require.Equal(t, test.want, got)
		})
	}
}
