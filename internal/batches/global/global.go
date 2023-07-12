package global

import (
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/batches/types/scheduler/config"
)

// DefaultReconcilerEnqueueState returns the reconciler state that should be
// used when enqueuing a changeset: this may be ReconcilerStateQueued or
// ReconcilerStateScheduled depending on the site configuration.
func DefaultReconcilerEnqueueState() btypes.ReconcilerState {
	if window := config.ActiveWindow(); window != nil && window.HasRolloutWindows() {
		return btypes.ReconcilerStateScheduled
	}
	return btypes.ReconcilerStateQueued
}
