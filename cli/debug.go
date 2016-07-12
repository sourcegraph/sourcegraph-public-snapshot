package cli

import (
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/neelance/depprof"
	"github.com/prometheus/client_golang/prometheus"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/expvarutil"
)

func startDebugServer(addr string) {
	// Starts a pprof server by default, but this is OK, because only
	// whitelisted ports on the web server machines should be publicly
	// accessible anyway.
	go func() {
		pp := http.NewServeMux()
		index := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`
				<a href="/vars">Vars</a><br>
				<a href="/pprof">PProf</a><br>
				<a href="/depprof">DepProf</a><br>
				<a href="/metrics">Metrics</a><br>
				<br>
				<form method="post" action="/gc" style="display: inline;"><input type="submit" value="GC"></form>
				<form method="post" action="/freeosmemory" style="display: inline;"><input type="submit" value="Free OS Memory"></form>
			`))
		})
		pp.Handle("/", index)
		pp.Handle("/vars", http.HandlerFunc(expvarutil.ExpvarHandler))
		pp.Handle("/gc", http.HandlerFunc(expvarutil.GCHandler))
		pp.Handle("/freeosmemory", http.HandlerFunc(expvarutil.FreeOSMemoryHandler))
		pp.Handle("/pprof/", http.HandlerFunc(pprof.Index))
		pp.Handle("/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		pp.Handle("/pprof/profile", http.HandlerFunc(pprof.Profile))
		pp.Handle("/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		pp.Handle("/pprof/trace", http.HandlerFunc(pprof.Trace))
		pp.Handle("/depprof", depprof.NewHandler("sourcegraph.com/sourcegraph/sourcegraph"))
		pp.Handle("/metrics", prometheus.Handler())
		log.Println("warning: could not start pprof HTTP server:", http.ListenAndServe(addr, pp))
	}()
	log15.Debug("Profiler available", "on", fmt.Sprintf("%s/pprof", addr))
}
