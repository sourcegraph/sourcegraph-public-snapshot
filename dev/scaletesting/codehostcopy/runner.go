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

func (r *Runner) List(ctx context.Context, limit int) error {
	out := output.NewOutput(os.Stdout, output.OutputOpts{})

	// Load existing repositories.
	srcRepos, err := r.store.Load()
	if err != nil {
		r.logger.Error("failed to open state database", log.Error(err))
		return err
	}
	loadedFromDB := true

	// If we're starting fresh, really fetch them.
	if len(srcRepos) == 0 {
		loadedFromDB = false
		r.logger.Info("No existing state found, creating ...")
		out.WriteLine(output.Line(output.EmojiHourglass, output.StyleBold, "Listing repos"))
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
	}
	block := out.Block(output.Line(output.EmojiInfo, output.StyleBold, fmt.Sprintf("List of repos (db: %v limit: %d total: %d)", loadedFromDB, limit, len(srcRepos))))
	if limit != 0 && limit < len(srcRepos) {
		srcRepos = srcRepos[:limit]
	}
	for _, r := range srcRepos {
		block.Writef("Name: %s Created: %v Pushed: %v GitURL: %s ToGitURL: %s Failed: %s", r.Name, r.Created, r.Pushed, r.GitURL, r.ToGitURL, r.Failed)
	}
	block.Close()
	return nil
}

func (r *Runner) Copy(ctx context.Context, concurrency int) error {
	out := output.NewOutput(os.Stdout, output.OutputOpts{})

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
					// If we resumed and this repo previously failed, we need to clear the failed status as it succeeded now
					repo.Failed = ""
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

	// we add the repo name so that we ensure we cd to the right repo directory
	// if we don't do this, there is no guarantee that the repo name and the git url are the same
	err = run.Bash(ctx, "git clone --mirror", repo.GitURL, repo.Name).Dir(tmpDir).Run().Wait()
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
        // --force, with mirror we want the remote to look exactly as we have it
		err = run.Bash(ctx, "git push --mirror --force destination").Dir(dir).Run().Wait()
		if err != nil {
			if strings.Contains(err.Error(), "timed out") {
				continue
			}
			return err
		}
	}
	return nil
}
