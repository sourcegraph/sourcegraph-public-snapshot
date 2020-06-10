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

type Resources []Resource

func (r Resources) Less(i, j int) bool { return r[i].Created.After(r[j].Created) }
func (r Resources) Len() int           { return len(r) }
func (r Resources) Swap(i, j int) {
	tmp := r[i]
	r[i] = r[j]
	r[j] = tmp
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
	// resources are sorted by creation beforehand
	highlightSince := time.Now().Add(-*opts.highlightWindow).UTC()
	var highlighted int
	for i, r := range resources {
		if r.Created.UTC().After(highlightSince) {
			highlighted = i + 1
		} else {
			break
		}
	}

	// populate google sheet with data
	if *opts.sheetID != "" {
		if err := updateSheet(ctx, *opts.sheetID, resources, highlighted); err != nil {
			return fmt.Errorf("sheets: %w", err)
		}
	}

	// generate message to deliver
	if *opts.slackWebhook != "" {
		buttons := []slackBlock{
			newSlackButtonSheet(*opts.sheetID),
			newSlackButtonDocs(),
		}
		if *opts.runID != "" {
			buttons = append(buttons, newSlackButtonRun(*opts.runID))
		}
		blocks := []slackBlock{
			{
				"type": "section",
				"text": &slackText{
					Type: slackTextMarkdown,
					Text: fmt.Sprintf(":package: I've found %d resources created in the past %s, with %d resources created in the past %s",
						len(resources), opts.window, highlighted, opts.highlightWindow),
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
	}

	return nil
}
