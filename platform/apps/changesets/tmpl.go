package changesets

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"reflect"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/platform/apps/changesets/assets"
	"src.sourcegraph.com/sourcegraph/platform/pctx"
	"src.sourcegraph.com/sourcegraph/platform/putil"
	"src.sourcegraph.com/sourcegraph/util/timeutil"
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
	"timeAgo":        timeutil.TimeAgo,
	"urlToUser":      urlToUser,
	"urlToChangeset": urlToChangeset,
	"urlToCompare":   urlToCompare,
	"json":           toJSON,
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

// toJSON converts v into JSON and returns it as JS.
func toJSON(v interface{}) (template.JS, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return template.JS(b), nil
}

// urlToChangeset returns the URL to the changeset with the given id within the
// repository that is found in ctx.
func urlToChangeset(ctx context.Context, id int64) string {
	return fmt.Sprintf("%s/%d", pctx.BaseURI(ctx), id)
}

// urlToCompare returns the URL to compare view for the repository that is found
// in ctx.
func urlToCompare(ctx context.Context) string {
	// TODO(x): Relative path and compare view usage like this is bad.
	return fmt.Sprintf("%s/../.compare/master", pctx.BaseURI(ctx))
}

// urlToUser returns the route to a user's home page.
func urlToUser(user string) string {
	// TODO(x): Use a less hard-coded way.
	return "/~" + user
}
