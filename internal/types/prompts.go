package types

import "time"

// Prompt represents a prompt.
type Prompt struct {
	ID               int32     // the globally unique DB ID
	Name             string    // the prompt name, unique among the owner's prompts
	Description      string    // a human-readable description of the prompt
	DefinitionText   string    // the prompt template text
	Draft            bool      // whether the prompt is a draft
	Owner            Namespace // the prompt's owner
	VisibilitySecret bool      // the prompt visibility state (if false, public)

	CreatedAt     time.Time // when this prompt was created
	CreatedByUser *int32    // the user that created this prompt
	UpdatedAt     time.Time // when this prompt was last updated
	UpdatedByUser *int32    // the user that last updated this prompt

	NameWithOwner string // only set when when getting and listing (not creating/inserting)
}
