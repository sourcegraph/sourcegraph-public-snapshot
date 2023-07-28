package squirrel

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/fatih/color"
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

// SymbolName is a nominal type for symbol names.
type SymbolName string

// Scope is a mapping from symbol name to symbol.
type Scope = map[SymbolName]*PartialSymbol // pointer for mutability

// PartialSymbol is the same as types.Symbol, but with the refs stored in a map to deduplicate.
type PartialSymbol struct {
	Name  string
	Hover string
	Def   types.Range
	// Store refs as a set to avoid duplicates from some tree-sitter queries.
	Refs map[types.Range]struct{}
}

// LocalCodeIntel computes the local code intel payload, which is a list of symbols.
func (s *SquirrelService) LocalCodeIntel(ctx context.Context, repoCommitPath types.RepoCommitPath) (*types.LocalCodeIntelPayload, error) {
	// Parse the file.
	root, err := s.parse(ctx, repoCommitPath)
	if err != nil {
		return nil, err
	}

	// Collect scopes
	scopes := map[NodeId]Scope{}
	forEachCapture(root.LangSpec.localsQuery, *root, func(nameToNode map[string]Node) {
		if node, ok := nameToNode["scope"]; ok {
			scopes[nodeId(node.Node)] = map[SymbolName]*PartialSymbol{}
			return
		}
	})

	// Collect defs
	forEachCapture(root.LangSpec.localsQuery, *root, func(nameToNode map[string]Node) {
		for captureName, node := range nameToNode {
			// Only collect "definition*" captures.
			if strings.HasPrefix(captureName, "definition") {
				// Find the nearest scope (if it exists).
				for cur := node.Node; cur != nil; cur = cur.Parent() {
					// Found the scope.
					if scope, ok := scopes[nodeId(cur)]; ok {
						// Get the symbol name.
						symbolName := SymbolName(strings.ToValidUTF8(node.Content(node.Contents), "�"))

						// Skip the symbol if it's already defined.
						if _, ok := scope[symbolName]; ok {
							break
						}

						// Put the symbol in the scope.
						scope[symbolName] = &PartialSymbol{
							Name:  string(symbolName),
							Hover: findHover(node),
							Def:   nodeToRange(node.Node),
							Refs:  map[types.Range]struct{}{},
						}

						// Stop walking up the tree.
						break
					}
				}
			}
		}
	})

	// Collect refs by walking the entire tree.
	walk(root.Node, func(node *sitter.Node) {
		// Only collect identifiers.
		if !strings.Contains(node.Type(), "identifier") {
			return
		}

		// Get the symbol name.
		symbolName := SymbolName(node.Content(root.Contents))

		// Find the nearest scope (if it exists).
		for cur := node; cur != nil; cur = cur.Parent() {
			if scope, ok := scopes[nodeId(cur)]; ok {
				// Check if it's in the scope.
				if _, ok := scope[symbolName]; !ok {
					// It's not in this scope, so keep walking up the tree.
					continue
				}

				// Put the ref in the scope.
				scope[symbolName].Refs[nodeToRange(node)] = struct{}{}

				// Done.
				return
			}
		}

		// Did not find the symbol in this file, so ignore it.
	})

	// Collect the symbols.
	symbols := []types.Symbol{}
	for _, scope := range scopes {
		for _, partialSymbol := range scope {
			refs := []types.Range{}
			for ref := range partialSymbol.Refs {
				refs = append(refs, ref)
			}
			symbols = append(symbols, types.Symbol{
				Name:  partialSymbol.Name,
				Hover: partialSymbol.Hover,
				Def:   partialSymbol.Def,
				Refs:  refs,
			})
		}
	}

	return &types.LocalCodeIntelPayload{Symbols: symbols}, nil
}

// Pretty prints the local code intel payload for debugging.
func prettyPrintLocalCodeIntelPayload(w io.Writer, payload types.LocalCodeIntelPayload, contents string) {
	lines := strings.Split(contents, "\n")

	// Sort payload.Symbols by Def Row then Column.
	sort.Slice(payload.Symbols, func(i, j int) bool {
		return isLessRange(payload.Symbols[i].Def, payload.Symbols[j].Def)
	})

	// Print all symbols.
	for _, symbol := range payload.Symbols {
		defColor := color.New(color.FgMagenta)
		refColor := color.New(color.FgCyan)
		fmt.Fprintf(w, "Hover %q, %s, %s\n", symbol.Hover, defColor.Sprint("def"), refColor.Sprint("refs"))

		// Convert each def and ref into a rangeColor.
		type rangeColor struct {
			rnge   types.Range
			color_ *color.Color
		}

		rnges := []rangeColor{}
		rnges = append(rnges, rangeColor{rnge: symbol.Def, color_: defColor})
		for _, ref := range symbol.Refs {
			rnges = append(rnges, rangeColor{rnge: ref, color_: refColor})
		}

		// How to print a range in color.
		printRange := func(rnge types.Range, c *color.Color) {
			line := lines[rnge.Row]
			lineWithSpaces := tabsToSpaces(line)
			column := lengthInSpaces(line[:rnge.Column])
			length := lengthInSpaces(line[rnge.Column : rnge.Column+rnge.Length])
			fmt.Fprint(w, color.New(color.FgBlack).Sprintf("%4d | ", rnge.Row))
			fmt.Fprint(w, color.New(color.FgBlack).Sprint(lineWithSpaces[:column]))
			fmt.Fprint(w, c.Sprint(lineWithSpaces[column:column+length]))
			fmt.Fprint(w, color.New(color.FgBlack).Sprint(lineWithSpaces[column+length:]))
			fmt.Fprintln(w)
		}

		// Sort ranges by row, then column.
		sort.Slice(rnges, func(i, j int) bool {
			if rnges[i].rnge.Row == rnges[j].rnge.Row {
				return rnges[i].rnge.Column < rnges[j].rnge.Column
			}
			return rnges[i].rnge.Row < rnges[j].rnge.Row
		})

		// Print each range.
		for _, rnge := range rnges {
			printRange(rnge.rnge, rnge.color_)
		}

		fmt.Fprintln(w)
	}
}
