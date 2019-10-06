package threads

import (
	"context"
	"log"
	"reflect"
	"strings"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/gituri"
)

func NewGQLThreadUpdatePreviewForCreation(input graphqlbackend.CreateThreadInput, repoComparison graphqlbackend.RepositoryComparison) graphqlbackend.ThreadUpdatePreview {
	return &gqlThreadUpdatePreview{new: NewGQLThreadPreview(input, repoComparison)}
}

func NewGQLThreadUpdatePreviewForUpdate(ctx context.Context, old graphqlbackend.Thread, newInput graphqlbackend.CreateThreadInput, newRepoComparison graphqlbackend.RepositoryComparison) (graphqlbackend.ThreadUpdatePreview, error) {
	// Determine if the update will actually change the thread.
	//
	// TODO!(sqs): handle more kinds of changes
	var changed bool
	if old.Title() != newInput.Title {
		changed = true
	}
	if !changed {
		oldRepoComparison, err := old.RepositoryComparison(ctx)
		if err != nil {
			return nil, err
		}
		if equal, err := RepoComparisonDiffEqual(ctx, oldRepoComparison, newRepoComparison); err != nil {
			return nil, err
		} else if !equal {
			changed = true
		}
	}

	if changed {
		return &gqlThreadUpdatePreview{old: old, new: NewGQLThreadPreview(newInput, newRepoComparison)}, nil
	}
	return nil, nil
}

func RepoComparisonDiffEqual(ctx context.Context, a, b graphqlbackend.RepositoryComparison) (bool, error) {
	// TODO!(sqs): check all fields
	aDiff, err := a.FileDiffs(&graphqlutil.ConnectionArgs{}).RawDiff(ctx)
	if err != nil {
		return false, err
	}
	bDiff, err := b.FileDiffs(&graphqlutil.ConnectionArgs{}).RawDiff(ctx)
	if err != nil {
		return false, err
	}

	// Treat 2 diffs as equal even if they are to/between different revisions.
	aDiff, err = StripDiffPathPrefixes(aDiff, true)
	if err != nil {
		return false, err
	}
	bDiff, err = StripDiffPathPrefixes(bDiff, false)
	if err != nil {
		return false, err
	}

	// TODO!(sqs): this doesnt always work because 2 diffs can be equivalent in a way that is hard to
	// determine, such as one has more lines of context than the other.
	getChangedLines := func(rawDiff string) []string {
		var l []string
		for _, line := range strings.Split(rawDiff, "\n") {
			if len(line) > 0 && (line[0] == '-' || line[0] == '+') {
				l = append(l, line)
			}
		}
		return l
	}

	log.Printf("======== aDiff\n%s\n\n======== bDiff\n%s\n\n", aDiff, bDiff)

	return reflect.DeepEqual(getChangedLines(aDiff), getChangedLines(bDiff)), nil
}

func StripDiffPathPrefixes(rawDiff string, alsoStripAOrBPrefix bool) (string, error) {
	fileDiffs, err := diff.ParseMultiFileDiff([]byte(rawDiff))
	if err != nil {
		return "", err
	}
	stripPathPrefix := func(uriStr string) string {
		strip := func(s string) string {
			if alsoStripAOrBPrefix {
				s = strings.TrimPrefix(strings.TrimPrefix(s, "b/"), "a/") // HACK TODO!(sqs)
			}
			return s
		}
		u, err := gituri.Parse(uriStr)
		if err != nil {
			return strip(uriStr)
		}
		return strip(u.FilePath())
	}
	for _, fd := range fileDiffs {
		fd.Extended = nil
		fd.OrigName = stripPathPrefix(fd.OrigName)
		fd.NewName = stripPathPrefix(fd.NewName)
	}
	b, err := diff.PrintMultiFileDiff(fileDiffs)
	return string(b), err
}

func NewGQLThreadUpdatePreviewForDeletion(old graphqlbackend.Thread) graphqlbackend.ThreadUpdatePreview {
	return &gqlThreadUpdatePreview{old: old}
}

type gqlThreadUpdatePreview struct {
	old graphqlbackend.Thread
	new graphqlbackend.ThreadPreview
}

func (v *gqlThreadUpdatePreview) OldThread() graphqlbackend.Thread { return v.old }

func (v *gqlThreadUpdatePreview) NewThread() graphqlbackend.ThreadPreview { return v.new }

func (v *gqlThreadUpdatePreview) Operation() graphqlbackend.ThreadUpdateOperation {
	switch {
	case v.old == nil && v.new != nil:
		return graphqlbackend.ThreadUpdateOperationCreation
	case v.old != nil && v.new != nil:
		return graphqlbackend.ThreadUpdateOperationUpdate
	case v.old != nil && v.new == nil:
		return graphqlbackend.ThreadUpdateOperationDeletion
	default:
		panic("unexpected")
	}
}

func (v *gqlThreadUpdatePreview) titleChanged() bool {
	return v.old != nil && v.new != nil && v.old.Title() != v.new.Title()
}

func (v *gqlThreadUpdatePreview) OldTitle() *string {
	if v.titleChanged() {
		return strPtr(v.old.Title())
	}
	return nil
}

func (v *gqlThreadUpdatePreview) NewTitle() *string {
	if v.titleChanged() {
		return strPtr(v.new.Title())
	}
	return nil
}
