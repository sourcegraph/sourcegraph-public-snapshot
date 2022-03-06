package squirrel

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
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
	"github.com/smacker/go-tree-sitter/protobuf"
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

	if debug {
		fmt.Println("👉 repo:", repo, "commit:", commit, "path:", path)
	}

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
		prettyPrintLocalCodeIntelPayload(args, *result, string(contents))
	}
	if err != nil {
		_ = json.NewEncoder(w).Encode(nil)
		log15.Error("failed to get definition", "err", err)
		return
	}

	if debug {
		fmt.Println("✅ repo:", repo, "commit:", commit, "path:", path)
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

	nvimDir, ok := langToNvimQueryDir[langName]
	if !ok {
		return nil, errors.Newf("neovim-treesitter does not have queries for the language %s", langName)
	}

	localsPath := path.Join("nvim-treesitter", "queries", nvimDir, "locals.scm")
	queriesBytes, err := queriesFs.ReadFile(localsPath)
	if err != nil {
		return nil, errors.Newf("could not read %d: %s", localsPath, err)
	}
	queryString := string(queriesBytes)

	lang, ok := langToSitterLanguage[langName]
	if !ok {
		return nil, fmt.Errorf("no tree-sitter parser for language %s", langName)
	}

	parser := sitter.NewParser()
	parser.SetLanguage(lang)

	tree, err := parser.ParseCtx(context.Background(), nil, []byte(contents))
	if err != nil {
		return nil, fmt.Errorf("failed to parse file contents: %s", err)
	}
	root := tree.RootNode()

	getId := newGetIdFunc()

	// Collect all scopes, defs, and refs
	scopes := map[Id]Scope{}
	forEachCapture(queryString, root, lang, func(captureName string, node *sitter.Node) {
		// Record the scope
		if captureName == "scope" {
			scopes[getId(node)] = map[string]*PartialSymbol{}
			return
		}

		if node.IsNamed() {
			for cur := node; cur != nil; cur = cur.Parent() {
				if scope, ok := scopes[getId(cur)]; ok {
					symbolName := node.Content([]byte(contents))
					if _, ok := scope[symbolName]; !ok {
						scope[symbolName] = &PartialSymbol{
							Hover: nil,
							Def:   nil,
							Refs:  []types.Range{},
						}
					}

					// Put the def in the scope
					if strings.HasPrefix(captureName, "definition") {
						rnge := nodeToRange(node)
						(*scope[symbolName]).Def = &rnge
					}

					// Put the ref in the scope
					(*scope[symbolName]).Refs = append(scope[symbolName].Refs, nodeToRange(node))
				}
			}
		}
	})

	// Collect the symbols
	symbols := []types.Symbol{}
	for _, scope := range scopes {
		for _, partialSymbol := range scope {
			if partialSymbol.Def != nil {
				symbols = append(symbols, types.Symbol{
					Hover: partialSymbol.Hover,
					Def:   *partialSymbol.Def,
					Refs:  partialSymbol.Refs,
				})
			}
		}
	}

	return &types.LocalCodeIntelPayload{Symbols: symbols}, nil
}

// var goIdentifiers = []string{"identifier", "type_identifier"}

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

var langToNvimQueryDir = map[string]string{
	"cpp":        "cpp",
	"csharp":     "c_sharp",
	"css":        "css",
	"dockerfile": "dockerfile",
	"elm":        "elm",
	"go":         "go",
	"hcl":        "hcl",
	"html":       "html",
	"java":       "java",
	"javascript": "javascript",
	"lua":        "lua",
	"ocaml":      "ocaml",
	"php":        "php",
	"python":     "python",
	"ruby":       "ruby",
	"rust":       "rust",
	"scala":      "scala",
	"shell":      "bash",
	"svelte":     "svelte",
	"toml":       "toml",
	"typescript": "typescript",
	"yaml":       "yaml",
}

var langToSitterLanguage = map[string]*sitter.Language{
	// Sourcegraph's language map makes no distinction between c and cpp.
	// "c":          c.GetLanguage(),
	"cpp":        cpp.GetLanguage(),
	"csharp":     csharp.GetLanguage(),
	"css":        css.GetLanguage(),
	"dockerfile": dockerfile.GetLanguage(),
	"elm":        elm.GetLanguage(),
	"go":         golang.GetLanguage(),
	"hcl":        hcl.GetLanguage(),
	"html":       html.GetLanguage(),
	"java":       java.GetLanguage(),
	"javascript": javascript.GetLanguage(),
	"lua":        lua.GetLanguage(),
	"ocaml":      ocaml.GetLanguage(),
	"php":        php.GetLanguage(),
	"protobuf":   protobuf.GetLanguage(),
	"python":     python.GetLanguage(),
	"ruby":       ruby.GetLanguage(),
	"rust":       rust.GetLanguage(),
	"scala":      scala.GetLanguage(),
	"shell":      bash.GetLanguage(),
	"svelte":     svelte.GetLanguage(),
	"toml":       toml.GetLanguage(),
	"typescript": typescript.GetLanguage(),
	"yaml":       yaml.GetLanguage(),
}

func prettyPrintLocalCodeIntelPayload(args types.RepoCommitPath, payload types.LocalCodeIntelPayload, contents string) {
	sb := &strings.Builder{}

	blue := color.New(color.FgBlue).SprintFunc()
	fmt.Fprintf(sb, blue("repo %s, commit %s, path %s"), args.Repo, args.Commit, args.Path)
	fmt.Fprintln(sb)
	// gutter := fmt.Sprintf("%5d | ", 3)

	// fmt.Fprintf(sb, "%s%s%s %s\n", gutterPadding, space, arrows, messageColor(breadcrumb.message)(breadcrumb.message))
	fmt.Fprintln(sb)

	fmt.Println(bracket(sb.String()))
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

func newGetIdFunc() func(node *sitter.Node) Id {
	// TODO get the ID directly from tree-sitter for convenience

	// String IDs look like this: <startByte>-<endByte>
	sringId := func(node *sitter.Node) string {
		return fmt.Sprintf("%d-%d", node.StartByte(), node.EndByte())
	}

	nextId := 0
	stringIdToId := map[string]Id{}
	return func(node *sitter.Node) Id {
		if id, ok := stringIdToId[sringId(node)]; ok {
			return id
		}
		stringIdToId[sringId(node)] = nextId
		nextId++
		return stringIdToId[sringId(node)]
	}
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
type Id = int

type Scope = map[string]*PartialSymbol

type PartialSymbol struct {
	Hover *string
	Def   *types.Range
	Refs  []types.Range
}

type Scope2 struct {
	parent  *Scope2
	symbols map[string]types.Symbol
}

func newScope(parent *Scope2) *Scope2 {
	return &Scope2{
		parent:  parent,
		symbols: map[string]types.Symbol{},
	}
}

func (s *Scope2) set(name string, symbol types.Symbol) {
	s.symbols[name] = symbol
}

func (s *Scope2) get(name string) *types.Symbol {
	if symbol, ok := s.symbols[name]; ok {
		return &symbol
	}

	if s.parent != nil {
		return s.parent.get(name)
	}

	return nil
}
