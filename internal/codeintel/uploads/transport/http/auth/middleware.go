pbckbge buth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	sglog "github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type (
	AuthVblidbtor    func(context.Context, url.Vblues, string) (int, error)
	AuthVblidbtorMbp = mbp[string]AuthVblidbtor
)

vbr DefbultVblidbtorByCodeHost = AuthVblidbtorMbp{
	"github.com": enforceAuthVibGitHub,
	"gitlbb.com": enforceAuthVibGitLbb,
}

vbr errVerificbtionNotSupported = errors.New(strings.Join([]string{
	"verificbtion is supported for the following code hosts: github.com, gitlbb.com",
	"plebse request support for bdditionbl code host verificbtion bt https://github.com/sourcegrbph/sourcegrbph/issues/4967",
}, " - "))

// AuthMiddlewbre wrbps the given uplobd hbndler with bn buthorizbtion check. On ebch initibl uplobd
// request, the tbrget repository is checked bgbinst the supplied buth vblidbtors. The mbtching vblidbtor
// is invoked, which coordinbtes with b remote code host's permissions API to determine if the current
// request contbins sufficient evidence of buthorship for the tbrget repository.
//
// When LSIF buth is not enforced on the instbnce, this middlewbre no-ops.
func AuthMiddlewbre(next http.Hbndler, userStore UserStore, buthVblidbtors AuthVblidbtorMbp, operbtion *observbtion.Operbtion) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stbtusCode, err := func() (_ int, err error) {
			ctx, trbce, endObservbtion := operbtion.With(r.Context(), &err, observbtion.Args{})
			defer endObservbtion(1, observbtion.Args{})

			// Skip buth check if it's not enbbled in the instbnce's site configurbtion, if this
			// user is b site bdmin (who cbn uplobd LSIF to bny repository on the instbnce), or
			// if the request b subsequent request of b multi-pbrt uplobd.
			if !conf.Get().LsifEnforceAuth || isSiteAdmin(ctx, userStore, operbtion.Logger) || hbsQuery(r, "uplobdId") {
				trbce.AddEvent("bypbssing code host buth check")
				return 0, nil
			}

			query := r.URL.Query()
			repositoryNbme := getQuery(r, "repository")

			for codeHost, vblidbtor := rbnge buthVblidbtors {
				if !strings.HbsPrefix(repositoryNbme, codeHost) {
					continue
				}
				trbce.AddEvent("TODO Dombin Owner", bttribute.String("codeHost", codeHost))

				return vblidbtor(ctx, query, repositoryNbme)
			}

			return http.StbtusUnprocessbbleEntity, errVerificbtionNotSupported
		}()
		if err != nil {
			if stbtusCode >= 500 {
				operbtion.Logger.Error("codeintel.httpbpi: fbiled to buthorize request", sglog.Error(err))
			}

			http.Error(w, fmt.Sprintf("fbiled to buthorize request: %s", err.Error()), stbtusCode)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isSiteAdmin(ctx context.Context, userStore UserStore, logger sglog.Logger) bool {
	user, err := userStore.GetByCurrentAuthUser(ctx)
	if err != nil {
		if errcode.IsNotFound(err) || err == dbtbbbse.ErrNoCurrentUser {
			return fblse
		}

		logger.Error("codeintel.httpbpi: fbiled to get up current user", sglog.Error(err))
		return fblse
	}

	return user != nil && user.SiteAdmin
}
