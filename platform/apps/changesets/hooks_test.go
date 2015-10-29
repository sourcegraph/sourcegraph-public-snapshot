package changesets

import (
	"errors"
	"strings"
	"testing"

	"github.com/AaronO/go-git-http"

	"src.sourcegraph.com/sourcegraph/notif/githooks"
)

// TestChangesetHook_couldAffectChangesets tests if a list of events is correctly
// filtered by couldAffectChangesets.
func TestChangesetHook_couldAffectChangesets(t *testing.T) {
	for _, tc := range []struct {
		in  githooks.Payload
		out bool
	}{
		{
			// contains an error
			githooks.Payload{
				Type: githooks.GitPushEvent,
				Event: githttp.Event{
					Error:  errors.New("some error"),
					Branch: "branch",
					Type:   githttp.PUSH,
					Last:   strings.Repeat("X", 40),
					Commit: strings.Repeat("Y", 40),
				},
			}, false,
		}, {
			// is a new branch
			githooks.Payload{
				Type: githooks.GitCreateEvent,
				Event: githttp.Event{
					Error:  nil,
					Branch: "branch",
					Type:   githttp.PUSH,
					Last:   strings.Repeat("0", 40),
					Commit: strings.Repeat("Y", 40),
				},
			}, false,
		}, {
			// invalid commit value
			githooks.Payload{
				Type: githooks.GitPushEvent,
				Event: githttp.Event{
					Error:  nil,
					Branch: "branch",
					Type:   githttp.PUSH,
					Last:   strings.Repeat("X", 40),
					Commit: "HEAD",
				},
			}, false,
		}, {
			// push commit
			githooks.Payload{
				Type: githooks.GitPushEvent,
				Event: githttp.Event{
					Error:  nil,
					Branch: "branch",
					Type:   githttp.PUSH,
					Last:   strings.Repeat("X", 40),
					Commit: strings.Repeat("Y", 40),
				},
			}, true,
		}, {
			// push branch deletion
			githooks.Payload{
				Type: githooks.GitDeleteEvent,
				Event: githttp.Event{
					Error:  nil,
					Branch: "branch",
					Type:   githttp.PUSH,
					Last:   strings.Repeat("X", 40),
					Commit: strings.Repeat("0", 40),
				},
			}, true,
		},
	} {
		if out := couldAffectChangesets(tc.in); out != tc.out {
			t.Errorf("Expected %v for %v, got %v", tc.out, tc.in, out)
		}
	}
}
