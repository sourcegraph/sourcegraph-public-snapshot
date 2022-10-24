package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"cuelang.org/go/cue/errors"

	"github.com/sourcegraph/log"

	"github.com/urfave/cli/v2"
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

var app = &cli.App{
	Usage:       "Copy organizations across code hosts",
	Description: "https://handbook.sourcegraph.com/departments/engineering/dev/tools/scaletesting/",
	Compiled:    time.Now(),
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "state",
			Usage: "Path to the file storing state, to resume work from",
			Value: "codehostcopy.db",
		},
		&cli.StringFlag{
			Name:  "config",
			Usage: "Path to the config file defining what to copy",
		},
	},
	Action: func(cmd *cli.Context) error {
		return doRun(cmd.Context, log.Scoped("runner", ""), cmd.String("state"), cmd.String("config"))
	},
	Commands: []*cli.Command{
		{
			Name:        "new",
			Description: "Create a new config file to start from",
			Action: func(ctx *cli.Context) error {
				return nil
			},
		},
	},
}

func bitbucketTest(ctx context.Context, logger log.Logger, cfg *Config) {
	logger.Info("creating bitbucket codehost")
	c, err := NewBitbucketCodeHost(ctx, logger, &cfg.From)
	if err != nil {
		return
	}

	logger.Info("listing repos")
	repos, err := c.ListRepos(ctx)
	if err != nil {
		logger.Error("list failed", log.Error(err))
	}

	for _, r := range repos {
		fmt.Printf("repo %+v", r)
	}
}

func doRun(ctx context.Context, logger log.Logger, state string, config string) error {
	cfg, err := loadConfig(config)
	if err != nil {
		var cueErr errors.Error
		if errors.As(err, &cueErr) {
			logger.Info(errors.Details(err, nil))
		}
		logger.Fatal("failed to load config", log.Error(err))
	}

	logger.Info("config loaded")

	bitbucketTest(ctx, logger, cfg)
	return nil
	s, err := newState(state)
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

	runner := NewRunner(logger, s, gh, gl)
	return runner.Run(ctx, 20)
}

func main() {
	cb := log.Init(log.Resource{
		Name: "codehostcopy",
	})
	defer cb.Sync()
	logger := log.Scoped("main", "")
	logger.Info("starting up")

	if err := app.RunContext(context.Background(), os.Args); err != nil {
		logger.Fatal("failed to run", log.Error(err))
	}
}
