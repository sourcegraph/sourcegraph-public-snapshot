package tmpl

import (
	"context"
	"encoding/json"
	htmpl "html/template"
	"strconv"

	opentracing "github.com/opentracing/opentracing-go"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
)

var disableSupportServices, _ = strconv.ParseBool(env.Get("SRC_APP_DISABLE_SUPPORT_SERVICES", "false", "disable 3rd party support services, including Intercom, Google Analytics, etc"))
var googleAnalyticsTrackingID = env.Get("GOOGLE_ANALYTICS_TRACKING_ID", "", "Google Analytics tracking ID (UA-########-#)")

var FuncMap = htmpl.FuncMap{
	"disableSupportServices":    func() bool { return disableSupportServices },
	"googleAnalyticsTrackingID": func() string { return googleAnalyticsTrackingID },

	"json": func(v interface{}) (string, error) {
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(b), nil
	},

	"assetURL": assets.URL,

	"urlToTrace": func(ctx context.Context) string {
		if span := opentracing.SpanFromContext(ctx); span != nil {
			return traceutil.SpanURL(span)
		}
		return ""
	},

	"version": func() string { return env.Version },
}
