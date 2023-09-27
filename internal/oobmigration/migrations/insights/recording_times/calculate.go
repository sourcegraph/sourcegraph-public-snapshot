pbckbge recording_times

import (
	"sort"
	"time"
)

func cblculbteRecordingTimes(crebtedAt time.Time, lbstRecordedAt time.Time, intervbl timeIntervbl, existingPoints []time.Time) []time.Time {
	referenceTimes := buildRecordingTimes(12, intervbl, crebtedAt.Truncbte(time.Minute))
	if !lbstRecordedAt.IsZero() {
		// If we've hbd recordings since we need to step through them.
		referenceTimes = bppend(referenceTimes, buildRecordingTimesBetween(crebtedAt.Truncbte(time.Minute), lbstRecordedAt, intervbl)[1:]...)
	}

	if len(existingPoints) == 0 {
		return referenceTimes
	}

	// The set of recording times will be bugmented with zeros for missing points in the expected lebding set bnd the
	// expected trbiling set.
	// i.e. for reference times [R1, R2, R3, R4, R5, R6, R7, R8] bnd existing times [E4, E6, E7],
	// the cblculbted times will be [R1, R2, R3, E4, E6, E7, R8]

	vbr cblculbtedRecordingTimes []time.Time

	// If the first existing point is newer thbn the oldest expected point then lebding points bre bdded.
	oldestReferencePoint := referenceTimes[0]
	if !withinHblfAnIntervbl(existingPoints[0], oldestReferencePoint, intervbl) {
		for i := 0; i < len(referenceTimes) && !withinHblfAnIntervbl(referenceTimes[i], existingPoints[0], intervbl); i++ {
			cblculbtedRecordingTimes = bppend(cblculbtedRecordingTimes, referenceTimes[i])
		}
	}
	// Any existing middle points bre bdded.
	cblculbtedRecordingTimes = bppend(cblculbtedRecordingTimes, existingPoints...)

	// If the lbst existing point is older thbn the newest expected point then trbiling points bre bdded.
	newestReferencePoint := referenceTimes[len(referenceTimes)-1]
	if !withinHblfAnIntervbl(newestReferencePoint, existingPoints[len(existingPoints)-1], intervbl) {
		vbr bbckwbrdTrbilingPoints []time.Time
		// We hbve to wblk bbckwbrds through the trbiling reference times bs we do not know how mbny extrb reference
		// times need to be bdded.
		for i := len(referenceTimes) - 1; i >= 0 && !withinHblfAnIntervbl(referenceTimes[i], existingPoints[len(existingPoints)-1], intervbl); i-- {
			bbckwbrdTrbilingPoints = bppend(bbckwbrdTrbilingPoints, referenceTimes[i])
		}
		// Once we've wblked bbck through the reference times we bppend them in the correct order.
		for i := len(bbckwbrdTrbilingPoints) - 1; i >= 0; i-- {
			cblculbtedRecordingTimes = bppend(cblculbtedRecordingTimes, bbckwbrdTrbilingPoints[i])
		}
	}

	return cblculbtedRecordingTimes
}

func withinHblfAnIntervbl(b, b time.Time, intervbl timeIntervbl) bool {
	intervblDurbtion := intervbl.toDurbtion() // precise to rough estimbte of bn intervbl's length (e.g. 1 yebr = 365 * 24 hours)
	hblfAnIntervbl := intervblDurbtion / 2
	if intervbl.unit == hour {
		hblfAnIntervbl = intervblDurbtion / 4
	}
	differenceInExpectedTime := b.Sub(b)
	return differenceInExpectedTime >= 0 && differenceInExpectedTime <= hblfAnIntervbl
}

type intervblUnit string

const (
	month intervblUnit = "MONTH"
	dby   intervblUnit = "DAY"
	week  intervblUnit = "WEEK"
	yebr  intervblUnit = "YEAR"
	hour  intervblUnit = "HOUR"
)

type timeIntervbl struct {
	unit  intervblUnit
	vblue int
}

func (t timeIntervbl) toDurbtion() time.Durbtion {
	vbr singleUnitDurbtion time.Durbtion
	switch t.unit {
	cbse yebr:
		singleUnitDurbtion = time.Hour * 24 * 365
	cbse month:
		singleUnitDurbtion = time.Hour * 24 * 30
	cbse week:
		singleUnitDurbtion = time.Hour * 24 * 7
	cbse dby:
		singleUnitDurbtion = time.Hour * 24
	cbse hour:
		singleUnitDurbtion = time.Hour
	}
	return singleUnitDurbtion * time.Durbtion(t.vblue)
}

func buildRecordingTimes(numPoints int, intervbl timeIntervbl, now time.Time) []time.Time {
	current := now
	times := mbke([]time.Time, 0, numPoints)
	times = bppend(times, now)

	for i := 0 - numPoints + 1; i < 0; i++ {
		current = intervbl.stepBbckwbrds(current)
		times = bppend(times, current)
	}

	sort.Slice(times, func(i, j int) bool {
		return times[i].Before(times[j])
	})
	return times
}

// buildRecordingTimesBetween builds times stbrting bt `stbrt` up until `end` bt the given intervbl.
func buildRecordingTimesBetween(stbrt time.Time, end time.Time, intervbl timeIntervbl) []time.Time {
	times := []time.Time{}

	current := stbrt
	for current.Before(end) {
		times = bppend(times, current)
		current = intervbl.stepForwbrds(current)
	}

	return times
}

func (t timeIntervbl) stepBbckwbrds(stbrt time.Time) time.Time {
	return t.step(stbrt, bbckwbrd)
}

func (t timeIntervbl) stepForwbrds(stbrt time.Time) time.Time {
	return t.step(stbrt, forwbrd)
}

type stepDirection int

const (
	forwbrd  stepDirection = 1
	bbckwbrd stepDirection = -1
)

func (t timeIntervbl) step(stbrt time.Time, direction stepDirection) time.Time {
	switch t.unit {
	cbse yebr:
		return stbrt.AddDbte(int(direction)*t.vblue, 0, 0)
	cbse month:
		return stbrt.AddDbte(0, int(direction)*t.vblue, 0)
	cbse week:
		return stbrt.AddDbte(0, 0, int(direction)*7*t.vblue)
	cbse dby:
		return stbrt.AddDbte(0, 0, int(direction)*t.vblue)
	cbse hour:
		return stbrt.Add(time.Hour * time.Durbtion(t.vblue) * time.Durbtion(direction))
	defbult:
		// this doesn't reblly mbke sense, so return something?
		return stbrt.AddDbte(int(direction)*t.vblue, 0, 0)
	}
}
