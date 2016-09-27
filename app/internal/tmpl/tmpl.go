// Package tmpl defines, loads, and renders the app's templates.
package tmpl

import (
	"fmt"
	htmpl "html/template"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"sync"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/appconf"
	"sourcegraph.com/sourcegraph/sourcegraph/app/jscontext"
	tmpldata "sourcegraph.com/sourcegraph/sourcegraph/app/templates"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httputil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httputil/httpctx"
)

var (
	templates   = map[string]*htmpl.Template{}
	templatesMu sync.Mutex
	loadOnce    sync.Once
)

// Get gets a template by name, if it exists (and has previously been
// parsed, either by Load or by add).
// Templates generally bare the name of the first file in their set.
func Get(name string) *htmpl.Template {
	templatesMu.Lock()
	t := templates[name]
	templatesMu.Unlock()
	return t
}

// add adds a parsed template. It will be available to callers of Exec
// and Get.
func add(name string, tmpl *htmpl.Template) {
	templatesMu.Lock()
	templates[name] = tmpl
	templatesMu.Unlock()
}

// commonTemplates returns all common templates such as user pages,
// etc., if successful.
func commonTemplates() error {
	return parseHTMLTemplates([][]string{{"error/error.html"}},
		[]string{"layout.html", "nav.html", "footer.html", "scripts.html", "styles.css"})
}

// Load loads (or re-loads) all template files from disk.
func Load() {
	if err := commonTemplates(); err != nil {
		log.Fatal(err)
	}
	if err := parseHTMLTemplates([][]string{
		{"ui.html", "head_from_meta.html", "layout.html", "scripts.html", "styles.css"},
		{"deflanding.html", "head_from_meta.html", "layout.html", "scripts.html", "styles.css"},
		{"repolanding.html", "head_from_meta.html", "layout.html", "scripts.html", "styles.css"},
		{"repoindex.html", "head_from_meta.html", "layout.html", "scripts.html", "styles.css"},
	}, nil); err != nil {
		log.Fatal(err)
	}
}

// LoadOnce loads all templates unless they have been already loaded.
func LoadOnce() {
	loadOnce.Do(func() {
		Load()
	})
}

// Common holds fields that are available at the top level in every
// template executed by Exec.
type Common struct {
	AuthInfo *sourcegraph.AuthInfo

	CurrentRoute string

	// TemplateName is the filename of the template being rendered
	// (e.g., "repo/main.html").
	TemplateName string

	Ctx context.Context

	// Debug is whether to show debugging info on the rendered page.
	Debug bool

	// ExternalLinks decides if we should include links to things like
	// sourcegraph.com and the issue tracker on github.com
	DisableExternalLinks bool

	// ErrorID is a randomly generated string used to identify a specific instance
	// of app error in the error logs.
	ErrorID string

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
	if data != nil {
		field := reflect.ValueOf(data).Elem().FieldByName("Common")
		existingCommon := field.Interface().(Common)

		jsctx, err := jscontext.NewJSContextFromRequest(req)
		if err != nil {
			return err
		}

		field.Set(reflect.ValueOf(Common{
			AuthInfo:             auth.ActorFromContext(req.Context()).AuthInfo(),
			TemplateName:         name,
			CurrentRoute:         httpctx.RouteName(req),
			Ctx:                  req.Context(),
			Debug:                handlerutil.DebugMode,
			DisableExternalLinks: appconf.Flags.DisableExternalLinks,
			ErrorID:              existingCommon.ErrorID,
			JSCtx:                jsctx,
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
			add(set[0], t)
		}()
	}
	wg.Wait()
	return nil
}

// FuncMap is the template func map passed to each template.
var FuncMap = htmpl.FuncMap{}
