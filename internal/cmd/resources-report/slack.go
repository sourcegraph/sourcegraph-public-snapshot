package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// see https://api.slack.com/reference/block-kit/blocks
type slackMessage struct {
	Blocks []slackBlock `json:"blocks"`
}

type slackBlock map[string]interface{}

const slackTextMarkdown = "mrkdwn"

type slackText struct {
	Type string `json:"type"` // just use `slackTextMarkdown`
	Text string `json:"text"`
}

func reportToSlack(ctx context.Context, webhook string, resources []Resource, since time.Time) error {
	// generate message to deliver
	blocks := []slackBlock{
		{
			"type": "section",
			"text": &slackText{
				Type: slackTextMarkdown,
				Text: fmt.Sprintf(":package: I've found %d resources created since %s!", len(resources), since.Format(time.RFC1123)),
			},
		},
		{
			"type": "divider",
		},
	}
	for _, resource := range resources {
		// if we have too many results, split up the message
		if len(blocks) > 40 {
			if err := sendSlackBlocks(ctx, webhook, blocks); err != nil {
				return err
			}
			blocks = nil
			break
		}
		block, err := resource.toSlackBlock()
		if err != nil {
			return err
		}
		blocks = append(blocks, block)
	}

	// send remaining blocks
	if len(blocks) > 0 {
		if err := sendSlackBlocks(ctx, webhook, blocks); err != nil {
			return err
		}
	}

	return nil
}

func sendSlackBlocks(ctx context.Context, webhook string, blocks []slackBlock) error {
	b, err := json.Marshal(&slackMessage{blocks})
	if err != nil {
		return fmt.Errorf("failed to post report to slack: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, webhook, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to post report to slack: %w", err)
	}
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to post report to slack: %w", err)
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		message, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to post report to slack: %s", resp.Status)
		} else {
			return fmt.Errorf("failed to post report to slack: %s", string(message))
		}
	}
	return nil
}
