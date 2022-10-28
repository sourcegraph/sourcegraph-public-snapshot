package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/run"
	"github.com/sourcegraph/sourcegraph/dev/scaletesting/internal/store"
	"github.com/sourcegraph/sourcegraph/lib/group"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

type Runner struct {
	source      CodeHostSource
	destination CodeHostDestination
	store       *store.Store
	logger      log.Logger
}

func logRepo(r *store.Repo, fields ...log.Field) []log.Field {
	return append([]log.Field{
		log.Object("repo",
			log.String("name", r.Name),
			log.String("from", r.GitURL),
			log.String("to", r.ToGitURL),
		),
	}, fields...)
}

func NewRunner(logger log.Logger, s *store.Store, source CodeHostSource, dest CodeHostDestination) *Runner {
	return &Runner{
		logger:      logger,
		source:      source,
		destination: dest,
		store:       s,
	}
}

func (r *Runner) Run(ctx context.Context, concurrency int) error {
	out := output.NewOutput(os.Stdout, output.OutputOpts{})
	r.logger.Info("test")

	// Load existing repositories.
	srcRepos, err := r.store.Load()
	if err != nil {
		r.logger.Error("failed to open state database", log.Error(err))
		return err
	}

	// If we're starting fresh, really fetch them.
	if len(srcRepos) == 0 {
		r.logger.Info("No existing state found, creating ...")
		repos, err := r.source.ListRepos(ctx)
		if err != nil {
			r.logger.Error("failed to list repositories from source", log.Error(err))
			return err
		}
		srcRepos = repos
		if err := r.store.Insert(repos); err != nil {
			r.logger.Error("failed to insert repositories from source", log.Error(err))
			return err
		}
		r.logger.Info(fmt.Sprintf("Found %d repos in source", len(srcRepos)))
	} else {
		r.logger.Info(fmt.Sprintf("Resuming work (%d repos)", len(srcRepos)))
	}

	bars := []output.ProgressBar{
		{Label: "Copying repos", Max: float64(len(srcRepos))},
	}
	progress := out.Progress(bars, nil)
	defer progress.Destroy()

	var done int64
	total := len(srcRepos)

	g := group.NewWithResults[error]().WithMaxConcurrency(20)
	for _, repo := range srcRepos {
		repo := repo
		g.Go(func() error {
			// Create the repo on destination.
			if !repo.Created {
				toGitURL, err := r.destination.CreateRepo(ctx, repo.Name)
				if err != nil {
					repo.Failed = err.Error()
					r.logger.Error("failed to create repo", logRepo(repo, log.Error(err))...)
				} else {
					repo.ToGitURL = toGitURL.String()
					repo.Created = true
				}
				if err := r.store.SaveRepo(repo); err != nil {
					r.logger.Error("failed to save repo", logRepo(repo, log.Error(err))...)
					return err
				}
			}

			// Push the repo on destination.
			if !repo.Pushed && repo.Created {
				err := pushRepo(ctx, repo)
				if err != nil {
					repo.Failed = err.Error()
					r.logger.Error("failed to push repo", logRepo(repo, log.Error(err))...)
				} else {
					repo.Pushed = true
				}
				if err := r.store.SaveRepo(repo); err != nil {
					r.logger.Error("failed to save repo", logRepo(repo, log.Error(err))...)
					return err
				}
			}
			atomic.AddInt64(&done, 1)
			progress.SetValue(0, float64(done))
			progress.SetLabel(0, fmt.Sprintf("Copying repos (%d/%d)", done, total))
			return nil
		})
	}
	errs := g.Wait()
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func pushRepo(ctx context.Context, repo *store.Repo) error {
	tmpDir, err := os.MkdirTemp(os.TempDir(), fmt.Sprintf("repo__%s", repo.Name))
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	err = run.Bash(ctx, "git clone", repo.GitURL).Dir(tmpDir).Run().Wait()
	if err != nil {
		return err
	}
	repoDir := filepath.Join(tmpDir, repo.Name)
	err = run.Bash(ctx, "git remote add destination", repo.ToGitURL).Dir(repoDir).Run().Wait()
	if err != nil {
		return err
	}
	return gitPushWithRetry(ctx, repoDir, 3)
}

func gitPushWithRetry(ctx context.Context, dir string, retry int) error {
	var err error
	for i := 0; i < retry; i++ {
		err = run.Bash(ctx, "git push destination").Dir(dir).Run().Wait()
		if err != nil {
			if strings.Contains(err.Error(), "timed out") {
				continue
			}
			return err
		}
	}
	return nil
}
