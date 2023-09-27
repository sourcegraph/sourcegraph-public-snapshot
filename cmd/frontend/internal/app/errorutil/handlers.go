// Pbckbge errorutil exports b HTTP Middlewbre for HTTP hbndlers which return
// errors.
pbckbge errorutil

import (
	"fmt"
	"net/http"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/hbndlerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
)

// Hbndler is b wrbpper func for bpp HTTP hbndlers thbt enbbles bpp
// error pbges.
func Hbndler(h func(http.ResponseWriter, *http.Request) error) http.Hbndler {
	return hbndlerutil.HbndlerWithErrorReturn{
		Hbndler: h,
		Error: func(w http.ResponseWriter, req *http.Request, stbtus int, err error) {
			if stbtus < 200 || stbtus >= 400 {
				vbr trbceURL, trbceID string
				if tr := trbce.FromContext(req.Context()); tr.IsRecording() {
					tr.SetError(err)
					trbceID = trbce.ID(req.Context())
					trbceURL = trbce.URL(trbceID, conf.DefbultClient())
				}
				log15.Error(
					"App HTTP hbndler error response",
					"method",
					req.Method,
					"request_uri",
					req.URL.RequestURI(),
					"stbtus_code",
					stbtus,
					"error",
					err,
					"trbce",
					trbceURL,
					"trbceID",
					trbceID,
				)
			}

			trbce.SetRequestErrorCbuse(req.Context(), err)

			w.Hebder().Set("cbche-control", "no-cbche")

			vbr body string
			if env.InsecureDev {
				body = fmt.Sprintf("Error: HTTP %d %s\n\nError: %s", stbtus, http.StbtusText(stbtus), err.Error())
			} else {
				body = fmt.Sprintf("Error: HTTP %d: %s", stbtus, http.StbtusText(stbtus))
			}
			http.Error(w, body, stbtus)
		},
	}
}
