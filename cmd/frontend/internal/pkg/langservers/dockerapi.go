package langservers

// health stores information about the container's healthcheck results
//
// Partially copied from https://sourcegraph.com/github.com/docker/docker@c4e93da8a6fcd206e3fbfb07b821b5743f90f437/-/blob/api/types/types.go#L290
type health struct {
	Status        string // Status is one of Starting, Healthy or Unhealthy
	FailingStreak int    // FailingStreak is the number of consecutive failures
}

// containerState stores container's running state
// it's part of ContainerJSONBase and will return by "inspect" command
//
// Partially copied from https://sourcegraph.com/github.com/docker/docker@c4e93da8a6fcd206e3fbfb07b821b5743f90f437/-/blob/api/types/types.go#L296
type containerState struct {
	Status     string // String representation of the container state. Can be one of "created", "running", "paused", "restarting", "removing", "exited", or "dead"
	Running    bool
	Paused     bool
	Restarting bool
	OOMKilled  bool
	Dead       bool
	Pid        int
	ExitCode   int
	Error      string
	StartedAt  string
	FinishedAt string
	Health     *health `json:",omitempty"`
}

// containerInspection reflects the JSON result from `docker inspect <container>`
//
// Partially copied from https://sourcegraph.com/github.com/docker/docker@c4e93da8a6fcd206e3fbfb07b821b5743f90f437/-/blob/api/types/types.go#L327
type containerInspection struct {
	State *containerState
	Image string
}

// imageInspection reflects the JSON result from `docker inspect <image>`
//
// Partially copied from https://sourcegraph.com/github.com/docker/docker@c4e93da/-/blob/api/types/types.go#L29
type imageInspection struct {
	ID     string `json:"Id"`
	Config *containerConfig
}

// containerConfig contains the configuration data about a container.
// It should hold only portable information about the container.
// Here, "portable" means "independent from the host we are running on".
// Non-portable information *should* appear in HostConfig.
// All fields added to this struct must be marked `omitempty` to keep getting
// predictable hashes from the old `v1Compatibility` configuration.
//
// Partially copied from https://sourcegraph.com/github.com/docker/docker@c4e93da/-/blob/api/types/container/config.go#L43
type containerConfig struct {
	Image string // Name of the image as it was passed by the operator (e.g. could be symbolic)
}
