package apiworker

import (
	"context"
	"errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/apiworker/apiclient"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type storeShim struct {
	queueName  string
	queueStore QueueStore
}

type QueueStore interface {
	Dequeue(ctx context.Context, queueName string, payload *apiclient.Job) (bool, error)
	SetLogContents(ctx context.Context, queueName string, jobID int, contents string) error
	MarkComplete(ctx context.Context, queueName string, jobID int) error
	MarkErrored(ctx context.Context, queueName string, jobID int, errorMessage string) error
	MarkFailed(ctx context.Context, queueName string, jobID int, errorMessage string) error
}

var _ workerutil.Store = &storeShim{}

func (s *storeShim) QueuedCount(ctx context.Context, extraArguments interface{}) (int, error) {
	return 0, errors.New("unimplemented")
}

func (s *storeShim) Dequeue(ctx context.Context, extraArguments interface{}) (workerutil.Record, workerutil.Store, bool, error) {
	var job apiclient.Job
	dequeued, err := s.queueStore.Dequeue(ctx, s.queueName, &job)
	if err != nil {
		return nil, nil, false, err
	}

	return job, s, dequeued, nil
}

func (s *storeShim) SetLogContents(ctx context.Context, id int, payload string) error {
	return s.queueStore.SetLogContents(ctx, s.queueName, id, payload)
}

func (s *storeShim) MarkComplete(ctx context.Context, id int) (bool, error) {
	return true, s.queueStore.MarkComplete(ctx, s.queueName, id)
}

func (s *storeShim) MarkErrored(ctx context.Context, id int, errorMessage string) (bool, error) {
	return true, s.queueStore.MarkErrored(ctx, s.queueName, id, errorMessage)
}

func (s *storeShim) MarkFailed(ctx context.Context, id int, errorMessage string) (bool, error) {
	return true, s.queueStore.MarkFailed(ctx, s.queueName, id, errorMessage)
}

func (s *storeShim) Done(err error) error {
	return err
}
