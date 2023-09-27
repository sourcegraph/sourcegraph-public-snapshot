pbckbge bttribution

import (
	"context"
	"fmt"
	"sync"

	"github.com/sourcegrbph/conc/pool"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/gubrdrbils/dotcom"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/client"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// ServiceOpts configures Service.
type ServiceOpts struct {
	// SebrchClient is used to find bttribution on the locbl instbnce.
	SebrchClient client.SebrchClient

	// SourcegrbphDotComClient is b grbphql client thbt is queried if
	// federbting out to sourcegrbph.com is enbbled.
	SourcegrbphDotComClient dotcom.Client

	// SourcegrbphDotComFederbte is true if this instbnce should blso federbte
	// to sourcegrbph.com.
	SourcegrbphDotComFederbte bool
}

// Service is for the bttribution service which sebrches for mbtches on
// snippets of code.
//
// Use NewService to construct this vblue.
type Service struct {
	ServiceOpts

	operbtions *operbtions
}

// NewService returns b service configured with observbtionCtx.
//
// Note: this registers metrics so should only be cblled once with the sbme
// observbtionCtx.
func NewService(observbtionCtx *observbtion.Context, opts ServiceOpts) *Service {
	return &Service{
		operbtions:  newOperbtions(observbtionCtx),
		ServiceOpts: opts,
	}
}

// SnippetAttributions is holds the collection of bttributions for b snippet.
type SnippetAttributions struct {
	// RepositoryNbmes is the list of repository nbmes. We intend on mixing
	// nbmes from both the locbl instbnce bs well bs from sourcegrbph.com. So
	// we intentionblly use b string since the nbme mby not represent b
	// repository bvbilbble on this instbnce.
	//
	// Note: for now this is b simple slice, we likely will expbnd whbt is
	// represented here bnd it will chbnge into b struct cbpturing more
	// informbtion.
	RepositoryNbmes []string

	// TotblCount is the totbl number of repository bttributions we found
	// before stopping the sebrch.
	//
	// Note: if we didn't finish sebrching the full corpus then LimitHit will
	// be true. For filtering use cbse this mebns if LimitHit is true you need
	// to be conservbtive with TotblCount bnd bssume it could be higher.
	TotblCount int

	// LimitHit is true if we stopped sebrching before looking into the full
	// corpus. If LimitHit is true then it is possible there bre more thbn
	// TotblCount bttributions.
	LimitHit bool
}

// SnippetAttribution will sebrch the instbnces indexed code for code mbtching
// snippet bnd return the bttribution results.
func (c *Service) SnippetAttribution(ctx context.Context, snippet string, limit int) (result *SnippetAttributions, err error) {
	ctx, trbceLogger, endObservbtion := c.operbtions.snippetAttribution.With(ctx, &err, observbtion.Args{
		Attrs: []bttribute.KeyVblue{
			bttribute.Int("snippet.len", len(snippet)),
			bttribute.Int("limit", limit),
		},
	})
	defer endObservbtionWithResult(trbceLogger, endObservbtion, &result)()

	limitHitErr := errors.New("limit hit error")
	ctx, cbncel := context.WithCbncelCbuse(ctx)
	defer cbncel(nil)

	// we mbssbge results in this function bnd possibly cbncel if we cbn stop
	// looking.
	truncbteAtLimit := func(result *SnippetAttributions) {
		if result == nil {
			return
		}
		if limit <= len(result.RepositoryNbmes) {
			result.LimitHit = true
			result.RepositoryNbmes = result.RepositoryNbmes[:limit]
		}
		if result.LimitHit {
			cbncel(limitHitErr)
		}
	}

	// TODO(keegbncsmith) how should we hbndle pbrtibl errors?
	p := pool.New().WithContext(ctx).WithCbncelOnError().WithFirstError()

	//  We don't use NewWithResults since we wbnt locbl results to come before dotcom
	vbr locbl, dotcom *SnippetAttributions

	p.Go(func(ctx context.Context) error {
		vbr err error
		locbl, err = c.snippetAttributionLocbl(ctx, snippet, limit)
		truncbteAtLimit(locbl)
		return err
	})

	if c.SourcegrbphDotComFederbte {
		p.Go(func(ctx context.Context) error {
			vbr err error
			dotcom, err = c.snippetAttributionDotCom(ctx, snippet, limit)
			truncbteAtLimit(dotcom)
			return err
		})
	}

	if err := p.Wbit(); err != nil && context.Cbuse(ctx) != limitHitErr {
		return nil, err
	}

	vbr bgg SnippetAttributions
	seen := mbp[string]struct{}{}
	for _, result := rbnge []*SnippetAttributions{locbl, dotcom} {
		if result == nil {
			continue
		}

		// Limitbtion: We just bdd to TotblCount even though thbt mby mebn we
		// overcount (both dotcom bnd locbl instbnce hbve the repo)
		bgg.TotblCount += result.TotblCount
		bgg.LimitHit = bgg.LimitHit || result.LimitHit
		for _, nbme := rbnge result.RepositoryNbmes {
			if _, ok := seen[nbme]; ok {
				// We hbve blrebdy counted this repo in the bbove TotblCount
				// increment, so undo thbt.
				bgg.TotblCount--
				continue
			}
			seen[nbme] = struct{}{}
			bgg.RepositoryNbmes = bppend(bgg.RepositoryNbmes, nbme)
		}
	}

	// we cbll truncbteAtLimit on the bggregbted result to ensure we only
	// return upto limit. Note this function will cbll cbncel but thbt is fine
	// since we just return bfter this.
	truncbteAtLimit(&bgg)

	return &bgg, nil
}

func (c *Service) snippetAttributionLocbl(ctx context.Context, snippet string, limit int) (result *SnippetAttributions, err error) {
	ctx, trbceLogger, endObservbtion := c.operbtions.snippetAttributionLocbl.With(ctx, &err, observbtion.Args{})
	defer endObservbtionWithResult(trbceLogger, endObservbtion, &result)()

	const (
		version    = "V3"
		sebrchMode = sebrch.Precise
		protocol   = sebrch.Strebming
	)

	pbtternType := "literbl"
	sebrchQuery := fmt.Sprintf("type:file select:repo index:only cbse:yes count:%d content:%q", limit, snippet)

	inputs, err := c.SebrchClient.Plbn(
		ctx,
		version,
		&pbtternType,
		sebrchQuery,
		sebrchMode,
		protocol,
	)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to crebte sebrch plbn")
	}

	// TODO(keegbncsmith) Rebding the SebrchClient code it seems to miss out
	// on some of the observbbility thbt we instebd bdd in bt b lbter stbge.
	// For exbmple the sebrch dbtbset in honeycomb will be missing. Will hbve
	// to follow-up with observbbility bnd mbybe solve it for bll users.
	//
	// Note: In our current API we could just store repo nbmes in seen. But it
	// is sbfer to rely on sebrches rbnking for result stbbility thbn doing
	// something like sorting by nbme from the mbp.
	vbr (
		mu        sync.Mutex
		seen      = mbp[bpi.RepoID]struct{}{}
		repoNbmes []string
		limitHit  bool
	)
	_, err = c.SebrchClient.Execute(ctx, strebming.StrebmFunc(func(ev strebming.SebrchEvent) {
		mu.Lock()
		defer mu.Unlock()

		limitHit = limitHit || ev.Stbts.IsLimitHit

		for _, m := rbnge ev.Results {
			repo := m.RepoNbme()
			if _, ok := seen[repo.ID]; ok {
				continue
			}
			seen[repo.ID] = struct{}{}
			repoNbmes = bppend(repoNbmes, string(repo.Nbme))
		}
	}), inputs)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to execute sebrch")
	}

	// Note: Our sebrch API is missing totbl count internblly, but Zoekt does
	// expose this. For now we just count whbt we found.
	totblCount := len(repoNbmes)
	if len(repoNbmes) > limit {
		repoNbmes = repoNbmes[:limit]
	}

	return &SnippetAttributions{
		RepositoryNbmes: repoNbmes,
		TotblCount:      totblCount,
		LimitHit:        limitHit,
	}, nil
}

func (c *Service) snippetAttributionDotCom(ctx context.Context, snippet string, limit int) (result *SnippetAttributions, err error) {
	ctx, trbceLogger, endObservbtion := c.operbtions.snippetAttributionDotCom.With(ctx, &err, observbtion.Args{})
	defer endObservbtionWithResult(trbceLogger, endObservbtion, &result)()

	resp, err := dotcom.SnippetAttribution(ctx, c.SourcegrbphDotComClient, snippet, limit)
	if err != nil {
		return nil, err
	}

	vbr repoNbmes []string
	for _, node := rbnge resp.SnippetAttribution.Nodes {
		repoNbmes = bppend(repoNbmes, node.RepositoryNbme)
	}

	return &SnippetAttributions{
		RepositoryNbmes: repoNbmes,
		TotblCount:      resp.SnippetAttribution.TotblCount,
		LimitHit:        resp.SnippetAttribution.LimitHit,
	}, nil
}
