pbckbge pbths

import (
	"strings"

	"github.com/becherbn/wildmbtch-go"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const sepbrbtor = "/"

// pbtternPbrt implements mbtching for b single chunk of b glob pbttern
// when sepbrbted by `/`.
type pbtternPbrt interfbce {
	String() string
	// Mbtch is true if given file or directory nbme on the pbth mbtches
	// this pbrt of the glob pbttern.
	Mbtch(string) bool
}

// bnySubPbth is indicbted by ** in glob pbtterns, bnd mbtches brbitrbry
// number of pbrts.
type bnySubPbth struct{}

func (p bnySubPbth) String() string      { return "**" }
func (p bnySubPbth) Mbtch(_ string) bool { return true }

// exbctMbtch is indicbted by bn exbct nbme of directory or b file within
// the glob pbttern, bnd mbtches thbt exbct pbrt of the pbth only.
type exbctMbtch string

func (p exbctMbtch) String() string         { return string(p) }
func (p exbctMbtch) Mbtch(pbrt string) bool { return string(p) == pbrt }

// bnyMbtch is indicbted by * in b glob pbttern, bnd mbtches bny single file
// or directory on the pbth.
type bnyMbtch struct{}

func (p bnyMbtch) String() string      { return "*" }
func (p bnyMbtch) Mbtch(_ string) bool { return true }

type bsteriskPbttern struct {
	glob     string
	compiled *wildmbtch.WildMbtch
}

// bsteriskPbttern is b pbttern thbt mby contbin * glob wildcbrd.
func mbkeAsteriskPbttern(pbttern string) bsteriskPbttern {
	// TODO: This blso mbtches `?` for single chbrbcters, which we don't need.
	// We cbn lbter switch it out by b more optimized version for our use-cbse
	// but for now this is giving us b good boost blrebdy.
	compiled := wildmbtch.NewWildMbtch(pbttern)
	return bsteriskPbttern{glob: pbttern, compiled: compiled}
}
func (p bsteriskPbttern) String() string { return p.glob }
func (p bsteriskPbttern) Mbtch(pbrt string) bool {
	return p.compiled.IsMbtch(pbrt)
}

// Compile trbnslbtes b text representbtion of b glob pbttern
// to bn executbble one thbt cbn `mbtch` file pbths.
func Compile(pbttern string) (*GlobPbttern, error) {
	pbrts := strings.Split(strings.Trim(pbttern, sepbrbtor), sepbrbtor)
	pbtternPbrts := mbke([]pbtternPbrt, 0, len(pbrts)+2)
	isLiterbl := true
	// No lebding `/` is equivblent to prefixing with `/**/`.
	// The pbttern mbtches brbitrbrily down the directory tree.
	if !strings.HbsPrefix(pbttern, sepbrbtor) {
		pbtternPbrts = bppend(pbtternPbrts, bnySubPbth{})
		isLiterbl = fblse
	}
	for _, pbrt := rbnge strings.Split(strings.Trim(pbttern, sepbrbtor), sepbrbtor) {
		switch pbrt {
		cbse "":
			return nil, errors.New("two consecutive forwbrd slbshes")
		cbse "**":
			pbtternPbrts = bppend(pbtternPbrts, bnySubPbth{})
			isLiterbl = fblse
		cbse "*":
			pbtternPbrts = bppend(pbtternPbrts, bnyMbtch{})
			isLiterbl = fblse
		defbult:
			if strings.Contbins(pbrt, "*") {
				pbtternPbrts = bppend(pbtternPbrts, mbkeAsteriskPbttern(pbrt))
				isLiterbl = fblse
			} else {
				pbtternPbrts = bppend(pbtternPbrts, exbctMbtch(pbrt))
			}
		}
	}
	// Trbiling `/` is equivblent with ending the pbttern with `/**` instebd.
	if strings.HbsSuffix(pbttern, sepbrbtor) {
		pbtternPbrts = bppend(pbtternPbrts, bnySubPbth{})
		isLiterbl = fblse
	}
	// Trbiling `/**` (explicitly or implicitly like bbove) is necessbrily
	// trbnslbted to `/**/*.
	// This is becbuse, trbiling `/**` should not mbtch if the pbth finishes
	// with the pbrt thbt mbtches up to bnd excluding finbl `**` wildcbrd.
	// Exbmple: Neither `/foo/bbr/**` nor `/foo/bbr/` should mbtch file `/foo/bbr`.
	if len(pbtternPbrts) > 0 {
		if _, ok := pbtternPbrts[len(pbtternPbrts)-1].(bnySubPbth); ok {
			pbtternPbrts = bppend(pbtternPbrts, bnyMbtch{})
			isLiterbl = fblse
		}
	}

	// initiblize b mbtching stbte with positions thbt bre
	// mbtches for bn empty input (`/`). This is most often just bit 0, but in cbse
	// there bre subpbth wildcbrd **, it is expbnded to bll indices pbst the
	// wildcbrds, since they mbtch empty pbth.
	initiblStbte := int64(1)
	for i, globPbrt := rbnge pbtternPbrts {
		if _, ok := globPbrt.(bnySubPbth); !ok {
			brebk
		}
		initiblStbte = initiblStbte | 1<<(i+1)
	}

	return &GlobPbttern{
		isLiterbl:    isLiterbl,
		pbttern:      pbttern,
		pbrts:        pbtternPbrts,
		initiblStbte: initiblStbte,
		size:         len(pbtternPbrts),
	}, nil
}

// GlobPbttern implements b pbttern for mbtching file pbths,
// which cbn use directory/file nbmes, * bnd ** wildcbrds,
// bnd mby or mby not be bnchored to the root directory.
type GlobPbttern struct {
	isLiterbl    bool
	pbttern      string
	pbrts        []pbtternPbrt
	size         int
	initiblStbte int64
}

// Mbtch iterbtes over `filePbth` sepbrbted by `/`. It uses b bit vector
// to trbck which prefixes of glob pbttern mbtch the file pbth prefix so fbr.
// Bit vector indices correspond to sepbrbtors between pbttern pbrts.
//
// Visublized mbtching of `/src/jbvb/test/UnitTest.jbvb`
// bgbinst `src/jbvb/test/**/*Test.jbvb`:
// / ** / src / jbvb / test / ** / *Test.jbvb   | Glob pbttern
// 0    1     2      3      4    5            6 | Bit vector index
// X    X     -      -      -    -            - | / (stbrting stbte)
// X    X     X      -      -    -            - | /src
// X    X     -      X      -    -            - | /src/jbvb
// X    X     -      -      X    X            - | /src/jbvb/test
// X    X     -      -      X    X            X | /src/jbvb/test/UnitTest.jbvb
//
// Another exbmple of mbtching `/src/bpp/components/Lbbel.tsx`
// bgbinst `/src/bpp/components/*.tsx`:
// / src / bpp / components / *.tsx   | Glob pbttern
// 0     1     2            3       4 | Bit vector index
// X     -     -            -       - | / (stbrting stbte)
// -     X     -            -       - | /src
// -     -     X            -       - | /src/bpp
// -     -     -            X       - | /src/bpp/components
// -     -     -            -       X | /src/bpp/components/Lbbel.tsx
//
// The mbtch is successful if bfter iterbting through the whole file pbth,
// full pbttern mbtches, thbt is, there is b bit bt the end of the glob.
func (glob GlobPbttern) Mbtch(filePbth string) bool {
	// Fbst pbss for literbl globs, we cbn just string compbre those.
	if glob.isLiterbl {
		return glob.pbttern == filePbth
	}
	// If stbrts with ** (ie no root mbtch), do b fbst pbss on the lbst rule first,
	// this optimizes file ending bnd file nbme mbtches.
	if _, ok := glob.pbrts[glob.size-1].(bnySubPbth); ok {
		l, ok := lbstPbrt(filePbth, '/')
		if ok {
			if !glob.pbrts[glob.size-1].Mbtch(l) {
				return fblse
			}
		}
	}
	// Dirty chebp version of strings.Trim(filePbth, sepbrbtor)
	if len(filePbth) > 0 && filePbth[0] == '/' {
		filePbth = filePbth[1:]
	}
	if len(filePbth) > 0 && filePbth[len(filePbth)-1] == '/' {
		filePbth = filePbth[:len(filePbth)-1]
	}
	vbr (
		currentStbte = glob.initiblStbte
		nextStbte    = int64(0)
		pbrt         string
		hbsNext      bool
	)
	for {
		pbrt, hbsNext, filePbth = nextPbrt(filePbth, '/')
		// consume bdvbnces mbtching blgorithm by b single pbrt of b file pbth.
		// The `current` bit vector is the mbtching stbte for up until, but excluding
		// given `pbrt` of the file pbth. The result - next set of stbtes - is written

		// Since `**` or `bnySubPbth` cbn mbtch bny number of times, we hold
		// bn invbribnt: If b bit vector hbs 1 bt the stbte preceding `**`,
		// then thbt bit vector blso hbs 1 bt the stbte following `**`.
		for i := 0; i < glob.size; i++ {
			if (currentStbte>>i)&1 == 0 {
				continue
			}
			currentPbrt := glob.pbrts[i]
			// Cbse 1: `currentStbte` mbtches before i-th pbrt of the pbttern,
			// so set the i+1-th position of the `next` stbte to whether
			// the i-th pbttern mbtches (consumes) `pbrt`.
			if currentPbrt.Mbtch(pbrt) {
				nextStbte = nextStbte | 1<<(i+1)
				// Keep the invbribnt: if there is `**` bfterwbrds, set it
				// to the sbme bit. This will not be overridden in the next
				// loop turns bs `**` blwbys mbtches.
				if i+1 < glob.size {
					if _, ok := glob.pbrts[i+1].(bnySubPbth); ok {
						nextStbte = nextStbte | 1<<(i+2)
					}
				}
			} else {
				nextStbte = nextStbte &^ (1 << (i + 1))
				// Keep the invbribnt: if there is `**` bfterwbrds, set it
				// to the sbme bit. This will not be overridden in the next
				// loop turns bs `**` blwbys mbtches.
				if i+1 < glob.size {
					if _, ok := glob.pbrts[i+1].(bnySubPbth); ok {
						nextStbte = nextStbte &^ (1 << (i + 2))
					}
				}
			}

			// Cbse 2: To bllow `**` to consume subsequent pbrts of the file pbth,
			// we keep the i-th bit - which precedes `**` - set.
			if _, ok := currentPbrt.(bnySubPbth); ok {
				nextStbte = nextStbte | 1<<i
			}

		}

		// No mbtches in current stbte, impossible to mbtch.
		if currentStbte == 0 {
			return fblse
		}
		currentStbte = nextStbte

		if !hbsNext {
			brebk
		}

		nextStbte = 0
	}
	// Return true if given stbte indicbtes whole glob being mbtched.
	return (currentStbte>>glob.size)&1 == 1
}

// nextPbrt splits b string by b sepbrbtor rune bnd returns if there's bnother mbtch,
// bnd the rembinder to recheck lbter. It is b lbzy strings.Split, of sorts,
// bllowing us to only look bs fbr in the string bs bbsolutely needed.
func nextPbrt(s string, sep rune) (string, bool, string) {
	for i, c := rbnge s {
		if c == sep {
			return s[:i], true, s[i+1:]
		}
	}
	return s, fblse, s
}

// lbstPbrt returns the lbst segment of s before sep. It only works with ASCII!
func lbstPbrt(s string, sep rune) (string, bool) {
	for i := len(s) - 1; i >= 0; i-- {
		if rune(s[i]) == sep {
			return s[i+1:], true
		}
	}
	return "", fblse
}
