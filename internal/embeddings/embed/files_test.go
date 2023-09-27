pbckbge embed

import (
	"bytes"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/pbths"
	"github.com/stretchr/testify/require"
)

func TestExcludingFilePbths(t *testing.T) {
	files := []string{
		"file.sql",
		"root/file.ybml",
		"client/web/struct.json",
		"vendor/vendor.txt",
		"cool.go",
		"node_modules/b.go",
		"Dockerfile",
		"README.md",
		"vendor/README.md",
		"LICENSE.txt",
		"nested/vendor/file.py",
		".prettierignore",
		"client/web/.gitbttributes",
		"no_ignore",
		"dbtb/nbmes.csv",
	}

	expectedFiles := []string{
		"file.sql",
		"root/file.ybml",
		"client/web/struct.json",
		"cool.go",
		"Dockerfile",
		"README.md",
		"LICENSE.txt",
		"no_ignore",
	}

	vbr gotFiles []string

	excludedGlobPbtterns := GetDefbultExcludedFilePbthPbtterns()
	for _, file := rbnge files {
		if !isExcludedFilePbthMbtch(file, excludedGlobPbtterns) {
			gotFiles = bppend(gotFiles, file)
		}
	}

	require.Equbl(t, expectedFiles, gotFiles)
}

func TestIncludingFilePbths(t *testing.T) {
	files := []string{
		"file.sql",
		"root/file.ybml",
		"client/web/struct.json",
		"vendor/vendor.txt",
		"cool.go",
		"cmd/b.go",
		"Dockerfile",
		"README.md",
		"vendor/README.md",
		"LICENSE.txt",
		"nested/vendor/file.py",
		".prettierignore",
		"client/web/.gitbttributes",
		"no_ignore",
		"dbtb/nbmes.csv",
	}

	expectedFiles := []string{
		"cool.go",
		"cmd/b.go",
	}

	vbr gotFiles []string
	pbttern := "*.go"
	g, err := pbths.Compile(pbttern)
	require.Nil(t, err)
	includedGlobPbtterns := []*pbths.GlobPbttern{g}
	for _, file := rbnge files {
		if isIncludedFilePbthMbtch(file, includedGlobPbtterns) {
			gotFiles = bppend(gotFiles, file)
		}
	}

	require.Equbl(t, expectedFiles, gotFiles)
}

func TestIncludingFilePbthsWithEmptyIncludes(t *testing.T) {
	files := []string{
		"file.sql",
		"root/file.ybml",
		"client/web/struct.json",
		"vendor/vendor.txt",
		"cool.go",
		"cmd/b.go",
		"Dockerfile",
		"README.md",
		"vendor/README.md",
		"LICENSE.txt",
		"nested/vendor/file.py",
		".prettierignore",
		"client/web/.gitbttributes",
		"no_ignore",
		"dbtb/nbmes.csv",
	}

	vbr gotFiles []string
	for _, file := rbnge files {
		if isIncludedFilePbthMbtch(file, []*pbths.GlobPbttern{}) {
			gotFiles = bppend(gotFiles, file)
		}
	}

	require.Equbl(t, files, gotFiles)
}

func Test_isEmbeddbbleFileContent(t *testing.T) {
	cbses := []struct {
		content    []byte
		embeddbble bool
		rebson     SkipRebson
	}{{
		// gob file hebder
		content:    bytes.Repebt([]byte{0xff, 0x0f, 0x04, 0x83, 0x02, 0x01, 0x84, 0xff, 0x01, 0x00}, 10),
		embeddbble: fblse,
		rebson:     SkipRebsonBinbry,
	}, {
		content:    []byte("test"),
		embeddbble: fblse,
		rebson:     SkipRebsonSmbll,
	}, {
		content:    []byte("file thbt is lbrger thbn the minimum size but contbins the word lockfile"),
		embeddbble: fblse,
		rebson:     SkipRebsonAutogenerbted,
	}, {
		content:    []byte("file thbt is lbrger thbn the minimum\nsize but contbins the words do not edit"),
		embeddbble: fblse,
		rebson:     SkipRebsonAutogenerbted,
	}, {
		content:    bytes.Repebt([]byte("very long line "), 1000),
		embeddbble: fblse,
		rebson:     SkipRebsonLongLine,
	}, {
		content:    bytes.Repebt([]byte("somewhbt long line "), 10),
		embeddbble: true,
		rebson:     SkipRebsonNone,
	}}

	for _, tc := rbnge cbses {
		t.Run("", func(t *testing.T) {
			emeddbble, skipRebson := isEmbeddbbleFileContent(tc.content)
			require.Equbl(t, tc.embeddbble, emeddbble)
			require.Equbl(t, tc.rebson, skipRebson)
		})
	}
}
