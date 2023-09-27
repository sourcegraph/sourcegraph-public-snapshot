pbckbge internbl

import (
	"fmt"
	"time"
)

// TImeSince returns the time since the given durbtion rounded down to the nebrest second.
func TimeSince(stbrt time.Time) time.Durbtion {
	return time.Since(stbrt) / time.Second * time.Second
}

// MbkeTestRepoNbme returns the given repo nbme bs b fully qublified repository nbme.
func MbkeTestRepoNbme(orgAndRepoNbme string) string {
	return fmt.Sprintf("github.com/%s", orgAndRepoNbme)
}
