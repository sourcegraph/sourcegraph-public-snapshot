package htmlutil

import (
	"github.com/microcosm-cc/bluemonday"
	"sourcegraph.com/sqs/pbtypes"
)

var sanitizePolicy = bluemonday.UGCPolicy().AllowAttrs("class").Matching(bluemonday.SpaceSeparatedTokens).Globally()

// EmptyForPB returns the type *pbtypes.HTML with an empty HTML string.
func EmptyForPB() *pbtypes.HTML {
	return &pbtypes.HTML{}
}

// SanitizeForPB makes sure that the given HTML code is safe against XSS attacks
// and returns it wrapped in the *pbtypes.HTML type to designate that this is
// safe HTML. *pbtypes.HTML should always be created via this package.
func SanitizeForPB(html string) *pbtypes.HTML {
	return &pbtypes.HTML{
		HTML: sanitizePolicy.Sanitize(html),
	}
}
