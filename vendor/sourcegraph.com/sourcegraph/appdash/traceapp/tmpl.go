package traceapp

import (
	"bytes"
	"errors"
	"fmt"
	htmpl "html/template"
	"io/ioutil"
	"net/http"
	"net/url"

	"reflect"
	"strconv"
	"strings"

	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/appdash/traceapp/tmpl"

	"github.com/gorilla/mux"
)

var (
	// ReloadTemplates is whether to reload html/template templates
	// before each request. It is useful during development.
	ReloadTemplates = true
)

var templates = [][]string{
	{"root.html", "layout.html"},
	{"trace.html", "layout.html"},
	{"traces.html", "layout.html"},
	{"dashboard.html", "layout.html"},
	{"aggregate.html", "layout.html"},
}

// TemplateCommon is data that is passed to (and available to) all templates.
type TemplateCommon struct {
	CurrentRoute  string
	CurrentURI    *url.URL
	BaseURL       *url.URL
	HaveDashboard bool
}

func (a *App) renderTemplate(w http.ResponseWriter, r *http.Request, name string, status int, data interface{}) error {
	a.tmplLock.Lock()
	defer a.tmplLock.Unlock()

	if a.tmpls == nil || ReloadTemplates {
		if err := a.parseHTMLTemplates(templates); err != nil {
			return err
		}
	}

	w.WriteHeader(status)
	if ct := w.Header().Get("content-type"); ct == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	t := a.tmpls[name]
	if t == nil {
		return fmt.Errorf("Template %s not found", name)
	}

	if data != nil {
		// Set TemplateCommon values.
		reflect.ValueOf(data).Elem().FieldByName("TemplateCommon").Set(reflect.ValueOf(TemplateCommon{
			CurrentRoute:  mux.CurrentRoute(r).GetName(),
			CurrentURI:    r.URL,
			BaseURL:       a.baseURL,
			HaveDashboard: a.Aggregator != nil,
		}))
	}

	// Write to a buffer to properly catch errors and avoid partial output written to the http.ResponseWriter
	var buf bytes.Buffer
	err := t.Execute(&buf, data)
	if err != nil {
		return err
	}
	_, err = buf.WriteTo(w)
	return err
}

// parseHTMLTemplates parses the HTML templates from their source.
func (a *App) parseHTMLTemplates(sets [][]string) error {
	a.tmpls = map[string]*htmpl.Template{}
	for _, set := range sets {
		t := htmpl.New("")
		t.Funcs(htmpl.FuncMap{
			"urlTo":             a.URLTo,
			"urlToTrace":        a.URLToTrace,
			"itoa":              strconv.Itoa,
			"str":               func(v interface{}) string { return fmt.Sprintf("%s", v) },
			"durationClass":     durationClass,
			"filterAnnotations": filterAnnotations,
			"descendTraces":     func() bool { return false },
			"dict":              dict,
		})
		for _, tmp := range set {
			tmplFile, err := tmpl.Data.Open("/" + tmp)
			if err != nil {
				return fmt.Errorf("template %v: %s", set, err)
			}
			tmplBytes, err := ioutil.ReadAll(tmplFile)
			tmplFile.Close()
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

func durationClass(usec int64) string {
	msec := usec / 1000
	if msec < 30 {
		return "d0"
	} else if msec < 60 {
		return "d1"
	} else if msec < 90 {
		return "d2"
	} else if msec < 150 {
		return "d3"
	} else if msec < 250 {
		return "d4"
	} else if msec < 400 {
		return "d5"
	} else if msec < 600 {
		return "d6"
	} else if msec < 900 {
		return "d7"
	} else if msec < 1300 {
		return "d8"
	} else if msec < 1900 {
		return "d9"
	}
	return "d10"
}

func filterAnnotations(anns appdash.Annotations) appdash.Annotations {
	var anns2 appdash.Annotations
	for _, ann := range anns {
		if ann.Key != "" && !strings.HasPrefix(ann.Key, "_") {
			anns2 = append(anns2, ann)
		}
	}
	return anns2

}

// dict builds a map of paired items, allowing you to invoke a template with
// multiple parameters.
func dict(pairs ...interface{}) (map[string]interface{}, error) {
	if len(pairs)%2 != 0 {
		return nil, errors.New("expected pairs")
	}
	m := make(map[string]interface{}, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		m[pairs[i].(string)] = pairs[i+1]
	}
	return m, nil
}
