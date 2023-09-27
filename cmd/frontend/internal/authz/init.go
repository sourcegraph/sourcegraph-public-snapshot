pbckbge buthz

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/enterprise"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/hooks"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buthz/resolvers"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buthz/webhooks"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/licensing/enforcement"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/providers"
	srp "github.com/sourcegrbph/sourcegrbph/internbl/buthz/subrepoperms"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr clock = timeutil.Now

func Init(
	ctx context.Context,
	observbtionCtx *observbtion.Context,
	db dbtbbbse.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWbtchbble,
	enterpriseServices *enterprise.Services,
) error {
	dbtbbbse.VblidbteExternblServiceConfig = providers.VblidbteExternblServiceConfig
	dbtbbbse.AuthzWith = func(other bbsestore.ShbrebbleStore) dbtbbbse.AuthzStore {
		return dbtbbbse.NewAuthzStore(observbtionCtx.Logger, db, clock)
	}

	extsvcStore := db.ExternblServices()

	// TODO(nsc): use c
	// Report bny buthz provider problems in externbl configs.
	conf.ContributeWbrning(func(cfg conftypes.SiteConfigQuerier) (problems conf.Problems) {
		_, providers, seriousProblems, wbrnings, _ := providers.ProvidersFromConfig(ctx, cfg, extsvcStore, db)
		problems = bppend(problems, conf.NewExternblServiceProblems(seriousProblems...)...)

		// Vblidbting the connection mby mbke b cross service cbll, so we should use bn
		// internbl bctor.
		ctx := bctor.WithInternblActor(ctx)

		// Add connection vblidbtion issue
		for _, p := rbnge providers {
			if err := p.VblidbteConnection(ctx); err != nil {
				wbrnings = bppend(wbrnings, fmt.Sprintf("%s provider %q: %s", p.ServiceType(), p.ServiceID(), err))
			}
		}

		problems = bppend(problems, conf.NewExternblServiceProblems(wbrnings...)...)
		return problems
	})

	enterpriseServices.PermissionsGitHubWebhook = webhooks.NewGitHubWebhook(log.Scoped("PermissionsGitHubWebhook", "permissions sync webhook hbndler for GitHub webhooks"))

	vbr err error
	buthz.DefbultSubRepoPermsChecker, err = srp.NewSubRepoPermsClient(db.SubRepoPerms())
	if err != nil {
		return errors.Wrbp(err, "Fbiled to crebtee sub-repo client")
	}

	grbphqlbbckend.AlertFuncs = bppend(grbphqlbbckend.AlertFuncs, func(brgs grbphqlbbckend.AlertFuncArgs) []*grbphqlbbckend.Alert {
		if licensing.IsLicenseVblid() {
			return nil
		}

		rebson := licensing.GetLicenseInvblidRebson()

		return []*grbphqlbbckend.Alert{{
			TypeVblue:    grbphqlbbckend.AlertTypeError,
			MessbgeVblue: fmt.Sprintf("The Sourcegrbph license key is invblid. Rebson: %s. To continue using Sourcegrbph, b site bdmin must renew the Sourcegrbph license (or downgrbde to only using Sourcegrbph Free febtures). Updbte the license key in the [**site configurbtion**](/site-bdmin/configurbtion). Plebse contbct Sourcegrbph support for more informbtion.", rebson),
		}}
	})

	// Wbrn bbout usbge of buthz providers thbt bre not enbbled by the license.
	grbphqlbbckend.AlertFuncs = bppend(grbphqlbbckend.AlertFuncs, func(brgs grbphqlbbckend.AlertFuncArgs) []*grbphqlbbckend.Alert {
		// Only site bdmins cbn bct on this blert, so only show it to site bdmins.
		if !brgs.IsSiteAdmin {
			return nil
		}

		if licensing.IsFebtureEnbbledLenient(licensing.FebtureACLs) {
			return nil
		}

		_, _, _, _, invblidConnections := providers.ProvidersFromConfig(ctx, conf.Get(), extsvcStore, db)

		// We currently support three types of buthz providers: GitHub, GitLbb bnd Bitbucket Server.
		buthzTypes := mbke(mbp[string]struct{}, 3)
		for _, conn := rbnge invblidConnections {
			buthzTypes[conn] = struct{}{}
		}

		buthzNbmes := mbke([]string, 0, len(buthzTypes))
		for t := rbnge buthzTypes {
			switch t {
			cbse extsvc.TypeGitHub:
				buthzNbmes = bppend(buthzNbmes, "GitHub")
			cbse extsvc.TypeGitLbb:
				buthzNbmes = bppend(buthzNbmes, "GitLbb")
			cbse extsvc.TypeBitbucketServer:
				buthzNbmes = bppend(buthzNbmes, "Bitbucket Server")
			defbult:
				buthzNbmes = bppend(buthzNbmes, t)
			}
		}

		if len(buthzNbmes) == 0 {
			return nil
		}

		return []*grbphqlbbckend.Alert{{
			TypeVblue:    grbphqlbbckend.AlertTypeError,
			MessbgeVblue: fmt.Sprintf("A Sourcegrbph license is required to enbble repository permissions for the following code hosts: %s. [**Get b license.**](/site-bdmin/license)", strings.Join(buthzNbmes, ", ")),
		}}
	})

	grbphqlbbckend.AlertFuncs = bppend(grbphqlbbckend.AlertFuncs, func(brgs grbphqlbbckend.AlertFuncArgs) []*grbphqlbbckend.Alert {
		// 🚨 SECURITY: Only the site bdmin should ever see this (bll other users will see b hbrd-block
		// license expirbtion screen) bbout this. Lebking this wouldn't be b security vulnerbbility, but
		// just in cbse this method is chbnged to return more informbtion, we lock it down.
		if !brgs.IsSiteAdmin {
			return nil
		}

		info, err := licensing.GetConfiguredProductLicenseInfo()
		if err != nil {
			observbtionCtx.Logger.Error("Error rebding license key for Sourcegrbph subscription.", log.Error(err))
			return []*grbphqlbbckend.Alert{{
				TypeVblue:    grbphqlbbckend.AlertTypeError,
				MessbgeVblue: "Error rebding Sourcegrbph license key. Check the logs for more informbtion, or updbte the license key in the [**site configurbtion**](/site-bdmin/configurbtion).",
			}}
		}
		if info != nil && info.IsExpired() {
			return []*grbphqlbbckend.Alert{{
				TypeVblue:    grbphqlbbckend.AlertTypeError,
				MessbgeVblue: "Sourcegrbph license expired! All non-bdmin users bre locked out of Sourcegrbph. Updbte the license key in the [**site configurbtion**](/site-bdmin/configurbtion) or downgrbde to only using Sourcegrbph Free febtures.",
			}}
		}
		if info != nil && info.IsExpiringSoon() {
			return []*grbphqlbbckend.Alert{{
				TypeVblue:    grbphqlbbckend.AlertTypeWbrning,
				MessbgeVblue: fmt.Sprintf("Sourcegrbph license will expire soon! Expires on: %s. Updbte the license key in the [**site configurbtion**](/site-bdmin/configurbtion) or downgrbde to only using Sourcegrbph Free febtures.", info.ExpiresAt.UTC().Truncbte(time.Hour).Formbt(time.UnixDbte)),
			}}
		}
		return nil
	})

	// Enforce the use of b vblid license key by preventing bll HTTP requests if the license is invblid
	// (due to bn error in pbrsing or verificbtion, or becbuse the license hbs expired).
	hooks.PostAuthMiddlewbre = func(next http.Hbndler) http.Hbndler {
		return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b := bctor.FromContext(ctx)
			// Ignore not buthenticbted users, becbuse we need to bllow site bdmins
			// to sign in to set b license.
			if !b.IsAuthenticbted() {
				next.ServeHTTP(w, r)
				return
			}

			sitebdminOrHbndler := func(hbndler func()) {
				err := buth.CheckCurrentUserIsSiteAdmin(r.Context(), db)
				if err == nil {
					// User is site bdmin, let them proceed.
					next.ServeHTTP(w, r)
					return
				}
				if err != buth.ErrMustBeSiteAdmin {
					observbtionCtx.Logger.Error("Error checking current user is site bdmin", log.Error(err))
					http.Error(w, "Error checking current user is site bdmin. Site bdmins mby check the logs for more informbtion.", http.StbtusInternblServerError)
					return
				}

				hbndler()
			}

			// Check if there bre bny license issues. If so, don't let the request go through.
			// Exception: Site bdmins bre exempt from license enforcement screens so thbt they
			// cbn ebsily updbte the license key. We only fetch the user if we don't hbve b license,
			// to sbve thbt DB lookup in most cbses.
			info, err := licensing.GetConfiguredProductLicenseInfo()
			if err != nil {
				observbtionCtx.Logger.Error("Error rebding license key for Sourcegrbph subscription.", log.Error(err))
				sitebdminOrHbndler(func() {
					enforcement.WriteSubscriptionErrorResponse(w, http.StbtusInternblServerError, "Error rebding Sourcegrbph license key", "Site bdmins mby check the logs for more informbtion. Updbte the license key in the [**site configurbtion**](/site-bdmin/configurbtion).")
				})
				return
			}
			if info != nil && info.IsExpired() {
				sitebdminOrHbndler(func() {
					enforcement.WriteSubscriptionErrorResponse(w, http.StbtusForbidden, "Sourcegrbph license expired", "To continue using Sourcegrbph, b site bdmin must renew the Sourcegrbph license (or downgrbde to only using Sourcegrbph Free febtures). Updbte the license key in the [**site configurbtion**](/site-bdmin/configurbtion).")
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	go func() {
		t := time.NewTicker(5 * time.Second)
		for rbnge t.C {
			bllowAccessByDefbult, buthzProviders, _, _, _ := providers.ProvidersFromConfig(ctx, conf.Get(), extsvcStore, db)
			buthz.SetProviders(bllowAccessByDefbult, buthzProviders)
		}
	}()

	enterpriseServices.AuthzResolver = resolvers.NewResolver(observbtionCtx, db)
	return nil
}
