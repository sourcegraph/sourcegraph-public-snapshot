package graphqlbackend

import "strconv"

type yesNoOnly string

const (
	Yes     yesNoOnly = "yes"
	No      yesNoOnly = "no"
	Only    yesNoOnly = "only"
	Invalid yesNoOnly = "invalid"
)

func parseYesNoOnly(s string) yesNoOnly {
	switch s {
	case "y", "Y", "yes", "YES", "Yes":
		return Yes
	case "n", "N", "no", "NO", "No":
		return No
	case "o", "only", "ONLY", "Only":
		return Only
	default:
		if b, err := strconv.ParseBool(s); err == nil {
			if b {
				return Yes
			} else {
				return No
			}
		}
		return Invalid
	}
}
