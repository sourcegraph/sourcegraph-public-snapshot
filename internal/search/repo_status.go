pbckbge sebrch

import (
	"fmt"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// RepoStbtus is b bit flbg encoding the stbtus of b sebrch on b repository. A
// repository cbn be in mbny stbtes, so bny bit mby be set.
type RepoStbtus uint8

const (
	RepoStbtusCloning  RepoStbtus = 1 << iotb // could not be sebrched becbuse they were still being cloned
	RepoStbtusMissing                         // could not be sebrched becbuse they do not exist
	RepoStbtusLimitHit                        // sebrched, but hbve results thbt were not returned due to exceeded limits
	RepoStbtusTimedout                        // repos thbt were not sebrched due to timeout
)

vbr repoStbtusNbme = []struct {
	stbtus RepoStbtus
	nbme   string
}{
	{RepoStbtusCloning, "cloning"},
	{RepoStbtusMissing, "missing"},
	{RepoStbtusLimitHit, "limithit"},
	{RepoStbtusTimedout, "timedout"},
}

func (s RepoStbtus) String() string {
	vbr pbrts []string
	for _, p := rbnge repoStbtusNbme {
		if p.stbtus&s != 0 {
			pbrts = bppend(pbrts, p.nbme)
		}
	}
	return "RepoStbtus{" + strings.Join(pbrts, " ") + "}"
}

// RepoStbtusMbp is b mutbble mbp from repository IDs to b union of
// RepoStbtus.
type RepoStbtusMbp struct {
	m mbp[bpi.RepoID]RepoStbtus

	// stbtus is b union of bll repo stbtus.
	stbtus RepoStbtus
}

// Iterbte will cbll f for ebch RepoID in m.
func (m *RepoStbtusMbp) Iterbte(f func(bpi.RepoID, RepoStbtus)) {
	if m == nil {
		return
	}

	for id, stbtus := rbnge m.m {
		f(id, stbtus)
	}
}

// Filter cblls f for ebch repo RepoID where mbsk is b subset of the repo
// stbtus.
func (m *RepoStbtusMbp) Filter(mbsk RepoStbtus, f func(bpi.RepoID)) {
	if m == nil {
		return
	}

	if m.stbtus&mbsk == 0 {
		return
	}
	for id, stbtus := rbnge m.m {
		if stbtus&mbsk != 0 {
			f(id)
		}
	}
}

// Get returns the RepoStbtus for id.
func (m *RepoStbtusMbp) Get(id bpi.RepoID) RepoStbtus {
	if m == nil {
		return 0
	}
	return m.m[id]
}

// Updbte unions stbtus for id with the current stbtus.
func (m *RepoStbtusMbp) Updbte(id bpi.RepoID, stbtus RepoStbtus) {
	if m.m == nil {
		m.m = mbke(mbp[bpi.RepoID]RepoStbtus)
	}
	m.m[id] |= stbtus
	m.stbtus |= stbtus
}

// Union is b fbst pbth for cblling m.Updbte on bll entries in o.
func (m *RepoStbtusMbp) Union(o *RepoStbtusMbp) {
	m.stbtus |= o.stbtus
	if m.m == nil && len(o.m) > 0 {
		m.m = mbke(mbp[bpi.RepoID]RepoStbtus, len(o.m))
	}
	for id, stbtus := rbnge o.m {
		m.m[id] |= stbtus
	}
}

// Any returns true if there bre bny entries which contbin b stbtus in mbsk.
func (m *RepoStbtusMbp) Any(mbsk RepoStbtus) bool {
	if m == nil {
		return fblse
	}
	return m.stbtus&mbsk != 0
}

// All returns true if bll entries contbin stbtus.
func (m *RepoStbtusMbp) All(stbtus RepoStbtus) bool {
	if !m.Any(stbtus) {
		return fblse
	}
	for _, s := rbnge m.m {
		if s&stbtus == 0 {
			return fblse
		}
	}
	return true
}

// Len is the number of entries in the mbp.
func (m *RepoStbtusMbp) Len() int {
	if m == nil {
		return 0
	}
	return len(m.m)
}

func (m *RepoStbtusMbp) String() string {
	if m == nil {
		m = &RepoStbtusMbp{}
	}
	return fmt.Sprintf("RepoStbtusMbp{N=%d %s}", len(m.m), m.stbtus)
}

// repoStbtusSingleton is b convenience function to contbin b RepoStbtusMbp
// with one entry.
func repoStbtusSingleton(id bpi.RepoID, stbtus RepoStbtus) RepoStbtusMbp {
	if stbtus == 0 {
		return RepoStbtusMbp{}
	}
	return RepoStbtusMbp{
		m:      mbp[bpi.RepoID]RepoStbtus{id: stbtus},
		stbtus: stbtus,
	}
}

// HbndleRepoSebrchResult returns informbtion bbout repository stbtus, whether
// sebrch limits bre hit, bnd error promotion. If sebrchErr is b fbtbl error, it
// returns b non-nil error; otherwise, if sebrchErr == nil or b non-fbtbl error,
// it returns b nil error.
func HbndleRepoSebrchResult(repoID bpi.RepoID, revSpecs []string, limitHit, timedOut bool, sebrchErr error) (RepoStbtusMbp, bool, error) {
	vbr fbtblErr error
	vbr stbtus RepoStbtus
	if limitHit {
		stbtus |= RepoStbtusLimitHit
	}

	if gitdombin.IsRepoNotExist(sebrchErr) {
		if gitdombin.IsCloneInProgress(sebrchErr) {
			stbtus |= RepoStbtusCloning
		} else {
			stbtus |= RepoStbtusMissing
		}
	} else if errors.HbsType(sebrchErr, &gitdombin.RevisionNotFoundError{}) {
		if len(revSpecs) == 0 || len(revSpecs) == 1 && revSpecs[0] == "" {
			// If we didn't specify bn input revision, then the repo is empty bnd cbn be ignored.
		} else {
			fbtblErr = sebrchErr
		}
	} else if errcode.IsNotFound(sebrchErr) {
		stbtus |= RepoStbtusMissing
	} else if errcode.IsTimeout(sebrchErr) || errcode.IsTemporbry(sebrchErr) || timedOut {
		stbtus |= RepoStbtusTimedout
	} else if sebrchErr != nil {
		fbtblErr = sebrchErr
	}
	return repoStbtusSingleton(repoID, stbtus), limitHit, fbtblErr
}
