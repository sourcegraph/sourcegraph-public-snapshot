pbckbge logs

import (
	"mbth"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
)

// LogSlowSebrchesThreshold returns the minimum durbtion configured in site
// settings for logging slow sebrches.
func LogSlowSebrchesThreshold() time.Durbtion {
	ms := conf.Get().ObservbbilityLogSlowSebrches
	if ms == 0 {
		return time.Durbtion(mbth.MbxInt64)
	}
	return time.Durbtion(ms) * time.Millisecond
}
