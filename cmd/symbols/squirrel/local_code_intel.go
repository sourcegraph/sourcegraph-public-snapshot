package squirrel

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

// Nominal type for symbol names.
type SymbolName string

// Scope is a mapping from symbol name to symbol.
type Scope = map[SymbolName]*PartialSymbol // pointer for mutability

// PartialSymbol is the same as types.Symbol, but with the refs stored in a map to deduplicate.
type PartialSymbol struct {
	Hover *string
	Def   *types.Range
	// Store refs as a set to avoid duplicates from some tree-sitter queries.
	Refs map[types.Range]struct{}
}

// Computes the local code intel payload, which is a list of symbols.
func (squirrel *SquirrelService) localCodeIntel(ctx context.Context, repoCommitPath types.RepoCommitPath) (*types.LocalCodeIntelPayload, error) {
	// Parse the file.
	root, err := squirrel.parse(ctx, repoCommitPath)
	if err != nil {
		return nil, err
	}

	debug := os.Getenv("SQUIRREL_DEBUG") == "true"

	// Collect scopes
	rootScopeId := nodeId(root.Node)
	scopes := map[Id]Scope{
		rootScopeId: {},
	}
	err = forEachCapture(root.LangSpec.query, *root, func(captureName string, node Node) {
		if captureName == "scope" {
			scopes[nodeId(node.Node)] = map[SymbolName]*PartialSymbol{}
			return
		}
	})
	if err != nil {
		return nil, err
	}

	// Collect defs
	err = forEachCapture(root.LangSpec.query, *root, func(captureName string, node Node) {
		// Only collect "definition*" captures.
		if strings.HasPrefix(captureName, "definition") {
			// Find the nearest scope (if it exists).
			for cur := node.Node; cur != nil; cur = cur.Parent() {
				// Found the scope.
				if scope, ok := scopes[nodeId(cur)]; ok {
					// Get the symbol name.
					symbolName := SymbolName(node.Content(node.Contents))

					// Print a debug message if the symbol is already defined.
					if symbol, ok := scope[symbolName]; ok && debug {
						lines := strings.Split(string(node.Contents), "\n")
						fmt.Printf("duplicate definition for %q in %s (skipping)\n", symbolName, repoCommitPath.Path)
						fmt.Printf("  %4d | %s\n", symbol.Def.Row, lines[symbol.Def.Row])
						fmt.Printf("  %4d | %s\n", node.StartPoint().Row, lines[node.StartPoint().Row])

						// Skip it.
						break
					}

					// Get the hover.
					hover := findHover(node)

					// Put the symbol in the scope.
					def := nodeToRange(node.Node)
					scope[symbolName] = &PartialSymbol{
						Hover: hover,
						Def:   &def,
						Refs:  map[types.Range]struct{}{},
					}

					// Stop walking up the tree.
					break
				}
			}
		}
	})
	if err != nil {
		return nil, err
	}

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
				(*scope[symbolName]).Refs[nodeToRange(node)] = struct{}{}

				// Done.
				return
			}
		}

		// Did not find the symbol in this file, so create a symbol at the root without a def for it.
		if _, ok := scopes[rootScopeId][symbolName]; !ok {
			scopes[rootScopeId][symbolName] = &PartialSymbol{Refs: map[types.Range]struct{}{}}
		}
		scopes[rootScopeId][symbolName].Refs[nodeToRange(node)] = struct{}{}
	})

	// Collect the symbols.
	symbols := []types.Symbol{}
	for _, scope := range scopes {
		for _, partialSymbol := range scope {
			if partialSymbol.Def == nil && len(partialSymbol.Refs) == 0 && debug {
				fmt.Println("no def or refs for", partialSymbol)
				continue
			}
			refs := []types.Range{}
			for ref := range partialSymbol.Refs {
				refs = append(refs, ref)
			}
			symbols = append(symbols, types.Symbol{
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
		if payload.Symbols[i].Def == nil && payload.Symbols[j].Def == nil {
			if len(payload.Symbols[i].Refs) == 0 && len(payload.Symbols[j].Refs) == 0 {
				fmt.Println("expected a definition or reference, sorting will be unstable")
				return true
			} else if len(payload.Symbols[i].Refs) == 0 {
				return false
			} else if len(payload.Symbols[j].Refs) == 0 {
				return true
			} else {
				return isLessRange(payload.Symbols[i].Refs[0], payload.Symbols[j].Refs[0])
			}
		} else if payload.Symbols[i].Def == nil {
			return false
		} else if payload.Symbols[j].Def == nil {
			return true
		} else {
			return isLessRange(*payload.Symbols[i].Def, *payload.Symbols[j].Def)
		}
	})

	// Print all symbols.
	for _, symbol := range payload.Symbols {
		// Print the hover.
		hover := "<no hover>"
		if symbol.Hover != nil {
			hover = *symbol.Hover
		}
		defColor := color.New(color.FgMagenta)
		refColor := color.New(color.FgCyan)
		fmt.Fprintf(w, "Hover %q, %s, %s\n", hover, defColor.Sprint("def"), refColor.Sprint("refs"))

		// Convert each def and ref into a rangeColor.
		type rangeColor struct {
			rnge   types.Range
			color_ *color.Color
		}

		rnges := []rangeColor{}

		if symbol.Def != nil {
			rnges = append(rnges, rangeColor{rnge: *symbol.Def, color_: defColor})
		}

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
