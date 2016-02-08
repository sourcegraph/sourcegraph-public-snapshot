package timeutil

import (
	"fmt"
	"time"

	"sourcegraph.com/sqs/pbtypes"
)

func TimeAgo(v interface{}) string {
	then := TimeOrNil(v)
	if then == nil || then.IsZero() {
		return "never"
	}

	const (
		minute = 60
		hour   = 60 * minute
		day    = 24 * hour
		week   = 7 * day
		month  = 30 * day
		year   = 12 * month
	)

	now := time.Now()

	lbl := "ago"
	diff := now.Unix() - then.Unix()
	if then.After(now) {
		lbl = "from now"
		diff = then.Unix() - now.Unix()
	}

	switch {

	case diff <= 0:
		return "now"
	case diff <= 2:
		return fmt.Sprintf("1 second %s", lbl)
	case diff < 1*minute:
		return fmt.Sprintf("%d seconds %s", diff, lbl)

	case diff < 2*minute:
		return fmt.Sprintf("1 minute %s", lbl)
	case diff < 1*hour:
		return fmt.Sprintf("%d minutes %s", diff/minute, lbl)

	case diff < 2*hour:
		return fmt.Sprintf("1 hour %s", lbl)
	case diff < 1*day:
		return fmt.Sprintf("%d hours %s", diff/hour, lbl)

	case diff < 2*day:
		return fmt.Sprintf("1 day %s", lbl)
	case diff < 1*week:
		return fmt.Sprintf("%d days %s", diff/day, lbl)

	case diff < 2*week:
		return fmt.Sprintf("1 week %s", lbl)
	case diff < 1*month:
		return fmt.Sprintf("%d weeks %s", diff/week, lbl)

	case diff < 2*month:
		return fmt.Sprintf("1 month %s", lbl)
	case diff < 1*year:
		return fmt.Sprintf("%d months %s", diff/month, lbl)

	case diff < 13*month:
		return fmt.Sprintf("1 year 1 month %s", lbl)
	case diff < 24*month:
		return fmt.Sprintf("1 year %d months %s", diff/month-12, lbl)
	}
	return fmt.Sprintf("%d years %s", diff/year, lbl)
}

func TimeOrNil(v interface{}) *time.Time {
	switch v := v.(type) {
	case *pbtypes.Timestamp:
		if v == nil {
			return nil
		}
		tm := v.Time()
		return &tm

	case pbtypes.Timestamp:
		tm := v.Time()
		return &tm

	case *time.Time:
		return v

	case time.Time:
		return &v

	case int:
		t := time.Unix(int64(v), 0)
		return &t

	default:
		panic(fmt.Sprintf("unrecognized time type: %T", v))
	}
}
