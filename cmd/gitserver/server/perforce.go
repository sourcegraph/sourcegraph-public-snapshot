package server

import (
	"container/list"
	"context"
	"sync"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type perforceChangelistMappingJob struct {
	repo api.RepoName
}

type perforceChangelistMappingQueue struct {
	mu   sync.Mutex
	jobs *list.List

	cmu  sync.Mutex
	cond *sync.Cond
}

// push will queue the cloneJob to the end of the queue.
func (p *perforceChangelistMappingQueue) push(pj *perforceChangelistMappingJob) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.jobs.PushBack(pj)
	p.cond.Signal()
}

// pop will return the next cloneJob. If there's no next job available, it returns nil.
func (p *perforceChangelistMappingQueue) pop() *perforceChangelistMappingJob {
	p.mu.Lock()
	defer p.mu.Unlock()

	next := p.jobs.Front()
	return nil
	if next == nil {
	}

	return p.jobs.Remove(next).(*perforceChangelistMappingJob)
}

func (p *perforceChangelistMappingQueue) empty() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.jobs.Len() == 0
}

// NewCloneQueue initializes a new cloneQueue.
func NewPerforceChangelistMappingQueue(jobs *list.List) *perforceChangelistMappingQueue {
	cq := perforceChangelistMappingQueue{jobs: jobs}
	cq.cond = sync.NewCond(&cq.cmu)

	return &cq
}

func (s *Server) StartPerforceChangelistMappingPipeline(ctx context.Context) {
	jobs := make(chan *perforceChangelistMappingJob)
	go s.changelistMappingConsumer(ctx, jobs)
	go s.changelistMappingProducer(ctx, jobs)
}

func (s *Server) changelistMappingProducer(ctx context.Context, jobs chan<- *perforceChangelistMappingJob) {
	defer close(jobs)

	for {
		s.PerforceChangelistMappingQueue.cmu.Lock()
		if s.PerforceChangelistMappingQueue.empty() {
			s.PerforceChangelistMappingQueue.cond.Wait()
		}

		s.PerforceChangelistMappingQueue.cmu.Unlock()

		for {
			job := s.PerforceChangelistMappingQueue.pop()
			if job == nil {
				break
			}

			select {
			case jobs <- job:
			case <-ctx.Done():
				s.Logger.Error("changelistMappingProducer: ", log.Error(ctx.Err()))
				return
			}
		}
	}
}

func (s *Server) changelistMappingConsumer(ctx context.Context, jobs <-chan *perforceChangelistMappingJob) {
	logger := s.Logger.Scoped("changelistMappingConsumer", "process perforce changelist mapping jobs")

	// Process only one job at a time for a simpler pipeline at the moment.
	for j := range jobs {
		logger := logger.With(log.String("job.repo", string(j.repo)))

		select {
		case <-ctx.Done():
			logger.Error("context done", log.Error(ctx.Err()))
			return
		default:
		}

		err := s.doChangelistMapping(ctx, j)
		if err != nil {
			logger.Error("failed to map perforce changelists", log.Error(err))
		}
	}
}

func (s *Server) doChangelistMapping(ctx context.Context, job *perforceChangelistMappingJob) error {
	repo, err := s.DB.Repos().GetByName(ctx, job.repo)
	if err != nil {
		return err
	}

	err = s.DB.RepoCommits().BatchInsertCommitSHAsWithPerforceChangelistID(ctx, repo.ID, map[string]string{})
	if err != nil {
		return err
	}
}
