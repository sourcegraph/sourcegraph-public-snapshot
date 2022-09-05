package jobutil

import (
	"strings"
	"testing"
	"time"
)

func TestLonger(t *testing.T) {
	N := 2
	noise := time.Nanosecond
	for dt := time.Millisecond + noise; dt < time.Hour; dt += time.Millisecond {
		dt2 := longer(N, dt)
		if dt2 < time.Duration(N)*dt {
			t.Fatalf("longer(%v)=%v < 2*%v, want more", dt, dt2, dt)
		}
		if strings.Contains(dt2.String(), ".") {
			t.Fatalf("longer(%v).String() = %q contains an unwanted decimal point, want a nice round duration", dt, dt2)
		}
		lowest := 2 * time.Second
		if dt2 < lowest {
			t.Fatalf("longer(%v) = %v < %s, too short", dt, dt2, lowest)
		}
	}
}
