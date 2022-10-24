package main

import (
	"context"
	_ "embed"
	"fmt"
	"net/url"
	"os"
	"time"

	"cuelang.org/go/cue/errors"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/log"
)

//go:embed config.example.cue
var exampleConfig string

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
			Name:     "config",
			Usage:    "Path to the config file defining what to copy",
			Required: true,
		},
	},
	Action: func(cmd *cli.Context) error {
		return doRun(cmd.Context, log.Scoped("runner", ""), cmd.String("state"), cmd.String("config"))
	},
	Commands: []*cli.Command{
		{
			Name:        "example",
			Description: "Create a new config file to start from",
			Action: func(ctx *cli.Context) error {
				fmt.Printf("%s", exampleConfig)
				return nil
			},
		},
	},
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

	if err := app.RunContext(context.Background(), os.Args); err != nil {
		logger.Fatal("failed to run", log.Error(err))
	}

}
