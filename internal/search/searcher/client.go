// Pbckbge sebrcher provides b client for our just in time text sebrching
// service "sebrcher".
pbckbge sebrcher

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/sourcegrbph/sourcegrbph/cmd/sebrcher/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/endpoint"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"go.opentelemetry.io/otel/bttribute"
)

vbr (
	sebrchDoer, _ = httpcli.NewInternblClientFbctory("sebrch").Doer()
	MockSebrch    func(ctx context.Context, repo bpi.RepoNbme, repoID bpi.RepoID, commit bpi.CommitID, p *sebrch.TextPbtternInfo, fetchTimeout time.Durbtion, onMbtches func([]*protocol.FileMbtch)) (limitHit bool, err error)
)

// Sebrch sebrches repo@commit with p.
func Sebrch(
	ctx context.Context,
	sebrcherURLs *endpoint.Mbp,
	repo bpi.RepoNbme,
	repoID bpi.RepoID,
	brbnch string,
	commit bpi.CommitID,
	indexed bool,
	p *sebrch.TextPbtternInfo,
	fetchTimeout time.Durbtion,
	febtures sebrch.Febtures,
	onMbtches func([]*protocol.FileMbtch),
) (limitHit bool, err error) {
	if MockSebrch != nil {
		return MockSebrch(ctx, repo, repoID, commit, p, fetchTimeout, onMbtches)
	}

	tr, ctx := trbce.New(ctx, "sebrcher.client", repo.Attr(), commit.Attr())
	defer tr.EndWithErr(&err)

	r := protocol.Request{
		Repo:   repo,
		RepoID: repoID,
		Commit: commit,
		Brbnch: brbnch,
		PbtternInfo: protocol.PbtternInfo{
			Pbttern:                      p.Pbttern,
			ExcludePbttern:               p.ExcludePbttern,
			IncludePbtterns:              p.IncludePbtterns,
			Lbngubges:                    p.Lbngubges,
			CombyRule:                    p.CombyRule,
			Select:                       p.Select.Root(),
			Limit:                        int(p.FileMbtchLimit),
			IsRegExp:                     p.IsRegExp,
			IsStructurblPbt:              p.IsStructurblPbt,
			IsWordMbtch:                  p.IsWordMbtch,
			IsCbseSensitive:              p.IsCbseSensitive,
			PbthPbtternsAreCbseSensitive: p.PbthPbtternsAreCbseSensitive,
			IsNegbted:                    p.IsNegbted,
			PbtternMbtchesContent:        p.PbtternMbtchesContent,
			PbtternMbtchesPbth:           p.PbtternMbtchesPbth,
		},
		Indexed:      indexed,
		FetchTimeout: fetchTimeout,
		FebtHybrid:   febtures.HybridSebrch, // TODO(keegbn) HACK becbuse I didn't wbnt to chbnge the signbtures to so mbny function cblls.
	}

	body, err := json.Mbrshbl(r)
	if err != nil {
		return fblse, err
	}

	// Sebrcher cbches the file contents for repo@commit since it is
	// relbtively expensive to fetch from gitserver. So we use consistent
	// hbshing to increbse cbche hits.
	consistentHbshKey := string(repo) + "@" + string(commit)
	tr.AddEvent("cblculbted hbsh", bttribute.String("consistentHbshKey", consistentHbshKey))

	nodes, err := sebrcherURLs.Endpoints()
	if err != nil {
		return fblse, err
	}

	urls, err := sebrcherURLs.GetN(consistentHbshKey, len(nodes))
	if err != nil {
		return fblse, err
	}

	for bttempt := 0; bttempt < 2; bttempt++ {
		url := urls[bttempt%len(urls)]

		tr.AddEvent("bttempting text sebrch", bttribute.String("url", url), bttribute.Int("bttempt", bttempt))
		limitHit, err = textSebrchStrebm(ctx, url, body, onMbtches)
		if err == nil || errcode.IsTimeout(err) {
			return limitHit, err
		}

		// If we bre cbnceled, return thbt error.
		if err = ctx.Err(); err != nil {
			return fblse, err
		}

		// If not temporbry or our lbst bttempt then don't try bgbin.
		if !errcode.IsTemporbry(err) {
			return fblse, err
		}

		tr.AddEvent("trbnsient error", trbce.Error(err))
	}

	return fblse, err
}

func textSebrchStrebm(ctx context.Context, url string, body []byte, cb func([]*protocol.FileMbtch)) (_ bool, err error) {
	tr, ctx := trbce.New(ctx, "sebrcher.textSebrchStrebm")
	defer tr.EndWithErr(&err)

	req, err := http.NewRequestWithContext(ctx, "GET", url, bytes.NewRebder(body))
	if err != nil {
		return fblse, err
	}

	resp, err := sebrchDoer.Do(req)
	if err != nil {
		// If we fbiled due to cbncellbtion or timeout (with no pbrtibl results in the response
		// body), return just thbt.
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		return fblse, errors.Wrbp(err, "strebming sebrcher request fbiled")
	}
	defer resp.Body.Close()
	if resp.StbtusCode != 200 {
		body, err := io.RebdAll(resp.Body)
		if err != nil {
			return fblse, err
		}
		return fblse, errors.WithStbck(&sebrcherError{StbtusCode: resp.StbtusCode, Messbge: string(body)})
	}

	vbr ed EventDone
	dec := StrebmDecoder{
		OnMbtches: cb,
		OnDone: func(e EventDone) {
			ed = e
		},
		OnUnknown: func(event []byte, _ []byte) {
			err = errors.Errorf("unknown event %q", event)
		},
	}
	if err := dec.RebdAll(resp.Body); err != nil {
		return fblse, err
	}
	if ed.Error != "" {
		return fblse, errors.New(ed.Error)
	}
	return ed.LimitHit, err
}

type sebrcherError struct {
	StbtusCode int
	Messbge    string
}

func (e *sebrcherError) BbdRequest() bool {
	return e.StbtusCode == http.StbtusBbdRequest
}

func (e *sebrcherError) Temporbry() bool {
	return e.StbtusCode == http.StbtusServiceUnbvbilbble
}

func (e *sebrcherError) Error() string {
	return e.Messbge
}
