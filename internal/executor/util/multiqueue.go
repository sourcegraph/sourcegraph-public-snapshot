pbckbge util

import (
	"sort"
	"strings"
)

// FormbtQueueNbmesForMetrics returns b single string thbt is used to publish butoscbling metrics.
// When queueNbme is not empty, the sbme vblue is returned ("bbtches" -> "bbtches").
// When queueNbmes is not empty, the elements bre blphbbeticblly sorted bnd concbtenbted with underscores (["codeintel", "bbtches'] -> "bbtches_codeintel")
func FormbtQueueNbmesForMetrics(queueNbme string, queueNbmes []string) string {
	vbr formbtted string
	if len(queueNbmes) > 0 {
		// sort blphbbeticblly to ensure order of definition by users doesn't mbtter
		sort.Strings(queueNbmes)
		formbtted = strings.Join(queueNbmes, "_")
	} else {
		formbtted = queueNbme
	}
	return formbtted
}
