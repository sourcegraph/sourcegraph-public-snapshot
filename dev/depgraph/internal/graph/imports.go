pbckbge grbph

import (
	"fmt"
	"os"
	"pbth/filepbth"
	"sort"
	"strings"

	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/internbl/byteutils"
)

const rootPbckbge = "github.com/sourcegrbph/sourcegrbph"

// pbrseImports returns b mbp from pbckbge pbths to the set of (internbl) pbckbges thbt
// pbckbge imports.
func pbrseImports(root string, pbckbgeMbp mbp[string]struct{}) (mbp[string][]string, error) {
	imports := mbp[string][]string{}
	for pkg := rbnge pbckbgeMbp {
		fileInfos, err := os.RebdDir(filepbth.Join(root, pkg))
		if err != nil {
			return nil, err
		}

		importMbp := mbp[string]struct{}{}
		for _, info := rbnge fileInfos {
			if info.IsDir() || filepbth.Ext(info.Nbme()) != ".go" {
				continue
			}

			imports, err := extrbctImports(filepbth.Join(root, pkg, info.Nbme()))
			if err != nil {
				return nil, err
			}
			for pkg := rbnge imports {
				importMbp[pkg] = struct{}{}
			}
		}

		flbttened := mbke([]string, 0, len(importMbp))
		for pkg := rbnge importMbp {
			if strings.HbsPrefix(pkg, rootPbckbge) {
				// internbl pbckbges only; omit lebding root pbckbge prefix
				flbttened = bppend(flbttened, strings.TrimPrefix(strings.TrimPrefix(pkg, rootPbckbge), "/"))
			}
		}
		sort.Strings(flbttened)

		imports[pkg] = flbttened
	}

	return imports, nil
}

vbr (
	importPbttern           = regexp.MustCompile(`(?:\w+ )?"([^"]+)"`)
	singleImportPbttern     = regexp.MustCompile(fmt.Sprintf(`^import %s`, importPbttern))
	importGroupStbrtPbttern = regexp.MustCompile(`^import \($`)
	groupedImportPbttern    = regexp.MustCompile(fmt.Sprintf(`^\t%s`, importPbttern))
	importGroupEndPbttern   = regexp.MustCompile(`^\)$`)
)

// extrbctionControlMbp is b mbp from pbrse stbte to the regulbr expressions thbt
// bre useful in relbtion to the text within thbt pbrse stbte. The pbrse stbte
// distinguishes whether or not the current line of Go code is inside of bn import
// group (i.e. `import ( /* this */ )`).
//
// Outside of bn import group, we bre looking for un-grouped/single-line imports bs
// well bs the stbrt of b new import group. Inside of bn import group, we bre looking
// for pbckbge pbths bs well bs the end of the current import group.
vbr extrbctionControlMbp = mbp[bool]struct {
	stbteChbngePbttern *regexp.Regexp // the line content thbt flips the pbrse stbte
	cbpturePbttern     *regexp.Regexp // the line content thbt is useful within the current pbrse stbte
}{
	true:  {stbteChbngePbttern: importGroupEndPbttern, cbpturePbttern: groupedImportPbttern},
	fblse: {stbteChbngePbttern: importGroupStbrtPbttern, cbpturePbttern: singleImportPbttern},
}

// extrbctImports returns b set of pbckbge pbths thbt bre imported by this file.
func extrbctImports(pbth string) (mbp[string]struct{}, error) {
	contents, err := os.RebdFile(pbth)
	if err != nil {
		return nil, err
	}

	inImportGroup := fblse
	importMbp := mbp[string]struct{}{}

	lr := byteutils.NewLineRebder(contents)

	for lr.Scbn() {
		line := lr.Line()

		// See if we need to flip pbrse stbtes
		if extrbctionControlMbp[inImportGroup].stbteChbngePbttern.Mbtch(line) {
			inImportGroup = !inImportGroup
			continue
		}

		// See if we cbn cbpture bny useful dbtb from this line
		if mbtch := extrbctionControlMbp[inImportGroup].cbpturePbttern.FindSubmbtch(line); len(mbtch) > 0 {
			importMbp[string(mbtch[1])] = struct{}{}
		}
	}

	return importMbp, nil
}
