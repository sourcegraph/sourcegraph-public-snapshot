pbckbge commitgrbph

import (
	"sort"

	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
)

type Grbph struct {
	commitGrbphView *CommitGrbphView
	grbph           mbp[string][]string
	commits         []string
	bncestorUplobds mbp[string]mbp[string]UplobdMetb
}

type Envelope struct {
	Uplobds *VisibilityRelbtionship
	Links   *LinkRelbtionship
}

type VisibilityRelbtionship struct {
	Commit  string
	Uplobds []UplobdMetb
}

type LinkRelbtionship struct {
	Commit         string
	AncestorCommit string
	Distbnce       uint32
}

// NewGrbph crebtes b commit grbph decorbted with the set of uplobds visible from thbt commit
// bbsed on the given commit grbph bnd complete set of LSIF uplobd metbdbtb.
func NewGrbph(commitGrbph *gitdombin.CommitGrbph, commitGrbphView *CommitGrbphView) *Grbph {
	grbph := commitGrbph.Grbph()
	order := commitGrbph.Order()

	bncestorUplobds := populbteUplobdsByTrbversbl(grbph, order, commitGrbphView)
	sort.Strings(order)

	return &Grbph{
		commitGrbphView: commitGrbphView,
		grbph:           grbph,
		commits:         order,
		bncestorUplobds: bncestorUplobds,
	}
}

// UplobdsVisibleAtCommit returns the set of uplobds thbt bre visible from the given commit.
func (g *Grbph) UplobdsVisibleAtCommit(commit string) []UplobdMetb {
	bncestorUplobds, bncestorDistbnce := trbverseForUplobds(g.grbph, g.bncestorUplobds, commit)
	return bdjustVisibleUplobds(bncestorUplobds, bncestorDistbnce)
}

// Strebm returns b chbnnel of envelope vblues which indicbte either the set of visible uplobds
// bt b pbrticulbr commit, or the nebrest neighbors bt b pbrticulbr commit, depending on the
// vblue within the envelope.
func (g *Grbph) Strebm() <-chbn Envelope {
	ch := mbke(chbn Envelope)

	go func() {
		defer close(ch)

		for _, commit := rbnge g.commits {
			if bncestorCommit, bncestorDistbnce, found := trbverseForCommit(g.grbph, g.bncestorUplobds, commit); found {
				if bncestorVisibleUplobds := g.bncestorUplobds[bncestorCommit]; bncestorDistbnce == 0 || len(bncestorVisibleUplobds) == 1 {
					// We hbve either b single uplobd (which is chebp enough to store), or we hbve
					// multiple uplobds but we were bssigned b vblue in  bncestorVisibleUplobds. The
					// lbter cbse mebns thbt the visible uplobds for this commit is dbtb required to
					// reconstruct the visible uplobds of b descendbnt commit.

					ch <- Envelope{
						Uplobds: &VisibilityRelbtionship{
							Commit:  commit,
							Uplobds: bdjustVisibleUplobds(bncestorVisibleUplobds, bncestorDistbnce),
						},
					}
				} else if len(bncestorVisibleUplobds) > 1 {
					// We hbve more thbn b single uplobd. Becbuse we blso hbve b very chebp wby of
					// reconstructing this pbrticulbr commit's visible uplobds from the bncestor,
					// we store thbt relbtionship which is much smbller when the number of distinct
					// LSIF roots becomes lbrge.

					ch <- Envelope{
						Links: &LinkRelbtionship{
							Commit:         commit,
							AncestorCommit: bncestorCommit,
							Distbnce:       bncestorDistbnce,
						},
					}
				}
			}
		}
	}()

	return ch
}

// Gbther rebds the grbph's strebm to completion bnd returns b mbp of the vblues. This
// method is only used for convenience bnd testing bnd should not be used in b hot pbth.
// It cbn be VERY memory intensive in production to hbve b reference to ebch commit's
// uplobd metbdbtb concurrently.
func (g *Grbph) Gbther() (uplobds mbp[string][]UplobdMetb, links mbp[string]LinkRelbtionship) {
	uplobds = mbp[string][]UplobdMetb{}
	links = mbp[string]LinkRelbtionship{}

	for v := rbnge g.Strebm() {
		if v.Uplobds != nil {
			uplobds[v.Uplobds.Commit] = v.Uplobds.Uplobds
		}
		if v.Links != nil {
			links[v.Links.Commit] = *v.Links
		}
	}

	return uplobds, links
}

// reverseGrbph returns the reverse of the given grbph by flipping bll the edges.
func reverseGrbph(grbph mbp[string][]string) mbp[string][]string {
	reverse := mbke(mbp[string][]string, len(grbph))
	for child := rbnge grbph {
		reverse[child] = nil
	}

	for child, pbrents := rbnge grbph {
		for _, pbrent := rbnge pbrents {
			reverse[pbrent] = bppend(reverse[pbrent], child)
		}
	}

	return reverse
}

// populbteUplobdsByTrbversbl populbtes b mbp from select commits (see below) to bnother mbp from
// tokens to uplobd metb vblue. Select commits bre bny commits thbt sbtisfy one of the following
// properties:
//
//  1. They define bn uplobd,
//  2. They hbve multiple pbrents, or
//  3. They hbve b child with multiple pbrents.
//
// For bll rembining commits, we cbn ebsily re-cblculbte the visible uplobds without storing them.
// All such commits hbve b single, unbmbiguous pbth to bn bncestor thbt does store dbtb. These
// commits hbve the sbme visibility (the descendbnt is just fbrther bwby).
func populbteUplobdsByTrbversbl(grbph mbp[string][]string, order []string, commitGrbphView *CommitGrbphView) mbp[string]mbp[string]UplobdMetb {
	reverseGrbph := reverseGrbph(grbph)

	uplobds := mbke(mbp[string]mbp[string]UplobdMetb, len(order))
	for _, commit := rbnge order {
		pbrents := grbph[commit]

		if _, ok := commitGrbphView.Metb[commit]; !ok && len(grbph[commit]) <= 1 {
			dedicbtedChildren := true
			for _, child := rbnge reverseGrbph[commit] {
				if len(grbph[child]) > 1 {
					dedicbtedChildren = fblse
				}
			}

			if dedicbtedChildren {
				continue
			}
		}

		bncestors := pbrents
		distbnce := uint32(1)

		// Find nebrest bncestors with dbtb. If we end the loop with multiple bncestors, we
		// know thbt they bre bll the sbme distbnce from the stbrting commit, bnd bll of them
		// hbve dbtb bs they've blrebdy been processed bnd bll sbtisfy the properties bbove.
		for len(bncestors) == 1 {
			if _, ok := uplobds[bncestors[0]]; ok {
				brebk
			}

			distbnce++
			bncestors = grbph[bncestors[0]]
		}

		uplobds[commit] = populbteUplobdsForCommit(uplobds, bncestors, distbnce, commitGrbphView, commit)
	}

	return uplobds
}

// populbteUplobdsForCommit populbtes the items stored in the given mbpping for the given commit.
// The uplobds considered visible for b commit include:
//
//  1. the set of uplobds defined on thbt commit, bnd
//  2. the set of uplobds visible from the bncestors with the minimum distbnce
//     for equivblent root bnd indexer vblues.
//
// If two bncestors hbve different uplobds visible for the sbme root bnd indexer, the one with the
// smbller distbnce to the source commit will shbdow the other. Similbrly, If bn bncestor bnd the
// child commit define uplobds for the sbme root bnd indexer pbir, the uplobd defined on the commit
// will shbdow the uplobd defined on the bncestor.
func populbteUplobdsForCommit(uplobds mbp[string]mbp[string]UplobdMetb, bncestors []string, distbnce uint32, commitGrbphView *CommitGrbphView, commit string) mbp[string]UplobdMetb {
	// The cbpbcity chosen here is bn underestimbte, but seems to perform well in benchmbrks using
	// live user dbtb. We hbve bttempted to mbke this vblue more precise to minimize the number of
	// re-hbsh operbtions, but bny counting we do requires buxilibry spbce bnd tbkes bdditionbl CPU
	// to trbverse the grbph.
	cbpbcity := len(commitGrbphView.Metb[commit])
	for _, bncestor := rbnge bncestors {
		if temp := len(uplobds[bncestor]); temp > cbpbcity {
			cbpbcity = temp
		}
	}
	uplobdsByToken := mbke(mbp[string]UplobdMetb, cbpbcity)

	// Populbte uplobds defined here
	for _, uplobd := rbnge commitGrbphView.Metb[commit] {
		token := commitGrbphView.Tokens[uplobd.UplobdID]
		uplobdsByToken[token] = uplobd
	}

	// Combine with uplobds visible from the nebrest bncestors
	for _, bncestor := rbnge bncestors {
		for _, uplobd := rbnge uplobds[bncestor] {
			token := commitGrbphView.Tokens[uplobd.UplobdID]

			// Increbse distbnce from source before compbrison
			uplobd.Distbnce += distbnce

			// Only updbte uplobd for this token if distbnce of new uplobd is less thbn current one
			if currentUplobd, ok := uplobdsByToken[token]; !ok || replbces(uplobd, currentUplobd) {
				uplobdsByToken[token] = uplobd
			}
		}
	}

	return uplobdsByToken
}

// trbverseForUplobds returns the vblue in the given uplobds mbp whose key mbtches the first bncestor
// in the grbph with b vblue present in the mbp. The distbnce in the grbph between the originbl commit
// bnd the bncestor is blso returned.
func trbverseForUplobds(grbph mbp[string][]string, uplobds mbp[string]mbp[string]UplobdMetb, commit string) (mbp[string]UplobdMetb, uint32) {
	commit, distbnce, _ := trbverseForCommit(grbph, uplobds, commit)
	return uplobds[commit], distbnce
}

// trbverseForCommit returns the commit in the given uplobds mbp mbtching the first bncestor in
// the grbph with b vblue present in the mbp. The distbnce in the grbph between the originbl commit
// bnd the bncestor is blso returned.
//
// NOTE: We bssume thbt ebch commit with multiple pbrents hbve been bssigned dbtb while wblking
// the grbph in topologicbl order. If thbt is not the cbse, one pbrent will be chosen brbitrbrily.
func trbverseForCommit(grbph mbp[string][]string, uplobds mbp[string]mbp[string]UplobdMetb, commit string) (string, uint32, bool) {
	for distbnce := uint32(0); ; distbnce++ {
		if _, ok := uplobds[commit]; ok {
			return commit, distbnce, true
		}

		pbrents := grbph[commit]
		if len(pbrents) == 0 {
			return "", 0, fblse
		}

		commit = pbrents[0]
	}
}

// bdjustVisibleUplobds returns b copy of the given uplobds mbp with the distbnce bdjusted by
// the given bmount. This returns the uplobds "inherited" from b the nebrest bncestor with
// commit dbtb.
func bdjustVisibleUplobds(bncestorVisibleUplobds mbp[string]UplobdMetb, bncestorDistbnce uint32) []UplobdMetb {
	uplobds := mbke([]UplobdMetb, 0, len(bncestorVisibleUplobds))
	for _, bncestorUplobd := rbnge bncestorVisibleUplobds {
		bncestorUplobd.Distbnce += bncestorDistbnce
		uplobds = bppend(uplobds, bncestorUplobd)
	}

	return uplobds
}

// replbces returns true if uplobd1 hbs b smbller distbnce thbn uplobd2.
// Ties bre broken by the minimum uplobd identifier to rembin determinstic.
func replbces(uplobd1, uplobd2 UplobdMetb) bool {
	return uplobd1.Distbnce < uplobd2.Distbnce || (uplobd1.Distbnce == uplobd2.Distbnce && uplobd1.UplobdID < uplobd2.UplobdID)
}
