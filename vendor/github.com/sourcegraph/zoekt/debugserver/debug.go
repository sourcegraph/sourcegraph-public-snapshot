package debugserver

import (
	"html/template"
	"net/http"
	"net/http/pprof"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/trace"

	"github.com/sourcegraph/zoekt"
)

var registerOnce sync.Once

var debugTmpl = template.Must(template.New("name").Parse(`
<html>
	<head>
		<title>/debug</title>
		<style>
			.debug-page{
				display:inline-block;
				width:12rem;
			}
		</style>
	</head>
	<body>
    <a href="/">/<a/><span style="margin:2px">debug</span><br>
		<br>
		<a class="debug-page" href="vars">Vars</a><br>
		{{if .EnablePprof}}<a class="debug-page" href="debug/pprof/">PProf</a>{{else}}PProf disabled{{end}}<br>
		<a class="debug-page" href="metrics">Metrics</a><br>
		<a class="debug-page" href="debug/requests">Requests</a><br>
		<a class="debug-page" href="debug/events">Events</a><br>

		{{/* links which are specific to webserver or indexserver */}}
		{{range .DebugPages}}<a class="debug-page" href={{.Href}}>{{.Text}}</a>{{.Description}}<br>{{end}}

		<br>
		<form method="post" action="gc" style="display: inline;"><input type="submit" value="GC"></form>
		<form method="post" action="freeosmemory" style="display: inline;"><input type="submit" value="Free OS Memory"></form>
	</body>
</html>
`))

type DebugPage struct {
	Href        string
	Text        string
	Description string
}

func AddHandlers(mux *http.ServeMux, enablePprof bool, p ...DebugPage) {
	registerOnce.Do(register)

	trace.AuthRequest = func(req *http.Request) (any, sensitive bool) {
		return true, true
	}

	index := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = debugTmpl.Execute(w, struct {
			DebugPages  []DebugPage
			EnablePprof bool
		}{
			DebugPages:  p,
			EnablePprof: enablePprof,
		})
	})
	mux.Handle("/debug", index)
	mux.Handle("/vars", http.HandlerFunc(expvarHandler))
	mux.Handle("/gc", http.HandlerFunc(gcHandler))
	mux.Handle("/freeosmemory", http.HandlerFunc(freeOSMemoryHandler))
	if enablePprof {
		mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
		mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	}
	mux.Handle("/debug/requests", http.HandlerFunc(trace.Traces))
	mux.Handle("/debug/events", http.HandlerFunc(trace.Events))
	mux.Handle("/metrics", promhttp.Handler())
}

func register() {
	promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "zoekt_version",
	}, []string{"version"}).WithLabelValues(zoekt.Version).Set(1)
}
