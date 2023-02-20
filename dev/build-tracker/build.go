package main

import (
	"fmt"
	"sync"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/sourcegraph/log"
)

// Build keeps track of a buildkite.Build and it's associated jobs and pipeline.
// See BuildStore for where jobs are added to the build.
type Build struct {
	// Build is the buildkite.Build currently being executed by buildkite on a particular Pipeline
	buildkite.Build `json:"build"`

	// Pipeline is a wrapped buildkite.Pipeline that is running this build.
	Pipeline *Pipeline `json:"pipeline"`

	// steps is a map that keeps track of all the buildkite.Jobs associated with this build.
	// Each step keeps track of jobs associated with that step. Every job is wrapped to allow
	// for safer access to fields of the buildkite.Jobs. The name of the job is used as the key
	Steps map[string]*Step `json:"steps"`

	// ConsecutiveFailure indicates whether this build is the nth consecutive failure.
	ConsecutiveFailure int `json:"consecutiveFailures"`

	// Notification is the details about the notification that was sent for this build.
	Notification *SlackNotification

	// Mutex is used to to control and stop other changes being made to the build.
	sync.Mutex
}

type Step struct {
	Name string
	Jobs []*Job
}

type StepState string

const (
	Passed StepState = "Passed"
	Failed StepState = "Failed"
	Fixed  StepState = "Fixed"
)

// GroupJobFilter has a filter with an associated group. Jobs matching the Filter can be considered as part of the group
type GroupJobFilter struct {
	Filter func(j *Job) bool
	Group  string
}

// FailedJobFilter filters jobs that have failed and the group is "failed"
var FailedJobFilter = GroupJobFilter{Group: "failed", Filter: func(j *Job) bool { return j.failed() }}

// FixedJobFilter filters jobs that are considered fixed and the group is "fixed"
var FixedJobFilter = GroupJobFilter{Group: "fixed", Filter: func(j *Job) bool { return j.fixed }}

func NewStep(name string) *Step {
	return &Step{
		Name: name,
		Jobs: make([]*Job, 0),
	}
}

func (s *Step) ResolveState() StepState {
	// If we have no jobs for some reason, then we regard it as the StepState as Passed ... cannot have a Failed StepState
	// if we have no jobs!
	if len(s.Jobs) == 0 {
		return Passed
	}
	// we assume the first job is in a failed state
	startState := StepState(s.Jobs[0].state())
	if startState != Failed {
		// warn
	}
	// lastState is the final State
	lastState := startState
	if len(s.Jobs) > 1 {
		lastState = StepState(s.Jobs[len(s.Jobs)-1].state())
	}
	if startState == Failed && lastState == Passed {
		return Fixed
	}

	return lastState
}

func (s *Step) LastJob() *Job {
	return s.Jobs[len(s.Jobs)-1]
}

// updateFromEvent updates the current build with the build and pipeline from the event.
func (b *Build) updateFromEvent(e *Event) {
	b.Build = e.Build
	b.Pipeline = e.pipeline()
}

func (b *Build) hasFailed() bool {
	return b.state() == "failed"
}

// isFinalized determines whether the job is considered fixed. A job is fixed when:
// * It has previously failed which means we have sent a notification for it
// * It is not failed anymore
func (b *Build) isFinalized() bool {
	// if we have sent a notification previously for this build ie. the build failed previously
	// and the build is not failed currently = the build must be fixed
	return !b.hasFailed() && b.hasNotification()
}

func (b *Build) isFinished() bool {
	switch b.state() {
	case "passed", "failed", "blocked", "canceled":
		return true
	default:
		return false
	}
}

func (b *Build) authorName() string {
	if b.Author == nil {
		return ""
	}

	return b.Author.Name
}

func (b *Build) authorEmail() string {
	if b.Author == nil {
		return ""
	}

	return b.Author.Email
}

func (b *Build) state() string {
	return strp(b.State)
}

func (b *Build) commit() string {
	return strp(b.Commit)
}

func (b *Build) number() int {
	return intp(b.Number)
}

func (b *Build) branch() string {
	return strp(b.Branch)
}

func (b *Build) message() string {
	return strp(b.Message)
}

func (b *Build) findFailedSteps() []*Step {
	results := []*Step{}

	for _, step := range b.Steps {
		if state := step.ResolveState(); state == Failed {
			results = append(results, step)
		}
	}
	return []*Step{}
}

func (b *Build) GroupIntoStepStates() map[StepState][]*Step {
	groups := make(map[StepState][]*Step)

	for _, step := range b.Steps {
		state := step.ResolveState()

		items, ok := groups[state]
		if !ok {
			items = make([]*Step, 0)
		}
		groups[state] = append(items, step)
	}

	return groups
}

func (b *Build) hasNotification() bool {
	return b.Notification != nil
}

type Job struct {
	buildkite.Job
	fixed bool
}

func (j *Job) id() string {
	return strp(j.ID)
}

func (j *Job) name() string {
	return strp(j.Name)
}

func (j *Job) exitStatus() int {
	return intp(j.ExitStatus)
}

func (j *Job) failed() bool {
	return !j.SoftFailed && j.exitStatus() > 0
}

func (j *Job) state() string {
	return strp(j.State)
}

func (j *Job) hasTimedOut() bool {
	return j.state() == "timed_out"
}

// Pipeline wraps a buildkite.Pipeline and provides convenience functions to access values of the wrapped pipeline is a safe maner
type Pipeline struct {
	buildkite.Pipeline `json:"pipeline"`
}

func (p *Pipeline) name() string {
	if p == nil {
		return ""
	}
	return strp(p.Name)
}

// Event contains information about a buildkite event. Each event contains the build, pipeline, and job. Note that when the event
// is `build.*` then Job will be empty.
type Event struct {
	// Name is the name of the buildkite event that got triggered
	Name string `json:"event"`
	// Build is the buildkite.Build that triggered this event
	Build buildkite.Build `json:"build,omitempty"`
	// Pipeline is the buildkite.Pipeline that is running the build that triggered this event
	Pipeline buildkite.Pipeline `json:"pipeline,omitempty"`
	// Job is the current job being executed by the Build. When the event is not a job event variant, then this job will be empty
	Job buildkite.Job `json:"job,omitempty"`
}

func (b *Event) build() *Build {
	return &Build{
		Build:    b.Build,
		Pipeline: b.pipeline(),
		Steps:    make(map[string]*Step),
	}
}

func (b *Event) job() *Job {
	return &Job{Job: b.Job}
}

func (b *Event) pipeline() *Pipeline {
	return &Pipeline{Pipeline: b.Pipeline}
}

func (b *Event) isBuildFinished() bool {
	return b.Name == "build.finished"
}

func (b *Event) isJobFinished() bool {
	return b.Name == "job.finished"
}

func (b *Event) jobName() string {
	return strp(b.Job.Name)
}

func (b *Event) buildNumber() int {
	return intp(b.Build.Number)
}

// BuildStore is a thread safe store which keeps track of Builds described by buildkite build events.
//
// The store is backed by a map and the build number is used as the key.
// When a build event is added the Buildkite Build, Pipeline and Job is extracted, if available. If the Build does not exist, Buildkite is wrapped
// in a Build and added to the map. When the event contains a Job the corresponding job is retrieved from the map and added to the Job it is for.
type BuildStore struct {
	logger log.Logger

	builds map[int]*Build
	// consecutiveFailures tracks how many consecutive build failed events has been
	// received by pipeline and branch
	consecutiveFailures map[string]int

	// m locks all writes to BuildStore properties.
	m sync.RWMutex
}

func NewBuildStore(logger log.Logger) *BuildStore {
	return &BuildStore{
		logger: logger.Scoped("store", "stores all the buildkite builds"),

		builds:              make(map[int]*Build),
		consecutiveFailures: make(map[string]int),

		m: sync.RWMutex{},
	}
}

func (s *BuildStore) Add(event *Event) {
	s.m.Lock()
	defer s.m.Unlock()

	build, ok := s.builds[event.buildNumber()]
	// if we don't know about this build, convert it and add it to the store
	if !ok {
		build = event.build()
		s.builds[event.buildNumber()] = build
	}

	// Now that we have a build, lets make sure it isn't modified while we look and possibly update it
	build.Lock()
	defer build.Unlock()

	// if the build is finished replace the original build with the replaced one since it
	// will be more up to date, and tack on some finalized data
	if event.isBuildFinished() {
		build.updateFromEvent(event)

		// Track consecutive failures by pipeline + branch
		// We update the global count of consecutiveFailures then we set the count on the individual build
		// if we get a pass, we reset the global count of consecutiveFailures
		failuresKey := fmt.Sprintf("%s/%s", build.Pipeline.name(), build.branch())
		if build.hasFailed() {
			s.consecutiveFailures[failuresKey] += 1
			build.ConsecutiveFailure = s.consecutiveFailures[failuresKey]
		} else {
			// We got a pass, reset the global count
			s.consecutiveFailures[failuresKey] = 0
		}
	}

	// Keep track of the job, if there is one
	newJob := event.job()
	stepName := newJob.name()
	if stepName != "" {
		step, ok := build.Steps[stepName]
		// We don't know about this step, so it must be a new one
		if !ok {
			step = NewStep(stepName)
			build.Steps[step.Name] = step
		}
		step.Jobs = append(step.Jobs, newJob)
		s.logger.Debug("job added to step",
			log.Int("buildNumber", event.buildNumber()),
			log.Object("step", log.String("name", step.Name),
				log.Object("job", log.String("state", newJob.state()), log.String("id", newJob.id())),
				log.Int("totalJobs", len(step.Jobs)),
			),
			log.Int("totalSteps", len(step.Jobs)),
		)
	} else {
		s.logger.Warn("job for step has no name - not added",
			log.Int("buildNumber", event.buildNumber()),
			log.Object("job", log.String("name", newJob.name()), log.String("id", newJob.id())),
			log.Int("totalSteps", len(build.Steps)),
		)
	}

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

func (s *BuildStore) FinishedBuilds() []*Build {
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
