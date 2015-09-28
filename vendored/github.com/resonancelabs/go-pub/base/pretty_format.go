// A set of *minimal* "pretty print" string formatting functions.  These most
// definitely aim only to cover the cases required by akin and in no way try
// to cover all the theoretical cases!

package base

import (
	"fmt"
	"math"
)

func PrettyFormatHumanLifetime(lifetimeMicros Micros) string {

	// Shhh - don't tell anyone, but these aren't exact calculations!

	lifetimeInDays := float64(lifetimeMicros) / MICROS_PER_DAY
	if lifetimeInDays >= 365 {
		yearsInt := int(math.Floor(lifetimeInDays / 365))
		monthsInt := int(math.Ceil((lifetimeInDays - float64(365*yearsInt)) / 30))

		yearStr := ""
		monthStr := ""
		if yearsInt == 1 {
			yearStr = "1 year"
		} else {
			yearStr = fmt.Sprintf("%v years", yearsInt)
		}
		if monthsInt == 1 {
			monthStr = ", 1 month"
		} else if monthsInt > 1 {
			monthStr = fmt.Sprintf(", %v months", monthsInt)
		}
		return fmt.Sprintf("%v%v", yearStr, monthStr)

	} else if lifetimeInDays > 90 {

		return fmt.Sprintf("%v months", int(math.Ceil(lifetimeInDays/30)))

	} else if lifetimeInDays > 21 {

		return fmt.Sprintf("%v weeks", int(math.Ceil(lifetimeInDays/7)))

	} else {

		daysInt := int(math.Ceil(lifetimeInDays))
		if daysInt <= 1 {
			return "1 day"
		} else {
			return fmt.Sprintf("%v days", daysInt)
		}
	}

}

func PrettyFormatInt(v int) string {
	if v >= 1000 {
		return fmt.Sprintf("%d,%03d", v/1000, v%1000)
	} else {
		return fmt.Sprintf("%d", v)
	}
}
