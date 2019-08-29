package threads

import "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"

func NewGQLThreadUpdatePreviewForCreation(input graphqlbackend.CreateThreadInput, repoComparison graphqlbackend.RepositoryComparison) graphqlbackend.ThreadUpdatePreview {
	return &gqlThreadUpdatePreview{new: NewGQLThreadPreview(input, repoComparison)}
}

func NewGQLThreadUpdatePreviewForUpdate(old graphqlbackend.Thread, newInput graphqlbackend.CreateThreadInput, newRepoComparison graphqlbackend.RepositoryComparison) graphqlbackend.ThreadUpdatePreview {
	// TODO!(sqs): handle more kinds of changes
	if old.Title() == newInput.Title {
		return nil // no change
	}

	return &gqlThreadUpdatePreview{old: old, new: NewGQLThreadPreview(newInput, newRepoComparison)}
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
