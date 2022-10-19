package main

import (
	"context"
	"flag"
	"log"
	"os"
	"sync/atomic"

	"github.com/google/go-github/v41/github"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sourcegraph/sourcegraph/lib/group"
	"github.com/sourcegraph/sourcegraph/lib/output"
	"golang.org/x/oauth2"
)

type config struct {
	githubToken    string
	githubOrg      string
	githubURL      string
	githubUser     string
	githubPassword string

	count  int
	prefix string
	resume string
}

func main() {
	var cfg config

	flag.StringVar(&cfg.githubToken, "github.token", "", "(required) GitHub personal access token for the destination GHE instance")
	flag.StringVar(&cfg.githubURL, "github.url", "", "(required) GitHub base URL for the destination GHE instance")
	flag.StringVar(&cfg.githubOrg, "github.org", "", "(required) GitHub organization for the destination GHE instance to add the repos")
	flag.StringVar(&cfg.githubUser, "github.login", "", "(required) GitHub organization for the destination GHE instance to add the repos")
	flag.StringVar(&cfg.githubPassword, "github.password", "", "(required) GitHub organization for the destination GHE instance to add the repos")
	flag.IntVar(&cfg.count, "count", 100, "Amount of blank repos to create")
	flag.StringVar(&cfg.prefix, "prefix", "repo", "Prefix to use when naming the repo, ex '[prefix]000042'")
	flag.StringVar(&cfg.resume, "resume", "state.db", "Temporary state to use to resume progress if interrupted")

	flag.Parse()

	out := output.NewOutput(os.Stdout, output.OutputOpts{})

	ctx := context.Background()
	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.githubToken},
	))

	gh, err := github.NewEnterpriseClient(cfg.githubURL, cfg.githubURL, tc)
	if err != nil {
		writeFailure(out, "Failed to sign-in to GHE")
		log.Fatal(err)
	}

	if cfg.githubOrg == "" {
		writeFailure(out, "-github.org must be provided")
		flag.Usage()
		os.Exit(-1)
	}
	if cfg.githubURL == "" {
		writeFailure(out, "-github.URL must be provided")
		flag.Usage()
		os.Exit(-1)
	}
	if cfg.githubToken == "" {
		writeFailure(out, "-github.token must be provided")
		flag.Usage()
		os.Exit(-1)
	}
	if cfg.githubUser == "" {
		writeFailure(out, "-github.login must be provided")
		flag.Usage()
		os.Exit(-1)
	}
	if cfg.githubPassword == "" {
		writeFailure(out, "-github.password must be provided")
		flag.Usage()
		os.Exit(-1)
	}

	blank, err := newBlankRepo(cfg.githubUser, cfg.githubPassword)
	if err != nil {
		writeFailure(out, "Failed to create folder for repository")
		log.Fatal(err)
	}
	err = blank.init(ctx)
	if err != nil {
		writeFailure(out, "Failed to initialize blank repository")
		log.Fatal(err)
	}

	state, err := newState(cfg.resume)
	if err != nil {
		log.Fatal(err)
	}
	var repos []*repo
	if repos, err = state.load(); err != nil {
		log.Fatal(err)
	}

	if len(repos) == 0 {
		repos, err = state.generate(cfg)
		if err != nil {
			log.Fatal(err)
		}
		writeSuccess(out, "generated jobs in %s", cfg.resume)
	} else {
		writeSuccess(out, "resuming jobs from %s", cfg.resume)
	}

	if _, _, err := gh.Organizations.Get(ctx, cfg.githubOrg); err != nil {
		writeFailure(out, "organization does not exists")
		log.Fatal(err)
	}

	bars := []output.ProgressBar{
		{Label: "CreatingRepos", Max: float64(cfg.count)},
		{Label: "Adding remotes", Max: float64(cfg.count)},
		{Label: "Pushing branches", Max: float64(cfg.count)},
	}
	progress := out.Progress(bars, nil)

	g := group.New().WithMaxConcurrency(20)
	var done int64
	for _, repo := range repos {
		repo := repo
		if repo.Created {
			atomic.AddInt64(&done, 1)
			progress.SetValue(0, float64(done))
			continue
		}
		g.Go(func() {
			newRepo, _, err := gh.Repositories.Create(ctx, cfg.githubOrg, &github.Repository{Name: github.String(repo.Name)})
			if err != nil {
				writeFailure(out, "Failed to create repository %s", repo.Name)
				repo.Failed = err.Error()
				if err := state.saveRepo(repo); err != nil {
					log.Fatal(err)
				}
				return
			}
			repo.GitURL = newRepo.GetGitURL()
			repo.Created = true
			repo.Failed = ""
			if err = state.saveRepo(repo); err != nil {
				log.Fatal(err)
			}

			atomic.AddInt64(&done, 1)
			progress.SetValue(0, float64(done))
		})
	}
	g.Wait()

	done = 0
	// Adding a remote will lock git configuration.
	for _, repo := range repos {
		err = blank.addRemote(ctx, repo.Name, repo.GitURL)
		if err != nil {
			writeFailure(out, "Failed to add remote to repository %s", repo.Name)
			log.Fatal(err)
		}
		done++
		progress.SetValue(1, float64(done))
	}

	done = 0
	g = group.New().WithMaxConcurrency(20)
	for _, repo := range repos {
		repo := repo
		if !repo.Created {
			continue
		}
		if repo.Pushed {
			atomic.AddInt64(&done, 1)
			progress.SetValue(2, float64(done))
			continue
		}
		g.Go(func() {
			err := blank.pushRemote(ctx, repo.Name)
			if err != nil {
				writeFailure(out, "Failed to push to repository %s", repo.Name)
				repo.Failed = err.Error()
				if err := state.saveRepo(repo); err != nil {
					log.Fatal(err)
				}
				return
			}
			repo.Pushed = true
			repo.Failed = ""
			if err := state.saveRepo(repo); err != nil {
				log.Fatal(err)
			}
			atomic.AddInt64(&done, 1)
			progress.SetValue(2, float64(done))
		})
	}
	g.Wait()

	progress.Destroy()
	all, err := state.countAllRepos()
	if err != nil {
		log.Fatal(err)
	}
	completed, err := state.countCompletedRepos()
	if err != nil {
		log.Fatal(err)
	}

	writeSuccess(out, "Successfully added %d repositories on $GHE/%s (%d failures)", completed, cfg.githubOrg, all-completed)
	defer blank.teardown()
}

func writeSuccess(out *output.Output, format string, a ...any) {
	out.WriteLine(output.Linef("✅", output.StyleSuccess, format, a...))
}

func writeFailure(out *output.Output, format string, a ...any) {
	out.WriteLine(output.Linef("❌", output.StyleFailure, format, a...))
}
