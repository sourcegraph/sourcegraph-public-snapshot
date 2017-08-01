package gobuildserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/go-langserver/langserver"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspext"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"
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

// determineEnvironment will setup the LS InitializeParams based what it can
// detect from the filesystem and what it received from the client's
// InitializeParams.
//
// It is expected that fs will be mounted at the returns
// InitializeParams.RootPath.
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

	langInitParams.RootPath = rootPath
	langInitParams.RootURI = lsp.DocumentURI("file://" + rootPath)
	langInitParams.RootImportPath = rootImportPath

	return langInitParams, nil
}

// detectCustomGOPATH tries to detect monorepos which require their own custom
// GOPATH. We want to support monorepos as described in
// https://blog.gopheracademy.com/advent-2015/go-in-a-monorepo/ We use
// .vscode/settings.json to be informed of the custom GOPATH.
//
// This is best-effort. If any errors occur or we do not detect a custom
// gopath, an empty result is returned.
func detectCustomGOPATH(ctx context.Context, fs ctxvfs.FileSystem) []string {
	b, err := ctxvfs.ReadFile(ctx, fs, "/.vscode/settings.json")
	if err != nil {
		return nil
	}
	settings := struct {
		GOPATH string `json:"go.gopath"`
	}{}
	_ = json.Unmarshal(b, &settings)

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
	u, err := uri.Parse(string(originalRootURI))
	if err != nil {
		return "", err
	}
	switch u.Scheme {
	case "git":
		// TODO(keegancsmith) umami has .git paths in their import
		// paths. This normalization may in fact be incorrect. This
		// codeblock is to test if we really need this stripping in
		// production.
		if strings.HasSuffix(u.Path, ".git") {
			pathHasGitSuffix.Inc()
			defer func() {
				log.Printf("WARN: determineRootImportPath has .git suffix. before=%q after=%q %s", originalRootURI, rootImportPath, err)
			}()
		}
		rootImportPath = path.Join(u.Host, strings.TrimSuffix(u.Path, ".git"), u.FilePath())
	default:
		return "", fmt.Errorf("unrecognized originalRootPath: %q", u)
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
	const maxSlashes = 3 // heuristic, shouldn't need to traverse too deep to find this out
	const maxFiles = 25  // heuristic, shouldn't need to read too many files to find this out
	numFiles := 0
	for w.Step() {
		if err := w.Err(); err != nil {
			return "", err
		}
		fi := w.Stat()
		if fi.Mode().IsDir() && ((fi.Name() != "." && strings.HasPrefix(fi.Name(), ".")) || fi.Name() == "cmd" || fi.Name() == "examples" || fi.Name() == "Godeps" || fi.Name() == "vendor" || fi.Name() == "third_party" || strings.HasPrefix(fi.Name(), "_") || strings.Count(w.Path(), "/") >= maxSlashes) {
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

var pathHasGitSuffix = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "golangserver",
	Subsystem: "build",
	Name:      "path_has_git_suffix",
	Help:      "Temporary counter to determine if paths have a git suffix.",
})

func init() {
	prometheus.MustRegister(pathHasGitSuffix)
}
