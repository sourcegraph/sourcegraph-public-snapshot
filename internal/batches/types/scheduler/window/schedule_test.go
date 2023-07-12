package window

import (
	"testing"
	"time"
)

func TestScheduleLimited(t *testing.T) {
	t.Parallel()

	base := time.Now()
	rate := rate{n: 100, unit: ratePerSecond}
	schedule := newSchedule(base, 1*time.Minute, rate)

	t.Run("Take", func(t *testing.T) {
		// We don't want to block the tests for any real length of time, but we
		// do want to validate that some sort of rate limiting is occurring.
		// Given the rate we set up, it _should_ take at least 10 ms to take two
		// slots out of the schedule (since the first Take() will be more or
		// less instant, and then the second should be 1/100 seconds later).
		if testing.Short() {
			t.Skip("Take tests blocking behaviour, and is therefore not necessarily fast")
		}

		start := time.Now()
		first, err := schedule.Take()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		second, err := schedule.Take()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		end := time.Now()

		if !end.After(start) {
			t.Fatalf("something funky is happening with the clock, as the end time is not after the start time: start=%v end=%v", start, end)
		}

		if !first.Before(second) {
			t.Errorf("Take return values are not sequential: first=%v second=%v", first, second)
		}

		if duration := end.Sub(start); duration < 10*time.Millisecond {
			t.Errorf("duration was less than the expected 10ms: %v", duration)
		}
	})

	t.Run("ValidUntil", func(t *testing.T) {
		have := schedule.ValidUntil()
		want := base.Add(1 * time.Minute)
		if have != want {
			t.Errorf("unexpected validity: have=%v want=%v", have, want)
		}
	})

	t.Run("total", func(t *testing.T) {
		have := schedule.total()
		want := 100 * 60
		if have != want {
			t.Errorf("unexpected total: have=%v want=%v", have, want)
		}
	})
}

func TestScheduleUnlimited(t *testing.T) {
	t.Parallel()

	base := time.Now()
	schedule := newSchedule(base, 1*time.Minute, makeUnlimitedRate())

	t.Run("Take", func(t *testing.T) {
		// There isn't really a sensible way to validate that no blocking occurs
		// here, so we'll just validate that the return value seems sensible.
		have, err := schedule.Take()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		} else if have.Before(base) {
			t.Errorf("unexpected take time before base: %v", have)
		}
	})

	t.Run("ValidUntil", func(t *testing.T) {
		have := schedule.ValidUntil()
		want := base.Add(1 * time.Minute)
		if have != want {
			t.Errorf("unexpected validity: have=%v want=%v", have, want)
		}
	})

	t.Run("total", func(t *testing.T) {
		have := schedule.total()
		want := -1
		if have != want {
			t.Errorf("unexpected total: have=%v want=%v", have, want)
		}
	})
}

func TestScheduleZero(t *testing.T) {
	t.Parallel()

	base := time.Now()
	schedule := newSchedule(base, 1*time.Minute, rate{n: 0})

	t.Run("Take", func(t *testing.T) {
		_, err := schedule.Take()
		if err != ErrZeroSchedule {
			t.Errorf("unexpected error: have=%v want=%v", err, ErrZeroSchedule)
		}
	})

	t.Run("ValidUntil", func(t *testing.T) {
		have := schedule.ValidUntil()
		want := base.Add(1 * time.Minute)
		if have != want {
			t.Errorf("unexpected validity: have=%v want=%v", have, want)
		}
	})

	t.Run("total", func(t *testing.T) {
		have := schedule.total()
		want := 0
		if have != want {
			t.Errorf("unexpected total: have=%v want=%v", have, want)
		}
	})
}
