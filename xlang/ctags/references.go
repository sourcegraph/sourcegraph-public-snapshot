package ctags

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/sourcegraph-go/pkg/lsp"
)

func (h *Handler) handleReferences(ctx context.Context, params lsp.ReferenceParams) ([]lsp.Location, error) {
	// Filter for interesting files.
	var filter = func(info os.FileInfo) bool {
		ext := filepath.Ext(params.TextDocument.URI)
		return filepath.Ext(info.Name()) == ext
	}

	files, err := ctxvfs.ReadAllFiles(ctx, h.fs, "", filter)
	if err != nil {
		return nil, err
	}

	word, _, err := wordAtPosition(ctx, h.fs, params.TextDocumentPositionParams)
	if err != nil {
		return nil, err
	}

	var locs []lsp.Location
	for filename, fileBytes := range files {
		file := string(fileBytes)
		locs = append(locs, locationsOfWordInFile(file, filename, word)...)
	}

	return locs, nil
}

func locationsOfWordInFile(file, filename, word string) []lsp.Location {
	var locs []lsp.Location
	re := regexp.MustCompile(".*" + word + ".*")
	for lineNumber, line := range strings.Split(file, "\n") {
		ranges := re.FindAllStringIndex(line, -1)
		for _, wordIndices := range ranges {
			locs = append(locs, lsp.Location{
				Range: lsp.Range{
					Start: lsp.Position{
						Line:      lineNumber,
						Character: wordIndices[0],
					},
					End: lsp.Position{
						Line:      lineNumber,
						Character: wordIndices[1],
					},
				},
				URI: filename, // TODO ADD URI STUFF
			})
		}
	}
	return locs
}
