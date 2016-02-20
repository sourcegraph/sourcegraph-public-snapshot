package apidocs

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"path"
	"reflect"
	"strings"

	"sourcegraph.com/sqs/pbtypes"

	"golang.org/x/net/context"

	"src.sourcegraph.com/apps/apidocs/assets"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/platform/pctx"
	"src.sourcegraph.com/sourcegraph/platform/putil"
)

// TmplCommon stores common information that will be exposed to all templates
// that serve data with this structure embedded (via executeTemplate).
type TmplCommon struct {
	// BaseURI holds the base URI provided by the platform for this app.
	BaseURI string

	// CSRFToken holds the CSRF token to validate the app with the platform.
	CSRFToken string

	// Ctx holds the context for this request.
	Ctx context.Context
}

var funcMap = template.FuncMap{
	"html":       html,
	"trimPrefix": strings.TrimPrefix,
	"kinds":      kinds,
}

// executeTemplate executes the template name with the given data and writes it
// to w.
func executeTemplate(w http.ResponseWriter, r *http.Request, name string, data interface{}) error {
	f, err := assets.Assets.Open("/" + name)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(f)
	f.Close()
	if err != nil {
		return err
	}
	t := template.New(name).Funcs(funcMap)
	t, err = t.Parse(string(b))
	if err != nil {
		return err
	}
	ctx := putil.Context(r)
	putCommon(ctx, data)
	return t.Execute(w, data)
}

// putCommon fills in the embedded TmplCommon structure inside data, if it exists.
func putCommon(ctx context.Context, data interface{}) {
	if data == nil {
		return
	}
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Ptr {
		panic("in executeTemplate: struct needs to be addressable (reflect.Ptr)")
	}
	f := v.Elem().FieldByName("TmplCommon")
	if !f.IsValid() {
		return
	}
	f.Set(reflect.ValueOf(TmplCommon{
		BaseURI:   pctx.BaseURI(ctx),
		CSRFToken: pctx.CSRFToken(ctx),
		Ctx:       ctx,
	}))
}

// html returns template.HTML for a safely pre-escaped pbtypes.HTML value.
func html(h *pbtypes.HTML) template.HTML {
	return template.HTML(h.HTML)
}

func kinds(kind string, defs []*sourcegraph.Def, requestDir string) (out []*sourcegraph.Def) {
	for _, def := range defs {
		// TODO(slimsag): srclib-go emits filepaths like "enum.go" so path.Dir("enum.go") == "."
		// but srclib-java emits full filepaths! ... investigate and confirm this is the real
		// issue / fix it.
		if def.Kind == kind && (path.Dir(def.File) == "." || path.Dir(def.File) == requestDir) {
			out = append(out, def)
		}
	}
	return
}
