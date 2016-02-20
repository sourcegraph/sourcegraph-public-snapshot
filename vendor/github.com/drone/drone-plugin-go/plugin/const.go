package plugin

const (
	StatePending = "pending"
	StateRunning = "running"
	StateSuccess = "success"
	StateFailure = "failure"
	StateKilled  = "killed"
	StateError   = "error"
)

const (
	EventPush   = "push"
	EventPull   = "pull_request"
	EventTag    = "tag"
	EventDeploy = "deploy"
)
