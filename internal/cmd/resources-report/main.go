package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

const resultsBuffer = 5

type options struct {
	slackWebhook *string
	gcp          *bool
	aws          *bool
	window       *time.Duration

	runID   *string
	dry     *bool
	verbose *bool
	timeout *time.Duration
}

func main() {
	help := flag.Bool("help", false, "Show help text")
	opts := options{
		slackWebhook: flag.String("slack.webhook", os.Getenv("SLACK_WEBHOOK"), "Slack webhook to post updates to"),
		gcp:          flag.Bool("gcp", false, "Report on Google Cloud resources"),
		aws:          flag.Bool("aws", false, "Report on Amazon Web Services resources"),
		window:       flag.Duration("window", 48*time.Hour, "Restrict results to resources created within a period"),

		runID:   flag.String("run.id", os.Getenv("GITHUB_RUN_ID"), "ID of workflow run"),
		dry:     flag.Bool("dry", false, "Do not post updates to slack, but print them to stdout"),
		verbose: flag.Bool("verbose", false, "Print debug output to stdout"),
		timeout: flag.Duration("timeout", time.Minute, "Set a timeout for report generation"),
	}
	flag.Parse()
	if *help {
		flag.CommandLine.Usage()
		return
	}
	if err := run(opts); err != nil {
		log.Fatal(err)
	}
}

func run(opts options) error {
	ctx, cancel := context.WithTimeout(context.Background(), *opts.timeout)
	defer cancel()

	// collect resources
	var resources []Resource
	since := time.Now().UTC().Add(-*opts.window)
	if *opts.gcp {
		rs, err := collectGCPResources(ctx, since, *opts.verbose)
		if err != nil {
			return fmt.Errorf("gcp: %w", err)
		}
		resources = append(resources, rs...)
	}
	if *opts.aws {
		rs, err := collectAWSResources(ctx, since, *opts.verbose)
		if err != nil {
			return fmt.Errorf("aws: %w", err)
		}
		resources = append(resources, rs...)
	}

	// report results
	if *opts.dry {
		log.Println("dry run - collected resources:")
		log.Println(reportString(resources))
	} else {
		if err := reportToSlack(ctx, *opts.slackWebhook, resources, since, *opts.runID); err != nil {
			return fmt.Errorf("slack: %w", err)
		}
	}

	log.Printf("done - collected a total of %d resources created since %s", len(resources), since.String())
	return nil
}

func reportString(resources []Resource) string {
	var output string
	for _, r := range resources {
		output += fmt.Sprintf(" * %+v\n", r)
	}
	return output
}
