package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/google/go-github/v41/github"
	"github.com/slack-go/slack"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/dev/team"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
	flag.StringVar(&checkFlags.slackToken, "slack.token", "", "mandatory slack api token")
	flag.StringVar(&checkFlags.slackAnnounceWebhooks, "slack.announce-webhook", "", "Slack Webhook URL to post the results on (comma-delimited for multiple values)")
	flag.StringVar(&checkFlags.slackDebugWebhook, "slack.debug-webhook", "", "Slack Webhook URL to post debug results on")
	flag.StringVar(&checkFlags.slackDiscussionChannel, "slack.discussion-channel", "#buildkite-main", "Slack channel to ask everyone to head over to for discusison")

	historyFlags := &cmdHistoryFlags{}
	flag.StringVar(&historyFlags.createdFromDate, "created.from", "", "date in YYYY-MM-DD format")
	flag.StringVar(&historyFlags.createdToDate, "created.to", "", "date in YYYY-MM-DD format")
	flag.StringVar(&historyFlags.loadFrom, "load-from", "", "file to load builds from")
	flag.StringVar(&historyFlags.writeTo, "write-to", "builds.json", "file to write builds to (unused if loading from file)")

	flags.Parse()

	switch cmd := flag.Arg(0); cmd {
	case "history":
		log.Println("buildchecker history")
		cmdHistory(ctx, flags, historyFlags)

	case "check":
		log.Println("buildchecker check")
		cmdCheck(ctx, flags, checkFlags)

	default:
		log.Printf("unknown command %q - available commands: 'history', 'check'", cmd)
		os.Exit(1)
	}
}

type cmdCheckFlags struct {
	slackToken  string
	githubToken string

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

	// Slack client
	slc := slack.New(checkFlags.slackToken)

	// Newest is returned first https://buildkite.com/docs/apis/rest-api/builds#list-builds-for-a-pipeline
	builds, _, err := bkc.Builds.ListByPipeline("sourcegraph", flags.Pipeline, &buildkite.BuildsListOptions{
		Branch: flags.Branch,
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
		team.NewTeammateResolver(ghc, slc),
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
		summary := slackSummary(results.LockBranch, flags.Branch, checkFlags.slackDiscussionChannel, results.FailedCommits)
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

type cmdHistoryFlags struct {
	createdFromDate string
	createdToDate   string
	loadFrom        string
	writeTo         string
}

func cmdHistory(ctx context.Context, flags *Flags, historyFlags *cmdHistoryFlags) {

	// Time range
	var err error
	createdFrom := time.Now().Add(-24 * time.Hour)
	if historyFlags.createdFromDate != "" {
		createdFrom, err = time.Parse("2006-01-02", historyFlags.createdFromDate)
		if err != nil {
			log.Fatal("time.Parse createdFromDate: ", err)
		}
	}
	createdTo := time.Now()
	if historyFlags.createdToDate != "" {
		createdTo, err = time.Parse("2006-01-02", historyFlags.createdToDate)
		if err != nil {
			log.Fatal("time.Parse createdFromDate: ", err)
		}
	}
	log.Printf("listing createdFrom: %s, createdTo: %s\n", createdFrom.Format(time.RFC3339), createdTo.Format(time.RFC3339))

	// Get builds
	var builds []buildkite.Build
	if historyFlags.loadFrom == "" {
		log.Println("fetching builds from Buildkite")

		// Buildkite client
		config, err := buildkite.NewTokenConfig(flags.BuildkiteToken, false)
		if err != nil {
			log.Fatal("buildkite.NewTokenConfig: ", err)
		}
		bkc := buildkite.NewClient(config.Client())

		// Paginate results
		var nextPage = 1
		var pages int
		for nextPage > 0 {
			pages++

			// Newest is returned first https://buildkite.com/docs/apis/rest-api/builds#list-builds-for-a-pipeline
			pageBuilds, resp, err := bkc.Builds.ListByPipeline("sourcegraph", flags.Pipeline, &buildkite.BuildsListOptions{
				Branch:             flags.Branch,
				CreatedFrom:        createdFrom,
				CreatedTo:          createdTo,
				IncludeRetriedJobs: false,
				ListOptions: buildkite.ListOptions{
					Page: nextPage,
					// Fix to high page size just in case, default is 30
					// https://buildkite.com/docs/apis/rest-api#pagination
					PerPage: 99,
				},
			})
			if err != nil {
				log.Fatal("Builds.ListByPipeline: ", err)
			}

			builds = append(builds, pageBuilds...)
			nextPage = resp.NextPage
		}

		buildsJSON, err := json.Marshal(&builds)
		if err != nil {
			log.Fatal("json.Marshal(&builds): ", err)
		}
		if historyFlags.writeTo != "" {
			if err := os.WriteFile(historyFlags.writeTo, buildsJSON, os.ModePerm); err != nil {
				log.Fatal("os.WriteFile: ", err)
			}
			log.Println("wrote to " + historyFlags.writeTo)
		}
	} else {
		log.Printf("loading builds from %s\n", historyFlags.loadFrom)
		data, err := os.ReadFile(historyFlags.loadFrom)
		if err != nil {
			log.Fatal("os.ReadFile: ", err)
		}
		var cachedBuilds []buildkite.Build
		if err := json.Unmarshal(data, &cachedBuilds); err != nil {
			log.Fatal("json.Unmarshal: ", err)
		}
		for _, b := range cachedBuilds {
			if b.CreatedAt.Before(createdFrom) || b.CreatedAt.After(createdTo) {
				continue
			}
			builds = append(builds, b)
		}
	}
	log.Printf("loaded %d builds\n", len(builds))

	// Mark retried builds as failed
	var inferredFail int
	for _, b := range builds {
		for _, j := range b.Jobs {
			if j.RetriesCount > 0 {
				failed := "failed"
				b.State = &failed
				inferredFail += 1
			}
		}
	}
	log.Printf("inferred %d builds as failed", inferredFail)

	// Generate history
	totals, flakes, incidents := generateHistory(builds, createdTo, CheckOptions{
		FailuresThreshold: flags.FailuresThreshold,
		BuildTimeout:      time.Duration(flags.FailuresTimeoutMins) * time.Minute,
	})

	// Write to files
	var errs error
	errs = errors.CombineErrors(errs, writeCSV("./totals.csv", mapToRecords(totals)))
	errs = errors.CombineErrors(errs, writeCSV("./flakes.csv", mapToRecords(flakes)))
	errs = errors.CombineErrors(errs, writeCSV("./incidents.csv", mapToRecords(incidents)))
	if errs != nil {
		log.Fatal("csv.WriteAll: ", errs)
	}

	log.Println("done!")
}

func writeCSV(p string, records [][]string) error {
	f, err := os.Create(p)
	if err != nil {
		log.Fatal("os.OpenFile: ", err)
	}
	fCsv := csv.NewWriter(f)
	return fCsv.WriteAll(records)
}
