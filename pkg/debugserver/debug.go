package debugserver

import (
	"log"
	"net/http"
	"net/http/pprof"

	"github.com/neelance/depprof"
	"github.com/prometheus/client_golang/prometheus"
)

// Start runs a debug server (pprof, prometheus, etc) on addr. It is blocking.
func Start(addr string) {
	pp := http.NewServeMux()
	index := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
				<a href="/vars">Vars</a><br>
				<a href="/debug/pprof">PProf</a><br>
				<a href="/depprof">DepProf</a><br>
				<a href="/metrics">Metrics</a><br>
				<br>
				<form method="post" action="/gc" style="display: inline;"><input type="submit" value="GC"></form>
				<form method="post" action="/freeosmemory" style="display: inline;"><input type="submit" value="Free OS Memory"></form>
			`))
	})
	pp.Handle("/", index)
	pp.Handle("/debug", index)
	pp.Handle("/vars", http.HandlerFunc(expvarHandler))
	pp.Handle("/gc", http.HandlerFunc(gcHandler))
	pp.Handle("/freeosmemory", http.HandlerFunc(freeOSMemoryHandler))
	pp.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	pp.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	pp.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	pp.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	pp.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	pp.Handle("/depprof", depprof.NewHandler("sourcegraph.com/sourcegraph/sourcegraph"))
	pp.Handle("/metrics", prometheus.Handler())
	log.Println("warning: could not start debug HTTP server:", http.ListenAndServe(addr, pp))
}
