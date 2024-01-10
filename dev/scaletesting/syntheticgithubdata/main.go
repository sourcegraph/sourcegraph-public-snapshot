package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/google/go-github/v55/github"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/lib/output"
)

type config struct {
	githubToken    string
	githubURL      string
	githubUser     string
	githubPassword string

	userCount      int
	teamCount      int
	subOrgCount    int
	reposSourceOrg string
	orgAdmin       string
	action         string
	resume         string
	generateTokens bool
}

var (
	emailDomain = "scaletesting.sourcegraph.com"

	out      *output.Output
	store    *state
	gh       *github.Client
	progress output.Progress
)

type userToken struct {
	login string
	token string
}

func main() {
	var cfg config

	flag.StringVar(&cfg.githubToken, "github.token", "", "(required) GitHub personal access token for the destination GHE instance")
	flag.StringVar(&cfg.githubURL, "github.url", "", "(required) GitHub base URL for the destination GHE instance")
	flag.StringVar(&cfg.githubUser, "github.login", "", "(required) GitHub user to authenticate with")
	flag.StringVar(&cfg.githubPassword, "github.password", "", "(required) password of the GitHub user to authenticate with")
	flag.IntVar(&cfg.userCount, "user.count", 100, "Amount of users to create or delete")
	flag.IntVar(&cfg.teamCount, "team.count", 20, "Amount of teams to create or delete")
	flag.IntVar(&cfg.subOrgCount, "suborg.count", 1, "Amount of sub-orgs to create or delete")
	flag.StringVar(&cfg.orgAdmin, "org.admin", "", "(required) Login of admin of orgs")
	flag.StringVar(&cfg.reposSourceOrg, "repos.sourceOrgName", "blank200k", "The org that contains the imported repositories to transfer")

	flag.StringVar(&cfg.action, "action", "create", "Whether to 'create', 'delete', or 'validate' the synthetic data")
	flag.StringVar(&cfg.resume, "resume", "state.db", "Temporary state to use to resume progress if interrupted")
	flag.BoolVar(&cfg.generateTokens, "generateTokens", false, "Generate new impersonation OAuth tokens for users")

	flag.Parse()

	ctx := context.Background()
	out = output.NewOutput(os.Stdout, output.OutputOpts{})

	var err error
	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.githubToken},
	))
	gh, err = github.NewClient(tc).WithEnterpriseURLs(cfg.githubURL, cfg.githubURL)
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
	if cfg.orgAdmin == "" {
		writeFailure(out, "-org.admin must be provided")
		flag.Usage()
		os.Exit(-1)
	}

	store, err = newState(cfg.resume)
	if err != nil {
		log.Fatal(err)
	}

	// load or generate orgs (used by both create and delete actions)
	var orgs []*org
	if orgs, err = store.loadOrgs(); err != nil {
		log.Fatal(err)
	}

	if len(orgs) == 0 {
		if orgs, err = store.generateOrgs(cfg); err != nil {
			log.Fatal(err)
		}
		writeSuccess(out, "generated org jobs in %s", cfg.resume)
	} else {
		writeSuccess(out, "resuming org jobs from %s", cfg.resume)
	}

	start := time.Now()

	switch cfg.action {
	case "create":
		create(ctx, orgs, cfg)

	case "delete":
		delete(ctx, cfg)

	case "validate":
		validate(ctx)
	}

	end := time.Now()
	writeInfo(out, "Started at %s, finished at %s", start.String(), end.String())
}

func writeSuccess(out *output.Output, format string, a ...any) {
	out.WriteLine(output.Linef("✅", output.StyleSuccess, format, a...))
}

func writeInfo(out *output.Output, format string, a ...any) {
	out.WriteLine(output.Linef("ℹ️", output.StyleYellow, format, a...))
}

func writeFailure(out *output.Output, format string, a ...any) {
	out.WriteLine(output.Linef("❌", output.StyleFailure, format, a...))
}
