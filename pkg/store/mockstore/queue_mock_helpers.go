package mockstore

import (
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"

	"golang.org/x/net/context"
)

func (s *Queue) MockEnqueue(t *testing.T, wantJob *store.Job) (called *bool) {
	called = new(bool)
	s.Enqueue_ = func(ctx context.Context, job *store.Job) error {
		*called = true
		if !reflect.DeepEqual(job, wantJob) {
			t.Errorf("got job %+v, want %+v", job, wantJob)
		}
		return nil
	}
	return
}

func (s *Queue) MockLockJob_Return(t *testing.T, job *store.Job) (called, calledSuccess, calledError *bool) {
	called = new(bool)
	calledSuccess = new(bool)
	calledError = new(bool)
	j := store.NewLockedJob(
		job,
		func() error {
			*calledSuccess = true
			return nil
		},
		func(_ string) error {
			*calledError = true
			return nil
		},
	)
	if job == nil {
		j = nil
	}
	s.LockJob_ = func(ctx context.Context) (*store.LockedJob, error) {
		*called = true
		return j, nil
	}
	return
}
