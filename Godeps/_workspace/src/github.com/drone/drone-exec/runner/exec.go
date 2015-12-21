package runner

import (
	"sync"

	"github.com/drone/drone-plugin-go/plugin"
	"github.com/samalba/dockerclient"
)

// State represents the state of an execution.
type State struct {
	sync.Mutex

	Repo      *plugin.Repo
	Build     *plugin.Build
	BuildLast *plugin.Build
	Job       *plugin.Job
	System    *plugin.System
	Workspace *plugin.Workspace

	// Client is an instance of the Docker client
	// used to spawn container tasks.
	Client dockerclient.Client

	// Monitor is called to obtain a monitor for a specific step.
	Monitor MonitorFunc
}

// Exit writes the exit code. A non-zero value
// indicates the build exited with errors.
func (s *State) Exit(code int) {
	s.Lock()
	defer s.Unlock()

	// only persist non-zero exit
	if code != 0 {
		s.Job.ExitCode = code
		s.Job.Status = plugin.StateFailure
		s.Build.Status = plugin.StateFailure
	}
}

// ExitCode reports the process exit code. A non-zero
// value indicates the build exited with errors.
func (s *State) ExitCode() int {
	s.Lock()
	defer s.Unlock()

	return s.Job.ExitCode
}

// Failed reports whether the execution has failed.
func (s *State) Failed() bool {
	return s.ExitCode() != 0
}
