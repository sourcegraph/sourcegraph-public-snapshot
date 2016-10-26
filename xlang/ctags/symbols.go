package ctags

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/ctags/parser"
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
	"func":        lsp.SKFunction,
}

func (h *Handler) handleSymbol(ctx context.Context, params lsp.WorkspaceSymbolParams) (symbols []lsp.SymbolInformation, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ctags.handleSymbol")
	if params.Query != "" {
		span.SetTag("query", params.Query)
	}
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogEvent(fmt.Sprintf("error: %v", err))
		}
		span.SetTag("returned symbols count", len(symbols))
		span.Finish()
	}()

	tags, err := h.getTags(ctx)
	if err != nil {
		return nil, err
	}

	span.SetTag("tags count", len(tags))
	vslog("Total definitions found: ", strconv.Itoa(len(tags)))

	tags = filterRankTags(ctx, params.Query, tags, params.Limit)
	span.SetTag("filtered tags count", len(tags))

	symbols = make([]lsp.SymbolInformation, 0, len(tags))
	for _, tag := range tags {
		if symbol := tagToSymbol(&tag); symbol != nil {
			symbols = append(symbols, *symbol)
		}
	}
	vslog("Returning definitions: ", strconv.Itoa(len(symbols)))

	return symbols, nil
}

func filterRankTags(ctx context.Context, query string, tags []parser.Tag, limit int) []parser.Tag {
	filterSpan, ctx := opentracing.StartSpanFromContext(ctx, "filter tags")
	defer filterSpan.Finish()

	if query != "" {
		q := strings.ToLower(query)
		exact, prefix, contains := []parser.Tag{}, []parser.Tag{}, []parser.Tag{}
		for _, t := range tags {
			name := strings.ToLower(t.Name)
			if name == q {
				exact = append(exact, t)
			} else if strings.HasPrefix(name, q) {
				prefix = append(prefix, t)
			} else if strings.Contains(name, q) {
				contains = append(contains, t)
			}
		}
		tags = append(append(exact, prefix...), contains...)
	}

	if len(tags) < limit {
		return tags
	}
	return tags[:limit]
}

// tagToSymbol converts a Tag to a SymbolInformation. In some cases,
// this conversion isn't valid, in which case the return value is nil.
func tagToSymbol(tag *parser.Tag) *lsp.SymbolInformation {
	loc := tagToLocation(tag)
	if loc == nil {
		return nil
	}
	kind := nameToSymbolKind[tag.Kind]
	if kind == 0 {
		kind = lsp.SKVariable
	}
	return &lsp.SymbolInformation{
		Name:     tag.Name,
		Kind:     kind,
		Location: *loc,
	}
}

func tagToLocation(tag *parser.Tag) *lsp.Location {
	nameIdx := strings.Index(tag.DefLinePrefix, tag.Name)
	if nameIdx < 0 {
		// Indicate no location if we couldn't find the name in the def line prefix.
		return nil
	}
	return &lsp.Location{
		URI: "file:///" + tag.File,
		Range: lsp.Range{
			Start: lsp.Position{Line: tag.LineNumber - 1, Character: nameIdx},
			End:   lsp.Position{Line: tag.LineNumber - 1, Character: nameIdx + len(tag.Name)},
		},
	}
}
