package main

import (
	"flag"
	"testing"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/sourcegraph/log/logtest"
)

var RunIntegrationTest *bool = flag.Bool("RunIntegrationTest", false, "Run integrations tests")

func newJob(name string, exit int) *buildkite.Job {
	return &buildkite.Job{
		Name:       &name,
		ExitStatus: &exit,
	}
}

func TestSlack(t *testing.T) {
	flag.Parse()
	if !*RunIntegrationTest {
		t.Skip("Integration test not enabled")
	}
	logger := logtest.NoOp(t)

	token := MustEnvVar("SLACK_TOKEN")

	client := NewSlackClient(logger, token)

	num := 160000
	url := "http://www.google.com"
	commit := "78926a5b3b836a8a104a5d5adf891e5626b1e405"
	pipelineID := "sourcegraph"
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
					ID:   &pipelineID,
					Name: &pipelineID,
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
				*newJob(":one: fake step", exit),
				*newJob(":two: fake step", exit),
				*newJob(":three: fake step", exit),
				*newJob(":four: fake step", exit),
			},
		},
	)

	if err != nil {
		t.Fatalf("failed to send slack notification: %v", err)
	}
}
