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
	StandardFormat     LsifFormat = 1
	FlatFormat         LsifFormat = 2
	FlatProtobufFormat LsifFormat = 3
)

type Dump struct {
	Format LsifFormat
	Reader io.Reader
}

func DetectFormat(file string) (LsifFormat, error) {
	if strings.HasSuffix(file, ".lsif") {
		return StandardFormat, nil
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
	case StandardFormat:
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

type ResultIDs struct {
	ResultSet        int
	DefinitionResult int
	ReferenceResult  int
}

type graph struct {
	ID       int
	Elements []Element
	idCache  map[string]ResultIDs
}

func (g *graph) ResultIDs(moniker string) ResultIDs {
	ids, ok := g.idCache[moniker]
	if !ok {
		ids = ResultIDs{
			ResultSet:        g.AddVertex("resultSet", ResultSet{}),
			DefinitionResult: g.AddVertex("definitionResult", nil),
			ReferenceResult:  g.AddVertex("referenceResult", nil),
		}
		g.AddEdge("textDocument/definition", Edge{OutV: ids.ResultSet, InV: ids.DefinitionResult})
		g.AddEdge("textDocument/references", Edge{OutV: ids.ResultSet, InV: ids.ReferenceResult})
		g.idCache[moniker] = ids
	}
	return ids
}

func (g *graph) Add(ty, label string, payload interface{}) int {
	g.ID++
	g.Elements = append(g.Elements, Element{
		ID:      g.ID,
		Type:    ty,
		Label:   label,
		Payload: payload,
	})
	return g.ID
}

func (g *graph) AddVertex(label string, payload interface{}) int {
	return g.Add("vertex", label, payload)
}

func (g *graph) AddEdge(label string, payload Edge) int {
	return g.Add("edge", label, payload)
}

func (g *graph) AddPackage(doc *proto.Package) {}

func (g *graph) AddDocument(doc *proto.Document) {
	documentID := g.AddVertex("document", "file:///"+doc.Uri)
	rangeIDs := []int{}
	for _, occ := range doc.Occurrences {
		rangeID := g.AddVertex("range", Range{
			RangeData: protocol.RangeData{
				Start: protocol.Pos{
					Line:      int(occ.Range.Start.Line),
					Character: int(occ.Range.Start.Character),
				},
				End: protocol.Pos{
					Line:      int(occ.Range.End.Line),
					Character: int(occ.Range.End.Character),
				},
			},
		})
		rangeIDs = append(rangeIDs, rangeID)
		ids := g.ResultIDs(occ.MonikerId)
		g.AddEdge("next", Edge{OutV: rangeID, InV: ids.ResultSet})
		switch occ.Role {
		case proto.MonikerOccurrence_ROLE_DEFINITION:
			g.AddEdge("item", Edge{OutV: ids.DefinitionResult, InVs: []int{rangeID}, Document: documentID})
		case proto.MonikerOccurrence_ROLE_REFERENCE:
			g.AddEdge("item", Edge{OutV: ids.ReferenceResult, InVs: []int{rangeID}, Document: documentID})
		default:
		}
	}
	g.AddEdge("contains", Edge{OutV: documentID, InVs: rangeIDs})
}

func (g *graph) AddMoniker(doc *proto.Moniker) {}

func WriteNDJSON(elements []interface{}, out io.Writer) {
	w := writer.NewJSONWriter(out)
	for _, e := range elements {
		w.Write(e)
	}
	w.Flush()
}

func ConvertFlatToGraph(vals *proto.LsifValues) []Element {
	g := graph{ID: 0, Elements: []Element{}, idCache: map[string]ResultIDs{}}
	g.AddVertex(
		"metaData",
		MetaData{
			Version:     "0.1.0",
			ProjectRoot: "file:///",
		},
	)
	for _, lsifValue := range vals.Values {
		switch value := lsifValue.Value.(type) {
		case *proto.LsifValue_Package:
			g.AddPackage(value.Package)
		case *proto.LsifValue_Document:
			g.AddDocument(value.Document)
		case *proto.LsifValue_Moniker:
			g.AddMoniker(value.Moniker)
		default:
		}

	}
	return g.Elements
}

func ElementsToEmptyInterfaces(els []Element) []interface{} {
	r := []interface{}{}
	for _, el := range els {
		r = append(r, el)
	}
	return r
}
