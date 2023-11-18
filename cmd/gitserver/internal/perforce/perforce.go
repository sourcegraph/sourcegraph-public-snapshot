package perforce

import (
	"bufio"
	"bytes"
	"container/list"
	"context"
	"database/sql"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/perforce"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type changelistMappingJob struct {
	RepoName api.RepoName
	RepoDir  common.GitDir
}

func NewChangelistMappingJob(repoName api.RepoName, repoDir common.GitDir) *changelistMappingJob {
	return &changelistMappingJob{
		RepoName: repoName,
		RepoDir:  repoDir,
	}
}

// Service is used to manage perforce depot related interactions from gitserver.
//
// NOTE: Use NewService to instantiate a new service to ensure all other side effects of creating a
// new service are taken care of.
type Service struct {
	Logger log.Logger
	DB     database.DB

	ctx                    context.Context
	changelistMappingQueue *common.Queue[*changelistMappingJob]
}

// NewService initializes a new service with a queue and starts a producer-consumer pipeline that
// will read jobs from the queue and "produce" them for "consumption".
func NewService(ctx context.Context, obctx *observation.Context, logger log.Logger, db database.DB, jobs *list.List) *Service {
	queue := common.NewQueue[*changelistMappingJob](obctx, "perforce-changelist-mapper", jobs)

	s := &Service{
		Logger: logger,
		DB:     db,

		ctx:                    ctx,
		changelistMappingQueue: queue,
	}

	s.startPerforceChangelistMappingPipeline(ctx)

	return s
}

// EnqueueChangelistMappingJob will push the ChangelistMappingJob onto the queue iff the
// experimental config for PerforceChangelistMapping is enabled and if the repo belongs to a code
// host of type PERFORCE.
func (s *Service) EnqueueChangelistMappingJob(job *changelistMappingJob) {
	if c := conf.Get(); c.ExperimentalFeatures == nil || c.ExperimentalFeatures.PerforceChangelistMapping != "enabled" {
		return
	}

	if r, err := s.DB.Repos().GetByName(s.ctx, job.RepoName); err != nil {
		s.Logger.Warn("failed to retrieve repo from DB (this could be a data inconsistency)", log.Error(err))
	} else if r.ExternalRepo.ServiceType == extsvc.VariantPerforce.AsType() {
		s.changelistMappingQueue.Push(job)
	}
}

func (s *Service) startPerforceChangelistMappingPipeline(ctx context.Context) {
	tasks := make(chan *changelistMappingTask)

	// Protect against panics.
	goroutine.Go(func() { s.changelistMappingConsumer(ctx, tasks) })
	goroutine.Go(func() { s.changelistMappingProducer(ctx, tasks) })
}

// changelistMappingProducer "pops" jobs from the FIFO queue of the "Service" and produce them
// for "consumption".
func (s *Service) changelistMappingProducer(ctx context.Context, tasks chan<- *changelistMappingTask) {
	defer close(tasks)

	for {
		s.changelistMappingQueue.Mutex.Lock()
		if s.changelistMappingQueue.Empty() {
			s.changelistMappingQueue.Cond.Wait()
		}

		s.changelistMappingQueue.Mutex.Unlock()

		for {
			job, doneFunc := s.changelistMappingQueue.Pop()
			if job == nil {
				break
			}

			select {
			case tasks <- &changelistMappingTask{
				changelistMappingJob: *job,
				done:                 doneFunc,
			}:
			case <-ctx.Done():
				s.Logger.Error("changelistMappingProducer: ", log.Error(ctx.Err()))
				return
			}
		}
	}
}

// changelistMappingConsumer "consumes" jobs "produced" by the producer.
func (s *Service) changelistMappingConsumer(ctx context.Context, tasks <-chan *changelistMappingTask) {
	logger := s.Logger.Scoped("changelistMappingConsumer")

	// Process only one job at a time for a simpler pipeline at the moment.
	for task := range tasks {
		logger := logger.With(log.String("job.repo", string(task.RepoName)))

		select {
		case <-ctx.Done():
			logger.Error("context done", log.Error(ctx.Err()))
			return
		default:
		}

		err := s.doChangelistMapping(ctx, task.changelistMappingJob)
		if err != nil {
			logger.Error("failed to map perforce changelists", log.Error(err))
		}

		timeTaken := task.done()
		// NOTE: Hardcoded to log for tasks that run longer than 1 minute. Will revisit this if it
		// becomes noisy under production loads.
		if timeTaken > time.Duration(time.Second*60) {
			s.Logger.Warn("mapping job took long to complete", log.Duration("duration", timeTaken))
		}
	}
}

// doChangelistMapping performs the commits -> changelist ID mapping for a new or existing repo.
func (s *Service) doChangelistMapping(ctx context.Context, job *changelistMappingJob) error {
	logger := s.Logger.Scoped("doChangelistMapping").With(
		log.String("repo", string(job.RepoName)),
	)

	logger.Debug("started")

	repo, err := s.DB.Repos().GetByName(ctx, job.RepoName)
	if err != nil {
		return errors.Wrap(err, "Repos.GetByName")
	}

	if repo.ExternalRepo.ServiceType != extsvc.TypePerforce {
		logger.Warn("skipping non-perforce depot (this is not a regression but someone is likely pushing non perforce depots into the queue and creating NOOP jobs)")
		return nil
	}

	dir := job.RepoDir

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

// getCommitsToInsert returns a list of commitsSHA -> changelistID for each commit that is yet to
// be "mapped" in the DB. For new repos, this will contain all the commits and for existing repos it
// will only return the commits yet to be mapped in the DB.
//
// It returns an error if any.
func (s *Service) getCommitsToInsert(ctx context.Context, logger log.Logger, repoID api.RepoID, dir common.GitDir) (commitsMap []types.PerforceChangelist, err error) {
	latestRowCommit, err := s.DB.RepoCommitsChangelists().GetLatestForRepo(ctx, repoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// This repo has not been imported into the RepoCommits table yet. Start from the beginning.
			results, err := newMappableCommits(ctx, dir, "", "")
			return results, errors.Wrap(err, "failed to import new repo (perforce changelists will have limited functionality)")
		}

		return nil, errors.Wrap(err, "RepoCommits.GetLatestForRepo")
	}

	head, err := headCommitSHA(ctx, dir)
	if err != nil {
		return nil, errors.Wrap(err, "headCommitSHA")
	}

	if latestRowCommit != nil && string(latestRowCommit.CommitSHA) == head {
		logger.Info("repo commits already mapped upto HEAD, skipping", log.String("HEAD", head))
		return nil, nil
	}

	results, err := newMappableCommits(ctx, dir, string(latestRowCommit.CommitSHA), head)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to import existing repo's commits after HEAD: %q", head)
	}

	return results, nil
}

// headCommitSHA returns the commitSHA at HEAD of the repo.
func headCommitSHA(ctx context.Context, dir common.GitDir) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	dir.Set(cmd)

	output, err := cmd.Output()
	if err != nil {
		return "", &common.GitCommandError{Err: err, Output: string(output)}
	}

	return string(bytes.TrimSpace(output)), nil
}

// logFormatWithCommitSHAAndCommitBodyOnly prints the commit SHA and the commit subject and body
// separated by a single space. These are the only three fields that we need to parse the changelist
// ID from the commit.
//
// Normally just the commit SHA and the commit body would suffice, but in practice we have noticed
// some commits generated by p4-fusion end up not having a blank line between the subject and the
// body. As a result, the body is empty and the subject contains the metadata that we're looking
// for.
//
// For example:
//
// 48485 - test-5386
// [p4-fusion: depot-paths = "//go/": change = 48485]
//
// VS:
//
// 1012 - :boar:
//
// [p4-fusion: depot-paths = "//go/": change = 1012]
//
// To handle such edge cases we print both the subject and the body together without any spaces.
// This still ensures that we have only two parts per commit in the output:
// <commit SHA, (commit subject + commit body)>
//
// And the regex pattern can still work because it is not anchored to look for the start or end of a
// line and we search for a sub-string match so anything before the actual metadata is ignored.
//
// For both the cases when the p4-fusion metadata has or does not have a blank line between the
// subject and the body, the output will look like:
//
// $ git log --format='format:%H %s%b'
// 4e5b9dbc6393b195688a93ea04b98fada50bfa03 83733 - Removing this from the workspace[p4-fusion: depot-paths = "//rhia-depot-test/": change = 83733]
// e2f6d6e306490831b0fdd908fdbee702d7074a66 83732 - adding sourcegraph repos[p4-fusion: depot-paths = "//rhia-depot-test/": change = 83732]
// 90b9b9574517f30810346f0ab07f66c49c77ab0f 83731 - "initial commit"[p4-fusion: depot-paths = "//rhia-depot-test/": change = 83731]
var logFormatWithCommitSHAAndCommitBodyOnly = "--format=format:%H %s%b%x00"

// newMappableCommits executes git log with "logFormatWithCommitSHAAndCommitBodyOnly" as the format
// specifier and return a list of commitsSHA -> changelistID for each commit between the range
// "lastMappedCommit..HEAD".
//
// If "lastMappedCommit" is empty, it will return the list for all commits of this repo.
//
// newMappableCommits will read the output one commit at a time to avoid an unbounded memory growth.
func newMappableCommits(ctx context.Context, dir common.GitDir, lastMappedCommit, head string) ([]types.PerforceChangelist, error) {
	// ensure we cleanup command when returning
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "log", logFormatWithCommitSHAAndCommitBodyOnly)
	// FIXME: When lastMappedCommit..head is an invalid range.
	// TODO: Follow up in a separate PR.
	if lastMappedCommit != "" {
		cmd.Args = append(cmd.Args, fmt.Sprintf("%s..%s", lastMappedCommit, head))
	}
	dir.Set(cmd)

	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create stdout pipe for command")
	}

	if err := cmd.Start(); err != nil {
		return nil, errors.Wrap(err, "failed to start command")
	}

	scan := bufio.NewScanner(out)
	scan.Split(scanNull)

	commitMaps := []types.PerforceChangelist{}
	for scan.Scan() {
		c, err := parseGitLogLine(strings.TrimSpace(scan.Text()))
		if err != nil {
			return nil, err
		}

		commitMaps = append(commitMaps, *c)
	}

	return commitMaps, errors.Wrap(cmd.Wait(), "command execution pipeline failed")
}

func scanNull(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, 0); i >= 0 {
		return i + 1, data[0:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

// parseGitLogLine will parse the a line from the git-log output and return the commitSHA and changelistID.
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

// changelistMappingTask is a thin wrapper around a changelistMappingJob to associate the
// doneFunc with each job.
type changelistMappingTask struct {
	*changelistMappingJob
	done func() time.Duration
}
