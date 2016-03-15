package app

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	htmpl "html/template"
	"net/url"
	"os"
	"strings"
	"text/template"
	"time"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/sourcegraph/server/accesscontrol"

	"sourcegraph.com/sourcegraph/sourcegraph/app/appconf"
	"sourcegraph.com/sourcegraph/sourcegraph/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/schemautil"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/authutil"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/idkey"
	"sourcegraph.com/sourcegraph/sourcegraph/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/sgx/buildvar"
	"sourcegraph.com/sourcegraph/sourcegraph/sourcecode"
	"sourcegraph.com/sourcegraph/sourcegraph/util/envutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/metricutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/textutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/timeutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/traceutil/appdashctx"
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
	"personLabel": personLabel,

	"repoBasename": repoBasename,
	"repoLink":     repoLink,

	"repoMetaDescription": repoMetaDescription,

	"defQualifiedName":            sourcecode.DefQualifiedName,
	"defQualifiedNameAndType":     sourcecode.DefQualifiedNameAndType,
	"overrideStyleViaRegexpFlags": sourcecode.OverrideStyleViaRegexpFlags,

	"appconf":   func() interface{} { return &appconf.Flags },
	"authFlags": func() *authutil.Flags { return &authutil.ActiveFlags },

	"buildClass":  buildClass,
	"buildStatus": buildStatus,

	"add": func(a, b int) int { return a + b },
	"json": func(v interface{}) (string, error) {
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(b), nil
	},
	"rawJSON": func(v interface{}) (htmpl.JS, error) {
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return htmpl.JS(b), nil
	},

	"customFeedbackForm": func() htmpl.HTML { return appconf.Flags.CustomFeedbackForm },
	"uiBuild":            func() bool { return !appconf.Flags.NoUIBuild },

	"urlTo":              router.Rel.URLTo,
	"urlToUserSubroute":  router.Rel.URLToUserSubroute,
	"urlToRepo":          router.Rel.URLToRepo,
	"urlToRepoRev":       router.Rel.URLToRepoRev,
	"urlToRepoBuild":     router.Rel.URLToRepoBuild,
	"urlToRepoSubroute":  router.Rel.URLToRepoSubroute,
	"urlToRepoTreeEntry": router.Rel.URLToRepoTreeEntry,
	"urlToRepoCommit":    router.Rel.URLToRepoCommit,
	"urlWithSchema":      schemautil.URLWithSchema,
	"urlToDef":           router.Rel.URLToDef,
	"urlToDefAtRev":      router.Rel.URLToDefAtRev,
	"urlToDefSubroute":   router.Rel.URLToDefSubroute,
	"urlToWithReturnTo":  urlToWithReturnTo,

	"fileToBreadcrumb":    fileToBreadcrumb,
	"snippetToBreadcrumb": snippetToBreadcrumb,
	"router":              func() *router.Router { return router.Rel },

	"schemaMatchesExceptListAndSortOptions": schemautil.SchemaMatchesExceptListAndSortOptions,

	"classForRoute": func(route string) string {
		parts := strings.Split(route, ".")
		classes := make([]string, len(parts))
		for i := range parts {
			classes[i] = "route-" + strings.Join(parts[:i+1], "-")
		}
		return strings.Join(classes, " ")
	},

	// coalesce returns the first item in vals that is truthy (using
	// template.IsTrue), or nil if all items are nil.
	"coalesce": func(vals ...interface{}) interface{} {
		for _, v := range vals {
			truth, ok := template.IsTrue(v)
			if truth && ok {
				return v
			}
		}
		return nil
	},

	"commitSummary":       commitSummary,
	"commitRestOfMessage": commitRestOfMessage,

	"sanitizeHTML": sanitizeHTML,
	"textFromHTML": textutil.TextFromHTML,
	"timeOrNil":    timeutil.TimeOrNil,
	"timeAgo":      timeutil.TimeAgo,
	"now":          time.Now,
	"duration":     duration,
	"isNil":        isNil,
	"minTime":      minTime,
	"toInt": func(v interface{}) (int, error) {
		switch v := v.(type) {
		case int:
			return v, nil
		case uint32:
			return int(v), nil
		case int32:
			return int(v), nil
		case uint:
			return int(v), nil
		case uint64:
			return int(v), nil
		case int64:
			return int(v), nil
		}
		return 0, fmt.Errorf("toInt: unexpected type %T", v)
	},

	"truncate":         textutil.Truncate,
	"truncateCommitID": truncateCommitID,
	"maxLen":           maxLen,

	"assetURL": assets.URL,

	"getClientIDOrHostName": func(ctx context.Context, appURL *url.URL) string {
		clientID := idkey.FromContext(ctx).ID
		if clientID != "" {
			// return a shortened clientID, to match the clientID logged
			// in eventsutil/events.go:getShortClientID.
			if len(clientID) > 6 {
				return clientID[:6]
			}
			return clientID
		}
		if appURL == nil {
			return "unknown-host"
		}
		return appURL.Host
	},

	"hasField": hasStructField,

	"googleAnalyticsTrackingID": func() string { return appconf.Flags.GoogleAnalyticsTrackingID },

	"deployedGitCommitID": func() string { return envutil.GitCommitID },
	"hostname":            func() string { return hostname },

	"fileSearchDisabled": func() bool { return appconf.Flags.DisableSearch },

	"isAdmin": func(ctx context.Context, method string) bool {
		return accesscontrol.VerifyUserHasAdminAccess(ctx, method) == nil
	},

	"publicRavenDSN": func() string { return conf.PublicRavenDSN },

	"urlToAppdashTrace": func(ctx context.Context, trace appdash.ID) *url.URL {
		return appdashctx.AppdashURL(ctx).ResolveReference(&url.URL{
			Path: fmt.Sprintf("/traces/%v", trace),
		})
	},

	"buildvar": func() buildvar.Vars { return buildvar.All },

	"showDataCollectionMessage": func() bool { return !metricutil.DisableMetricsCollection() },

	"intercomHMAC": func(email string) string {
		mac := hmac.New(sha256.New, []byte(os.Getenv("SG_INTERCOM_SECRET_KEY")))
		mac.Write([]byte(email))
		return string(mac.Sum(nil))
	},
}
