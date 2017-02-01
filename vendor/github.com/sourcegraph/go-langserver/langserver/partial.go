package langserver

import "github.com/sourcegraph/go-langserver/pkg/lsp"

// This file contains types and functions related to $/partialResult streaming

// RewriteURIer is a type that implements RewriteURI. The typical use of
// RewriteURI is a build server which translates workspace URIs into URIs for
// other systems to consume.
type RewriteURIer interface {
	// RewriteURI will update all URIs in the type using the rewrite
	// function.
	RewriteURI(rewrite func(string) string)
}

// referenceAddOp is a JSON Patch operation used by
// textDocument/references. The only other patch operation is to create the
// empty location list.
type referenceAddOp struct {
	// OP should always be "add"
	OP    string       `json:"op"`
	Path  string       `json:"path"`
	Value lsp.Location `json:"value"`
}

type referencePatch []referenceAddOp

func (p referencePatch) RewriteURI(rewrite func(string) string) {
	for i := range p {
		p[i].Value.URI = rewrite(p[i].Value.URI)
	}
}

// xreferenceAddOp is a JSON Patch operation used by
// workspace/xreferences. The only other patch operation is to create the
// empty location list.
type xreferenceAddOp struct {
	// OP should always be "add"
	OP    string               `json:"op"`
	Path  string               `json:"path"`
	Value referenceInformation `json:"value"`
}

type xreferencePatch []xreferenceAddOp

func (p xreferencePatch) RewriteURI(rewrite func(string) string) {
	for i := range p {
		p[i].Value.Reference.URI = rewrite(p[i].Value.Reference.URI)
	}
}
