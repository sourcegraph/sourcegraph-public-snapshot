package writer

import (
	"sync/atomic"

	"github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/lsif/protocol"
)

// Emitter creates vertex and edge values and passes them to the underlying
// JSONWriter instance. Use of this struct guarantees that unique identifiers
// are generated for each constructed element.
type Emitter struct {
	writer JSONWriter
	id     uint64
}

func NewEmitter(writer JSONWriter) *Emitter {
	return &Emitter{
		writer: writer,
	}
}

func (e *Emitter) EmitMetaData(root string, info protocol.ToolInfo) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewMetaData(id, root, info))
	return id
}

func (e *Emitter) EmitProject(languageID string) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewProject(id, languageID))
	return id
}

func (e *Emitter) EmitDocument(languageID, path string) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewDocument(id, languageID, "file://"+path))
	return id
}

func (e *Emitter) EmitRange(start, end protocol.Pos) uint64 {
	return e.EmitRangeWithTag(start, end, nil)
}

// EmitRangeWithTag emits a range with a "tag" property describing a symbol.
func (e *Emitter) EmitRangeWithTag(start, end protocol.Pos, tag *protocol.RangeTag) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewRange(id, start, end, tag))
	return id
}

func (e *Emitter) EmitResultSet() uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewResultSet(id))
	return id
}

func (e *Emitter) EmitDocumentSymbolResult(result []*protocol.RangeBasedDocumentSymbol) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewDocumentSymbolResult(id, result))
	return id
}

func (e *Emitter) EmitDocumentSymbolEdge(resultV, docV uint64) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewDocumentSymbolEdge(id, resultV, docV))
	return id
}

func (e *Emitter) EmitHoverResult(contents []protocol.MarkedString) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewHoverResult(id, contents))
	return id
}

func (e *Emitter) EmitTextDocumentHover(outV, inV uint64) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewTextDocumentHover(id, outV, inV))
	return id
}

func (e *Emitter) EmitDefinitionResult() uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewDefinitionResult(id))
	return id
}

func (e *Emitter) EmitTypeDefinitionResult() uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewTypeDefinitionResult(id))
	return id
}

func (e *Emitter) EmitTextDocumentDefinition(outV, inV uint64) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewTextDocumentDefinition(id, outV, inV))
	return id
}

func (e *Emitter) EmitTextDocumentTypeDefinition(outV, inV uint64) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewTextDocumentTypeDefinition(id, outV, inV))
	return id
}

func (e *Emitter) EmitReferenceResult() uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewReferenceResult(id))
	return id
}

func (e *Emitter) EmitTextDocumentReferences(outV, inV uint64) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewTextDocumentReferences(id, outV, inV))
	return id
}

func (e *Emitter) EmitItem(outV uint64, inVs []uint64, docID uint64) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewItem(id, outV, inVs, docID))
	return id
}

func (e *Emitter) EmitItemOfDefinitions(outV uint64, inVs []uint64, docID uint64) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewItemOfDefinitions(id, outV, inVs, docID))
	return id
}

func (e *Emitter) EmitItemOfReferences(outV uint64, inVs []uint64, docID uint64) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewItemOfReferences(id, outV, inVs, docID))
	return id
}

func (e *Emitter) EmitMoniker(kind, scheme, identifier string) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewMoniker(id, kind, scheme, identifier))
	return id
}

func (e *Emitter) EmitMonikerEdge(outV, inV uint64) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewMonikerEdge(id, outV, inV))
	return id
}

func (e *Emitter) EmitPackageInformation(packageName, scheme, version string) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewPackageInformation(id, packageName, scheme, version))
	return id
}

func (e *Emitter) EmitPackageInformationEdge(outV, inV uint64) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewPackageInformationEdge(id, outV, inV))
	return id
}

func (e *Emitter) EmitContains(outV uint64, inVs []uint64) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewContains(id, outV, inVs))
	return id
}

func (e *Emitter) EmitNext(outV, inV uint64) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewNext(id, outV, inV))
	return id
}

func (e *Emitter) NumElements() uint64 {
	return atomic.LoadUint64(&e.id)
}

func (e *Emitter) Flush() error {
	return e.writer.Flush()
}

func (e *Emitter) nextID() uint64 {
	return atomic.AddUint64(&e.id, 1)
}
