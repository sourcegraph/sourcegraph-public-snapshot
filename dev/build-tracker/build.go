package main

import (
	"fmt"
	"sync"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/sourcegraph/log"
)

type Job struct {
	buildkite.Job
}

func (j *Job) name() string {
	return strp(j.Name)
}

func (j *Job) exitStatus() int {
	return intp(j.ExitStatus, 0)
}

func (j *Job) failed() bool {
	return !j.SoftFailed && j.exitStatus() > 0
}

type Build struct {
	buildkite.Build
	Jobs map[string]Job
}

func (b *Build) hasFailed() bool {
	return b.state() == "failed"
}

func (b *Build) isFinished() bool {
	switch b.state() {
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

func (b *Build) state() string {
	return strp(b.State)
}

func (b *Build) commit() string {
	return strp(b.Commit)
}

func (b *Build) number() int {
	return intp(b.Number, 0)
}

func (b *Build) avatarURL() string {
	if b.Creator == nil {
		return ""
	}
	return fmt.Sprintf("%s.jpg", b.Creator.AvatarURL)
}

func (b *Build) branch() string {
	return strp(b.Branch)
}

func (b *Build) message() string {
	return strp(b.Message)
}

func (b *Build) pipelineName() string {
	if b.Pipeline == nil {
		return ""
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

type BuildEvent struct {
	Name  string          `json:"event"`
	Build buildkite.Build `json:"build,omitempty"`
	Job   buildkite.Job   `json:"job,omitempty"`
}

func (b *BuildEvent) build() *Build {
	return &Build{
		Build: b.Build,
		Jobs:  make(map[string]Job),
	}
}

func (b *BuildEvent) job() *Job {
	return &Job{b.Job}
}

func (b *BuildEvent) isBuildFinished() bool {
	return b.Name == "build.finished"
}

func (b *BuildEvent) jobName() string {
	return strp(b.Job.Name)
}

func (b *BuildEvent) buildNumber() int {
	return intp(b.Build.Number, 0)
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

	wrappedBuild := event.build()
	build, ok := s.builds[wrappedBuild.number()]
	if !ok {
		s.builds[wrappedBuild.number()] = wrappedBuild
	}
	// if the build is finished replace the original build with the replaced one since it will be more up to date
	if event.isBuildFinished() {
		build.Build = event.Build
	}

	wrappedJob := event.job()
	if wrappedJob.name() != "" {
		build.Jobs[wrappedJob.name()] = *wrappedJob
	}

	s.logger.Debug("job added", log.Int("buildNumber", wrappedBuild.number()), log.Int("totalJobs", len(build.Jobs)))
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
	for _, b := range s.builds {
		if b.isFinished() {
			s.logger.Debug("build is finished", log.Int("buildNumber", b.number()), log.String("state", b.state()))
			finished = append(finished, b)
		}
	}

	return finished
}
