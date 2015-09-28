package devdoc

import (
	"html/template"
	"net/http"
	"path"
	"strings"
	"sync"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"sourcegraph.com/sourcegraph/prototools/tmpl"
	"src.sourcegraph.com/sourcegraph/devdoc/assets"
	tmplassets "src.sourcegraph.com/sourcegraph/devdoc/tmpl"
)

// cacheController wraps the given HTTP handler and sets the Cache-Control
// header to cc (e.g. "max-age=300, public").
func cacheController(f http.Handler, cc string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", cc)
	})
}

// App represents the Sourcegraph developer application.
type App struct {
	*Router

	docs      *plugin.CodeGeneratorRequest
	generator *tmpl.Generator

	tmplLock sync.Mutex
	tmpls    map[string]*template.Template
}

// New returns a new app using the given base router, or creating a new one if
// it is nil.
func New(r *Router) *App {
	if r == nil {
		r = NewRouter(nil)
	}
	a := &App{
		Router: r,
	}
	a.r.Get(RootRoute).Handler(a.handler(a.serveBasic("root.html")))
	a.r.Get(LibrariesRoute).Handler(a.handler(a.serveBasic("libraries.html")))
	a.r.Get(CommunityRoute).Handler(a.handler(a.serveBasic("community.html")))
	a.r.Get(EnableRoute).Handler(a.handler(a.serveBasic("enable.html")))
	a.r.Get(APIRoute).Handler(a.handler(a.serveAPI))

	// Static file serving.
	u, err := a.Router.URLTo(RootRoute)
	if err != nil {
		panic(err)
	}
	staticPath := path.Join(u.Path, "static/")
	staticHandler := http.StripPrefix(staticPath, http.FileServer(&assetfs.AssetFS{
		Asset:    assets.Asset,
		AssetDir: assets.AssetDir,
		Prefix:   "",
	}))
	r.r.Get(StaticRoute).Handler(cacheController(staticHandler, "max-age=300, public"))

	// Try to initialize the doc generator, if we can't then we serve without
	// API docs (e.g. if it's not a release binary).
	if err := a.initGenerator(); err != nil {
		log15.Error("Serving without API documentation", "app", "devdoc", "error", err)
	}

	return a
}

// ServeHTTP implements http.Handler.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.r.ServeHTTP(w, r)
}

// initGenerator initializes the documentation generator. If any error is
// returned documentation cannot be generated (most likely due to a binary
// without sourcegraph.dump built in).
func (a *App) initGenerator() error {
	a.generator = tmpl.New()

	// Load all templates from the template assets.
	a.generator.ReadFile = tmplassets.Asset

	// Unmarshal the Protobuf-encoded request.
	a.docs = new(plugin.CodeGeneratorRequest)
	protoRequest, err := assets.Asset("sourcegraph.dump")
	if err != nil {
		return err
	}
	if err := proto.Unmarshal(protoRequest, a.docs); err != nil {
		return err
	}

	// Find the base URL path for generated docs.
	baseURL, err := a.URLTo(APIRoute)
	if err != nil {
		return err
	}
	a.generator.RootDir = baseURL.Path

	// APIHost is the prefix to append to gateway paths, which in our case we
	// just want to be root (we don't want to render the full host in gateway
	// paths).
	a.generator.APIHost = "/"

	// Set the request for the generator.
	if err := a.generator.SetRequest(a.docs); err != nil {
		return err
	}

	// Load the filemap from the template assets.
	fileMap, err := tmplassets.Asset("doc/filemap.xml")
	if err != nil {
		return err
	}

	return a.generator.ParseFileMap("doc/", string(fileMap))
}

// serveBasic performs non-specialized serving of a template.
func (a *App) serveBasic(tmplName string) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		return a.renderTemplate(w, r, tmplName, http.StatusOK, &struct {
			TemplateCommon
		}{})
	}
}

// serveAPI serves the API route.
func (a *App) serveAPI(w http.ResponseWriter, r *http.Request) error {
	if a.generator == nil {
		// Serving without docs, redirect to enable docs page.
		u, err := a.Router.URLTo(EnableRoute)
		if err != nil {
			return err
		}
		http.Redirect(w, r, u.String(), http.StatusSeeOther)
		return nil
	}

	// Redirect to the index file if it's the root route.
	u, err := a.Router.URLTo(APIRoute)
	if err != nil {
		return err
	}
	p := r.URL.Path
	if p == u.String() || p == u.String()+"/" {
		// Modify the path such that /api and /api/ is really a request for the
		// index page.
		p = path.Join(u.String(), "index.html")
	}

	// Modify the path to match the output file listed in the filemap generator.
	p = strings.TrimPrefix(p, u.Path)
	p = strings.TrimPrefix(p, "/")

	// Create a new TemplateCommon struct as our context for doc. generation
	// below.
	tc, err := a.newTemplateCommon(r)
	if err != nil {
		return err
	}

	// Generate an output HTML file for the named path (p) with the input $.Ctx
	// for the template invocation.
	//
	// TODO(slimsag): add rootDir fetching to prototools/tmpl/util.go instead of
	// using BaseURL inside our templates here. Static doc users will need it
	// probably, and it would make our templates more usable for generating
	// static docs.
	output, err := a.generator.GenerateOutput(p, tc)
	if err != nil {
		return err
	}

	// Serve the auto-generated documentation file through the API page-layout
	// template.
	return a.renderTemplate(w, r, "api.html", http.StatusOK, &struct {
		TemplateCommon
		HTML template.HTML
	}{
		HTML: template.HTML(output.GetContent()),
	})
}
