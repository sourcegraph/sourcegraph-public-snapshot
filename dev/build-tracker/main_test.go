package main

import (
	"testing"

	"github.com/buildkite/go-buildkite/v3/buildkite"

	"github.com/sourcegraph/sourcegraph/dev/build-tracker/build"
)

func TestToBuildNotification(t *testing.T) {
	num := 160000
	url := "http://www.google.com"
	commit := "78926a5b3b836a8a104a5d5adf891e5626b1e405"
	pipelineID := "sourcegraph"
	exit := 999
	msg := "this is a test"
	t.Run("2 failed jobs", func(t *testing.T) {
		b := &build.Build{
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
			Pipeline: &build.Pipeline{buildkite.Pipeline{
				Name: &pipelineID,
			}},
			Steps: map[string]*build.Step{
				":one: fake step": build.NewStepFromJob(newJob(t, ":one: fake step", exit)),
				":two: fake step": build.NewStepFromJob(newJob(t, ":two: fake step", exit)),
			},
		}

		notification := determineBuildStatusNotification(b)

		if len(notification.Failed) != 2 {
			t.Errorf("got %d, wanted %d for failed jobs in BuildNotification", len(notification.Failed), 2)
		}
		if notification.BuildStatus != string(build.BuildFailed) {
			t.Errorf("got %s, wanted %s for Build Status in Notification", notification.BuildStatus, build.BuildFailed)
		}
	})
	t.Run("2 failed jobs initially and a late job should be 3 total jobs", func(t *testing.T) {
		b := &build.Build{
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
			Pipeline: &build.Pipeline{buildkite.Pipeline{
				Name: &pipelineID,
			}},
			Steps: map[string]*build.Step{
				":one: fake step": build.NewStepFromJob(newJob(t, ":one: fake step", exit)),
				":two: fake step": build.NewStepFromJob(newJob(t, ":two: fake step", exit)),
			},
		}

		notification := determineBuildStatusNotification(b)
		if len(notification.Failed) != 2 {
			t.Errorf("got %d, wanted %d for failed jobs in BuildNotification", len(notification.Failed), 2)
		}
		if notification.BuildStatus != string(build.BuildFailed) {
			t.Errorf("got %s, wanted %s for Build Status in Notification", notification.BuildStatus, build.BuildFailed)
		}

		err := b.AddJob(newJob(t, ":three: fake step", exit))
		if err != nil {
			t.Fatalf("failed to add job to build: %v", err)
		}

		notification = determineBuildStatusNotification(b)
		if len(notification.Failed) != 3 {
			t.Errorf("got %d, wanted %d for failed jobs in BuildNotification", len(notification.Failed), 3)
		}
		if notification.BuildStatus != string(build.BuildFailed) {
			t.Errorf("got %s, wanted %s for Build Status in Notification", notification.BuildStatus, build.BuildFailed)
		}
	})
	t.Run("2 failed jobs initially and both jobs passed should a fixed build", func(t *testing.T) {
		b := &build.Build{
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
			Pipeline: &build.Pipeline{buildkite.Pipeline{
				Name: &pipelineID,
			}},
			Steps: map[string]*build.Step{
				":one: fake step": build.NewStepFromJob(newJob(t, ":one: fake step", 999)),
				":two: fake step": build.NewStepFromJob(newJob(t, ":two: fake step", 999)),
			},
		}

		notification := determineBuildStatusNotification(b)
		if len(notification.Failed) != 2 {
			t.Errorf("got %d, wanted %d for failed jobs in BuildNotification", len(notification.Failed), 2)
		}
		if notification.BuildStatus != string(build.BuildFailed) {
			t.Errorf("got %s, wanted %s for Build Status in Notification", notification.BuildStatus, build.BuildFailed)
		}

		// Add the fixed job
		err := b.AddJob(newJob(t, ":one: fake step", 0))
		if err != nil {
			t.Fatalf("failed to add job to build: %v", err)
		}

		notification = determineBuildStatusNotification(b)
		if len(notification.Failed) != 1 {
			t.Errorf("got %d, wanted %d for failed jobs in BuildNotification", len(notification.Failed), 1)
		}
		if len(notification.Fixed) != 1 {
			t.Errorf("got %d, wanted %d for fixed jobs in BuildNotification", len(notification.Fixed), 1)
		}
		// Build should still be in a failed state ... since on job is still failing
		if notification.BuildStatus != string(build.BuildFailed) {
			t.Errorf("got %s, wanted %s for Build Status in Notification", notification.BuildStatus, build.BuildFailed)
		}

		// Add the fixed job
		err = b.AddJob(newJob(t, ":two: fake step", 0))
		if err != nil {
			t.Fatalf("failed to add job to build: %v", err)
		}

		notification = determineBuildStatusNotification(b)
		// All jobs should be fixed now
		if len(notification.Failed) != 0 {
			t.Errorf("got %d, wanted %d for failed jobs in BuildNotification", len(notification.Failed), 2)
		}
		if len(notification.Fixed) != 2 {
			t.Errorf("got %d, wanted %d for fixed jobs in BuildNotification", len(notification.Fixed), 2)
		}
		// All Jobs are fixed, so build should be in fixed state
		if notification.BuildStatus != string(build.BuildFixed) {
			t.Errorf("got %s, wanted %s for Build Status in Notification", notification.BuildStatus, build.BuildFixed)
		}
	})
}
