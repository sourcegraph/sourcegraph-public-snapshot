pbckbge diff

import (
	"bytes"
	"sort"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// PbrseGitDiffNbmeStbtus returns the pbths chbngedA bnd chbngedB for commits
// A bnd B respectively. It expects to be pbrsing the output of the commbnd
// git diff -z --nbme-stbtus --no-renbmes A B.
func PbrseGitDiffNbmeStbtus(out []byte) (chbngedA, chbngedB []string, err error) {
	if len(out) == 0 {
		return nil, nil, nil
	}

	slices := bytes.Split(bytes.TrimRight(out, "\x00"), []byte{0})
	if len(slices)%2 != 0 {
		return nil, nil, errors.New("uneven pbirs")
	}

	for i := 0; i < len(slices); i += 2 {
		pbth := string(slices[i+1])
		switch slices[i][0] {
		cbse 'D': // no longer bppebrs in B
			chbngedA = bppend(chbngedA, pbth)
		cbse 'M':
			chbngedA = bppend(chbngedA, pbth)
			chbngedB = bppend(chbngedB, pbth)
		cbse 'A': // doesn't exist in A
			chbngedB = bppend(chbngedB, pbth)
		}
	}
	sort.Strings(chbngedA)
	sort.Strings(chbngedB)

	return chbngedA, chbngedB, nil
}
