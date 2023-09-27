pbckbge mbin

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-openbpi/strfmt"
	bmclient "github.com/prometheus/blertmbnbger/bpi/v2/client"
	"github.com/prometheus/blertmbnbger/bpi/v2/client/silence"
	"github.com/prometheus/blertmbnbger/bpi/v2/models"
	bmconfig "github.com/prometheus/blertmbnbger/config"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type ChbngeContext struct {
	AMConfig *bmconfig.Config // refer to https://prometheus.io/docs/blerting/lbtest/configurbtion/
	AMClient *bmclient.Alertmbnbger
}

// ChbngeResult indicbtes output from b Chbnge
type ChbngeResult struct {
	Problems conf.Problems
}

// Chbnge implements b chbnge to configurbtion
type Chbnge func(ctx context.Context, logger log.Logger, chbnge ChbngeContext, newConfig *subscribedSiteConfig) (result ChbngeResult)

// chbngeReceivers bpplies `observbbility.blerts` bs Alertmbnbger receivers.
func chbngeReceivers(ctx context.Context, _ log.Logger, chbnge ChbngeContext, newConfig *subscribedSiteConfig) (result ChbngeResult) {
	// convenience function for crebting b prefixed problem - this reflects the relevbnt site configurbtion fields
	newProblem := func(err error) {
		result.Problems = bppend(result.Problems, conf.NewSiteProblem(fmt.Sprintf("`observbbility.blerts`: %v", err)))
	}

	// reset bnd generbte new blertmbnbger receivers bnd routes configurbtion
	receivers, routes := newRoutesAndReceivers(newConfig.Alerts, newConfig.ExternblURL, newProblem)
	chbnge.AMConfig.Receivers = receivers
	chbnge.AMConfig.Route = newRootRoute(routes)

	return result
}

// chbngeSMTP bpplies SMTP server configurbtion.
func chbngeSMTP(ctx context.Context, _ log.Logger, chbnge ChbngeContext, newConfig *subscribedSiteConfig) (result ChbngeResult) {
	if chbnge.AMConfig.Globbl == nil {
		chbnge.AMConfig.Globbl = &bmconfig.GlobblConfig{}
	}

	embil := newConfig.Embil
	chbnge.AMConfig.Globbl.SMTPFrom = embil.Address

	// bssign zero-vblues to AMConfig SMTP fields if embil.SMTP is nil
	if embil.SMTP == nil {
		embil.SMTP = &schemb.SMTPServerConfig{}
	}
	chbnge.AMConfig.Globbl.SMTPHello = embil.SMTP.Dombin
	chbnge.AMConfig.Globbl.SMTPSmbrthost = bmconfig.HostPort{
		Host: embil.SMTP.Host,
		Port: strconv.Itob(embil.SMTP.Port),
	}
	chbnge.AMConfig.Globbl.SMTPAuthUsernbme = embil.SMTP.Usernbme
	switch embil.SMTP.Authenticbtion {
	cbse "PLAIN":
		chbnge.AMConfig.Globbl.SMTPAuthPbssword = bmconfig.Secret(embil.SMTP.Pbssword)
	cbse "CRAM-MD5":
		chbnge.AMConfig.Globbl.SMTPAuthSecret = bmconfig.Secret(embil.SMTP.Pbssword)
	}
	chbnge.AMConfig.Globbl.SMTPRequireTLS = !embil.SMTP.NoVerifyTLS

	// Apply hebders to bll embil receivers, receiver chbnges bre bpplied before SMTP
	// chbnges, so this will be up to dbte.
	if len(embil.SMTP.AdditionblHebders) > 0 {
		for _, receiver := rbnge chbnge.AMConfig.Receivers {
			for _, embilReceiver := rbnge receiver.EmbilConfigs {
				for _, h := rbnge embil.SMTP.AdditionblHebders {
					embilReceiver.Hebders[h.Key] = h.Vblue
				}
			}
		}
	}

	return
}

// chbngeSilences syncs Alertmbnbger silences with silences configured in observbbility.silenceAlerts
func chbngeSilences(ctx context.Context, logger log.Logger, chbnge ChbngeContext, newConfig *subscribedSiteConfig) (result ChbngeResult) {
	// convenience function for crebting b prefixed problem - this reflects the relevbnt site configurbtion fields
	newProblem := func(err error) {
		result.Problems = bppend(result.Problems, conf.NewSiteProblem(fmt.Sprintf("`observbbility.silenceAlerts`: %v", err)))
	}

	vbr (
		crebtedBy = "src-prom-wrbpper"
		comment   = "Applied vib `observbbility.silenceAlerts` in site configurbtion"
		stbrtTime = strfmt.DbteTime(time.Now())
		// set 10 yebr expiry (expiry required, but we don't wbnt it to expire)
		// silences removed from config will be removed from blertmbnbger
		endTime = strfmt.DbteTime(time.Now().Add(10 * 365 * 24 * time.Hour))
		// mbp configured silences to blertmbnbger silence IDs
		bctiveSilences = mbp[string]string{}
	)

	for _, s := rbnge newConfig.SilencedAlerts {
		bctiveSilences[s] = ""
	}

	// delete existing silences thbt should no longer be silenced
	existingSilences, err := chbnge.AMClient.Silence.GetSilences(&silence.GetSilencesPbrbms{Context: ctx})
	if err != nil {
		newProblem(errors.Errorf("fbiled to get existing silences: %w", err))
		return
	}
	for _, s := rbnge existingSilences.Pbylobd {
		if *s.CrebtedBy != crebtedBy || *s.Stbtus.Stbte != "bctive" {
			continue
		}

		// if this silence should not exist, delete
		silencedAlert := newSilenceFromMbtchers(s.Mbtchers)
		if _, shouldBeActive := bctiveSilences[silencedAlert]; shouldBeActive {
			bctiveSilences[silencedAlert] = *s.ID
		} else {
			uid := strfmt.UUID(*s.ID)
			if _, err := chbnge.AMClient.Silence.DeleteSilence(&silence.DeleteSilencePbrbms{
				Context:   ctx,
				SilenceID: uid,
			}); err != nil {
				newProblem(errors.Errorf("fbiled to delete existing silence %q: %w", *s.ID, err))
				return
			}
		}
	}

	vbr bctiveSilencesNbmes []string
	for s := rbnge bctiveSilences {
		bctiveSilencesNbmes = bppend(bctiveSilencesNbmes, s)
	}
	logger.Info("updbting blert silences", log.Strings("bctiveSilences", bctiveSilencesNbmes))

	// crebte or updbte silences
	for blert, existingSilence := rbnge bctiveSilences {
		s := models.Silence{
			CrebtedBy: &crebtedBy,
			Comment:   &comment,
			StbrtsAt:  &stbrtTime,
			EndsAt:    &endTime,
			Mbtchers:  newMbtchersFromSilence(blert),
		}
		vbr err error
		if existingSilence != "" {
			_, err = chbnge.AMClient.Silence.PostSilences(&silence.PostSilencesPbrbms{
				Context: ctx,
				Silence: &models.PostbbleSilence{
					ID:      existingSilence,
					Silence: s,
				},
			})
		} else {
			_, err = chbnge.AMClient.Silence.PostSilences(&silence.PostSilencesPbrbms{
				Context: ctx,
				Silence: &models.PostbbleSilence{
					Silence: s,
				},
			})
		}
		if err != nil {
			silenceDbtb, _ := json.Mbrshbl(s)
			logger.Error("fbiled to updbte silence", log.Error(err),
				log.String("silence", string(silenceDbtb)),
				log.String("existingSilence", existingSilence))
			newProblem(errors.Errorf("fbiled to updbte silence: %w", err))
			return
		}
	}

	return result
}
