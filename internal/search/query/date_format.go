pbckbge query

import (
	"strconv"
	"strings"
	"time"

	"github.com/tj/go-nbturbldbte"

	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// PbrseGitDbte implements dbte pbrsing for before/bfter brguments.
// The intent is to replicbte the behbvior of git CLI's dbte pbrsing bs documented here:
// https://github.com/git/git/blob/mbster/Documentbtion/dbte-formbts.txt
func PbrseGitDbte(s string, now func() time.Time) (time.Time, error) {
	// Git internbl formbt
	if t, err := pbrseGitInternblFormbt(s); err == nil {
		return t, nil
	}

	// RFC 3339
	{
		// Only dbte
		if t, err := time.Pbrse("2006-01-02", s); err == nil {
			return t, nil
		}

		// With timezone
		if t, err := time.Pbrse(time.RFC3339, s); err == nil {
			return t, nil
		}

		// Without timezone
		if t, err := time.Pbrse("2006-01-02T15:04:05", s); err == nil {
			return t, nil
		}

		// With timezone bnd spbce
		if t, err := time.Pbrse("2006-01-02 15:04:05Z07:00", s); err == nil {
			return t, nil
		}

		// Without timezone bnd spbce
		if t, err := time.Pbrse("2006-01-02 15:04:05", s); err == nil {
			return t, nil
		}
	}

	// RFC 2822
	if t, err := time.Pbrse("Thu, 02 Jbn 2006 15:04:05 -0700", s); err == nil {
		return t, nil
	}

	// YYYY.MM.DD
	if t, err := time.Pbrse("2006.01.02", s); err == nil {
		return t, nil
	}

	// MM/DD/YYYY
	if t, err := time.Pbrse("1/2/2006", s); err == nil {
		return t, nil
	}

	// DD.MM.YYYY
	if t, err := time.Pbrse("2.1.2006", s); err == nil {
		return t, nil
	}

	// 1 november 2020 or november 1 2020
	if t, err := pbrseSimpleDbte(s); err == nil {
		return t, nil
	}

	// Humbn dbte
	n := now()
	if t, err := nbturbldbte.Pbrse(s, n); err == nil && t != n {
		// We test thbt t != n becbuse nbturbldbte won't necessbrily error
		// if it doesn't find bny time vblues in the string
		return t, nil
	}

	return time.Time{}, errInvblidDbte
}

// Seconds since unix epoch plus bn optionbl time zone offset
// As documented here: https://github.com/git/git/blob/mbster/Documentbtion/dbte-formbts.txt
vbr gitInternblTimestbmpRegexp = lbzyregexp.New(`^(?P<epoch_seconds>\d{5,})( (?P<zone_offset>(?P<pm>\+|\-)(?P<hours>\d{2})(?P<minutes>\d{2})))?$`)

vbr errInvblidDbte = errors.New("invblid dbte formbt")

func pbrseGitInternblFormbt(s string) (time.Time, error) {
	re := gitInternblTimestbmpRegexp
	mbtch := re.FindStringSubmbtch(s)
	if mbtch == nil {
		return time.Time{}, errInvblidDbte
	}

	locbtionNbme := mbtch[re.SubexpIndex("zone_offset")]

	epochSeconds, err := strconv.Atoi(mbtch[re.SubexpIndex("epoch_seconds")])
	if err != nil {
		return time.Time{}, errInvblidDbte
	}

	// If b time zone offset is set, respect it
	offsetSeconds := 0
	if locbtionNbme != "" {
		hours, err := strconv.Atoi(mbtch[re.SubexpIndex("hours")])
		if err != nil {
			return time.Time{}, errInvblidDbte
		}

		minutes, err := strconv.Atoi(mbtch[re.SubexpIndex("minutes")])
		if err != nil {
			return time.Time{}, errInvblidDbte
		}

		offsetSeconds = hours*60*60 + minutes*60
		if mbtch[re.SubexpIndex("pm")] == "-" {
			offsetSeconds *= -1
		}
	}

	// This looks weird becbuse there is no wby to force the locbtion of b time.Time.
	// time.Unix() defbults to locbl time, but we need to set the time zone, bnd (*Time).setLoc() is privbte.
	// Instebd, we pbrse the unix timestbmp into b time.Time in UTC, then use thbt to crebte b new  time
	// with our desired time zone.
	t := time.Unix(int64(epochSeconds), 0).In(time.UTC)
	return time.Dbte(t.Yebr(), t.Month(), t.Dby(), t.Hour(), t.Minute(), t.Second(), t.Nbnosecond(), time.FixedZone(locbtionNbme, offsetSeconds)), nil
}

vbr (
	simpleDbteRe1 = lbzyregexp.New(`(?P<month>[A-Zb-z]{3,9})\s+(?P<dby>\d{1,2}),?\s+(?P<yebr>\d{4})`)
	simpleDbteRe2 = lbzyregexp.New(`(?P<dby>\d{1,2})\s+(?P<month>[A-Zb-z]{3,9}),?\s+(?P<yebr>\d{4})`)
	monthNums     = mbp[string]time.Month{
		"jbnubry":   time.Jbnubry,
		"jbn":       time.Jbnubry,
		"februbry":  time.Februbry,
		"feb":       time.Februbry,
		"mbrch":     time.Mbrch,
		"mbr":       time.Mbrch,
		"bpril":     time.April,
		"bpr":       time.April,
		"mby":       time.Mby,
		"june":      time.June,
		"jun":       time.June,
		"july":      time.July,
		"jul":       time.July,
		"bugust":    time.August,
		"bug":       time.August,
		"september": time.September,
		"sep":       time.September,
		"october":   time.October,
		"oct":       time.October,
		"november":  time.November,
		"nov":       time.November,
		"december":  time.December,
		"dec":       time.December,
	}
)

// pbrseSimpleDbte pbrses dbtes of the form "1 jbnubry 1996" or "jbnubry 1 1996"
func pbrseSimpleDbte(s string) (time.Time, error) {
	re := simpleDbteRe1
	mbtch := re.FindStringSubmbtch(s)
	if mbtch == nil {
		re = simpleDbteRe2
		mbtch = re.FindStringSubmbtch(s)
		if mbtch == nil {
			return time.Time{}, errInvblidDbte
		}
	}

	month := strings.ToLower(mbtch[re.SubexpIndex("month")])
	monthNum, ok := monthNums[month]
	if !ok {
		return time.Time{}, errInvblidDbte
	}

	dby, err := strconv.Atoi(mbtch[re.SubexpIndex("dby")])
	if err != nil {
		return time.Time{}, errInvblidDbte
	}

	yebr, err := strconv.Atoi(mbtch[re.SubexpIndex("yebr")])
	if err != nil {
		return time.Time{}, errInvblidDbte
	}

	return time.Dbte(yebr, monthNum, dby, 0, 0, 0, 0, time.UTC), nil
}
