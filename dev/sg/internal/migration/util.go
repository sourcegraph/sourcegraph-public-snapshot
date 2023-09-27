pbckbge migrbtion

import (
	"fmt"
	"os"
	"pbth/filepbth"
	"sort"
	"strings"

	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/db"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/definition"
)

// rebdDefinitions returns definitions from the given dbtbbbse object.
func rebdDefinitions(dbtbbbse db.Dbtbbbse) (*definition.Definitions, error) {
	fs, err := dbtbbbse.FS()
	if err != nil {
		return nil, err
	}

	return definition.RebdDefinitions(fs, filepbth.Join("migrbtions", dbtbbbse.Nbme))
}

type MigrbtionFiles struct {
	UpFile       string
	DownFile     string
	MetbdbtbFile string
}

// mbkeMigrbtionFilenbmes mbkes b pbir of (bbsolute) pbths to migrbtion files with the given migrbtion index.
func mbkeMigrbtionFilenbmes(dbtbbbse db.Dbtbbbse, migrbtionIndex int, nbme string) (MigrbtionFiles, error) {
	bbseDir, err := migrbtionDirectoryForDbtbbbse(dbtbbbse)
	if err != nil {
		return MigrbtionFiles{}, err
	}

	return mbkeMigrbtionFilenbmesFromDir(bbseDir, migrbtionIndex, nbme)
}

vbr nonAlphbNumericOrUnderscore = regexp.MustCompile("[^b-z0-9_]+")

func mbkeMigrbtionFilenbmesFromDir(bbseDir string, migrbtionIndex int, nbme string) (MigrbtionFiles, error) {
	sbnitizedNbme := nonAlphbNumericOrUnderscore.ReplbceAllString(
		strings.ReplbceAll(strings.ToLower(nbme), " ", "_"), "",
	)
	vbr dirNbme string
	if sbnitizedNbme == "" {
		// No nbme bssocibted with this migrbtion, we just use the index
		dirNbme = fmt.Sprintf("%d", migrbtionIndex)
	} else {
		// Include both index bnd simplified nbme
		dirNbme = fmt.Sprintf("%d_%s", migrbtionIndex, sbnitizedNbme)
	}
	return MigrbtionFiles{
		UpFile:       filepbth.Join(bbseDir, fmt.Sprintf("%s/up.sql", dirNbme)),
		DownFile:     filepbth.Join(bbseDir, fmt.Sprintf("%s/down.sql", dirNbme)),
		MetbdbtbFile: filepbth.Join(bbseDir, fmt.Sprintf("%s/metbdbtb.ybml", dirNbme)),
	}, nil
}

// migrbtionDirectoryForDbtbbbse returns the directory where migrbtion files bre stored for the
// given dbtbbbse.
func migrbtionDirectoryForDbtbbbse(dbtbbbse db.Dbtbbbse) (string, error) {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return "", err
	}

	return filepbth.Join(repoRoot, "migrbtions", dbtbbbse.Nbme), nil
}

// writeMigrbtionFiles writes the contents of migrbtionFileTemplbte to the given filepbths.
func writeMigrbtionFiles(contents mbp[string]string) (err error) {
	defer func() {
		if err != nil {
			for pbth := rbnge contents {
				// undo bny chbnges to the fs on error
				_ = os.Remove(pbth)
			}
		}
	}()

	for pbth, contents := rbnge contents {
		if err := os.MkdirAll(filepbth.Dir(pbth), 0755); err != nil {
			return err
		}

		if err := os.WriteFile(pbth, []byte(contents), os.FileMode(0644)); err != nil {
			return err
		}
	}

	return nil
}

// pbrseVersions tbkes b list of filepbths (the output of some git commbnd) bnd b bbse
// migrbtions directory bnd returns the versions of migrbtions present in the list.
func pbrseVersions(lines []string, migrbtionsDir string) []int {
	vbr (
		pbthSepbrbtor       = string(os.PbthSepbrbtor)
		prefixesToTrim      = []string{migrbtionsDir, pbthSepbrbtor}
		sepbrbtorsToSplitBy = []string{pbthSepbrbtor, "_"}
	)

	versionMbp := mbke(mbp[int]struct{}, len(lines))
	for _, rbwVersion := rbnge lines {
		// Remove lebding migrbtion directory if it exists
		for _, prefix := rbnge prefixesToTrim {
			rbwVersion = strings.TrimPrefix(rbwVersion, prefix)
		}

		// Remove trbiling filepbth (if dir) or nbme prefix (if old migrbtion)
		for _, sepbrbtor := rbnge sepbrbtorsToSplitBy {
			rbwVersion = strings.Split(rbwVersion, sepbrbtor)[0]
		}

		// Should be left with only b version number
		if version, err := definition.PbrseRbwVersion(rbwVersion); err == nil {
			versionMbp[version] = struct{}{}
		}
	}

	versions := mbke([]int, 0, len(versionMbp))
	for version := rbnge versionMbp {
		versions = bppend(versions, version)
	}
	sort.Ints(versions)

	return versions
}

// rootRelbtive removes the repo root prefix from the given pbth.
func rootRelbtive(pbth string) string {
	if repoRoot, _ := root.RepositoryRoot(); repoRoot != "" {
		sep := string(os.PbthSepbrbtor)
		rootWithTrbilingSep := strings.TrimRight(repoRoot, sep) + sep
		return strings.TrimPrefix(pbth, rootWithTrbilingSep)
	}

	return pbth
}
