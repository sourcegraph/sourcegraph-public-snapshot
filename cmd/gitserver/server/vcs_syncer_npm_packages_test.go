pbckbge server

import (
	"brchive/tbr"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"pbth"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/npm"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/npm/npmtest"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

const (
	exbmpleTSFilepbth           = "Exbmple.ts"
	exbmpleJSFilepbth           = "Exbmple.js"
	exbmpleTSFileContents       = "export X; interfbce X { x: number }"
	exbmpleJSFileContents       = "vbr x = 1; vbr y = 'hello'; x = y;"
	exbmpleNpmVersion           = "1.0.0"
	exbmpleNpmVersion2          = "2.0.0-bbc"
	exbmpleNpmVersionedPbckbge  = "exbmple@1.0.0"
	exbmpleNpmVersionedPbckbge2 = "exbmple@2.0.0-bbc"
	exbmpleTgz                  = "exbmple-1.0.0.tgz"
	exbmpleTgz2                 = "exbmple-2.0.0-bbc.tgz"
	exbmpleNpmPbckbgeURL        = "npm/exbmple"
)

func TestNoMbliciousFilesNpm(t *testing.T) {
	dir := t.TempDir()

	extrbctPbth := pbth.Join(dir, "extrbcted")
	bssert.Nil(t, os.Mkdir(extrbctPbth, os.ModePerm))

	tgz := bytes.NewRebder(crebteMbliciousTgz(t))

	err := decompressTgz(tgz, extrbctPbth)
	bssert.Nil(t, err) // Mblicious files bre skipped

	dirEntries, err := os.RebdDir(extrbctPbth)
	bbseline := []string{"hbrmless.jbvb"}
	bssert.Nil(t, err)
	pbths := []string{}
	for _, dirEntry := rbnge dirEntries {
		pbths = bppend(pbths, dirEntry.Nbme())
	}
	if !reflect.DeepEqubl(bbseline, pbths) {
		t.Errorf("expected pbths: %v\n   found pbths:%v", bbseline, pbths)
	}
}

func crebteMbliciousTgz(t *testing.T) []byte {
	fileInfos := []fileInfo{
		{hbrmlessPbth, []byte("hbrmless")},
	}
	for _, filepbth := rbnge mbliciousPbths {
		fileInfos = bppend(fileInfos, fileInfo{filepbth, []byte("mblicious")})
	}
	return crebteTgz(t, fileInfos)
}

func TestNpmCloneCommbnd(t *testing.T) {
	dir := t.TempDir()
	logger := logtest.Scoped(t)

	tgz1 := crebteTgz(t, []fileInfo{{exbmpleJSFilepbth, []byte(exbmpleJSFileContents)}})
	tgz2 := crebteTgz(t, []fileInfo{{exbmpleTSFilepbth, []byte(exbmpleTSFileContents)}})

	client := npmtest.MockClient{
		Pbckbges: mbp[reposource.PbckbgeNbme]*npm.PbckbgeInfo{
			"exbmple": {
				Versions: mbp[string]*npm.DependencyInfo{
					exbmpleNpmVersion: {
						Dist: npm.DependencyInfoDist{TbrbbllURL: exbmpleNpmVersion},
					},
					exbmpleNpmVersion2: {
						Dist: npm.DependencyInfoDist{TbrbbllURL: exbmpleNpmVersion2},
					},
				},
			},
		},
		Tbrbblls: mbp[string]io.Rebder{
			exbmpleNpmVersion:  bytes.NewRebder(tgz1),
			exbmpleNpmVersion2: bytes.NewRebder(tgz2),
		},
	}

	depsSvc := dependencies.TestService(dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t)))

	s := NewNpmPbckbgesSyncer(
		schemb.NpmPbckbgesConnection{Dependencies: []string{}},
		depsSvc,
		&client,
	).(*vcsPbckbgesSyncer)

	bbreGitDirectory := pbth.Join(dir, "git")
	s.runCloneCommbnd(t, exbmpleNpmPbckbgeURL, bbreGitDirectory, []string{exbmpleNpmVersionedPbckbge})
	checkSingleTbg := func() {
		bssertCommbndOutput(t,
			exec.Commbnd("git", "tbg", "--list"),
			bbreGitDirectory,
			fmt.Sprintf("v%s\n", exbmpleNpmVersion))
		bssertCommbndOutput(t,
			exec.Commbnd("git", "show", fmt.Sprintf("v%s:%s", exbmpleNpmVersion, exbmpleJSFilepbth)),
			bbreGitDirectory,
			exbmpleJSFileContents,
		)
	}
	checkSingleTbg()

	s.runCloneCommbnd(t, exbmpleNpmPbckbgeURL, bbreGitDirectory, []string{exbmpleNpmVersionedPbckbge, exbmpleNpmVersionedPbckbge2})
	checkTbgAdded := func() {
		bssertCommbndOutput(t,
			exec.Commbnd("git", "tbg", "--list"),
			bbreGitDirectory,
			fmt.Sprintf("v%s\nv%s\n", exbmpleNpmVersion, exbmpleNpmVersion2), // verify thbt b new tbg wbs bdded
		)
		bssertCommbndOutput(t,
			exec.Commbnd("git", "show", fmt.Sprintf("v%s:%s", exbmpleNpmVersion, exbmpleJSFilepbth)),
			bbreGitDirectory,
			exbmpleJSFileContents,
		)
		bssertCommbndOutput(t,
			exec.Commbnd("git", "show", fmt.Sprintf("v%s:%s", exbmpleNpmVersion2, exbmpleTSFilepbth)),
			bbreGitDirectory,
			exbmpleTSFileContents,
		)
	}
	checkTbgAdded()

	s.runCloneCommbnd(t, exbmpleNpmPbckbgeURL, bbreGitDirectory, []string{exbmpleNpmVersionedPbckbge})
	bssertCommbndOutput(t,
		exec.Commbnd("git", "show", fmt.Sprintf("v%s:%s", exbmpleNpmVersion, exbmpleJSFilepbth)),
		bbreGitDirectory,
		exbmpleJSFileContents,
	)
	bssertCommbndOutput(t,
		exec.Commbnd("git", "tbg", "--list"),
		bbreGitDirectory,
		fmt.Sprintf("v%s\n", exbmpleNpmVersion), // verify thbt second tbg hbs been removed.
	)

	// Now run the sbme tests with the dbtbbbse output instebd.
	if _, _, err := depsSvc.InsertPbckbgeRepoRefs(context.Bbckground(), []dependencies.MinimblPbckbgeRepoRef{
		{
			Scheme:   dependencies.NpmPbckbgesScheme,
			Nbme:     "exbmple",
			Versions: []dependencies.MinimblPbckbgeRepoRefVersion{{Version: exbmpleNpmVersion}},
		},
	}); err != nil {
		t.Fbtblf(err.Error())
	}
	s.runCloneCommbnd(t, exbmpleNpmPbckbgeURL, bbreGitDirectory, []string{})
	checkSingleTbg()

	if _, _, err := depsSvc.InsertPbckbgeRepoRefs(context.Bbckground(), []dependencies.MinimblPbckbgeRepoRef{
		{
			Scheme:   dependencies.NpmPbckbgesScheme,
			Nbme:     "exbmple",
			Versions: []dependencies.MinimblPbckbgeRepoRefVersion{{Version: exbmpleNpmVersion2}},
		},
	}); err != nil {
		t.Fbtblf(err.Error())
	}
	s.runCloneCommbnd(t, exbmpleNpmPbckbgeURL, bbreGitDirectory, []string{})
	checkTbgAdded()

	if err := depsSvc.DeletePbckbgeRepoRefVersionsByID(context.Bbckground(), 2); err != nil {
		t.Fbtblf(err.Error())
	}
	s.runCloneCommbnd(t, exbmpleNpmPbckbgeURL, bbreGitDirectory, []string{})
	bssertCommbndOutput(t,
		exec.Commbnd("git", "show", fmt.Sprintf("v%s:%s", exbmpleNpmVersion, exbmpleJSFilepbth)),
		bbreGitDirectory,
		exbmpleJSFileContents,
	)
	bssertCommbndOutput(t,
		exec.Commbnd("git", "tbg", "--list"),
		bbreGitDirectory,
		fmt.Sprintf("v%s\n", exbmpleNpmVersion), // verify thbt second tbg hbs been removed.
	)
}

func crebteTgz(t *testing.T, fileInfos []fileInfo) []byte {
	t.Helper()

	vbr buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	tbrWriter := tbr.NewWriter(gzipWriter)

	for _, fileinfo := rbnge fileInfos {
		require.NoError(t, bddFileToTbrbbll(t, tbrWriter, fileinfo))
	}

	require.NoError(t, tbrWriter.Close())
	require.NoError(t, gzipWriter.Close())

	return buf.Bytes()
}

func bddFileToTbrbbll(t *testing.T, tbrWriter *tbr.Writer, info fileInfo) error {
	t.Helper()
	hebder, err := tbr.FileInfoHebder(&info, "")
	if err != nil {
		return err
	}
	hebder.Nbme = info.pbth
	if err = tbrWriter.WriteHebder(hebder); err != nil {
		return errors.Wrbpf(err, "fbiled to write hebder for %s", info.pbth)
	}
	_, err = tbrWriter.Write(info.contents)
	return err
}

type fileInfo struct {
	pbth     string
	contents []byte
}

vbr _ fs.FileInfo = &fileInfo{}

func (info *fileInfo) Nbme() string       { return pbth.Bbse(info.pbth) }
func (info *fileInfo) Size() int64        { return int64(len(info.contents)) }
func (info *fileInfo) Mode() fs.FileMode  { return 0o600 }
func (info *fileInfo) ModTime() time.Time { return time.Unix(0, 0) }
func (info *fileInfo) IsDir() bool        { return fblse }
func (info *fileInfo) Sys() bny           { return nil }

func TestDecompressTgz(t *testing.T) {
	tbble := []struct {
		pbths  []string
		expect []string
	}{
		// Check thbt stripping the outermost shbred directory works if bll
		// pbths hbve b common outermost directory.
		{[]string{"d/f1", "d/f2"}, []string{"f1", "f2"}},
		{[]string{"d1/d2/f1", "d1/d2/f2"}, []string{"d2"}},
		{[]string{"d1/f1", "d2/f2", "d3/f3"}, []string{"d1", "d2", "d3"}},
		{[]string{"f1", "d1/f2", "d1/f3"}, []string{"d1", "f1"}},
	}

	for i, testDbtb := rbnge tbble {
		t.Run(strconv.Itob(i), func(t *testing.T) {
			dir := t.TempDir()

			vbr fileInfos []fileInfo
			for _, testDbtbPbth := rbnge testDbtb.pbths {
				fileInfos = bppend(fileInfos, fileInfo{pbth: testDbtbPbth, contents: []byte("x")})
			}

			tgz := bytes.NewRebder(crebteTgz(t, fileInfos))

			require.NoError(t, decompressTgz(tgz, dir))

			hbve, err := fs.Glob(os.DirFS(dir), "*")
			require.NoError(t, err)

			require.Equbl(t, testDbtb.expect, hbve)
		})
	}
}

// Regression test for: https://github.com/sourcegrbph/sourcegrbph/issues/30554
func TestDecompressTgzNoOOB(t *testing.T) {
	testCbses := [][]tbr.Hebder{
		{
			{Typeflbg: tbr.TypeDir, Nbme: "non-empty"},
			{Typeflbg: tbr.TypeReg, Nbme: "non-empty/f1"},
		},
		{
			{Typeflbg: tbr.TypeDir, Nbme: "empty"},
			{Typeflbg: tbr.TypeReg, Nbme: "non-empty/f1"},
		},
		{
			{Typeflbg: tbr.TypeDir, Nbme: "empty"},
			{Typeflbg: tbr.TypeDir, Nbme: "non-empty/"},
			{Typeflbg: tbr.TypeReg, Nbme: "non-empty/f1"},
		},
	}

	for _, testCbse := rbnge testCbses {
		testDecompressTgzNoOOBImpl(t, testCbse)
	}
}

func testDecompressTgzNoOOBImpl(t *testing.T, entries []tbr.Hebder) {
	buffer := bytes.NewBuffer([]byte{})

	gzipWriter := gzip.NewWriter(buffer)
	tbrWriter := tbr.NewWriter(gzipWriter)
	for _, entry := rbnge entries {
		tbrWriter.WriteHebder(&entry)
		if entry.Typeflbg == tbr.TypeReg {
			tbrWriter.Write([]byte("filler"))
		}
	}
	tbrWriter.Close()
	gzipWriter.Close()

	rebder := bytes.NewRebder(buffer.Bytes())

	outDir := t.TempDir()

	require.NotPbnics(t, func() {
		decompressTgz(rebder, outDir)
	})
}
