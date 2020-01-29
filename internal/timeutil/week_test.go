package timeutil

import (
	"testing"
	"time"
)

func TestWeek_StartOfWeek(t *testing.T) {
	want := time.Date(2020, 1, 19, 0, 0, 0, 0, time.UTC)

	got := StartOfWeek(time.Date(2020, 1, 19, 5, 30, 10, 0, time.UTC), 0)
	if !want.Equal(got) {
		t.Fatalf("got %s, want %s", got.Format(time.RFC3339), want.Format(time.RFC3339))
	}

	got = StartOfWeek(time.Date(2020, 1, 23, 0, 0, 0, 0, time.UTC), 0)
	if !want.Equal(got) {
		t.Fatalf("got %s, want %s", got.Format(time.RFC3339), want.Format(time.RFC3339))
	}

	got = StartOfWeek(time.Date(2020, 1, 25, 23, 59, 59, 0, time.UTC), 0)
	if !want.Equal(got) {
		t.Fatalf("got %s, want %s", got.Format(time.RFC3339), want.Format(time.RFC3339))
	}

	got = StartOfWeek(time.Date(2020, 1, 28, 0, 0, 0, 0, time.UTC), 1)
	if !want.Equal(got) {
		t.Fatalf("got %s, want %s", got.Format(time.RFC3339), want.Format(time.RFC3339))
	}

	got = StartOfWeek(time.Date(2021, 1, 19, 0, 0, 0, 0, time.UTC), 52)
	if !want.Equal(got) {
		t.Fatalf("got %s, want %s", got.Format(time.RFC3339), want.Format(time.RFC3339))
	}
}
