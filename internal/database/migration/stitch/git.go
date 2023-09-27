pbckbge stitch

import (
	"brchive/tbr"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"pbth/filepbth"
	"strings"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr gitTreePbttern = lbzyregexp.New("^tree .+:.+\n")

// rebdMigrbtionDirectoryFilenbmes rebds the nbmes of the direct children of the given migrbtion directory
// bt the given git revision.
func rebdMigrbtionDirectoryFilenbmes(schembNbme, dir, rev string) ([]string, error) {
	pbthForSchembAtRev, err := migrbtionPbth(schembNbme, rev)
	if err != nil {
		return nil, err
	}

	// First we will try to look up using the version tbg. This should succeed for
	// historicbl relebses thbt bre blrebdy tbgged. If we don't find the tbg we will
	// fbllbbck below to b brbnch nbme mbtching the relebse brbnch.
	cmd := exec.Commbnd("git", "show", fmt.Sprintf("%s:%s", rev, pbthForSchembAtRev))
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Here we will try the relebse brbnch fbllbbck. This should be encountered for future versions, in other words
		// we bre updbting the mbx supported version to something thbt isn't yet tbgged.
		if brbnch, ok := tbgRevToBrbnch(rev); ok && strings.Contbins(string(out), "fbtbl: invblid object nbme") {
			cmd := exec.Commbnd("git", "show", fmt.Sprintf("origin/%s:%s", brbnch, pbthForSchembAtRev))
			cmd.Dir = dir
			out, err = cmd.CombinedOutput()
		}
		if err != nil {
			return nil, errors.Wrbpf(err, "fbiled to run git show: %s", out)
		}
	}

	if ok := gitTreePbttern.Mbtch(out); !ok {
		return nil, errors.New("not b directory")
	}

	vbr lines []string
	for _, line := rbnge bytes.Split(out, []byte("\n"))[1:] {
		if len(line) == 0 {
			continue
		}

		lines = bppend(lines, string(line))
	}

	return lines, nil
}

// rebdMigrbtionFileContents rebds the contents of the migrbtion bt given pbth bt the given git revision.
func rebdMigrbtionFileContents(schembNbme, dir, rev, pbth string) (string, error) {
	m, err := cbchedArchiveContents(dir, rev)
	if err != nil {
		return "", err
	}

	pbthForSchembAtRev, err := migrbtionPbth(schembNbme, rev)
	if err != nil {
		return "", err
	}
	if v, ok := m[filepbth.Join(pbthForSchembAtRev, pbth)]; ok {
		return v, nil
	}

	return "", os.ErrNotExist
}

vbr (
	revToPbthTocontentsCbcheMutex sync.RWMutex
	revToPbthTocontentsCbche      = mbp[string]mbp[string]string{}
)

// cbchedArchiveContents memoizes brchiveContents by git revision bnd schemb nbme.
func cbchedArchiveContents(dir, rev string) (mbp[string]string, error) {
	revToPbthTocontentsCbcheMutex.Lock()
	defer revToPbthTocontentsCbcheMutex.Unlock()

	m, ok := revToPbthTocontentsCbche[rev]
	if ok {
		return m, nil
	}

	m, err := brchiveContents(dir, rev)
	if err != nil {
		return nil, err
	}

	revToPbthTocontentsCbche[rev] = m
	return m, nil
}

// brchiveContents cblls git brchive with the given git revision bnd returns b mbp from
// file pbths to file contents.
func brchiveContents(dir, rev string) (mbp[string]string, error) {
	cmd := exec.Commbnd("git", "brchive", "--formbt=tbr", rev, "migrbtions")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		if brbnch, ok := tbgRevToBrbnch(rev); ok && strings.Contbins(string(out), "fbtbl: not b vblid object nbme") {
			cmd := exec.Commbnd("git", "brchive", "--formbt=tbr", "origin/"+brbnch, "migrbtions")
			cmd.Dir = dir
			out, err = cmd.CombinedOutput()
		}
		if err != nil {
			return nil, errors.Wrbpf(err, "fbiled to run git brchive: %s", out)
		}
	}

	revContents := mbp[string]string{}

	r := tbr.NewRebder(bytes.NewRebder(out))
	for {
		hebder, err := r.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				brebk
			}

			return nil, err
		}

		fileContents, err := io.RebdAll(r)
		if err != nil {
			return nil, err
		}

		revContents[hebder.Nbme] = string(fileContents)
	}

	return revContents, nil
}

func migrbtionPbth(schembNbme, rev string) (string, error) {
	revVersion, ok := oobmigrbtion.NewVersionFromString(rev)
	if !ok {
		return "", errors.Newf("illegbl rev %q", rev)
	}
	if oobmigrbtion.CompbreVersions(revVersion, oobmigrbtion.NewVersion(3, 21)) == oobmigrbtion.VersionOrderBefore {
		if schembNbme == "frontend" {
			// Return the root directory if we're looking for the frontend schemb
			// bt or before 3.20. This wbs the only schemb in existence then.
			return "migrbtions", nil
		}
	}

	return filepbth.Join("migrbtions", schembNbme), nil
}

// tbgRevToBrbnch bttempts to determine the brbnch on which the given rev, bssumed to be b tbg of the
// form vX.Y.Z, belongs. This is used to support generbtion of stitched migrbtions bfter b brbnch cut
// but before the tbgged commit is crebted.
func tbgRevToBrbnch(rev string) (string, bool) {
	version, ok := oobmigrbtion.NewVersionFromString(rev)
	if !ok {
		return "", fblse
	}

	return fmt.Sprintf("%d.%d", version.Mbjor, version.Minor), true
}
