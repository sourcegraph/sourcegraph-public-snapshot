package main

import (
	"context"
	"net/url"
	"os"
	"strings"
	"sync/atomic"

	"github.com/sourcegraph/run"
	"github.com/sourcegraph/sourcegraph/lib/group"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

type Runner struct {
	source      CodeHostSource
	destination CodeHostDestination
}

func NewRunner(source CodeHostSource, dest CodeHostDestination) *Runner {
	return &Runner{
		source:      source,
		destination: dest,
	}
}

func (r *Runner) Run(ctx context.Context, concurrency int) error {
	out := output.NewOutput(os.Stdout, output.OutputOpts{})

	srcRepos, err := r.source.ListRepos(ctx)
	if err != nil {
		return err
	}

	bars := []output.ProgressBar{
		{Label: "Copying repos", Max: float64(len(srcRepos))},
	}
	progress := out.Progress(bars, nil)
	defer progress.Destroy()

	var done int64
	err = inTempFolder(func() error {
		g := group.NewWithResults[error]().WithMaxConcurrency(20)
		for _, repo := range srcRepos {
			repo := repo
			g.Go(func() error {
				gitURL, err := r.destination.CreateRepo(ctx, repo.name)
				if err != nil {
					return err
				}

				err = uploadRepo(ctx, repo, gitURL)
				if err != nil {
					return err
				}
				atomic.AddInt64(&done, 1)
				progress.SetValue(0, float64(done))
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
	})
	if err != nil {
		return err
	}
	return nil
}

func uploadRepo(ctx context.Context, repo *Repo, gitURL *url.URL) error {
	err := run.Bash(ctx, "git clone", repo.url).Run().Wait()
	if err != nil {
		return err
	}
	err = run.Bash(ctx, "git remote add destination", gitURL.String()).Dir(repo.name).Run().Wait()
	if err != nil {
		return err
	}
	return gitPushWithRetry(ctx, repo.name)
}

func gitPushWithRetry(ctx context.Context, name string) error {
	var err error
	for i := 0; i < 3; i++ {
		err = run.Bash(ctx, "git push destination").Dir(name).Run().Wait()
		if err != nil {
			if strings.Contains("timed out", err.Error()) {
				continue
			}
			return err
		}
	}
	return nil
}

func inTempFolder(f func() error) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()

	path, err := os.MkdirTemp(os.TempDir(), "repo")
	if err != nil {
		return err
	}
	err = os.Chdir(path)
	if err != nil {
		return err
	}
	return f()
}
