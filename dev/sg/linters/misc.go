package linters

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/internal/honey"
)

var noLocalHost = runScript("Check for localhost usage", "dev/check/no-localhost-guard.sh") // CI:LOCALHOST_OK

func timeCheck(check *check.Check[*repo.State]) *check.Check[*repo.State] {
	if os.Getenv("CI") != "true" {
		return check
	}

	old := check.Check
	check.Check = func(ctx context.Context, out *std.Output, args *repo.State) error {
		t1 := time.Now()
		event := honey.NewEventWithFields("sg-lint", map[string]any{
			"check":         check.Name,
			"commit":        os.Getenv("COMMIT_SHA"),
			"target_branch": os.Getenv("BUILDKITE_PULL_REQUEST_BASE_BRANCH"),
			"is_aspect":     os.Getenv("ASPECT_WORKFLOWS_BUILD") == "1",
		})

		defer func() {
			if err := event.Send(); err != nil {
				fmt.Fprintf(os.Stderr, "Error sending honeycomb event: %v\n", err)
			}
		}()

		err := old(honey.WithEvent(ctx, event), out, args)
		t2 := time.Since(t1)
		event.AddField("duration", t2.Seconds())
		event.AddField("duration_ms", t2.Milliseconds())
		event.AddField("error", err)
		return err
	}

	return check
}
