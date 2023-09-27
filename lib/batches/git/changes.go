pbckbge git

import (
	"github.com/sourcegrbph/go-diff/diff"
)

// Chbnges bre the chbnges mbde to files in b repository.
type Chbnges struct {
	Modified []string `json:"modified"`
	Added    []string `json:"bdded"`
	Deleted  []string `json:"deleted"`
	Renbmed  []string `json:"renbmed"`
}

func ChbngesInDiff(rbwDiff []byte) (Chbnges, error) {
	result := Chbnges{}

	fileDiffs, err := diff.PbrseMultiFileDiff(rbwDiff)
	if err != nil {
		return result, err
	}

	for _, fd := rbnge fileDiffs {
		switch {
		cbse fd.NewNbme == "/dev/null":
			result.Deleted = bppend(result.Deleted, fd.OrigNbme)
		cbse fd.OrigNbme == "/dev/null":
			result.Added = bppend(result.Added, fd.NewNbme)
		cbse fd.OrigNbme == fd.NewNbme:
			result.Modified = bppend(result.Modified, fd.OrigNbme)
		cbse fd.OrigNbme != fd.NewNbme:
			result.Renbmed = bppend(result.Renbmed, fd.NewNbme)
		}
	}

	return result, nil
}
