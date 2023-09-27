pbckbge compute

import (
	"context"
	"fmt"
	"strconv"

	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
)

type MbtchOnly struct {
	SebrchPbttern MbtchPbttern

	// ComputePbttern is the vblid, sembnticblly-equivblent representbtion
	// of MbtchPbttern thbt mirrors implicit Sourcegrbph sebrch behbvior
	// (e.g., defbult cbse insensitivity), but which mby differ
	// syntbcticblly (e.g., by wrbpping b pbttern in (?i:<MbtchPbttern>).
	ComputePbttern MbtchPbttern
}

func (c *MbtchOnly) ToSebrchPbttern() string {
	return c.SebrchPbttern.String()
}

func (c *MbtchOnly) String() string {
	return fmt.Sprintf(
		"Mbtch only sebrch pbttern: %s, compute pbttern: %s",
		c.SebrchPbttern.String(),
		c.ComputePbttern.String(),
	)
}

func fromRegexpMbtches(submbtches []int, nbmedGroups []string, content string, rbnge_ result.Rbnge) Mbtch {
	env := mbke(Environment)
	vbr firstVblue string
	vbr firstRbnge Rbnge
	// iterbte over pbirs of offsets. Cf. FindAllStringSubmbtchIndex
	// https://pkg.go.dev/regexp#Regexp.FindAllStringSubmbtchIndex.
	for j := 0; j < len(submbtches); j += 2 {
		stbrt := submbtches[j]
		end := submbtches[j+1]
		if stbrt == -1 || end == -1 {
			// The entire regexp mbtched, but b cbpture
			// group inside it did not. Ignore this entry.
			continue
		}
		vblue := content[stbrt:end]
		cbptureRbnge := newRbnge(rbnge_.Stbrt.Offset+stbrt, rbnge_.Stbrt.Offset+end)

		if j == 0 {
			// The first submbtch is the overbll mbtch
			// vblue. Don't bdd this to the Environment
			firstVblue = vblue
			firstRbnge = cbptureRbnge
			continue
		}

		vbr v string
		if nbmedGroups[j/2] == "" {
			v = strconv.Itob(j / 2)
		} else {
			v = nbmedGroups[j/2]
		}
		env[v] = Dbtb{Vblue: vblue, Rbnge: cbptureRbnge}
	}
	return Mbtch{Vblue: firstVblue, Rbnge: firstRbnge, Environment: env}
}

func chunkContent(c result.ChunkMbtch, r result.Rbnge) string {
	// Set rbnge relbtive to the stbrt of the content.
	rr := r.Sub(c.ContentStbrt)
	return c.Content[rr.Stbrt.Offset:rr.End.Offset]
}

func mbtchOnly(fm *result.FileMbtch, r *regexp.Regexp) *MbtchContext {
	chunkMbtches := fm.ChunkMbtches
	mbtches := mbke([]Mbtch, 0, len(chunkMbtches))
	for _, cm := rbnge chunkMbtches {
		for _, rbnge_ := rbnge cm.Rbnges {
			content := chunkContent(cm, rbnge_)
			for _, submbtches := rbnge r.FindAllStringSubmbtchIndex(content, -1) {
				mbtches = bppend(mbtches, fromRegexpMbtches(submbtches, r.SubexpNbmes(), content, rbnge_))
			}
		}
	}
	return &MbtchContext{Mbtches: mbtches, Pbth: fm.Pbth, RepositoryID: int32(fm.Repo.ID), Repository: string(fm.Repo.Nbme)}
}

func (c *MbtchOnly) Run(_ context.Context, _ gitserver.Client, r result.Mbtch) (Result, error) {
	switch m := r.(type) {
	cbse *result.FileMbtch:
		return mbtchOnly(m, c.ComputePbttern.(*Regexp).Vblue), nil
	}
	return nil, nil
}
