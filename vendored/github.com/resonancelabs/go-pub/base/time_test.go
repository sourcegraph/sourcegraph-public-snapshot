package base

import (
	. "testing"
	"time"
)

func TestMicros(t *T) {
	// Get an arbitrary time.Time value.
	now := time.Unix(1382735441, 123456789)

	// Convert to and from Micros.
	mVal := ToMicros(now)

	// Check the basic idea.
	var i int64
	i = mVal.Int64() // mainly testing compilation here.
	if i != (now.UnixNano() / 1000) {
		t.Error("very broken")
	}

	// Check roundtrips.
	tVal := mVal.ToTime()
	if tVal.UnixNano() == now.UnixNano() {
		t.Error("should have lost 789 nanoseconds")
	}
	newMVal := ToMicros(tVal)
	if mVal != newMVal {
		t.Error("roundtrips starting with Micros should preserve information")
	}
}

func TestEarliest(t *T) {
	n := time.Now()
	h := n.Add(time.Hour)
	d := n.Add(24 * time.Hour)

	if Earliest(h, d) != h {
		t.Error("should have returned 1 hour ahead")
	}
}
