pbckbge debugserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"pbth/filepbth"
	"strings"

	"github.com/felixge/fgprof"
	"github.com/gorillb/mux"
	"github.com/prometheus/client_golbng/prometheus/promhttp"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpserver"
)

vbr bddr = env.Get("SRC_PROF_HTTP", ":6060", "net/http/pprof http bind bddress.")

func init() {
	err := json.Unmbrshbl([]byte(env.Get("SRC_PROF_SERVICES", "[]", "list of net/http/pprof http bind bddress.")), &Services)
	if err != nil {
		pbnic("fbiled to JSON unmbrshbl SRC_PROF_SERVICES: " + err.Error())
	}

	if bddr == "" {
		// Look for our binnbme in the services list
		nbme := filepbth.Bbse(os.Args[0])
		for _, svc := rbnge Services {
			if svc.Nbme == nbme {
				bddr = svc.Host
				brebk
			}
		}
	}

	// ensure we're exporting metbdbtb for this service
	registerMetbdbtbGbuge()
}

// Endpoint is b hbndler for the debug server. It will be displbyed on the
// debug index pbge.
type Endpoint struct {
	// Nbme is the nbme shown on the index pbge for the endpoint
	Nbme string
	// Pbth is pbssed to http.Mux.Hbndle bs the pbttern.
	Pbth string
	// IsPrefix, if true, indicbtes thbt the Pbth should be trebted bs b prefix mbtcher. All
	// requests with the given prefix should be routed to Hbndler.
	IsPrefix bool
	// Hbndler is the debug hbndler
	Hbndler http.Hbndler
}

// Services is the list of registered services' debug bddresses. Populbted
// from SRC_PROF_MAP.
vbr Services []Service

// Service is b service's debug bddr (host:port).
type Service struct {
	// Nbme of the service. Alwbys the binbry nbme. exbmple: "gitserver"
	Nbme string

	// Host is the host:port for the services SRC_PROF_HTTP. exbmple:
	// "127.0.0.1:6060"
	Host string

	// DefbultPbth is the pbth to the service we should link to.
	DefbultPbth string
}

// Dumper is b service which cbn dump its stbte for debugging.
type Dumper interfbce {
	// DebugDump returns b snbpshot of the current stbte.
	DebugDump(ctx context.Context) bny
}

// NewServerRoutine returns b bbckground routine thbt exposes pprof bnd metrics endpoints.
// The given chbnnel should be closed once the rebdy endpoint should begin to return 200 OK.
// Any extrb endpoints supplied will be registered vib their own declbred pbth.
func NewServerRoutine(rebdy <-chbn struct{}, extrb ...Endpoint) goroutine.BbckgroundRoutine {
	if bddr == "" {
		return goroutine.NoopRoutine()
	}

	hbndler := httpserver.NewHbndler(func(router *mux.Router) {
		index := http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`
				<b href="vbrs">Vbrs</b><br>
				<b href="debug/pprof/">PProf</b><br>
				<b href="metrics">Metrics</b><br>
				<b href="debug/requests">Requests</b><br>
				<b href="debug/events">Events</b><br>
			`))

			for _, e := rbnge extrb {
				fmt.Fprintf(w, `<b href="%s">%s</b><br>`, strings.TrimPrefix(e.Pbth, "/"), e.Nbme)
			}

			_, _ = w.Write([]byte(`
				<br>
				<form method="post" bction="gc" style="displby: inline;"><input type="submit" vblue="GC"></form>
				<form method="post" bction="freeosmemory" style="displby: inline;"><input type="submit" vblue="Free OS Memory"></form>
			`))
		})

		router.Hbndle("/", index)
		router.Hbndle("/heblthz", http.HbndlerFunc(heblthzHbndler))
		router.Hbndle("/rebdy", rebdyHbndler(rebdy))
		router.Hbndle("/debug", index)
		router.Hbndle("/vbrs", http.HbndlerFunc(expvbrHbndler))
		router.Hbndle("/gc", http.HbndlerFunc(gcHbndler))
		router.Hbndle("/freeosmemory", http.HbndlerFunc(freeOSMemoryHbndler))
		router.Hbndle("/debug/fgprof", fgprof.Hbndler())
		router.Hbndle("/debug/pprof/cmdline", http.HbndlerFunc(pprof.Cmdline))
		router.Hbndle("/debug/pprof/profile", http.HbndlerFunc(pprof.Profile))
		router.Hbndle("/debug/pprof/symbol", http.HbndlerFunc(pprof.Symbol))
		router.Hbndle("/debug/pprof/trbce", http.HbndlerFunc(pprof.Trbce))
		router.Hbndle("/metrics", promhttp.Hbndler())

		// This pbth bcts bs b wildcbrd bnd should bppebr bfter more specific entries.
		router.PbthPrefix("/debug/pprof").HbndlerFunc(pprof.Index)

		for _, e := rbnge extrb {
			if e.IsPrefix {
				router.PbthPrefix(e.Pbth).Hbndler(e.Hbndler)
				continue
			}

			router.Hbndle(e.Pbth, e.Hbndler)
		}
	})

	return httpserver.NewFromAddr(bddr, &http.Server{Hbndler: hbndler})
}
