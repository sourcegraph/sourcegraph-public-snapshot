package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sync/atomic"
	"time"

	"github.com/google/go-github/github"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/dev/scaletesting/internal/store"
	"github.com/sourcegraph/sourcegraph/lib/group"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var app = &cli.App{
	Usage:       "Edit repository settings in bulk",
	Description: "https://handbook.sourcegraph.com/departments/engineering/dev/tools/scaletesting/",
	Compiled:    time.Now(),
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "github.token",
			Usage:    "GitHub token",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "github.org",
			Usage:    "Organization holding the repositories that are to be edited",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "github.url",
			Usage: "Base URL to the GitHub instance",
			Value: "https://github.com",
		},
		&cli.StringFlag{
			Name:  "state",
			Usage: "Path to a database file to store the state (will be created if doesn't exist)",
			Value: "bulkreposettings.db",
		},
		&cli.IntFlag{
			Name:  "retry",
			Usage: "Max retry count",
			Value: 3,
		},
	},
	Commands: []*cli.Command{
		{
			Name:        "visibility",
			Description: "change visibility of repositories",
			Subcommands: []*cli.Command{
				{
					Name:        "private",
					Description: "Set repo visibility to private",
					Action: func(cmd *cli.Context) error {
						logger := log.Scoped("runner", "")
						ctx := context.Background()
						tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
							&oauth2.Token{AccessToken: cmd.String("github.token")},
						))
						baseURL, err := url.Parse(cmd.String("github.url"))
						if err != nil {
							return err
						}
						baseURL.Path = "/api/v3"
						gh, err := github.NewEnterpriseClient(baseURL.String(), baseURL.String(), tc)
						if err != nil {
							logger.Fatal("failed to sign-in to GitHub", log.Error(err))
						}

						org := cmd.String("github.org")

						s, err := store.New(cmd.String("state"))
						if err != nil {
							logger.Fatal("failed to init state", log.Error(err))
						}

						repos, err := s.Load()
						if err != nil {
							logger.Error("failed to open state database", log.Error(err))
							return err
						}

						if len(repos) == 0 {
							logger.Info("No existing state found, creating ...")
							repos, err = fetchRepos(cmd.Context, org, gh)
							if err != nil {
								logger.Error("failed to fetch repositories from org", log.Error(err), log.String("github.org", org))
								return err
							}
							if err := s.Insert(repos); err != nil {
								logger.Error("failed to insert repositories from org", log.Error(err), log.String("github.org", org))
								return err
							}
						}

						out := output.NewOutput(os.Stdout, output.OutputOpts{})
						bars := []output.ProgressBar{
							{Label: "Updating repos", Max: float64(len(repos))},
						}
						progress := out.Progress(bars, nil)
						defer progress.Destroy()

						var done int64
						total := len(repos)

						g := group.NewWithResults[error]().WithMaxConcurrency(20)
						for _, r := range repos {
							r := r
							g.Go(func() error {
								if r.Pushed {
									return nil
								}
								var err error
								settings := &github.Repository{Private: github.Bool(true)}
								for i := 0; i < cmd.Int("retry"); i++ {
									_, _, err = gh.Repositories.Edit(cmd.Context, org, r.Name, settings)
									if err != nil {
										r.Failed = err.Error()
									} else {
										r.Failed = ""
										r.Pushed = true
										break
									}
								}
								if err := s.SaveRepo(r); err != nil {
									logger.Fatal("could not save repo", log.Error(err), log.String("repo", r.Name))
								}
								atomic.AddInt64(&done, 1)
								progress.SetValue(0, float64(done))
								progress.SetLabel(0, fmt.Sprintf("Updating repos (%d/%d)", done, total))
								return err
							})
						}
						return nil
					},
				},
			},
		},
	},
}

func fetchRepos(ctx context.Context, org string, gh *github.Client) ([]*store.Repo, error) {
	opts := github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{},
	}
	var repos []*github.Repository
	for {
		rs, resp, err := gh.Repositories.ListByOrg(ctx, org, &opts)
		if err != nil {
			return nil, err
		}
		repos = append(repos, rs...)

		if resp.NextPage == 0 {
			break
		}
		opts.ListOptions.Page = resp.NextPage
	}

	res := make([]*store.Repo, 0, len(repos))
	for _, repo := range repos {
		res = append(res, &store.Repo{
			Name:   repo.GetName(),
			GitURL: repo.GetGitURL(),
		})
	}

	return res, nil
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
