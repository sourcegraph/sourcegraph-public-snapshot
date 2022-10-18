package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"

	"github.com/google/go-github/v41/github"
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

	r, err := newBlankRepo(cfg.githubUser, cfg.githubPassword)
	if err != nil {
		writeFailure(out, "Failed to create folder for repository")
		log.Fatal(err)
	}
	err = r.init(ctx)
	if err != nil {
		writeFailure(out, "Failed to initialize blank repository")
		log.Fatal(err)
	}

	names := generateNames(cfg.prefix, cfg.count)
	var mu sync.Mutex

	bars := []output.ProgressBar{
		{Label: "CreatingRepos", Max: float64(cfg.count)},
		{Label: "Adding remotes", Max: float64(cfg.count)},
		{Label: "Pushing branches", Max: float64(cfg.count)},
	}
	progress := out.Progress(bars, nil)

	g := group.New().WithMaxConcurrency(10)
	var done int64
	for name := range names {
		name := name
		g.Go(func() {
			newRepo, _, err := gh.Repositories.Create(ctx, cfg.githubOrg, &github.Repository{Name: github.String(name)})
			if err != nil {
				writeFailure(out, "Failed to create repository %s", name)
				log.Fatal(err)
			}
			mu.Lock()
			names[name] = newRepo.GetGitURL()
			mu.Unlock()
			atomic.AddInt64(&done, 1)
			progress.SetValue(0, float64(done))
		})
	}
	g.Wait()

	done = 0
	// Adding a remote will lock git configuration.
	for name, gitURL := range names {
		err = r.addRemote(ctx, name, gitURL)
		if err != nil {
			writeFailure(out, "Failed to add remote to repository %s", name)
			log.Fatal(err)
		}
		done++
		progress.SetValue(1, float64(done))
	}

	done = 0
	g = group.New().WithMaxConcurrency(10)
	for name := range names {
		name := name
		g.Go(func() {
			err = r.pushRemote(ctx, name)
			if err != nil {
				writeFailure(out, "Failed to push to repository %s", name)
				log.Fatal(err)
			}
			atomic.AddInt64(&done, 1)
			progress.SetValue(2, float64(done))
		})
	}
	g.Wait()

	progress.Destroy()
	writeSuccess(out, "Successfully added %d repositories on $GHE/%s", cfg.count, cfg.githubOrg)
	defer r.teardown()
}

func writeSuccess(out *output.Output, format string, a ...any) {
	out.WriteLine(output.Linef("✅", output.StyleSuccess, format, a...))
}

func writeFailure(out *output.Output, format string, a ...any) {
	out.WriteLine(output.Linef("❌", output.StyleFailure, format, a...))
}

func generateNames(prefix string, count int) map[string]string {
	names := make(map[string]string, count)
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s%09d", prefix, i)
		names[name] = ""
	}
	return names
}
