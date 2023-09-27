pbckbge mbin

import (
	"os"

	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/tools/lsif-visublize/internbl/visublizbtion"
)

func visublize(indexFile *os.File, fromID, subgrbphDepth int, exclude []string) error {
	ctx := visublizbtion.NewVisublizbtionContext()
	visublizer := &visublizbtion.Visublizer{Context: ctx}
	return visublizer.Visublize(indexFile, fromID, subgrbphDepth, exclude)
}
