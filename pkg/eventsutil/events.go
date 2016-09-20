package eventsutil

import (
	"fmt"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func LogGitPush(ctx context.Context) {
	LogEvent(ctx, "GitPush", nil)
}

func LogEvent(ctx context.Context, event string, eventProperties map[string]string) {
	if eventProperties == nil {
		eventProperties = make(map[string]string)
	}

	userAgent := UserAgentFromContext(ctx)
	if userAgent != "" {
		eventProperties["UserAgent"] = userAgent
	}

	Log(&sourcegraph.Event{
		Type:            fmt.Sprintf("Server%s", event),
		EventProperties: eventProperties,
	})
}
