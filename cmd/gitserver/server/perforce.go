package server

import (
	"container/list"
	"context"
	"fmt"
	"os/exec"
	"sync"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
	logger := s.Logger.Scoped("doChangelistMapping", "").With(
		log.String("repo", string(job.repo)),
	)

	repo, err := s.DB.Repos().GetByName(ctx, job.repo)
	if err != nil {
		return errors.Wrap(err, "Repos.GetByName")
	}

	dir := s.dir(protocol.NormalizeRepo(repo.Name))

	latestRowCommit, err := s.DB.RepoCommits().GetLatestForRepo(ctx, repo.ID)
	if err != nil {
		return errors.Wrap(err, "RepoCommits.GetLatestForRepo")
	}

	var commitsMap []types.PerforceChangelist
	// This repo has not been imported into the RepoCommits table yet. Start from the beginning.
	if latestRowCommit != nil {
		head, err := headCommitSHA(ctx, logger, dir)
		if err != nil {
			return errors.Wrap(err, "headCommitSHA")
		}

		if latestRowCommit.CommitSHA == head {
			logger.Info("repo commits already mapped upto HEAD, skipping", log.String("HEAD", head))
			return nil
		}

		commitsMap, err = newMappableCommits(ctx, logger, dir, latestRowCommit.CommitSHA, head)
		if err != nil {
			return nil
		}
	} else {
		commitsMap, err = newMappableCommits(ctx, logger, dir, "", "")
		if err != nil {
			return nil
		}
	}

	totalCommits := len(commitsMap)

	// TODO: Do we want to make this configurable?
	step := 1000
	for i := 0; i < totalCommits; i += step {
		seek := i + step
		if seek > totalCommits {
			seek = totalCommits
		}

		chunk := commitsMap[i:seek]

		err = s.DB.RepoCommits().BatchInsertCommitSHAsWithPerforceChangelistID(ctx, repo.ID, chunk)
		if err != nil {
			return err
		}
	}

	return nil
}

func headCommitSHA(ctx context.Context, logger log.Logger, dir GitDir) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	dir.Set(cmd)
	output, err := runWith(ctx, wrexec.Wrap(ctx, logger, cmd), false, nil)
	if err != nil {
		return "", &GitCommandError{Err: err, Output: string(output)}
	}

	return string(output), nil
}

func newMappableCommits(ctx context.Context, logger log.Logger, dir GitDir, lastMappedCommit, head string) ([]types.PerforceChangelist, error) {
	cmd := exec.CommandContext(ctx, "git", "log")
	if lastMappedCommit != "" {
		cmd.Args = append(cmd.Args, fmt.Sprintf("%s..%s", lastMappedCommit, head))
	}

	dir.Set(cmd)
	output, err := runWith(ctx, wrexec.Wrap(ctx, logger, cmd), false, nil)
	if err != nil {
		return nil, &GitCommandError{Err: err, Output: string(output)}
	}

	commits, err := gitserver.ParseGitLogOutput(output)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse git log output")
	}

	data := make([]types.PerforceChangelist, len(commits))
	for i, commit := range commits {
		cid, err := getP4ChangelistID(commit.Message.Body())
		if err != nil {
			return nil, errors.Wrap(err, "getP4ChangelistID")
		}

		data[i] = types.PerforceChangelist{CommitSHA: commit.ID, ChangelistID: cid}
	}

	return data, nil
}

var gitP4Pattern = lazyregexp.New(`\[(?:git-p4|p4-fusion): depot-paths = "(.*?)"\: change = (\d+)\]`)

func getP4ChangelistID(body string) (string, error) {
	matches := gitP4Pattern.FindStringSubmatch(body)
	if len(matches) != 3 {
		return "", errors.Newf("failed to retrieve changelist ID from commit body: %q", body)
	}

	return matches[2], nil
}
