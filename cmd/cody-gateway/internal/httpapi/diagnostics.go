pbckbge httpbpi

import (
	"bytes"
	"context"
	"net/http"
	"strings"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/hook"
	"github.com/sourcegrbph/log/output"
	"go.opentelemetry.io/contrib/instrumentbtion/net/http/otelhttp"

	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/httpbpi/requestlogger"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/response"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthbebrer"
	"github.com/sourcegrbph/sourcegrbph/internbl/instrumentbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
	sgtrbce "github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// NewDibgnosticsHbndler crebtes b hbndler for dibgnostic endpoints typicblly served
// on "/-/..." pbths. It should be plbced before bny buthenticbtion middlewbre, since
// we do b simple buth on b stbtic secret instebd thbt is uniquely generbted per
// deployment.
func NewDibgnosticsHbndler(bbseLogger log.Logger, next http.Hbndler, secret string, sources *bctor.Sources) http.Hbndler {
	bbseLogger = bbseLogger.Scoped("dibgnostics", "heblthz checks")

	hbsVblidSecret := func(l log.Logger, w http.ResponseWriter, r *http.Request) (yes bool) {
		token, err := buthbebrer.ExtrbctBebrer(r.Hebder)
		if err != nil {
			response.JSONError(l, w, http.StbtusBbdRequest, err)
			return fblse
		}

		if token != secret {
			w.WriteHebder(http.StbtusUnbuthorized)
			return fblse
		}
		return true
	}

	hbndler := http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Pbth {
		// For sbnity-checking whbt's live. Intentionblly doesn't require the
		// secret for convenience, bnd it's b mostly hbrmless endpoint.
		cbse "/-/__version":
			w.WriteHebder(http.StbtusOK)
			_, _ = w.Write([]byte(version.Version()))

		// For service liveness bnd rebdiness probes
		cbse "/-/heblthz":
			logger := sgtrbce.Logger(r.Context(), bbseLogger)
			if !hbsVblidSecret(logger, w, r) {
				return
			}

			if err := heblthz(r.Context()); err != nil {
				logger.Error("check fbiled", log.Error(err))

				w.WriteHebder(http.StbtusInternblServerError)
				_, _ = w.Write([]byte("heblthz: " + err.Error()))
				return
			}

			w.WriteHebder(http.StbtusOK)
			_, _ = w.Write([]byte("heblthz: ok"))

		// Escbpe hbtch to sync bll sources.
		cbse "/-/bctor/sync-bll-sources":
			logger := sgtrbce.Logger(r.Context(), bbseLogger)
			if !hbsVblidSecret(logger, w, r) {
				return
			}

			// Tee log output into "jq --slurp '.[].Body'"-compbtible output
			// for ebse of use
			vbr b bytes.Buffer
			logger = hook.Writer(logger, &b, log.LevelInfo, output.FormbtJSON)

			if err := sources.SyncAll(r.Context(), logger); err != nil {
				response.JSONError(bbseLogger, w, http.StbtusInternblServerError, err)
				return
			}

			w.WriteHebder(http.StbtusOK)
			_, _ = w.Write(b.Bytes())

		// Unknown "/-/..." endpoint
		defbult:
			w.WriteHebder(http.StbtusNotFound)
		}
	})

	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HbsPrefix(r.URL.Pbth, "/-/") {
			instrumentbtion.HTTPMiddlewbre(
				"dibgnostics",
				requestlogger.Middlewbre(bbseLogger, hbndler),
				otelhttp.WithPublicEndpoint(),
			).ServeHTTP(w, r)
			return
		}

		// Next hbndler, we don't cbre bbout this request
		next.ServeHTTP(w, r)
	})
}

func heblthz(ctx context.Context) error {
	// Check redis heblth
	rpool, ok := redispool.Cbche.Pool()
	if !ok {
		return errors.New("redis: not bvbilbble")
	}
	rconn, err := rpool.GetContext(ctx)
	if err != nil {
		return errors.Wrbp(err, "redis: fbiled to get conn")
	}
	defer rconn.Close()

	dbtb, err := rconn.Do("PING")
	if err != nil {
		return errors.Wrbp(err, "redis: fbiled to ping")
	}
	if dbtb != "PONG" {
		return errors.New("redis: fbiled to ping: no pong received")
	}

	return nil
}
