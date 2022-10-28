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

	"github.com/sourcegraph/sourcegraph/dev/scaletesting/internal/store"
	sgerr "github.com/sourcegraph/sourcegraph/lib/errors"
)

const Concurrency = 20

//go:embed config.example.cue
var exampleConfig string

type CodeHostSource interface {
	ListRepos(ctx context.Context) ([]*store.Repo, error)
}

type CodeHostDestination interface {
	CreateRepo(ctx context.Context, name string) (*url.URL, error)
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
			Action: func(_ *cli.Context) error {
				fmt.Printf("%s", exampleConfig)
				return nil
			},
		},
		{
			Name:        "list",
			Description: "list repos from the 'from' codehost defined in the configuration",
			Action: func(cmd *cli.Context) error {
				return doList(cmd.Context, log.Scoped("list", ""), cmd.String("state"), cmd.String("config"), cmd.Int("limit"))
			},
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:        "limit",
					DefaultText: "limit the amount of repos that gets printed. Use 0 to print all repos",
					Value:       10,
				},
			},
		},
		{
			Name:        "retry",
			Description: "retry repos that failed",
			Action: func(cmd *cli.Context) error {
				return doRetry(cmd.Context, log.Scoped("retry", ""), cmd.String("state"), cmd.String("config"))
			},
		},
	},
}

func createDestinationCodeHost(ctx context.Context, logger log.Logger, cfg CodeHostDefinition) (CodeHostDestination, error) {
	switch cfg.Kind {
	case "bitbucket":
		{
			return NewBitbucketCodeHost(ctx, logger, &cfg)
		}
	case "gitlab":
		{
			return NewGitLabCodeHost(ctx, &cfg)
		}
	default:
		{
			return nil, sgerr.Newf("unknown code host %q", cfg.Kind)
		}
	}
}

func createSourceCodeHost(ctx context.Context, logger log.Logger, cfg CodeHostDefinition) (CodeHostSource, error) {
	switch cfg.Kind {
	case "bitbucket":
		{
			return NewBitbucketCodeHost(ctx, logger, &cfg)
		}
	case "github":
		{
			return NewGithubCodeHost(ctx, &cfg)
		}
	default:
		{
			return nil, sgerr.Newf("unknown code host %q", cfg.Kind)
		}
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

	s, err := store.New(state)
	if err != nil {
		logger.Fatal("failed to init state", log.Error(err))
	}
	from, err := createSourceCodeHost(ctx, logger, cfg.From)
	if err != nil {
		logger.Fatal("failed to create from code host", log.Error(err))
	}

	dest, err := createDestinationCodeHost(ctx, logger, cfg.Destination)
	if err != nil {
		logger.Fatal("failed to create destination code host", log.Error(err))
	}
	runner := NewRunner(logger, s, from, dest)
	return runner.Copy(ctx, Concurrency)
}

func doList(ctx context.Context, logger log.Logger, state string, config string, limit int) error {
	cfg, err := loadConfig(config)
	if err != nil {
		var cueErr errors.Error
		if errors.As(err, &cueErr) {
			logger.Info(errors.Details(err, nil))
		}
		logger.Fatal("failed to load config", log.Error(err))
	}
	s, err := store.New(state)
	if err != nil {
		logger.Fatal("failed to init state", log.Error(err))
	}

	from, err := createSourceCodeHost(ctx, logger, cfg.From)
	if err != nil {
		logger.Fatal("failed to create from code host", log.Error(err))
	}

	runner := NewRunner(logger, s, from, nil)
	return runner.List(ctx, limit)

}

func doRetry(ctx context.Context, logger log.Logger, state string, config string) error {
	cfg, err := loadConfig(config)
	if err != nil {
		var cueErr errors.Error
		if errors.As(err, &cueErr) {
			logger.Info(errors.Details(err, nil))
		}
		logger.Fatal("failed to load config", log.Error(err))
	}
	s, err := store.New(state)
	if err != nil {
		logger.Fatal("failed to init state", log.Error(err))
	}

	from, err := createSourceCodeHost(ctx, logger, cfg.From)
	if err != nil {
		logger.Fatal("failed to create from code host", log.Error(err))
	}

	runner := NewRunner(logger, s, from, nil)
	return runner.Retry(ctx, Concurrency)

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
