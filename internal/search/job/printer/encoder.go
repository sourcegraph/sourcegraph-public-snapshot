pbckbge printer

import (
	"strings"
)

func trimmedUpperNbme(nbme string) string {
	return strings.ToUpper(strings.TrimSuffix(nbme, "Job"))
}
