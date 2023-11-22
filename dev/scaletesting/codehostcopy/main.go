package main

import (
	"context"
	_ "embed"
	"fmt"
	"net/url"
	"os"
	"time"

	cueErrs "cuelang.org/go/cue/errors"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/dev/scaletesting/internal/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

//go:embed config.example.cue
var exampleConfig string

// SSHKeyHandler enables one to add and remove SSH keys
type SSHKeyHandler interface {
	AddSSHKey(ctx context.Context) (int64, error)
	DropSSHKey(ctx context.Context, keyID int64) error
}

type CodeHostSource interface {
	GitOpts() []GitOpt
	SSHKeyHandler
	InitializeFromState(ctx context.Context, stateRepos []*store.Repo) (int, int, error)
	Iterator() Iterator[[]*store.Repo]
}

type CodeHostDestination interface {
	GitOpts() []GitOpt
	SSHKeyHandler
	CreateRepo(ctx context.Context, name string) (*url.URL, error)
}

type Iterator[T any] interface {
	Err() error
	Next(ctx context.Context) T
	Done() bool
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
		&cli.PathFlag{
			Name:     "ssh-key",
			Usage:    "path to ssh key to use for cloning",
			Value:    "",
			Required: false,
		},
	},
	Action: func(cmd *cli.Context) error {
		return doRun(cmd.Context, log.Scoped("runner"), cmd.String("state"), cmd.String("config"))
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
				return doList(cmd.Context, log.Scoped("list"), cmd.String("state"), cmd.String("config"), cmd.Int("limit"))
			},
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:        "limit",
					DefaultText: "limit the amount of repos that gets printed. Use 0 to print all repos",
					Value:       10,
				},
			},
		},
	},
}

func createDestinationCodeHost(ctx context.Context, logger log.Logger, cfg CodeHostDefinition) (CodeHostDestination, error) {
	switch cfg.Kind {
	case "dummy":
		return NewDummyCodeHost(logger, &cfg), nil
	case "bitbucket":
		return NewBitbucketCodeHost(logger, &cfg)
	case "gitlab":
		return NewGitLabCodeHost(ctx, &cfg)
	case "github":
		return NewGitHubCodeHost(ctx, &cfg)
	default:
		return nil, errors.Newf("unknown code host %q", cfg.Kind)
	}
}

func createSourceCodeHost(ctx context.Context, logger log.Logger, cfg CodeHostDefinition) (CodeHostSource, error) {
	switch cfg.Kind {
	case "bitbucket":
		return NewBitbucketCodeHost(logger, &cfg)
	case "github":
		return NewGitHubCodeHost(ctx, &cfg)
	case "gitlab":
		return NewGitLabCodeHost(ctx, &cfg)
	default:
		return nil, errors.Newf("unknown code host %q", cfg.Kind)
	}
}

func doRun(ctx context.Context, logger log.Logger, state string, config string) error {
	cfg, err := loadConfig(config)
	if err != nil {
		var cueErr cueErrs.Error
		if errors.As(err, &cueErr) {
			logger.Info(cueErrs.Details(err, nil))
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
	return runner.Copy(ctx, cfg.MaxConcurrency)
}

func doList(ctx context.Context, logger log.Logger, state string, config string, limit int) error {
	cfg, err := loadConfig(config)
	if err != nil {
		var cueErr cueErrs.Error
		if errors.As(err, &cueErr) {
			logger.Info(cueErrs.Details(err, nil))
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

func main() {
	cb := log.Init(log.Resource{
		Name: "codehostcopy",
	})
	defer cb.Sync()
	logger := log.Scoped("main")

	if err := app.RunContext(context.Background(), os.Args); err != nil {
		logger.Fatal("failed to run", log.Error(err))
	}
}
