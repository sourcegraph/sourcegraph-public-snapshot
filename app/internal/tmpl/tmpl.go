// Package tmpl defines, loads, and renders the app's templates.
package tmpl

import (
	"fmt"
	htmpl "html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"

	"golang.org/x/net/context"

	"github.com/gorilla/mux"
	"github.com/justinas/nosurf"
	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/sourcegraph/app/appconf"
	appauth "sourcegraph.com/sourcegraph/sourcegraph/app/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/app/jscontext"
	tmpldata "sourcegraph.com/sourcegraph/sourcegraph/app/templates"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
	"sourcegraph.com/sourcegraph/sourcegraph/util/traceutil"
)

var (
	templates   = map[string]*htmpl.Template{}
	templatesMu sync.Mutex
)

// Get gets a template by name, if it exists (and has previously been
// parsed, either by Load or by Add).
// Templates generally bare the name of the first file in their set.
func Get(name string) *htmpl.Template {
	templatesMu.Lock()
	t := templates[name]
	templatesMu.Unlock()
	return t
}

// Add adds a parsed template. It will be available to callers of Exec
// and Get.
//
// TODO(sqs): is this necessary?
func Add(name string, tmpl *htmpl.Template) {
	templatesMu.Lock()
	templates[name] = tmpl
	templatesMu.Unlock()
}

// Delete removes the named template.
func Delete(name string) {
	templatesMu.Lock()
	delete(templates, name)
	templatesMu.Unlock()
}

// commonTemplates returns all common templates such as user pages,
// etc., if successful.
func commonTemplates() error {
	return parseHTMLTemplates([][]string{
		{"error/error.html"},
	}, []string{
		"layout.html",
		"nav.html",
		"footer.html",
		"scripts.html",
	})
}

// Load loads (or re-loads) all template files from disk.
func Load() {
	if err := commonTemplates(); err != nil {
		log.Fatal(err)
	}
	if err := parseHTMLTemplates([][]string{{"ui.html", "layout.html", "scripts.html"}}, nil); err != nil {
		log.Fatal(err)
	}
}

// Common holds fields that are available at the top level in every
// template executed by Exec.
type Common struct {
	RequestHost string // the request's Host header

	Session   *appauth.Session // the session cookie
	CSRFToken string

	Actor auth.Actor

	CurrentRoute  string
	CurrentURI    *url.URL
	CurrentURL    *url.URL
	CurrentQuery  url.Values
	CurrentSpanID appdash.SpanID

	// TemplateName is the filename of the template being rendered
	// (e.g., "repo/main.html").
	TemplateName string

	// AppURL is the conf.AppURL(ctx) value for the current context.
	AppURL       *url.URL
	CanonicalURL *url.URL
	HostName     string

	Ctx context.Context

	CurrentRouteVars map[string]string

	// Debug is whether to show debugging info on the rendered page.
	Debug bool

	// ReturnTo is the URL to the page that the user should be returned to if
	// the user initiates a signup or login process from this page. Usually this
	// is the same as CurrentURI. The exceptions are when there are tracking
	// querystring parameters (we want to remove from the URL that the user
	// visits after signing up), and when the user is on the signup or login
	// pages themselves (otherwise we could get into a loop).
	//
	// The ReturnTo field is overridden by serveSignUp and other handlers that want
	// to set a ReturnTo different from CurrentURI.
	ReturnTo string

	// ExternalLinks decides if we should include links to things like
	// sourcegraph.com and the issue tracker on github.com
	DisableExternalLinks bool

	// Features is a struct containing feature toggles. See conf/feature
	Features interface{}

	// ErrorID is a randomly generated string used to identify a specific instance
	// of app error in the error logs.
	ErrorID string

	// CacheControl is the HTTP cache-control header value that should be set in all
	// AJAX requests originating from this page.
	CacheControl string

	// DeviceID is the correlation id given to user activity from a particular
	// device, pre- and post- authentication.
	DeviceID string

	JSCtx jscontext.JSContext
}

func executeTemplateBase(w http.ResponseWriter, templateName string, data interface{}) error {
	t := Get(templateName)
	if t == nil {
		return fmt.Errorf("Template %s not found", templateName)
	}
	return t.Execute(w, data)
}

// Exec executes the template (named by `name`) using the template data.
func Exec(req *http.Request, resp http.ResponseWriter, name string, status int, header http.Header, data interface{}) error {
	ctx := httpctx.FromRequest(req)

	if data != nil {
		sess, err := appauth.ReadSessionCookie(req)
		if err != nil && err != appauth.ErrNoSession {
			return err
		}

		field := reflect.ValueOf(data).Elem().FieldByName("Common")
		existingCommon := field.Interface().(Common)

		currentURL := conf.AppURL(ctx).ResolveReference(req.URL)

		// Propagate Cache-Control no-cache and max-age=0 directives
		// to the requests made by our client-side JavaScript. This is
		// not a perfect parser, but it catches the important cases.
		var cacheControl string
		if cc := req.Header.Get("cache-control"); strings.Contains(cc, "no-cache") || strings.Contains(cc, "max-age=0") {
			cacheControl = "no-cache"
		}

		jsctx, err := jscontext.NewJSContextFromRequest(ctx, req)
		if err != nil {
			return err
		}

		field.Set(reflect.ValueOf(Common{
			Actor: auth.ActorFromContext(ctx),

			RequestHost: req.Host,

			Session:   sess,
			CSRFToken: nosurf.Token(req),

			TemplateName: name,

			CurrentRoute: httpctx.RouteName(req),
			CurrentURI:   req.URL,
			CurrentURL:   currentURL,
			CurrentQuery: req.URL.Query(),

			AppURL: conf.AppURL(ctx),

			Ctx: ctx,

			CurrentSpanID:    traceutil.SpanID(req),
			CurrentRouteVars: mux.Vars(req),
			Debug:            handlerutil.DebugMode(req),

			DisableExternalLinks: appconf.Flags.DisableExternalLinks,
			Features:             feature.Features,

			ErrorID: existingCommon.ErrorID,

			CacheControl: cacheControl,

			JSCtx: jsctx,
		}))
	}

	// Buffer HTTP response so that if the template execution returns
	// an error (e.g., a template calls a template func that panics or
	// returns an error), we can return an HTTP error status code and
	// page to the browser. If we don't buffer it here, then the HTTP
	// response is already partially written to the client by the time
	// the error is detected, so the page rendering is aborted halfway
	// through with an error message, AND the HTTP status is 200
	// (which makes it hard to detect failures in tests).
	var bw httputil.ResponseBuffer

	for k, v := range header {
		bw.Header()[k] = v
	}
	if ct := bw.Header().Get("content-type"); ct == "" {
		bw.Header().Set("Content-Type", "text/html; charset=utf-8")
	}

	bw.WriteHeader(status)
	if status == http.StatusNotModified {
		return nil
	}

	if err := executeTemplateBase(&bw, name, data); err != nil {
		return err
	}

	return bw.WriteTo(resp)
}

// parseHTMLTemplates takes a list of template file sets. For each set in the
// list it creates a template containing all of the definitions found in that
// set. The name of each template will be the same as the first file in each
// set.
//
// A list of layout templates may also be provided. These will be shared
// amongst all templates.
func parseHTMLTemplates(sets [][]string, layout []string) error {
	var wg sync.WaitGroup
	for _, setv := range sets {
		set := setv
		if layout != nil {
			set = append(setv, layout...)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()

			t := htmpl.New("")
			t.Funcs(FuncMap)

			for _, tname := range set {
				f, err := tmpldata.Data.Open("/" + tname)
				if err != nil {
					log.Fatalf("read template asset %s: %s", tname, err)
				}
				tmpl, err := ioutil.ReadAll(f)
				f.Close()
				if err != nil {
					log.Fatalf("read template asset %s: %s", tname, err)
				}
				if _, err := t.Parse(string(tmpl)); err != nil {
					log.Fatalf("template %v: %s", set, err)
				}
			}

			t = t.Lookup("ROOT")
			if t == nil {
				log.Fatalf("ROOT template not found in %v", set)
			}
			Add(set[0], t)
		}()
	}
	wg.Wait()
	return nil
}

// FuncMap is the template func map passed to each template.
var FuncMap = htmpl.FuncMap{}
