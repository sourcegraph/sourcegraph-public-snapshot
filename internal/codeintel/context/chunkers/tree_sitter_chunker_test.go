package chunkers

import (
	"context"
	"fmt"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
	ctxt "github.com/sourcegraph/sourcegraph/internal/codeintel/context"
	"github.com/stretchr/testify/suite"
	"io"
	"os"
	"strings"
	"testing"
)

type TreeSitterChunkerSuite struct {
	suite.Suite
}

func TestTreeSitterChunkerSuite(t *testing.T) {
	suite.Run(t, new(TreeSitterChunkerSuite))
}

func (ts *TreeSitterChunkerSuite) TestTreeSitterChunker() {
	chunkOptions := &ChunkOptions{
		ChunkTokensThreshold: 10,
		CoalesceThreshold:    5,
	}
	tsc := NewTreeSitterChunker(chunkOptions)

	content := `function fib(n) {
  if (n < 2) {
    return 1
  } else {
    return fib(n - 2) + fib(n - 1)
  }
}`
	content = strings.ReplaceAll(content, "\n", " ")
	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, []byte(content))
	node := tree.RootNode()
	ts.NoError(err)

	for i := 0; uint32(i) < node.ChildCount(); i++ {
		c := node.Child(i)
		fmt.Println("child", content[c.StartByte():c.EndByte()])
		if c.ChildCount() > 0 {
			fmt.Println("     has children")
		}
		for j := 0; uint32(j) < c.ChildCount(); j++ {
			cc := c.Child(j)
			fmt.Println("         child", content[cc.StartByte():cc.EndByte()])
			if cc.ChildCount() > 0 {
				fmt.Println("             has children")
			}

			for k := 0; uint32(k) < cc.ChildCount(); k++ {
				ccc := cc.Child(k)
				fmt.Println("                 child", content[ccc.StartByte():ccc.EndByte()])

				if ccc.ChildCount() > 0 {
					fmt.Println("               has children")
				}
			}
		}
	}

	chunks := tsc.Chunk(content, "foo.js")
	fmt.Println()

	fmt.Println("\nchunks")
	for _, chunk := range chunks {
		fmt.Println(chunk)
	}

	reconstructedContent := ""
	for _, chunk := range chunks {
		reconstructedContent += chunk.Content
	}

	ts.Equal(content, reconstructedContent)
}

func (ts *TreeSitterChunkerSuite) TestCases() {

	opts := &ChunkOptions{
		ChunkTokensThreshold:           100,
		NoSplitTokensThreshold:         150,
		ChunkEarlySplitTokensThreshold: 120,
		CoalesceThreshold:              50,
	}

	files, err := os.ReadDir("testdata")
	ts.NoError(err)
	for _, file := range files {
		ts.Run(file.Name(), func() {
			ts.testTreeSitterChunker("testdata/"+file.Name(), nil, opts)
			// ts.testClassicChunker("testdata/"+file.Name(), nil, opts)
		})
	}
}

func (ts *TreeSitterChunkerSuite) testTreeSitterChunker(filename string, expectedChunks []ctxt.EmbeddableChunk, chunkOptions *ChunkOptions) {
	tsc := NewTreeSitterChunker(chunkOptions)

	file, err := os.Open(filename)
	ts.NoError(err)
	defer file.Close()
	fileBytes, err := io.ReadAll(file)
	ts.NoError(err)
	content := string(fileBytes)

	chunks := tsc.Chunk(content, filename)
	reconstructedContent := ""
	for _, chunk := range chunks {
		// ts.Equal(expectedChunks[i], chunk)
		reconstructedContent += chunk.Content
		// fmt.Printf("chunk (len %d, start line %d, end line %d): ```\n%s```\n", len(chunk.Content), chunk.StartLine, chunk.EndLine, chunk.Content)
	}

	ts.Equal(strings.TrimSpace(content), strings.TrimSpace(reconstructedContent))
}
