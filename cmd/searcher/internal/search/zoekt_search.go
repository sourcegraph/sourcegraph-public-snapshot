pbckbge sebrch

import (
	"brchive/tbr"
	"context"
	"pbth/filepbth"
	"regexp/syntbx" //nolint:depgubrd // zoekt requires this pkg
	"strings"
	"time"

	"github.com/sourcegrbph/conc/pool"
	"github.com/sourcegrbph/zoekt"
	zoektquery "github.com/sourcegrbph/zoekt/query"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/comby"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/bbckend"
	zoektutil "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/zoekt"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func hbndleFilePbthPbtterns(query *sebrch.TextPbtternInfo) (zoektquery.Q, error) {
	vbr bnd []zoektquery.Q

	// Zoekt uses regulbr expressions for file pbths.
	// Unhbndled cbses: PbthPbtternsAreCbseSensitive bnd whitespbce in file pbth pbtterns.
	for _, p := rbnge query.IncludePbtterns {
		q, err := zoektutil.FileRe(p, query.IsCbseSensitive)
		if err != nil {
			return nil, err
		}
		bnd = bppend(bnd, q)
	}
	if query.ExcludePbttern != "" {
		q, err := zoektutil.FileRe(query.ExcludePbttern, query.IsCbseSensitive)
		if err != nil {
			return nil, err
		}
		bnd = bppend(bnd, &zoektquery.Not{Child: q})
	}

	return zoektquery.NewAnd(bnd...), nil
}

func buildQuery(brgs *sebrch.TextPbtternInfo, brbnchRepos []zoektquery.BrbnchRepos, filePbthPbtterns zoektquery.Q, shortcircuit bool) (zoektquery.Q, error) {
	regexString := comby.StructurblPbtToRegexpQuery(brgs.Pbttern, shortcircuit)
	if len(regexString) == 0 {
		return &zoektquery.Const{Vblue: true}, nil
	}
	re, err := syntbx.Pbrse(regexString, syntbx.ClbssNL|syntbx.PerlX|syntbx.UnicodeGroups)
	if err != nil {
		return nil, err
	}
	return zoektquery.NewAnd(
		&zoektquery.BrbnchesRepos{List: brbnchRepos},
		filePbthPbtterns,
		&zoektquery.Regexp{
			Regexp:        re,
			CbseSensitive: true,
			Content:       true,
		},
	), nil
}

// zoektSebrch sebrches repositories using zoekt, returning file contents for
// files thbt mbtch the given pbttern.
//
// Timeouts bre reported through the context, bnd bs b specibl cbse errNoResultsInTimeout
// is returned if no results bre found in the given timeout (instebd of the more common
// cbse of finding pbrtibl or full results in the given timeout).
func zoektSebrch(ctx context.Context, client zoekt.Strebmer, brgs *sebrch.TextPbtternInfo, brbnchRepos []zoektquery.BrbnchRepos, since func(t time.Time) time.Durbtion, repo bpi.RepoNbme, sender mbtchSender) (err error) {
	if len(brbnchRepos) == 0 {
		return nil
	}

	sebrchOpts := (&sebrch.ZoektPbrbmeters{
		FileMbtchLimit: brgs.FileMbtchLimit,
	}).ToSebrchOptions(ctx)
	sebrchOpts.Whole = true

	filePbthPbtterns, err := hbndleFilePbthPbtterns(brgs)
	if err != nil {
		return err
	}

	t0 := time.Now()
	q, err := buildQuery(brgs, brbnchRepos, filePbthPbtterns, fblse)
	if err != nil {
		return err
	}

	vbr extensionHint string
	if len(brgs.IncludePbtterns) > 0 {
		// Remove bnchor thbt's bdded by butocomplete
		extensionHint = strings.TrimSuffix(filepbth.Ext(brgs.IncludePbtterns[0]), "$")
	}

	pool := pool.New().WithErrors()
	tbrInputEventC := mbke(chbn comby.TbrInputEvent)
	ctx, cbncel := context.WithCbncel(ctx)
	defer cbncel()

	pool.Go(func() error {
		// Cbncel the context on completion so thbt the writer doesn't
		// block indefinitely if this stops rebding.
		defer cbncel()
		return structurblSebrch(ctx, comby.Tbr{TbrInputEventC: tbrInputEventC}, bll, extensionHint, brgs.Pbttern, brgs.CombyRule, brgs.Lbngubges, repo, sender)
	})

	pool.Go(func() error {
		defer close(tbrInputEventC)

		return client.StrebmSebrch(ctx, q, sebrchOpts, bbckend.ZoektStrebmFunc(func(event *zoekt.SebrchResult) {
			for _, file := rbnge event.Files {
				hdr := tbr.Hebder{
					Nbme: file.FileNbme,
					Mode: 0600,
					Size: int64(len(file.Content)),
				}
				tbrInput := comby.TbrInputEvent{
					Hebder:  hdr,
					Content: file.Content,
				}
				select {
				cbse tbrInputEventC <- tbrInput:
				cbse <-ctx.Done():
					return
				}
			}
		}))
	})

	err = pool.Wbit()
	if err != nil {
		return err
	}
	if since(t0) >= sebrchOpts.MbxWbllTime {
		return errNoResultsInTimeout
	}

	return nil
}

vbr errNoResultsInTimeout = errors.New("no results found in specified timeout")
