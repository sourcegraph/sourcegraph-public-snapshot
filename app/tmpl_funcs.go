package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	htmpl "html/template"
	"net/url"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/appdash"

	"sourcegraph.com/sourcegraph/sourcegraph/app/appconf"
	"sourcegraph.com/sourcegraph/sourcegraph/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/buildvar"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/envutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil/appdashctx"
)

func init() {
	for name, fn := range tmplFuncs {
		if _, present := tmpl.FuncMap[name]; present {
			panic("template func already exists: " + name)
		}
		tmpl.FuncMap[name] = fn
	}
}

var tmplFuncs = htmpl.FuncMap{
	"appconf": func() interface{} { return &appconf.Flags },

	"json": func(v interface{}) (string, error) {
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(b), nil
	},
	"rawJSON": rawJSON,

	"customFeedbackForm": func() htmpl.HTML { return appconf.Flags.CustomFeedbackForm },

	"maxLen": func(maxLen int, s string) string {
		if len(s) <= maxLen {
			return s
		}
		return s[:maxLen]
	},

	"assetURL": assets.URL,

	"googleAnalyticsTrackingID": func() string { return appconf.Flags.GoogleAnalyticsTrackingID },

	"deployedGitCommitID": func() string { return envutil.GitCommitID },
	"fileSearchDisabled":  func() bool { return appconf.Flags.DisableSearch },

	"publicRavenDSN": func() string { return conf.PublicRavenDSN },

	"urlToAppdashTrace": func(ctx context.Context, trace appdash.ID) *url.URL {
		return appdashctx.AppdashURL(ctx).ResolveReference(&url.URL{
			Path: fmt.Sprintf("/traces/%v", trace),
		})
	},

	"buildvar": func() buildvar.Vars { return buildvar.All },
}

func rawJSON(v *json.RawMessage) htmpl.JS {
	if v == nil || len(*v) == 0 {
		return "null"
	}

	// SECURITY: Run through Go's JSON encoder to ensure this is
	// properly escaped JSON. Specifically, if it contains "<" or
	// ">" chars, we must escape those, or else they could be
	// interpreted as ending a <script> tag in an HTML page.
	var buf bytes.Buffer
	json.HTMLEscape(&buf, *v)

	return htmpl.JS(buf.String())
}
