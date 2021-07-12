package writer

import "github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol"

// This file contains emitters for the Sourcegraph documentation LSIF extension.

// EmitDocumentationResultEdge emits a "documentationResult" edge, see protocol.DocumentationResultEdge for info.
func (e *Emitter) EmitDocumentationResultEdge(inV, outV uint64) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewDocumentationResultEdge(id, inV, outV))
	return id
}

// EmitDocumentationChildrenEdge emits a "documentationChildren" edge, see protocol.DocumentationChildrenEdge for info.
func (e *Emitter) EmitDocumentationChildrenEdge(inVs []uint64, outV uint64) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewDocumentationChildrenEdge(id, inVs, outV))
	return id
}

// EmitDocumentationResult emits a "documentationResult" vertex, see protocol.DocumentationResult for info.
func (e *Emitter) EmitDocumentationResult(result protocol.Documentation) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewDocumentationResult(id, result))
	return id
}

// EmitDocumentationStringEdge emits a "documentationString" edge, see protocol.DocumentationStringEdge for info.
func (e *Emitter) EmitDocumentationStringEdge(inV, outV uint64, kind protocol.DocumentationStringKind) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewDocumentationStringEdge(id, inV, outV, kind))
	return id
}

// EmitDocumentationString emits a "documentationString" vertex, see protocol.DocumentationString for info.
func (e *Emitter) EmitDocumentationString(result protocol.MarkupContent) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewDocumentationString(id, result))
	return id
}
