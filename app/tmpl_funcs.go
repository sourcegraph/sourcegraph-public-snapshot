package app

import (
	"encoding/json"
	"fmt"
	htmpl "html/template"
	"net"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/appconf"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/schemautil"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/authutil"
	"sourcegraph.com/sourcegraph/sourcegraph/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/search"
	"sourcegraph.com/sourcegraph/sourcegraph/sgx/buildvar"
	"sourcegraph.com/sourcegraph/sourcegraph/sourcecode"
	"sourcegraph.com/sourcegraph/sourcegraph/util/envutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/textutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/timeutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/traceutil/appdashctx"
	"sourcegraph.com/sourcegraph/srclib/toolchain"
)

func init() {
	tmpl.FuncMap = TemplateFunctions
}

var TemplateFunctions = htmpl.FuncMap{
	"personLabel":         personLabel,
	"userMetaDescription": userMetaDescription,
	"userStat":            userStat,

	"repoBasename":      repoBasename,
	"repoLink":          repoLink,
	"absRepoLink":       absRepoLink,
	"repoLabelForOwner": repoLabelForOwner,

	"pathBase": path.Base,

	"repoMetaDescription": repoMetaDescription,
	"repoStat":            repoStat,

	"defQualifiedName":            sourcecode.DefQualifiedName,
	"defQualifiedNameAndType":     sourcecode.DefQualifiedNameAndType,
	"overrideStyleViaRegexpFlags": sourcecode.OverrideStyleViaRegexpFlags,

	"appconf":   func() *appconf.Flags { return &appconf.Current },
	"authFlags": func() *authutil.Flags { return &authutil.ActiveFlags },

	"buildClass":  buildClass,
	"buildStatus": buildStatus,

	"number":        number,
	"pluralizeWord": pluralizeWord,
	"pluralize":     pluralize,
	"firstSentence": textutil.FirstSentence,
	"firstChars":    textutil.FirstChars,
	"add":           func(a, b int) int { return a + b },
	"add32":         func(a, b int32) int32 { return a + b },
	"min": func(a, b int) int {
		if a < b {
			return a
		}
		return b
	},
	"json": func(v interface{}) string {
		b, _ := json.Marshal(v)
		return string(b)
	},

	// map creates a map of string keys and interface{} values given pairs. It can
	// be used to invoke templates with multiple parameters:
	//
	//  {{template "foo" (map "A" $a "B" $b)}}
	//
	// There must be an even number of values (i.e. pairs), with each first item
	// in the pair being a string, or else this function will panic.
	"map": func(values ...interface{}) map[string]interface{} {
		m := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			m[values[i].(string)] = values[i+1]
		}
		return m
	},

	"customLogo":         func() htmpl.HTML { return appconf.Current.CustomLogo },
	"motd":               func() htmpl.HTML { return appconf.Current.MOTD },
	"customFeedbackForm": func() htmpl.HTML { return appconf.Current.CustomFeedbackForm },
	"autoBuild":          func() bool { return !appconf.Current.NoAutoBuild },

	"trimPrefix": strings.TrimPrefix,

	"absURL": func(appURL, other *url.URL) *url.URL { return appURL.ResolveReference(other) },

	// isLocalhost is true if the server is running on localhost.
	"isLocalhost": func(host string) bool { return host == "localhost" || strings.HasPrefix(host, "localhost:") },

	"urlTo":                      router.Rel.URLTo,
	"urlToBlogPost":              router.Rel.URLToBlogPost,
	"urlToBlogAtomFeed":          router.Rel.URLToBlogAtomFeed,
	"urlToUser":                  router.Rel.URLToUser,
	"urlToUserSubroute":          router.Rel.URLToUserSubroute,
	"urlToRepo":                  router.Rel.URLToRepo,
	"urlToRepoRev":               router.Rel.URLToRepoRev,
	"urlToRepoChangesets":        router.Rel.URLToRepoChangesets,
	"urlToRepoChangeset":         router.Rel.URLToRepoChangeset,
	"urlToRepoBuild":             router.Rel.URLToRepoBuild,
	"urlToRepoSubroute":          router.Rel.URLToRepoSubroute,
	"urlToRepoSubrouteRev":       router.Rel.URLToRepoSubrouteRev,
	"urlToRepoTreeEntry":         router.Rel.URLToRepoTreeEntry,
	"urlToRepoTreeEntrySubroute": router.Rel.URLToRepoTreeEntrySubroute,
	"urlToRepoCommit":            router.Rel.URLToRepoCommit,
	"urlToRepoCompare":           router.Rel.URLToRepoCompare,
	"urlToRepoGoDoc":             router.Rel.URLToRepoGoDoc,
	"urlToRepoApp":               router.Rel.URLToRepoApp,
	"urlWithSchema":              schemautil.URLWithSchema,
	"urlToDef":                   router.Rel.URLToDef,
	"urlToDefAtRev":              router.Rel.URLToDefAtRev,
	"urlToDefSubroute":           router.Rel.URLToDefSubroute,
	"urlToWithReturnTo":          urlToWithReturnTo,
	"urlToRepoBuildSubroute":     router.Rel.URLToRepoBuildSubroute,
	"urlToRepoBuildTaskSubroute": router.Rel.URLToRepoBuildTaskSubroute,

	"fileToBreadcrumb":       FileToBreadcrumb,
	"fileLinesToBreadcrumb":  FileLinesToBreadcrumb,
	"snippetToBreadcrumb":    SnippetToBreadcrumb,
	"absSnippetToBreadcrumb": AbsSnippetToBreadcrumb,
	"router":                 func() *router.Router { return router.Rel },

	"searchFormInfo": searchFormInfo,

	"flattenName":     handlerutil.FlattenName,
	"flattenNameHTML": handlerutil.FlattenNameHTML,

	"repoMaybeUnsupported": repoMaybeUnsupported,

	"schemaMatchesExceptListAndSortOptions": schemautil.SchemaMatchesExceptListAndSortOptions,

	"classForRoute": func(route string) string {
		parts := strings.Split(route, ".")
		classes := make([]string, len(parts))
		for i := range parts {
			classes[i] = "route-" + strings.Join(parts[:i+1], "-")
		}
		return strings.Join(classes, " ")
	},
	"nextPageURL": func(currentURI *url.URL, inc int) string {
		values := currentURI.Query()

		pageField, exists := values["Page"]
		if !exists || len(pageField) != 1 {
			pageField = []string{"1"}
		}
		page, _ := strconv.Atoi(pageField[0])
		values["Page"] = []string{strconv.Itoa(page + inc)}
		delete(values, "_pjax")

		return "?" + values.Encode()
	},
	"effectivePage": func(p int) int {
		if p == 0 {
			return 1
		}
		return p
	},

	"commitSummary":       commitSummary,
	"commitRestOfMessage": commitRestOfMessage,

	"toString2":             func(v interface{}) string { return fmt.Sprintf("%s", v) },
	"bytesToString":         func(v []byte) string { return string(v) },
	"sanitizeHTML":          sanitizeHTML,
	"sanitizeFormattedCode": sanitizeFormattedCode,
	"textFromHTML":          textutil.TextFromHTML,
	"timeOrNil":             timeutil.TimeOrNil,
	"timeAgo":               timeutil.TimeAgo,
	"now":                   time.Now,
	"duration":              duration,
	"isNil":                 isNil,
	"minTime":               minTime,
	"pathJoin":              filepath.Join,
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
	"displayURL": func(urlStr string) string {
		return strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(urlStr, "https://"), "http://"), "/")
	},

	"queryDescribe": func(pbtoks []sourcegraph.PBToken) string {
		return search.Describe(sourcegraph.PBTokens(pbtoks))
	},
	"queryRevision": func(pbtoks []sourcegraph.PBToken) string {
		return search.Revision(sourcegraph.PBTokens(pbtoks))
	},

	"assetURL": assetURL,

	"hasField": hasStructField,

	"hasPrefix":                 strings.HasPrefix,
	"ifTemplate":                ifTemplate,
	"googleAnalyticsTrackingID": func() string { return appconf.Current.GoogleAnalyticsTrackingID },
	"heapAnalyticsID":           func() string { return appconf.Current.HeapAnalyticsID },

	"deployedGitCommitID": func() string { return envutil.GitCommitID },
	"hostname":            func() string { return hostname },

	"nl2br": func(s string) htmpl.HTML {
		return htmpl.HTML(strings.Replace(htmpl.HTMLEscapeString(s), "\n", "<br>", -1))
	},

	"gosrcBaseURL": func(appURL *url.URL) string {
		_, port, _ := net.SplitHostPort(appURL.Host)
		if port != "" {
			port = ":" + port
		}
		return (&url.URL{
			Scheme: "http", // TODO(sqs): get ssl cert for gosrc.org
			Host:   "gosrc.org" + port,
		}).String()
	},
	"gosrcBookmarklet": func() htmpl.URL { return htmpl.URL(gosrcBookmarklet) },

	"showRepoRevSwitcher": showRepoRevSwitcher,

	"repoEnabledFrames":          repoEnabledFrames,
	"repoEnabledFrameChangesets": repoEnabledFrameChangesets,
	"showSearchForm":             showSearchForm,
	"fileSearchDisabled":         func() bool { return appconf.Current.DisableSearch },
	"disableCloneURL":            func() bool { return appconf.Current.DisableCloneURL },

	// Returns whether or not any srclib toolchains are installed.
	"haveToolchain": func() bool {
		infos, err := toolchain.List()
		// note: err != nil if e.g. $SRCLIBPATH is not a directory.
		return err == nil && len(infos) > 0
	},

	"repoEnabled": func(c *sourcegraph.RepoConfig) bool {
		return c != nil && c.Enabled
	},

	"repoPerms": func(p *sourcegraph.RepoPermissions) sourcegraph.RepoPermissions {
		if p == nil {
			return sourcegraph.RepoPermissions{}
		}
		return *p
	},

	"publicRavenDSN": func() string { return conf.PublicRavenDSN },

	"urlToAppdashTrace": func(ctx context.Context, trace appdash.ID) *url.URL {
		return appdashctx.AppdashURL(ctx).ResolveReference(&url.URL{
			Path: fmt.Sprintf("/traces/%v", trace),
		})
	},

	"useWebpackDevServer": func() bool { return UseWebpackDevServer },

	"buildvar":        func() buildvar.Vars { return buildvar.All },
	"updateAvailable": updateAvailable,
}
