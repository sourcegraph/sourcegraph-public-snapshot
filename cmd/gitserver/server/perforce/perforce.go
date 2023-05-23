package perforce

import (
	"bufio"
	"bytes"
	"container/list"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server/common"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/perforce"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	Logger                 log.Logger
	DB                     database.DB
	changelistMappingQueue *changelistMappingQueue
}

func (s *Service) EnqueueChangelistMappingJob(job *ChangelistMappingJob) {
	s.changelistMappingQueue.push(job)
}

// NewCloneQueue initializes a new cloneQueue.
func NewService(logger log.Logger, db database.DB, jobs *list.List) Service {
	cq := changelistMappingQueue{jobs: jobs}
	cq.cond = sync.NewCond(&cq.cmu)

	return Service{
		changelistMappingQueue: &cq,
	}
}

type ChangelistMappingJob struct {
	RepoName api.RepoName
	RepoDir  common.GitDir
}

// FIXME: Use generics.
type changelistMappingQueue struct {
	mu   sync.Mutex
	jobs *list.List

	cmu  sync.Mutex
	cond *sync.Cond
}

// push will queue the cloneJob to the end of the queue.
func (p *changelistMappingQueue) push(pj *ChangelistMappingJob) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.jobs.PushBack(pj)
	p.cond.Signal()
}

// pop will return the next cloneJob. If there's no next job available, it returns nil.
func (p *changelistMappingQueue) pop() *ChangelistMappingJob {
	p.mu.Lock()
	defer p.mu.Unlock()

	next := p.jobs.Front()
	if next == nil {
		return nil
	}

	return p.jobs.Remove(next).(*ChangelistMappingJob)
}

func (p *changelistMappingQueue) empty() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.jobs.Len() == 0
}

func (s *Service) StartPerforceChangelistMappingPipeline(ctx context.Context) {
	jobs := make(chan *ChangelistMappingJob)
	go s.changelistMappingConsumer(ctx, jobs)
	go s.changelistMappingProducer(ctx, jobs)
}

func (s *Service) changelistMappingProducer(ctx context.Context, jobs chan<- *ChangelistMappingJob) {
	defer close(jobs)

	for {
		s.changelistMappingQueue.cmu.Lock()
		if s.changelistMappingQueue.empty() {
			s.changelistMappingQueue.cond.Wait()
		}

		s.changelistMappingQueue.cmu.Unlock()

		for {
			job := s.changelistMappingQueue.pop()
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

func (s *Service) changelistMappingConsumer(ctx context.Context, jobs <-chan *ChangelistMappingJob) {
	logger := s.Logger.Scoped("changelistMappingConsumer", "process perforce changelist mapping jobs")

	// Process only one job at a time for a simpler pipeline at the moment.
	for j := range jobs {
		logger := logger.With(log.String("job.repo", string(j.RepoName)))

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

func (s *Service) doChangelistMapping(ctx context.Context, job *ChangelistMappingJob) error {
	logger := s.Logger.Scoped("doChangelistMapping", "").With(
		log.String("repo", string(job.RepoName)),
	)

	logger.Warn("started")

	repo, err := s.DB.Repos().GetByName(ctx, job.RepoName)
	if err != nil {
		return errors.Wrap(err, "Repos.GetByName")
	}

	if repo.ExternalRepo.ServiceType != extsvc.TypePerforce {
		logger.Warn("skipping non-perforce depot (this is not a regression but someone is likely pushing non perforce depots into the queue and creating NOOP jobs)")
		return nil
	}

	logger.Warn("repo received from DB")

	// dir := s.dir(protocol.NormalizeRepo(repo.Name))
	dir := job.RepoDir

	logger.Warn("latestRowCommit")

	commitsMap, err := s.getCommitsToInsert(ctx, logger, repo.ID, dir)
	if err != nil {
		return err
	}

	// We want to write all the commits or nothing at all in a single transaction to avoid partially
	// succesful mapping jobs which will make it difficult to determine missing commits that need to
	// be mapped. This makes it easy to have a reliable start point for the next time this job is
	// attempted, knowing for sure that the latest commit in the DB is indeed the last point from
	// which we need to resume the mapping.
	err = s.DB.RepoCommitsChangelists().BatchInsertCommitSHAsWithPerforceChangelistID(ctx, repo.ID, commitsMap)
	if err != nil {
		return err
	}

	return nil
}

// FIXME
func runWith(ctx context.Context, cmd wrexec.Cmder, configRemoteOpts bool, progress io.Writer) ([]byte, error) {
	return nil, nil
}

func headCommitSHA(ctx context.Context, logger log.Logger, dir common.GitDir) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	dir.Set(cmd)
	output, err := runWith(ctx, wrexec.Wrap(ctx, logger, cmd), false, nil)
	if err != nil {
		return "", &common.GitCommandError{Err: err, Output: string(output)}
	}

	return string(bytes.TrimSpace(output)), nil
}

func (s *Service) getCommitsToInsert(ctx context.Context, logger log.Logger, repoID api.RepoID, dir common.GitDir) (commitsMap []types.PerforceChangelist, err error) {
	latestRowCommit, err := s.DB.RepoCommitsChangelists().GetLatestForRepo(ctx, repoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// This repo has not been imported into the RepoCommits table yet. Start from the beginning.
			results, err := newMappableCommits(ctx, logger, dir, "", "")
			return results, errors.Wrap(err, "failed to import new repo (perforce changelists will have limited functionality)")
		}

		return nil, errors.Wrap(err, "RepoCommits.GetLatestForRepo")
	}

	logger.Warn("continuing from latestCommit")

	head, err := headCommitSHA(ctx, logger, dir)
	if err != nil {
		return nil, errors.Wrap(err, "headCommitSHA")
	}

	if latestRowCommit != nil && string(latestRowCommit.CommitSHA) == head {
		logger.Info("repo commits already mapped upto HEAD, skipping", log.String("HEAD", head))
		return nil, nil
	}

	results, err := newMappableCommits(ctx, logger, dir, string(latestRowCommit.CommitSHA), head)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to import existing repo's commits after HEAD: %q", head)
	}

	return results, nil
}

// logFormatWithCommitSHAAndCommitBodyOnly will print the commit SHA and the commit body (skips the
// subject) separated by a space. These are the only two fields that we need to parse the changelist
// ID from the commit.
//
// Example:
// $ git log --format='format:%H %b'
// 4e5b9dbc6393b195688a93ea04b98fada50bfa03 [p4-fusion: depot-paths = "//rhia-depot-test/": change = 83733]
// e2f6d6e306490831b0fdd908fdbee702d7074a66 [p4-fusion: depot-paths = "//rhia-depot-test/": change = 83732]
// 90b9b9574517f30810346f0ab07f66c49c77ab0f [p4-fusion: depot-paths = "//rhia-depot-test/": change = 83731]
var logFormatWithCommitSHAAndCommitBodyOnly = "--format=format:%H %b"

func newMappableCommits(ctx context.Context, logger log.Logger, dir common.GitDir, lastMappedCommit, head string) ([]types.PerforceChangelist, error) {
	cmd := exec.CommandContext(ctx, "git", "log")
	// FIXME: When lastMappedCommit..head is an invalid range.
	// TODO: Follow up in a separate PR.
	if lastMappedCommit != "" {
		cmd.Args = append(cmd.Args, fmt.Sprintf("%s..%s", lastMappedCommit, head))
	}

	cmd.Args = append(cmd.Args, logFormatWithCommitSHAAndCommitBodyOnly)
	dir.Set(cmd)

	progressReader, progressWriter := io.Pipe()

	logLineResults := make(chan string)
	g, ctx := errgroup.WithContext(ctx)

	// Start reading the output of the command in a goroutine.
	g.Go(func() error {
		defer close(logLineResults)
		return readGitLogOutput(ctx, logger, progressReader, logLineResults)
	})

	// Run the command in a goroutine. It will start writing the output to progressWriter.
	g.Go(func() error {
		defer progressWriter.Close()

		output, err := runWith(ctx, wrexec.Wrap(ctx, logger, cmd), false, progressWriter)
		if err != nil {
			return &common.GitCommandError{Err: err, Output: string(output)}
		}
		return nil
	})

	go func() {
		g.Wait()
	}()

	// Collect the results.
	commitMaps := []types.PerforceChangelist{}
	for line := range logLineResults {
		// FIXME: Something about this is generating an empty newline after each log line in tests.
		// Skip an empty newline for the next output.
		if line == "" {
			continue
		}

		c, err := parseGitLogLine(line)
		if err != nil {
			return nil, err
		}

		commitMaps = append(commitMaps, *c)
	}

	// In case any of the goroutines failed, collect and return the error.
	return commitMaps, errors.Wrap(g.Wait(), "command exeuction pipeline failed")
}

func readGitLogOutput(ctx context.Context, logger log.Logger, reader io.Reader, logLineResults chan<- string) error {
	scan := bufio.NewScanner(reader)
	scan.Split(bufio.ScanLines)
	for scan.Scan() {
		line := scan.Text()

		select {
		case logLineResults <- line:
			// return errors.New("early exit")
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return errors.Wrap(scan.Err(), "scanning git-log output failed")
}

func parseGitLogLine(line string) (*types.PerforceChangelist, error) {
	// Expected format: "<commitSHA> <commitBody>"
	parts := strings.SplitN(line, " ", 2)
	if len(parts) != 2 {
		return nil, errors.Newf("failed to split line %q from git log output into commitSHA and commit body, parts after splitting: %d", line, len(parts))
	}

	parsedCID, err := perforce.GetP4ChangelistID(parts[1])
	if err != nil {
		return nil, err
	}

	cid, err := strconv.ParseInt(parsedCID, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse changelist ID to int64")
	}

	return &types.PerforceChangelist{
		CommitSHA:    api.CommitID(parts[0]),
		ChangelistID: cid,
	}, nil
}
