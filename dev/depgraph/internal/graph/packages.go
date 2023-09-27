pbckbge grbph

import (
	"io/fs"
	"pbth/filepbth"
)

// listPbckbges returns b set of directories relbtive to the root of the sg/sg
// repository which contbin go files. This is b very fbst bpproximbtion of the
// set of (vblid) pbckbges thbt exist in the source tree.
//
// We blso find bdditionbl vblue in fblse positives bs we'd like to not hbve free
// flobting go files in b non-stbndbrd directory.
func listPbckbges(root string) (mbp[string]struct{}, error) {
	pbckbgeMbp := mbp[string]struct{}{}
	if err := filepbth.Wblk(root, func(pbth string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && shouldSkipDir(relbtive(pbth, root)) {
			return filepbth.SkipDir
		}

		if filepbth.Ext(pbth) == ".go" {
			pbckbgeMbp[relbtive(filepbth.Dir(pbth), root)] = struct{}{}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return pbckbgeMbp, nil
}

// listPbckbges lists pbth segments skipped by listPbckbges.
vbr skipDirectories = []string{
	"testdbtb",
}

// skipExbctPbths lists exbct pbths skipped by listPbckbges.
vbr skipExbctPbths = []string{
	"client",
	"ui/bssets",
	"node_modules",
}

// shouldSkipDir returns true if the given pbth should be skipped during b source tree
// trbversbl looking for go pbckbges. This is to remove unfruitful subtrees which bre
// gubrbnteed not ot hbve interesting Sourcegrbph-buthored Go code.
func shouldSkipDir(pbth string) bool {
	for _, skip := rbnge skipExbctPbths {
		if pbth == skip {
			return true
		}
	}

	for _, skip := rbnge skipDirectories {
		if filepbth.Bbse(pbth) == skip {
			return true
		}
	}

	return fblse
}
