package a8n

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

func TestCalcCounts(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	daysAgo := func(days int) time.Time { return now.AddDate(0, 0, -days) }

	tests := []struct {
		codehosts  string
		name       string
		changesets []*a8n.Changeset
		start      time.Time
		end        time.Time
		events     []Event
		want       []*ChangesetCounts
	}{
		{
			codehosts: "github",
			name:      "single changeset open merged",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(2)),
			},
			start: daysAgo(2),
			events: []Event{
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(2), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(1), Total: 1, Merged: 1},
				{Time: daysAgo(0), Total: 1, Merged: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "single changeset open merged",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(2)),
			},
			start: daysAgo(2),
			events: []Event{
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindBitbucketServerMerged, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(2), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(1), Total: 1, Merged: 1},
				{Time: daysAgo(0), Total: 1, Merged: 1},
			},
		},
		{
			name: "start end time on subset of events",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(3)),
			},
			start: daysAgo(4),
			end:   daysAgo(2),
			events: []Event{
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(4), Total: 0, Open: 0},
				{Time: daysAgo(3), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(2), Total: 1, Open: 1, OpenPending: 1},
			},
		},
		{
			codehosts: "github",
			name:      "single changeset created and closed before start time",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(8)),
			},
			start: daysAgo(4),
			end:   daysAgo(2),
			events: []Event{
				fakeEvent{t: daysAgo(7), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(4), Total: 1, Merged: 1},
				{Time: daysAgo(3), Total: 1, Merged: 1},
				{Time: daysAgo(2), Total: 1, Merged: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "single changeset created and closed before start time",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(8)),
			},
			start: daysAgo(4),
			end:   daysAgo(2),
			events: []Event{
				fakeEvent{t: daysAgo(7), kind: a8n.ChangesetEventKindBitbucketServerMerged, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(4), Total: 1, Merged: 1},
				{Time: daysAgo(3), Total: 1, Merged: 1},
				{Time: daysAgo(2), Total: 1, Merged: 1},
			},
		},
		{
			name: "start time not even x*24hours before end time",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(2)),
			},
			start: daysAgo(3),
			end:   now.Add(-18 * time.Hour),
			events: []Event{
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(2).Add(-18 * time.Hour), Total: 0, Merged: 0},
				{Time: daysAgo(1).Add(-18 * time.Hour), Total: 1, Open: 1, OpenPending: 1},
				{Time: now.Add(-18 * time.Hour), Total: 1, Merged: 1},
			},
		},
		{
			codehosts: "github",
			name:      "multiple changesets open merged",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(2)),
				ghChangeset(2, daysAgo(2)),
			},
			start: daysAgo(2),
			events: []Event{
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 2},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(2), Total: 2, Open: 2, OpenPending: 2},
				{Time: daysAgo(1), Total: 2, Merged: 2},
				{Time: daysAgo(0), Total: 2, Merged: 2},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "multiple changesets open merged",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(2)),
				bbsChangeset(2, daysAgo(2)),
			},
			start: daysAgo(2),
			events: []Event{
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindBitbucketServerMerged, id: 1},
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindBitbucketServerMerged, id: 2},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(2), Total: 2, Open: 2, OpenPending: 2},
				{Time: daysAgo(1), Total: 2, Merged: 2},
				{Time: daysAgo(0), Total: 2, Merged: 2},
			},
		},
		{
			codehosts: "github",
			name:      "multiple changesets open merged different times",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(3)),
				ghChangeset(2, daysAgo(2)),
			},
			start: daysAgo(4),
			events: []Event{
				fakeEvent{t: daysAgo(2), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 2},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(4), Total: 0, Open: 0},
				{Time: daysAgo(3), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(2), Total: 2, Open: 1, OpenPending: 1, Merged: 1},
				{Time: daysAgo(1), Total: 2, Merged: 2},
				{Time: daysAgo(0), Total: 2, Merged: 2},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "multiple changesets open merged different times",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(3)),
				bbsChangeset(2, daysAgo(2)),
			},
			start: daysAgo(4),
			events: []Event{
				fakeEvent{t: daysAgo(2), kind: a8n.ChangesetEventKindBitbucketServerMerged, id: 1},
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindBitbucketServerMerged, id: 2},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(4), Total: 0, Open: 0},
				{Time: daysAgo(3), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(2), Total: 2, Open: 1, OpenPending: 1, Merged: 1},
				{Time: daysAgo(1), Total: 2, Merged: 2},
				{Time: daysAgo(0), Total: 2, Merged: 2},
			},
		},
		{
			codehosts: "github",
			name:      "changeset merged and closed at same time",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(2)),
			},
			start: daysAgo(2),
			events: []Event{
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubClosed, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(2), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(1), Total: 1, Merged: 1},
				{Time: daysAgo(0), Total: 1, Merged: 1},
			},
		},
		{
			codehosts: "github",
			name:      "changeset merged and closed at same time, reversed order in slice",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(2)),
			},
			start: daysAgo(2),
			events: []Event{
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubClosed, id: 1},
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(2), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(1), Total: 1, Merged: 1},
				{Time: daysAgo(0), Total: 1, Merged: 1},
			},
		},
		{
			codehosts: "github",
			name:      "single changeset open closed reopened merged",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(4)),
			},
			start: daysAgo(5),
			events: []Event{
				fakeEvent{t: daysAgo(3), kind: a8n.ChangesetEventKindGitHubClosed, id: 1},
				fakeEvent{t: daysAgo(2), kind: a8n.ChangesetEventKindGitHubReopened, id: 1},
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(5), Total: 0, Open: 0},
				{Time: daysAgo(4), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(3), Total: 1, Open: 0, Closed: 1},
				{Time: daysAgo(2), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(1), Total: 1, Merged: 1},
				{Time: daysAgo(0), Total: 1, Merged: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "single changeset open declined reopened merged",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(4)),
			},
			start: daysAgo(5),
			events: []Event{
				fakeEvent{t: daysAgo(3), kind: a8n.ChangesetEventKindBitbucketServerDeclined, id: 1},
				fakeEvent{t: daysAgo(2), kind: a8n.ChangesetEventKindBitbucketServerReopened, id: 1},
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindBitbucketServerMerged, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(5), Total: 0, Open: 0},
				{Time: daysAgo(4), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(3), Total: 1, Open: 0, Closed: 1},
				{Time: daysAgo(2), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(1), Total: 1, Merged: 1},
				{Time: daysAgo(0), Total: 1, Merged: 1},
			},
		},
		{
			codehosts: "github",
			name:      "multiple changesets open closed reopened merged different times",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(5)),
				ghChangeset(2, daysAgo(4)),
			},
			start: daysAgo(6),
			events: []Event{
				fakeEvent{t: daysAgo(4), kind: a8n.ChangesetEventKindGitHubClosed, id: 1},
				fakeEvent{t: daysAgo(3), kind: a8n.ChangesetEventKindGitHubClosed, id: 2},
				fakeEvent{t: daysAgo(3), kind: a8n.ChangesetEventKindGitHubReopened, id: 1},
				fakeEvent{t: daysAgo(2), kind: a8n.ChangesetEventKindGitHubReopened, id: 2},
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
				fakeEvent{t: daysAgo(0), kind: a8n.ChangesetEventKindGitHubMerged, id: 2},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(6), Total: 0, Open: 0},
				{Time: daysAgo(5), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(4), Total: 2, Open: 1, OpenPending: 1, Closed: 1},
				{Time: daysAgo(3), Total: 2, Open: 1, OpenPending: 1, Closed: 1},
				{Time: daysAgo(2), Total: 2, Open: 2, OpenPending: 2},
				{Time: daysAgo(1), Total: 2, Open: 1, OpenPending: 1, Merged: 1},
				{Time: daysAgo(0), Total: 2, Merged: 2},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "multiple changesets open declined reopened merged different times",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(5)),
				bbsChangeset(2, daysAgo(4)),
			},
			start: daysAgo(6),
			events: []Event{
				fakeEvent{t: daysAgo(4), kind: a8n.ChangesetEventKindBitbucketServerDeclined, id: 1},
				fakeEvent{t: daysAgo(3), kind: a8n.ChangesetEventKindBitbucketServerDeclined, id: 2},
				fakeEvent{t: daysAgo(3), kind: a8n.ChangesetEventKindBitbucketServerReopened, id: 1},
				fakeEvent{t: daysAgo(2), kind: a8n.ChangesetEventKindBitbucketServerReopened, id: 2},
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindBitbucketServerMerged, id: 1},
				fakeEvent{t: daysAgo(0), kind: a8n.ChangesetEventKindBitbucketServerMerged, id: 2},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(6), Total: 0, Open: 0},
				{Time: daysAgo(5), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(4), Total: 2, Open: 1, OpenPending: 1, Closed: 1},
				{Time: daysAgo(3), Total: 2, Open: 1, OpenPending: 1, Closed: 1},
				{Time: daysAgo(2), Total: 2, Open: 2, OpenPending: 2},
				{Time: daysAgo(1), Total: 2, Open: 1, OpenPending: 1, Merged: 1},
				{Time: daysAgo(0), Total: 2, Merged: 2},
			},
		},
		{
			codehosts: "github",
			name:      "single changeset open closed reopened merged, unsorted events",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(4)),
			},
			start: daysAgo(5),
			events: []Event{
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
				fakeEvent{t: daysAgo(3), kind: a8n.ChangesetEventKindGitHubClosed, id: 1},
				fakeEvent{t: daysAgo(2), kind: a8n.ChangesetEventKindGitHubReopened, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(5), Total: 0, Open: 0},
				{Time: daysAgo(4), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(3), Total: 1, Open: 0, Closed: 1},
				{Time: daysAgo(2), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(1), Total: 1, Merged: 1},
				{Time: daysAgo(0), Total: 1, Merged: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "single changeset open closed reopened merged, unsorted events",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(4)),
			},
			start: daysAgo(5),
			events: []Event{
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindBitbucketServerMerged, id: 1},
				fakeEvent{t: daysAgo(3), kind: a8n.ChangesetEventKindBitbucketServerDeclined, id: 1},
				fakeEvent{t: daysAgo(2), kind: a8n.ChangesetEventKindBitbucketServerReopened, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(5), Total: 0, Open: 0},
				{Time: daysAgo(4), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(3), Total: 1, Open: 0, Closed: 1},
				{Time: daysAgo(2), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(1), Total: 1, Merged: 1},
				{Time: daysAgo(0), Total: 1, Merged: 1},
			},
		},
		{
			codehosts: "github",
			name:      "single changeset open, approved, merged",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(3)),
			},
			start: daysAgo(4),
			events: []Event{
				ghReview(1, daysAgo(2), "user1", "APPROVED"),
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(4), Total: 0, Open: 0},
				{Time: daysAgo(3), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(2), Total: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
				{Time: daysAgo(1), Total: 1, Merged: 1},
				{Time: daysAgo(0), Total: 1, Merged: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "single changeset open, approved, merged",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(3)),
			},
			start: daysAgo(4),
			events: []Event{
				bbsActivity(1, daysAgo(2), "user1", a8n.ChangesetEventKindBitbucketServerApproved),
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindBitbucketServerMerged, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(4), Total: 0, Open: 0},
				{Time: daysAgo(3), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(2), Total: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
				{Time: daysAgo(1), Total: 1, Merged: 1},
				{Time: daysAgo(0), Total: 1, Merged: 1},
			},
		},
		{
			codehosts: "github",
			name:      "single changeset open, approved, closed, reopened",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(3)),
			},
			start: daysAgo(3),
			events: []Event{
				ghReview(1, daysAgo(2), "user1", "APPROVED"),
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubClosed, id: 1},
				fakeEvent{t: daysAgo(0), kind: a8n.ChangesetEventKindGitHubReopened, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(3), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(2), Total: 1, Open: 1, OpenApproved: 1},
				{Time: daysAgo(1), Total: 1, Closed: 1},
				{Time: daysAgo(0), Total: 1, Open: 1, OpenApproved: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "single changeset open, approved, declined, reopened",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(3)),
			},
			start: daysAgo(3),
			events: []Event{
				bbsActivity(1, daysAgo(2), "user1", a8n.ChangesetEventKindBitbucketServerApproved),
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindBitbucketServerDeclined, id: 1},
				fakeEvent{t: daysAgo(0), kind: a8n.ChangesetEventKindBitbucketServerReopened, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(3), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(2), Total: 1, Open: 1, OpenApproved: 1},
				{Time: daysAgo(1), Total: 1, Closed: 1},
				{Time: daysAgo(0), Total: 1, Open: 1, OpenApproved: 1},
			},
		},
		{
			codehosts: "github",
			name:      "single changeset open, approved, closed, merged",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(3)),
			},
			start: daysAgo(3),
			events: []Event{
				ghReview(1, daysAgo(2), "user1", "APPROVED"),
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubClosed, id: 1},
				fakeEvent{t: daysAgo(0), kind: a8n.ChangesetEventKindGitHubReopened, id: 1},
				fakeEvent{t: daysAgo(0), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(3), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(2), Total: 1, Open: 1, OpenApproved: 1},
				{Time: daysAgo(1), Total: 1, Closed: 1},
				{Time: daysAgo(0), Total: 1, Merged: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "single changeset open, approved, closed, merged",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(3)),
			},
			start: daysAgo(3),
			events: []Event{
				bbsActivity(1, daysAgo(2), "user1", a8n.ChangesetEventKindBitbucketServerApproved),
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindBitbucketServerDeclined, id: 1},
				fakeEvent{t: daysAgo(0), kind: a8n.ChangesetEventKindBitbucketServerReopened, id: 1},
				fakeEvent{t: daysAgo(0), kind: a8n.ChangesetEventKindBitbucketServerMerged, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(3), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(2), Total: 1, Open: 1, OpenApproved: 1},
				{Time: daysAgo(1), Total: 1, Closed: 1},
				{Time: daysAgo(0), Total: 1, Merged: 1},
			},
		},
		{
			codehosts: "github",
			name:      "single changeset open, changes-requested, closed, reopened",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(3)),
			},
			start: daysAgo(3),
			events: []Event{
				ghReview(1, daysAgo(2), "user1", "CHANGES_REQUESTED"),
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubClosed, id: 1},
				fakeEvent{t: daysAgo(0), kind: a8n.ChangesetEventKindGitHubReopened, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(3), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(2), Total: 1, Open: 1, OpenChangesRequested: 1},
				{Time: daysAgo(1), Total: 1, Closed: 1},
				{Time: daysAgo(0), Total: 1, Open: 1, OpenChangesRequested: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "single changeset open, changes-requested, closed, reopened",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(3)),
			},
			start: daysAgo(3),
			events: []Event{
				bbsActivity(1, daysAgo(2), "user1", a8n.ChangesetEventKindBitbucketServerReviewed),
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindBitbucketServerDeclined, id: 1},
				fakeEvent{t: daysAgo(0), kind: a8n.ChangesetEventKindBitbucketServerReopened, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(3), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(2), Total: 1, Open: 1, OpenChangesRequested: 1},
				{Time: daysAgo(1), Total: 1, Closed: 1},
				{Time: daysAgo(0), Total: 1, Open: 1, OpenChangesRequested: 1},
			},
		},
		{
			codehosts: "github",
			name:      "single changeset open, changes-requested, closed, merged",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(3)),
			},
			start: daysAgo(3),
			events: []Event{
				ghReview(1, daysAgo(2), "user1", "CHANGES_REQUESTED"),
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubClosed, id: 1},
				fakeEvent{t: daysAgo(0), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(3), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(2), Total: 1, Open: 1, OpenChangesRequested: 1},
				{Time: daysAgo(1), Total: 1, Closed: 1},
				{Time: daysAgo(0), Total: 1, Merged: 1},
			},
		},
		{
			codehosts: "github",
			name:      "single changeset open, comment review, approved, merged",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(3)),
			},
			start: daysAgo(4),
			events: []Event{
				ghReview(1, daysAgo(3), "user1", "COMMENTED"),
				ghReview(1, daysAgo(2), "user2", "APPROVED"),
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(4), Total: 0, Open: 0},
				{Time: daysAgo(3), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(2), Total: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
				{Time: daysAgo(1), Total: 1, Merged: 1},
				{Time: daysAgo(0), Total: 1, Merged: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "single changeset open, comment review, approved, merged",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(3)),
			},
			start: daysAgo(4),
			events: []Event{
				bbsActivity(1, daysAgo(3), "user1", a8n.ChangesetEventKindBitbucketServerCommented),
				bbsActivity(1, daysAgo(2), "user2", a8n.ChangesetEventKindBitbucketServerApproved),
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindBitbucketServerMerged, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(4), Total: 0, Open: 0},
				{Time: daysAgo(3), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(2), Total: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
				{Time: daysAgo(1), Total: 1, Merged: 1},
				{Time: daysAgo(0), Total: 1, Merged: 1},
			},
		},
		{
			codehosts: "github",
			name:      "single changeset multiple approvals counting once",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(1)),
			},
			start: daysAgo(1),
			events: []Event{
				ghReview(1, daysAgo(1), "user1", "APPROVED"),
				ghReview(1, daysAgo(0), "user2", "APPROVED"),
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(1), Total: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
				{Time: daysAgo(0), Total: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "single changeset multiple approvals counting once",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(1)),
			},
			start: daysAgo(1),
			events: []Event{
				bbsActivity(1, daysAgo(1), "user1", a8n.ChangesetEventKindBitbucketServerApproved),
				bbsActivity(1, daysAgo(0), "user2", a8n.ChangesetEventKindBitbucketServerApproved),
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(1), Total: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
				{Time: daysAgo(0), Total: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
			},
		},
		{
			codehosts: "github",
			name:      "single changeset multiple changes-requested reviews counting once",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(1)),
			},
			start: daysAgo(1),
			events: []Event{
				ghReview(1, daysAgo(1), "user1", "CHANGES_REQUESTED"),
				ghReview(1, daysAgo(0), "user2", "CHANGES_REQUESTED"),
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(1), Total: 1, Open: 1, OpenPending: 0, OpenChangesRequested: 1},
				{Time: daysAgo(0), Total: 1, Open: 1, OpenPending: 0, OpenChangesRequested: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "single changeset multiple changes-requested reviews counting once",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(1)),
			},
			start: daysAgo(1),
			events: []Event{
				bbsActivity(1, daysAgo(1), "user1", a8n.ChangesetEventKindBitbucketServerReviewed),
				bbsActivity(1, daysAgo(0), "user2", a8n.ChangesetEventKindBitbucketServerReviewed),
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(1), Total: 1, Open: 1, OpenPending: 0, OpenChangesRequested: 1},
				{Time: daysAgo(0), Total: 1, Open: 1, OpenPending: 0, OpenChangesRequested: 1},
			},
		},
		{
			codehosts: "github",
			name:      "single changeset open, changes-requested, merged",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(3)),
			},
			start: daysAgo(4),
			events: []Event{
				ghReview(1, daysAgo(2), "user1", "CHANGES_REQUESTED"),
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(4), Total: 0, Open: 0},
				{Time: daysAgo(3), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(2), Total: 1, Open: 1, OpenPending: 0, OpenChangesRequested: 1},
				{Time: daysAgo(1), Total: 1, Merged: 1},
				{Time: daysAgo(0), Total: 1, Merged: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "single changeset open, changes-requested, merged",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(3)),
			},
			start: daysAgo(4),
			events: []Event{
				bbsActivity(1, daysAgo(2), "user1", a8n.ChangesetEventKindBitbucketServerReviewed),
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindBitbucketServerMerged, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(4), Total: 0, Open: 0},
				{Time: daysAgo(3), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(2), Total: 1, Open: 1, OpenPending: 0, OpenChangesRequested: 1},
				{Time: daysAgo(1), Total: 1, Merged: 1},
				{Time: daysAgo(0), Total: 1, Merged: 1},
			},
		},
		{
			codehosts: "github",
			name:      "multiple changesets open different review stages before merge",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(6)),
				ghChangeset(2, daysAgo(6)),
				ghChangeset(3, daysAgo(6)),
			},
			start: daysAgo(7),
			events: []Event{
				ghReview(1, daysAgo(5), "user1", "APPROVED"),
				fakeEvent{t: daysAgo(3), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
				ghReview(2, daysAgo(4), "user1", "APPROVED"),
				ghReview(2, daysAgo(3), "user2", "APPROVED"),
				fakeEvent{t: daysAgo(2), kind: a8n.ChangesetEventKindGitHubMerged, id: 2},
				ghReview(3, daysAgo(2), "user1", "CHANGES_REQUESTED"),
				ghReview(3, daysAgo(1), "user2", "CHANGES_REQUESTED"),
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 3},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(7), Total: 0, Open: 0},
				{Time: daysAgo(6), Total: 3, Open: 3, OpenPending: 3},
				{Time: daysAgo(5), Total: 3, Open: 3, OpenPending: 2, OpenApproved: 1},
				{Time: daysAgo(4), Total: 3, Open: 3, OpenPending: 1, OpenApproved: 2},
				{Time: daysAgo(3), Total: 3, Open: 2, OpenPending: 1, OpenApproved: 1, Merged: 1},
				{Time: daysAgo(2), Total: 3, Open: 1, OpenPending: 0, OpenChangesRequested: 1, Merged: 2},
				{Time: daysAgo(1), Total: 3, Merged: 3},
				{Time: daysAgo(0), Total: 3, Merged: 3},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "multiple changesets open different review stages before merge",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(6)),
				bbsChangeset(2, daysAgo(6)),
				bbsChangeset(3, daysAgo(6)),
			},
			start: daysAgo(7),
			events: []Event{
				bbsActivity(1, daysAgo(5), "user1", a8n.ChangesetEventKindBitbucketServerApproved),
				fakeEvent{t: daysAgo(3), kind: a8n.ChangesetEventKindBitbucketServerMerged, id: 1},
				bbsActivity(2, daysAgo(4), "user1", a8n.ChangesetEventKindBitbucketServerApproved),
				bbsActivity(2, daysAgo(3), "user2", a8n.ChangesetEventKindBitbucketServerApproved),
				fakeEvent{t: daysAgo(2), kind: a8n.ChangesetEventKindBitbucketServerMerged, id: 2},
				bbsActivity(3, daysAgo(2), "user1", a8n.ChangesetEventKindBitbucketServerReviewed),
				bbsActivity(3, daysAgo(1), "user2", a8n.ChangesetEventKindBitbucketServerReviewed),
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindBitbucketServerMerged, id: 3},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(7), Total: 0, Open: 0},
				{Time: daysAgo(6), Total: 3, Open: 3, OpenPending: 3},
				{Time: daysAgo(5), Total: 3, Open: 3, OpenPending: 2, OpenApproved: 1},
				{Time: daysAgo(4), Total: 3, Open: 3, OpenPending: 1, OpenApproved: 2},
				{Time: daysAgo(3), Total: 3, Open: 2, OpenPending: 1, OpenApproved: 1, Merged: 1},
				{Time: daysAgo(2), Total: 3, Open: 1, OpenPending: 0, OpenChangesRequested: 1, Merged: 2},
				{Time: daysAgo(1), Total: 3, Merged: 3},
				{Time: daysAgo(0), Total: 3, Merged: 3},
			},
		},
		{
			codehosts: "github",
			name:      "time slice of multiple changesets in different stages before merge",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(6)),
				ghChangeset(2, daysAgo(6)),
				ghChangeset(3, daysAgo(6)),
			},
			// Same test as above, except we only look at 3 days in the middle
			start: daysAgo(4),
			end:   daysAgo(2),
			events: []Event{
				ghReview(1, daysAgo(5), "user1", "APPROVED"),
				fakeEvent{t: daysAgo(3), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
				ghReview(2, daysAgo(4), "user1", "APPROVED"),
				fakeEvent{t: daysAgo(2), kind: a8n.ChangesetEventKindGitHubMerged, id: 2},
				ghReview(3, daysAgo(2), "user1", "CHANGES_REQUESTED"),
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 3},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(4), Total: 3, Open: 3, OpenPending: 1, OpenApproved: 2},
				{Time: daysAgo(3), Total: 3, Open: 2, OpenPending: 1, OpenApproved: 1, Merged: 1},
				{Time: daysAgo(2), Total: 3, Open: 1, OpenPending: 0, OpenChangesRequested: 1, Merged: 2},
			},
		},
		{
			codehosts: "github",
			name:      "single changeset with changes-requested then approved by same person",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(1)),
			},
			start: daysAgo(1),
			events: []Event{
				ghReview(1, daysAgo(1), "user1", "CHANGES_REQUESTED"),
				ghReview(1, daysAgo(0), "user1", "APPROVED"),
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(1), Total: 1, Open: 1, OpenChangesRequested: 1},
				{Time: daysAgo(0), Total: 1, Open: 1, OpenApproved: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "single changeset with changes-requested then approved by same person",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(1)),
			},
			start: daysAgo(1),
			events: []Event{
				bbsActivity(1, daysAgo(1), "user1", a8n.ChangesetEventKindBitbucketServerReviewed),
				bbsActivity(1, daysAgo(0), "user1", a8n.ChangesetEventKindBitbucketServerApproved),
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(1), Total: 1, Open: 1, OpenChangesRequested: 1},
				{Time: daysAgo(0), Total: 1, Open: 1, OpenApproved: 1},
			},
		},
		{
			codehosts: "github",
			name:      "single changeset with approved then changes-requested by same person",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(1)),
			},
			start: daysAgo(1),
			events: []Event{
				ghReview(1, daysAgo(1), "user1", "APPROVED"),
				ghReview(1, daysAgo(0), "user1", "CHANGES_REQUESTED"),
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(1), Total: 1, Open: 1, OpenApproved: 1},
				{Time: daysAgo(0), Total: 1, Open: 1, OpenChangesRequested: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "single changeset with approved then changes-requested by same person",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(1)),
			},
			start: daysAgo(1),
			events: []Event{
				bbsActivity(1, daysAgo(1), "user1", a8n.ChangesetEventKindBitbucketServerApproved),
				bbsActivity(1, daysAgo(0), "user1", a8n.ChangesetEventKindBitbucketServerReviewed),
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(1), Total: 1, Open: 1, OpenApproved: 1},
				{Time: daysAgo(0), Total: 1, Open: 1, OpenChangesRequested: 1},
			},
		},
		{
			codehosts: "github",
			name:      "single changeset with approval by one person then changes-requested by another",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(1)),
			},
			start: daysAgo(1),
			events: []Event{
				ghReview(1, daysAgo(1), "user1", "APPROVED"),
				ghReview(1, daysAgo(0), "user2", "CHANGES_REQUESTED"), // This has higher precedence
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(1), Total: 1, Open: 1, OpenApproved: 1},
				{Time: daysAgo(0), Total: 1, Open: 1, OpenChangesRequested: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "single changeset with approval by one person then changes-requested by another",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(1)),
			},
			start: daysAgo(1),
			events: []Event{
				bbsActivity(1, daysAgo(1), "user1", a8n.ChangesetEventKindBitbucketServerApproved),
				bbsActivity(1, daysAgo(0), "user2", a8n.ChangesetEventKindBitbucketServerReviewed),
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(1), Total: 1, Open: 1, OpenApproved: 1},
				{Time: daysAgo(0), Total: 1, Open: 1, OpenChangesRequested: 1},
			},
		},
		{
			codehosts: "github",
			name:      "single changeset with changes-requested by one person then approval by another",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(1)),
			},
			start: daysAgo(1),
			events: []Event{
				ghReview(1, daysAgo(1), "user1", "CHANGES_REQUESTED"),
				ghReview(1, daysAgo(0), "user2", "APPROVED"),
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(1), Total: 1, Open: 1, OpenChangesRequested: 1},
				{Time: daysAgo(0), Total: 1, Open: 1, OpenChangesRequested: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "single changeset with changes-requested by one person then approval by another",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(1)),
			},
			start: daysAgo(1),
			events: []Event{
				bbsActivity(1, daysAgo(1), "user1", a8n.ChangesetEventKindBitbucketServerReviewed),
				bbsActivity(1, daysAgo(0), "user2", a8n.ChangesetEventKindBitbucketServerApproved),
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(1), Total: 1, Open: 1, OpenChangesRequested: 1},
				{Time: daysAgo(0), Total: 1, Open: 1, OpenChangesRequested: 1},
			},
		},
		{
			codehosts: "github",
			name:      "single changeset with changes-requested by one person, approval by another, then approval by first person",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(2)),
			},
			start: daysAgo(2),
			events: []Event{
				ghReview(1, daysAgo(2), "user1", "CHANGES_REQUESTED"),
				ghReview(1, daysAgo(1), "user2", "APPROVED"),
				ghReview(1, daysAgo(0), "user1", "APPROVED"),
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(2), Total: 1, Open: 1, OpenChangesRequested: 1},
				{Time: daysAgo(1), Total: 1, Open: 1, OpenChangesRequested: 1},
				{Time: daysAgo(0), Total: 1, Open: 1, OpenApproved: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "single changeset with changes-requested by one person, approval by another, then approval by first person",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(2)),
			},
			start: daysAgo(2),
			events: []Event{
				bbsActivity(1, daysAgo(2), "user1", a8n.ChangesetEventKindBitbucketServerReviewed),
				bbsActivity(1, daysAgo(1), "user2", a8n.ChangesetEventKindBitbucketServerApproved),
				bbsActivity(1, daysAgo(0), "user1", a8n.ChangesetEventKindBitbucketServerApproved),
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(2), Total: 1, Open: 1, OpenChangesRequested: 1},
				{Time: daysAgo(1), Total: 1, Open: 1, OpenChangesRequested: 1},
				{Time: daysAgo(0), Total: 1, Open: 1, OpenApproved: 1},
			},
		},
		{
			codehosts: "github",
			name:      "single changeset with approval by one person, changes-requested by another, then changes-requested by first person",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(2)),
			},
			start: daysAgo(2),
			events: []Event{
				ghReview(1, daysAgo(2), "user1", "APPROVED"),
				ghReview(1, daysAgo(1), "user2", "CHANGES_REQUESTED"),
				ghReview(1, daysAgo(0), "user1", "CHANGES_REQUESTED"),
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(2), Total: 1, Open: 1, OpenApproved: 1},
				{Time: daysAgo(1), Total: 1, Open: 1, OpenChangesRequested: 1},
				{Time: daysAgo(0), Total: 1, Open: 1, OpenChangesRequested: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "single changeset with approval by one person, changes-requested by another, then changes-requested by first person",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(2)),
			},
			start: daysAgo(2),
			events: []Event{
				bbsActivity(1, daysAgo(2), "user1", a8n.ChangesetEventKindBitbucketServerApproved),
				bbsActivity(1, daysAgo(1), "user2", a8n.ChangesetEventKindBitbucketServerReviewed),
				bbsActivity(1, daysAgo(0), "user1", a8n.ChangesetEventKindBitbucketServerReviewed),
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(2), Total: 1, Open: 1, OpenApproved: 1},
				{Time: daysAgo(1), Total: 1, Open: 1, OpenChangesRequested: 1},
				{Time: daysAgo(0), Total: 1, Open: 1, OpenChangesRequested: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "single changeset open, approved, unapproved",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(3)),
			},
			start: daysAgo(4),
			events: []Event{
				bbsActivity(1, daysAgo(2), "user1", a8n.ChangesetEventKindBitbucketServerApproved),
				bbsActivity(1, daysAgo(1), "user1", a8n.ChangesetEventKindBitbucketServerUnapproved),
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(4), Total: 0, Open: 0},
				{Time: daysAgo(3), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(2), Total: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
				{Time: daysAgo(1), Total: 1, Open: 1, OpenPending: 1, OpenApproved: 0},
				{Time: daysAgo(0), Total: 1, Open: 1, OpenPending: 1, OpenApproved: 0},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "single changeset open, changes requested, approved, unapproved",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(3)),
			},
			start: daysAgo(4),
			events: []Event{
				bbsActivity(1, daysAgo(2), "user1", a8n.ChangesetEventKindBitbucketServerReviewed),
				bbsActivity(1, daysAgo(1), "user1", a8n.ChangesetEventKindBitbucketServerApproved),
				bbsActivity(1, daysAgo(0), "user1", a8n.ChangesetEventKindBitbucketServerUnapproved),
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(4), Total: 0, Open: 0},
				{Time: daysAgo(3), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(2), Total: 1, Open: 1, OpenPending: 0, OpenChangesRequested: 1},
				{Time: daysAgo(1), Total: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
				{Time: daysAgo(0), Total: 1, Open: 1, OpenPending: 1, OpenApproved: 0},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "single changeset open, approved, unapproved, approved by another person",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(3)),
			},
			start: daysAgo(4),
			events: []Event{
				bbsActivity(1, daysAgo(2), "user1", a8n.ChangesetEventKindBitbucketServerApproved),
				bbsActivity(1, daysAgo(1), "user1", a8n.ChangesetEventKindBitbucketServerUnapproved),
				bbsActivity(1, daysAgo(0), "user2", a8n.ChangesetEventKindBitbucketServerApproved),
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(4), Total: 0, Open: 0},
				{Time: daysAgo(3), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(2), Total: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
				{Time: daysAgo(1), Total: 1, Open: 1, OpenPending: 1, OpenApproved: 0},
				{Time: daysAgo(0), Total: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			name:      "single changeset open, approved, then approved and unapproved by another person",
			changesets: []*a8n.Changeset{
				bbsChangeset(1, daysAgo(3)),
			},
			start: daysAgo(4),
			events: []Event{
				bbsActivity(1, daysAgo(2), "user1", a8n.ChangesetEventKindBitbucketServerApproved),
				bbsActivity(1, daysAgo(1), "user2", a8n.ChangesetEventKindBitbucketServerApproved),
				bbsActivity(1, daysAgo(0), "user2", a8n.ChangesetEventKindBitbucketServerUnapproved),
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(4), Total: 0, Open: 0},
				{Time: daysAgo(3), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(2), Total: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
				{Time: daysAgo(1), Total: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
				{Time: daysAgo(0), Total: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
			},
		},
		{
			codehosts: "github and bitbucketserver",
			name:      "multiple changesets on different code hosts in different review stages before merge",
			changesets: []*a8n.Changeset{
				ghChangeset(1, daysAgo(6)),
				bbsChangeset(1, daysAgo(6)),
				ghChangeset(2, daysAgo(6)),
				bbsChangeset(2, daysAgo(6)),
				ghChangeset(3, daysAgo(6)),
				bbsChangeset(3, daysAgo(6)),
			},
			start: daysAgo(7),
			events: []Event{
				// GitHub Events
				ghReview(1, daysAgo(5), "user1", "APPROVED"),
				fakeEvent{t: daysAgo(3), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
				ghReview(2, daysAgo(4), "user1", "APPROVED"),
				ghReview(2, daysAgo(3), "user2", "APPROVED"),
				fakeEvent{t: daysAgo(2), kind: a8n.ChangesetEventKindGitHubMerged, id: 2},
				ghReview(3, daysAgo(2), "user1", "CHANGES_REQUESTED"),
				ghReview(3, daysAgo(1), "user2", "CHANGES_REQUESTED"),
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 3},
				// Bitbucket Server Events
				bbsActivity(1, daysAgo(5), "user1", a8n.ChangesetEventKindBitbucketServerApproved),
				fakeEvent{t: daysAgo(3), kind: a8n.ChangesetEventKindBitbucketServerMerged, id: 1},
				bbsActivity(2, daysAgo(4), "user1", a8n.ChangesetEventKindBitbucketServerApproved),
				bbsActivity(2, daysAgo(3), "user2", a8n.ChangesetEventKindBitbucketServerApproved),
				fakeEvent{t: daysAgo(2), kind: a8n.ChangesetEventKindBitbucketServerMerged, id: 2},
				bbsActivity(3, daysAgo(2), "user1", a8n.ChangesetEventKindBitbucketServerReviewed),
				bbsActivity(3, daysAgo(1), "user2", a8n.ChangesetEventKindBitbucketServerReviewed),
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindBitbucketServerMerged, id: 3},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(7), Total: 0, Open: 0},
				{Time: daysAgo(6), Total: 6, Open: 6, OpenPending: 6},
				{Time: daysAgo(5), Total: 6, Open: 6, OpenPending: 4, OpenApproved: 2},
				{Time: daysAgo(4), Total: 6, Open: 6, OpenPending: 2, OpenApproved: 4},
				{Time: daysAgo(3), Total: 6, Open: 4, OpenPending: 2, OpenApproved: 2, Merged: 2},
				{Time: daysAgo(2), Total: 6, Open: 2, OpenPending: 0, OpenChangesRequested: 2, Merged: 4},
				{Time: daysAgo(1), Total: 6, Merged: 6},
				{Time: daysAgo(0), Total: 6, Merged: 6},
			},
		},
	}

	for _, tc := range tests {
		if tc.codehosts != "" {
			tc.name = tc.codehosts + "/" + tc.name
		}
		t.Run(tc.name, func(t *testing.T) {
			if tc.end.IsZero() {
				tc.end = now
			}

			have, err := CalcCounts(tc.start, tc.end, tc.changesets, tc.events...)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(have, tc.want) {
				t.Errorf("wrong counts calculated. diff=%s", cmp.Diff(have, tc.want))
			}
		})
	}
}

type fakeEvent struct {
	t    time.Time
	kind a8n.ChangesetEventKind
	id   int64
}

func (e fakeEvent) Timestamp() time.Time         { return e.t }
func (e fakeEvent) Type() a8n.ChangesetEventKind { return e.kind }
func (e fakeEvent) Changeset() int64             { return e.id }

func ghChangeset(id int64, t time.Time) *a8n.Changeset {
	return &a8n.Changeset{ID: id, Metadata: &github.PullRequest{CreatedAt: t}}
}

func bbsChangeset(id int64, t time.Time) *a8n.Changeset {
	return &a8n.Changeset{
		ID:       id,
		Metadata: &bitbucketserver.PullRequest{CreatedDate: timeToUnixMilli(t)},
	}
}

func timeToUnixMilli(t time.Time) int {
	return int(t.UnixNano()) / int(time.Millisecond)
}

func ghReview(id int64, t time.Time, login, state string) *a8n.ChangesetEvent {
	return &a8n.ChangesetEvent{
		ChangesetID: id,
		Kind:        a8n.ChangesetEventKindGitHubReviewed,
		Metadata: &github.PullRequestReview{
			UpdatedAt: t,
			State:     state,
			Author: github.Actor{
				Login: login,
			},
		},
	}
}

func bbsActivity(id int64, t time.Time, username string, kind a8n.ChangesetEventKind) *a8n.ChangesetEvent {
	return &a8n.ChangesetEvent{
		ChangesetID: id,
		Kind:        kind,
		Metadata: &bitbucketserver.Activity{
			CreatedDate: timeToUnixMilli(t),
			User: bitbucketserver.User{
				Name: username,
			},
		},
	}
}
