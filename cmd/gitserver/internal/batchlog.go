package internal

import (
	"bytes"
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/executil"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/gitserverfs"
	"github.com/sourcegraph/sourcegraph/internal/api"
	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Note: The BatchLog endpoint has been deprecated. This file shall be removed after
// the 5.3 release has been cut.

func (s *Server) batchGitLogInstrumentedHandler(ctx context.Context, req *proto.BatchLogRequest) (resp *proto.BatchLogResponse, err error) {
	// Perform requests in each repository in the input batch. We perform these commands
	// concurrently, but only allow for so many commands to be in-flight at a time so that
	// we don't overwhelm a shard with either a large request or too many concurrent batch
	// requests.

	g, ctx := errgroup.WithContext(ctx)
	results := make([]*proto.BatchLogResult, len(req.GetRepoCommits())) //nolint:staticcheck

	if s.GlobalBatchLogSemaphore == nil {
		return &proto.BatchLogResponse{}, errors.New("s.GlobalBatchLogSemaphore not initialized")
	}

	for i, repoCommit := range req.GetRepoCommits() { //nolint:staticcheck
		// Avoid capture of loop variables
		i, repoCommit := i, repoCommit

		if err := s.GlobalBatchLogSemaphore.Acquire(ctx, 1); err != nil {
			return resp, err
		}

		g.Go(func() error {
			defer s.GlobalBatchLogSemaphore.Release(1)

			output, isRepoCloned, gitLogErr := s.performGitLogCommand(ctx, repoCommit, req.GetFormat()) //nolint:staticcheck
			if gitLogErr == nil && !isRepoCloned {
				gitLogErr = errors.Newf("repo not found")
			}

			// Concurrently write results to shared slice. This slice is already properly
			// sized, and each goroutine writes to a unique index exactly once. There should
			// be no data race conditions possible here.

			results[i] = &proto.BatchLogResult{
				RepoCommit:    repoCommit,
				CommandOutput: output,
			}

			if gitLogErr != nil {
				errMessage := gitLogErr.Error()
				results[i].CommandError = &errMessage //nolint:staticcheck
			}

			return nil
		})
	}

	if err = g.Wait(); err != nil {
		return
	}
	return &proto.BatchLogResponse{Results: results}, nil
}

func (s *Server) performGitLogCommand(ctx context.Context, repoCommit *proto.RepoCommit, format string) (output string, isRepoCloned bool, err error) {
	dir := gitserverfs.RepoDirFromName(s.ReposDir, api.RepoName(repoCommit.GetRepo())) //nolint:staticcheck
	if !repoCloned(dir) {
		return "", false, nil
	}

	var buf bytes.Buffer

	commitId := repoCommit.GetCommit() //nolint:staticcheck
	// make sure CommitID is not an arg
	if commitId[0] == '-' {
		return "", true, errors.New("commit ID starting with - is not allowed")
	}

	cmd := s.RecordingCommandFactory.Command(ctx, s.Logger, repoCommit.GetRepo(), "git", "log", "-n", "1", "--name-only", format, commitId) //nolint:staticcheck
	dir.Set(cmd.Unwrap())
	cmd.Unwrap().Stdout = &buf

	if _, err := executil.RunCommand(ctx, cmd); err != nil {
		return "", true, err
	}

	return buf.String(), true, nil
}
