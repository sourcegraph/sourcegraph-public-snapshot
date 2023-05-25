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
	*gerrit.Change
	CodeHostURL *url.URL
}
