package main

import (
	"flag"
	"fmt"
	"testing"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/hexops/autogold"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/dev/team"
)

var RunSlackIntegrationTest = flag.Bool("RunSlackIntegrationTest", false, "Run Slack integration tests")
var RunGitHubIntegrationTest = flag.Bool("RunGitHubIntegrationTest", false, "Run Github integration tests")

func newJob(name string, exit int) *Job {
	return &Job{buildkite.Job{
		Name:       &name,
		ExitStatus: &exit,
	}}

}

func TestLargeAmountOfFailures(t *testing.T) {
	num := 160000
	commit := "ca7c44f79984ff8d645b580bfaaf08ce9a37a05d"
	url := "http://www.google.com"
	pipelineID := "sourcegraph"
	msg := "Large amount of failures test"
	build := Build{
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
		Pipeline: &Pipeline{buildkite.Pipeline{
			Name: &pipelineID,
		}},
		Jobs: map[string]Job{},
	}
	for i := 1; i <= 30; i++ {
		job := *newJob(fmt.Sprintf("%d fake step", i), i)
		build.Jobs[job.name()] = job
	}

	flag.Parse()
	if !*RunSlackIntegrationTest {
		t.Skip("Slack Integration test not enabled")
	}
	logger := logtest.NoOp(t)

	conf, err := configFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	client := NewNotificationClient(logger, conf.SlackToken, conf.GithubToken, DefaultChannel)

	_, err = client.sendNewMessage(&build)
	if err != nil {
		t.Fatalf("failed to send build: %s", err)
	}
}

func TestSlackMention(t *testing.T) {
	t.Run("If SlackID is empty, ask people to update their team.yml", func(t *testing.T) {
		result := slackMention(&team.Teammate{
			SlackID: "",
			Name:    "Bob Burgers",
			Email:   "bob@burgers.com",
			GitHub:  "bobbyb",
		})

		require.Equal(t, "Bob Burgers (bob@burgers.com) - We could not locate your Slack ID. Please check that your information in the Handbook team.yml file is correct", result)
	})
	t.Run("Use SlackID if it exists", func(t *testing.T) {
		result := slackMention(&team.Teammate{
			SlackID: "USE_ME",
			Name:    "Bob Burgers",
			Email:   "bob@burgers.com",
			GitHub:  "bobbyb",
		})
		require.Equal(t, "<@USE_ME>", result)
	})
}

func TestGetTeammateFromBuild(t *testing.T) {
	flag.Parse()
	if !*RunGitHubIntegrationTest {
		t.Skip("Github Integration test not enabled")
	}

	logger := logtest.NoOp(t)
	config, err := configFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("with nil author, commit author is still retrieved", func(t *testing.T) {
		client := NewNotificationClient(logger, config.SlackToken, config.GithubToken, DefaultChannel)

		num := 160000
		commit := "ca7c44f79984ff8d645b580bfaaf08ce9a37a05d"
		pipelineID := "sourcegraph"
		build := &Build{
			Build: buildkite.Build{
				Pipeline: &buildkite.Pipeline{
					ID:   &pipelineID,
					Name: &pipelineID,
				},
				Number: &num,
				Commit: &commit,
			},
			Pipeline: &Pipeline{buildkite.Pipeline{
				Name: &pipelineID,
			}},
			Jobs: map[string]Job{},
		}

		teammate, err := client.getTeammateForBuild(build)
		require.NoError(t, err)
		require.NotEqual(t, teammate.SlackID, "")
		require.Equal(t, teammate.Name, "Leo Papaloizos")
	})
	t.Run("commit author preferred over build author", func(t *testing.T) {
		client := NewNotificationClient(logger, config.SlackToken, config.GithubToken, DefaultChannel)

		num := 160000
		commit := "78926a5b3b836a8a104a5d5adf891e5626b1e405"
		pipelineID := "sourcegraph"
		build := &Build{
			Build: buildkite.Build{
				Pipeline: &buildkite.Pipeline{
					ID:   &pipelineID,
					Name: &pipelineID,
				},
				Number: &num,
				Commit: &commit,
				Author: &buildkite.Author{
					Name:  "William Bezuidenhout",
					Email: "william.bezuidenhout@sourcegraph.com",
				},
			},
			Pipeline: &Pipeline{buildkite.Pipeline{
				Name: &pipelineID,
			}},
			Jobs: map[string]Job{},
		}

		teammate, err := client.getTeammateForBuild(build)
		require.NoError(t, err)
		require.Equal(t, teammate.Name, "Ryan Slade")
	})
}

func TestSlackNotification(t *testing.T) {
	flag.Parse()
	if !*RunSlackIntegrationTest {
		t.Skip("Slack Integration test not enabled")
	}
	logger := logtest.NoOp(t)

	config, err := configFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	client := NewNotificationClient(logger, config.SlackToken, config.GithubToken, DefaultChannel)

	num := 160000
	url := "http://www.google.com"
	commit := "78926a5b3b836a8a104a5d5adf891e5626b1e405"
	pipelineID := "sourcegraph"
	exit := 999
	msg := "this is a test"
	t.Run("send new notification", func(t *testing.T) {
		build := &Build{
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
			Pipeline: &Pipeline{buildkite.Pipeline{
				Name: &pipelineID,
			}},
			Jobs: map[string]Job{
				":one: fake step":   *newJob(":one: fake step", exit),
				":two: fake step":   *newJob(":two: fake step", exit),
				":three: fake step": *newJob(":three: fake step", exit),
				":four: fake step":  *newJob(":four: fake step", exit),
			},
		}

		notification, err := client.sendNewMessage(build)
		if err != nil {
			t.Fatalf("failed to send slack notification: %v", err)
		}

		if notification == nil {
			t.Fatalf("expected not nil notificaiton after new notification")
		}
		if notification.ID == "" {
			t.Error("expected notification id to not be empty")
		}
		if notification.ChannelID == "" {
			t.Error("expected notification channel id to not be empty")
		}
	})
	t.Run("update notification", func(t *testing.T) {
		build := &Build{
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
			Pipeline: &Pipeline{buildkite.Pipeline{
				Name: &pipelineID,
			}},
			Jobs: map[string]Job{
				":one: fake step": *newJob(":one: fake step", exit),
			},
		}

		// post a new notification
		newNotification, err := client.sendNewMessage(build)
		if err != nil {
			t.Fatalf("failed to send slack notification: %v", err)
		}
		if newNotification == nil {
			t.Errorf("expected not nil notification after new message")
		}
		// now update the notification with additional jobs that failed
		build.Jobs[":alarm_clock: delayed job"] = *newJob(":clock: delayed job", exit)
		updatedNotification, err := client.sendUpdatedMessage(build, newNotification)
		if err != nil {
			t.Fatalf("failed to send slack notification: %v", err)
		}
		if updatedNotification == nil {
			t.Errorf("expected not nil notification after updated message")
		}
		if newNotification.Equals(updatedNotification) {
			t.Errorf("expected new and updated notifications to differ - new '%v' updated '%v'", newNotification, updatedNotification)
		}
	})
}
func TestServerNotify(t *testing.T) {
	flag.Parse()
	if !*RunSlackIntegrationTest {
		t.Skip("Slack Integration test not enabled")
	}
	logger := logtest.NoOp(t)

	config, err := configFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	server := NewServer(logger, *config)

	num := 160000
	url := "http://www.google.com"
	commit := "78926a5b3b836a8a104a5d5adf891e5626b1e405"
	pipelineID := "sourcegraph"
	exit := 999
	msg := "this is a test"
	build := &Build{
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
		Pipeline: &Pipeline{buildkite.Pipeline{
			Name: &pipelineID,
		}},
		Jobs: map[string]Job{
			":one: fake step": *newJob(":one: fake step", exit),
		},
	}

	// post a new notification
	err = server.notifyIfFailed(build)
	if err != nil {
		t.Fatalf("failed to send slack notification: %v", err)
	}
	if !build.hasNotification() {
		t.Fatalf("expected build to have a notification after failure notification")
	}
	// now update the notification with additional jobs that failed
	build.Jobs[":alarm_clock: delayed job"] = *newJob(":alarm_clock: delayed job", exit)
	// even though we call the same method, it has a different flow because the build has a previous notification present
	oldNotification := build.Notification
	err = server.notifyIfFailed(build)
	if err != nil {
		t.Fatalf("failed to send slack notification: %v", err)
	}
	if oldNotification.Equals(build.Notification) {
		t.Fatalf("expected notification to be different after new/update sending another notification")
	}
}

func TestGenerateHeader(t *testing.T) {
	for _, tc := range []struct {
		build *Build
		want  autogold.Value // use 'go test -update' to update
	}{
		{
			build: &Build{
				ConsecutiveFailure: 0,
			},
			want: autogold.Want("first failure", ":red_circle: Build 0 failed"),
		},
		{
			build: &Build{
				ConsecutiveFailure: 1,
			},
			want: autogold.Want("second failure", ":red_circle: Build 0 failed"),
		},
		{
			build: &Build{
				ConsecutiveFailure: 4,
			},
			want: autogold.Want("fifth failure", ":red_circle: Build 0 failed (:bangbang: 4th failure)"),
		},
	} {
		t.Run(tc.want.Name(), func(t *testing.T) {
			got := generateSlackHeader(tc.build)
			tc.want.Equal(t, got)
		})
	}
}
