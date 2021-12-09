package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/google/go-github/v31/github"
	"golang.org/x/oauth2"
)

func main() {
	var (
		ctx            = context.Background()
		buildkiteToken string
		githubToken    string
		pipeline       string
		branch         string
		threshold      int
		timeoutMins    int
	)

	flag.StringVar(&buildkiteToken, "buildkite.token", "", "mandatory buildkite token")
	flag.StringVar(&githubToken, "github.token", "", "mandatory github token")
	flag.StringVar(&pipeline, "pipeline", "sourcegraph", "name of the pipeline to inspect")
	flag.StringVar(&branch, "branch", "main", "name of the branch to inspect")
	flag.IntVar(&threshold, "failures.threshold", 3, "failures required to trigger an incident")
	flag.IntVar(&timeoutMins, "failures.timeout", 40, "duration of a run required to be considered a failure (minutes)")

	config, err := buildkite.NewTokenConfig(buildkiteToken, false)
	if err != nil {
		panic(err)
	}
	// Buildkite client
	bkc := buildkite.NewClient(config.Client())

	// GitHub client
	ghc := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)))

	// Newest is returned first https://buildkite.com/docs/apis/rest-api/builds#list-builds-for-a-pipeline
	builds, _, err := bkc.Builds.ListByPipeline("sourcegraph", pipeline, &buildkite.BuildsListOptions{
		Branch: branch,
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := buildsherrif(ctx, ghc, builds, sherrifOptions{
		Branch:            newBranchLocker(ghc, "sourcegraph", "sourcegraph", branch),
		FailuresThreshold: threshold,
		BuildTimeout:      time.Duration(timeoutMins) * time.Minute,
	}); err != nil {
		log.Fatal(err)
	}
}
