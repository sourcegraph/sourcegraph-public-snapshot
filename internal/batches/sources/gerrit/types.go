package gerrit

import (
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
)

// AnnotatedChange adds metadata we need that lives outside the main
// Change type returned by the Gerrit API.
// This type is used as the primary metadata type for Gerrit
// changesets.
type AnnotatedChange struct {
	Change      *gerrit.Change    `json:"change"`
	Reviewers   []gerrit.Reviewer `json:"reviewers"`
	CodeHostURL url.URL           `json:"codeHostURL"`
}
