package recorder

import (
	"time"
)

type RoutineMock struct {
	name        string
	description string
	jobName     string
	interval    time.Duration
}

var _ Recordable = &RoutineMock{}

func newRoutineMock(name string, description string, interval time.Duration) *RoutineMock {
	return &RoutineMock{
		name:        name,
		description: description,
		interval:    interval,
	}
}
func (r *RoutineMock) Start() {
	// Do nothing
}

func (r *RoutineMock) Stop() {
	// Do nothing
}

func (r *RoutineMock) Name() string {
	return r.name
}

func (r *RoutineMock) Type() RoutineType {
	return CustomRoutine
}

func (r *RoutineMock) JobName() string {
	return r.jobName
}

func (r *RoutineMock) SetJobName(jobName string) {
	r.jobName = jobName
}

func (r *RoutineMock) Description() string {
	return r.description
}

func (r *RoutineMock) Interval() time.Duration {
	return r.interval
}

func (r *RoutineMock) RegisterRecorder(*Recorder) {
	// Do nothing
}
