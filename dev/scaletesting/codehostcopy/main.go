package main

import (
	"context"
	"flag"
	"net/url"

	"cuelang.org/go/cue/errors"

	"github.com/sourcegraph/log"
)

type CodeHostSource interface {
	ListRepos(ctx context.Context) ([]*Repo, error)
}

type CodeHostDestination interface {
	CreateRepo(ctx context.Context, name string) (*url.URL, error)
}

type Repo struct {
	FromGitURL string
	ToGitURL   string
	Name       string
	Failed     string
	Created    bool
	Pushed     bool
}

type Flags struct {
	resume string
	config string
}

func main() {
	var flags Flags
	flag.StringVar(&flags.resume, "resume", "state.db", "Temporary state to use to resume progress if interrupted")
	flag.StringVar(&flags.config, "config", "", "TODO")
	flag.Parse()

	cb := log.Init(log.Resource{
		Name: "codehostcopy",
	})
	defer cb.Sync()
	logger := log.Scoped("main", "")

	ctx := context.Background()
	cfg, err := loadConfig(flags.config)
	if err != nil {
		var cueErr errors.Error
		if errors.As(err, &cueErr) {
			logger.Info(errors.Details(err, nil))
		}
		logger.Fatal("failed to load config", log.Error(err))
	}

	state, err := newState(flags.resume)
	if err != nil {
		logger.Fatal("failed to init state", log.Error(err))
	}
	gh, err := NewGithubCodeHost(ctx, &cfg.From)
	if err != nil {
		logger.Fatal("failed to init GitHub code host", log.Error(err))
	}
	gl, err := NewGitLabCodeHost(ctx, &cfg.Destination)
	if err != nil {
		logger.Fatal("failed to init GitLab code host", log.Error(err))
	}

	runner := NewRunner(logger, state, gh, gl)
	if err := runner.Run(ctx, 20); err != nil {
		logger.Fatal("failed runnerr", log.Error(err))
	}
}
