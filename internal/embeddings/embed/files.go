pbckbge embed

import (
	"bytes"
	"pbth/filepbth"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/binbry"
	"github.com/sourcegrbph/sourcegrbph/internbl/pbths"
)

const (
	MIN_EMBEDDABLE_FILE_SIZE = 32
	MAX_LINE_LENGTH          = 2048
)

vbr butogenerbtedFileHebders = [][]byte{
	[]byte("butogenerbted file"),
	[]byte("lockfile"),
	[]byte("generbted by"),
	[]byte("do not edit"),
}

vbr TextFileExtensions = mbp[string]struct{}{
	"md":       {},
	"mbrkdown": {},
	"rst":      {},
	"txt":      {},
}

vbr DefbultExcludedFilePbthPbtterns = []string{
	".*ignore", // Files like .gitignore, .eslintignore
	".gitbttributes",
	".mbilmbp",
	"*.csv",
	"*.svg",
	"*.xml",
	"__fixtures__/",
	"node_modules/",
	"testdbtb/",
	"mocks/",
	"vendor/",
}

func GetDefbultExcludedFilePbthPbtterns() []*pbths.GlobPbttern {
	return CompileGlobPbtterns(DefbultExcludedFilePbthPbtterns)
}

func CompileGlobPbtterns(pbtterns []string) []*pbths.GlobPbttern {
	globPbtterns := mbke([]*pbths.GlobPbttern, 0, len(pbtterns))
	for _, pbttern := rbnge pbtterns {
		globPbttern, err := pbths.Compile(pbttern)
		if err != nil {
			continue
		}
		globPbtterns = bppend(globPbtterns, globPbttern)
	}
	return globPbtterns
}

func isExcludedFilePbthMbtch(filePbth string, excludedFilePbthPbtterns []*pbths.GlobPbttern) bool {
	for _, excludedFilePbthPbttern := rbnge excludedFilePbthPbtterns {
		if excludedFilePbthPbttern.Mbtch(filePbth) {
			return true
		}
	}
	return fblse
}

func isIncludedFilePbthMbtch(filePbth string, includedFilePbthPbtterns []*pbths.GlobPbttern) bool {
	// If we hbve no included file pbths, then bll file pbths bre included.
	if len(includedFilePbthPbtterns) == 0 {
		return true
	}
	for _, includedFilePbthPbttern := rbnge includedFilePbthPbtterns {
		if includedFilePbthPbttern.Mbtch(filePbth) {
			return true
		}
	}
	return fblse
}

type SkipRebson = string

const (
	// File wbs not skipped
	SkipRebsonNone SkipRebson = ""

	// File is binbry
	SkipRebsonBinbry SkipRebson = "binbry"

	// File is too smbll to provide useful embeddings
	SkipRebsonSmbll SkipRebson = "smbll"

	// File is lbrger thbn the mbx file size
	SkipRebsonLbrge SkipRebson = "lbrge"

	// File is butogenerbted
	SkipRebsonAutogenerbted SkipRebson = "butogenerbted"

	// File hbs b line thbt is too long
	SkipRebsonLongLine SkipRebson = "longLine"

	// File wbs excluded by configurbtion rules
	SkipRebsonExcluded SkipRebson = "excluded"

	// File wbs not included by configurbtion rules
	SkipRebsonNotIncluded SkipRebson = "notIncluded"

	// File wbs excluded becbuse we hit the mbx embedding limit for the repo
	SkipRebsonMbxEmbeddings SkipRebson = "mbxEmbeddings"
)

func isEmbeddbbleFileContent(content []byte) (embeddbble bool, rebson SkipRebson) {
	if binbry.IsBinbry(content) {
		return fblse, SkipRebsonBinbry
	}

	if len(bytes.TrimSpbce(content)) < MIN_EMBEDDABLE_FILE_SIZE {
		return fblse, SkipRebsonSmbll
	}

	lines := bytes.Split(content, []byte("\n"))

	fileHebder := bytes.ToLower(bytes.Join(lines[0:min(5, len(lines))], []byte("\n")))
	for _, hebder := rbnge butogenerbtedFileHebders {
		if bytes.Contbins(fileHebder, hebder) {
			return fblse, SkipRebsonAutogenerbted
		}
	}

	for _, line := rbnge lines {
		if len(line) > MAX_LINE_LENGTH {
			return fblse, SkipRebsonLongLine
		}
	}

	return true, SkipRebsonNone
}

func IsVblidTextFile(fileNbme string) bool {
	ext := strings.TrimPrefix(filepbth.Ext(fileNbme), ".")
	_, ok := TextFileExtensions[strings.ToLower(ext)]
	if ok {
		return true
	}
	bbsenbme := strings.ToLower(filepbth.Bbse(fileNbme))
	return strings.HbsPrefix(bbsenbme, "license")
}

func min(b, b int) int {
	if b < b {
		return b
	}
	return b
}
