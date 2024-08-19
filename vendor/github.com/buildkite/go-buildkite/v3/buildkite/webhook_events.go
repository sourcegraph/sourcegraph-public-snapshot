package buildkite

// agentEvent is a wrapper for an agent event notification
//
// Buildkite API docs: https://buildkite.com/docs/apis/webhooks/agent-events
type AgentEvent struct {
	Event  *string `json:"event"`
	Agent  *Agent  `json:"agent"`
	Sender *User   `json:"sender"`
}

// AgentConnectedEvent is triggered when an agent has connected to the API
//
// Buildkite API docs: https://buildkite.com/docs/apis/webhooks/agent-events
type AgentConnectedEvent struct {
	AgentEvent
}

// AgentDisconnectedEvent is triggered when an agent has disconnected.
//
// Buildkite API docs: https://buildkite.com/docs/apis/webhooks/agent-events
type AgentDisconnectedEvent struct {
	AgentEvent
}

// AgentLostEvent is triggered when an agent has been marked as lost.
//
// Buildkite API docs: https://buildkite.com/docs/apis/webhooks/agent-events
type AgentLostEvent struct {
	AgentEvent
}

// AgentStoppedEvent is triggered when an agent has stopped.
//
// Buildkite API docs: https://buildkite.com/docs/apis/webhooks/agent-events
type AgentStoppedEvent struct {
	AgentEvent
}

// AgentStoppingEvent is triggered when an agent is stopping.
//
// Buildkite API docs: https://buildkite.com/docs/apis/webhooks/agent-events
type AgentStoppingEvent struct {
	AgentEvent
}

// buildEvent is a wrapper for a build event notification
//
// Buildkite API docs: https://buildkite.com/docs/apis/webhooks/build-events
type BuildEvent struct {
	Event    *string   `json:"event"`
	Build    *Build    `json:"build"`
	Pipeline *Pipeline `json:"pipeline"`
	Sender   *User     `json:"sender"`
}

// BuildFailingEvent is triggered when a build enters a failing state
//
// Buildkite API docs: https://buildkite.com/docs/apis/webhooks/build-events
type BuildFailingEvent struct {
	BuildEvent
}

// BuildFinishedEvent is triggered when a build finishes
//
// Buildkite API docs: https://buildkite.com/docs/apis/webhooks/build-events
type BuildFinishedEvent struct {
	BuildEvent
}

// BuildRunningEvent is triggered when a build starts running
//
// Buildkite API docs: https://buildkite.com/docs/apis/webhooks/build-events
type BuildRunningEvent struct {
	BuildEvent
}

// BuildScheduledEvent is triggered when a build is scheduled
//
// Buildkite API docs: https://buildkite.com/docs/apis/webhooks/build-events
type BuildScheduledEvent struct {
	BuildEvent
}

// jobEvent is a wrapper for a job event notification
//
// Buildkite API docs: https://buildkite.com/docs/apis/webhooks/job-events
type JobEvent struct {
	Event    *string   `json:"event"`
	Build    *Build    `json:"build"`
	Job      *Job      `json:"job"`
	Pipeline *Pipeline `json:"pipeline"`
	Sender   *User     `json:"sender"`
}

// JobActivatedEvent is triggered when a job is activated
//
// Buildkite API docs: https://buildkite.com/docs/apis/webhooks/job-events
type JobActivatedEvent struct {
	JobEvent
}

// JobFinishedEvent is triggered when a job is finished
//
// Buildkite API docs: https://buildkite.com/docs/apis/webhooks/job-events
type JobFinishedEvent struct {
	JobEvent
}

// JobScheduledEvent is triggered when a job is scheduled
//
// Buildkite API docs: https://buildkite.com/docs/apis/webhooks/job-events
type JobScheduledEvent struct {
	JobEvent
}

// JobStartedEvent is triggered when a job is started
//
// Buildkite API docs: https://buildkite.com/docs/apis/webhooks/job-events
type JobStartedEvent struct {
	JobEvent
}

// PingEvent is triggered when a webhook notification setting is changed
//
// Buildkite API docs: https://buildkite.com/docs/apis/webhooks/ping-events
type PingEvent struct {
	Event        *string       `json:"event"`
	Organization *Organization `json:"organization"`
	Sender       *User         `json:"sender"`
}
