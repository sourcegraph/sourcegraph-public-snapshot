package changesets

import (
	"errors"
	"strings"
	"testing"

	"github.com/AaronO/go-git-http"

	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/notif"
)

// TestChangesetHook_couldAffectChangesets tests if a list of events is correctly
// filtered by couldAffectChangesets.
func TestChangesetHook_couldAffectChangesets(t *testing.T) {
	for _, tc := range []struct {
		id  events.EventID
		in  notif.GitPayload
		out bool
	}{
		{
			// contains an error
			notif.GitPushEvent,
			notif.GitPayload{
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
			notif.GitCreateEvent,
			notif.GitPayload{
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
			notif.GitPushEvent,
			notif.GitPayload{
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
			notif.GitPushEvent,
			notif.GitPayload{
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
			notif.GitDeleteEvent,
			notif.GitPayload{
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
		if out := couldAffectChangesets(tc.id, tc.in); out != tc.out {
			t.Errorf("Expected %v for %v, got %v", tc.out, tc.in, out)
		}
	}
}
