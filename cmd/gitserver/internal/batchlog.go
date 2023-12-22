package internal

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/executil"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/gitserverfs"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (s *Server) batchGitLogInstrumentedHandler(ctx context.Context, req protocol.BatchLogRequest) (resp protocol.BatchLogResponse, err error) {
	ctx, _, endObservation := s.operations.batchLog.With(ctx, &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{Attrs: []attribute.KeyValue{
			attribute.String("results", fmt.Sprintf("%+v", resp.Results)),
		}})
	}()

	// Perform requests in each repository in the input batch. We perform these commands
	// concurrently, but only allow for so many commands to be in-flight at a time so that
	// we don't overwhelm a shard with either a large request or too many concurrent batch
	// requests.

	g, ctx := errgroup.WithContext(ctx)
	results := make([]protocol.BatchLogResult, len(req.RepoCommits))

	if s.GlobalBatchLogSemaphore == nil {
		return protocol.BatchLogResponse{}, errors.New("s.GlobalBatchLogSemaphore not initialized")
	}

	for i, repoCommit := range req.RepoCommits {
		// Avoid capture of loop variables
		i, repoCommit := i, repoCommit

		start := time.Now()
		if err := s.GlobalBatchLogSemaphore.Acquire(ctx, 1); err != nil {
			return resp, err
		}
		s.operations.batchLogSemaphoreWait.Observe(time.Since(start).Seconds())

		g.Go(func() error {
			defer s.GlobalBatchLogSemaphore.Release(1)

			output, isRepoCloned, gitLogErr := s.performGitLogCommand(ctx, repoCommit, req.Format)
			if gitLogErr == nil && !isRepoCloned {
				gitLogErr = errors.Newf("repo not found")
			}
			var errMessage string
			if gitLogErr != nil {
				errMessage = gitLogErr.Error()
			}

			// Concurrently write results to shared slice. This slice is already properly
			// sized, and each goroutine writes to a unique index exactly once. There should
			// be no data race conditions possible here.

			results[i] = protocol.BatchLogResult{
				RepoCommit:    repoCommit,
				CommandOutput: output,
				CommandError:  errMessage,
			}
			return nil
		})
	}

	if err = g.Wait(); err != nil {
		return
	}
	return protocol.BatchLogResponse{Results: results}, nil
}

func (s *Server) performGitLogCommand(ctx context.Context, repoCommit api.RepoCommit, format string) (output string, isRepoCloned bool, err error) {
	ctx, _, endObservation := s.operations.batchLogSingle.With(ctx, &err, observation.Args{
		Attrs: append(
			[]attribute.KeyValue{
				attribute.String("format", format),
			},
			repoCommit.Attrs()...,
		),
	})
	defer func() {
		endObservation(1, observation.Args{Attrs: []attribute.KeyValue{
			attribute.Bool("isRepoCloned", isRepoCloned),
		}})
	}()

	dir := gitserverfs.RepoDirFromName(s.ReposDir, repoCommit.Repo)
	if !repoCloned(dir) {
		return "", false, nil
	}

	var buf bytes.Buffer

	commitId := string(repoCommit.CommitID)
	// make sure CommitID is not an arg
	if commitId[0] == '-' {
		return "", true, errors.New("commit ID starting with - is not allowed")
	}

	cmd := s.RecordingCommandFactory.Command(ctx, s.Logger, string(repoCommit.Repo), "git", "log", "-n", "1", "--name-only", format, commitId)
	dir.Set(cmd.Unwrap())
	cmd.Unwrap().Stdout = &buf

	if _, err := executil.RunCommand(ctx, cmd); err != nil {
		return "", true, err
	}

	return buf.String(), true, nil
}
