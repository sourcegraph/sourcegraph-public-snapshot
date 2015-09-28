package devdoc

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	"reflect"

	"github.com/sourcegraph/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/devdoc/tmpl"
)

var (
	// ReloadTemplates is whether to reload html/template templates
	// before each request. It is useful during development.
	ReloadTemplates = true
)

var templates = [][]string{
	{"error.html", "layout.html"},
	{"root.html", "layout.html"},
	{"libraries.html", "layout.html"},
	{"api.html", "layout.html"},
	{"community.html", "layout.html"},
	{"enable.html", "layout.html"},
}

// TemplateCommon is data that is passed to (and available to) all templates.
type TemplateCommon struct {
	CurrentRoute string
	CurrentURI   *url.URL
	BaseURL      *url.URL
}

// newTemplateCommon returns a new TemplateCommon initialized for the given
// request.
func (a *App) newTemplateCommon(r *http.Request) (*TemplateCommon, error) {
	baseURL, err := a.URLTo(RootRoute)
	if err != nil {
		return nil, err
	}
	return &TemplateCommon{
		CurrentRoute: mux.CurrentRoute(r).GetName(),
		CurrentURI:   r.URL,
		BaseURL:      baseURL,
	}, nil
}

// renderTemplate renders the named template file out to w with the given data.
func (a *App) renderTemplate(w http.ResponseWriter, r *http.Request, name string, status int, data interface{}) error {
	// Prepare the template.
	t, err := a.prepTemplate(name)
	if err != nil {
		return err
	}

	if ct := w.Header().Get("content-type"); ct == "" {
		w.Header().Set("content-type", "text/html; charset=utf-8")
	}
	w.WriteHeader(status)

	if data != nil {
		// Set TemplateCommon values.
		tc, err := a.newTemplateCommon(r)
		if err != nil {
			return err
		}
		reflect.ValueOf(data).Elem().FieldByName("TemplateCommon").Set(reflect.ValueOf(*tc))
	}

	// Write to a buffer to properly catch errors and avoid partial output written to the http.ResponseWriter
	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	if err != nil {
		return err
	}
	_, err = buf.WriteTo(w)
	return err
}

// prepTemplate prepare the named template or returns an error. It is used such
// that a.tmplLock is not held during the entire length of the request.
func (a *App) prepTemplate(name string) (*template.Template, error) {
	a.tmplLock.Lock()
	defer a.tmplLock.Unlock()

	if a.tmpls == nil || ReloadTemplates {
		if err := a.parseHTMLTemplates(templates); err != nil {
			return nil, err
		}
	}

	t := a.tmpls[name]
	if t == nil {
		return nil, fmt.Errorf("Template %s not found", name)
	}
	return t, nil
}

// parseHTMLTemplates parses the HTML templates from their source.
func (a *App) parseHTMLTemplates(sets [][]string) error {
	a.tmpls = map[string]*template.Template{}
	for _, set := range sets {
		t := template.New("")
		t.Funcs(template.FuncMap{
			"urlTo": a.URLTo,
		})
		for _, tmp := range set {
			tmplBytes, err := tmpl.Asset(tmp)
			if err != nil {
				return fmt.Errorf("template %v: %s", set, err)
			}
			if _, err := t.Parse(string(tmplBytes)); err != nil {
				return fmt.Errorf("template %v: %s", set, err)
			}
		}
		t = t.Lookup("ROOT")
		if t == nil {
			return fmt.Errorf("ROOT template not found in %v", set)
		}
		a.tmpls[set[0]] = t
	}
	return nil
}
