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
		Minute = 60
		Hour   = 60 * Minute
		Day    = 24 * Hour
		Week   = 7 * Day
		Month  = 30 * Day
		Year   = 12 * Month
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
	case diff < 1*Minute:
		return fmt.Sprintf("%d seconds %s", diff, lbl)

	case diff < 2*Minute:
		return fmt.Sprintf("1 minute %s", lbl)
	case diff < 1*Hour:
		return fmt.Sprintf("%d minutes %s", diff/Minute, lbl)

	case diff < 2*Hour:
		return fmt.Sprintf("1 hour %s", lbl)
	case diff < 1*Day:
		return fmt.Sprintf("%d hours %s", diff/Hour, lbl)

	case diff < 2*Day:
		return fmt.Sprintf("1 day %s", lbl)
	case diff < 1*Week:
		return fmt.Sprintf("%d days %s", diff/Day, lbl)

	case diff < 2*Week:
		return fmt.Sprintf("1 week %s", lbl)
	case diff < 1*Month:
		return fmt.Sprintf("%d weeks %s", diff/Week, lbl)

	case diff < 2*Month:
		return fmt.Sprintf("1 month %s", lbl)
	case diff < 1*Year:
		return fmt.Sprintf("%d months %s", diff/Month, lbl)

	case diff < 13*Month:
		return fmt.Sprintf("1 year 1 month %s", lbl)
	case diff < 24*Month:
		return fmt.Sprintf("1 year %d months %s", diff/Month-12, lbl)
	}
	return fmt.Sprintf("%d years %s", diff/Year, lbl)
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
