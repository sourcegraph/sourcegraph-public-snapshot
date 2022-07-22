package main

import (
	"fmt"
	"sync"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/sourcegraph/log"
)

type Build struct {
	buildkite.Build
	Jobs map[string]buildkite.Job
}

func (b *Build) HasFailed(logger log.Logger) bool {
	state := ""
	if b.State != nil {
		state = *b.State
	}

	// no need to check the jobs if the overall build hasn't failed
	if state != "failed" {
		logger.Debug("build state not failed", log.String("State", state))
		return false
	}

	for n, j := range b.Jobs {
		failed := j.ExitStatus != nil && !j.SoftFailed && *j.ExitStatus > 0
		logger.Debug("checking job", log.String("Name", n), log.Bool("failed", failed))
		if failed {
			return true
		}
	}
	return false
}

func (b *Build) IsFinished() bool {
	state := ""
	if b.State != nil {
		state = *b.State
	}

	switch state {
	case "passed":
		fallthrough
	case "failed":
		fallthrough
	case "blocked":
		fallthrough
	case "canceled":
		return true
	}

	return false

}

func (b *Build) AvatarURL() string {
	if b.Creator == nil {
		return ""
	}
	return fmt.Sprintf("%s.jpg", b.Creator.AvatarURL)
}

func (b *Build) PipelineName() string {
	if b.Pipeline == nil {
		return "N/A"
	}

	var slug, name string
	if b.Pipeline.Name != nil {
		name = *b.Pipeline.Name
	}
	if b.Pipeline.Slug != nil {
		slug = *b.Pipeline.Slug
	}

	if name == "" {
		return name
	}
	return slug

}

func NewBuildFrom(event *BuildEvent) *Build {
	return &Build{
		Build: event.Build,
		Jobs:  make(map[string]buildkite.Job),
	}
}

type BuildEvent struct {
	Event string          `json:"event"`
	Build buildkite.Build `json:"build,omitempty"`
	Job   buildkite.Job   `json:"job,omitempty"`
}

func (b *BuildEvent) IsBuildFinished() bool {
	return b.Event == "build.finished"
}

func (b *BuildEvent) BuildNumber() int {
	if b.Build.Number == nil {
		return -1
	}
	return *b.Build.Number
}

func (b *BuildEvent) JobName() string {
	if b.Job.Name == nil {
		return ""
	}
	return *b.Job.Name
}

type BuildStore struct {
	logger log.Logger
	builds map[int]*Build
	m      sync.RWMutex
}

func NewBuildStore(logger log.Logger) *BuildStore {
	return &BuildStore{
		logger: logger.Scoped("store", "stores all the builds"),
		builds: make(map[int]*Build),
		m:      sync.RWMutex{},
	}
}

func (s *BuildStore) Add(event *BuildEvent) {
	s.m.Lock()
	defer s.m.Unlock()
	build, ok := s.builds[event.BuildNumber()]
	if !ok {
		build = NewBuildFrom(event)
		s.builds[event.BuildNumber()] = build
	}
	// if the build is finished replace the original build with the replaced one since it will be more up to date
	if event.IsBuildFinished() {
		build.Build = event.Build
	}

	if event.JobName() != "" {
		build.Jobs[event.JobName()] = event.Job
	}

	s.logger.Debug("job added", log.Int("buildNumber", event.BuildNumber()), log.Int("totalJobs", len(build.Jobs)))
}

func (s *BuildStore) GetByBuildNumber(num int) *Build {
	s.m.RLock()
	defer s.m.RUnlock()

	return s.builds[num]
}

func (s *BuildStore) DelByBuildNumber(buildNumbers ...int) {
	s.m.Lock()
	defer s.m.Unlock()

	for _, num := range buildNumbers {
		delete(s.builds, num)
	}
	s.logger.Info("deleted builds", log.Int("totalBuilds", len(buildNumbers)))
}

func (s *BuildStore) AllFinishedBuilds() []*Build {
	s.m.RLock()
	defer s.m.RUnlock()

	finished := make([]*Build, 0)
	for _, v := range s.builds {
		if v.IsFinished() {
			s.logger.Debug("build is finished", log.Int("buildNumber", *v.Number), log.String("State", *v.State))
			finished = append(finished, v)
		}
	}

	return finished
}
