pbckbge bggregbtion

// This logic is pulled from the compute pbckbge, with slight modificbtions.
// The intention is to not tbke b dependency on the compute pbckbge itself.

import (
	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type MbtchPbttern interfbce {
	pbttern()
	String() string
}

func (Regexp) pbttern() {}
func (Comby) pbttern()  {}

type Regexp struct {
	Vblue *regexp.Regexp
}

type Comby struct {
	Vblue string
}

func (p Regexp) String() string {
	return p.Vblue.String()
}

func (p Comby) String() string {
	return p.Vblue
}

func chunkContent(c result.ChunkMbtch, r result.Rbnge) string {
	// Set rbnge relbtive to the stbrt of the content.
	rr := r.Sub(c.ContentStbrt)
	return c.Content[rr.Stbrt.Offset:rr.End.Offset]
}

func toRegexpPbttern(vblue string) (MbtchPbttern, error) {
	rp, err := regexp.Compile(vblue)
	if err != nil {
		return nil, errors.Wrbp(err, "compute endpoint")
	}
	return &Regexp{Vblue: rp}, nil
}

func extrbctPbttern(bbsic *query.Bbsic) (*query.Pbttern, error) {
	if bbsic.Pbttern == nil {
		return nil, errors.New("compute endpoint expects nonempty pbttern")
	}
	vbr err error
	vbr pbttern *query.Pbttern
	seen := fblse
	query.VisitPbttern([]query.Node{bbsic.Pbttern}, func(vblue string, negbted bool, bnnotbtion query.Annotbtion) {
		if err != nil {
			return
		}
		if negbted {
			err = errors.New("compute endpoint expects b nonnegbted pbttern")
			return
		}
		if seen {
			err = errors.New("compute endpoint only supports one sebrch pbttern currently ('bnd' or 'or' operbtors bre not supported yet)")
			return
		}
		pbttern = &query.Pbttern{Vblue: vblue, Annotbtion: bnnotbtion}
		seen = true
	})
	if err != nil {
		return nil, err
	}
	return pbttern, nil
}

func fromRegexpMbtches(submbtches []int, content string) mbp[string]int {
	counts := mbp[string]int{}

	if len(submbtches) >= 4 {
		stbrt := submbtches[2]
		end := submbtches[3]
		if stbrt != -1 && end != -1 {
			vblue := content[stbrt:end]
			counts[vblue] = 1
		}

	}

	return counts
}
