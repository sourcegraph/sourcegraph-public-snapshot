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

// GitOpt is an option which changes the git command that gets invoked
type GitOpt func(cmd *run.Command) *run.Command

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

func (r *Runner) addSSHKey(ctx context.Context) (func(), error) {
	// Add SSH Key to source and dest
	srcKey, err := r.source.AddSSHKey(ctx)
	if err != nil {
		return nil, err
	}

	destKey, err := r.destination.AddSSHKey(ctx)
	if err != nil {
		// Have to remove the source since it was added earlier
		r.source.DropSSHKey(ctx, srcKey)
		return nil, err
	}

	// create a func that cleans the ssh keys up when called
	return func() {
		r.source.DropSSHKey(ctx, srcKey)
		r.destination.DropSSHKey(ctx, destKey)
	}, nil
}

func (r *Runner) Run(ctx context.Context, concurrency int) error {
	out := output.NewOutput(os.Stdout, output.OutputOpts{})

	out.WriteLine(output.Line(output.EmojiInfo, output.StyleGrey, "Adding codehost ssh key"))
	cleanup, err := r.addSSHKey(ctx)
	if err != nil {
		return err
	}
	defer func() {
		out.WriteLine(output.Line(output.EmojiInfo, output.StyleGrey, "Removing codehost ssh key"))
		cleanup()
	}()

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

	g := group.NewWithResults[error]().WithMaxConcurrency(concurrency)
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
				err := pushRepo(ctx, repo, r.source.GitOpts(), r.destination.GitOpts())
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

func pushRepo(ctx context.Context, repo *store.Repo, srcOpts []GitOpt, destOpts []GitOpt) error {
	tmpDir, err := os.MkdirTemp(os.TempDir(), fmt.Sprintf("repo__%s", repo.Name))
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	cmd := run.Bash(ctx, "git clone", repo.GitURL).Dir(tmpDir)
	for _, opt := range srcOpts {
		cmd = opt(cmd)
	}
	err = cmd.Run().Wait()
	if err != nil {
		return err
	}
	repoDir := filepath.Join(tmpDir, repo.Name)
	cmd = run.Bash(ctx, "git remote add destination", repo.ToGitURL).Dir(repoDir)
	for _, opt := range destOpts {
		cmd = opt(cmd)
	}
	err = cmd.Run().Wait()
	if err != nil {
		return err
	}
	return gitPushWithRetry(ctx, repoDir, 3, destOpts...)
}

func gitPushWithRetry(ctx context.Context, dir string, retry int, destOpts ...GitOpt) error {
	var err error
	for i := 0; i < retry; i++ {
		cmd := run.Bash(ctx, "git push destination").Dir(dir)
		for _, opt := range destOpts {
			cmd = opt(cmd)
		}
		err = cmd.Run().Wait()
		if err != nil {
			if strings.Contains(err.Error(), "timed out") {
				continue
			}
			return err
		}
	}
	return nil
}
