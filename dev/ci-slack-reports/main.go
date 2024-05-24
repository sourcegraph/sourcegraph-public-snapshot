package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/olekukonko/tablewriter"
	"github.com/sourcegraph/log"
	"golang.org/x/oauth2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/slack-go/slack"

	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime/contract"
)

func main() {
	ctx := context.Background()
	flag.Parse()

	liblog := log.Init(log.Resource{
		Name: "ci-slack-reports",
	}, log.NewSentrySink())
	defer liblog.Sync()
	logger := log.Scoped("run")

	env, err := contract.ParseEnv(os.Environ())
	if err != nil {
		logger.Fatal("failed to parse env", log.Error(err))
	}

	contract := contract.New(logger.Scoped("msp"), job{}, env)
	if contract.Diagnostics.ConfigureSentry(liblog) {
		logger.Info("Sentry integration enabled")
	}

	config := new(Config)
	config.Load(env)

	slackClient := slack.New(config.SlackToken)

	// for testing only
	token := oauth2.Token{
		AccessToken: os.Getenv("GCP_TOKEN"),
	}
	bq, err := bigquery.NewClient(ctx, config.BigQueryProjectID, option.WithTokenSource(oauth2.StaticTokenSource(&token)))
	if err != nil {
		logger.Fatal("failed to create BigQuery client", log.Error(err))
	}
	defer bq.Close()

	endTime := time.Now().Truncate(time.Hour * 24)
	startTime := endTime.Add(-time.Hour * 24 * 7)

	rows, err := runQuery(ctx, *bq, topThreeSumTime, []bigquery.QueryParameter{
		{
			Name:  "start_time",
			Value: startTime,
		},
		{
			Name:  "end_time",
			Value: endTime,
		},
	})
	if err != nil {
		logger.Fatal("error fetching top-threes", log.Error(err))
	}

	type Row struct {
		Target    string `bigquery:"target"`
		LastOwner string `bigquery:"last_owner"`
		// TotalTime in the week in minutes
		TotalTime  float32 `bigquery:"total_time"`
		CacheRatio float32 `bigquery:"cache_ratio"`
		Runs       int     `bigquery:"runs"`
	}

	teamToTests := make(map[string][]Row)

	for {
		var row Row
		err := rows.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			logger.Fatal("failed to read query results", log.Error(err))
		}

		teamToTests[row.LastOwner] = append(teamToTests[row.LastOwner], row)
	}

	teamToChannel := map[string]string{
		"owner_infra_devinfra": DefaultChannel,
	}

	for team, tests := range teamToTests {
		channel, ok := teamToChannel[team]
		if !ok {
			// continue
			channel = DefaultChannel
		}

		// TOOD: include more information such as links to notion, redash dashboard etc.
		message := fmt.Sprintf(":bazel: *Your team's top %d tests with most CI time in the past week* _for %s_\n", len(tests), strings.ReplaceAll(strings.TrimPrefix(team, "owner_"), "_", " "))

		var out strings.Builder
		out.WriteString("```\n")
		table := tablewriter.NewWriter(&out)
		table.SetHeader([]string{"Target", "Time", "Cache %", "Runs"})

		for _, test := range tests {
			duration, _ := time.ParseDuration(strconv.FormatFloat(float64(test.TotalTime), 'f', 2, 32) + "m")
			table.Append([]string{test.Target, duration.String(), strconv.Itoa(int(test.CacheRatio * 100)), strconv.Itoa(test.Runs)})
		}

		table.SetBorder(false)
		table.SetAutoWrapText(false)
		table.SetColWidth(80)
		table.Render()
		out.WriteString("\n```")

		_, err := slackClient.UploadFile(slack.FileUploadParameters{
			Content:        out.String(),
			Filetype:       "markdown",
			InitialComment: message,
			Channels:       []string{channel},
		})
		if err != nil {
			logger.Fatal("failed to upload table", log.Error(err))
		}
	}
}

type job struct{}

func (s job) Name() string { return "ci-slack-reports" }

func (s job) Version() string { return version.Version() }
