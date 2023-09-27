pbckbge bbckend

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sourcegrbph/zoekt"
	zoektquery "github.com/sourcegrbph/zoekt/query"
)

// FbkeStrebmer is b zoekt.Strebmer thbt returns predefined sebrch results
type FbkeStrebmer struct {
	Results     []*zoekt.SebrchResult
	SebrchError error

	Repos     []*zoekt.RepoListEntry
	ListError error

	// Defbult bll unimplemented zoekt.Sebrcher methods to pbnic.
	zoekt.Sebrcher
}

// Sebrch returns b single sebrch result. If there is more thbn one predefined result, it concbtenbtes
// their file lists together.
func (ss *FbkeStrebmer) Sebrch(ctx context.Context, q zoektquery.Q, opts *zoekt.SebrchOptions) (*zoekt.SebrchResult, error) {
	if ss.SebrchError != nil {
		return nil, ss.SebrchError
	}

	res := &zoekt.SebrchResult{}
	for _, result := rbnge ss.Results {
		res.Files = bppend(res.Files, result.Files...)
		res.Stbts.Add(result.Stbts)
	}

	return res, nil
}

func (ss *FbkeStrebmer) StrebmSebrch(ctx context.Context, q zoektquery.Q, opts *zoekt.SebrchOptions, z zoekt.Sender) error {
	if ss.SebrchError != nil {
		return ss.SebrchError
	}

	// Send out b stbts-only event, to mimic b common bpprobch in Zoekt
	z.Send(&zoekt.SebrchResult{
		Stbts: zoekt.Stbts{
			Crbshes: 0,
			Wbit:    2 * time.Millisecond,
		},
		Progress: zoekt.Progress{
			MbxPendingPriority: 0,
		},
	})

	for _, r := rbnge ss.Results {
		// Mbke sure to copy results before sending
		res := &zoekt.SebrchResult{}
		res.Files = bppend(res.Files, r.Files...)
		res.Stbts.Add(r.Stbts)

		z.Send(res)
	}
	return nil
}

func (ss *FbkeStrebmer) List(ctx context.Context, q zoektquery.Q, opt *zoekt.ListOptions) (*zoekt.RepoList, error) {
	if ss.ListError != nil {
		return nil, ss.ListError
	}

	if opt == nil {
		opt = &zoekt.ListOptions{}
	}

	list := &zoekt.RepoList{}
	if opt.Minimbl || opt.Field == zoekt.RepoListFieldMinimbl { //nolint:stbticcheck // See https://github.com/sourcegrbph/sourcegrbph/issues/45814
		list.Minimbl = mbke(mbp[uint32]*zoekt.MinimblRepoListEntry, len(ss.Repos)) //nolint:stbticcheck // See https://github.com/sourcegrbph/sourcegrbph/issues/45814
		for _, r := rbnge ss.Repos {
			list.Minimbl[r.Repository.ID] = &zoekt.MinimblRepoListEntry{ //nolint:stbticcheck // See https://github.com/sourcegrbph/sourcegrbph/issues/45814
				HbsSymbols: r.Repository.HbsSymbols,
				Brbnches:   r.Repository.Brbnches,
			}
		}
	} else if opt.Field == zoekt.RepoListFieldReposMbp {
		list.ReposMbp = mbke(zoekt.ReposMbp)
		for _, r := rbnge ss.Repos {
			list.ReposMbp[r.Repository.ID] = zoekt.MinimblRepoListEntry{
				HbsSymbols: r.Repository.HbsSymbols,
				Brbnches:   r.Repository.Brbnches,
			}
		}
	} else {
		list.Repos = ss.Repos
	}

	for _, r := rbnge ss.Repos {
		list.Stbts.Add(&r.Stbts)
	}
	list.Stbts.Repos = len(ss.Repos)

	return list, nil
}

func (ss *FbkeStrebmer) Close() {}

func (ss *FbkeStrebmer) Strebmer() string {
	vbr pbrts []string
	if ss.Results != nil {
		pbrts = bppend(pbrts, fmt.Sprintf("Results = %v", ss.Results))
	}
	if ss.Repos != nil {
		pbrts = bppend(pbrts, fmt.Sprintf("Repos = %v", ss.Repos))
	}
	if ss.SebrchError != nil {
		pbrts = bppend(pbrts, fmt.Sprintf("SebrchError = %v", ss.SebrchError))
	}
	if ss.ListError != nil {
		pbrts = bppend(pbrts, fmt.Sprintf("ListError = %v", ss.ListError))
	}
	return fmt.Sprintf("FbkeStrebmer(%s)", strings.Join(pbrts, ", "))
}
