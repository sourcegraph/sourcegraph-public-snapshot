package main

import (
	"flag"
	"log"
	"os"
)

type options struct {
	slackWebhook *string
	gcp          *bool
	aws          *bool

	dry     *bool
	verbose *bool
}

func main() {
	if err := run(options{
		slackWebhook: flag.String("slack.webhook", os.Getenv("SLACK_WEBHOOK"), "Slack webhook to post updates to"),
		gcp:          flag.Bool("gcp", false, "If true, report on Google Cloud resources"),
		aws:          flag.Bool("aws", false, "If true, report on Amazon Web Services resources"),

		dry:     flag.Bool("dry", false, "If true, do not post updates to slack, but print them to stdout"),
		verbose: flag.Bool("verbose", false, "If true, print debug output to stdout"),
	}); err != nil {
		log.Fatal(err)
	}
}

func run(opts options) error {
	return nil
}
