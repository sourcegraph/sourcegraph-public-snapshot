package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/google/go-github/v41/github"
	libhoney "github.com/honeycombio/libhoney-go"
	"github.com/slack-go/slack"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/dev/okay"
	"github.com/sourcegraph/sourcegraph/dev/team"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Flags denotes shared Buildchecker flags.
type Flags struct {
	BuildkiteToken      string
	Pipeline            string
	Branch              string
	SlackToken          string
	FailuresThreshold   int
	FailuresTimeoutMins int
}

func (f *Flags) Parse() {
	flag.StringVar(&f.BuildkiteToken, "buildkite.token", "", "mandatory buildkite token")
	flag.StringVar(&f.Pipeline, "pipeline", "sourcegraph", "name of the pipeline to inspect")
	flag.StringVar(&f.Branch, "branch", "main", "name of the branch to inspect")
	flag.StringVar(&f.SlackToken, "slack.token", "", "mandatory slack api token")

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
	flag.StringVar(&checkFlags.slackDebugWebhook, "slack.debug-webhook", "", "Slack Webhook URL to post debug results on")
	flag.StringVar(&checkFlags.slackDiscussionChannel, "slack.discussion-channel", "#buildkite-main", "Slack channel to ask everyone to head over to for discusison")

	historyFlags := &cmdHistoryFlags{}
	flag.StringVar(&historyFlags.createdFromDate, "created.from", "", "date in YYYY-MM-DD format")
	flag.StringVar(&historyFlags.createdToDate, "created.to", "", "date in YYYY-MM-DD format")
	flag.StringVar(&historyFlags.buildsLoadFrom, "builds.load-from", "", "file to load builds from - if unset, fetches from Buildkite")
	flag.StringVar(&historyFlags.buildsWriteTo, "builds.write-to", "", "file to write builds to (unused if loading from file)")
	flag.StringVar(&historyFlags.resultsCsvPath, "csv", "", "path for CSV results exports")
	flag.StringVar(&historyFlags.honeycombDataset, "honeycomb.dataset", "", "honeycomb dataset to publish to")
	flag.StringVar(&historyFlags.honeycombToken, "honeycomb.token", "", "honeycomb API token")
	flag.StringVar(&historyFlags.okayHQToken, "okayhq.token", "", "okayhq API token")
	flag.StringVar(&historyFlags.slackReportWebHook, "slack.report-webhook", "", "Slack Webhook URL to post weekly report on ")

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
	slc := slack.New(flags.SlackToken)

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

	buildsLoadFrom string
	buildsWriteTo  string

	resultsCsvPath   string
	honeycombDataset string
	honeycombToken   string

	okayHQToken string

	slackReportWebHook string
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
	if historyFlags.buildsLoadFrom == "" {
		// Load builds from Buildkite if no cached builds configured
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
		log.Printf("request paging progress:")
		for nextPage > 0 {
			pages++
			fmt.Printf(" %d", pages)

			// Newest is returned first https://buildkite.com/docs/apis/rest-api/builds#list-builds-for-a-pipeline
			pageBuilds, resp, err := bkc.Builds.ListByPipeline("sourcegraph", flags.Pipeline, &buildkite.BuildsListOptions{
				Branch:             flags.Branch,
				CreatedFrom:        createdFrom,
				CreatedTo:          createdTo,
				IncludeRetriedJobs: false,
				ListOptions: buildkite.ListOptions{
					Page:    nextPage,
					PerPage: 50,
				},
			})
			if err != nil {
				log.Fatal("Builds.ListByPipeline: ", err)
			}

			builds = append(builds, pageBuilds...)
			nextPage = resp.NextPage
		}
		fmt.Println() // end line for progress spinner

		if historyFlags.buildsWriteTo != "" {
			// Cache builds for ease of re-running analyses
			log.Printf("Caching discovered builds in %s\n", historyFlags.buildsWriteTo)
			buildsJSON, err := json.Marshal(&builds)
			if err != nil {
				log.Fatal("json.Marshal(&builds): ", err)
			}
			if err := os.WriteFile(historyFlags.buildsWriteTo, buildsJSON, os.ModePerm); err != nil {
				log.Fatal("os.WriteFile: ", err)
			}
			log.Println("wrote to " + historyFlags.buildsWriteTo)
		}
	} else {
		// Load builds from configured path
		log.Printf("loading builds from %s\n", historyFlags.buildsLoadFrom)
		data, err := os.ReadFile(historyFlags.buildsLoadFrom)
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
	checkOpts := CheckOptions{
		FailuresThreshold: flags.FailuresThreshold,
		BuildTimeout:      time.Duration(flags.FailuresTimeoutMins) * time.Minute,
	}
	log.Printf("running analysis with options: %+v\n", checkOpts)
	totals, flakes, incidents := generateHistory(builds, createdTo, checkOpts)

	// Prepare output
	if historyFlags.resultsCsvPath != "" {
		// Write to files
		log.Printf("Writing CSV results to %s\n", historyFlags.resultsCsvPath)
		var errs error
		errs = errors.CombineErrors(errs, writeCSV(filepath.Join(historyFlags.resultsCsvPath, "totals.csv"), mapToRecords(totals)))
		errs = errors.CombineErrors(errs, writeCSV(filepath.Join(historyFlags.resultsCsvPath, "flakes.csv"), mapToRecords(flakes)))
		errs = errors.CombineErrors(errs, writeCSV(filepath.Join(historyFlags.resultsCsvPath, "incidents.csv"), mapToRecords(incidents)))
		if errs != nil {
			log.Fatal("csv.WriteAll: ", errs)
		}
	}
	if historyFlags.honeycombDataset != "" {
		// Send to honeycomb
		log.Printf("Sending results to honeycomb dataset %q\n", historyFlags.honeycombDataset)
		hc, err := libhoney.NewClient(libhoney.ClientConfig{
			Dataset: historyFlags.honeycombDataset,
			APIKey:  historyFlags.honeycombToken,
		})
		if err != nil {
			log.Fatal("libhoney.NewClient: ", err)
		}
		var events []*libhoney.Event
		for _, record := range mapToRecords(totals) {
			recordDateString := record[0]
			ev := hc.NewEvent()
			ev.Timestamp, _ = time.Parse(dateFormat, recordDateString)
			ev.AddField("build_count", totals[recordDateString])         // date:count
			ev.AddField("incident_minutes", incidents[recordDateString]) // date:minutes
			ev.AddField("flake_count", flakes[recordDateString])         // date:count
			events = append(events, ev)
		}

		// send all at once
		log.Printf("Sending %d events to Honeycomb\n", len(events))
		var errs error
		for _, ev := range events {
			if err := ev.Send(); err != nil {
				errs = errors.Append(errs, err)
			}
		}
		hc.Close()
		if err != nil {
			log.Fatal("honeycomb.Send: ", err)
		}
		// log events that do not send
		for _, ev := range events {
			if strings.Contains(ev.String(), "sent:false") {
				log.Printf("An event did not send: %s", ev.String())
			}
		}
	}
	if historyFlags.okayHQToken != "" {
		okayCli := okay.NewClient(http.DefaultClient, historyFlags.okayHQToken)

		for _, record := range mapToRecords(totals) {
			recordDateString := record[0]
			eventTime, err := time.Parse("2006-01-02T00:00:00Z", recordDateString+"T00:00:00Z")
			if err != nil {
				log.Fatal("time.Parse: ", err)
			}

			metrics := map[string]okay.Metric{
				"totalCount":       okay.Count(totals[recordDateString]),
				"incidentDuration": okay.Duration(time.Duration(incidents[recordDateString]) * time.Minute),
				"flakeCount":       okay.Count(flakes[recordDateString]),
			}
			event := okay.Event{
				Name:      "buildStats",
				Timestamp: eventTime,
				UniqueKey: []string{"ts", "pipeline", "branch"},
				Properties: map[string]string{
					"ts":           eventTime.Format(time.RFC3339),
					"organization": "sourcegraph",
					"pipeline":     "sourcegraph",
					"branch":       "main",
				},
				Metrics: metrics,
			}

			err = okayCli.Push(&event)
			if err != nil {
				log.Fatal("Error storing OKAYHQ event okay.Push: ", err.Error())
			}
		}
		if err := okayCli.Flush(); err != nil {
			log.Fatal("Error posting to OKAYHQ okay.Flush: ", err.Error())
		}
	}
	if historyFlags.slackReportWebHook != "" {
		var totalBuilds int
		var totalTime int
		var totalFlakes int

		for _, total := range totals {
			totalBuilds += total
		}

		for _, incident := range incidents {
			totalTime += incident
		}

		for _, flake := range flakes {
			totalFlakes += flake
		}

		avgFlakes := math.Round(float64(totalFlakes) / float64(totalBuilds) * 100)

		message := generateSummaryMessage(historyFlags.createdFromDate, historyFlags.createdToDate, totalBuilds, totalFlakes, avgFlakes, time.Duration(totalTime*int(time.Minute)))

		if _, err := postSlackUpdate([]string{historyFlags.slackReportWebHook}, message); err != nil {
			log.Fatal("postSlackUpdate: ", err)
		}
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

func generateSummaryMessage(dateFrom, dateTo string, builds, flakes int, avgFlakes float64, downtime time.Duration) string {

	return fmt.Sprintf(`:bar_chart: Welcome to your weekly CI report for period *%s* to *%s*!
	• Total builds: *%d*
	• Total flakes: *%d*
	• Average %% of build flakes: *%v%%*
	• Total incident duration: *%v*

	For more information, view the dashboards at <https://app.okayhq.com/dashboards/3856903d-33ea-4d60-9719-68fec0eb4313/build-stats-kpis|OkayHQ>.
`, dateFrom, dateTo, builds, flakes, avgFlakes, downtime)
}
