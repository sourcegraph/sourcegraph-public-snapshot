package threads

import (
	"context"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/pkg/gituri"
)

func NewGQLThreadUpdatePreviewForCreation(input graphqlbackend.CreateThreadInput, repoComparison graphqlbackend.RepositoryComparison) graphqlbackend.ThreadUpdatePreview {
	return &gqlThreadUpdatePreview{new: NewGQLThreadPreview(input, repoComparison)}
}

func NewGQLThreadUpdatePreviewForUpdate(ctx context.Context, old graphqlbackend.Thread, newInput graphqlbackend.CreateThreadInput, newRepoComparison graphqlbackend.RepositoryComparison) (graphqlbackend.ThreadUpdatePreview, error) {
	// Determine if the update will actually change the thread.
	//
	// TODO!(sqs): handle more kinds of changes
	var changed bool
	if old.Title() == newInput.Title {
		changed = true
	}
	if !changed {
		oldRepoComparison, err := old.RepositoryComparison(ctx)
		if err != nil {
			return nil, err
		}
		if equal, err := repoComparisonDiffEqual(ctx, oldRepoComparison, newRepoComparison); err != nil {
			return nil, err
		} else if !equal {
			changed = true
		}
	}

	return &gqlThreadUpdatePreview{old: old, new: NewGQLThreadPreview(newInput, newRepoComparison)}, nil
}

func repoComparisonDiffEqual(ctx context.Context, a, b graphqlbackend.RepositoryComparison) (bool, error) {
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
	aDiff, err = StripDiffPathPrefixes(aDiff)
	if err != nil {
		return false, err
	}
	bDiff, err = StripDiffPathPrefixes(bDiff)
	if err != nil {
		return false, err
	}

	return aDiff == bDiff, nil
}

// TODO!(sqs): this doesnt work because 2 diffs can be equivalent in a way that is hard to
// determine, such as one has more lines of context than the other.
func StripDiffPathPrefixes(rawDiff string) (string, error) {
	fileDiffs, err := diff.ParseMultiFileDiff([]byte(rawDiff))
	if err != nil {
		return "", err
	}
	stripPathPrefix := func(uriStr string) string {
		u, err := gituri.Parse(uriStr)
		if err != nil {
			return uriStr
		}
		return u.FilePath()
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
