package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
)

func main() {
	var (
		ctx            = context.Background()
		buildkiteToken string
		githubToken    string
		slackWebhook   string
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
	flag.StringVar(&slackWebhook, "slack", "", "Slack Webhook URL to post the results on")

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

	opts := sherrifOptions{
		FailuresThreshold: threshold,
		BuildTimeout:      time.Duration(timeoutMins) * time.Minute,
	}
	fmt.Printf("running buildsherrif over %d builds with option: %+v\n", len(builds), opts)
	results, err := buildsherrif(
		ctx,
		newBranchLocker(ghc, "sourcegraph", "sourcegraph", branch),
		builds,
		opts,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Only post an update if the lock has been modified
	if results.LockModified {
		if err := postSlackUpdate(slackWebhook, slackSummary(results.Locked, results.FailedCommits)); err != nil {
			log.Fatal(err)
		}
	}
}
