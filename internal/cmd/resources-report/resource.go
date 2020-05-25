package main

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Platform string

const (
	PlatformGCP Platform = "gcp"
	PlatformAWS Platform = "aws"
)

type Resource struct {
	Platform   Platform
	Identifier string
	Type       string
	Location   string
	Owner      string
	Created    time.Time
	Meta       map[string]interface{}
}

func hasPrefix(value string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func generateReport(ctx context.Context, opts options, resources []Resource) error {
	// populate google sheet with data
	if err := updateSheet(ctx, *opts.sheetID, resources); err != nil {
		return fmt.Errorf("sheets: %w", err)
	}

	// generate message to deliver
	buttons := []slackBlock{
		newSlackButtonSheet(*opts.sheetID),
	}
	if *opts.runID != "" {
		buttons = append(buttons, newSlackButtonRun(*opts.runID))
	}
	blocks := []slackBlock{
		{
			"type": "section",
			"text": &slackText{
				Type: slackTextMarkdown,
				Text: fmt.Sprintf(":package: I've found %d resources created in the past %s!", len(resources), opts.window),
			},
		},
		{
			"type":     "actions",
			"elements": buttons,
		},
	}
	if err := sendSlackBlocks(ctx, *opts.slackWebhook, blocks); err != nil {
		return fmt.Errorf("slack: %w", err)
	}

	return nil
}
