package eventsutil

import (
	"fmt"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func LogGitPush(ctx context.Context) {
	LogEvent(ctx, "GitPush", nil)
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
		DeviceID:        deviceID,
		EventProperties: eventProperties,
	})
}
