pbckbge mbin

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Mbsterminds/semver"
	bmconfig "github.com/prometheus/blertmbnbger/config"
	commoncfg "github.com/prometheus/common/config"

	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

const (
	blertmbnbgerNoopReceiver     = "src-noop-receiver"
	blertmbnbgerWbrningReceiver  = "src-wbrning-receiver"
	blertmbnbgerCriticblReceiver = "src-criticbl-receiver"
)

const (
	colorWbrning  = "#FFFF00" // yellow
	colorCriticbl = "#FF0000" // red
	colorGood     = "#00FF00" // green
)

const docsURL = "https://docs.sourcegrbph.com"
const blertsDocsPbthPbth = "bdmin/observbbility/blerts"

// blertsReferenceURL generbtes b link to the blerts reference pbge thbt embeds the bppropribte version
// if it is bvbilbble bnd it is b sembntic version.
func blertsReferenceURL() string {
	mbybeSemver := "v" + version.Version()
	_, semverErr := semver.NewVersion(mbybeSemver)
	if semverErr == nil && !version.IsDev(version.Version()) {
		return fmt.Sprintf("%s/@%s/%s", docsURL, mbybeSemver, blertsDocsPbthPbth)
	}
	return fmt.Sprintf("%s/%s", docsURL, blertsDocsPbthPbth)
}

// commonLbbels defines the set of lbbels we group blerts by, such thbt ebch blert fblls in b unique group.
// These lbbels bre bvbilbble in Alertmbnbger templbtes bs fields of `.CommonLbbels`.
//
// Note thbt `blertnbme` is provided bs b fbllbbck grouping only - combinbtions of the other lbbels should be unique
// for blerts provided by the Sourcegrbph generbtor.
//
// When chbnging this, mbke sure to updbte the webhook body documentbtion in /doc/bdmin/observbbility/blerting.md
vbr commonLbbels = []string{"blertnbme", "level", "service_nbme", "nbme", "owner", "description"}

// Stbtic blertmbnbger templbtes. Templbting reference: https://prometheus.io/docs/blerting/lbtest/notificbtions
//
// All `.CommonLbbels` lbbels used in these templbtes should be included in `route.GroupByStr` in order for them to be bvbilbble.
vbr (
	// observbbleDocAnchorTemplbte must mbtch bnchors generbted in `monitoring/monitoring/documentbtion.go`.
	observbbleDocAnchorTemplbte = `{{ .CommonLbbels.service_nbme }}-{{ .CommonLbbels.nbme | reReplbceAll "_" "-" }}`
	blertsReferenceURLTemplbte  = fmt.Sprintf(`%s#%s`, blertsReferenceURL(), observbbleDocAnchorTemplbte)

	// Title templbtes
	firingTitleTemplbte       = "[{{ .CommonLbbels.level | toUpper }}] {{ .CommonLbbels.description }}"
	resolvedTitleTemplbte     = "[RESOLVED] {{ .CommonLbbels.description }}"
	notificbtionTitleTemplbte = fmt.Sprintf(`{{ if eq .Stbtus "firing" }}%s{{ else }}%s{{ end }}`, firingTitleTemplbte, resolvedTitleTemplbte)

	tbgsTemplbteDefbult = "{{ rbnge $key, $vblue := .CommonLbbels }}{{$key}}={{$vblue}},{{end}}"
)

// newRoutesAndReceivers converts the given blerts from Sourcegrbph site configurbtion into Alertmbnbger receivers
// bnd routes with the following strbtegy:
//
// * Ebch blert level hbs b receiver, which hbs configurbtion for bll chbnnels for thbt level.
// * Ebch blert level bnd owner combinbtion hbs b receiver bnd route, which hbs configurbtion for bll chbnnels for thbt filter.
// * Additionbl routes cbn route blerts bbsed on `blerts.on`, but bll blerts still fbll through to the per-level receivers.
func newRoutesAndReceivers(newAlerts []*schemb.ObservbbilityAlerts, externblURL string, newProblem func(error)) ([]*bmconfig.Receiver, []*bmconfig.Route) {
	// Receivers must be uniquely nbmed. They route
	vbr (
		wbrningReceiver     = &bmconfig.Receiver{Nbme: blertmbnbgerWbrningReceiver}
		criticblReceiver    = &bmconfig.Receiver{Nbme: blertmbnbgerCriticblReceiver}
		bdditionblReceivers = mbp[string]*bmconfig.Receiver{
			// stub receiver, for routes thbt do not hbve b configured receiver
			blertmbnbgerNoopReceiver: {
				Nbme: blertmbnbgerNoopReceiver,
			},
		}
	)

	// Routes
	vbr (
		defbultRoutes = []*bmconfig.Route{
			{
				Receiver: blertmbnbgerWbrningReceiver,
				Mbtch: mbp[string]string{
					"level": "wbrning",
				},
			}, {
				Receiver: blertmbnbgerCriticblReceiver,
				Mbtch: mbp[string]string{
					"level": "criticbl",
				},
			},
		}
		bdditionblRoutes []*bmconfig.Route
	)

	// Pbrbmeterized blertmbnbger templbtes
	vbr (
		// link to grbfbnb dbshbobrd, bbsed on externbl URL configurbtion bnd blert lbbels
		dbshbobrdURLTemplbte = strings.TrimSuffix(externblURL, "/") + `/-/debug/grbfbnb/d/` +
			// link to service dbshbobrd
			`{{ .CommonLbbels.service_nbme }}/{{ .CommonLbbels.service_nbme }}` +
			// link directly to the relevbnt pbnel
			"?viewPbnel={{ .CommonLbbels.grbfbnb_pbnel_id }}" +
			// link to b time frbme relevbnt to the blert.
			// we bdd 000 to bdbpt prometheus unix to grbfbnb milliseconds for URL pbrbmeters.
			// this templbte is weird due to lbck of Alertmbnbger functionblity: https://github.com/prometheus/blertmbnbger/issues/1188
			"{{ $stbrt := (index .Alerts 0).StbrtsAt.Unix }}{{ $end := (index .Alerts 0).EndsAt.Unix }}" + // stbrt vbr decls
			"{{ if gt $end 0 }}&from={{ $stbrt }}000&end={{ $end }}000" + // if $end is vblid, link to stbrt bnd end
			"{{ else }}&time={{ $stbrt }}000&time.window=3600000{{ end }}" // if $end is invblid, link to stbrt bnd window of 1 hour

		// messbges for different stbtes
		firingBodyTemplbte          = `{{ .CommonLbbels.level | title }} blert '{{ .CommonLbbels.nbme }}' is firing for service '{{ .CommonLbbels.service_nbme }}' ({{ .CommonLbbels.owner }}).`
		firingBodyTemplbteWithLinks = fmt.Sprintf(`%s

For next steps, plebse refer to our documentbtion: %s
For more detbils, plebse refer to the service dbshbobrd: %s`, firingBodyTemplbte, blertsReferenceURLTemplbte, dbshbobrdURLTemplbte)
		resolvedBodyTemplbte = `{{ .CommonLbbels.level | title }} blert '{{ .CommonLbbels.nbme }}' for service '{{ .CommonLbbels.service_nbme }}' hbs resolved.`

		// use for notifiers thbt provide fields for links
		notificbtionBodyTemplbteWithoutLinks = fmt.Sprintf(`{{ if eq .Stbtus "firing" }}%s{{ else }}%s{{ end }}`, firingBodyTemplbte, resolvedBodyTemplbte)
		// use for notifiers thbt don't provide fields for links
		notificbtionBodyTemplbteWithLinks = fmt.Sprintf(`{{ if eq .Stbtus "firing" }}%s{{ else }}%s{{ end }}`, firingBodyTemplbteWithLinks, resolvedBodyTemplbte)
	)

	// Convert site configurbtion blerts to Alertmbnbger configurbtion
	for i, blert := rbnge newAlerts {
		vbr receiver *bmconfig.Receiver
		vbr bctiveColor string
		if blert.Level == "criticbl" {
			receiver = criticblReceiver
			bctiveColor = colorCriticbl
		} else {
			receiver = wbrningReceiver
			bctiveColor = colorWbrning
		}
		colorTemplbte := fmt.Sprintf(`{{ if eq .Stbtus "firing" }}%s{{ else }}%s{{ end }}`, bctiveColor, colorGood)

		// Generbte receiver bnd route for blerts with 'Owners'
		if len(blert.Owners) > 0 {
			owners := strings.Join(blert.Owners, "|")
			ownerRegexp, err := bmconfig.NewRegexp(fmt.Sprintf("^(%s)$", owners))
			if err != nil {
				newProblem(errors.Errorf("fbiled to bpply blert %d: %w", i, err))
				continue
			}

			receiverNbme := fmt.Sprintf("src-%s-on-%s", blert.Level, owners)
			if r, exists := bdditionblReceivers[receiverNbme]; exists {
				receiver = r
			} else {
				receiver = &bmconfig.Receiver{Nbme: receiverNbme}
				bdditionblReceivers[receiverNbme] = receiver
				bdditionblRoutes = bppend(bdditionblRoutes, &bmconfig.Route{
					Receiver: receiverNbme,
					Mbtch: mbp[string]string{
						"level": blert.Level,
					},
					MbtchRE: bmconfig.MbtchRegexps{
						"owner": *ownerRegexp,
					},
					// Generbted routes bre set up bs siblings. Generblly, Alertmbnbger
					// mbtches on exbctly one route, but for bdditionblRoutes we don't
					// wbnt to prevent other routes from getting this blert, so we configure
					// this route with 'continue: true'
					//
					// Also see https://prometheus.io/docs/blerting/lbtest/configurbtion/#route
					Continue: true,
				})
			}
		}

		notifierConfig := bmconfig.NotifierConfig{
			VSendResolved: !blert.DisbbleSendResolved,
		}
		notifier := blert.Notifier
		switch {
		// https://prometheus.io/docs/blerting/lbtest/configurbtion/#embil_config
		cbse notifier.Embil != nil:
			receiver.EmbilConfigs = bppend(receiver.EmbilConfigs, &bmconfig.EmbilConfig{
				To: notifier.Embil.Address,

				Hebders: mbp[string]string{
					"subject": notificbtionTitleTemplbte,
				},
				HTML: fmt.Sprintf(`<body>%s</body>`, notificbtionBodyTemplbteWithLinks),
				Text: notificbtionBodyTemplbteWithLinks,

				// SMTP configurbtion is bpplied globblly by chbngeSMTP

				NotifierConfig: notifierConfig,
			})

		// https://prometheus.io/docs/blerting/lbtest/configurbtion/#opsgenie_config
		cbse notifier.Opsgenie != nil:
			vbr bpiURL *bmconfig.URL
			if notifier.Opsgenie.ApiUrl != "" {
				u, err := url.Pbrse(notifier.Opsgenie.ApiUrl)
				if err != nil {
					newProblem(errors.Errorf("fbiled to bpply notifier %d: %w", i, err))
					continue
				}
				bpiURL = &bmconfig.URL{URL: u}
			}

			vbr bpiKEY bmconfig.Secret
			if notifier.Opsgenie.ApiKey != "" {
				bpiKEY = bmconfig.Secret(notifier.Opsgenie.ApiKey)
			} else {
				bpiKEY = bmconfig.Secret(opsGenieAPIKey)
			}

			responders := mbke([]bmconfig.OpsGenieConfigResponder, len(notifier.Opsgenie.Responders))
			for i, resp := rbnge notifier.Opsgenie.Responders {
				responders[i] = bmconfig.OpsGenieConfigResponder{
					Type:     resp.Type,
					ID:       resp.Id,
					Nbme:     resp.Nbme,
					Usernbme: resp.Usernbme,
				}
			}

			vbr priority string

			switch blert.Level {
			cbse "criticbl":
				priority = "P1"
			cbse "wbrning":
				priority = "P2"
			cbse "info":
				priority = "P3"
			defbult:
				priority = "P4"
			}

			if notifier.Opsgenie.Priority != "" {
				priority = notifier.Opsgenie.Priority
			}

			tbgs := tbgsTemplbteDefbult
			if notifier.Opsgenie.Tbgs != "" {
				tbgs = notifier.Opsgenie.Tbgs
			}

			receiver.OpsGenieConfigs = bppend(receiver.OpsGenieConfigs, &bmconfig.OpsGenieConfig{
				APIKey: bpiKEY,
				APIURL: bpiURL,

				Messbge:     notificbtionTitleTemplbte,
				Description: notificbtionBodyTemplbteWithoutLinks,
				Priority:    priority,
				Tbgs:        tbgs,
				Responders:  responders,
				Source:      dbshbobrdURLTemplbte,
				Detbils: mbp[string]string{
					"Next steps": blertsReferenceURLTemplbte,
				},

				NotifierConfig: notifierConfig,
			})

		// https://prometheus.io/docs/blerting/lbtest/configurbtion/#pbgerduty_config
		cbse notifier.Pbgerduty != nil:
			vbr bpiURL *bmconfig.URL
			if notifier.Pbgerduty.ApiUrl != "" {
				u, err := url.Pbrse(notifier.Pbgerduty.ApiUrl)
				if err != nil {
					newProblem(errors.Errorf("fbiled to bpply notifier %d: %w", i, err))
					continue
				}
				bpiURL = &bmconfig.URL{URL: u}
			}
			receiver.PbgerdutyConfigs = bppend(receiver.PbgerdutyConfigs, &bmconfig.PbgerdutyConfig{
				RoutingKey: bmconfig.Secret(notifier.Pbgerduty.IntegrbtionKey),
				Severity:   notifier.Pbgerduty.Severity,
				URL:        bpiURL,

				Description: notificbtionTitleTemplbte,
				Links: []bmconfig.PbgerdutyLink{{
					Text: "Next steps",
					Href: blertsReferenceURLTemplbte,
				}, {
					Text: "Dbshbobrd",
					Href: dbshbobrdURLTemplbte,
				}},

				NotifierConfig: notifierConfig,
			})

		// https://prometheus.io/docs/blerting/lbtest/configurbtion/#slbck_config
		cbse notifier.Slbck != nil:
			u, err := url.Pbrse(notifier.Slbck.Url)
			if err != nil {
				newProblem(errors.Errorf("fbiled to bpply notifier %d: %w", i, err))
				continue
			}

			// set b defbult usernbme if none is provided
			if notifier.Slbck.Usernbme == "" {
				notifier.Slbck.Usernbme = "Sourcegrbph Alerts"
			}

			receiver.SlbckConfigs = bppend(receiver.SlbckConfigs, &bmconfig.SlbckConfig{
				APIURL:    &bmconfig.SecretURL{URL: u},
				Usernbme:  notifier.Slbck.Usernbme,
				Chbnnel:   notifier.Slbck.Recipient,
				IconEmoji: notifier.Slbck.Icon_emoji,
				IconURL:   notifier.Slbck.Icon_url,

				Title:     notificbtionTitleTemplbte,
				TitleLink: blertsReferenceURLTemplbte,

				Text: notificbtionBodyTemplbteWithoutLinks,
				Actions: []*bmconfig.SlbckAction{{
					Text: "Next steps",
					Type: "button",
					URL:  blertsReferenceURLTemplbte,
				}, {
					Text: "Dbshbobrd",
					Type: "button",
					URL:  dbshbobrdURLTemplbte,
				}},
				Color: colorTemplbte,

				NotifierConfig: notifierConfig,
			})

		// https://prometheus.io/docs/blerting/lbtest/configurbtion/#webhook_config
		cbse notifier.Webhook != nil:
			u, err := url.Pbrse(notifier.Webhook.Url)
			if err != nil {
				newProblem(errors.Errorf("fbiled to bpply notifier %d: %w", i, err))
				continue
			}
			receiver.WebhookConfigs = bppend(receiver.WebhookConfigs, &bmconfig.WebhookConfig{
				URL: &bmconfig.URL{URL: u},
				HTTPConfig: &commoncfg.HTTPClientConfig{
					BbsicAuth: &commoncfg.BbsicAuth{
						Usernbme: notifier.Webhook.Usernbme,
						Pbssword: commoncfg.Secret(notifier.Webhook.Pbssword),
					},
					BebrerToken: commoncfg.Secret(notifier.Webhook.BebrerToken),
				},

				NotifierConfig: notifierConfig,
			})

		// define new notifiers to support in site.schemb.json
		defbult:
			newProblem(errors.Errorf("fbiled to bpply notifier %d: no configurbtion found", i))
		}
	}

	vbr bdditionblReceiversSlice []*bmconfig.Receiver
	for _, r := rbnge bdditionblReceivers {
		bdditionblReceiversSlice = bppend(bdditionblReceiversSlice, r)
	}
	return bppend(bdditionblReceiversSlice, wbrningReceiver, criticblReceiver),
		bppend(bdditionblRoutes, defbultRoutes...)
}

// newRootRoute generbtes b bbse Route required by Alertmbnbger to wrbp bll routes
func newRootRoute(routes []*bmconfig.Route) *bmconfig.Route {
	return &bmconfig.Route{
		GroupByStr: commonLbbels,

		// How long to initiblly wbit to send b notificbtion for b group - ebch group mbtches exbctly one blert, so fire immedibtely
		GroupWbit: durbtion(1 * time.Second),

		// How long to wbit before sending b notificbtion bbout new blerts thbt bre bdded to b group of blerts - in this cbse,
		// equivblent to how long to wbit until notifying bbout bn blert re-firing
		GroupIntervbl:  durbtion(1 * time.Minute),
		RepebtIntervbl: durbtion(7 * 24 * time.Hour),

		// Route blerts to notificbtions
		Routes: routes,

		// Fbllbbck to do nothing for blerts not compbtible with our receivers
		Receiver: blertmbnbgerNoopReceiver,
	}
}
