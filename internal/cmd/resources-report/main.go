package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

type options struct {
	slackWebhook *string
	gcp          *bool
	aws          *bool

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
	resources := make([]Resource, 0)
	if *opts.gcp {
		rs, err := collectGCPResources(ctx, *opts.verbose)
		if err != nil {
			return fmt.Errorf("gcp: %w", err)
		}
		resources = append(resources, rs...)
	}
	if *opts.aws {
		rs, err := collectAWSResources(ctx, *opts.verbose)
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
		if err := reportToSlack(ctx, *opts.slackWebhook, resources); err != nil {
			return fmt.Errorf("slack: %w", err)
		}
	}

	log.Printf("done - collected a total of %d resources\n", len(resources))
	return nil
}
