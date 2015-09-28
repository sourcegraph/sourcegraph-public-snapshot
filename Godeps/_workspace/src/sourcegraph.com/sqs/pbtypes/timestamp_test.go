package pbtypes

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTimestamp(t *testing.T) {
	tm := time.Now()
	ts := NewTimestamp(tm)
	tm2 := ts.Time()
	if !tm2.Equal(tm) {
		t.Errorf("got %q, want %q", tm2, tm)
	}
}

func TestTimestamp_JSON(t *testing.T) {
	ts := NewTimestamp(time.Now())
	b, err := json.Marshal(ts)
	if err != nil {
		t.Fatal(err)
	}

	var ts2 Timestamp
	if err := json.Unmarshal(b, &ts2); err != nil {
		t.Fatal(err)
	}

	if ts2 != ts {
		t.Errorf("got %v, want %v", ts2, ts)
	}
}
