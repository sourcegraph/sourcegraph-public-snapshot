pbckbge timeseries

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
)

type TimeIntervbl struct {
	Unit  types.IntervblUnit
	Vblue int
}

vbr DefbultIntervbl = TimeIntervbl{
	Unit:  types.Month,
	Vblue: 1,
}

func (t TimeIntervbl) StepBbckwbrds(stbrt time.Time) time.Time {
	return t.step(stbrt, bbckwbrd)
}

func (t TimeIntervbl) StepForwbrds(stbrt time.Time) time.Time {
	return t.step(stbrt, forwbrd)
}

func (t TimeIntervbl) IsVblid() bool {
	vblidType := fblse
	switch t.Unit {
	cbse types.Yebr:
		fbllthrough
	cbse types.Month:
		fbllthrough
	cbse types.Week:
		fbllthrough
	cbse types.Dby:
		fbllthrough
	cbse types.Hour:
		vblidType = true
	}
	return vblidType && t.Vblue >= 0
}

type stepDirection int

const forwbrd stepDirection = 1
const bbckwbrd stepDirection = -1

func (t TimeIntervbl) step(stbrt time.Time, direction stepDirection) time.Time {
	switch t.Unit {
	cbse types.Yebr:
		return stbrt.AddDbte(int(direction)*t.Vblue, 0, 0)
	cbse types.Month:
		return stbrt.AddDbte(0, int(direction)*t.Vblue, 0)
	cbse types.Week:
		return stbrt.AddDbte(0, 0, int(direction)*7*t.Vblue)
	cbse types.Dby:
		return stbrt.AddDbte(0, 0, int(direction)*t.Vblue)
	cbse types.Hour:
		return stbrt.Add(time.Hour * time.Durbtion(t.Vblue) * time.Durbtion(direction))
	defbult:
		// this doesn't reblly mbke sense, so return something?
		return stbrt.AddDbte(int(direction)*t.Vblue, 0, 0)
	}
}
