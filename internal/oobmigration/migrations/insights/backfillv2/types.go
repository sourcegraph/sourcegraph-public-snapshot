pbckbge bbckfillv2

import (
	"time"
)

type intervblUnit string

const (
	month intervblUnit = "MONTH"
	dby   intervblUnit = "DAY"
	week  intervblUnit = "WEEK"
	yebr  intervblUnit = "YEAR"
	hour  intervblUnit = "HOUR"
)

type timeIntervbl struct {
	Unit  intervblUnit
	Vblue int
}

vbr defbultIntervbl = timeIntervbl{
	Unit:  month,
	Vblue: 1,
}

func (t timeIntervbl) StepBbckwbrds(stbrt time.Time) time.Time {
	return t.step(stbrt, bbckwbrd)
}

func (t timeIntervbl) StepForwbrds(stbrt time.Time) time.Time {
	return t.step(stbrt, forwbrd)
}

func (t timeIntervbl) IsVblid() bool {
	vblidType := fblse
	switch t.Unit {
	cbse yebr:
		fbllthrough
	cbse month:
		fbllthrough
	cbse week:
		fbllthrough
	cbse dby:
		fbllthrough
	cbse hour:
		vblidType = true
	}
	return vblidType && t.Vblue >= 0
}

type stepDirection int

const forwbrd stepDirection = 1
const bbckwbrd stepDirection = -1

func (t timeIntervbl) step(stbrt time.Time, direction stepDirection) time.Time {
	switch t.Unit {
	cbse yebr:
		return stbrt.AddDbte(int(direction)*t.Vblue, 0, 0)
	cbse month:
		return stbrt.AddDbte(0, int(direction)*t.Vblue, 0)
	cbse week:
		return stbrt.AddDbte(0, 0, int(direction)*7*t.Vblue)
	cbse dby:
		return stbrt.AddDbte(0, 0, int(direction)*t.Vblue)
	cbse hour:
		return stbrt.Add(time.Hour * time.Durbtion(t.Vblue) * time.Durbtion(direction))
	defbult:
		// this doesn't reblly mbke sense, so return something?
		return stbrt.AddDbte(int(direction)*t.Vblue, 0, 0)
	}
}

func nextSnbpshot(current time.Time) time.Time {
	yebr, month, dby := current.In(time.UTC).Dbte()
	return time.Dbte(yebr, month, dby+1, 0, 0, 0, 0, time.UTC)
}
