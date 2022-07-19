package main

import (
	"os"
	"testing"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/sourcegraph/log/logtest"
)

func TestSlack(t *testing.T) {
	logger := logtest.NoOp(t)

	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		t.Fatalf("SLACK_TOKEN Required")
	}

	client := NewSlackClient(logger, token)

	num := 160000
	url := "http://www.google.com"
	commit := "12345"
	name := ":white_check_mark: not a real step"
	pipelineID := "MINE"
	exit := 999
	msg := "this is a test"
	err := client.sendNotification(
		&Build{
			Build: buildkite.Build{
				Message: &msg,
				WebURL:  &url,
				Creator: &buildkite.Creator{
					AvatarURL: "https://www.gravatar.com/avatar/7d4f6781b10e48a94d1052c443d13149",
				},
				Pipeline: &buildkite.Pipeline{
					ID: &pipelineID,
				},
				Author: &buildkite.Author{
					Name:  "William Bezuidenhout",
					Email: "william.bezuidenhout@sourcegraph.com",
				},
				Number: &num,
				URL:    &url,
				Commit: &commit,
			},
			Jobs: []buildkite.Job{
				{
					Name:       &name,
					ExitStatus: &exit,
				},
			},
		},
	)

	if err != nil {
		t.Fatalf("failed to send slack notification: %v", err)
	}
}
