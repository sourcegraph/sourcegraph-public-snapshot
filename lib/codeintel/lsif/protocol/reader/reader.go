package reader

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"runtime"
	"strings"
	"sync"

	pb "github.com/golang/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/tools/lsif-flat/proto"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol/writer"
)

type ElementAndErr struct {
	Element Element
	Err     error
}

type LsifFormat int32

const (
	GraphFormat        LsifFormat = 1
	FlatFormat         LsifFormat = 2
	FlatProtobufFormat LsifFormat = 3
)

type Dump struct {
	Format LsifFormat
	Reader io.Reader
}

func DetectFormat(file string) (LsifFormat, error) {
	if strings.HasSuffix(file, ".lsif") {
		return GraphFormat, nil
	} else if strings.HasSuffix(file, ".lsif-flat") {
		return FlatFormat, nil
	} else if strings.HasSuffix(file, ".lsif-flat.pb") {
		return FlatProtobufFormat, nil
	} else {
		return 0, fmt.Errorf("unrecognized format %s, expected one of these: [.lsif, .lsif-flat, .lsif-flat.pb]", file)
	}
}

// Read unmarshals the given LSIF dump one element at a time and sends them back through a channel.
func Read(ctx context.Context, dump Dump) <-chan ElementAndErr {
	switch dump.Format {
	case FlatProtobufFormat:
		ch := make(chan ElementAndErr, ChannelBufferSize)

		bytes, err := ioutil.ReadAll(dump.Reader)
		if err != nil {
			fmt.Println("Read err", err)
			ch <- ElementAndErr{Err: err}
		}
		values := proto.LsifValues{}
		err = pb.Unmarshal(bytes, &values)
		if err != nil {
			fmt.Println("Read err", err)
			ch <- ElementAndErr{Err: err}
		}
		go func() {
			defer close(ch)
			for _, el := range ConvertFlatToGraph(&values) {
				ch <- ElementAndErr{Element: el}
			}
		}()
		return ch
	case FlatFormat:
		// TODO
		ch := make(chan ElementAndErr, ChannelBufferSize)
		close(ch)
		return ch
	case GraphFormat:
		interner := NewInterner()
		return readLines(ctx, dump.Reader, func(line []byte) (Element, error) {
			return unmarshalElement(interner, line)
		})
	default:
		// gohno
		return make(chan ElementAndErr, ChannelBufferSize)
	}
}

// LineBufferSize is the maximum size of the buffer used to read each line of a raw LSIF index. Lines in
// LSIF can get very long as it include escaped hover text (package documentation), as well as large edges
// such as the contains edge of large documents.
//
// This corresponds a 10MB buffer that can accommodate 10 million characters.
const LineBufferSize = 1e7

// ChannelBufferSize is the number sources lines that can be read ahead of the correlator.
const ChannelBufferSize = 512

// NumUnmarshalGoRoutines is the number of goroutines launched to unmarshal individual lines.
var NumUnmarshalGoRoutines = runtime.GOMAXPROCS(0)

// readLines reads the given content as line-separated objects which are unmarshallable by the given function
// and returns a channel of Pair values for each non-empty line.
func readLines(ctx context.Context, r io.Reader, unmarshal func(line []byte) (Element, error)) <-chan ElementAndErr {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	scanner.Buffer(make([]byte, LineBufferSize), LineBufferSize)

	// Pool of buffers used to transfer copies of the scanner slice to unmarshal workers
	pool := sync.Pool{New: func() interface{} { return new(bytes.Buffer) }}

	// Read the document in a separate go-routine.
	lineCh := make(chan *bytes.Buffer, ChannelBufferSize)
	go func() {
		defer close(lineCh)

		for scanner.Scan() {
			if line := scanner.Bytes(); len(line) != 0 {
				buf := pool.Get().(*bytes.Buffer)
				_, _ = buf.Write(line)

				select {
				case lineCh <- buf:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	pairCh := make(chan ElementAndErr, ChannelBufferSize)
	go func() {
		defer close(pairCh)

		// Unmarshal workers receive work assignments as indices into a shared
		// slice and put the result into the same index in a second shared slice.
		work := make(chan int, NumUnmarshalGoRoutines)
		defer close(work)

		// Each unmarshal worker sends a zero-length value on this channel
		// to signal completion of a unit of work.
		signal := make(chan struct{}, NumUnmarshalGoRoutines)
		defer close(signal)

		// The input slice
		lines := make([]*bytes.Buffer, NumUnmarshalGoRoutines)

		// The result slice
		pairs := make([]ElementAndErr, NumUnmarshalGoRoutines)

		for i := 0; i < NumUnmarshalGoRoutines; i++ {
			go func() {
				for idx := range work {
					element, err := unmarshal(lines[idx].Bytes())
					pairs[idx].Element = element
					pairs[idx].Err = err
					signal <- struct{}{}
				}
			}()
		}

		done := false
		for !done {
			i := 0

			// Read a new "batch" of lines from the reader routine and fill the
			// shared array. Each index that receives a new value is queued in
			// the unmarshal worker channel and can be immediately processed.
			for i < NumUnmarshalGoRoutines {
				line, ok := <-lineCh
				if !ok {
					done = true
					break
				}

				lines[i] = line
				work <- i
				i++
			}

			// Wait until the current batch has been completely unmarshalled
			for j := 0; j < i; j++ {
				<-signal
			}

			// Return each buffer to the pool for reuse
			for j := 0; j < i; j++ {
				lines[j].Reset()
				pool.Put(lines[j])
			}

			// Read the result array in order. If the caller context has completed,
			// we'll abandon any additional values we were going to send on this
			// channel (as well as any additional errors from the scanner).
			for j := 0; j < i; j++ {
				select {
				case pairCh <- pairs[j]:
				case <-ctx.Done():
					return
				}
			}
		}

		// If there was an error reading from the source, output it here
		if err := scanner.Err(); err != nil {
			pairCh <- ElementAndErr{Err: err}
		}
	}()

	return pairCh
}

func ConvertFlatToGraph(vals *proto.LsifValues) []Element {
	g := NewGraph()

	g.EmitVertex("metaData", MetaData{Version: "0.1.0", ProjectRoot: "file:///"})

	for _, lsifValue := range vals.Values {
		if value, ok := lsifValue.Value.(*proto.LsifValue_Package); ok {
			g.EmitPackage(value.Package)
		}
	}
	for _, lsifValue := range vals.Values {
		if value, ok := lsifValue.Value.(*proto.LsifValue_Moniker); ok {
			g.EmitMoniker(value.Moniker)
		}
	}
	for _, lsifValue := range vals.Values {
		if value, ok := lsifValue.Value.(*proto.LsifValue_Document); ok {
			g.EmitDocument(value.Document)
		}
	}

	// Implementations
	for _, lsifValue := range vals.Values {
		if moniker, ok := lsifValue.Value.(*proto.LsifValue_Moniker); ok {
			for _, impl := range moniker.Moniker.ImplementationMonikers {
				// Local implementations
				for _, rngeDoc := range g.monikerToDefinition[impl] {
					g.EmitEdge("item", Edge{
						OutV:     g.monikerToResults[moniker.Moniker.Id].ImplementationResult,
						InVs:     []int{rngeDoc.rnge},
						Document: rngeDoc.doc},
					)
				}

				// Remote implementations
				if _, ok := g.monikerToKindToGID[impl]; ok {
					if g_ID, ok := g.monikerToKindToGID[impl]["implementation"]; ok {
						resultSet := g.monikerToResults[moniker.Moniker.Id].ResultSet
						g.EmitEdge("moniker", Edge{OutV: resultSet, InV: g_ID})
					}
				}
			}
		}
	}

	return g.Elements
}

type ResultIDs struct {
	ResultSet            int
	DefinitionResult     int
	ReferenceResult      int
	ImplementationResult int
	HoverResult          int
}

type graph struct {
	ID                  int
	Elements            []Element
	monikerToResults    map[string]ResultIDs
	f2g_package         map[string]int
	monikerToKindToGID  map[string]map[string]int
	monikerToDefinition map[string][]RangeDoc
}

type RangeDoc struct {
	rnge int
	doc  int
}

func NewGraph() graph {
	return graph{
		ID:                  0,
		Elements:            []Element{},
		monikerToResults:    map[string]ResultIDs{},
		f2g_package:         map[string]int{},
		monikerToKindToGID:  map[string]map[string]int{},
		monikerToDefinition: map[string][]RangeDoc{},
	}
}

func (g *graph) EmitPackage(pkg *proto.Package) {
	g.f2g_package[pkg.Id] = g.EmitVertex("packageInformation", PackageInformation{
		Name:    pkg.Name,
		Version: pkg.Version,
	})
}

func (g *graph) EmitMoniker(moniker *proto.Moniker) {
	if moniker.Kind == "" {
		// It's a moniker for a local symbol (i.e. neither imported nor exported).
		return
	}

	if pkgid, ok := g.f2g_package[moniker.PackageId]; ok {
		g_id := g.EmitVertex("moniker", Moniker{
			Kind:       moniker.Kind,
			Scheme:     moniker.Scheme,
			Identifier: moniker.Id,
		})
		g.EmitEdge("packageInformation", Edge{OutV: g_id, InV: pkgid})
		if _, ok := g.monikerToKindToGID[moniker.Id]; !ok {
			g.monikerToKindToGID[moniker.Id] = map[string]int{}
		}
		g.monikerToKindToGID[moniker.Id][moniker.Kind] = g_id
	} else {
		fmt.Printf("when emitting moniker %s: package %v not found", moniker.PackageId, moniker.Id)
	}
}

func (g *graph) EmitDocument(doc *proto.Document) {
	documentID := g.EmitVertex("document", "file:///"+doc.Uri)
	rangeIDs := []int{}
	for _, occ := range doc.Occurrences {
		rangeID := g.EmitRange(occ.Range)
		rangeIDs = append(rangeIDs, rangeID)
		results := g.EmitResults(occ.MonikerId, strings.Join(occ.MarkdownHover, "\n"))
		g.EmitEdge("next", Edge{OutV: rangeID, InV: results.ResultSet})
		switch occ.Role {
		case proto.MonikerOccurrence_ROLE_DEFINITION:
			g.EmitEdge("item", Edge{OutV: results.DefinitionResult, InVs: []int{rangeID}, Document: documentID})
			if _, ok := g.monikerToDefinition[occ.MonikerId]; !ok {
				g.monikerToDefinition[occ.MonikerId] = []RangeDoc{}
			}
			g.monikerToDefinition[occ.MonikerId] = append(g.monikerToDefinition[occ.MonikerId], RangeDoc{rnge: rangeID, doc: documentID})
		case proto.MonikerOccurrence_ROLE_REFERENCE:
			g.EmitEdge("item", Edge{OutV: results.ReferenceResult, InVs: []int{rangeID}, Document: documentID})
		default:
		}
	}
	g.EmitEdge("contains", Edge{OutV: documentID, InVs: rangeIDs})
}

func (g *graph) EmitResults(moniker string, hover string) ResultIDs {
	if ids, ok := g.monikerToResults[moniker]; ok {
		return ids
	}

	ids := ResultIDs{
		ResultSet:            g.EmitVertex("resultSet", ResultSet{}),
		DefinitionResult:     g.EmitVertex("definitionResult", nil),
		ReferenceResult:      g.EmitVertex("referenceResult", nil),
		ImplementationResult: g.EmitVertex("implementationResult", nil),
		HoverResult:          g.EmitVertex("hoverResult", hover),
	}
	g.EmitEdge("textDocument/definition", Edge{OutV: ids.ResultSet, InV: ids.DefinitionResult})
	g.EmitEdge("textDocument/references", Edge{OutV: ids.ResultSet, InV: ids.ReferenceResult})
	g.EmitEdge("textDocument/implementation", Edge{OutV: ids.ResultSet, InV: ids.ImplementationResult})
	g.EmitEdge("textDocument/hover", Edge{OutV: ids.ResultSet, InV: ids.HoverResult})

	if _, ok := g.monikerToKindToGID[moniker]; ok {
		if g_ID, ok := g.monikerToKindToGID[moniker]["import"]; ok {
			g.EmitEdge("moniker", Edge{OutV: ids.ResultSet, InV: g_ID})
		}
		if g_ID, ok := g.monikerToKindToGID[moniker]["export"]; ok {
			g.EmitEdge("moniker", Edge{OutV: ids.ResultSet, InV: g_ID})
		}
	}

	g.monikerToResults[moniker] = ids

	return ids
}

func (g *graph) EmitRange(rnge *proto.Range) int {
	return g.Emit("vertex", "range", Range{
		RangeData: protocol.RangeData{
			Start: protocol.Pos{
				Line:      int(rnge.Start.Line),
				Character: int(rnge.Start.Character),
			},
			End: protocol.Pos{
				Line:      int(rnge.End.Line),
				Character: int(rnge.End.Character),
			},
		},
	})
}

func (g *graph) EmitVertex(label string, payload interface{}) int {
	return g.Emit("vertex", label, payload)
}

func (g *graph) EmitEdge(label string, payload Edge) int {
	return g.Emit("edge", label, payload)
}

func (g *graph) Emit(ty, label string, payload interface{}) int {
	g.ID++
	g.Elements = append(g.Elements, Element{
		ID:      g.ID,
		Type:    ty,
		Label:   label,
		Payload: payload,
	})
	return g.ID
}

func WriteNDJSON(elements []interface{}, out io.Writer) {
	w := writer.NewJSONWriter(out)
	for _, e := range elements {
		w.Write(e)
	}
	w.Flush()
}

func ElementsToEmptyInterfaces(els []Element) []interface{} {
	r := []interface{}{}
	for _, el := range els {
		r = append(r, el)
	}
	return r
}
