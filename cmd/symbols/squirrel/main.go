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

	symbolsTypes "github.com/sourcegraph/sourcegraph/cmd/symbols/types"
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

	result, err := localCodeIntel(r.Context(), args, readFileFromGitserver)
	if result != nil && debug {
		fmt.Fprintln(debugStringBuilder, "👉 repo:", repo, "commit:", commit, "path:", path)
		contents, err := readFileFromGitserver(r.Context(), args)
		if err != nil {
			log15.Error("failed to read file from gitserver", "err", err)
		} else {
			prettyPrintLocalCodeIntelPayload(debugStringBuilder, args, *result, string(contents))
			fmt.Fprintln(debugStringBuilder, "✅ repo:", repo, "commit:", commit, "path:", path)

			fmt.Println(" ")
			fmt.Println(bracket(debugStringBuilder.String()))
			fmt.Println(" ")
		}
	}
	if err != nil {
		_ = json.NewEncoder(w).Encode(nil)
		log15.Error("failed to generate local code intel payload", "err", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		log15.Error("failed to write response: %s", "error", err)
		http.Error(w, fmt.Sprintf("failed to generate local code intel payload: %s", err), http.StatusInternalServerError)
		return
	}
}

func NewSymbolInfoHandler(symbolSearch symbolsTypes.SearchFunc) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log15.Error("failed to read request body", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var args types.RepoCommitPathPoint
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&args); err != nil {
			log15.Error("failed to decode request body", "err", err, "body", string(body))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		repo := args.Repo
		commit := args.Commit
		path := args.Path
		row := args.Row
		column := args.Column

		debug := os.Getenv("SQUIRREL_DEBUG") == "true"
		debugStringBuilder := &strings.Builder{}

		squirrel := NewSquirrel(readFileFromGitserver, symbolSearch)

		result, err := squirrel.symbolInfo(r.Context(), args)
		if debug {
			fmt.Fprintln(debugStringBuilder, "👉 repo:", repo, "commit:", commit, "path:", path, "row:", row, "column:", column)
			prettyPrintBreadcrumbs(debugStringBuilder, squirrel.breadcrumbs, readFileFromGitserver)
			if result == nil {
				fmt.Fprintln(debugStringBuilder, "❌ no definition found")
			} else {
				fmt.Fprintln(debugStringBuilder, "✅ found definition: ", *result)
			}

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
}

func localCodeIntel(ctx context.Context, repoCommitPath types.RepoCommitPath, readFile ReadFileFunc) (*types.LocalCodeIntelPayload, error) {
	root, contents, langSpec, err := parse(ctx, repoCommitPath, readFile)

	localsPath := path.Join("nvim-treesitter", "queries", langSpec.nvimQueryDir, "locals.scm")
	queriesBytes, err := queriesFs.ReadFile(localsPath)
	if err != nil {
		return nil, errors.Newf("could not read %d: %s", localsPath, err)
	}
	queryString := string(queriesBytes)

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
					symbolName := node.Content(contents)

					// Print a debug message if the symbol is already defined.
					if symbol, ok := scope[symbolName]; ok && debug {
						lines := strings.Split(string(contents), "\n")
						fmt.Printf("duplicate definition for %q in %s (using second)\n", symbolName, repoCommitPath.Path)
						fmt.Printf("  %4d | %s\n", symbol.Def.Row, lines[symbol.Def.Row])
						fmt.Printf("  %4d | %s\n", node.StartPoint().Row, lines[node.StartPoint().Row])
					}

					// Get the hover.
					hover := getHover(node, langSpec.commentStyle, string(contents))

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

type ReadFileFunc func(context.Context, types.RepoCommitPath) ([]byte, error)

type Squirrel struct {
	readFile     ReadFileFunc
	symbolSearch symbolsTypes.SearchFunc
	breadcrumbs  []Breadcrumb
}

func NewSquirrel(readFile ReadFileFunc, symbolSearch symbolsTypes.SearchFunc) *Squirrel {
	return &Squirrel{
		readFile:     readFile,
		symbolSearch: symbolSearch,
		breadcrumbs:  []Breadcrumb{},
	}
}

func (s *Squirrel) symbolInfo(ctx context.Context, location types.RepoCommitPathPoint) (*types.SymbolInfo, error) {
	def, err := s.getDefAtLocation(ctx, location)
	if err != nil {
		return nil, err
	}
	if def == nil {
		return nil, nil
	}

	hover, err := s.getHoverOnLine(ctx, *def)
	if err != nil {
		log15.Error("failed to get hover on a line", "error", err, "location", location, "deflocation", def)
	}

	return &types.SymbolInfo{
		Definition: *def,
		Hover:      hover,
	}, nil
}

func (s *Squirrel) getDefAtLocation(ctx context.Context, point types.RepoCommitPathPoint) (*types.RepoCommitPathRange, error) {
	root, _, langSpec, err := parse(ctx, point.RepoCommitPath, s.readFile)
	if err != nil {
		return nil, err
	}

	startNode := root.NamedDescendantForPointRange(
		sitter.Point{Row: uint32(point.Row), Column: uint32(point.Column)},
		sitter.Point{Row: uint32(point.Row), Column: uint32(point.Column)},
	)

	if startNode == nil {
		return nil, errors.New("node is nil")
	}

	foundPkgOrNode, err := s.getDef(ctx, langSpec.language, point.RepoCommitPath, startNode)
	if err != nil {
		return nil, err
	}
	if foundPkgOrNode == nil {
		return nil, nil
	}

	if foundPkgOrNode.Node != nil {
		return &types.RepoCommitPathRange{
			RepoCommitPath: foundPkgOrNode.Node.RepoCommitPath,
			Range:          nodeToRange(foundPkgOrNode.Node.Node),
		}, nil
	}

	return nil, nil
}

type PkgOrNode struct {
	Pkg  *types.RepoCommitPath
	Node *NodeWithRepoCommitPath
}

type NodeWithRepoCommitPath struct {
	RepoCommitPath types.RepoCommitPath
	Node           *sitter.Node
}

func (s *Squirrel) getDef(ctx context.Context, lang *sitter.Language, repoCommitPath types.RepoCommitPath, node *sitter.Node) (*PkgOrNode, error) {
	if node == nil {
		return nil, nil
	}

	s.breadcrumbs = append(s.breadcrumbs, Breadcrumb{
		RepoCommitPathRange: types.RepoCommitPathRange{
			RepoCommitPath: repoCommitPath,
			Range:          nodeToRange(node),
		},
		length:  nodeLength(node),
		message: "getDef",
	})

	contents, err := s.readFile(ctx, repoCommitPath)
	if err != nil {
		return nil, err
	}

	switch node.Type() {
	case "identifier":
		for cur := node; cur != nil; cur = cur.Parent() {
			parent := cur.Parent()
			if parent == nil {
				break
			}
			switch parent.Type() {
			case "block":
				for cur2 := cur; cur2 != nil; cur2 = cur2.PrevNamedSibling() {
					if cur2.Type() == "var_declaration" {
						if cur2.NamedChild(0).Type() == "var_spec" {
							found := cur2.NamedChild(0).ChildByFieldName("name")
							if found.Content(contents) == node.Content(contents) {
								return &PkgOrNode{Node: &NodeWithRepoCommitPath{RepoCommitPath: repoCommitPath, Node: found}}, nil
							}
						}
					}
				}
			}
		}
	case "type_identifier":
		parent := node.Parent()
		if parent == nil {
			break
		}
		switch parent.Type() {
		case "qualified_type":
			return s.getField(ctx, lang, repoCommitPath, parent.ChildByFieldName("package"), node.Content(contents))
		default:
			return nil, errors.Newf("unrecognized parent type %s", parent.Type())
		}
	case "field_identifier":
		parent := node.Parent()
		if parent == nil {
			return nil, nil
		}

		switch parent.Type() {
		case "selector_expression":
			return s.getField(ctx, lang, repoCommitPath, parent.ChildByFieldName("operand"), node.Content(contents))
		default:
			return nil, errors.Newf("unexpected parent type %s", parent.Type())
		}
	case "package_identifier":
		top := getRoot(node)
		pkg := node.Content(contents)
		dir := ""
		forEachCapture("(import_spec path: (interpreted_string_literal) @import)", top, lang, func(name string, node *sitter.Node) {
			path := node.Content(contents)
			path = strings.TrimPrefix(path, `"`)
			path = strings.TrimSuffix(path, `"`)

			if !strings.HasSuffix(path, "/"+pkg) {
				return
			}

			components := strings.Split(path, "/")
			if len(components) < 3 {
				return
			}

			dir = strings.Join(components[3:], "/")
		})
		if dir == "" {
			return nil, nil
		}
		return &PkgOrNode{Pkg: &types.RepoCommitPath{
			Repo:   repoCommitPath.Repo,
			Commit: repoCommitPath.Commit,
			Path:   dir,
		}}, nil
	}

	return nil, nil
}

func (s *Squirrel) getField(ctx context.Context, lang *sitter.Language, repoCommitPath types.RepoCommitPath, node *sitter.Node, field string) (*PkgOrNode, error) {
	if node == nil {
		return nil, nil
	}

	s.breadcrumbs = append(s.breadcrumbs, Breadcrumb{
		RepoCommitPathRange: types.RepoCommitPathRange{
			RepoCommitPath: repoCommitPath,
			Range:          nodeToRange(node),
		},
		length:  nodeLength(node),
		message: fmt.Sprintf("getField(%s)", field),
	})

	typePkgOrNode, err := s.getTypeDef(ctx, lang, repoCommitPath, node)
	if err != nil {
		return nil, err
	}
	if typePkgOrNode == nil {
		return nil, nil
	}

	if typePkgOrNode.Pkg != nil {
		result, err := s.getDefInRepoDir(ctx, typePkgOrNode.Pkg.Repo, typePkgOrNode.Pkg.Commit, field, typePkgOrNode.Pkg.Path)
		if err != nil {
			return nil, err
		}
		return &PkgOrNode{Node: result}, nil
	}

	typeDef := typePkgOrNode.Node.Node

	parent := typeDef.Parent()
	if parent == nil {
		return nil, nil
	}
	switch parent.Type() {
	case "type_spec":
		ty := parent.ChildByFieldName("type")
		if ty == nil {
			return nil, nil
		}

		contents, err := s.readFile(ctx, typePkgOrNode.Node.RepoCommitPath)
		if err != nil {
			return nil, err
		}

		var foundMethod *sitter.Node
		forEachCapture("(method_declaration name: (field_identifier) @method)", getRoot(ty), lang, func(captureName string, node *sitter.Node) {
			if node.Content(contents) == field {
				foundMethod = node
			}
		})
		if foundMethod == nil {
			return nil, nil
		}
		return &PkgOrNode{Node: &NodeWithRepoCommitPath{RepoCommitPath: typePkgOrNode.Node.RepoCommitPath, Node: foundMethod}}, nil
	default:
		return nil, errors.Newf("unrecognized type %s", typeDef.Type())
	}
}

func (s *Squirrel) getTypeDef(ctx context.Context, lang *sitter.Language, repoCommitPath types.RepoCommitPath, node *sitter.Node) (*PkgOrNode, error) {
	if node == nil {
		return nil, nil
	}

	s.breadcrumbs = append(s.breadcrumbs, Breadcrumb{
		RepoCommitPathRange: types.RepoCommitPathRange{
			RepoCommitPath: repoCommitPath,
			Range:          nodeToRange(node),
		},
		length:  nodeLength(node),
		message: "getTypeDef",
	})

	_, err := s.readFile(ctx, repoCommitPath)
	if err != nil {
		return nil, err
	}

	defPkgOrNode, err := s.getDef(ctx, lang, repoCommitPath, node)
	if err != nil {
		return nil, err
	}
	if defPkgOrNode == nil {
		return nil, nil
	}

	if defPkgOrNode.Pkg != nil {
		return defPkgOrNode, nil
	}

	def := defPkgOrNode.Node.Node

	parent := def.Parent()
	if parent == nil {
		return nil, nil
	}

	switch parent.Type() {
	case "var_spec":
		ty := parent.ChildByFieldName("type")
		if ty == nil {
			return nil, nil
		}
		if ty.Type() == "pointer_type" {
			ty = ty.NamedChild(0)
			if ty == nil {
				return nil, nil
			}
		}
		switch ty.Type() {
		case "qualified_type":
			return s.getTypeDef(ctx, lang, defPkgOrNode.Node.RepoCommitPath, ty.ChildByFieldName("name"))
		}
	case "type_spec":
		return defPkgOrNode, nil
	default:
		return nil, errors.Newf("unrecognized parent type %s", parent.Type())
	}

	return nil, nil
}

func (s *Squirrel) getDefInRepoDir(ctx context.Context, repo, commit, symbolName, dir string) (*NodeWithRepoCommitPath, error) {
	defSymbols, err := s.symbolSearch(ctx, symbolsTypes.SearchArgs{
		Repo:            api.RepoName(repo),
		CommitID:        api.CommitID(commit),
		Query:           fmt.Sprintf("^%s$", symbolName),
		IsRegExp:        true,
		IsCaseSensitive: true,
		IncludePatterns: []string{"^" + dir},
		ExcludePattern:  "",
		First:           1,
	})
	if err != nil {
		return nil, err
	}

	if len(defSymbols) == 0 {
		return nil, nil
	}

	defSymbol := defSymbols[0]

	def := types.RepoCommitPathRange{
		RepoCommitPath: types.RepoCommitPath{
			Repo:   repo,
			Commit: commit,
			Path:   defSymbol.Path,
		},
		Range: types.Range{
			Row:    int(defSymbol.Line - 1),
			Column: 0, // TODO symbol search should also return the character
			Length: len(defSymbol.Name),
		},
	}

	contents, err := s.readFile(ctx, def.RepoCommitPath)
	if err != nil {
		return nil, err
	}

	root, _, _, err := parse(ctx, def.RepoCommitPath, s.readFile)
	lines := strings.Split(string(contents), "\n")
	column := strings.Index(lines[def.Range.Row], defSymbol.Name)
	if column == -1 {
		return nil, nil
	}

	node := root.NamedDescendantForPointRange(
		sitter.Point{Row: uint32(def.Range.Row), Column: uint32(column)},
		sitter.Point{Row: uint32(def.Range.Row), Column: uint32(column)},
	)

	if node == nil {
		return nil, nil
	}

	return &NodeWithRepoCommitPath{RepoCommitPath: def.RepoCommitPath, Node: node}, nil
}

func (s *Squirrel) getHoverOnLine(ctx context.Context, rnge types.RepoCommitPathRange) (*string, error) {
	root, endContents, langSpec, err := parse(ctx, rnge.RepoCommitPath, s.readFile)
	if err != nil {
		return nil, err
	}

	endNode := root.NamedDescendantForPointRange(
		sitter.Point{Row: uint32(rnge.Row), Column: uint32(rnge.Column)},
		sitter.Point{Row: uint32(rnge.Row), Column: uint32(rnge.Column)},
	)
	if endNode == nil {
		return nil, errors.Newf("no node at %d:%d", rnge.Row, rnge.Column)
	}

	return getHover(endNode, langSpec.commentStyle, string(endContents)), nil
}

func parse(ctx context.Context, repoCommitPath types.RepoCommitPath, readFile ReadFileFunc) (*sitter.Node, []byte, *LangSpec, error) {
	ext := strings.TrimPrefix(filepath.Ext(repoCommitPath.Path), ".")

	langName, ok := extToLang[ext]
	if !ok {
		return nil, nil, nil, errors.Newf("unrecognized file extension %s", ext)
	}

	langSpec, ok := langToLangSpec[langName]
	if !ok {
		return nil, nil, nil, errors.Newf("unsupported language %s", langName)
	}

	parser := sitter.NewParser()
	parser.SetLanguage(langSpec.language)

	contents, err := readFile(ctx, repoCommitPath)
	if err != nil {
		return nil, nil, nil, err
	}

	tree, err := parser.ParseCtx(context.Background(), nil, contents)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse file contents: %s", err)
	}

	root := tree.RootNode()
	if root == nil {
		return nil, nil, nil, errors.New("root is nil")
	}

	return root, contents, &langSpec, nil
}

type Breadcrumb struct {
	types.RepoCommitPathRange
	length  int
	message string
}

func prettyPrintBreadcrumbs(w *strings.Builder, breadcrumbs []Breadcrumb, readFile ReadFileFunc) {
	m := map[types.RepoCommitPath]map[int][]Breadcrumb{}
	for _, breadcrumb := range breadcrumbs {
		path := breadcrumb.RepoCommitPath

		if _, ok := m[path]; !ok {
			m[path] = map[int][]Breadcrumb{}
		}

		m[path][int(breadcrumb.Row)] = append(m[path][int(breadcrumb.Row)], breadcrumb)
	}

	for repoCommitPath, lineToBreadcrumb := range m {
		blue := color.New(color.FgBlue).SprintFunc()
		grey := color.New(color.FgBlack).SprintFunc()
		fmt.Fprintf(w, blue("repo %s, commit %s, path %s"), repoCommitPath.Repo, repoCommitPath.Commit, repoCommitPath.Path)
		fmt.Fprintln(w)

		contents, err := readFile(context.Background(), repoCommitPath)
		if err != nil {
			fmt.Println("Error reading file: ", err)
			return
		}
		lines := strings.Split(string(contents), "\n")
		for lineNumber, line := range lines {
			breadcrumbs, ok := lineToBreadcrumb[lineNumber]
			if !ok {
				continue
			}

			fmt.Fprintln(w)

			gutter := fmt.Sprintf("%5d | ", lineNumber)

			columnToMessage := map[int]string{}
			for _, breadcrumb := range breadcrumbs {
				for column := int(breadcrumb.Column); column < int(breadcrumb.Column)+breadcrumb.length; column++ {
					columnToMessage[lengthInSpaces(line[:column])] = breadcrumb.message
				}

				gutterPadding := strings.Repeat(" ", len(gutter))

				space := strings.Repeat(" ", lengthInSpaces(line[:breadcrumb.Column]))

				arrows := messageColor(breadcrumb.message)(strings.Repeat("v", breadcrumb.length))

				fmt.Fprintf(w, "%s%s%s %s\n", gutterPadding, space, arrows, messageColor(breadcrumb.message)(breadcrumb.message))
			}

			fmt.Fprint(w, grey(gutter))
			lineWithSpaces := strings.ReplaceAll(line, "\t", "    ")
			for c := 0; c < len(lineWithSpaces); c++ {
				if message, ok := columnToMessage[c]; ok {
					fmt.Fprint(w, messageColor(message)(string(lineWithSpaces[c])))
				} else {
					fmt.Fprint(w, grey(string(lineWithSpaces[c])))
				}
			}
			fmt.Fprintln(w)
		}
	}
}

func pickBreadcrumbs(breadcrumbs []Breadcrumb, messages []string) []Breadcrumb {
	var picked []Breadcrumb
	for _, breadcrumb := range breadcrumbs {
		for _, message := range messages {
			if strings.Contains(breadcrumb.message, message) {
				picked = append(picked, breadcrumb)
				break
			}
		}
	}
	return picked
}

func messageColor(message string) colorSprintfFunc {
	if message == "start" {
		return color.New(color.FgHiCyan).SprintFunc()
	} else if message == "found" {
		return color.New(color.FgRed).SprintFunc()
	} else if message == "correct" {
		return color.New(color.FgGreen).SprintFunc()
	} else if strings.Contains(message, "scope") {
		return color.New(color.FgHiYellow).SprintFunc()
	} else {
		return color.New(color.FgHiMagenta).SprintFunc()
	}
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

func forEachCapture(query string, root *sitter.Node, lang *sitter.Language, f func(captureName string, node *sitter.Node)) error {
	sitterQuery, err := sitter.NewQuery([]byte(query), lang)
	if err != nil {
		return errors.Newf("failed to parse query: %s\n%s", err, query)
	}
	cursor := sitter.NewQueryCursor()
	cursor.Exec(sitterQuery, root)

	match, _, hasCapture := cursor.NextCapture()
	for hasCapture {
		for _, capture := range match.Captures {
			captureName := sitterQuery.CaptureNameForId(capture.Index)
			f(captureName, capture.Node)
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

func nodeLength(node *sitter.Node) int {
	length := 1
	if node.StartPoint().Row == node.EndPoint().Row {
		length = int(node.EndPoint().Column - node.StartPoint().Column)
	}
	return length
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

func readFileFromGitserver(ctx context.Context, repoCommitPath types.RepoCommitPath) ([]byte, error) {
	cmd := gitserver.DefaultClient.Command("git", "cat-file", "blob", repoCommitPath.Commit+":"+repoCommitPath.Path)
	cmd.Repo = api.RepoName(repoCommitPath.Repo)
	stdout, stderr, err := cmd.DividedOutput(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get file contents: %s\n\nstdout:\n\n%s\n\nstderr:\n\n%s", err, stdout, stderr)
	}
	return stdout, nil
}

func getRoot(node *sitter.Node) *sitter.Node {
	var top *sitter.Node
	for cur := node; cur != nil; cur = cur.Parent() {
		top = cur
	}
	return top
}
