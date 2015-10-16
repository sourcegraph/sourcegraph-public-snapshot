package app

import (
	"encoding/json"
	"fmt"
	htmpl "html/template"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/srclib/toolchain"

	"src.sourcegraph.com/sourcegraph/app/appconf"
	"src.sourcegraph.com/sourcegraph/app/internal/schemautil"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/search"
	"src.sourcegraph.com/sourcegraph/sgx/buildvar"
	"src.sourcegraph.com/sourcegraph/sourcecode"
	"src.sourcegraph.com/sourcegraph/util/envutil"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/textutil"
	"src.sourcegraph.com/sourcegraph/util/timeutil"
	"src.sourcegraph.com/sourcegraph/util/traceutil/appdashctx"
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

	"repoMetaDescription": repoMetaDescription,
	"repoStat":            repoStat,

	"defQualifiedName":            sourcecode.DefQualifiedName,
	"defQualifiedNameAndType":     sourcecode.DefQualifiedNameAndType,
	"overrideStyleViaRegexpFlags": sourcecode.OverrideStyleViaRegexpFlags,

	"appconf":   func() interface{} { return &appconf.Flags },
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
	"json": func(v interface{}) (string, error) {
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(b), nil
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

	"customLogo":         func() htmpl.HTML { return appconf.Flags.CustomLogo },
	"motd":               func() htmpl.HTML { return appconf.Flags.MOTD },
	"customFeedbackForm": func() htmpl.HTML { return appconf.Flags.CustomFeedbackForm },
	"uiBuild":            func() bool { return !appconf.Flags.NoUIBuild },

	"trimPrefix": strings.TrimPrefix,

	"absURL": func(appURL, other *url.URL) *url.URL { return appURL.ResolveReference(other) },

	"urlTo":                      router.Rel.URLTo,
	"urlToBlogPost":              router.Rel.URLToBlogPost,
	"urlToBlogAtomFeed":          router.Rel.URLToBlogAtomFeed,
	"urlToUser":                  router.Rel.URLToUser,
	"urlToUserSubroute":          router.Rel.URLToUserSubroute,
	"urlToRepo":                  router.Rel.URLToRepo,
	"urlToRepoRev":               router.Rel.URLToRepoRev,
	"urlToRepoDiscussion":        router.Rel.URLToRepoDiscussion,
	"urlToRepoBuild":             router.Rel.URLToRepoBuild,
	"urlToRepoSubroute":          router.Rel.URLToRepoSubroute,
	"urlToRepoSubrouteRev":       router.Rel.URLToRepoSubrouteRev,
	"urlToRepoTreeEntry":         router.Rel.URLToRepoTreeEntry,
	"urlToRepoTreeEntrySubroute": router.Rel.URLToRepoTreeEntrySubroute,
	"urlToRepoCommit":            router.Rel.URLToRepoCommit,
	"urlToRepoCompare":           router.Rel.URLToRepoCompare,
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

	"isOAuthInitPage": func(tmplName string) bool { return tmplName == "oauth-client/initiate.html" },

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

	"ifTemplate":                ifTemplate,
	"googleAnalyticsTrackingID": func() string { return appconf.Flags.GoogleAnalyticsTrackingID },
	"heapAnalyticsID":           func() string { return appconf.Flags.HeapAnalyticsID },

	"deployedGitCommitID": func() string { return envutil.GitCommitID },
	"hostname":            func() string { return hostname },

	"nl2br": func(s string) htmpl.HTML {
		return htmpl.HTML(strings.Replace(htmpl.HTMLEscapeString(s), "\n", "<br>", -1))
	},

	"showRepoRevSwitcher": showRepoRevSwitcher,

	"repoEnabledFrames":  repoEnabledFrames,
	"showSearchForm":     showSearchForm,
	"fileSearchDisabled": func() bool { return appconf.Flags.DisableSearch },
	"disableCloneURL":    func() bool { return appconf.Flags.DisableCloneURL },

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
