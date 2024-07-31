package chunkers

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/bash"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/kotlin"
	"github.com/smacker/go-tree-sitter/php"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/scala"
	"github.com/smacker/go-tree-sitter/sql"
	"github.com/smacker/go-tree-sitter/swift"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"

	ctxt "github.com/sourcegraph/sourcegraph/internal/codeintel/context"
)

type TreeSitterChunker struct {
	chunkOptions *ChunkOptions
}

func NewTreeSitterChunker(chunkOptions *ChunkOptions) *TreeSitterChunker {
	if chunkOptions == nil {
		chunkOptions = &defaultChunkOptions
	}
	return &TreeSitterChunker{chunkOptions: chunkOptions}
}

func (tsc *TreeSitterChunker) Chunk(content, filename string) []ctxt.EmbeddableChunk {
	chunks, err := tsc.doChunk(content, filename)
	if err != nil {
		// Fall back to classic chunker in case of error
		return NewClassicChunker(tsc.chunkOptions).Chunk(content, filename)
	}
	return chunks
}

func (tsc *TreeSitterChunker) doChunk(content, filename string) ([]ctxt.EmbeddableChunk, error) {
	parser := sitter.NewParser()
	language := detectLanguage(filename)
	if language == nil {
		return nil, errors.New("no tree-sitter language detected")
	}
	parser.SetLanguage(language)
	tree, err := parser.ParseCtx(context.Background(), nil, []byte(content))
	if err != nil {
		return nil, err
	}

	if tree.RootNode().ChildCount() == 0 || tree.RootNode().Child(0).HasError() {
		return nil, errors.New("error parsing with tree-sitter")
	}

	// Chunk the code
	spans := tsc.chunkNode(tree.RootNode())

	// Fill gaps
	for i := 0; i < len(spans)-1; i++ {
		spans[i] = Span{
			start: spans[i].start,
			end:   spans[i+1].start,
		}
	}

	// Combine small spans into bigger ones
	newSpans := make([]Span, 0)
	curSpan := Span{0, 0}
	for _, span := range spans {
		curSpan = curSpan.Add(span)

		if int(curSpan.Len()) > tsc.chunkOptions.CoalesceThreshold {
			newSpans = append(newSpans, curSpan)
			curSpan = Span{span.end, span.end}
		}
	}
	if curSpan.Len() > 0 {
		newSpans = append(newSpans, curSpan)
	}

	spans = newSpans

	// Transform spans into chunks and filter out empty ones
	chunks := make([]ctxt.EmbeddableChunk, 0)
	line := 1

	for _, span := range spans {
		if span.Len() < 1 {
			continue
		}

		chunk := content[span.start:span.end]
		numLinesInChunk := strings.Count(chunk, "\n")
		chunks = append(chunks, ctxt.EmbeddableChunk{
			FileName:  filename,
			StartLine: line,
			EndLine:   line + numLinesInChunk,
			Content:   chunk,
		})
		line += numLinesInChunk
	}

	return chunks, nil
}

type Span struct {
	start uint32
	end   uint32
}

func (s Span) Len() uint32 {
	return s.end - s.start
}

func (s Span) Add(s2 Span) Span {
	return Span{s.start, s2.end}
}

func (tsc *TreeSitterChunker) chunkNode(node *sitter.Node) []Span {
	chunks := make([]Span, 0)
	curSpan := Span{start: node.StartByte(), end: node.EndByte()}

	maxBytes := uint32(tsc.chunkOptions.ChunkTokensThreshold * ctxt.CHARS_PER_TOKEN)

	for idx := range node.ChildCount() {
		child := node.Child(int(idx))
		if child.EndByte()-child.StartByte() > maxBytes {
			// Recursively chunk the child
			chunks = append(chunks, curSpan)
			curSpan = Span{child.EndByte(), child.EndByte()}
			chunks = append(chunks, tsc.chunkNode(child)...)
		} else if child.EndByte()-child.StartByte()+curSpan.Len() > maxBytes {
			// Start a new span
			chunks = append(chunks, curSpan)
			curSpan = Span{child.StartByte(), child.EndByte()}
		} else {
			// Add to the current span
			curSpan = curSpan.Add(Span{child.EndByte(), child.EndByte()})
		}
	}

	return append(chunks, curSpan)
}

// TODO expand or find a better way to do this
func detectLanguage(filename string) *sitter.Language {
	switch filepath.Ext(filename) {
	case ".py":
		return python.GetLanguage()
	case ".js":
		return javascript.GetLanguage()
	case ".ts":
		return typescript.GetLanguage()
	case ".tsx":
		return tsx.GetLanguage()
	case ".java":
		return java.GetLanguage()
	case ".scala":
		return scala.GetLanguage()
	case ".kt":
		return kotlin.GetLanguage()
	case ".c":
		return c.GetLanguage()
	case ".cc", ".cpp", ".cxx", ".h", ".hh", ".hxx":
		return cpp.GetLanguage()
	case ".cs":
		return csharp.GetLanguage()
	case ".go":
		return golang.GetLanguage()
	case ".rb":
		return ruby.GetLanguage()
	case ".rs":
		return rust.GetLanguage()
	case ".php":
		return php.GetLanguage()
	case ".swift":
		return swift.GetLanguage()
	case ".sh":
		return bash.GetLanguage()
	case ".sql":
		return sql.GetLanguage()
	default:
		return nil
	}

}
