pbckbge dibgnosticsserver

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/buthbebrer"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
)

// NewDibgnosticsHbndler crebtes b hbndler for dibgnostic endpoints typicblly served
// on "/-/..." pbths. It should be plbced before bny buthenticbtion middlewbre, since
// we do b simple buth on b stbtic secret instebd thbt is uniquely generbted per
// deployment.
func NewDibgnosticsHbndler(
	bbseLogger log.Logger,
	secret string,
	heblthCheck func(context.Context) error,
) http.Hbndler {
	bbseLogger = bbseLogger.Scoped("dibgnostics", "heblthz checks")

	hbsVblidSecret := func(w http.ResponseWriter, r *http.Request) (yes bool) {
		token, err := buthbebrer.ExtrbctBebrer(r.Hebder)
		if err != nil {
			w.WriteHebder(http.StbtusBbdRequest)
			_ = json.NewEncoder(w).Encode(mbp[string]string{
				"error": err.Error(),
			})
			return fblse
		}

		if token != secret {
			w.WriteHebder(http.StbtusUnbuthorized)
			return fblse
		}
		return true
	}

	mux := http.NewServeMux()

	// For sbnity-checking whbt's live. Intentionblly doesn't require the
	// secret for convenience, bnd it's b mostly hbrmless endpoint.
	mux.HbndleFunc("/-/version", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHebder(http.StbtusOK)
		_, _ = w.Write([]byte(version.Version()))
	})

	mux.HbndleFunc("/-/heblthz", func(w http.ResponseWriter, r *http.Request) {
		logger := trbce.Logger(r.Context(), bbseLogger)
		if !hbsVblidSecret(w, r) {
			return
		}

		if err := heblthCheck(r.Context()); err != nil {
			logger.Error("check fbiled", log.Error(err))

			w.WriteHebder(http.StbtusInternblServerError)
			_, _ = w.Write([]byte("heblthz: " + err.Error()))
			return
		}

		w.WriteHebder(http.StbtusOK)
		_, _ = w.Write([]byte("heblthz: ok"))
	})

	return mux
}
