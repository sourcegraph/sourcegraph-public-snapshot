package validation

import (
	"net/url"
	"strings"

	protocolReader "github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol/reader"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/reader"
)

// validateMetaDataVertex ensures that the given metadata vertex has a valid project root. The
// project root is stashed in the validation context for use by validateDocumentVertex.
func validateMetaDataVertex(ctx *ValidationContext, lineContext reader.LineContext) bool {
	if ctx.ProjectRoot != nil {
		ctx.AddError("metaData defined multiple times").AddContext(lineContext)
	}

	metaData, ok := lineContext.Element.Payload.(protocolReader.MetaData)
	if !ok {
		ctx.AddError("illegal payload").AddContext(lineContext)
		return false
	}

	projectRootURL, err := url.Parse(metaData.ProjectRoot)
	if err != nil {
		ctx.AddError("project root is not a valid URL").AddContext(lineContext)
		return false
	}
	if projectRootURL.Scheme == "" {
		ctx.AddError("project root is not a valid URL").AddContext(lineContext)
		return false
	}

	ctx.ProjectRoot = projectRootURL
	return true
}

// validateMetaDataVertex ensures that the given document vertex has a valid URI which is
// relative to the project root.
func validateDocumentVertex(ctx *ValidationContext, lineContext reader.LineContext) bool {
	uri, ok := lineContext.Element.Payload.(string)
	if !ok {
		ctx.AddError("illegal payload").AddContext(lineContext)
		return false
	}

	documentUrl, err := url.Parse(uri)
	if err != nil {
		ctx.AddError("document uri is not a valid URL").AddContext(lineContext)
		return false
	}
	if documentUrl.Scheme == "" {
		ctx.AddError("document uri is not a valid URL").AddContext(lineContext)
		return false
	}

	if ctx.ProjectRoot != nil && !strings.HasPrefix(documentUrl.String(), ctx.ProjectRoot.String()) {
		ctx.AddError("document is not relative to project root").AddContext(lineContext)
		return false
	}

	return true
}

// validateRangeVertex ensures that the given range vertex has valid bounds and extents.
func validateRangeVertex(ctx *ValidationContext, lineContext reader.LineContext) bool {
	r, ok := lineContext.Element.Payload.(protocolReader.Range)
	if !ok {
		ctx.AddError("illegal payload").AddContext(lineContext)
		return false
	}

	if r.Start.Line < 0 || r.Start.Character < 0 || r.End.Line < 0 || r.End.Character < 0 {
		ctx.AddError("illegal range bounds").AddContext(lineContext)
		return false
	}

	if r.Start.Line > r.End.Line {
		ctx.AddError("illegal range extents").AddContext(lineContext)
		return false
	}
	if r.Start.Line == r.End.Line && r.Start.Character > r.End.Character {
		ctx.AddError("illegal range extents").AddContext(lineContext)
		return false
	}

	return true
}
