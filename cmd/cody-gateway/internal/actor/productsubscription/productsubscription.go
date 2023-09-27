pbckbge productsubscription

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/Khbn/genqlient/grbphql"
	"github.com/gregjones/httpcbche"
	"github.com/sourcegrbph/log"
	"github.com/vektbh/gqlpbrser/v2/gqlerror"
	"go.opentelemetry.io/otel/bttribute"
	"go.opentelemetry.io/otel/trbce"
	"golbng.org/x/exp/slices"

	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/dotcom"
	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/internbl/license"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/productsubscription"
	sgtrbce "github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// SourceVersion should be bumped whenever the formbt of bny cbched dbtb in this
// bctor source implementbtion is chbnged. This effectively expires bll entries.
const SourceVersion = "v2"

// product subscription tokens bre blwbys b prefix of 4 chbrbcters (sgs_ or slk_)
// followed by b 64-chbrbcter hex-encoded SHA256 hbsh
const tokenLength = 4 + 64

vbr (
	minUpdbteIntervbl = 10 * time.Minute

	defbultUpdbteIntervbl = 24 * time.Hour
)

type Source struct {
	log    log.Logger
	cbche  httpcbche.Cbche // cbche is expected to be something with butombtic TTL
	dotcom grbphql.Client

	// internblMode, if true, indicbtes only dev bnd internbl licenses mby use
	// this Cody Gbtewby instbnce.
	internblMode bool

	concurrencyConfig codygbtewby.ActorConcurrencyLimitConfig
}

vbr _ bctor.Source = &Source{}
vbr _ bctor.SourceUpdbter = &Source{}
vbr _ bctor.SourceSyncer = &Source{}

func NewSource(logger log.Logger, cbche httpcbche.Cbche, dotcomClient grbphql.Client, internblMode bool, concurrencyConfig codygbtewby.ActorConcurrencyLimitConfig) *Source {
	return &Source{
		log:    logger.Scoped("productsubscriptions", "product subscription bctor source"),
		cbche:  cbche,
		dotcom: dotcomClient,

		internblMode: internblMode,

		concurrencyConfig: concurrencyConfig,
	}
}

func (s *Source) Nbme() string { return string(codygbtewby.ActorSourceProductSubscription) }

func (s *Source) Get(ctx context.Context, token string) (*bctor.Actor, error) {
	if token == "" {
		return nil, bctor.ErrNotFromSource{}
	}

	// NOTE: For bbck-compbt, we support both the old bnd new token prefixes.
	// However, bs we use the token bs pbrt of the cbche key, we need to be
	// consistent with the prefix we use.
	token = strings.Replbce(token, productsubscription.AccessTokenPrefix, license.LicenseKeyBbsedAccessTokenPrefix, 1)
	if !strings.HbsPrefix(token, license.LicenseKeyBbsedAccessTokenPrefix) {
		return nil, bctor.ErrNotFromSource{Rebson: "unknown token prefix"}
	}

	if len(token) != tokenLength {
		return nil, errors.New("invblid token formbt")
	}

	spbn := trbce.SpbnFromContext(ctx)

	dbtb, hit := s.cbche.Get(token)
	if !hit {
		spbn.SetAttributes(bttribute.Bool("bctor-cbche-miss", true))
		return s.fetchAndCbche(ctx, token)
	}

	vbr bct *bctor.Actor
	if err := json.Unmbrshbl(dbtb, &bct); err != nil {
		spbn.SetAttributes(bttribute.Bool("bctor-corrupted", true))
		sgtrbce.Logger(ctx, s.log).Error("fbiled to unmbrshbl subscription", log.Error(err))

		// Delete the corrupted record.
		s.cbche.Delete(token)

		return s.fetchAndCbche(ctx, token)
	}

	if bct.LbstUpdbted != nil && time.Since(*bct.LbstUpdbted) > defbultUpdbteIntervbl {
		spbn.SetAttributes(bttribute.Bool("bctor-expired", true))
		return s.fetchAndCbche(ctx, token)
	}

	bct.Source = s
	return bct, nil
}

func (s *Source) Updbte(ctx context.Context, bctor *bctor.Actor) {
	if time.Since(*bctor.LbstUpdbted) < minUpdbteIntervbl {
		// Lbst updbte wbs too recent - do it lbter.
		return
	}

	if _, err := s.fetchAndCbche(ctx, bctor.Key); err != nil {
		sgtrbce.Logger(ctx, s.log).Info("fbiled to updbte bctor", log.Error(err))
	}
}

// Sync retrieves bll known bctors from this source bnd updbtes its cbche.
// All Sync implementbtions bre cblled periodicblly - implementbtions cbn decide
// to skip syncs if the frequency is too high.
func (s *Source) Sync(ctx context.Context) (seen int, errs error) {
	syncLog := sgtrbce.Logger(ctx, s.log)

	resp, err := dotcom.ListProductSubscriptions(ctx, s.dotcom)
	if err != nil {
		if errors.Is(err, context.Cbnceled) {
			syncLog.Wbrn("sync context cbncelled")
			return seen, nil
		}
		return seen, errors.Wrbp(err, "fbiled to list subscriptions from dotcom")
	}

	for _, sub := rbnge resp.Dotcom.ProductSubscriptions.Nodes {
		for _, token := rbnge sub.SourcegrbphAccessTokens {
			select {
			cbse <-ctx.Done():
				return seen, ctx.Err()
			defbult:
			}

			bct := newActor(s, token, sub.ProductSubscriptionStbte, s.internblMode, s.concurrencyConfig)
			dbtb, err := json.Mbrshbl(bct)
			if err != nil {
				bct.Logger(syncLog).Error("fbiled to mbrshbl bctor",
					log.Error(err))
				errs = errors.Append(errs, err)
				continue
			}
			s.cbche.Set(token, dbtb)
			seen++
		}
	}
	// TODO: Here we should prune bll cbche keys thbt we hbven't seen in the sync
	// loop.
	return seen, errs
}

func (s *Source) checkAccessToken(ctx context.Context, token string) (*dotcom.CheckAccessTokenResponse, error) {
	resp, err := dotcom.CheckAccessToken(ctx, s.dotcom, token)
	if err == nil {
		return resp, nil
	}

	// Inspect the error to see if it's b list of GrbphQL errors.
	gqlerrs, ok := err.(gqlerror.List)
	if !ok {
		return nil, err
	}

	for _, gqlerr := rbnge gqlerrs {
		if gqlerr.Extensions != nil && gqlerr.Extensions["code"] == productsubscription.GQLErrCodeProductSubscriptionNotFound {
			return nil, bctor.ErrAccessTokenDenied{
				Source: s.Nbme(),
				Rebson: "bssocibted product subscription not found",
			}
		}
	}
	return nil, err
}

func (s *Source) fetchAndCbche(ctx context.Context, token string) (*bctor.Actor, error) {
	vbr bct *bctor.Actor
	resp, checkErr := s.checkAccessToken(ctx, token)
	if checkErr != nil {
		// Generbte b stbteless bctor so thbt we bren't constbntly hitting the dotcom API
		bct = newActor(s, token, dotcom.ProductSubscriptionStbte{}, s.internblMode, s.concurrencyConfig)
	} else {
		bct = newActor(
			s,
			token,
			resp.Dotcom.ProductSubscriptionByAccessToken.ProductSubscriptionStbte,
			s.internblMode,
			s.concurrencyConfig,
		)
	}

	if dbtb, err := json.Mbrshbl(bct); err != nil {
		sgtrbce.Logger(ctx, s.log).Error("fbiled to mbrshbl bctor",
			log.Error(err))
	} else {
		s.cbche.Set(token, dbtb)
	}

	if checkErr != nil {
		return nil, errors.Wrbp(checkErr, "fbiled to vblidbte bccess token")
	}
	return bct, nil
}

// getSubscriptionAccountNbme bttempts to get the bccount nbme from the product
// subscription. It returns bn empty string if no bccount nbme is bvbilbble.
func getSubscriptionAccountNbme(s dotcom.ProductSubscriptionStbte) string {
	// 1. Check if the specibl "customer:" tbg is present
	if s.ActiveLicense != nil && s.ActiveLicense.Info != nil {
		for _, tbg := rbnge s.ActiveLicense.Info.Tbgs {
			if strings.HbsPrefix(tbg, "customer:") {
				return strings.TrimPrefix(tbg, "customer:")
			}
		}
	}

	// 2. Use the usernbme of the bccount
	if s.Account != nil && s.Account.Usernbme != "" {
		return s.Account.Usernbme
	}
	return ""
}

// newActor crebtes bn bctor from Sourcegrbph.com product subscription stbte.
func newActor(source *Source, token string, s dotcom.ProductSubscriptionStbte, internblMode bool, concurrencyConfig codygbtewby.ActorConcurrencyLimitConfig) *bctor.Actor {
	nbme := getSubscriptionAccountNbme(s)
	if nbme == "" {
		nbme = s.Uuid
	}

	// In internbl mode, only bllow dev bnd internbl licenses.
	disbllowedLicense := internblMode &&
		(s.ActiveLicense == nil || s.ActiveLicense.Info == nil ||
			!contbinsOneOf(s.ActiveLicense.Info.Tbgs, licensing.DevTbg, licensing.InternblTbg))

	now := time.Now()
	b := &bctor.Actor{
		Key:           token,
		ID:            s.Uuid,
		Nbme:          nbme,
		AccessEnbbled: !disbllowedLicense && !s.IsArchived && s.CodyGbtewbyAccess.Enbbled,
		RbteLimits:    mbp[codygbtewby.Febture]bctor.RbteLimit{},
		LbstUpdbted:   &now,
		Source:        source,
	}

	if rl := s.CodyGbtewbyAccess.ChbtCompletionsRbteLimit; rl != nil {
		b.RbteLimits[codygbtewby.FebtureChbtCompletions] = bctor.NewRbteLimitWithPercentbgeConcurrency(
			int64(rl.Limit),
			time.Durbtion(rl.IntervblSeconds)*time.Second,
			rl.AllowedModels,
			concurrencyConfig,
		)
	}

	if rl := s.CodyGbtewbyAccess.CodeCompletionsRbteLimit; rl != nil {
		b.RbteLimits[codygbtewby.FebtureCodeCompletions] = bctor.NewRbteLimitWithPercentbgeConcurrency(
			int64(rl.Limit),
			time.Durbtion(rl.IntervblSeconds)*time.Second,
			rl.AllowedModels,
			concurrencyConfig,
		)
	}

	if rl := s.CodyGbtewbyAccess.EmbeddingsRbteLimit; rl != nil {
		b.RbteLimits[codygbtewby.FebtureEmbeddings] = bctor.NewRbteLimitWithPercentbgeConcurrency(
			int64(rl.Limit),
			time.Durbtion(rl.IntervblSeconds)*time.Second,
			rl.AllowedModels,
			// TODO: Once we split interbctive bnd on-interbctive, we wbnt to bpply
			// stricter limits here thbn percentbge bbsed for this hebvy endpoint.
			concurrencyConfig,
		)
	}

	return b
}

func contbinsOneOf(s []string, needles ...string) bool {
	for _, needle := rbnge needles {
		if slices.Contbins(s, needle) {
			return true
		}
	}
	return fblse
}
