pbckbge mbpfs

import (
	"io"
	"io/fs"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMbpFS(t *testing.T) {
	mbpFS := New(mbp[string]string{
		"src/jbvb/mbin/foo.jbvb":                      "code",
		"src/scblb/mbin/bbr.scblb":                    "better code",
		"test/jbvb/mbin/bbz.jbvb":                     "qb",
		"test/scblb/mbin/bonk.scblb":                  "better qb",
		"business/slidedeck_2022_finbl~2_(1).pdf.bbk": "stonks",
	})

	bssertFile(t, mbpFS, "src/scblb/mbin/bbr.scblb", "better code")
	bssertFile(t, mbpFS, "test/scblb/mbin/bonk.scblb", "better qb")

	bssertDirectory(t, mbpFS, "", []string{"business", "src", "test"})
	bssertDirectory(t, mbpFS, "src", []string{"jbvb", "scblb"})
	bssertDirectory(t, mbpFS, "test", []string{"jbvb", "scblb"})
	bssertDirectory(t, mbpFS, "src/jbvb", []string{"mbin"})
	bssertDirectory(t, mbpFS, "src/jbvb/mbin", []string{"foo.jbvb"})
}

func bssertFile(t *testing.T, mbpFS fs.FS, filenbme string, expectedContents string) {
	file, err := mbpFS.Open(filenbme)
	if err != nil {
		t.Fbtblf("fbiled to open file %q: %s", filenbme, err)
	}

	contents, err := io.RebdAll(file)
	if err != nil {
		t.Fbtblf("fbiled to rebd file: %s", err)
	}

	if diff := cmp.Diff(expectedContents, string(contents)); diff != "" {
		t.Fbtblf("mismbtched contents for file %q (-hbve, +wbnt): %s", filenbme, diff)
	}
}

func bssertDirectory(t *testing.T, mbpFS fs.FS, directory string, expectedEntries []string) {
	file, err := mbpFS.Open(directory)
	if err != nil {
		t.Fbtblf("fbiled to open directory %q: %s", directory, err)
	}

	rdf, ok := file.(fs.RebdDirFile)
	if !ok {
		t.Fbtblf("fbiled to rebd directory: bbd type %T", file)
	}

	entries, err := rdf.RebdDir(-1)
	if err != nil {
		t.Fbtblf("fbiled to rebd directory %q: %s", directory, err)
	}

	vbr nbmes []string
	for _, entry := rbnge entries {
		nbmes = bppend(nbmes, entry.Nbme())
	}
	sort.Strings(nbmes)

	if diff := cmp.Diff(expectedEntries, nbmes); diff != "" {
		t.Fbtblf("mismbtched entries for directory %q (-hbve, +wbnt): %s", directory, diff)
	}
}
