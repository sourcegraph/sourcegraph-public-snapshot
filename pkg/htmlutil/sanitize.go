package htmlutil

import (
	"bytes"
	"go/doc"

	"github.com/microcosm-cc/bluemonday"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
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

// ComputeDocHTML computes the DocHTML field of the Def
// from its internal Docs field, and sanitizes it.
func ComputeDocHTML(dc *sourcegraph.Def) {
	if len(dc.Docs) < 1 {
		return
	}
	defDoc := dc.Docs[0]
	var docHTML string
	switch defDoc.Format {
	case "text/html":
		docHTML = defDoc.Data
	// TODO "text/x-markdown"
	// TODO "text/x-rst"
	default: // including "text/plain"
		var buf bytes.Buffer
		doc.ToHTML(&buf, defDoc.Data, nil)
		docHTML = buf.String()
	}
	dc.DocHTML = SanitizeForPB(docHTML)
}
