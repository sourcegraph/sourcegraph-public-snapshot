package squirrel

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/grafana/regexp"
	"github.com/inconshreveable/log15"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/bash"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/css"
	"github.com/smacker/go-tree-sitter/dockerfile"
	"github.com/smacker/go-tree-sitter/elm"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/hcl"
	"github.com/smacker/go-tree-sitter/html"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/lua"
	"github.com/smacker/go-tree-sitter/ocaml"
	"github.com/smacker/go-tree-sitter/php"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/scala"
	"github.com/smacker/go-tree-sitter/svelte"
	"github.com/smacker/go-tree-sitter/toml"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
	"github.com/smacker/go-tree-sitter/yaml"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func LocalCodeIntelHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log15.Error("failed to read request body", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var args types.RepoCommitPath
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&args); err != nil {
		log15.Error("failed to decode request body", "err", err, "body", string(body))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	repo := args.Repo
	commit := args.Commit
	path := args.Path

	debug := os.Getenv("SQUIRREL_DEBUG") == "true"
	debugStringBuilder := &strings.Builder{}

	cmd := gitserver.DefaultClient.Command("git", "cat-file", "blob", commit+":"+path)
	cmd.Repo = api.RepoName(repo)
	contents, stderr, err := cmd.DividedOutput(r.Context())
	if err != nil {
		log15.Error("failed to get file contents", "stdout", contents, "stderr", stderr)
		http.Error(w, fmt.Sprintf("failed to get file contents: %s", err), http.StatusInternalServerError)
		return
	}

	result, err := localCodeIntel(path, string(contents))
	if result != nil && debug {
		fmt.Fprintln(debugStringBuilder, "👉 repo:", repo, "commit:", commit, "path:", path)
		prettyPrintLocalCodeIntelPayload(debugStringBuilder, args, *result, string(contents))
		fmt.Fprintln(debugStringBuilder, "✅ repo:", repo, "commit:", commit, "path:", path)

		fmt.Println(" ")
		fmt.Println(bracket(debugStringBuilder.String()))
		fmt.Println(" ")
	}
	if err != nil {
		_ = json.NewEncoder(w).Encode(nil)
		log15.Error("failed to get definition", "err", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		log15.Error("failed to write response: %s", "error", err)
		http.Error(w, fmt.Sprintf("failed to get definition: %s", err), http.StatusInternalServerError)
		return
	}
}

func localCodeIntel(fullPath string, contents string) (*types.LocalCodeIntelPayload, error) {
	ext := strings.TrimPrefix(filepath.Ext(fullPath), ".")

	langName, ok := extToLang[ext]
	if !ok {
		return nil, errors.Newf("unrecognized file extension %s", ext)
	}

	langSpec, ok := langToLangSpec[langName]
	if !ok {
		return nil, errors.Newf("unsupported language %s", langName)
	}

	localsPath := path.Join("nvim-treesitter", "queries", langSpec.nvimQueryDir, "locals.scm")
	queriesBytes, err := queriesFs.ReadFile(localsPath)
	if err != nil {
		return nil, errors.Newf("could not read %d: %s", localsPath, err)
	}
	queryString := string(queriesBytes)

	parser := sitter.NewParser()
	parser.SetLanguage(langSpec.language)

	tree, err := parser.ParseCtx(context.Background(), nil, []byte(contents))
	if err != nil {
		return nil, fmt.Errorf("failed to parse file contents: %s", err)
	}
	root := tree.RootNode()

	debug := os.Getenv("SQUIRREL_DEBUG") == "true"

	// Collect scopes
	rootScopeId := nodeId(root)
	scopes := map[Id]Scope{
		rootScopeId: {},
	}
	err = forEachCapture(queryString, root, langSpec.language, func(captureName string, node *sitter.Node) {
		if captureName == "scope" {
			scopes[nodeId(node)] = map[string]*PartialSymbol{}
			return
		}
	})
	if err != nil {
		return nil, err
	}

	// Collect defs
	err = forEachCapture(queryString, root, langSpec.language, func(captureName string, node *sitter.Node) {
		// Only collect "definition*" captures.
		if strings.HasPrefix(captureName, "definition") {
			// Find the nearest scope (if it exists).
			for cur := node; cur != nil; cur = cur.Parent() {
				// Found the scope.
				if scope, ok := scopes[nodeId(cur)]; ok {
					// Get the symbol name.
					symbolName := node.Content([]byte(contents))

					// Print a debug message if the symbol is already defined.
					if symbol, ok := scope[symbolName]; ok && debug {
						lines := strings.Split(contents, "\n")
						fmt.Printf("duplicate definition for %q in %s (using second)\n", symbolName, fullPath)
						fmt.Printf("  %4d | %s\n", symbol.Def.Row, lines[symbol.Def.Row])
						fmt.Printf("  %4d | %s\n", node.StartPoint().Row, lines[node.StartPoint().Row])
					}

					// Get the hover.
					hover := getHover(node, langSpec.commentStyle, contents)

					// Put the symbol in the scope.
					def := nodeToRange(node)
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
	walk(root, func(node *sitter.Node) {
		// Only collect identifiers.
		if !strings.Contains(node.Type(), "identifier") {
			return
		}

		// Get the symbol name.
		symbolName := node.Content([]byte(contents))

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

	// Collect the symbols
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

func getHover(node *sitter.Node, style CommentStyle, contents string) *string {
	hover := ""
	hover += "```" + style.codeFenceName + "\n"
	hover += strings.Split(contents, "\n")[node.StartPoint().Row] + "\n"
	hover += "```"

	for cur := node; cur != nil && cur.StartPoint().Row == node.StartPoint().Row; cur = cur.Parent() {
		prev := cur.PrevNamedSibling()

		// Skip over Java annotations and the like.
		for ; prev != nil; prev = prev.PrevNamedSibling() {
			if !contains(style.skipNodeTypes, prev.Type()) {
				break
			}
		}

		// Collect comments backwards.
		comments := []string{}
		lastStartRow := -1
		for ; prev != nil && contains(style.nodeTypes, prev.Type()); prev = prev.PrevNamedSibling() {
			if lastStartRow == -1 {
				lastStartRow = int(prev.StartPoint().Row)
			} else if lastStartRow != int(prev.EndPoint().Row+1) {
				break
			} else {
				lastStartRow = int(prev.StartPoint().Row)
			}

			comment := prev.Content([]byte(contents))

			// Strip line noise and delete garbage lines.
			lines := []string{}
			allLines := strings.Split(comment, "\n")
			for _, line := range allLines {
				if style.ignoreRegex != nil && style.ignoreRegex.MatchString(line) {
					continue
				}

				if style.stripRegex != nil {
					line = style.stripRegex.ReplaceAllString(line, "")
				}

				lines = append(lines, line)
			}

			// Remove shared leading spaces.
			spaces := math.MaxInt32
			for _, line := range lines {
				spaces = min(spaces, len(line)-len(strings.TrimLeft(line, " ")))
			}
			for i := range lines {
				lines[i] = strings.TrimLeft(lines[i], " ")
			}

			// Join lines.
			comments = append(comments, strings.Join(lines, "\n"))
		}

		if len(comments) == 0 {
			continue
		}

		// Reverse comments
		for i, j := 0, len(comments)-1; i < j; i, j = i+1, j-1 {
			comments[i], comments[j] = comments[j], comments[i]
		}

		hover = hover + "\n\n---\n\n" + strings.Join(comments, "\n") + "\n"
	}

	return &hover
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

type CommentStyle struct {
	placedBelow   bool
	ignoreRegex   *regexp.Regexp
	stripRegex    *regexp.Regexp
	skipNodeTypes []string
	nodeTypes     []string
	codeFenceName string
}

//go:embed nvim-treesitter
var queriesFs embed.FS

//go:embed language-file-extensions.json
var languageFileExtensionsJson string

var langToExts = func() map[string][]string {
	var m map[string][]string
	err := json.Unmarshal([]byte(languageFileExtensionsJson), &m)
	if err != nil {
		panic(err)
	}
	return m
}()

var extToLang = func() map[string]string {
	m := map[string]string{}
	for lang, exts := range langToExts {
		for _, ext := range exts {
			if _, ok := m[ext]; ok {
				panic(fmt.Sprintf("duplicate file extension %s", ext))
			}
			m[ext] = lang
		}
	}
	return m
}()

type LangSpec struct {
	nvimQueryDir string
	language     *sitter.Language
	commentStyle CommentStyle
}

var langToLangSpec = map[string]LangSpec{
	"cpp": {
		nvimQueryDir: "cpp",
		language:     cpp.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"csharp": {
		nvimQueryDir: "c_sharp",
		language:     csharp.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"css": {
		nvimQueryDir: "css",
		language:     css.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"dockerfile": {
		nvimQueryDir: "dockerfile",
		language:     dockerfile.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"elm": {
		nvimQueryDir: "elm",
		language:     elm.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"go": {
		nvimQueryDir: "go",
		language:     golang.GetLanguage(),
		commentStyle: CommentStyle{
			nodeTypes:     []string{"comment"},
			stripRegex:    regexp.MustCompile(`^//`),
			codeFenceName: "go",
		}, // TODO
	},
	"hcl": {
		nvimQueryDir: "hcl",
		language:     hcl.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"html": {
		nvimQueryDir: "html",
		language:     html.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"java": {
		nvimQueryDir: "java",
		language:     java.GetLanguage(),
		commentStyle: CommentStyle{
			nodeTypes:     []string{"line_comment", "block_comment"},
			stripRegex:    regexp.MustCompile(`(^//|^\s*\*|^/\*\*|\*/$)`),
			ignoreRegex:   regexp.MustCompile(`^\s*(/\*\*|\*/)\s*$`),
			codeFenceName: "java",
			skipNodeTypes: []string{"modifiers"},
		}, // TODO
	},
	"javascript": {
		nvimQueryDir: "javascript",
		language:     javascript.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"lua": {
		nvimQueryDir: "lua",
		language:     lua.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"ocaml": {
		nvimQueryDir: "ocaml",
		language:     ocaml.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"php": {
		nvimQueryDir: "php",
		language:     php.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"python": {
		nvimQueryDir: "python",
		language:     python.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"ruby": {
		nvimQueryDir: "ruby",
		language:     ruby.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"rust": {
		nvimQueryDir: "rust",
		language:     rust.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"scala": {
		nvimQueryDir: "scala",
		language:     scala.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"shell": {
		nvimQueryDir: "bash",
		language:     bash.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"svelte": {
		nvimQueryDir: "svelte",
		language:     svelte.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"toml": {
		nvimQueryDir: "toml",
		language:     toml.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"typescript": {
		nvimQueryDir: "typescript",
		language:     typescript.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"yaml": {
		nvimQueryDir: "yaml",
		language:     yaml.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
}

func prettyPrintLocalCodeIntelPayload(w io.Writer, args types.RepoCommitPath, payload types.LocalCodeIntelPayload, contents string) {
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

func isLessRange(a, b types.Range) bool {
	if a.Row == b.Row {
		return a.Column < b.Column
	}
	return a.Row < b.Row
}

func tabsToSpaces(s string) string {
	return strings.Replace(s, "\t", "    ", -1)
}

func lengthInSpaces(s string) int {
	total := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\t' {
			total += 4
		} else {
			total++
		}
	}
	return total
}

func spacesToColumn(s string, ix int) int {
	total := 0
	for i := 0; i < len(s); i++ {
		if total >= ix {
			return i
		}

		if s[i] == '\t' {
			total += 4
		} else {
			total++
		}
	}
	return total
}

type colorSprintfFunc func(a ...interface{}) string

func bracket(text string) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	if len(lines) == 1 {
		return "- " + text
	}

	for i, line := range lines {
		if i == 0 {
			lines[i] = "┌ " + line
		} else if i < len(lines)-1 {
			lines[i] = "│ " + line
		} else {
			lines[i] = "└ " + line
		}
	}

	return strings.Join(lines, "\n")
}

func forEachCapture(query string, root *sitter.Node, lang *sitter.Language, f func(name string, node *sitter.Node)) error {
	sitterQuery, err := sitter.NewQuery([]byte(query), lang)
	if err != nil {
		return errors.Newf("failed to parse query: %s\n%s", err, query)
	}
	cursor := sitter.NewQueryCursor()
	cursor.Exec(sitterQuery, root)

	match, _, hasCapture := cursor.NextCapture()
	for hasCapture {
		for _, capture := range match.Captures {
			name := sitterQuery.CaptureNameForId(capture.Index)
			f(name, capture.Node)
		}
		// Next capture
		match, _, hasCapture = cursor.NextCapture()
	}

	return nil
}

func nodeToRange(node *sitter.Node) types.Range {
	length := 1
	if node.StartPoint().Row == node.EndPoint().Row {
		length = int(node.EndPoint().Column - node.StartPoint().Column)
	}
	return types.Range{
		Row:    int(node.StartPoint().Row),
		Column: int(node.StartPoint().Column),
		Length: length,
	}
}

// The ID of a tree-sitter node.
type Id = string

type SymbolName = string

// Scope is a mapping from symbol name to symbol.
type Scope = map[SymbolName]*PartialSymbol // pointer for mutability

// PartialSymbol is the same as types.Symbol, but with the refs stored in a map to deduplicate.
type PartialSymbol struct {
	Hover *string
	Def   *types.Range
	// Store refs as a set to avoid duplicates from some tree-sitter queries.
	Refs map[types.Range]struct{}
}

// walk walks every node in the tree-sitter tree, calling f on each node.
func walk(node *sitter.Node, f func(node *sitter.Node)) {
	f(node)
	for i := 0; i < int(node.ChildCount()); i++ {
		walk(node.Child(i), f)
	}
}

func nodeId(node *sitter.Node) Id {
	return fmt.Sprint(nodeToRange(node))
}
