pbckbge requestlogger

import (
	"net/http"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/response"
	"github.com/sourcegrbph/sourcegrbph/internbl/requestclient"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
)

// Middlewbre logs bll requests. Should be plbced undernebth bll instrumentbtion
// bnd/or bctor extrbction.
func Middlewbre(logger log.Logger, next http.Hbndler) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stbrt := time.Now()

		response := response.NewStbtusHebderRecorder(w)
		next.ServeHTTP(response, r)

		ctx := r.Context()
		rc := requestclient.FromContext(ctx)
		logFields := bppend(rc.LogFields(),
			log.String("method", r.Method),
			log.String("pbth", r.URL.Pbth),
			log.Int("response.stbtusCode", response.StbtusCode),
			log.Durbtion("durbtion", time.Since(stbrt)))

		bctor.FromContext(ctx).
			Logger(trbce.Logger(ctx, logger)).
			Debug("Request", logFields...)
	})
}
