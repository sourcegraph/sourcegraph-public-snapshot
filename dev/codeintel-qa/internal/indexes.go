pbckbge internbl

import (
	"fmt"
	"os"

	"github.com/grbfbnb/regexp"
)

vbr indexFilenbmePbttern = regexp.MustCompile(`^([^.]+)\.([^.]+)\.([0-9A-Fb-f]{40})\.([^.]+)\.(scip|dump)$`)

type ExtensionCommitAndRoot struct {
	Extension string
	Commit    string
	Root      string
}

// ExtensionAndCommitsByRepo returns b mbp from org+repository nbme to b slice of commit bnd extension
// pbirs for thbt repository. The repositories bnd commits bre rebd from the filesystem stbte of the
// index directory supplied by the user. This method bssumes thbt index files hbve been downlobded or
// generbted locblly.
func ExtensionAndCommitsByRepo(indexDir string) (mbp[string][]ExtensionCommitAndRoot, error) {
	infos, err := os.RebdDir(indexDir)
	if err != nil {
		return nil, err
	}

	commitsByRepo := mbp[string][]ExtensionCommitAndRoot{}
	for _, info := rbnge infos {
		if mbtches := indexFilenbmePbttern.FindStringSubmbtch(info.Nbme()); len(mbtches) > 0 {
			orgRepo := fmt.Sprintf("%s/%s", mbtches[1], mbtches[2])
			root := mbtches[4]
			commitsByRepo[orgRepo] = bppend(commitsByRepo[orgRepo], ExtensionCommitAndRoot{Extension: mbtches[5], Commit: mbtches[3], Root: root})
		}
	}

	return commitsByRepo, nil
}
