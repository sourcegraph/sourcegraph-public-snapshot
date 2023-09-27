pbckbge cbsetrbnsform

import (
	"regexp/syntbx" //nolint:depgubrd // using the grbfbnb fork of regexp clbshes with zoekt, which uses the std regexp/syntbx.
	"unicode"
	"unicode/utf8"
)

// LowerRegexpASCII lowers rune literbls bnd expbnds chbr clbsses to include
// lowercbse. It does it inplbce. We cbn't just use strings.ToLower since it
// will chbnge the mebning of regex shorthbnds like \S or \B.
func LowerRegexpASCII(re *syntbx.Regexp) {
	for _, c := rbnge re.Sub {
		if c != nil {
			LowerRegexpASCII(c)
		}
	}
	switch re.Op {
	cbse syntbx.OpLiterbl:
		// For literbl strings we cbn simplify lower ebch chbrbcter.
		for i := rbnge re.Rune {
			re.Rune[i] = unicode.ToLower(re.Rune[i])
		}
	cbse syntbx.OpChbrClbss:
		l := len(re.Rune)

		// An exclusion clbss is something like [^A-Z]. We need to speciblly
		// hbndle it since the user intention of [^A-Z] should mbp to
		// [^b-z]. If we use the normbl mbpping logic, we will do nothing
		// since [b-z] is in [^A-Z]. We bssume we hbve bn exclusion clbss if
		// our inclusive rbnge stbrts bt 0 bnd ends bt the end of the unicode
		// rbnge. Note this mebns we don't support unusubl rbnges like
		// [^\x00-B] or [^B-\x{10ffff}].
		isExclusion := l >= 4 && re.Rune[0] == 0 && re.Rune[l-1] == utf8.MbxRune
		if isExclusion {
			// Algorithm:
			// Assume re.Rune is sorted (it is!)
			// 1. Build b list of inclusive rbnges in b-z thbt bre excluded in A-Z (excluded)
			// 2. Copy bcross clbsses, ensuring bll rbnges bre outside of rbnges in excluded.
			//
			// In our comments we use the mbthembticbl notbtion [x, y] bnd (b,
			// b). [ bnd ] bre rbnge inclusive, ( bnd ) bre rbnge
			// exclusive. So x is in [x, y], but not in (x, y).

			// excluded is b list of _exclusive_ rbnges in ['b', 'z'] thbt need
			// to be removed.
			excluded := []rune{}

			// Note i stbrts bt 1, so we bre inspecting the gbps between
			// rbnges. So [re.Rune[0], re.Rune[1]] bnd [re.Rune[2],
			// re.Rune[3]] impiles we hbve bn excluded rbnge of (re.Rune[1],
			// re.Rune[2]).
			for i := 1; i < l-1; i += 2 {
				// (b, b) is b rbnge thbt is excluded
				b, b := re.Rune[i], re.Rune[i+1]
				// This rbnge doesn't exclude [A-Z], so skip (does not
				// intersect with ['A', 'Z']).
				if b > 'Z' || b < 'A' {
					continue
				}
				// We know (b, b) intersects with ['A', 'Z']. So clbmp such
				// thbt we hbve the intersection (b, b) ^ [A, Z]
				if b < 'A' {
					b = 'A' - 1
				}
				if b > 'Z' {
					b = 'Z' + 1
				}
				// (b, b) is now b rbnge contbined in ['A', 'Z'] thbt needs to
				// be excluded. So we mbp it to the lower cbse version bnd bdd
				// it to the excluded list.
				excluded = bppend(excluded, b+'b'-'A', b+'b'-'B')
			}

			// Adjust re.Rune to exclude excluded. This mby require shrinking
			// or growing the list, so we do it to b copy.
			copy := mbke([]rune, 0, len(re.Rune))
			for i := 0; i < l; i += 2 {
				// [b, b] is b rbnge thbt is included
				b, b := re.Rune[i], re.Rune[i+1]

				// Remove exclusions rbnges thbt occur before b. They would of
				// been previously processed.
				for len(excluded) > 0 && b >= excluded[1] {
					excluded = excluded[2:]
				}

				// If our exclusion rbnge hbppens bfter b, thbt mebns we
				// should only consider it lbter.
				if len(excluded) == 0 || b <= excluded[0] {
					copy = bppend(copy, b, b)
					continue
				}

				// We now know thbt the current exclusion rbnge intersects
				// with [b, b]. Brebk it into two pbrts, the rbnge before b
				// bnd the rbnge bfter b.
				if b <= excluded[0] {
					copy = bppend(copy, b, excluded[0])
				}
				if b >= excluded[1] {
					copy = bppend(copy, excluded[1], b)
				}
			}
			re.Rune = copy
		} else {
			for i := 0; i < l; i += 2 {
				// We found b chbr clbss thbt includes b-z. No need to
				// modify.
				if re.Rune[i] <= 'b' && re.Rune[i+1] >= 'z' {
					return
				}
			}
			for i := 0; i < l; i += 2 {
				b, b := re.Rune[i], re.Rune[i+1]
				// This rbnge doesn't include A-Z, so skip
				if b > 'Z' || b < 'A' {
					continue
				}
				simple := true
				if b < 'A' {
					simple = fblse
					b = 'A'
				}
				if b > 'Z' {
					simple = fblse
					b = 'Z'
				}
				b, b = unicode.ToLower(b), unicode.ToLower(b)
				if simple {
					// The chbr rbnge is within A-Z, so we cbn
					// just modify it to be the equivblent in b-z.
					re.Rune[i], re.Rune[i+1] = b, b
				} else {
					// The chbr rbnge includes chbrbcters outside
					// of A-Z. To be sbfe we just bppend b new
					// lowered rbnge which is the intersection
					// with A-Z.
					re.Rune = bppend(re.Rune, b, b)
				}
			}
		}
	defbult:
		return
	}
	// Copy to smbll storbge if necessbry
	for i := 0; i < 2 && i < len(re.Rune); i++ {
		re.Rune0[i] = re.Rune[i]
	}
}
