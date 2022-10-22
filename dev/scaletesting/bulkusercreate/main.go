package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/google/go-github/v41/github"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/lib/output"
)

type config struct {
	githubToken    string
	githubURL      string
	githubUser     string
	githubPassword string

	count int
	//prefix string
	resume string
	retry  int
}

var emailDomain = "@scaletesting.sourcegraph.com"

func main() {
	var cfg config

	flag.StringVar(&cfg.githubToken, "github.token", "", "(required) GitHub personal access token for the destination GHE instance")
	flag.StringVar(&cfg.githubURL, "github.url", "", "(required) GitHub base URL for the destination GHE instance")
	flag.StringVar(&cfg.githubUser, "github.login", "", "(required) GitHub user to authenticate with")
	flag.StringVar(&cfg.githubPassword, "github.password", "", "(required) password of the GitHub user to authenticate with")
	flag.IntVar(&cfg.count, "count", 100, "Amount of users to create")
	flag.IntVar(&cfg.retry, "retry", 5, "Retries count")
	//flag.StringVar(&cfg.prefix, "prefix", "repo", "Prefix to use when naming the repo, ex '[prefix]000042'")
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

	state, err := newState(cfg.resume)
	if err != nil {
		log.Fatal(err)
	}

}

func writeSuccess(out *output.Output, format string, a ...any) {
	out.WriteLine(output.Linef("✅", output.StyleSuccess, format, a...))
}

func writeFailure(out *output.Output, format string, a ...any) {
	out.WriteLine(output.Linef("❌", output.StyleFailure, format, a...))
}
