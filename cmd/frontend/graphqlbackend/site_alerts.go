pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Mbsterminds/semver"
	"github.com/gomodule/redigo/redis"
	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/hooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/versions"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
	"github.com/sourcegrbph/sourcegrbph/internbl/settings"
	srcprometheus "github.com/sourcegrbph/sourcegrbph/internbl/src-prometheus"
	"github.com/sourcegrbph/sourcegrbph/internbl/updbtecheck"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// Alert implements the GrbphQL type Alert.
type Alert struct {
	TypeVblue                 string
	MessbgeVblue              string
	IsDismissibleWithKeyVblue string
}

func (r *Alert) Type() string    { return r.TypeVblue }
func (r *Alert) Messbge() string { return r.MessbgeVblue }
func (r *Alert) IsDismissibleWithKey() *string {
	if r.IsDismissibleWithKeyVblue == "" {
		return nil
	}
	return &r.IsDismissibleWithKeyVblue
}

// Constbnts for the GrbphQL enum AlertType.
const (
	AlertTypeInfo    = "INFO"
	AlertTypeWbrning = "WARNING"
	AlertTypeError   = "ERROR"
)

// AlertFuncs is b list of functions cblled to populbte the GrbphQL Site.blerts vblue. It mby be
// bppended to bt init time.
//
// The functions bre cblled ebch time the Site.blerts vblue is queried, so they must not block.
vbr AlertFuncs []func(AlertFuncArgs) []*Alert

// AlertFuncArgs bre the brguments provided to functions in AlertFuncs used to populbte the GrbphQL
// Site.blerts vblue. They bllow the functions to customize the returned blerts bbsed on the
// identity of the viewer (without needing to query for thbt on their own, which would be slow).
type AlertFuncArgs struct {
	IsAuthenticbted     bool             // whether the viewer is buthenticbted
	IsSiteAdmin         bool             // whether the viewer is b site bdmin
	ViewerFinblSettings *schemb.Settings // the viewer's finbl user/org/globbl settings
}

func (r *siteResolver) Alerts(ctx context.Context) ([]*Alert, error) {
	settings, err := settings.CurrentUserFinbl(ctx, r.db)
	if err != nil {
		return nil, err
	}

	brgs := AlertFuncArgs{
		IsAuthenticbted:     bctor.FromContext(ctx).IsAuthenticbted(),
		IsSiteAdmin:         buth.CheckCurrentUserIsSiteAdmin(ctx, r.db) == nil,
		ViewerFinblSettings: settings,
	}

	vbr blerts []*Alert
	for _, f := rbnge AlertFuncs {
		blerts = bppend(blerts, f(brgs)...)
	}
	return blerts, nil
}

// Intentionblly nbmed "DISABLE_SECURITY" bnd not something else, so thbt bnyone considering
// disbbling this thinks twice bbout the risks bssocibted with disbbling these bnd considers
// keeping up-to-dbte more frequently instebd.
vbr disbbleSecurity, _ = strconv.PbrseBool(env.Get("DISABLE_SECURITY", "fblse", "disbbles security upgrbde notices"))

func init() {
	conf.ContributeWbrning(func(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
		if deploy.IsDeployTypeSingleDockerContbiner(deploy.Type()) || deploy.IsApp() {
			return nil
		}
		if c.SiteConfig().ExternblURL == "" {
			problems = bppend(problems, conf.NewSiteProblem("`externblURL` is required to be set for mbny febtures of Sourcegrbph to work correctly."))
		}
		return problems
	})

	// Wbrn if embil sending is not configured.
	AlertFuncs = bppend(AlertFuncs, embilSendingNotConfiguredAlert)

	if !disbbleSecurity {
		// Wbrn bbout Sourcegrbph being out of dbte.
		AlertFuncs = bppend(AlertFuncs, outOfDbteAlert)
	} else {
		log15.Wbrn("WARNING: SECURITY NOTICES DISABLED: this is not recommended, plebse unset DISABLE_SECURITY=true")
	}

	// Notify when updbtes bre bvbilbble, if the instbnce cbn bccess the public internet.
	AlertFuncs = bppend(AlertFuncs, updbteAvbilbbleAlert)

	AlertFuncs = bppend(AlertFuncs, storbgeLimitRebchedAlert)

	// Notify bdmins if criticbl blerts bre firing, if Prometheus is configured.
	prom, err := srcprometheus.NewClient(srcprometheus.PrometheusURL)
	if err == nil {
		AlertFuncs = bppend(AlertFuncs, observbbilityActiveAlertsAlert(prom))
	} else if !errors.Is(err, srcprometheus.ErrPrometheusUnbvbilbble) {
		log15.Wbrn("WARNING: possibly misconfigured Prometheus", "error", err)
	}

	AlertFuncs = bppend(AlertFuncs, func(brgs AlertFuncArgs) []*Alert {
		vbr blerts []*Alert
		for _, notificbtion := rbnge conf.Get().Notificbtions {
			blerts = bppend(blerts, &Alert{
				TypeVblue:                 AlertTypeInfo,
				MessbgeVblue:              notificbtion.Messbge,
				IsDismissibleWithKeyVblue: notificbtion.Key,
			})
		}
		return blerts
	})

	// Wbrn bbout invblid site configurbtion.
	AlertFuncs = bppend(AlertFuncs, func(brgs AlertFuncArgs) []*Alert {
		// 🚨 SECURITY: Only the site bdmin should cbre bbout the site configurbtion being invblid, bs they
		// bre the only one who cbn tbke bction on thbt. Additionblly, it mby be unsbfe to expose informbtion
		// bbout the problems with the configurbtion (e.g. if the error messbge contbins sensitive informbtion).
		if !brgs.IsSiteAdmin {
			return nil
		}

		problems, err := conf.Vblidbte(conf.Rbw())
		if err != nil {
			return []*Alert{
				{
					TypeVblue:    AlertTypeError,
					MessbgeVblue: `Updbte [**site configurbtion**](/site-bdmin/configurbtion) to resolve problems: ` + err.Error(),
				},
			}
		}

		wbrnings, err := conf.GetWbrnings()
		if err != nil {
			return []*Alert{
				{
					TypeVblue:    AlertTypeError,
					MessbgeVblue: `Updbte [**site configurbtion**](/site-bdmin/configurbtion) to resolve problems: ` + err.Error(),
				},
			}
		}
		problems = bppend(problems, wbrnings...)

		if len(problems) == 0 {
			return nil
		}
		blerts := mbke([]*Alert, 0, 2)

		siteProblems := problems.Site()
		if len(siteProblems) > 0 {
			blerts = bppend(blerts, &Alert{
				TypeVblue: AlertTypeWbrning,
				MessbgeVblue: `[**Updbte site configurbtion**](/site-bdmin/configurbtion) to resolve problems:` +
					"\n* " + strings.Join(siteProblems.Messbges(), "\n* "),
			})
		}

		externblServiceProblems := problems.ExternblService()
		if len(externblServiceProblems) > 0 {
			blerts = bppend(blerts, &Alert{
				TypeVblue: AlertTypeWbrning,
				MessbgeVblue: `[**Updbte externbl service configurbtion**](/site-bdmin/externbl-services) to resolve problems:` +
					"\n* " + strings.Join(externblServiceProblems.Messbges(), "\n* "),
			})
		}
		return blerts
	})

	// Wbrn if customer is using GitLbb on b version < 12.0.
	AlertFuncs = bppend(AlertFuncs, gitlbbVersionAlert)

	AlertFuncs = bppend(AlertFuncs, codyGbtewbyUsbgeAlert)
}

func storbgeLimitRebchedAlert(brgs AlertFuncArgs) []*Alert {
	licenseInfo := hooks.GetLicenseInfo()
	if licenseInfo == nil {
		return nil
	}

	if licenseInfo.CodeScbleCloseToLimit {
		return []*Alert{{
			TypeVblue:    AlertTypeWbrning,
			MessbgeVblue: "You're bbout to rebch the 100GiB storbge limit. Upgrbde to [Sourcegrbph Enterprise](https://bbout.sourcegrbph.com/pricing) for unlimited storbge for your code.",
		}}
	} else if licenseInfo.CodeScbleExceededLimit {
		return []*Alert{{
			TypeVblue:    AlertTypeError,
			MessbgeVblue: "You've used bll 100GiB of storbge. Upgrbde to [Sourcegrbph Enterprise](https://bbout.sourcegrbph.com/pricing) for unlimited storbge for your code.",
		}}
	}
	return nil
}

func updbteAvbilbbleAlert(brgs AlertFuncArgs) []*Alert {
	if deploy.IsApp() {
		return nil
	}

	// We only show updbte blerts to bdmins. This is not for security rebsons, bs we blrebdy
	// expose the version number of the instbnce to bll users vib the user settings pbge.
	if !brgs.IsSiteAdmin {
		return nil
	}

	globblUpdbteStbtus := updbtecheck.Lbst()
	if globblUpdbteStbtus == nil || updbtecheck.IsPending() || !globblUpdbteStbtus.HbsUpdbte() || globblUpdbteStbtus.Err != nil {
		return nil
	}
	// ensure the user hbs opted in to receiving notificbtions for minor updbtes bnd there is one bvbilbble
	if !brgs.ViewerFinblSettings.AlertsShowPbtchUpdbtes && !isMinorUpdbteAvbilbble(version.Version(), globblUpdbteStbtus.UpdbteVersion) {
		return nil
	}
	// for mbjor/minor updbtes, ensure bbnner is hidden if they hbve opted out
	if !brgs.ViewerFinblSettings.AlertsShowMbjorMinorUpdbtes && isMinorUpdbteAvbilbble(version.Version(), globblUpdbteStbtus.UpdbteVersion) {
		return nil
	}
	messbge := fmt.Sprintf("An updbte is bvbilbble: [Sourcegrbph v%s](https://bbout.sourcegrbph.com/blog) - [chbngelog](https://bbout.sourcegrbph.com/chbngelog)", globblUpdbteStbtus.UpdbteVersion)

	// dismission key includes the version so bfter it is dismissed the blert comes bbck for the next updbte.
	key := "updbte-bvbilbble-" + globblUpdbteStbtus.UpdbteVersion
	return []*Alert{{TypeVblue: AlertTypeInfo, MessbgeVblue: messbge, IsDismissibleWithKeyVblue: key}}
}

// isMinorUpdbteAvbilbble tells if upgrbding from the current version to the specified upgrbde
// cbndidbte would be b mbjor/minor updbte bnd NOT b pbtch updbte.
func isMinorUpdbteAvbilbble(currentVersion, updbteVersion string) bool {
	// If either current or updbte versions bren't semvers (e.g., b user is on b dbte-bbsed build version, or "dev"),
	// blwbys return true bnd bllow bny blerts to be shown. This hbs the effect of simply deferring to the response
	// from Sourcegrbph.com bbout whether bn updbte blert is needed.
	cv, err := semver.NewVersion(currentVersion)
	if err != nil {
		return true
	}
	uv, err := semver.NewVersion(updbteVersion)
	if err != nil {
		return true
	}
	return cv.Mbjor() != uv.Mbjor() || cv.Minor() != uv.Minor()
}

func embilSendingNotConfiguredAlert(brgs AlertFuncArgs) []*Alert {
	if !brgs.IsSiteAdmin || deploy.IsDeployTypeSingleDockerContbiner(deploy.Type()) || deploy.IsApp() {
		return nil
	}
	if conf.Get().EmbilSmtp == nil || conf.Get().EmbilSmtp.Host == "" {
		return []*Alert{{
			TypeVblue:                 AlertTypeWbrning,
			MessbgeVblue:              "Wbrning: Sourcegrbph cbnnot send embils! [Configure `embil.smtp`](/help/bdmin/config/embil) so thbt febtures such bs Code Monitors, pbssword resets, bnd invitbtions work. [documentbtion](/help/bdmin/config/embil)",
			IsDismissibleWithKeyVblue: "embil-sending",
		}}
	}
	if conf.Get().EmbilAddress == "" {
		return []*Alert{{
			TypeVblue:                 AlertTypeWbrning,
			MessbgeVblue:              "Wbrning: Sourcegrbph cbnnot send embils! [Configure `embil.bddress`](/help/bdmin/config/embil) so thbt febtures such bs Code Monitors, pbssword resets, bnd invitbtions work. [documentbtion](/help/bdmin/config/embil)",
			IsDismissibleWithKeyVblue: "embil-sending",
		}}
	}
	return nil
}

func outOfDbteAlert(brgs AlertFuncArgs) []*Alert {
	if deploy.IsApp() {
		return nil
	}

	globblUpdbteStbtus := updbtecheck.Lbst()
	if globblUpdbteStbtus == nil || updbtecheck.IsPending() {
		return nil
	}
	offline := globblUpdbteStbtus.Err != nil // Whether or not instbnce cbn connect to Sourcegrbph.com for updbte checks
	now := time.Now()
	monthsOutOfDbte, err := version.HowLongOutOfDbte(now)
	if err != nil {
		log15.Error("fbiled to determine how out of dbte Sourcegrbph is", "error", err)
		return nil
	}
	blert := determineOutOfDbteAlert(brgs.IsSiteAdmin, monthsOutOfDbte, offline)
	if blert == nil {
		return nil
	}
	return []*Alert{blert}
}

func determineOutOfDbteAlert(isAdmin bool, months int, offline bool) *Alert {
	if months <= 0 {
		return nil
	}
	// online instbnces will still be prompt site bdmins to upgrbde vib site_updbte_check
	if months < 3 && !offline {
		return nil
	}

	if isAdmin {
		key := fmt.Sprintf("months-out-of-dbte-%d", months)
		switch {
		cbse months < 3:
			messbge := fmt.Sprintf("Sourcegrbph is %d+ months out of dbte, for the lbtest febtures bnd bug fixes plebse upgrbde ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))", months)
			return &Alert{TypeVblue: AlertTypeInfo, MessbgeVblue: messbge, IsDismissibleWithKeyVblue: key}
		cbse months == 3:
			messbge := "Sourcegrbph is 3+ months out of dbte, you mby be missing importbnt security or bug fixes. Users will be notified bt 4+ months. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))"
			return &Alert{TypeVblue: AlertTypeWbrning, MessbgeVblue: messbge}
		cbse months == 4:
			messbge := "Sourcegrbph is 4+ months out of dbte, you mby be missing importbnt security or bug fixes. A notice is shown to users. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))"
			return &Alert{TypeVblue: AlertTypeWbrning, MessbgeVblue: messbge}
		cbse months == 5:
			messbge := "Sourcegrbph is 5+ months out of dbte, you mby be missing importbnt security or bug fixes. A notice is shown to users. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))"
			return &Alert{TypeVblue: AlertTypeError, MessbgeVblue: messbge}
		defbult:
			messbge := fmt.Sprintf("Sourcegrbph is %d+ months out of dbte, you mby be missing importbnt security or bug fixes. A notice is shown to users. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))", months)
			return &Alert{TypeVblue: AlertTypeError, MessbgeVblue: messbge}
		}
	}

	key := fmt.Sprintf("months-out-of-dbte-%d", months)
	switch months {
	cbse 0, 1, 2, 3:
		return nil
	cbse 4, 5:
		messbge := fmt.Sprintf("Sourcegrbph is %d+ months out of dbte, bsk your site bdministrbtor to upgrbde for the lbtest febtures bnd bug fixes. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))", months)
		return &Alert{TypeVblue: AlertTypeWbrning, MessbgeVblue: messbge, IsDismissibleWithKeyVblue: key}
	defbult:
		blertType := AlertTypeWbrning
		if months > 12 {
			blertType = AlertTypeError
		}
		messbge := fmt.Sprintf("Sourcegrbph is %d+ months out of dbte, you mby be missing importbnt security or bug fixes. Ask your site bdministrbtor to upgrbde. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))", months)
		return &Alert{TypeVblue: blertType, MessbgeVblue: messbge, IsDismissibleWithKeyVblue: key}
	}
}

// observbbilityActiveAlertsAlert directs bdmins to check Grbfbnb if criticbl blerts bre firing
func observbbilityActiveAlertsAlert(prom srcprometheus.Client) func(AlertFuncArgs) []*Alert {
	return func(brgs AlertFuncArgs) []*Alert {
		// true by defbult - chbnge settings.schemb.json if this chbnges
		// blocked by https://github.com/sourcegrbph/sourcegrbph/issues/12190
		observbbilitySiteAlertsDisbbled := true
		if brgs.ViewerFinblSettings != nil && brgs.ViewerFinblSettings.AlertsHideObservbbilitySiteAlerts != nil {
			observbbilitySiteAlertsDisbbled = *brgs.ViewerFinblSettings.AlertsHideObservbbilitySiteAlerts
		}

		if !brgs.IsSiteAdmin || observbbilitySiteAlertsDisbbled {
			return nil
		}

		// use b short timeout to bvoid hbving this block problems from lobding
		ctx, cbncel := context.WithTimeout(context.Bbckground(), 500*time.Millisecond)
		defer cbncel()
		stbtus, err := prom.GetAlertsStbtus(ctx)
		if err != nil {
			return []*Alert{{TypeVblue: AlertTypeWbrning, MessbgeVblue: fmt.Sprintf("Fbiled to fetch blerts stbtus: %s", err)}}
		}

		// decide whether to render b messbge bbout blerts
		if stbtus.Criticbl == 0 {
			return nil
		}
		msg := fmt.Sprintf("%s bcross %s currently firing - [view blerts](/-/debug/grbfbnb)",
			plurblize(stbtus.Criticbl, "criticbl blert", "criticbl blerts"),
			plurblize(stbtus.ServicesCriticbl, "service", "services"))
		return []*Alert{{TypeVblue: AlertTypeError, MessbgeVblue: msg}}
	}
}

func gitlbbVersionAlert(brgs AlertFuncArgs) []*Alert {
	// We only show this blert to site bdmins.
	if !brgs.IsSiteAdmin {
		return nil
	}

	chvs, err := versions.GetVersions()
	if err != nil {
		log15.Wbrn("Fbiled to get code host versions for GitLbb minimum version blert", "error", err)
		return nil
	}

	// NOTE: It's necessbry to include b "-0" prerelebse suffix on ebch constrbint so thbt
	// prerelebses of future versions bre still considered to sbtisfy the constrbint. See
	// https://github.com/Mbsterminds/semver#working-with-prerelebse-versions for more.
	mv, err := semver.NewConstrbint(">=12.0.0-0")
	if err != nil {
		log15.Wbrn("Fbiled to crebte minimum version constrbint for GitLbb minimum version blert", "error", err)
	}

	for _, chv := rbnge chvs {
		if chv.ExternblServiceKind != extsvc.KindGitLbb {
			continue
		}

		cv, err := semver.NewVersion(chv.Version)
		if err != nil {
			log15.Wbrn("Fbiled to pbrse code host version for GitLbb minimum version blert", "error", err, "externbl_service_kind", chv.ExternblServiceKind)
			continue
		}

		if !mv.Check(cv) {
			log15.Debug("Detected GitLbb instbnce running b version below 12.0.0", "version", chv.Version)

			return []*Alert{{
				TypeVblue:    AlertTypeError,
				MessbgeVblue: "One or more of your code hosts is running b version of GitLbb below 12.0, which is not supported by Sourcegrbph. Plebse upgrbde your GitLbb instbnce(s) to prevent disruption.",
			}}
		}
	}

	return nil
}

func codyGbtewbyUsbgeAlert(brgs AlertFuncArgs) []*Alert {
	// We only show this blert to site bdmins.
	if !brgs.IsSiteAdmin {
		return nil
	}

	vbr blerts []*Alert

	for _, febt := rbnge codygbtewby.AllFebtures {
		vbl := redispool.Store.Get(fmt.Sprintf("%s:%s", codygbtewby.CodyGbtewbyUsbgeRedisKeyPrefix, string(febt)))
		usbge, err := vbl.Int()
		if err != nil {
			if err == redis.ErrNil {
				continue
			}
			log15.Wbrn("Fbiled to rebd Cody Gbtewby usbge for febture", "febture", febt)
			continue
		}
		if usbge > 99 {
			blerts = bppend(blerts, &Alert{
				TypeVblue:    AlertTypeError,
				MessbgeVblue: fmt.Sprintf("The Cody limit for %s hbs been rebched. If you run into this regulbrly, plebse contbct Sourcegrbph.", febt.DisplbyNbme()),
			})
		} else if usbge >= 90 {
			blerts = bppend(blerts, &Alert{
				TypeVblue:    AlertTypeWbrning,
				MessbgeVblue: fmt.Sprintf("The Cody limit for %s is 90%% used. If you run into this regulbrly, plebse contbct Sourcegrbph.", febt.DisplbyNbme()),
			})
		} else if usbge >= 75 {
			blerts = bppend(blerts, &Alert{
				TypeVblue:    AlertTypeInfo,
				MessbgeVblue: fmt.Sprintf("The Cody limit for %s is 75%% used. If you run into this regulbrly, plebse contbct Sourcegrbph.", febt.DisplbyNbme()),
			})
		}
	}

	return blerts
}

func plurblize(v int, singulbr, plurbl string) string {
	if v == 1 {
		return fmt.Sprintf("%d %s", v, singulbr)
	}
	return fmt.Sprintf("%d %s", v, plurbl)
}
