package ctags

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

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
	"func":        lsp.SKFunction,
}

func (h *Handler) handleDefinition(ctx context.Context, req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) ([]lsp.Location, error) {
	vslog(fmt.Sprintf("Requesting jump-to-def with params %+v", params))

	rootDir := h.init.RootPath
	filename := filepath.Join(rootDir, strings.TrimPrefix(params.TextDocument.URI, "file:///"))
	l, c := params.Position.Line, params.Position.Character

	tok, err := getTokenFromFile(filename, l, c)
	if err != nil {
		return nil, err
	}
	if tok.kind != tokName {
		return nil, nil
	}

	p, err := parser.Parse(ctx, rootDir, nil)
	if err != nil {
		return nil, err
	}

	matches := make([]lsp.Location, 0)
	for _, tag := range p.Tags() {
		if tag.Name == tok.token {
			if loc := tagToLocation(&tag, rootDir); loc != nil {
				matches = append(matches, *loc)
			}
		}
	}

	return matches, nil
}

func (h *Handler) handleReferences(ctx context.Context, req *jsonrpc2.Request, params lsp.ReferenceParams) ([]lsp.Location, error) {
	// TODO: deal with params.IncludeDeclaration
	vslog(fmt.Sprintf("Requesting references with params %+v", params))

	rootDir := h.init.RootPath
	file := filepath.Join(rootDir, strings.TrimPrefix(params.TextDocument.URI, "file:///"))
	l, c := params.Position.Line, params.Position.Character
	tok, err := getTokenFromFile(file, l, c)
	if err != nil {
		return nil, err
	}
	if tok.kind != tokName {
		return nil, nil
	}

	vslog(fmt.Sprintf("finding references for %s", tok.token))

	// Search for any token occurrence
	var reflocs []lsp.Location
	{
		grepCmd := exec.Command("pt", "--nogroup", "--numbers", "--nocolor", "--column", "-w", tok.token, "--ignore=tags")
		grepCmd.Dir = rootDir
		b, err := grepCmd.Output()
		if err != nil {
			return nil, fmt.Errorf("could not run `pt`: %s", err)
		}
		lines := strings.Split(strings.TrimSpace(string(b)), "\n")
		for _, line_ := range lines {
			if line_ == "" {
				continue
			}

			cmps := strings.SplitN(line_, ":", 4)
			if len(cmps) != 4 {
				return nil, fmt.Errorf("error parsing pt output line, expected format $FILE:$LINE:$COL:$LINE_CONTENT, but got %q", line_)
			}
			file, lineno_, colno_ := cmps[0], cmps[1], cmps[2]
			lineno, err := strconv.Atoi(lineno_)
			if err != nil {
				return nil, fmt.Errorf("could not parse line number from line %q", line_)
			}
			lineno--
			colno, err := strconv.Atoi(colno_)
			if err != nil {
				return nil, fmt.Errorf("could not parse column number from line %q", line_)
			}
			colno--

			reflocs = append(reflocs, lsp.Location{
				URI: "file://" + filepath.Join(rootDir, file),
				Range: lsp.Range{
					Start: lsp.Position{Line: lineno, Character: colno},
					End:   lsp.Position{Line: lineno, Character: colno + len(tok.token)},
				},
			})
		}
	}

	vslog(fmt.Sprintf("Returning %d references", len(reflocs)))

	return reflocs, nil
}

func (h *Handler) handleSymbol(ctx context.Context, req *jsonrpc2.Request, params lsp.WorkspaceSymbolParams) (symbols []lsp.SymbolInformation, err error) {
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

	rootDir := h.init.RootPath
	vslog("Requesting workspace symbols for ", rootDir, " with params ", fmt.Sprintf("%q", params))
	p, err := parser.Parse(ctx, rootDir, nil)
	if err != nil {
		return nil, err
	}

	tags := p.Tags()

	span.SetTag("tags count", len(tags))
	vslog("Total definitions found: ", strconv.Itoa(len(tags)))

	tags = filterRankTags(ctx, params.Query, tags)
	span.SetTag("filtered tags count", len(tags))

	symbols = make([]lsp.SymbolInformation, 0, len(tags))
	for _, tag := range tags {
		if symbol := tagToSymbol(&tag, rootDir); symbol != nil {
			symbols = append(symbols, *symbol)
		}
	}
	vslog("Returning definitions: ", strconv.Itoa(len(symbols)))

	return symbols, nil
}

func filterRankTags(ctx context.Context, query string, tags []parser.Tag) []parser.Tag {
	filterSpan, ctx := opentracing.StartSpanFromContext(ctx, "filter tags")
	defer filterSpan.Finish()

	// Limit the amount of symbols we serve to the client. Allowing an
	// excessively large amount to be returned will generate a huge response
	// object, which slows down the performance of the pipeline significantly.
	const limit = 100

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
func tagToSymbol(tag *parser.Tag, rootDir string) *lsp.SymbolInformation {
	loc := tagToLocation(tag, rootDir)
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

func tagToLocation(tag *parser.Tag, rootDir string) *lsp.Location {
	nameIdx := strings.Index(tag.DefLinePrefix, tag.Name)
	if nameIdx < 0 {
		// Indicate no location if we couldn't find the name in the def line prefix.
		return nil
	}
	return &lsp.Location{
		URI: "file://" + rootDir + "/" + tag.File,
		Range: lsp.Range{
			Start: lsp.Position{Line: tag.Line - 1, Character: nameIdx},
			End:   lsp.Position{Line: tag.Line - 1, Character: nameIdx + len(tag.Name)},
		},
	}
}
