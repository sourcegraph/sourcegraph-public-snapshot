package campaigns

import (
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"

	cmpgn "github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

func TestFindMergeCommitID(t *testing.T) {
	now := time.Now()

	addCommit := func(event *cmpgn.ChangesetEvent, commitID string) *cmpgn.ChangesetEvent {
		switch m := event.Metadata.(type) {
		case *github.MergedEvent:
			m.Commit = github.Commit{
				OID: commitID,
			}
		case *bitbucketserver.Activity:
			m.Commit = &bitbucketserver.Commit{
				ID: commitID,
			}
		}
		return event
	}

	for _, tc := range []struct {
		name   string
		events ChangesetEvents
		want   string
	}{
		{
			name:   "nil events",
			events: nil,
			want:   "",
		},
		{
			name:   "no events",
			events: ChangesetEvents{},
			want:   "",
		},
		{
			name: "one bitbucket merge event",
			events: ChangesetEvents{
				addCommit(event(t, now, cmpgn.ChangesetEventKindBitbucketServerMerged, 1), "deadbeef"),
			},
			want: "deadbeef",
		},
		{
			name: "multiple bitbucket events with merge",
			events: ChangesetEvents{
				event(t, now, cmpgn.ChangesetEventKindBitbucketServerReopened, 1),
				addCommit(event(t, now, cmpgn.ChangesetEventKindBitbucketServerMerged, 1), "deadbeef"),
			},
			want: "deadbeef",
		},
		{
			name: "multiple bitbucket events no merge",
			events: ChangesetEvents{
				event(t, now, cmpgn.ChangesetEventKindBitbucketServerReopened, 1),
				event(t, now, cmpgn.ChangesetEventKindBitbucketServerDeclined, 1),
			},
			want: "",
		},
		{
			name: "one github merge event",
			events: ChangesetEvents{
				addCommit(event(t, now, cmpgn.ChangesetEventKindGitHubMerged, 1), "deadbeef"),
			},
			want: "deadbeef",
		},
		{
			name: "multiple github events with merge",
			events: ChangesetEvents{
				event(t, now, cmpgn.ChangesetEventKindGitHubClosed, 1),
				addCommit(event(t, now, cmpgn.ChangesetEventKindGitHubMerged, 1), "deadbeef"),
			},
			want: "deadbeef",
		},
		{
			name: "multiple github events no merge",
			events: ChangesetEvents{
				event(t, now, cmpgn.ChangesetEventKindGitHubClosed, 1),
				event(t, now, cmpgn.ChangesetEventKindGitHubReopened, 1),
			},
			want: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			have := tc.events.FindMergeCommitID()
			if have != tc.want {
				t.Fatalf("Want %q, have %q", tc.want, have)
			}
		})
	}
}
