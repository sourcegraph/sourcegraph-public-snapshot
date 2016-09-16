package ctags

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/lang/ctags/parser"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

var nameToSymbolKind = map[string]lsp.SymbolKind{
	"file":        lsp.SKFile,
	"module":      lsp.SKModule,
	"namespace":   lsp.SKNamespace,
	"package":     lsp.SKPackage,
	"class":       lsp.SKClass,
	"method":      lsp.SKMethod,
	"property":    lsp.SKProperty,
	"field":       lsp.SKField,
	"constructor": lsp.SKConstructor,
	"enum":        lsp.SKEnum,
	"interface":   lsp.SKInterface,
	"function":    lsp.SKFunction,
	"variable":    lsp.SKVariable,
	"constant":    lsp.SKConstant,
	"string":      lsp.SKString,
	"number":      lsp.SKNumber,
	"boolean":     lsp.SKBoolean,
	"array":       lsp.SKArray,
}

func (h *Handler) handleSymbol(ctx context.Context, req *jsonrpc2.Request, params lsp.WorkspaceSymbolParams) ([]lsp.SymbolInformation, error) {
	rootDir := h.init.RootPath
	vslog("Requesting workspace symbols for ", rootDir, " with params ", fmt.Sprintf("%q", params))
	p, err := parser.Parse(ctx, rootDir, nil)
	if err != nil {
		return nil, err
	}

	tags := p.Tags()

	vslog("Total definitions found: ", strconv.Itoa(len(tags)))

	var symbols []lsp.SymbolInformation
	if params.Query == "" {
		symbols = make([]lsp.SymbolInformation, 0, len(tags))
	} else {
		symbols = make([]lsp.SymbolInformation, 0)
	}
	for _, tag := range tags {
		nameIdx := strings.Index(tag.DefLinePrefix, tag.Name)
		if nameIdx < 0 {
			// Drop this tag if we couldn't find the name in the def line prefix.
			// TODO(beyang): warn or error here?
			continue
		}
		kind := nameToSymbolKind[tag.Kind]
		if kind == 0 {
			kind = lsp.SKVariable
		}
		if params.Query == "" || strings.HasPrefix(strings.ToLower(tag.Name), strings.ToLower(params.Query)) {
			symbols = append(symbols, lsp.SymbolInformation{
				Name: tag.Name,
				Kind: kind,
				Location: lsp.Location{
					URI: "file://" + rootDir + "/" + tag.File,
					Range: lsp.Range{
						Start: lsp.Position{Line: tag.Line - 1, Character: nameIdx},
						End:   lsp.Position{Line: tag.Line - 1, Character: nameIdx + len(tag.Name)},
					},
				},
			})
		}
	}

	vslog("Returning definitions: ", strconv.Itoa(len(symbols)))

	return symbols, nil
}
