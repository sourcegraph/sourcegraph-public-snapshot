package eventsutil

import (
	"fmt"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
)

func LogGitPush(ctx context.Context) {
	LogEvent(ctx, "GitPush", nil)
}

func LogAddRepoCompleted(ctx context.Context, language string, mirror, private bool) {
	m := make(map[string]string)
	m["language"] = language
	if mirror {
		m["mirror"] = "true"
	} else {
		m["mirror"] = "false"
	}
	if private {
		m["private"] = "true"
	} else {
		m["private"] = "false"
	}

	LogEvent(ctx, "AddRepoCompleted", m)
}

func LogCreateAccountCompleted(ctx context.Context) {
	LogEvent(ctx, "CreateAccountCompleted", nil)
}

func LogLinkGitHubCompleted(ctx context.Context) {
	LogEvent(ctx, "LinkGitHubCompleted", nil)
}

func LogBuildCompleted(ctx context.Context, success bool) {
	m := make(map[string]string)
	if success {
		m["result"] = "success"
	} else {
		m["result"] = "failure"
	}
	LogEvent(ctx, "BuildCompleted", m)
}

func LogEvent(ctx context.Context, event string, eventProperties map[string]string) {
	login := auth.ActorFromContext(ctx).Login
	deviceID := sourcegraphClientID

	if eventProperties == nil {
		eventProperties = make(map[string]string)
	}

	userAgent := UserAgentFromContext(ctx)
	if userAgent != "" {
		eventProperties["UserAgent"] = userAgent
	}

	Log(&sourcegraph.Event{
		Type:            fmt.Sprintf("Server%s", event),
		UserID:          login,
		DeviceID:        deviceID,
		EventProperties: eventProperties,
	})
}
