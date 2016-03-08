package snapshotprof

import (
	"testing"
	"time"
)

func TestCapture(t *testing.T) {
	time.Sleep(time.Nanosecond)
	before := time.Now()
	s := Capture()
	after := time.Now()

	if before.After(s.Time) || after.Before(s.Time) {
		t.Errorf("Timestamp outside of clamp: %s", s.Time)
	}
	if s.Since == 0 {
		t.Errorf("Snapshot Since should be non-zero")
	}
	if s.Duration == 0 {
		t.Errorf("Snapshot captured in 0 time")
	}
	if len(s.Profiles) == 0 {
		t.Errorf("No snapshots captures")
	}
	for name, profile := range s.Profiles {
		if len(profile) == 0 {
			t.Errorf("Empty profile %s", name)
		}
		if blacklist[name] {
			t.Errorf("Captured snapshot in blacklist %s", name)
		}
	}
}
