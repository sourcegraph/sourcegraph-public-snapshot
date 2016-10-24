package ctags

import (
	"context"
	"go/scanner"
	"go/token"
	"os"

	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/sourcegraph-go/pkg/lsp"
)

func (h *Handler) handleReferences(ctx context.Context, params lsp.ReferenceParams) ([]lsp.Location, error) {
	// Filter for interesting files.
	var filter = func(info os.FileInfo) bool {
		return isSupportedFile(info.Name())
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
		locs = append(locs, locationsOfWordInFile(fileBytes, filename, word)...)
	}

	return locs, nil
}

func locationsOfWordInFile(b []byte, filename, word string) (locs []lsp.Location) {
	fs := token.NewFileSet()
	f := fs.AddFile(filename, fs.Base(), len(b))
	var sc scanner.Scanner
	var eh scanner.ErrorHandler
	sc.Init(f, b, eh, scanner.ScanComments)
	for pos, tok, lit := sc.Scan(); tok != token.EOF; pos, tok, lit = sc.Scan() {
		if tok == token.IDENT && lit == word {
			position := f.Position(pos)
			locs = append(locs, lsp.Location{
				URI: "file://" + filename,
				Range: lsp.Range{
					Start: lsp.Position{Line: position.Line - 1, Character: position.Column - 1},
					End:   lsp.Position{Line: position.Line - 1, Character: position.Column - 1 + len(word)},
				},
			})
		}
	}
	return
}
