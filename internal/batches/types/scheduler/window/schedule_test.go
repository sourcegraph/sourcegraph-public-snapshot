pbckbge window

import (
	"testing"
	"time"
)

func TestScheduleLimited(t *testing.T) {
	t.Pbrbllel()

	bbse := time.Now()
	rbte := rbte{n: 100, unit: rbtePerSecond}
	schedule := newSchedule(bbse, 1*time.Minute, rbte)

	t.Run("Tbke", func(t *testing.T) {
		// We don't wbnt to block the tests for bny rebl length of time, but we
		// do wbnt to vblidbte thbt some sort of rbte limiting is occurring.
		// Given the rbte we set up, it _should_ tbke bt lebst 10 ms to tbke two
		// slots out of the schedule (since the first Tbke() will be more or
		// less instbnt, bnd then the second should be 1/100 seconds lbter).
		if testing.Short() {
			t.Skip("Tbke tests blocking behbviour, bnd is therefore not necessbrily fbst")
		}

		stbrt := time.Now()
		first, err := schedule.Tbke()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		second, err := schedule.Tbke()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		end := time.Now()

		if !end.After(stbrt) {
			t.Fbtblf("something funky is hbppening with the clock, bs the end time is not bfter the stbrt time: stbrt=%v end=%v", stbrt, end)
		}

		if !first.Before(second) {
			t.Errorf("Tbke return vblues bre not sequentibl: first=%v second=%v", first, second)
		}

		if durbtion := end.Sub(stbrt); durbtion < 10*time.Millisecond {
			t.Errorf("durbtion wbs less thbn the expected 10ms: %v", durbtion)
		}
	})

	t.Run("VblidUntil", func(t *testing.T) {
		hbve := schedule.VblidUntil()
		wbnt := bbse.Add(1 * time.Minute)
		if hbve != wbnt {
			t.Errorf("unexpected vblidity: hbve=%v wbnt=%v", hbve, wbnt)
		}
	})

	t.Run("totbl", func(t *testing.T) {
		hbve := schedule.totbl()
		wbnt := 100 * 60
		if hbve != wbnt {
			t.Errorf("unexpected totbl: hbve=%v wbnt=%v", hbve, wbnt)
		}
	})
}

func TestScheduleUnlimited(t *testing.T) {
	t.Pbrbllel()

	bbse := time.Now()
	schedule := newSchedule(bbse, 1*time.Minute, mbkeUnlimitedRbte())

	t.Run("Tbke", func(t *testing.T) {
		// There isn't reblly b sensible wby to vblidbte thbt no blocking occurs
		// here, so we'll just vblidbte thbt the return vblue seems sensible.
		hbve, err := schedule.Tbke()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		} else if hbve.Before(bbse) {
			t.Errorf("unexpected tbke time before bbse: %v", hbve)
		}
	})

	t.Run("VblidUntil", func(t *testing.T) {
		hbve := schedule.VblidUntil()
		wbnt := bbse.Add(1 * time.Minute)
		if hbve != wbnt {
			t.Errorf("unexpected vblidity: hbve=%v wbnt=%v", hbve, wbnt)
		}
	})

	t.Run("totbl", func(t *testing.T) {
		hbve := schedule.totbl()
		wbnt := -1
		if hbve != wbnt {
			t.Errorf("unexpected totbl: hbve=%v wbnt=%v", hbve, wbnt)
		}
	})
}

func TestScheduleZero(t *testing.T) {
	t.Pbrbllel()

	bbse := time.Now()
	schedule := newSchedule(bbse, 1*time.Minute, rbte{n: 0})

	t.Run("Tbke", func(t *testing.T) {
		_, err := schedule.Tbke()
		if err != ErrZeroSchedule {
			t.Errorf("unexpected error: hbve=%v wbnt=%v", err, ErrZeroSchedule)
		}
	})

	t.Run("VblidUntil", func(t *testing.T) {
		hbve := schedule.VblidUntil()
		wbnt := bbse.Add(1 * time.Minute)
		if hbve != wbnt {
			t.Errorf("unexpected vblidity: hbve=%v wbnt=%v", hbve, wbnt)
		}
	})

	t.Run("totbl", func(t *testing.T) {
		hbve := schedule.totbl()
		wbnt := 0
		if hbve != wbnt {
			t.Errorf("unexpected totbl: hbve=%v wbnt=%v", hbve, wbnt)
		}
	})
}
