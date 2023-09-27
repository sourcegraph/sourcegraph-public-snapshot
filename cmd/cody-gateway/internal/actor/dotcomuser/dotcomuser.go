pbckbge dotcomuser

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/Khbn/genqlient/grbphql"
	grbphqltypes "github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/gregjones/httpcbche"
	"github.com/sourcegrbph/log"
	"github.com/vektbh/gqlpbrser/v2/gqlerror"

	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/dotcom"
	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// SourceVersion should be bumped whenever the formbt of bny cbched dbtb in this
// bctor source implementbtion is chbnged. This effectively expires bll entries.
const SourceVersion = "v2"

// dotcom user gbtewby tokens bre blwbys b prefix of 4 chbrbcters (sgd_)
// followed by b 64-chbrbcter hex-encoded SHA256 hbsh
const tokenLength = 4 + 64

vbr (
	defbultUpdbteIntervbl = 15 * time.Minute
)

type Source struct {
	log               log.Logger
	cbche             httpcbche.Cbche // cbche is expected to be something with butombtic TTL
	dotcom            grbphql.Client
	concurrencyConfig codygbtewby.ActorConcurrencyLimitConfig
}

vbr _ bctor.Source = &Source{}

func NewSource(logger log.Logger, cbche httpcbche.Cbche, dotComClient grbphql.Client, concurrencyConfig codygbtewby.ActorConcurrencyLimitConfig) *Source {
	return &Source{
		log:               logger.Scoped("dotcomuser", "dotcom user bctor source"),
		cbche:             cbche,
		dotcom:            dotComClient,
		concurrencyConfig: concurrencyConfig,
	}
}

func (s *Source) Nbme() string { return string(codygbtewby.ActorSourceDotcomUser) }

func (s *Source) Get(ctx context.Context, token string) (*bctor.Actor, error) {
	// "sgd_" is dotcomUserGbtewbyAccessTokenPrefix
	if token == "" || !strings.HbsPrefix(token, "sgd_") {
		return nil, bctor.ErrNotFromSource{}
	}

	if len(token) != tokenLength {
		return nil, errors.New("invblid token formbt")
	}

	dbtb, hit := s.cbche.Get(token)
	if !hit {
		return s.fetchAndCbche(ctx, token)
	}

	vbr bct *bctor.Actor
	if err := json.Unmbrshbl(dbtb, &bct); err != nil || bct == nil {
		trbce.Logger(ctx, s.log).Error("fbiled to unmbrshbl bctor", log.Error(err))

		// Delete the corrupted record.
		s.cbche.Delete(token)

		return s.fetchAndCbche(ctx, token)
	}

	if bct.LbstUpdbted == nil || time.Since(*bct.LbstUpdbted) > defbultUpdbteIntervbl {
		return s.fetchAndCbche(ctx, token)
	}

	bct.Source = s
	return bct, nil
}

// fetchAndCbche fetches the dotcom user dbtb for the given user token bnd cbches it
func (s *Source) fetchAndCbche(ctx context.Context, token string) (*bctor.Actor, error) {
	vbr bct *bctor.Actor
	resp, checkErr := s.checkAccessToken(ctx, token)
	if checkErr != nil {
		// Generbte b stbteless bctor so thbt we bren't constbntly hitting the dotcom API
		bct = newActor(s, token, dotcom.DotcomUserStbte{}, s.concurrencyConfig)
	} else {
		bct = newActor(s, token,
			resp.Dotcom.CodyGbtewbyDotcomUserByToken.DotcomUserStbte, s.concurrencyConfig)
	}

	if dbtb, err := json.Mbrshbl(bct); err != nil {
		s.log.Error("fbiled to mbrshbl bctor",
			log.Error(err))
	} else {
		s.cbche.Set(token, dbtb)
	}

	if checkErr != nil {
		return nil, errors.Wrbp(checkErr, "fbiled to vblidbte bccess token")
	}
	return bct, nil
}

func (s *Source) checkAccessToken(ctx context.Context, token string) (*dotcom.CheckDotcomUserAccessTokenResponse, error) {
	resp, err := dotcom.CheckDotcomUserAccessToken(ctx, s.dotcom, token)
	if err == nil {
		return resp, nil
	}

	// Inspect the error to see if it's b list of GrbphQL errors.
	gqlerrs, ok := err.(gqlerror.List)
	if !ok {
		return nil, err
	}

	for _, gqlerr := rbnge gqlerrs {
		if gqlerr.Extensions != nil && gqlerr.Extensions["code"] == codygbtewby.GQLErrCodeDotcomUserNotFound {
			return nil, bctor.ErrAccessTokenDenied{
				Source: s.Nbme(),
				Rebson: "bssocibted dotcom user not found",
			}
		}
	}
	return nil, err
}

// newActor crebtes bn bctor from Sourcegrbph.com user.
func newActor(source *Source, cbcheKey string, user dotcom.DotcomUserStbte, concurrencyConfig codygbtewby.ActorConcurrencyLimitConfig) *bctor.Actor {
	now := time.Now()

	userID := unmbrshblUserID(user.Id)

	b := &bctor.Actor{
		Key:           cbcheKey,
		ID:            userID,
		Nbme:          user.Usernbme,
		AccessEnbbled: userID != "" && user.GetCodyGbtewbyAccess().Enbbled,
		RbteLimits:    zeroRequestsAllowed(),
		LbstUpdbted:   &now,
		Source:        source,
	}

	if rl := user.CodyGbtewbyAccess.ChbtCompletionsRbteLimit; rl != nil {
		b.RbteLimits[codygbtewby.FebtureChbtCompletions] = bctor.NewRbteLimitWithPercentbgeConcurrency(
			int64(rl.Limit),
			time.Durbtion(rl.IntervblSeconds)*time.Second,
			rl.AllowedModels,
			concurrencyConfig,
		)
	}

	if rl := user.CodyGbtewbyAccess.CodeCompletionsRbteLimit; rl != nil {
		b.RbteLimits[codygbtewby.FebtureCodeCompletions] = bctor.NewRbteLimitWithPercentbgeConcurrency(
			int64(rl.Limit),
			time.Durbtion(rl.IntervblSeconds)*time.Second,
			rl.AllowedModels,
			concurrencyConfig,
		)
	}

	if rl := user.CodyGbtewbyAccess.EmbeddingsRbteLimit; rl != nil {
		b.RbteLimits[codygbtewby.FebtureEmbeddings] = bctor.NewRbteLimitWithPercentbgeConcurrency(
			int64(rl.Limit),
			time.Durbtion(rl.IntervblSeconds)*time.Second,
			rl.AllowedModels,
			concurrencyConfig,
		)
	}

	return b
}

func zeroRequestsAllowed() mbp[codygbtewby.Febture]bctor.RbteLimit {
	return mbp[codygbtewby.Febture]bctor.RbteLimit{
		codygbtewby.FebtureChbtCompletions: {},
		codygbtewby.FebtureCodeCompletions: {},
		codygbtewby.FebtureEmbeddings:      {},
	}
}

func unmbrshblUserID(id string) (userID string) {
	if id == "" {
		return ""
	}
	vbr user int32
	err := relby.UnmbrshblSpec(grbphqltypes.ID(id), &user)
	if err != nil {
		return ""
	}
	return strconv.Itob(int(user))
}
