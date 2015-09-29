// Package tmpl defines, loads, and renders the app's templates.
package tmpl

import (
	"fmt"
	htmpl "html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"sync"

	"golang.org/x/net/context"

	"github.com/justinas/nosurf"
	"github.com/sourcegraph/mux"
	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	appauth "src.sourcegraph.com/sourcegraph/app/auth"
	"src.sourcegraph.com/sourcegraph/app/internal/appconf"
	"src.sourcegraph.com/sourcegraph/app/internal/canonicalurl"
	"src.sourcegraph.com/sourcegraph/app/internal/godocsupport"
	"src.sourcegraph.com/sourcegraph/app/internal/returnto"
	tmpldata "src.sourcegraph.com/sourcegraph/app/templates"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/conf/feature"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
	"src.sourcegraph.com/sourcegraph/util/traceutil"
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

// repoTemplates returns all repository template pages if successful.
func repoTemplates() error {
	return parseHTMLTemplates([][]string{
		{"repo/main.html", "repo/readme.inc.html", "repo/tree.inc.html", "repo/tree/dir.inc.html", "repo/commit.inc.html"},
		{"repo/not_enabled.html"},
		{"repo/badges.html", "repo/badges_and_counters.html"},
		{"repo/counters.html", "repo/badges_and_counters.html"},
		{"repo/builds.html", "builds/build.inc.html"},
		{"repo/build.html", "builds/build.inc.html", "repo/commit.inc.html"},
		{"repo/tree/file.html"},
		{"repo/tree/doc.html", "repo/commit.inc.html"},
		{"repo/tree/share.html"},
		{"repo/tree/dir.html", "repo/tree/dir.inc.html", "repo/commit.inc.html"},
		{"repo/search_results.html", "search/results.inc.html"},
		{"repo/godoc.html", "error/common.html"},
		{"repo/frame.html", "error/common.html"},
		{"repo/commit.html", "repo/commit.inc.html"},
		{"repo/commits.html", "repo/commit.inc.html"},
		{"repo/branches.html"},
		{"repo/tags.html"},
		{"repo/compare.html", "repo/commit.inc.html"},
		{"repo/discussion.html", "repo/discussions.inc.html"},
		{"repo/discussions.html", "repo/discussions.inc.html"},
		{"repo/changeset.html"},
		{"repo/changeset.list.html"},
		{"repo/changeset.notfound.html"},
		{"repo/no_vcs_data.html"},
		{"repo/no_build.html"},

		{"def/share.html", "def/def.html"},
		{"def/examples.html", "def/examples.inc.html", "def/snippet.inc.html", "def/def.html"},
	}, []string{
		"repo/repo.html",
		"repo/subnav.html",

		"common.html",
		"layout.html",
		"nav.html",
		"footer.html",
	})
}

// commonTemplates returns all common templates such as user pages, blog, search,
// etc. if successful.
func commonTemplates() error {
	return parseHTMLTemplates([][]string{
		{"user/login.html"},
		{"user/signup.html"},
		{"user/logged_out.html"},
		{"user/forgot_password.html"},
		{"user/password_reset.html"},
		{"user/new_password.html"},
		{"user/owned_repos.html", "user/owned_repos.inc.html", "user/person.html", "user/profile.inc.html"},
		{"user/orgs.html", "user/orgs.inc.html", "user/person.html", "user/profile.inc.html"},
		{"user/settings/profile.html", "user/settings/common.inc.html", "user/profile.inc.html"},
		{"user/settings/notifications.html", "user/settings/common.inc.html", "user/profile.inc.html"},
		{"user/settings/emails.html", "user/settings/common.inc.html", "user/profile.inc.html"},
		{"user/settings/integrations.html", "user/settings/common.inc.html", "user/profile.inc.html"},
		{"user/settings/auth.html", "user/settings/common.inc.html", "user/profile.inc.html"},
		{"blog/index.html", "blog/blog.html", "blog/common.inc.html"},
		{"blog/post.html", "blog/blog.html", "blog/common.inc.html"},

		{"liveblog/index.html", "liveblog/layout.html", "liveblog/common.inc.html"},
		{"liveblog/post.html", "liveblog/layout.html", "liveblog/common.inc.html"},

		{"home/new.html"},
		{"home/dashboard.html"},
		{"home/welcome.html"},

		{"search/form.html"},
		{"search/results.html", "search/results.inc.html"},

		{"builds/builds.html", "builds/build.inc.html"},
		{"error/error.html", "error/common.html"},
		{"oauth-provider/authorize.html"},
		{"client-registration/register-client.html"},
		{"org/members.html", "org/members.inc.html", "user/person.html", "user/profile.inc.html"},
		{"godoc/home.html"},
	}, []string{
		"common.html",
		"layout.html",
		"nav.html",
		"footer.html",
	})
}

// standaloneTemplates returns a set of standalone templates (sourcebox, codebox,
// etc.) if successful.
func standaloneTemplates() error {
	return parseHTMLTemplates([][]string{
		{"sourcebox/sourcebox.js"},
		{"sourcebox/sourcebox.html"},
		{"def/popover.html"},
	}, []string{"common.html"})
}

// Load loads (or re-loads) all template files from disk.
func Load() {
	if err := repoTemplates(); err != nil {
		log.Fatal(err)
	}
	if err := commonTemplates(); err != nil {
		log.Fatal(err)
	}
	if err := standaloneTemplates(); err != nil {
		log.Fatal(err)
	}
}

// Common holds fields that are available at the top level in every
// template executed by Exec.
type Common struct {
	RequestHost string // the request's Host header

	Session   *appauth.Session // the session cookie
	CSRFToken string

	CurrentUser   *sourcegraph.User
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

	// FullWidth sets the main body and navigation to fluid.
	FullWidth bool

	// Configuration for connecting to Traceguide servers.
	TraceguideAccessToken string
	TraceguideServiceHost string
}

func executeTemplateBase(w http.ResponseWriter, templateName, templateSubName string, data interface{}) error {
	t := Get(templateName)
	if t == nil {
		return fmt.Errorf("Template %s not found", templateName)
	}
	if templateSubName != "" {
		t = t.Lookup(templateSubName)
		if t == nil {
			return fmt.Errorf("Template %s %s not found", templateName, templateSubName)
		}
	}
	return t.Execute(w, data)
}

// Exec executes the template (named by `name`) using the template data.
func Exec(req *http.Request, resp http.ResponseWriter, name string, status int, header http.Header, data interface{}) error {
	if data != nil {
		ctx := httpctx.FromRequest(req)

		sess, err := appauth.ReadSessionCookie(req)
		if err != nil && err != appauth.ErrNoSession {
			return err
		}

		field := reflect.ValueOf(data).Elem().FieldByName("Common")
		existingCommon := field.Interface().(Common)

		currentURL := conf.AppURL(ctx).ResolveReference(req.URL)
		canonicalURL := existingCommon.CanonicalURL
		if canonicalURL == nil {
			canonicalURL = canonicalurl.FromURL(currentURL)
		}

		returnTo, _ := returnto.BestGuess(req)

		field.Set(reflect.ValueOf(Common{
			CurrentUser: handlerutil.UserFromRequest(req),

			RequestHost: req.Host,

			Session:   sess,
			CSRFToken: nosurf.Token(req),

			TemplateName: name,

			CurrentRoute: httpctx.RouteName(req),
			CurrentURI:   req.URL,
			CurrentURL:   currentURL,
			CurrentQuery: req.URL.Query(),

			AppURL:       conf.AppURL(ctx),
			CanonicalURL: canonicalURL,

			Ctx: ctx,

			CurrentSpanID:    traceutil.SpanID(req),
			CurrentRouteVars: mux.Vars(req),
			Debug:            handlerutil.DebugMode(req),
			ReturnTo:         returnTo,

			DisableExternalLinks: appconf.Current.DisableExternalLinks,
			Features:             feature.Features,
			FullWidth:            existingCommon.FullWidth,

			TraceguideAccessToken: os.Getenv("SG_TRACEGUIDE_ACCESS_TOKEN"),
			TraceguideServiceHost: os.Getenv("SG_TRACEGUIDE_SERVICE_HOST"),
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

	// Cache pjax and normal responses separately.
	bw.Header().Add("vary", "X-Pjax")

	bw.WriteHeader(status)
	if status == http.StatusNotModified {
		return nil
	}

	templateSubName := ""
	if usePJAX, tmplName := checkPJAX(req); usePJAX {
		templateSubName = tmplName
	}

	if err := executeTemplateBase(&bw, name, templateSubName, data); err != nil {
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
			t.Funcs(godocsupport.TemplateFuncMap)

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
var FuncMap htmpl.FuncMap

// checkPJAX returns whether the request is for a partial PJAX page,
// and which PJAX html/template to use (default is "PJAX").
func checkPJAX(r *http.Request) (usePJAX bool, tmplName string) {
	v := r.Header.Get("x-pjax-container")
	if v == "" {
		return false, ""
	}
	if tmplName, ok := pjaxDOMIDToTemplate[v]; ok {
		return true, tmplName
	}
	return true, "PJAX"
}

// pjaxDOMIDToTemplate maps PJAX container names in the DOM (e.g.,
// #repo-pjax-container) to the html/template name that should replace
// that DOM element (e.g., RepoPJAX). If none is specified for a DOM
// ID, the default template name is "PJAX".
var pjaxDOMIDToTemplate = map[string]string{
	"#repo-pjax-container": "RepoPJAX",
}
