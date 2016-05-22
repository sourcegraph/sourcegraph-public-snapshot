// Package static implements an HTTP middleware that satisfies
// requests using existing static files from a directory, if
// configured.
package static

import (
	"fmt"
	htmpl "html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"gopkg.in/inconshreveable/log15.v2"

	"sync"

	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/httpfs"
	"sourcegraph.com/sourcegraph/sourcegraph/app/appconf"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httputil/httpctx"
)

var Flags struct {
	Dir   string `long:"app.static-dir"   description:"path to a plain directory from which to serve static files" env:"SRC_APP_STATIC_DIR"`
	Dev   bool   `long:"app.static-dev"   description:"when present static template files are reloaded upon each request" env:"SRC_APP_STATIC_DEV"`
	Debug bool   `long:"app.static-debug" description:"debug serving of static files" env:"SRC_APP_STATIC_DEBUG"`
}

func init() {
	cli.PostInit = append(cli.PostInit, func() {
		cli.Serve.AddGroup("Static File Hosting", "Static File Hosting flags", &Flags)
	})
}

func init() {
	internal.Middleware = append(internal.Middleware, Middleware)
}

// Middleware satisfies requests using static content, if
// available. Otherwise it delegates to next.
//
// Implementation note: unlike staticMiddleware, it needs no
// instantiation.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if Flags.Dir == "" {
			next.ServeHTTP(w, r)
			return
		}

		reuseMu.Lock()
		if reuse == nil {
			var err error
			reuse, err = newMiddleware()
			if err != nil {
				log15.Crit("static middleware init failed", "err", err, "Flags", Flags)
				http.Error(w, "static middleware error", http.StatusInternalServerError)
				return
			}
		}
		reuseMu.Unlock()

		reuse.Middleware(w, r, next.ServeHTTP)
	})
}

var (
	reuseMu sync.Mutex
	reuse   *staticMiddleware
)

// staticMiddleware is a middleware that serves requests for static files and
// template files above all else. If a static file for the request is not
// present, the next handler in the chain is invoked to handle the request.
//
// Files with the ".tmpl or .html" extension are treated specially as Go html/template
// files, whereas files of any other extension are simply served statically.
//
// To provide nice URLs (e.g. "/foo" not "/foo.html" or "/foo.tmpl") serving
// occurs by first searching for a file in the static directory named "foo",
// then with a suffix of ".html" and then again with ".tmpl", serving whichever
// is first found.
//
// For better structure inside static file directories, requests to
// sub-directories (e.g. "/dir") are handled by their respective index file
// "/dir/index.html" or "/dir/index.tmpl").
//
// Files are served from the underlying vfs chosen at static.NewMiddleware time,
// and as such it may either be a plain OS directory or a VCS repository (which
// is useful for pushing changes etc in team-based environments).
type staticMiddleware struct {
	vfs        vfs.FileSystem
	httpfs     http.FileSystem
	fileServer http.Handler

	// serveTemplateHandler is literally just mw.serveTemplate except it is
	// wrapped in an error rendering handler.
	serveTemplateHandler http.Handler

	mu              sync.Mutex
	root            *htmpl.Template
	loadedTemplates []string
}

// Middleware is the actual middleware handler function.
func (mw *staticMiddleware) Middleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	mw.debugf("request for %q from dir %q\n", r.URL.Path, Flags.Dir)

	// If user is logged in and visits the home page, redirect them to dashboard.
	if r.URL.Path == "/" {
		mw.debugf("request for %q, redirecting user to app home page\n", r.URL.Path)
		next(w, r)
		return
	}

	// Choose the file appropriate.
	p, ok := mw.chooseFile(r.URL.Path, true)
	if !ok {
		next(w, r)
		return
	}

	httpctx.SetRouteName(r, "static")

	// Modify the URL that we're requesting to account for changes to the path
	// by chooseFile.
	r.URL.Path = p

	// If it's an html template file we serve it as a Go html template.
	if ext := path.Ext(p); ext == ".tmpl" || ext == ".html" {
		mw.serveTemplateHandler.ServeHTTP(w, r)
		return
	}

	// At this point it can only be a static file.

	mw.debugf("serving static file %q\n", r.URL.Path)
	mw.fileServer.ServeHTTP(w, r)
}

// serveTemplate serves the named static template file.
func (mw *staticMiddleware) serveTemplate(w http.ResponseWriter, r *http.Request) error {
	mw.debugf("serving template file %q\n", r.URL.Path)

	// Reload content if needed.
	if err := mw.reloadContent(false); err != nil {
		return err
	}

	// Execute the template.
	return tmpl.Exec(r, w, r.URL.Path, http.StatusOK, nil, &struct{ tmpl.Common }{})
}

// debugf is like log.Printf but it only prints its inputs if the
// --app.static-debug command-line flag is present, and it has a prefix string.
func (mw *staticMiddleware) debugf(f string, args ...interface{}) {
	if !Flags.Debug {
		return
	}
	f = fmt.Sprintf("app.static: %s", f)
	log.Printf(f, args...)
}

// chooseFile chooses the appropriate file for the input path string. It first
// tries to find literally the input path string, then with an ".html" suffix,
// and then again with a ".tmpl" suffix.
//
// Additionally if the path found is a directory an attempt will be made to find
// an index file (dir/index, dir/index.html or dir/index.tmpl) using this very
// function. When calling initially recurse should always be true.
func (mw *staticMiddleware) chooseFile(input string, recurse bool) (p string, ok bool) {
	// Find a file extension that the FS actually has.
	var (
		ext  string
		exts = []string{"", ".html", ".tmpl"}
	)
	for i, e := range exts {
		// First check if the VFS has the file or not. We have to do this because
		// FileServer doesn't expose any errors and would simply respond to the
		// request with a 404 if there is no such file.
		epath := input + e
		mw.debugf("trying stat %q\n", epath)
		fi, err := mw.vfs.Stat(epath)
		if err != nil {
			if i == len(exts)-1 {
				// Last extension, we don't have any more to try.
				mw.debugf("no such file with any extension, skipping\n")
				return "", false
			}
			continue // Try the next extension
		}

		// If the file is a directory, we're not interested.
		if fi.IsDir() {
			if !recurse {
				return "", false
			}
			mw.debugf("is directory not file, re-trying for index file\n")
			return mw.chooseFile(path.Join(input, "index"), false)
		}

		// Store the extension for later and quit the search.
		ext = e
		break
	}
	return input + ext, true
}

// loadTemplate loads the named template from the VFS.
func (mw *staticMiddleware) loadTemplate(name string) error {
	// Create the template.
	t := mw.root.New(name).Delims("[#[", "]#]")
	t.Funcs(tmpl.FuncMap)

	// Open the file.
	f, err := mw.vfs.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	// Read the data.
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	// Parse the template data.
	_, err = t.Parse(string(data))
	if err != nil {
		return err
	}

	// Place the loaded template into the global template execution map and the
	// local loaded template map for removal later.
	tmpl.Add(name, t)
	mw.mu.Lock()
	mw.loadedTemplates = append(mw.loadedTemplates, name)
	mw.mu.Unlock()
	return nil
}

// walkVFS walks the given directory in mw.vfs invoking the given function with
// the path and info of every file (and directory) encountered. The first error
// encountered is returned, if any.
func (mw *staticMiddleware) walkVFS(dir string, fn func(path string, fi os.FileInfo) error) error {
	// Read the directory listing.
	infos, err := mw.vfs.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, f := range infos {
		fp := path.Join(dir, f.Name())

		// Invoke the walk function with the filepath.
		if err := fn(fp, f); err != nil {
			return err
		}

		// Recursively walk the sub-directory.
		if f.IsDir() {
			if err := mw.walkVFS(fp, fn); err != nil {
				return err
			}
		}
	}
	return nil
}

// buildContentMap walks the entire static content directory building a map of
// the paths. This allows us to perform fast checks as to whether or not a
// request is for a static content file, or for something else not handled by
// this middleware.
//
// The force parameter specifies whether or not to force reloading e.g. if both
// ReloadAssets==false && Flags.Dev==false (used for first initialization)
func (mw *staticMiddleware) reloadContent(force bool) error {
	if !appconf.Flags.ReloadAssets && !Flags.Dev && !force {
		// Don't need to reload yet.
		return nil
	}
	mw.debugf("reloading templates\n")

	// Remove our loaded templates from the templates map now, as they may have
	// been removed from the directory.
	mw.mu.Lock()
	mw.root = htmpl.New("root")
	for _, loaded := range mw.loadedTemplates {
		tmpl.Delete(loaded)
	}
	mw.loadedTemplates = mw.loadedTemplates[:0]
	mw.mu.Unlock()

	// Walk the VFS loading just template files that we find.
	return mw.walkVFS("/", func(p string, fi os.FileInfo) error {
		if fi.IsDir() || (path.Ext(p) != ".tmpl" && path.Ext(p) != ".html") {
			return nil
		}
		return mw.loadTemplate(p)
	})
}

// newMiddleware returns a new initialized static-file-serving
// middleware. An error is returned only due to opening the VFS.
//
// If no app.static-dir CLI flag is provided, a
// panic will occur (the caller should check first).
func newMiddleware() (*staticMiddleware, error) {
	mw := &staticMiddleware{}

	if Flags.Dir != "" {
		mw.debugf("serving a normal directory\n")
		mw.vfs = vfs.OS(Flags.Dir)
	} else {
		panic("no dir or repo specified")
	}

	mw.serveTemplateHandler = internal.Handler(mw.serveTemplate)
	mw.httpfs = httpfs.New(mw.vfs)

	mw.fileServer = http.FileServer(mw.httpfs)
	return mw, mw.reloadContent(true)
}
