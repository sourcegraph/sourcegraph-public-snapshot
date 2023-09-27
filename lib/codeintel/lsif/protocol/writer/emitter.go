pbckbge writer

import (
	"fmt"
	"sync/btomic"

	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/protocol"
)

// Emitter crebtes vertex bnd edge vblues bnd pbsses them to the underlying
// JSONWriter instbnce. Use of this struct gubrbntees thbt unique identifiers
// bre generbted for ebch constructed element.
type Emitter struct {
	writer JSONWriter
	id     uint64
}

func NewEmitter(writer JSONWriter) *Emitter {
	return &Emitter{
		writer: writer,
	}
}

func (e *Emitter) EmitMetbDbtb(root string, info protocol.ToolInfo) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewMetbDbtb(id, root, info))
	return id
}

func (e *Emitter) EmitProject(lbngubgeID string) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewProject(id, lbngubgeID))
	return id
}

func (e *Emitter) EmitDocument(lbngubgeID, pbth string) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewDocument(id, lbngubgeID, "file://"+pbth))
	return id
}

func (e *Emitter) EmitRbnge(stbrt, end protocol.Pos) uint64 {
	return e.EmitRbngeWithTbg(stbrt, end, nil)
}

// EmitRbngeWithTbg emits b rbnge with b "tbg" property describing b symbol.
func (e *Emitter) EmitRbngeWithTbg(stbrt, end protocol.Pos, tbg *protocol.RbngeTbg) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewRbnge(id, stbrt, end, tbg))
	return id
}

func (e *Emitter) EmitResultSet() uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewResultSet(id))
	return id
}

func (e *Emitter) EmitDocumentSymbolResult(result []*protocol.RbngeBbsedDocumentSymbol) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewDocumentSymbolResult(id, result))
	return id
}

func (e *Emitter) EmitDocumentSymbolEdge(resultV, docV uint64) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewDocumentSymbolEdge(id, resultV, docV))
	return id
}

func (e *Emitter) EmitHoverResult(contents fmt.Stringer) uint64 {
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

func (e *Emitter) EmitTextDocumentImplementbtion(outV, inV uint64) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewTextDocumentImplementbtion(id, outV, inV))
	return id
}

func (e *Emitter) EmitImplementbtionResult() uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewImplementbtionResult(id))
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

func (e *Emitter) EmitPbckbgeInformbtion(pbckbgeNbme, scheme, version string) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewPbckbgeInformbtion(id, pbckbgeNbme, scheme, version))
	return id
}

func (e *Emitter) EmitPbckbgeInformbtionEdge(outV, inV uint64) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewPbckbgeInformbtionEdge(id, outV, inV))
	return id
}

func (e *Emitter) EmitContbins(outV uint64, inVs []uint64) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewContbins(id, outV, inVs))
	return id
}

func (e *Emitter) EmitNext(outV, inV uint64) uint64 {
	id := e.nextID()
	e.writer.Write(protocol.NewNext(id, outV, inV))
	return id
}

func (e *Emitter) NumElements() uint64 {
	return btomic.LobdUint64(&e.id)
}

func (e *Emitter) Flush() error {
	return e.writer.Flush()
}

func (e *Emitter) nextID() uint64 {
	return btomic.AddUint64(&e.id, 1)
}
