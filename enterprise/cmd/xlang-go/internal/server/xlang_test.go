package server_test

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/enterprise/cmd/xlang-go/internal/server"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	lsext "github.com/sourcegraph/go-langserver/pkg/lspext"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
	"github.com/sourcegraph/sourcegraph/xlang/proxy"
	"github.com/sourcegraph/sourcegraph/xlang/uri"
)

func init() {
	server.UseRemoteFS = true
}

var update = flag.Bool("update", false, "update golden files on disk")

// lspTests runs all test suites for LSP functionality.
func lspTests(t testing.TB, ctx context.Context, c *jsonrpc2.Conn, root *uri.URI, wantHover, wantDefinition, wantXDefinition map[string]string, wantReferences, wantSymbols map[string][]string, wantXDependencies string, wantXReferences map[*lsext.WorkspaceReferencesParams][]string, wantXPackages []string) {
	for pos, want := range wantHover {
		tbRun(t, fmt.Sprintf("hover-%s", strings.Replace(pos, "/", "-", -1)), func(t testing.TB) {
			hoverTest(t, ctx, c, root, pos, want)
		})
	}

	for pos, want := range wantDefinition {
		tbRun(t, fmt.Sprintf("definition-%s", strings.Replace(pos, "/", "-", -1)), func(t testing.TB) {
			definitionTest(t, ctx, c, root, pos, want)
		})
	}
	for pos, want := range wantXDefinition {
		tbRun(t, fmt.Sprintf("xdefinition-%s", strings.Replace(pos, "/", "-", -1)), func(t testing.TB) {
			xdefinitionTest(t, ctx, c, root, pos, want)
		})
	}

	for pos, want := range wantReferences {
		tbRun(t, fmt.Sprintf("references-%s", pos), func(t testing.TB) {
			referencesTest(t, ctx, c, root, pos, want)
		})
	}

	for query, want := range wantSymbols {
		tbRun(t, fmt.Sprintf("symbols(q=%q)", query), func(t testing.TB) {
			symbolsTest(t, ctx, c, root, query, want)
		})
	}

	if wantXDependencies != "" {
		tbRun(t, fmt.Sprintf("xdependencies-"+wantXDependencies), func(t testing.TB) {
			var deps []lspext.DependencyReference
			err := c.Call(ctx, "workspace/xdependencies", struct{}{}, &deps)
			if err != nil {
				t.Fatal(err)
			}
			jsonTest(t, deps, "xdependencies-"+wantXDependencies)
		})
	}

	for params, want := range wantXReferences {
		tbRun(t, fmt.Sprintf("xreferences"), func(t testing.TB) {
			workspaceReferencesTest(t, ctx, c, root, *params, want)
		})
	}

	if wantXPackages != nil {
		tbRun(t, "xpackages", func(t testing.TB) {
			workspacePackagesTest(t, ctx, c, root, wantXPackages)
		})
	}
}

// tbRun calls (testing.T).Run or (testing.B).Run.
func tbRun(t testing.TB, name string, f func(testing.TB)) bool {
	switch tb := t.(type) {
	case *testing.B:
		return tb.Run(name, func(b *testing.B) { f(b) })
	case *testing.T:
		return tb.Run(name, func(t *testing.T) { f(t) })
	default:
		panic(fmt.Sprintf("unexpected %T, want *testing.B or *testing.T", tb))
	}
}

func hoverTest(t testing.TB, ctx context.Context, c *jsonrpc2.Conn, root *uri.URI, pos, want string) {
	file, line, char, err := parsePos(pos)
	if err != nil {
		t.Fatal(err)
	}
	hover, err := callHover(ctx, c, lsp.DocumentURI(root.WithFilePath(filepath.Join(root.FilePath(), file)).String()), line, char)
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasSuffix(want, "...") {
		// Allow specifying expected hover strings with "..." at the
		// end for ease of test creation.
		if len(hover) >= len(want)+3 {
			hover = hover[:len(want)-3] + "..."
		}
	}
	if !strings.Contains(hover, want) {
		t.Fatalf("got %q, want %q", hover, want)
	}
}

func definitionTest(t testing.TB, ctx context.Context, c *jsonrpc2.Conn, root *uri.URI, pos, want string) {
	file, line, char, err := parsePos(pos)
	if err != nil {
		t.Fatal(err)
	}
	definition, err := callDefinition(ctx, c, lsp.DocumentURI(root.WithFilePath(filepath.Join(root.FilePath(), file)).String()), line, char)
	if err != nil {
		t.Fatal(err)
	}
	definition = strings.TrimPrefix(definition, "file:///")
	if definition != want {
		t.Errorf("got %q, want %q", definition, want)
	}
}

func xdefinitionTest(t testing.TB, ctx context.Context, c *jsonrpc2.Conn, root *uri.URI, pos, want string) {
	file, line, char, err := parsePos(pos)
	if err != nil {
		t.Fatal(err)
	}
	xdefinition, err := callXDefinition(ctx, c, lsp.DocumentURI(root.WithFilePath(filepath.Join(root.FilePath(), file)).String()), line, char)
	if err != nil {
		t.Fatal(err)
	}
	xdefinition = strings.TrimPrefix(xdefinition, "file:///")
	if xdefinition != want {
		t.Errorf("got %q, want %q", xdefinition, want)
	}
}

func referencesTest(t testing.TB, ctx context.Context, c *jsonrpc2.Conn, root *uri.URI, pos string, want []string) {
	file, line, char, err := parsePos(pos)
	if err != nil {
		t.Fatal(err)
	}
	references, err := callReferences(ctx, c, lsp.DocumentURI(root.WithFilePath(filepath.Join(root.FilePath(), file)).String()), line, char)
	if err != nil {
		t.Fatal(err)
	}
	for i := range references {
		references[i] = strings.TrimPrefix(references[i], "file:///")
	}
	sort.Strings(references)
	sort.Strings(want)
	if !reflect.DeepEqual(references, want) {
		t.Errorf("got %q, want %q", references, want)
	}
}

func symbolsTest(t testing.TB, ctx context.Context, c *jsonrpc2.Conn, root *uri.URI, query string, want []string) {
	wg := sync.WaitGroup{}
	for i := 0; i < len(query)-1; i++ {
		wg.Add(1)
		go func(q string) {
			defer wg.Done()
			_, _ = callSymbols(ctx, c, q)
		}(query[:i])
		time.Sleep(time.Millisecond * 10)
	}
	symbols, err := callSymbols(ctx, c, query)
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	for i := range symbols {
		symbols[i] = strings.TrimPrefix(symbols[i], "file:///")
	}
	if !reflect.DeepEqual(symbols, want) {
		t.Errorf("got %#v, want %q", symbols, want)
	}
}

func jsonTest(t testing.TB, gotData interface{}, testName string) {
	got, err := json.MarshalIndent(gotData, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	wantFile := filepath.Join("testdata", "want-"+testName)
	want, err := ioutil.ReadFile(wantFile)
	if err != nil {
		t.Fatal(err)
		return
	}
	if strings.TrimSpace(string(got)) != strings.TrimSpace(string(want)) {
		if *update {
			t.Logf("updating %s", wantFile)
			ioutil.WriteFile(wantFile, got, 0777)
			return
		}
		gotFile := filepath.Join("testdata", "got-"+testName)
		ioutil.WriteFile(gotFile, got, 0777)

		cmd := exec.Command("git", "diff", "--color", "--no-index", wantFile, gotFile)
		cmd.Dir, err = os.Getwd()
		if err != nil {
			t.Fatal(err)
			return
		}
		diff, _ := cmd.CombinedOutput()
		if len(diff) > 1024 { // roughly 20 lines
			diff = []byte(string(diff[:1024]) + fmt.Sprintf("\x1B[0m ... (%d bytes omitted)", len(diff)-1024))
		}
		t.Error(string(diff))
		t.Fatalf("\ngot  %s\nwant %s", gotFile, wantFile)
	}
}

func workspaceReferencesTest(t testing.TB, ctx context.Context, c *jsonrpc2.Conn, rootURI *uri.URI, params lsext.WorkspaceReferencesParams, want []string) {
	references, err := callWorkspaceReferences(ctx, c, params)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(references, want) {
		t.Errorf("\ngot  %q\nwant %q", references, want)
	}
}

func workspacePackagesTest(t testing.TB, ctx context.Context, c *jsonrpc2.Conn, rootURI *uri.URI, want []string) {
	packages, err := callWorkspacePackages(ctx, c)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(packages, want) {
		t.Errorf("\ngot  %q\nwant %q", packages, want)
	}
}

func callWorkspacePackages(ctx context.Context, c *jsonrpc2.Conn) ([]string, error) {
	var packages []lspext.PackageInformation
	err := c.Call(ctx, "workspace/xpackages", nil, &packages)
	if err != nil {
		return nil, err
	}
	pkgs := make([]string, len(packages))
	for i, p := range packages {
		pkgs[i] = p.Package["package"].(string)
	}
	return pkgs, nil
}

func callWorkspaceReferences(ctx context.Context, c *jsonrpc2.Conn, params lsext.WorkspaceReferencesParams) ([]string, error) {
	var references []lsext.ReferenceInformation
	err := c.Call(ctx, "workspace/xreferences", params, &references)
	if err != nil {
		return nil, err
	}
	refs := make([]string, len(references))
	for i, r := range references {
		locationPath := strings.TrimPrefix(string(r.Reference.URI), "file://")
		start := r.Reference.Range.Start
		end := r.Reference.Range.End
		refs[i] = fmt.Sprintf("%s:%d:%d-%d:%d -> %v", locationPath, start.Line+1, start.Character+1, end.Line+1, end.Character+1, r.Symbol)
	}
	return refs, nil
}

func parsePos(s string) (file string, line, char int, err error) {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		err = fmt.Errorf("invalid pos %q (%d parts)", s, len(parts))
		return
	}
	file = parts[0]
	line, err = strconv.Atoi(parts[1])
	if err != nil {
		err = fmt.Errorf("invalid line in %q: %s", s, err)
		return
	}
	char, err = strconv.Atoi(parts[2])
	if err != nil {
		err = fmt.Errorf("invalid char in %q: %s", s, err)
		return
	}
	return file, line - 1, char - 1, nil // LSP is 0-indexed
}

func callHover(ctx context.Context, c *jsonrpc2.Conn, uri lsp.DocumentURI, line, char int) (string, error) {
	var res struct {
		Contents markedStrings `json:"contents"`
		lsp.Hover
	}
	err := c.Call(ctx, "textDocument/hover", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: uri},
		Position:     lsp.Position{Line: line, Character: char},
	}, &res)
	if err != nil {
		return "", err
	}
	var str string
	for i, ms := range res.Contents {
		if i != 0 {
			str += " "
		}
		str += ms.Value
	}
	return str, nil
}

func callDefinition(ctx context.Context, c *jsonrpc2.Conn, uri lsp.DocumentURI, line, char int) (string, error) {
	var res locations
	err := c.Call(ctx, "textDocument/definition", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: uri},
		Position:     lsp.Position{Line: line, Character: char},
	}, &res)
	if err != nil {
		return "", err
	}
	var str string
	for i, loc := range res {
		if loc.URI == "" {
			continue
		}
		if i != 0 {
			str += ", "
		}
		str += fmt.Sprintf("%s:%d:%d", loc.URI, loc.Range.Start.Line+1, loc.Range.Start.Character+1)
	}
	return str, nil
}

func callXDefinition(ctx context.Context, c *jsonrpc2.Conn, uri lsp.DocumentURI, line, char int) (string, error) {
	var res []lsext.SymbolLocationInformation
	err := c.Call(ctx, "textDocument/xdefinition", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: uri},
		Position:     lsp.Position{Line: line, Character: char},
	}, &res)
	if err != nil {
		return "", err
	}
	var str string
	for i, loc := range res {
		if loc.Location.URI == "" {
			continue
		}
		if i != 0 {
			str += ", "
		}
		str += fmt.Sprintf("%s:%d:%d %s", loc.Location.URI, loc.Location.Range.Start.Line+1, loc.Location.Range.Start.Character+1, loc.Symbol)
	}
	return str, nil
}

func callReferences(ctx context.Context, c *jsonrpc2.Conn, uri lsp.DocumentURI, line, char int) ([]string, error) {
	var res locations
	err := c.Call(ctx, "textDocument/references", lsp.ReferenceParams{
		Context: lsp.ReferenceContext{IncludeDeclaration: true},
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: uri},
			Position:     lsp.Position{Line: line, Character: char},
		},
	}, &res)
	if err != nil {
		return nil, err
	}
	str := make([]string, len(res))
	for i, loc := range res {
		str[i] = fmt.Sprintf("%s:%d:%d", loc.URI, loc.Range.Start.Line+1, loc.Range.Start.Character+1)
	}
	return str, nil
}

var symbolKindName = map[lsp.SymbolKind]string{
	lsp.SKFile:        "file",
	lsp.SKModule:      "module",
	lsp.SKNamespace:   "namespace",
	lsp.SKPackage:     "package",
	lsp.SKClass:       "class",
	lsp.SKMethod:      "method",
	lsp.SKProperty:    "property",
	lsp.SKField:       "field",
	lsp.SKConstructor: "constructor",
	lsp.SKEnum:        "enum",
	lsp.SKInterface:   "interface",
	lsp.SKFunction:    "function",
	lsp.SKVariable:    "variable",
	lsp.SKConstant:    "constant",
	lsp.SKString:      "string",
	lsp.SKNumber:      "number",
	lsp.SKBoolean:     "boolean",
	lsp.SKArray:       "array",
}

func callSymbols(ctx context.Context, c *jsonrpc2.Conn, query string) ([]string, error) {
	var symbols []lsp.SymbolInformation
	err := c.Call(ctx, "workspace/symbol", lsp.WorkspaceSymbolParams{Query: query}, &symbols)
	if err != nil {
		return nil, err
	}
	syms := make([]string, len(symbols))
	for i, s := range symbols {
		syms[i] = fmt.Sprintf("%s:%s:%s:%d:%d", s.Location.URI, symbolKindName[s.Kind], qualifiedName(s), s.Location.Range.Start.Line, s.Location.Range.Start.Character)
	}
	return syms, nil
}

func qualifiedName(s lsp.SymbolInformation) string {
	if s.ContainerName != "" {
		return s.ContainerName + "." + s.Name
	}
	return s.Name
}

type markedStrings []lsp.MarkedString

func (v *markedStrings) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return errors.New("invalid empty JSON")
	}
	if data[0] == '[' {
		var ms []markedString
		if err := json.Unmarshal(data, &ms); err != nil {
			return err
		}
		for _, ms := range ms {
			*v = append(*v, lsp.MarkedString(ms))
		}
		return nil
	}
	*v = []lsp.MarkedString{{}}
	return json.Unmarshal(data, &(*v)[0])
}

type markedString lsp.MarkedString

func (v *markedString) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return errors.New("invalid empty JSON")
	}
	if data[0] == '{' {
		return json.Unmarshal(data, (*lsp.MarkedString)(v))
	}

	// String
	*v = markedString{}
	return json.Unmarshal(data, &v.Value)
}

type locations []lsp.Location

func (v *locations) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return errors.New("invalid empty JSON")
	}
	if data[0] == '[' {
		return json.Unmarshal(data, (*[]lsp.Location)(v))
	}
	*v = []lsp.Location{{}}
	return json.Unmarshal(data, &(*v)[0])
}

// testRequest is a simplified version of jsonrpc2.Request for easier
// test expectation definition and checking of the fields that matter.
type testRequest struct {
	Method string
	Params interface{}
}

func (r testRequest) String() string {
	b, err := json.Marshal(r.Params)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s(%s)", r.Method, b)
}

func testRequestEqual(a, b testRequest) bool {
	if a.Method != b.Method {
		return false
	}

	// We want to see if a and b have identical canonical JSON
	// representations. They are NOT identical Go structures, since
	// one comes from the wire (as raw JSON) and one is an interface{}
	// of a concrete struct/slice type provided as a test expectation.
	ajson, err := json.Marshal(a.Params)
	if err != nil {
		panic(err)
	}
	bjson, err := json.Marshal(b.Params)
	if err != nil {
		panic(err)
	}
	var a2, b2 interface{}
	if err := json.Unmarshal(ajson, &a2); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(bjson, &b2); err != nil {
		panic(err)
	}
	return reflect.DeepEqual(a2, b2)
}

func testRequestsEqual(as, bs []testRequest) bool {
	if len(as) != len(bs) {
		return false
	}
	for i, a := range as {
		if !testRequestEqual(a, bs[i]) {
			return false
		}
	}
	return true
}

type testRequests []testRequest

func (v testRequests) Len() int      { return len(v) }
func (v testRequests) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v testRequests) Less(i, j int) bool {
	ii, err := json.Marshal(v[i])
	if err != nil {
		panic(err)
	}
	jj, err := json.Marshal(v[j])
	if err != nil {
		panic(err)
	}
	return string(ii) < string(jj)
}

func useMapFS(m map[string]string) func() {
	orig := proxy.NewRemoteRepoVFS
	proxy.NewRemoteRepoVFS = func(ctx context.Context, cloneURL *url.URL, commitID api.CommitID) (proxy.FileSystem, error) {
		return mapFS(m), nil
	}
	return func() { proxy.NewRemoteRepoVFS = orig }
}

// mapFS lets us easily instantiate a VFS with a map[string]string
// (which is less noisy than map[string][]byte in test fixtures).
func mapFS(m map[string]string) *stringMapFS {
	m2 := make(map[string][]byte, len(m))
	filenames := make([]string, 0, len(m))
	for k, v := range m {
		m2[k] = []byte(v)
		filenames = append(filenames, k)
	}
	return &stringMapFS{
		FileSystem: ctxvfs.Map(m2),
		filenames:  filenames,
	}
}

type stringMapFS struct {
	ctxvfs.FileSystem
	filenames []string
}

func (fs *stringMapFS) ListAllFiles(ctx context.Context) ([]string, error) {
	return fs.filenames, nil
}
