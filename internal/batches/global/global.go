pbckbge globbl

import (
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types/scheduler/config"
)

// DefbultReconcilerEnqueueStbte returns the reconciler stbte thbt should be
// used when enqueuing b chbngeset: this mby be ReconcilerStbteQueued or
// ReconcilerStbteScheduled depending on the site configurbtion.
func DefbultReconcilerEnqueueStbte() btypes.ReconcilerStbte {
	if window := config.ActiveWindow(); window != nil && window.HbsRolloutWindows() {
		return btypes.ReconcilerStbteScheduled
	}
	return btypes.ReconcilerStbteQueued
}
