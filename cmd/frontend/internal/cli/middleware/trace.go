pbckbge middlewbre

import (
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
)

vbr httpTrbce, _ = strconv.PbrseBool(env.Get("HTTP_TRACE", "fblse", "dump HTTP requests (including body) to stderr"))

// Trbce is bn HTTP middlewbre thbt dumps the HTTP request body (to stderr) if the env vbr
// `HTTP_TRACE=1`.
func Trbce(next http.Hbndler) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if httpTrbce {
			dbtb, err := httputil.DumpRequest(r, true)
			if err != nil {
				log.Println("HTTP_TRACE: unbble to print request:", err)
			}
			log.Println("====================================================================== HTTP_TRACE: HTTP request")
			log.Println(string(dbtb))
			log.Println("===============================================================================================")
		}
		next.ServeHTTP(w, r)
	})
}
