package a8n

import (
	"flag"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

var update = flag.Bool("update", false, "update testdata")

func TestCalcCounts(t *testing.T) {
	ghChangesetCreated := parse(t, "2019-10-02T14:49:31Z")
	ghChangesetMerged := parse(t, "2019-10-07T13:13:45Z")

	tests := []struct {
		name       string
		start      time.Time
		end        time.Time
		changesets []*a8n.Changeset
		events     []Event
		want       []*ChangesetCounts
	}{
		{
			name: "single github changeset",
			// We start exactly one day earlier
			start: ghChangesetCreated.Add(-24 * time.Hour),
			end:   ghChangesetMerged,
			changesets: []*a8n.Changeset{
				{
					Metadata: &github.PullRequest{CreatedAt: ghChangesetCreated},
				},
			},
			events: []Event{
				fakeEvent{
					kind:        a8n.ChangesetEventKindGitHubClosed,
					t:           parse(t, "2019-10-03T14:02:51Z"),
					changesetID: 1,
				},
				fakeEvent{
					kind:        a8n.ChangesetEventKindGitHubReopened,
					t:           parse(t, "2019-10-03T14:02:54Z"),
					changesetID: 1,
				},
				fakeEvent{
					kind:        a8n.ChangesetEventKindGitHubMerged,
					t:           parse(t, "2019-10-07T13:13:44Z"),
					changesetID: 1,
				},
				fakeEvent{
					kind:        a8n.ChangesetEventKindGitHubClosed,
					t:           parse(t, "2019-10-07T13:13:44Z"),
					changesetID: 1,
				},
			},
			want: []*ChangesetCounts{
				{
					Time:  ghChangesetMerged.Add(5 * -24 * time.Hour),
					Total: 0,
					Open:  0,
				},
				{
					Time:  ghChangesetMerged.Add(4 * -24 * time.Hour),
					Total: 1,
					Open:  1,
				},
				{
					Time:  ghChangesetMerged.Add(3 * -24 * time.Hour),
					Total: 1,
					Open:  1,
				},
				{
					Time:  ghChangesetMerged.Add(2 * -24 * time.Hour),
					Total: 1,
					Open:  1,
				},
				{
					Time:  ghChangesetMerged.Add(1 * -24 * time.Hour),
					Total: 1,
					Open:  1,
				},
				{
					Time:   ghChangesetMerged,
					Total:  1,
					Closed: 1,
					Merged: 1,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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

func parse(t testing.TB, ts string) time.Time {
	t.Helper()

	timestamp, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		t.Fatal(err)
	}

	return timestamp
}

type fakeEvent struct {
	t           time.Time
	kind        a8n.ChangesetEventKind
	changesetID int64
}

func (e fakeEvent) Timestamp() time.Time         { return e.t }
func (e fakeEvent) Type() a8n.ChangesetEventKind { return e.kind }
func (e fakeEvent) Changeset() int64             { return e.changesetID }
