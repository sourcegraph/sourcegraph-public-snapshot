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
	flag.StringVar(&slackWebhook, "slack.webhook", "", "Slack Webhook URL to post the results on")
	flag.StringVar(&pipeline, "pipeline", "sourcegraph", "name of the pipeline to inspect")
	flag.StringVar(&branch, "branch", "main", "name of the branch to inspect")
	flag.IntVar(&threshold, "failures.threshold", 3, "failures required to trigger an incident")
	flag.IntVar(&timeoutMins, "failures.timeout", 40, "duration of a run required to be considered a failure (minutes)")
	flag.Parse()

	config, err := buildkite.NewTokenConfig(buildkiteToken, false)
	if err != nil {
		log.Fatal("buildkite.NewTokenConfig: ", err)
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
		// Fix to high page size just in case, default is 30
		// https://buildkite.com/docs/apis/rest-api#pagination
		ListOptions: buildkite.ListOptions{PerPage: 99},
	})
	if err != nil {
		log.Fatal("Builds.ListByPipeline: ", err)
	}

	opts := CheckOptions{
		FailuresThreshold: threshold,
		BuildTimeout:      time.Duration(timeoutMins) * time.Minute,
	}
	log.Printf("running buildchecker over %d builds with option: %+v\n", len(builds), opts)
	results, err := CheckBuilds(
		ctx,
		NewBranchLocker(ghc, "sourcegraph", "sourcegraph", branch),
		builds,
		opts,
	)
	if err != nil {
		log.Fatal("CheckBuilds: ", err)
	}
	log.Printf("results: %+v\n", err)

	// Only post an update if the lock has been modified
	lockModified := results.Action != nil
	if lockModified {
		// Post update first to avoid invisible changes
		summary := slackSummary(results.LockBranch, results.FailedCommits)
		if err := postSlackUpdate(slackWebhook, summary); err != nil {
			// If action is an unlock, try to unlock anyway
			if !results.LockBranch {
				log.Println("slack update failed but action is an unlock, trying to unlock branch anyway")
				goto POST
			}
			log.Fatal("postSlackUpdate: ", err)
		}

	POST:
		// If post works, do the thing
		if err := results.Action(); err != nil {
			slackErr := postSlackUpdate(slackWebhook, fmt.Sprintf("Failed to execute action (%+v): %s", results, err))
			if slackErr != nil {
				log.Fatal("postSlackUpdate: ", err)
			}
			log.Fatal("results.Action: ", err)
		}
	}
}
