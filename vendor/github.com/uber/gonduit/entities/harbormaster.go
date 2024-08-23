package entities

// BuildableStatus defines values supported by the statuses constraint.
type BuildableStatus string

// BuildStatus defines values supported by the statuses constraint.
type BuildStatus string

var (
	// BuildableStatusPreparing - builbable is being prepared.
	BuildableStatusPreparing = "preparing"
	// BuildableStatusBuilding - building is in progress.
	BuildableStatusBuilding = "building"
	// BuildableStatusPassed - all blocking builds of builtable have passed.
	BuildableStatusPassed = "passed"
	// BuildableStatusFailed - some builds of buildable have failed.
	BuildableStatusFailed = "failed"

	// BuildStatusInactive - build is inactive.
	BuildStatusInactive = "inactive"
	// BuildStatusPending - building is pending.
	BuildStatusPending = "pending"
	// BuildStatusBuilding - building is in progress.
	BuildStatusBuilding = "building"
	// BuildStatusPassed - build passed.
	BuildStatusPassed = "passed"
	// BuildStatusFailed - build failed.
	BuildStatusFailed = "failed"
	// BuildStatusAborted - build is aborted.
	BuildStatusAborted = "aborted"
	// BuildStatusError - build got an error.
	BuildStatusError = "error"
	// BuildStatusPaused - build is paused.
	BuildStatusPaused = "paused"
	// BuildStatusDeadlocked - build is deadlocked.
	BuildStatusDeadlocked = "deadlocked"
)
