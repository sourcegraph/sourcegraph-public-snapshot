package campaigns

import (
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	cmpgn "github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

func TestChangesetEventsLabels(t *testing.T) {
	now := time.Now()
	labelEvent := func(name string, kind cmpgn.ChangesetEventKind, when time.Time) *cmpgn.ChangesetEvent {
		removed := kind == cmpgn.ChangesetEventKindGitHubUnlabeled
		return &cmpgn.ChangesetEvent{
			Kind:      kind,
			UpdatedAt: when,
			Metadata: &github.LabelEvent{
				Actor: github.Actor{},
				Label: github.Label{
					Name: name,
				},
				CreatedAt: when,
				Removed:   removed,
			},
		}
	}
	changeset := func(names []string, updated time.Time) *cmpgn.Changeset {
		meta := &github.PullRequest{}
		for _, name := range names {
			meta.Labels.Nodes = append(meta.Labels.Nodes, github.Label{
				Name: name,
			})
		}
		return &cmpgn.Changeset{
			UpdatedAt: updated,
			Metadata:  meta,
		}
	}
	labels := func(names ...string) []cmpgn.ChangesetLabel {
		var ls []cmpgn.ChangesetLabel
		for _, name := range names {
			ls = append(ls, cmpgn.ChangesetLabel{Name: name})
		}
		return ls
	}

	tests := []struct {
		name      string
		changeset *cmpgn.Changeset
		events    ChangesetEvents
		want      []cmpgn.ChangesetLabel
	}{
		{
			name: "zero values",
		},
		{
			name:      "no events",
			changeset: changeset([]string{"label1"}, time.Time{}),
			events:    ChangesetEvents{},
			want:      labels("label1"),
		},
		{
			name:      "remove event",
			changeset: changeset([]string{"label1"}, time.Time{}),
			events: ChangesetEvents{
				labelEvent("label1", cmpgn.ChangesetEventKindGitHubUnlabeled, now),
			},
			want: []cmpgn.ChangesetLabel{},
		},
		{
			name:      "add event",
			changeset: changeset([]string{"label1"}, time.Time{}),
			events: ChangesetEvents{
				labelEvent("label2", cmpgn.ChangesetEventKindGitHubLabeled, now),
			},
			want: labels("label1", "label2"),
		},
		{
			name:      "old add event",
			changeset: changeset([]string{"label1"}, now.Add(5*time.Minute)),
			events: ChangesetEvents{
				labelEvent("label2", cmpgn.ChangesetEventKindGitHubLabeled, now),
			},
			want: labels("label1"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			have := tc.events.UpdateLabelsSince(tc.changeset)
			want := tc.want
			sort.Slice(have, func(i, j int) bool { return have[i].Name < have[j].Name })
			sort.Slice(want, func(i, j int) bool { return want[i].Name < want[j].Name })
			if diff := cmp.Diff(have, want, cmpopts.EquateEmpty()); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
