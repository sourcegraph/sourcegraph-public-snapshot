package bundle

import (
	"bytes"
	htmltemplate "html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"reflect"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/jscontext"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httputil"
)

// LauncherEntrypoint is the HTML template that launches the app.
var LauncherEntrypoint *htmltemplate.Template

// WorkbenchEntrypoint is the HTML template that launches the standalone workbench. This is used when something (such as the browser extension) wants to embed only the workbench in an iframe, for example.
var WorkbenchEntrypoint *htmltemplate.Template

// RenderEntrypoint renders the entrypoint template to the HTTP
// response.
func RenderEntrypoint(w http.ResponseWriter, r *http.Request, statusCode int, header http.Header, data interface{}, standaloneWorkbench bool) error {
	if Data == nil || WorkbenchEntrypoint == nil || LauncherEntrypoint == nil {
		return errNoApp
	}

	if data != nil {
		field := reflect.ValueOf(data).Elem().FieldByName("Common")
		existingCommon := field.Interface().(tmpl.Common)

		jsctx := jscontext.NewJSContextFromRequest(r)

		// Clear out sensitive data that vscode does not use.
		jsctx.LegacyAccessToken = ""
		if t := jsctx.GitHubToken; t != nil {
			t.Token = ""
		}
		jsctx.XHRHeaders = nil

		if cacheKey != "" {
			jsctx.AppRoot += "/" + cacheKey
		}

		field.Set(reflect.ValueOf(tmpl.Common{
			AuthInfo: actor.FromContext(r.Context()).AuthInfo(),
			Ctx:      r.Context(),
			Debug:    handlerutil.DebugMode,
			ErrorID:  existingCommon.ErrorID,
			JSCtx:    jsctx,
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

	bw.WriteHeader(statusCode)
	if statusCode == http.StatusNotModified {
		return nil
	}

	// HTTP/2 push of resources the client will probably need.
	if pusher, ok := w.(http.Pusher); ok {
		// Get the app root from jsctx.
		jsctx := jscontext.NewJSContextFromRequest(r)
		if appRoot, err := url.Parse(jsctx.AppRoot); err == nil {
			opt := &http.PushOptions{
				Header: http.Header{
					"Accept":          r.Header["Accept"],
					"Accept-Encoding": r.Header["Accept-Encoding"],
					"Cookie":          r.Header["Cookie"],
					"Authorization":   r.Header["Authorization"],
				},
			}
			for _, r := range pushResources {
				p := path.Join(appRoot.Path, r)
				if err := pusher.Push(p, opt); err != nil {
					log.Printf("warning: HTTP/2 push %q failed: %s", p, err)
					break
				}
			}
		}
	}

	var template *htmltemplate.Template
	if standaloneWorkbench {
		template = WorkbenchEntrypoint
	} else {
		template = LauncherEntrypoint
	}
	if err := template.Execute(&bw, data); err != nil {
		return err
	}

	return bw.WriteTo(w)
}

const (
	launcherEntrypointPath  = "/out/vs/launcher/browser/bootstrap/index.html"
	workbenchEntrypointPath = "/out/vs/workbench/browser/bootstrap/index.html"
)

func init() {
	if Data != nil {
		LauncherEntrypoint = createEntrypointTemplate(launcherEntrypointPath)
		WorkbenchEntrypoint = createEntrypointTemplate(workbenchEntrypointPath)
	}
}

func createEntrypointTemplate(entrypoint string) *htmltemplate.Template {
	f, err := Data.Open(entrypoint)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	data = bytes.Replace(data, []byte("<!-- INSERT SOURCEGRAPH CONTEXT -->"), []byte(insertHead), 1)

	template := htmltemplate.New("entrypoint")
	template.Funcs(tmpl.FuncMap)
	if _, err := template.Parse(string(data)); err != nil {
		log.Fatal("parsing entrypoint template:", err)
	}

	return template
}

const insertHead = `
<script type="text/javascript">window.__sourcegraphJSContext = {{.JSCtx}};</script>
<base href="{{.JSCtx.AppRoot}}/out/vs/launcher/browser/bootstrap/index.html" />

<title>{{with .Meta.Title}}{{.}} - {{end}}Sourcegraph</title>

{{/* Common to all pages */}}
<meta name="twitter:site" content="@srcgraph">
<meta name="twitter:image" content="{{assetURL "/img/sourcegraph-mark.png"}}">
<meta name="og:site_name" content="Sourcegraph">
{{if .Meta.Title}}
	<meta property="og:type" content="object">
	<meta name="twitter:card" content="summary">
	<meta property="og:title" content="{{.Meta.ShortTitle}}">
	<meta name="twitter:title" content="{{.Meta.ShortTitle}}">
	<meta name="twitter:description" content="{{.Meta.Description}}">
	<meta property="og:description" content="{{.Meta.Description}}">
	<meta name="description" content="{{.Meta.Description}}">
{{else}}
	{{/* Site-wide */}}
	<meta property="og:type" content="website">
	<meta name="twitter:card" content="summary_large_image">
	<meta property="og:title" content="Sourcegraph">
	<meta name="twitter:title" content="Sourcegraph">
	{{$description := "How developers discover and understand code. Sourcegraph is a fast, global, semantic code search and cross-reference engine."}}
	<meta name="twitter:description" content="{{$description}}">
	<meta property="og:description" content="{{$description}}">
	<meta name="description" content="{{$description}}">
{{end}}

{{with .Meta.CanonicalURL}}
	<link rel="canonical" href="{{.}}">
	<meta property="og:url" content="{{.}}">
{{end}}

<meta name="robots" content="{{.Meta.RobotsMetaContent}}">

<!-- Start of Telligent scripts -->
<script>
	; (function (t, r, a, c, k, e, d) { if (!t[k]) { t.GlobalTelligentNamespace = t.GlobalTelligentNamespace || []; t.GlobalTelligentNamespace.push(k); t[k] = function () { (t[k].q = t[k].q || []).push(arguments) }; t[k].q = t[k].q || []; e = r.createElement(a); d = r.getElementsByTagName(a)[0]; e.async = 1; e.src = c; d.parentNode.insertBefore(e, d) } }(window, document, "script", "https://storage.googleapis.com/telligent-artifacts/tracker/tel.js", "telligent"));
</script>
<!-- End of Telligent scripts-->

{{if not disableSupportServices}}
	<!-- Start of HubSpot script -->
	<script type="text/javascript" id="hs-script-loader" async src="//js.hs-scripts.com/2762526.js"></script>
	<!-- End of HubSpot script -->

	{{if googleAnalyticsTrackingID}}
		<!-- Start of Google Analytics script -->
		<script>
			(function(i,s,o,g,r,a,m){i["GoogleAnalyticsObject"]=r;i[r]=i[r]||function(){(i[r].q=i[r].q||[]).push(arguments)},i[r].l=1*new Date();a=s.createElement(o),m=s.getElementsByTagName(o)[0];a.async=1;a.src=g;m.parentNode.insertBefore(a,m)})(window,document,"script","https://www.google-analytics.com/analytics.js","ga");
			window.ga("create", {{googleAnalyticsTrackingID}}, "auto");
			window.ga("require", "linkid", "https://www.google-analytics.com/plugins/ua/linkid.js");
			window.ga("require", "urlChangeTracker");
			window.ga("set", "dimension1", {{if .AuthInfo.UID}}"true"{{else}}"false"{{end}});
			window.ga("set", "dimension4", "web");
			window.ga("send", "pageview");
		</script>
		<!-- End of Google Analytics script -->
	{{end}}
{{end}}
`

// This list should be periodically updated to be in sync with the
// unpushed resources loaded over the network when a browser loads the
// app.
//
// TODO(sqs): It would be nice but might not be worth the effort to
// generate this list automatically.
var pushResources = []string{
	"out/vs/launcher/browser/bootstrap/index.js",

	"out/vs/workbench/browser/bootstrap/config.js",
	"out/vs/workbench/browser/bootstrap/index.js",
	"out/vs/loader.js",
	"out/vs/code/browser/main.js",
	"out/vs/code/browser/main.css",
	"out/vs/code/browser/main.nls.js",

	"extensions/diff/package.json",
	"extensions/diff/language-configuration.json",
	"extensions/docker/package.json",
	"extensions/file-links/package.json",
	"extensions/gitsyntax/package.json",
	"extensions/go/package.json",
	"extensions/json/package.json",
	"extensions/lsp/package.json",
	"extensions/markdown/package.json",
	"extensions/theme-abyss/package.json",
	"extensions/theme-defaults/package.json",
	"extensions/theme-kimbie-dark/package.json",
	"extensions/theme-monokai/package.json",
	"extensions/theme-monokai-dimmed/package.json",
	"extensions/theme-quietlight/package.json",
	"extensions/theme-red/package.json",
	"extensions/theme-seti/package.json",
	"extensions/theme-solarized-dark/package.json",
	"extensions/theme-solarized-light/package.json",
	"extensions/theme-sourcegraph/package.json",
	"extensions/theme-tomorrow-night-blue/package.json",

	"out/vs/workbench/browser/extensionHostProcess.js",
	"out/vs/workbench/browser/extensionHostProcess.nls.js",
	"out/browser_modules/lsp.js",
}
