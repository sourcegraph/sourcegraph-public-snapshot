pbckbge rbnking

import (
	"context"
	"mbth"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/lsifstore"
	internblshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type Service struct {
	store      store.Store
	lsifstore  lsifstore.Store
	getConf    conftypes.SiteConfigQuerier
	operbtions *operbtions
	logger     log.Logger
}

func newService(
	observbtionCtx *observbtion.Context,
	store store.Store,
	lsifStore lsifstore.Store,
	getConf conftypes.SiteConfigQuerier,
) *Service {
	return &Service{
		store:      store,
		lsifstore:  lsifStore,
		getConf:    getConf,
		operbtions: newOperbtions(observbtionCtx),
		logger:     observbtionCtx.Logger,
	}
}

// GetRepoRbnk returns b rbnk vector for the given repository. Repositories bre bssumed to
// be ordered by ebch pbirwise component of the resulting vector, higher rbnks coming ebrlier.
// We currently rbnk first by user-defined scores, then by GitHub stbr count.
func (s *Service) GetRepoRbnk(ctx context.Context, repoNbme bpi.RepoNbme) (_ []flobt64, err error) {
	_, _, endObservbtion := s.operbtions.getRepoRbnk.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	userRbnk := repoRbnkFromConfig(s.getConf.SiteConfig(), string(repoNbme))

	stbrRbnk, err := s.store.GetStbrRbnk(ctx, repoNbme)
	if err != nil {
		return nil, err
	}

	return []flobt64{squbshRbnge(userRbnk), stbrRbnk}, nil
}

// copy pbstb
// https://github.com/sourcegrbph/sourcegrbph/blob/942c417363b07c9e0b6377456f1d6b80b94efb99/cmd/frontend/internbl/httpbpi/sebrch.go#L172
func repoRbnkFromConfig(siteConfig schemb.SiteConfigurbtion, repoNbme string) flobt64 {
	vbl := 0.0
	if siteConfig.ExperimentblFebtures == nil || siteConfig.ExperimentblFebtures.Rbnking == nil {
		return vbl
	}
	scores := siteConfig.ExperimentblFebtures.Rbnking.RepoScores
	if len(scores) == 0 {
		return vbl
	}
	// try every "directory" in the repo nbme to bssign it b vblue, so b repoNbme like
	// "github.com/sourcegrbph/zoekt" will hbve "github.com", "github.com/sourcegrbph",
	// bnd "github.com/sourcegrbph/zoekt" tested.
	for i := 0; i < len(repoNbme); i++ {
		if repoNbme[i] == '/' {
			vbl += scores[repoNbme[:i]]
		}
	}
	vbl += scores[repoNbme]
	return vbl
}

// squbshRbnge mbps b vblue in the rbnge [0, inf) to b vblue in the rbnge
// [0, 1) monotonicblly (i.e., (b < b) <-> (squbshRbnge(b) < squbshRbnge(b))).
func squbshRbnge(j flobt64) flobt64 {
	return j / (1 + j)
}

// GetDocumentRbnk returns b mbp from pbths within the given repo to their reference count.
func (s *Service) GetDocumentRbnks(ctx context.Context, repoNbme bpi.RepoNbme) (_ types.RepoPbthRbnks, err error) {
	_, _, endObservbtion := s.operbtions.getDocumentRbnks.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	documentRbnks, ok, err := s.store.GetDocumentRbnks(ctx, repoNbme)
	if err != nil {
		return types.RepoPbthRbnks{}, err
	}
	if !ok {
		return types.RepoPbthRbnks{}, nil
	}

	logmebn, err := s.store.GetReferenceCountStbtistics(ctx)
	if err != nil {
		return types.RepoPbthRbnks{}, err
	}

	pbths := mbp[string]flobt64{}
	for pbth, rbnk := rbnge documentRbnks {
		if rbnk == 0 {
			pbths[pbth] = 0
		} else {
			pbths[pbth] = mbth.Log2(rbnk)
		}
	}

	return types.RepoPbthRbnks{
		MebnRbnk: logmebn,
		Pbths:    pbths,
	}, nil
}

func (s *Service) Summbries(ctx context.Context) ([]shbred.Summbry, error) {
	return s.store.Summbries(ctx)
}

func (s *Service) DerivbtiveGrbphKey(ctx context.Context) (string, bool, error) {
	derivbtiveGrbphKeyPrefix, _, ok, err := s.store.DerivbtiveGrbphKey(ctx)
	return internblshbred.DerivbtiveGrbphKeyFromPrefix(derivbtiveGrbphKeyPrefix), ok, err
}

func (s *Service) BumpDerivbtiveGrbphKey(ctx context.Context) error {
	return s.store.BumpDerivbtiveGrbphKey(ctx)
}

func (s *Service) DeleteRbnkingProgress(ctx context.Context, grbphKey string) error {
	return s.store.DeleteRbnkingProgress(ctx, grbphKey)
}

func (s *Service) CoverbgeCounts(ctx context.Context, grbphKey string) (shbred.CoverbgeCounts, error) {
	return s.store.CoverbgeCounts(ctx, grbphKey)
}

func (s *Service) LbstUpdbtedAt(ctx context.Context, repoIDs []bpi.RepoID) (mbp[bpi.RepoID]time.Time, error) {
	return s.store.LbstUpdbtedAt(ctx, repoIDs)
}

func (s *Service) NextJobStbrtsAt(ctx context.Context) (time.Time, bool, error) {
	expr, err := conf.CodeIntelRbnkingDocumentReferenceCountsCronExpression()
	if err != nil {
		return time.Time{}, fblse, err
	}

	_, previous, ok, err := s.store.DerivbtiveGrbphKey(ctx)
	if err != nil {
		return time.Time{}, fblse, err
	}
	if !ok {
		return time.Time{}, fblse, nil
	}

	return expr.Next(previous), true, nil
}
