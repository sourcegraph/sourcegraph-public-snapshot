package reconciler

import (
	"testing"

	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestPublicationStateCalculator(t *testing.T) {
	type want struct {
		published   bool
		draft       bool
		unpublished bool
	}

	for name, tc := range map[string]struct {
		spec batches.PublishedValue
		ui   *btypes.ChangesetUiPublicationState
		want want
	}{
		"unpublished; no ui": {
			spec: batches.PublishedValue{Val: false},
			ui:   nil,
			want: want{false, false, true},
		},
		"draft; no ui": {
			spec: batches.PublishedValue{Val: "draft"},
			ui:   nil,
			want: want{false, true, false},
		},
		"published; no ui": {
			spec: batches.PublishedValue{Val: true},
			ui:   nil,
			want: want{true, false, false},
		},
		"no published value; no ui": {
			spec: batches.PublishedValue{Val: nil},
			ui:   nil,
			want: want{false, false, true},
		},
		"no published value; unpublished ui": {
			spec: batches.PublishedValue{Val: nil},
			ui:   pointers.Ptr(btypes.ChangesetUiPublicationStateUnpublished),
			want: want{false, false, true},
		},
		"no published value; draft ui": {
			spec: batches.PublishedValue{Val: nil},
			ui:   pointers.Ptr(btypes.ChangesetUiPublicationStateDraft),
			want: want{false, true, false},
		},
		"no published value; published ui": {
			spec: batches.PublishedValue{Val: nil},
			ui:   pointers.Ptr(btypes.ChangesetUiPublicationStatePublished),
			want: want{true, false, false},
		},
	} {
		t.Run(name, func(t *testing.T) {
			calc := &publicationStateCalculator{tc.spec, tc.ui}

			if have, want := calc.IsPublished(), tc.want.published; have != want {
				t.Errorf("unexpected IsPublished result: have=%v want=%v", have, want)
			}
			if have, want := calc.IsDraft(), tc.want.draft; have != want {
				t.Errorf("unexpected IsDraft result: have=%v want=%v", have, want)
			}
			if have, want := calc.IsUnpublished(), tc.want.unpublished; have != want {
				t.Errorf("unexpected IsUnpublished result: have=%v want=%v", have, want)
			}
		})
	}
}
