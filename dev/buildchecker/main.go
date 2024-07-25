package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/google/go-github/v55/github"
	"github.com/slack-go/slack"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/dev/team"
)

// Flags denotes shared Buildchecker flags.
type Flags struct {
	BuildkiteToken      string
	Pipeline            string
	Branch              string
	FailuresThreshold   int
	FailuresTimeoutMins int
}

func (f *Flags) Parse() {
	flag.StringVar(&f.BuildkiteToken, "buildkite.token", "", "mandatory buildkite token")
	flag.StringVar(&f.Pipeline, "pipeline", "sourcegraph", "name of the pipeline to inspect")
	flag.StringVar(&f.Branch, "branch", "main", "name of the branch to inspect")

	flag.IntVar(&f.FailuresThreshold, "failures.threshold", 3, "failures required to trigger an incident")
	flag.IntVar(&f.FailuresTimeoutMins, "failures.timeout", 60, "duration of a run required to be considered a failure (minutes)")
	flag.Parse()
}

func main() {
	ctx := context.Background()

	// Define and parse all flags
	flags := &Flags{}

	checkFlags := &cmdCheckFlags{}
	flag.StringVar(&checkFlags.githubToken, "github.token", "", "mandatory github token")
	flag.StringVar(&checkFlags.slackAnnounceWebhooks, "slack.announce-webhook", "", "Slack Webhook URL to post the results on (comma-delimited for multiple values)")
	flag.StringVar(&checkFlags.slackToken, "slack.token", "", "Slack token used for resolving Slack handles to mention")
	flag.StringVar(&checkFlags.slackDebugWebhook, "slack.debug-webhook", "", "Slack Webhook URL to post debug results on")
	flag.StringVar(&checkFlags.slackDiscussionChannel, "slack.discussion-channel", "#buildkite-main", "Slack channel to ask everyone to head over to for discusison")

	flags.Parse()

	switch cmd := flag.Arg(0); cmd {
	case "check":
		log.Println("buildchecker check")
		cmdCheck(ctx, flags, checkFlags)

	default:
		log.Printf("unknown command %q - available commands: 'history', 'check'", cmd)
		os.Exit(1)
	}
}

type cmdCheckFlags struct {
	githubToken string

	slackToken             string
	slackAnnounceWebhooks  string
	slackDebugWebhook      string
	slackDiscussionChannel string
}

func cmdCheck(ctx context.Context, flags *Flags, checkFlags *cmdCheckFlags) {
	config, err := buildkite.NewTokenConfig(flags.BuildkiteToken, false)
	if err != nil {
		log.Fatal("buildkite.NewTokenConfig: ", err)
	}
	// Buildkite client
	bkc := buildkite.NewClient(config.Client())

	// GitHub client
	ghc := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: checkFlags.githubToken},
	)))

	// Newest is returned first https://buildkite.com/docs/apis/rest-api/builds#list-builds-for-a-pipeline
	builds, _, err := bkc.Builds.ListByPipeline("sourcegraph", flags.Pipeline, &buildkite.BuildsListOptions{
		Branch: []string{flags.Branch},
		// Fix to high page size just in case, default is 30
		// https://buildkite.com/docs/apis/rest-api#pagination
		ListOptions: buildkite.ListOptions{PerPage: 99},
	})
	if err != nil {
		log.Fatal("Builds.ListByPipeline: ", err)
	}

	opts := CheckOptions{
		FailuresThreshold: flags.FailuresThreshold,
		BuildTimeout:      time.Duration(flags.FailuresTimeoutMins) * time.Minute,
	}
	log.Printf("running buildchecker over %d builds with option: %+v\n", len(builds), opts)
	results, err := CheckBuilds(
		ctx,
		NewBranchLocker(ghc, "sourcegraph", "sourcegraph", flags.Branch),
		team.NewTeammateResolver(ghc, slack.New(checkFlags.slackToken)),
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
		summary := generateBranchEventSummary(results.LockBranch, flags.Branch, checkFlags.slackDiscussionChannel, results.FailedCommits)
		announceWebhooks := strings.Split(checkFlags.slackAnnounceWebhooks, ",")

		// Post update first to avoid invisible changes
		if oneSucceeded, err := postSlackUpdate(announceWebhooks, summary); !oneSucceeded {
			// If action is an unlock, try to unlock anyway
			if !results.LockBranch {
				log.Println("slack update failed but action is an unlock, trying to unlock branch anyway")
				goto POST
			}
			log.Fatal("postSlackUpdate: ", err)
		} else if err != nil {
			// At least one message succeeded, so we just log the error and continue
			log.Println("postSlackUpdate: ", err)
		}

	POST:
		// If post works, do the thing
		if err := results.Action(); err != nil {
			_, slackErr := postSlackUpdate([]string{checkFlags.slackDebugWebhook}, fmt.Sprintf("Failed to execute action (%+v): %s", results, err))
			if slackErr != nil {
				log.Fatal("postSlackUpdate: ", err)
			}

			log.Fatal("results.Action: ", err)
		}
	}
}
