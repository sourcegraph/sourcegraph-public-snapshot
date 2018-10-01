package graphqlbackend

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func TestDiscussionSelectionRelativeTo(t *testing.T) {
	i32 := func(i int32) *int32 {
		return &i
	}
	tests := []struct {
		name         string
		oldSelection *types.DiscussionThreadTargetRepo
		newContent   string
		want         *discussionSelectionRangeResolver
	}{
		{
			name: "added_content_before",
			oldSelection: &types.DiscussionThreadTargetRepo{
				StartLine: i32(3), StartCharacter: i32(0), EndLine: i32(4), EndCharacter: i32(1),
				LinesBefore: &[]string{"0", "1", "2"},
				Lines:       &[]string{"3"},
				LinesAfter:  &[]string{"4", "5", "6"},
			},
			newContent: "a\nb\nc\n0\n1\n2\n3\n4\n5\n6",
			want:       &discussionSelectionRangeResolver{startLine: 6, startCharacter: 0, endLine: 7, endCharacter: 1},
		},
		{
			name: "added_content_after",
			oldSelection: &types.DiscussionThreadTargetRepo{
				StartLine: i32(3), StartCharacter: i32(0), EndLine: i32(4), EndCharacter: i32(1),
				LinesBefore: &[]string{"0", "1", "2"},
				Lines:       &[]string{"3"},
				LinesAfter:  &[]string{"4", "5", "6"},
			},
			newContent: "0\n1\n2\n3\n4\n5\n6\na\nb\nc",
			want:       &discussionSelectionRangeResolver{startLine: 3, startCharacter: 0, endLine: 4, endCharacter: 1},
		},
		{
			name: "added_content_before_and_after",
			oldSelection: &types.DiscussionThreadTargetRepo{
				StartLine: i32(3), StartCharacter: i32(0), EndLine: i32(4), EndCharacter: i32(1),
				LinesBefore: &[]string{"0", "1", "2"},
				Lines:       &[]string{"3"},
				LinesAfter:  &[]string{"4", "5", "6"},
			},
			newContent: "a\nb\nc\n0\n1\n2\n3\n4\n5\n6\na\nb\nc",
			want:       &discussionSelectionRangeResolver{startLine: 6, startCharacter: 0, endLine: 7, endCharacter: 1},
		},
		{
			name: "removed_content_before",
			oldSelection: &types.DiscussionThreadTargetRepo{
				StartLine: i32(3), StartCharacter: i32(0), EndLine: i32(4), EndCharacter: i32(1),
				LinesBefore: &[]string{"0", "1", "2"},
				Lines:       &[]string{"3"},
				LinesAfter:  &[]string{"4", "5", "6"},
			},
			newContent: "3\n4\n5\n6",
			want:       &discussionSelectionRangeResolver{startLine: 0, startCharacter: 0, endLine: 1, endCharacter: 1},
		},
		{
			name: "removed_content_after",
			oldSelection: &types.DiscussionThreadTargetRepo{
				StartLine: i32(3), StartCharacter: i32(0), EndLine: i32(4), EndCharacter: i32(1),
				LinesBefore: &[]string{"0", "1", "2"},
				Lines:       &[]string{"3"},
				LinesAfter:  &[]string{"4", "5", "6"},
			},
			newContent: "0\n1\n2\n3\n",
			want:       &discussionSelectionRangeResolver{startLine: 0, startCharacter: 0, endLine: 1, endCharacter: 1},
		},
		{
			name: "no_match",
			oldSelection: &types.DiscussionThreadTargetRepo{
				StartLine: i32(3), StartCharacter: i32(0), EndLine: i32(4), EndCharacter: i32(1),
				LinesBefore: &[]string{"0", "1", "2"},
				Lines:       &[]string{"3"},
				LinesAfter:  &[]string{"4", "5", "6"},
			},
			newContent: "0\n2\n3\n1\n",
			want:       nil,
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			got := discussionSelectionRelativeTo(tst.oldSelection, tst.newContent)
			if !reflect.DeepEqual(got, tst.want) {
				t.Logf("got  %+v\n", got)
				t.Fatalf("want %+v\n", tst.want)
			}
		})
	}
}
