package ctags

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/sourcegraph-go/pkg/lsp"
)

func (h *Handler) handleDefinition(ctx context.Context, params lsp.TextDocumentPositionParams) ([]lsp.Location, error) {
	tags, err := h.getTags(ctx)
	if err != nil {
		return nil, err
	}
	word, _, err := wordAtPosition(ctx, h.fs, params)
	if err != nil {
		return nil, err
	}
	var locations []lsp.Location
	for _, tag := range tags {
		if tag.Name == word {
			loc := *tagToLocation(&tag)
			locations = append(locations, loc)
		}
	}
	return locations, nil
}

var ErrBadRequest = fmt.Errorf("invalid position argument")

// wordAtPosition finds the word and start character of the word at the given position.
func wordAtPosition(ctx context.Context, fs ctxvfs.FileSystem, params lsp.TextDocumentPositionParams) (string, int, error) {
	path := strings.TrimPrefix(params.TextDocument.URI, "file://")
	rsk, err := fs.Open(ctx, path)
	if err != nil {
		return "", 0, err
	}
	defer rsk.Close()

	b, err := ioutil.ReadAll(rsk)
	if err != nil {
		return "", 0, err
	}

	lineNumber := params.Position.Line
	col := params.Position.Character

	if lineNumber < 1 || col < 1 {
		return "", 0, ErrBadRequest
	}
	lines := bytes.Split(b, []byte("\n"))
	if lineNumber >= len(lines) {
		return "", 0, ErrBadRequest
	}
	line := string(lines[lineNumber])
	if col >= len(line) {
		return "", 0, ErrBadRequest
	}
	word := wordAtPoint(line, col)
	wordStart := strings.Index(line, word) + 1
	return word, wordStart, nil
}

// wordRE is a poor man's parser. We're looking for something that looks like an
// identifier.
var wordRE = regexp.MustCompile(`\w+`)

// wordAtPoint finds something that looks like an identifier, that is located at
// the one indexed column.
func wordAtPoint(line string, col int) string {
	indicies := wordRE.FindAllStringIndex(line, -1)

	col = col - 1 // LSP is one indexed
	for _, v := range indicies {
		if v[0] <= col && v[1] > col {
			return line[v[0]:v[1]]
		}
	}
	return ""
}
