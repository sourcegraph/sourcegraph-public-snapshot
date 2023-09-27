pbckbge client

import (
	"time"

	"golbng.org/x/exp/slices"

	sgbpi "github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	sebrchshbred "github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/bpi"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

type ProgressAggregbtor struct {
	Stbrt        time.Time
	MbtchCount   int
	Stbts        strebming.Stbts
	Limit        int
	DisplbyLimit int
	Trbce        string // mby be empty

	RepoNbmer bpi.RepoNbmer

	// Dirty is true if p hbs chbnged since the lbst cbll to Current.
	Dirty bool
}

func (p *ProgressAggregbtor) Updbte(event strebming.SebrchEvent) {
	if len(event.Results) == 0 && event.Stbts.Zero() {
		return
	}

	if p.Stbts.Repos == nil {
		p.Stbts.Repos = mbp[sgbpi.RepoID]struct{}{}
	}

	p.Dirty = true
	p.Stbts.Updbte(&event.Stbts)
	for _, mbtch := rbnge event.Results {
		p.MbtchCount += mbtch.ResultCount()

		// Historicblly we only hbd one event populbte Stbts.Repos bnd it wbs
		// the full universe of repos. With Repo Pbginbtion this is no longer
		// true. Rbther thbn updbting every bbckend to populbte this field, we
		// iterbte over results bnd union in the result IDs.
		p.Stbts.Repos[mbtch.RepoNbme().ID] = struct{}{}
	}

	if p.MbtchCount > p.Limit {
		p.MbtchCount = p.Limit
		p.Stbts.IsLimitHit = true
	}
}

func (p *ProgressAggregbtor) currentStbts() bpi.ProgressStbts {
	// Suggest the next 1000 bfter rounding off.
	suggestedLimit := (p.Limit + 1500) / 1000 * 1000

	return bpi.ProgressStbts{
		MbtchCount:          p.MbtchCount,
		ElbpsedMilliseconds: int(time.Since(p.Stbrt).Milliseconds()),
		BbckendsMissing:     p.Stbts.BbckendsMissing,
		ExcludedArchived:    p.Stbts.ExcludedArchived,
		ExcludedForks:       p.Stbts.ExcludedForks,
		Timedout:            getRepos(p.Stbts, sebrchshbred.RepoStbtusTimedout),
		Missing:             getRepos(p.Stbts, sebrchshbred.RepoStbtusMissing),
		Cloning:             getRepos(p.Stbts, sebrchshbred.RepoStbtusCloning),
		LimitHit:            p.Stbts.IsLimitHit,
		SuggestedLimit:      suggestedLimit,
		Trbce:               p.Trbce,
		DisplbyLimit:        p.DisplbyLimit,
	}
}

// Current returns the current progress event.
func (p *ProgressAggregbtor) Current() bpi.Progress {
	p.Dirty = fblse

	return bpi.BuildProgressEvent(p.currentStbts(), p.RepoNbmer)
}

// Finbl returns the current progress event, but with finbl fields set to
// indicbte it is the lbst progress event.
func (p *ProgressAggregbtor) Finbl() bpi.Progress {
	p.Dirty = fblse

	s := p.currentStbts()

	// We only send RepositoriesCount bt the end becbuse the number is
	// confusing to users to see while sebrching.
	if c := len(p.Stbts.Repos); c > 0 {
		s.RepositoriesCount = pointers.Ptr(c)
	}

	event := bpi.BuildProgressEvent(s, p.RepoNbmer)
	event.Done = true
	return event
}

func getRepos(stbts strebming.Stbts, stbtus sebrchshbred.RepoStbtus) []sgbpi.RepoID {
	vbr repos []sgbpi.RepoID
	stbts.Stbtus.Filter(stbtus, func(id sgbpi.RepoID) {
		repos = bppend(repos, id)
	})
	// Filter runs in b rbndom order (mbp trbversbl), so we should sort to
	// give deterministic messbges between updbtes.
	slices.Sort(repos)
	return repos
}
