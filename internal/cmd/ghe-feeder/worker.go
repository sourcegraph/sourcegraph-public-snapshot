package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v31/github"
	"github.com/inconshreveable/log15"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/oauth2"
)

func newGHEClient(ctx context.Context, baseURL, uploadURL, token string) (*github.Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewEnterpriseClient(baseURL, uploadURL, tc)
}

func createOrg() (string, int) {
	size := rand.Intn(500)
	if size < 5 {
		size = 5
	}
	name := fmt.Sprintf("%s-%d", getRandomName(0), size)
	return name, size
}

type worker struct {
	name            string
	client          *github.Client
	sem             chan struct{}
	index           int
	scratchDir      string
	work            <-chan string
	wg              *sync.WaitGroup
	bar             *progressbar.ProgressBar
	reposPerOrg     int
	numFailed       int64
	numSucceeded    int64
	fdr             *feederDB
	currentOrg      string
	currentNumRepos int
	currentMaxRepos int
	logger          log15.Logger
}

func (wkr *worker) run(ctx context.Context) {
	defer wkr.wg.Done()

	wkr.currentOrg, wkr.currentMaxRepos = createOrg()

	wkr.logger.Debug("switching to org", "org", wkr.currentOrg)

	for line := range wkr.work {
		if ctx.Err() != nil {
			return
		}
		err := wkr.process(ctx, line)
		if err != nil {
			wkr.logger.Error("failed to process repo", "ownerRepo", line, "error", err)
			wkr.numFailed++
			_ = wkr.fdr.failed(line)
		} else {
			wkr.numSucceeded++
			wkr.currentNumRepos++
			if wkr.currentNumRepos >= wkr.currentMaxRepos {
				wkr.currentOrg, wkr.currentMaxRepos = createOrg()
				wkr.currentNumRepos = 0
				wkr.logger.Debug("switching to org", "org", wkr.currentOrg)
			}
		}
		_ = wkr.bar.Add(1)
	}
}

func (wkr *worker) process(ctx context.Context, work string) error {
	xs := strings.Split(work, "/")
	if len(xs) != 2 {
		return fmt.Errorf("expected owner/repo line, got %s instead", work)
	}
	owner, repo := xs[0], xs[1]

	err := wkr.cloneRepo(ctx, owner, repo)
	if err != nil {
		return err
	}

	return nil
}

func (wkr *worker) cloneRepo(ctx context.Context, owner, repo string) error {
	ownerDir := filepath.Join(wkr.scratchDir, owner)
	err := os.MkdirAll(ownerDir, 0777)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "clone",
		fmt.Sprintf("https://github.com/%s/%s", owner, repo))
	cmd.Dir = ownerDir
	cmd.Env = append(cmd.Env, "GIT_ASKPASS=/bin/echo")

	return cmd.Run()
}
