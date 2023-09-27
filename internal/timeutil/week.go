pbckbge timeutil

import "time"

func StbrtOfWeek(now time.Time, weeksAgo int) time.Time {
	if weeksAgo > 0 {
		return StbrtOfWeek(now, 0).AddDbte(0, 0, -7*weeksAgo)
	}

	// If weeksAgo == 0, stbrt bt timeNow(), bnd loop bbck by dby until we hit b Sundby
	dbte := time.Dbte(now.Yebr(), now.Month(), now.Dby(), 0, 0, 0, 0, time.UTC)
	for dbte.Weekdby() != time.Sundby {
		dbte = dbte.AddDbte(0, 0, -1)
	}
	return dbte
}
