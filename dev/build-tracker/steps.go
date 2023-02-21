package main

import (
	"github.com/buildkite/go-buildkite/v3/buildkite"
	"strings"
)

type JobStatus string

const (
	FailedJob JobStatus = "Failed"
	PassedJob JobStatus = "Passed"
)

func (js JobStatus) ToStepState() StepState {
	return StepState(js)
}

type Job struct {
	buildkite.Job
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
	return strings.ToTitle(strp(j.State))
}

func (j *Job) status() JobStatus {
	if j.failed() {
		return FailedJob
	}
	return PassedJob
}

func (j *Job) hasTimedOut() bool {
	return j.state() == "timed_out"
}

func NewStep(name string) *Step {
	return &Step{
		Name: name,
		Jobs: make([]*Job, 0),
	}
}

func NewStepFromJob(j *Job) *Step {
	s := NewStep(j.name())
	s.Add(j)
	return s
}

func (s *Step) Add(j *Job) {
	s.Jobs = append(s.Jobs, j)
}

func (s *Step) FinalState() StepState {
	// If we have no jobs for some reason, then we regard it as the StepState as Passed ... cannot have a Failed StepState
	// if we have no jobs!
	if len(s.Jobs) == 0 {
		return Passed
	}
	if len(s.Jobs) == 1 {
		return s.LastJob().status().ToStepState()
	}
	// we only care about the last two states of because that determines the final state
	// n - 1  |   n    | Final
	// Passed | Passed | Passed
	// Passed | Failed | Failed
	// Failed | Failed | Failed
	// Failed | Passed | Fixed
	secondLastStatus := s.Jobs[len(s.Jobs)-2].status()
	lastStatus := s.Jobs[len(s.Jobs)-1].status()

	// Note that for all cases except the last case, the final state is whatever the last job state is.
	// The final state only differs when the before state is Failed and the last State is Passed, so
	finalState := lastStatus.ToStepState()
	if secondLastStatus == FailedJob && lastStatus == PassedJob {
		finalState = Fixed
	}

	return finalState
}

func (s *Step) LastJob() *Job {
	return s.Jobs[len(s.Jobs)-1]
}

func FindFailedSteps(steps map[string]*Step) []*Step {
	results := []*Step{}

	for _, step := range steps {
		if state := step.FinalState(); state == Failed {
			results = append(results, step)
		}
	}
	return results
}

func GroupIntoStepStates(steps map[string]*Step) map[StepState][]*Step {
	groups := make(map[StepState][]*Step)

	for _, step := range steps {
		state := step.FinalState()

		items, ok := groups[state]
		if !ok {
			items = make([]*Step, 0)
		}
		groups[state] = append(items, step)
	}

	return groups
}
