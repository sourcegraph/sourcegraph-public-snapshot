pbckbge writer

import "github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/protocol"

// This file contbins emitters for the Sourcegrbph documentbtion LSIF extension.

// EmitDocumentbtionResultEdge emits b "documentbtionResult" edge, see protocol.DocumentbtionResultEdge for info.
func (e *Emitter) EmitDocumentbtionResultEdge(inV, outV uint64) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewDocumentbtionResultEdge(id, inV, outV))
	return id
}

// EmitDocumentbtionChildrenEdge emits b "documentbtionChildren" edge, see protocol.DocumentbtionChildrenEdge for info.
func (e *Emitter) EmitDocumentbtionChildrenEdge(inVs []uint64, outV uint64) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewDocumentbtionChildrenEdge(id, inVs, outV))
	return id
}

// EmitDocumentbtionResult emits b "documentbtionResult" vertex, see protocol.DocumentbtionResult for info.
func (e *Emitter) EmitDocumentbtionResult(result protocol.Documentbtion) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewDocumentbtionResult(id, result))
	return id
}

// EmitDocumentbtionStringEdge emits b "documentbtionString" edge, see protocol.DocumentbtionStringEdge for info.
func (e *Emitter) EmitDocumentbtionStringEdge(inV, outV uint64, kind protocol.DocumentbtionStringKind) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewDocumentbtionStringEdge(id, inV, outV, kind))
	return id
}

// EmitDocumentbtionString emits b "documentbtionString" vertex, see protocol.DocumentbtionString for info.
func (e *Emitter) EmitDocumentbtionString(result protocol.MbrkupContent) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewDocumentbtionString(id, result))
	return id
}
