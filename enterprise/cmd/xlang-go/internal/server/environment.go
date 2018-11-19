package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"

	log15 "gopkg.in/inconshreveable/log15.v2"
	yaml "gopkg.in/yaml.v2"

	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/go-langserver/langserver"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/go-lsp/lspext"
	"github.com/sourcegraph/jsonx"
	"github.com/sourcegraph/sourcegraph/pkg/gituri"
)

const (
	gopath     = "/"
	goroot     = "/goroot"
	gocompiler = "gc"

	// TODO(sqs): allow these to be customized. They're
	// fine for now, though.
	goos   = "linux"
	goarch = "amd64"
)

// determineEnvironment will setup the language server InitializeParams based
// what it can detect from the filesystem and what it received from the client's
// InitializeParams.
//
// It is expected that fs will be mounted at InitializeParams.RootURI.
func determineEnvironment(ctx context.Context, fs ctxvfs.FileSystem, params lspext.InitializeParams) (*langserver.InitializeParams, error) {
	rootImportPath, err := determineRootImportPath(ctx, params.OriginalRootURI, fs)
	if err != nil {
		return nil, fmt.Errorf("unable to determine workspace's root Go import path: %s (original rootPath is %q)", err, params.OriginalRootURI)
	}
	// Sanity-check the import path.
	if rootImportPath == "" || rootImportPath != path.Clean(rootImportPath) || strings.Contains(rootImportPath, "..") || strings.HasPrefix(rootImportPath, string(os.PathSeparator)) || strings.HasPrefix(rootImportPath, "/") || strings.HasPrefix(rootImportPath, ".") {
		return nil, fmt.Errorf("empty or suspicious import path: %q", rootImportPath)
	}

	// Put all files in the workspace under a /src/IMPORTPATH
	// directory, such as /src/github.com/foo/bar, so that Go can
	// build it in GOPATH=/.
	var rootPath string
	if rootImportPath == "github.com/golang/go" {
		// stdlib means our rootpath is the GOPATH
		rootPath = goroot
		rootImportPath = ""
	} else {
		rootPath = "/src/" + rootImportPath
	}

	GOPATH := gopath
	if customGOPATH := detectCustomGOPATH(ctx, fs); len(customGOPATH) > 0 {
		// Convert list of relative GOPATHs into absolute. We can have
		// more than one so we root ourselves at /workspace. We still
		// append the default GOPATH of `/` at the end. Fetched
		// dependencies will be mounted at that location.
		rootPath = "/workspace"
		rootImportPath = ""
		for i := range customGOPATH {
			customGOPATH[i] = rootPath + customGOPATH[i]
		}
		customGOPATH = append(customGOPATH, gopath)
		GOPATH = strings.Join(customGOPATH, ":")
	}

	// Send "initialize" to the wrapped lang server.
	langInitParams := &langserver.InitializeParams{
		InitializeParams:     params.InitializeParams,
		NoOSFileSystemAccess: true,
		BuildContext: &langserver.InitializeBuildContextParams{
			GOOS:       goos,
			GOARCH:     goarch,
			GOPATH:     GOPATH,
			GOROOT:     goroot,
			CgoEnabled: false,
			Compiler:   gocompiler,

			// TODO(sqs): We'd like to set this to true only for
			// the package we're analyzing (or for the whole
			// repo), but go/loader is insufficiently
			// configurable, so it applies it to the entire
			// program, which takes a lot longer and causes weird
			// error messages in the runtime package, etc. Disable
			// it for now.
			UseAllFiles: false,
		},
	}

	langInitParams.RootURI = lsp.DocumentURI("file://" + rootPath)
	langInitParams.RootImportPath = rootImportPath

	return langInitParams, nil
}

// detectCustomGOPATH tries to detect monorepos which require their own custom
// GOPATH.
//
// This is best-effort. If any errors occur or we do not detect a custom
// gopath, an empty result is returned.
func detectCustomGOPATH(ctx context.Context, fs ctxvfs.FileSystem) (gopaths []string) {
	// If we detect any .sorucegraph/config.json GOPATHs then they take
	// absolute precedence and override all others.
	if paths := detectSourcegraphGOPATH(ctx, fs); len(paths) > 0 {
		return paths
	}

	// Check .vscode/config.json and .envrc files, giving them equal precedence.
	if paths := detectVSCodeGOPATH(ctx, fs); len(paths) > 0 {
		gopaths = append(gopaths, paths...)
	}
	if paths := detectEnvRCGOPATH(ctx, fs); len(paths) > 0 {
		gopaths = append(gopaths, paths...)
	}
	return
}

// detectVSCodeGOPATH tries to detect monorepos which require their own custom
// GOPATH. We want to support monorepos as described in
// https://blog.gopheracademy.com/advent-2015/go-in-a-monorepo/ We use
// .vscode/settings.json to be informed of the custom GOPATH.
//
// This is best-effort. If any errors occur or we do not detect a custom
// gopath, an empty result is returned.
func detectVSCodeGOPATH(ctx context.Context, fs ctxvfs.FileSystem) []string {
	const settingsPath = ".vscode/settings.json"
	b, err := ctxvfs.ReadFile(ctx, fs, "/"+settingsPath)
	if err != nil {
		return nil
	}
	settings := struct {
		GOPATH string `json:"go.gopath"`
	}{}
	if err := unmarshalJSONC(string(b), &settings); err != nil {
		log15.Warn("Failed to parse JSON in "+settingsPath+" file. Treating as empty.", "err", err)
	}

	var paths []string
	for _, p := range filepath.SplitList(settings.GOPATH) {
		// We only care about relative gopaths
		if !strings.HasPrefix(p, "${workspaceRoot}") {
			continue
		}
		paths = append(paths, p[len("${workspaceRoot}"):])
	}
	return paths
}

// unmarshalJSONC unmarshals the JSON using a fault-tolerant parser that allows comments
// and trailing commas. If any unrecoverable faults are found, an error is returned.
func unmarshalJSONC(text string, v interface{}) error {
	data, errs := jsonx.Parse(text, jsonx.ParseOptions{Comments: true, TrailingCommas: true})
	if len(errs) > 0 {
		return fmt.Errorf("failed to parse JSON: %v", errs)
	}
	return json.Unmarshal(data, v)
}

// detectEnvRCGOPATH tries to detect monorepos which require their own custom
// GOPATH. We want to support monorepos such as the ones described in
// http://tammersaleh.com/posts/manage-your-gopath-with-direnv/ We use
// $REPO_ROOT/.envrc to be informed of the custom GOPATH. We support any line
// matching one of two formats below (because we do not want to actually
// execute .envrc):
//
// 	export GOPATH=VALUE
// 	GOPATH_add VALUE
//
// Where "VALUE" may be any of:
//
// 	some/relative/path
// 	one/:two:three/
// 	${PWD}/path
// 	$(PWD)/path
// 	`pwd`/path
//
// Or any of the above with double or single quotes wrapped around them. We
// will ignore any absolute path values.
func detectEnvRCGOPATH(ctx context.Context, fs ctxvfs.FileSystem) (gopaths []string) {
	b, err := ctxvfs.ReadFile(ctx, fs, "/.envrc")
	if err != nil {
		return nil
	}
	scanner := bufio.NewScanner(bytes.NewReader(b))
	for scanner.Scan() {
		value := ""
		line := scanner.Text()
		if prefixStr := "export GOPATH="; strings.HasPrefix(line, prefixStr) {
			value = strings.TrimSpace(strings.TrimPrefix(line, prefixStr))
		} else if prefixStr := "GOPATH_add "; strings.HasPrefix(line, prefixStr) {
			value = strings.TrimSpace(strings.TrimPrefix(line, prefixStr))
		} else {
			continue // no value
		}
		value = unquote(value, `"`) // remove double quotes
		value = unquote(value, `'`) // remove single quotes
		for _, value := range strings.Split(value, ":") {
			if strings.HasPrefix(value, "/") {
				// Not interested in absolute paths.
				continue
			}

			// Replace any form of PWD with an empty string (so we get a path
			// relative to repo root).
			value = strings.Replace(value, "${PWD}", "", -1)
			value = strings.Replace(value, "$(PWD)", "", -1)
			value = strings.Replace(value, "`pwd`", "", -1)
			if !strings.HasPrefix(value, "/") {
				value = "/" + value
			}
			gopaths = append(gopaths, value)
		}
	}
	_ = scanner.Err() // discarded intentionally
	return
}

// unquote removes the given quote string (either `'` or `"`) from the given
// string if it is wrapped in them.
func unquote(s, quote string) string {
	if !strings.HasPrefix(s, quote) && !strings.HasSuffix(s, quote) {
		return s
	}
	s = strings.TrimPrefix(s, quote)
	s = strings.TrimSuffix(s, quote)
	return s
}

// detectSourcegraphGOPATH tries to detect monorepos which require their own custom
// GOPATH. We detect a .sourcegraph/config.json file with the following
// contents:
//
// 	{
// 	  "go": {
// 	    "GOPATH": ["gopathdir", "gopathdir2"]
// 	  }
// 	}
//
// See the sourcegraphConfig struct documentation for more info.
//
// This is best-effort. If any errors occur or we do not detect a custom
// gopath, an empty result is returned.
func detectSourcegraphGOPATH(ctx context.Context, fs ctxvfs.FileSystem) (gopaths []string) {
	cfg := readSourcegraphConfig(ctx, fs)
	for _, p := range cfg.Go.GOPATH {
		if !strings.HasPrefix(p, "/") {
			// Assume all paths are relative to repo root.
			p = "/" + p
		}
		gopaths = append(gopaths, p)
	}
	return
}

// sourcegraphConfig is a struct representing the Go portion of the
// .sourcegraph/config.json file that may be placed in the repo root.
type sourcegraphConfig struct {
	// Go is the go portion of the configuration. Any language server can
	// define their own schema for their language, hence the namespacing here.
	Go struct {
		// GOPATH is a list of GOPATHs to use for the repository. It is assumed
		// each GOPATH string value is a path relative to the repository root.
		//
		// This is to support monorepos such as the ones described in
		// https://blog.gopheracademy.com/advent-2015/go-in-a-monorepo/
		//
		// See https://docs.sourcegraph.com/extensions/language_servers/go#custom-gopaths-go-monorepos.
		GOPATH []string

		// RootImportPath defines what Go import path corresponds to the
		// repository root. Effectively, the Go language server will clone the
		// repository into $GOPATH/src/$ROOT_IMPORT_PATH when set.
		//
		// This overrides the heuristic-based approach to locate this
		// information (from glide.yml or canonical import path comments) and
		// gives you an opportunity to specify it directly.
		//
		// See https://docs.sourcegraph.com/extensions/language_servers/go#vanity-import-paths.
		RootImportPath string
	} `json:"go"`
}

// readSourcegraphConfig reads the .sourcegraph/config.json file from the
// repository root if it exists, otherwise an empty struct value (not nil).
func readSourcegraphConfig(ctx context.Context, fs ctxvfs.FileSystem) *sourcegraphConfig {
	config := sourcegraphConfig{}
	b, err := ctxvfs.ReadFile(ctx, fs, "/.sourcegraph/config.json")
	if err != nil {
		return &config
	}
	_ = json.Unmarshal(b, &config)
	return &config
}

// determineRootImportPath determines the root import path for the Go
// workspace. It looks at canonical import path comments and the
// repo's original clone URL to infer it.
//
// It's intended to handle cases like
// github.com/kubernetes/kubernetes, which has doc.go files that
// indicate its root import path is k8s.io/kubernetes.
func determineRootImportPath(ctx context.Context, originalRootURI lsp.DocumentURI, fs ctxvfs.FileSystem) (rootImportPath string, err error) {
	if originalRootURI == "" {
		return "", errors.New("unable to determine Go workspace root import path without due to empty root path")
	}
	u, err := gituri.Parse(string(originalRootURI))
	if err != nil {
		return "", err
	}
	switch u.Scheme {
	case "git":
		rootImportPath = path.Join(u.Host, strings.TrimSuffix(u.Path, ".git"), u.FilePath())
	default:
		return "", fmt.Errorf("unrecognized originalRootPath: %q", u)
	}

	// If .sourcegraph/config.json specifies a root import path to use, then
	// use that one above all else.
	cfg := readSourcegraphConfig(ctx, fs)
	if v := cfg.Go.RootImportPath; v != "" {
		return v, nil
	}

	// Glide provides a canonical import path for us, try that first if it
	// exists.
	yml, err := ctxvfs.ReadFile(ctx, fs, "/glide.yaml")
	if err == nil && len(yml) > 0 {
		glide := struct {
			Package string `yaml:"package"`
			// There are other fields, but we don't use them
		}{}
		// best effort, so ignore error if we have a badly formatted
		// yml file
		_ = yaml.Unmarshal(yml, &glide)
		if glide.Package != "" {
			return glide.Package, nil
		}
	}

	// Now scan for canonical import path comments. This is a
	// heuristic; it is not guaranteed to produce the right result
	// (e.g., you could have multiple files with different canonical
	// import path comments that don't share a prefix, which is weird
	// and would break this).
	//
	// Since we have not yet set h.FS, we need to use the passed in fs.
	w := ctxvfs.Walk(ctx, "/", fs)
	const maxSlashes = 4 // heuristic, shouldn't need to traverse too deep to find this out
	const maxFiles = 25  // heuristic, shouldn't need to read too many files to find this out
	numFiles := 0
	for w.Step() {
		if err := w.Err(); err != nil {
			return "", err
		}
		fi := w.Stat()
		if fi.Mode().IsDir() && ((fi.Name() != "." && strings.HasPrefix(fi.Name(), ".")) || fi.Name() == "examples" || fi.Name() == "Godeps" || fi.Name() == "vendor" || fi.Name() == "third_party" || strings.HasPrefix(fi.Name(), "_") || strings.Count(w.Path(), "/") >= maxSlashes) {
			w.SkipDir()
			continue
		}
		if strings.HasSuffix(fi.Name(), ".go") {
			if numFiles >= maxFiles {
				// Instead of breaking, we SkipDir here so that we
				// ensure we always read all files in the root dir (to
				// improve the heuristic hit rate). We will not read
				// any more subdir files after calling SkipDir, which
				// is what we want.
				w.SkipDir()
			}
			numFiles++

			// For perf, read for the canonical import path 1 file at
			// a time instead of using build.Import, which always
			// reads all the files.
			contents, err := ctxvfs.ReadFile(ctx, fs, w.Path())
			if err != nil {
				return "", err
			}
			canonImportPath, err := readCanonicalImportPath(contents)
			if err == nil && canonImportPath != "" {
				// Chop off the subpackage path.
				parts := strings.Split(canonImportPath, "/")
				popComponents := strings.Count(w.Path(), "/") - 1
				if len(parts) <= popComponents {
					return "", fmt.Errorf("invalid canonical import path %q in file at path %q", canonImportPath, w.Path())
				}
				return strings.Join(parts[:len(parts)-popComponents], "/"), nil
			}
		}
	}

	// No canonical import path found, using our heuristics. Use the
	// root import path derived from the repo's clone URL.
	return rootImportPath, nil
}

func readCanonicalImportPath(contents []byte) (string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", contents, parser.PackageClauseOnly|parser.ParseComments)
	if err != nil {
		return "", err
	}
	if len(f.Comments) > 0 {
		// Take the last comment (`package foo // import "xyz"`), so
		// that we avoid copyright notices and package comments.
		c := f.Comments[len(f.Comments)-1]
		txt := strings.TrimSpace(c.Text())
		if strings.HasPrefix(txt, `import "`) && strings.HasSuffix(txt, `"`) {
			return strings.TrimSuffix(strings.TrimPrefix(txt, `import "`), `"`), nil
		}
	}
	return "", nil
}
