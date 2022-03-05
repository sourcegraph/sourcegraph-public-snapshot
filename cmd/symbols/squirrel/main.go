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

func DefinitionHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log15.Error("failed to read request body", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var loc types.SquirrelLocation
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&loc); err != nil {
		log15.Error("failed to decode request body", "err", err, "body", string(body))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	repo := loc.Repo
	commit := loc.Commit
	path := loc.Path
	row := loc.Row
	column := loc.Column

	debug := os.Getenv("SQUIRREL_DEBUG") == "true"

	if debug {
		fmt.Println("👉 repo:", repo, "commit:", commit, "path:", path, "row:", row, "column:", column)
	}

	readFile := func(RepoCommitPath) ([]byte, error) {
		cmd := gitserver.DefaultClient.Command("git", "cat-file", "blob", commit+":"+path)
		cmd.Repo = api.RepoName(repo)
		stdout, stderr, err := cmd.DividedOutput(r.Context())
		if err != nil {
			return nil, fmt.Errorf("failed to get file contents: %s\n\nstdout:\n\n%s\n\nstderr:\n\n%s", err, stdout, stderr)
		}
		return stdout, nil
	}

	squirrel := NewSquirrel(readFile)

	result, breadcrumbs, err := squirrel.definition(Location{
		RepoCommitPath: RepoCommitPath{
			Repo:   repo,
			Commit: commit,
			Path:   path},
		Row:    uint32(row),
		Column: uint32(column),
	})
	if breadcrumbs != nil && debug {
		prettyPrintBreadcrumbs(pickBreadcrumbs(breadcrumbs, []string{"start", "found"}), readFile)
	}
	if err != nil {
		_ = json.NewEncoder(w).Encode(nil)
		log15.Error("failed to get definition", "err", err)
		return
	}

	if debug {
		fmt.Println("✅ repo:", result.Repo, "commit:", result.Commit, "path:", result.Path, "row:", result.Row, "column:", result.Column)
	}

	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		log15.Error("failed to write response: %s", err)
		http.Error(w, fmt.Sprintf("failed to get definition: %s", err), http.StatusInternalServerError)
	}
}

type RepoCommitPath struct {
	Repo   string `json:"repo"`
	Commit string `json:"commit"`
	Path   string `json:"path"`
}

type Location struct {
	RepoCommitPath
	Row    uint32 `json:"row"`
	Column uint32 `json:"column"`
}

type ReadFileFunc func(RepoCommitPath) ([]byte, error)

type Squirrel struct {
	readFile ReadFileFunc
}

func NewSquirrel(readFile ReadFileFunc) *Squirrel {
	return &Squirrel{readFile: readFile}
}

func (s *Squirrel) definition(location Location) (*Location, []Breadcrumb, error) {
	parser := sitter.NewParser()

	ext := strings.TrimPrefix(filepath.Ext(location.Path), ".")

	langName, ok := extToLang[ext]
	if !ok {
		return nil, nil, errors.Newf("unrecognized file extension %s", ext)
	}

	nvimDir, ok := langToNvimQueryDir[langName]
	if !ok {
		return nil, nil, errors.Newf("neovim-treesitter does not have queries for the language %s", langName)
	}

	localsPath := path.Join("nvim-treesitter", "queries", nvimDir, "locals.scm")
	queriesBytes, err := queriesFs.ReadFile(localsPath)
	if err != nil {
		return nil, nil, errors.Newf("could not read %d: %s", localsPath, err)
	}
	queryString := string(queriesBytes)

	lang, ok := langToSitterLanguage[langName]
	if !ok {
		return nil, nil, fmt.Errorf("no tree-sitter parser for language %s", langName)
	}

	parser.SetLanguage(lang)

	input, err := s.readFile(location.RepoCommitPath)
	if err != nil {
		return nil, nil, err
	}

	tree, err := parser.ParseCtx(context.Background(), nil, input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse file contents: %s", err)
	}
	root := tree.RootNode()

	startNode := root.NamedDescendantForPointRange(
		sitter.Point{Row: location.Row, Column: location.Column},
		sitter.Point{Row: location.Row, Column: location.Column},
	)

	if startNode == nil {
		return nil, nil, errors.New("node is nil")
	}

	breadcrumbs := []Breadcrumb{{
		Location: location,
		length:   1,
		message:  "start",
	}}

	// Execute the query
	query, err := sitter.NewQuery([]byte(queryString), lang)
	if err != nil {
		return nil, breadcrumbs, errors.Newf("failed to parse query: %s\n%s", err, queryString)
	}
	cursor := sitter.NewQueryCursor()
	cursor.Exec(query, root)

	// Collect all definitions into scopes
	scopes := map[string][]*sitter.Node{}
	match, _, hasCapture := cursor.NextCapture()
	for hasCapture {
		for _, capture := range match.Captures {
			name := query.CaptureNameForId(capture.Index)

			// Add to breadcrumbs
			length := 1
			if capture.Node.EndPoint().Row == capture.Node.StartPoint().Row && name != "scope" {
				length = int(capture.Node.EndPoint().Column - capture.Node.StartPoint().Column)
			}
			breadcrumbs = append(breadcrumbs, Breadcrumb{
				Location: Location{
					RepoCommitPath: location.RepoCommitPath,
					Row:            capture.Node.StartPoint().Row,
					Column:         capture.Node.StartPoint().Column,
				},
				length:  length,
				message: fmt.Sprintf("%s: %s", name, capture.Node.Type()),
			})

			// Add definition to nearest scope
			if strings.Contains(name, "definition") {
				for cur := capture.Node; cur != nil; cur = cur.Parent() {
					id := getId(cur)
					_, ok := scopes[id]
					if !ok {
						continue
					}
					scopes[id] = append(scopes[id], capture.Node)
					break
				}

				continue
			}

			// Record the scope
			if strings.Contains(name, "scope") {
				scopes[getId(capture.Node)] = []*sitter.Node{}
				continue
			}
		}

		// Next capture
		match, _, hasCapture = cursor.NextCapture()
	}

	// Walk up the tree to find the nearest definition
	for currentNode := startNode; currentNode != nil; currentNode = currentNode.Parent() {
		scope, ok := scopes[getId(currentNode)]
		if !ok {
			// This node isn't a scope, continue.
			continue
		}

		// Check if the scope contains the definition
		for _, def := range scope {
			if def.Content(input) == startNode.Content(input) {
				found := Location{
					RepoCommitPath: location.RepoCommitPath,
					Row:            def.StartPoint().Row,
					Column:         def.StartPoint().Column,
				}

				breadcrumbs := append(breadcrumbs, Breadcrumb{
					Location: found,
					length:   1,
					message:  "found",
				})

				return &found, breadcrumbs, nil
			}
		}
	}

	return nil, breadcrumbs, errors.New("could not find definition")
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

type Breadcrumb struct {
	Location
	length  int
	message string
}

// IDs are <startByte>-<endByte> as a proxy for node ID
func getId(node *sitter.Node) string {
	return fmt.Sprintf("%d-%d", node.StartByte(), node.EndByte())
}

func prettyPrintBreadcrumbs(breadcrumbs []Breadcrumb, readFile ReadFileFunc) {
	sb := &strings.Builder{}

	m := map[RepoCommitPath]map[int][]Breadcrumb{}
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
		fmt.Fprintf(sb, blue("repo %s, commit %s, path %s"), repoCommitPath.Repo, repoCommitPath.Commit, repoCommitPath.Path)
		fmt.Fprintln(sb)

		contents, err := readFile(repoCommitPath)
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

			fmt.Fprintln(sb)

			gutter := fmt.Sprintf("%5d | ", lineNumber)

			columnToMessage := map[int]string{}
			for _, breadcrumb := range breadcrumbs {
				for column := int(breadcrumb.Column); column < int(breadcrumb.Column)+breadcrumb.length; column++ {
					columnToMessage[lengthInSpaces(line[:column])] = breadcrumb.message
				}

				gutterPadding := strings.Repeat(" ", len(gutter))

				space := strings.Repeat(" ", lengthInSpaces(line[:breadcrumb.Column]))

				arrows := messageColor(breadcrumb.message)(strings.Repeat("v", breadcrumb.length))

				fmt.Fprintf(sb, "%s%s%s %s\n", gutterPadding, space, arrows, messageColor(breadcrumb.message)(breadcrumb.message))
			}

			fmt.Fprint(sb, grey(gutter))
			lineWithSpaces := strings.ReplaceAll(line, "\t", "    ")
			for c := 0; c < len(lineWithSpaces); c++ {
				if message, ok := columnToMessage[c]; ok {
					fmt.Fprint(sb, messageColor(message)(string(lineWithSpaces[c])))
				} else {
					fmt.Fprint(sb, grey(string(lineWithSpaces[c])))
				}
			}
			fmt.Fprintln(sb)
		}
	}

	fmt.Println(bracket(sb.String()))
}

type colorSprintfFunc func(a ...interface{}) string

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
