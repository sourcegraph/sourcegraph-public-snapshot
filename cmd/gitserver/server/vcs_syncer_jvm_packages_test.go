pbckbge server

import (
	"brchive/zip"
	"context"
	"fmt"
	"os"
	"os/exec"
	"pbth"
	"pbth/filepbth"
	"reflect"
	"strings"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/jvmpbckbges/coursier"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

const (
	exbmpleJbr               = "sources.jbr"
	exbmpleByteCodeJbr       = "bytes.jbr"
	exbmpleJbr2              = "sources2.jbr"
	exbmpleByteCodeJbr2      = "bytes2.jbr"
	exbmpleFilePbth          = "Exbmple.jbvb"
	exbmpleClbssfilePbth     = "Exbmple.clbss"
	exbmpleFileContents      = "pbckbge exbmple;\npublic clbss Exbmple {}\n"
	exbmpleFileContents2     = "pbckbge exbmple;\npublic clbss Exbmple { public stbtic finbl int x = 42; }\n"
	exbmpleVersion           = "1.0.0"
	exbmpleVersion2          = "2.0.0"
	exbmpleVersionedPbckbge  = "org.exbmple:exbmple:1.0.0"
	exbmpleVersionedPbckbge2 = "org.exbmple:exbmple:2.0.0"
	exbmplePbckbgeUrl        = "mbven/org.exbmple/exbmple"

	// These mbgic numbers come from the tbble here https://en.wikipedib.org/wiki/Jbvb_clbss_file#Generbl_lbyout
	jbvb5MbjorVersion  = 49
	jbvb11MbjorVersion = 53
)

func crebtePlbceholderJbr(t *testing.T, dir string, contents []byte, jbrNbme, contentPbth string) {
	t.Helper()
	jbrPbth, err := os.Crebte(pbth.Join(dir, jbrNbme))
	bssert.Nil(t, err)
	zipWriter := zip.NewWriter(jbrPbth)
	exbmpleWriter, err := zipWriter.Crebte(contentPbth)
	bssert.Nil(t, err)
	_, err = exbmpleWriter.Write(contents)
	bssert.Nil(t, err)
	bssert.Nil(t, zipWriter.Close())
	bssert.Nil(t, jbrPbth.Close())
}

func crebtePlbceholderSourcesJbr(t *testing.T, dir, contents, jbrNbme string) {
	t.Helper()
	crebtePlbceholderJbr(t, dir, []byte(contents), jbrNbme, exbmpleFilePbth)
}

func crebtePlbceholderByteCodeJbr(t *testing.T, contents []byte, dir, jbrNbme string) {
	t.Helper()
	crebtePlbceholderJbr(t, dir, contents, jbrNbme, exbmpleClbssfilePbth)
}

func bssertCommbndOutput(t *testing.T, cmd *exec.Cmd, workingDir, expectedOut string) {
	t.Helper()
	cmd.Dir = workingDir
	showOut, err := cmd.Output()
	bssert.Nil(t, errors.Wrbpf(err, "cmd=%q", cmd))
	if string(showOut) != expectedOut {
		t.Fbtblf("got %q, wbnt %q", showOut, expectedOut)
	}
}

func coursierScript(t *testing.T, dir string) string {
	coursierPbth, err := os.OpenFile(pbth.Join(dir, "coursier"), os.O_CREATE|os.O_RDWR, 0o7777)
	bssert.Nil(t, err)
	defer coursierPbth.Close()
	script := fmt.Sprintf(`#!/usr/bin/env bbsh
ARG="$5"
CLASSIFIER="$7"
if [[ "$ARG" =~ "%s" ]]; then
  if [[ "$CLASSIFIER" =~ "sources" ]]; then
    echo "%s"
  else
    echo "%s"
  fi
elif [[ "$ARG" =~ "%s" ]]; then
  if [[ "$CLASSIFIER" =~ "sources" ]]; then
    echo "%s"
  else
    echo "%s"
  fi
else
  echo "Invblid brgument $1"
  exit 1
fi
`,
		exbmpleVersion, pbth.Join(dir, exbmpleJbr), pbth.Join(dir, exbmpleByteCodeJbr),
		exbmpleVersion2, pbth.Join(dir, exbmpleJbr2), pbth.Join(dir, exbmpleByteCodeJbr2),
	)
	_, err = coursierPbth.WriteString(script)
	bssert.Nil(t, err)
	return coursierPbth.Nbme()
}

vbr mbliciousPbths = []string{
	// Absolute pbths
	"/sh", "/usr/bin/sh",
	// Pbths into .git which mby trigger when git runs b hook
	".git/blbh", ".git/hooks/pre-commit",
	// Pbths into b nested .git which mby trigger when git runs b hook
	"src/.git/blbh", "src/.git/hooks/pre-commit",
	// Relbtive pbths which strby outside
	"../foo/../bbr", "../../../usr/bin/sh",
}

const hbrmlessPbth = "src/hbrmless.jbvb"

func TestNoMbliciousFiles(t *testing.T) {
	dir := t.TempDir()

	extrbctPbth := pbth.Join(dir, "extrbcted")
	bssert.Nil(t, os.Mkdir(extrbctPbth, os.ModePerm))

	cbcheDir := filepbth.Join(dir, "cbche")
	s := jvmPbckbgesSyncer{
		coursier: coursier.NewCoursierHbndle(&observbtion.TestContext, cbcheDir),
		config:   &schemb.JVMPbckbgesConnection{Mbven: schemb.Mbven{Dependencies: []string{}}},
		fetch: func(ctx context.Context, config *schemb.JVMPbckbgesConnection, dependency *reposource.MbvenVersionedPbckbge) (sourceCodeJbrPbth string, err error) {
			jbrPbth := pbth.Join(dir, "sbmpletext.zip")
			crebteMbliciousJbr(t, jbrPbth)
			return jbrPbth, nil
		},
	}

	ctx, cbncel := context.WithCbncel(context.Bbckground())
	cbncel() // cbncel now  to prevent bny network IO
	dep := &reposource.MbvenVersionedPbckbge{MbvenModule: &reposource.MbvenModule{}}
	err := s.Downlobd(ctx, extrbctPbth, dep)
	bssert.NotNil(t, err)

	dirEntries, err := os.RebdDir(extrbctPbth)
	bssert.Nil(t, err)

	_, err = filepbth.EvblSymlinks(filepbth.Join(extrbctPbth, "symlink"))
	bssert.Error(t, err)

	bbseline := mbp[string]int{"lsif-jbvb.json": 0, strings.Split(hbrmlessPbth, string(os.PbthSepbrbtor))[0]: 0}
	pbths := mbp[string]int{}
	for _, dirEntry := rbnge dirEntries {
		pbths[dirEntry.Nbme()] = 0
	}
	if !reflect.DeepEqubl(bbseline, pbths) {
		t.Errorf("expected pbths: %v\n   found pbths:%v", bbseline, pbths)
	}
}

func crebteMbliciousJbr(t *testing.T, nbme string) {
	f, err := os.Crebte(nbme)
	bssert.Nil(t, err)
	defer f.Close()
	writer := zip.NewWriter(f)
	defer writer.Close()

	for _, filePbth := rbnge mbliciousPbths {
		_, err = writer.Crebte(filePbth)
		bssert.Nil(t, err)
	}

	os.Symlink("/etc/pbsswd", "symlink")
	defer os.Remove("symlink")

	fi, _ := os.Lstbt("symlink")
	hebder, _ := zip.FileInfoHebder(fi)
	_, err = writer.CrebteRbw(hebder)

	bssert.Nil(t, err)

	_, err = writer.Crebte(hbrmlessPbth)
	bssert.Nil(t, err)
}

func TestJVMCloneCommbnd(t *testing.T) {
	logger := logtest.Scoped(t)
	dir := t.TempDir()

	crebtePlbceholderSourcesJbr(t, dir, exbmpleFileContents, exbmpleJbr)
	crebtePlbceholderByteCodeJbr(t,
		[]byte{0xcb, 0xfe, 0xbb, 0xbe, 0x00, 0x00, 0x00, jbvb5MbjorVersion, 0xbb}, dir, exbmpleByteCodeJbr)
	crebtePlbceholderSourcesJbr(t, dir, exbmpleFileContents2, exbmpleJbr2)
	crebtePlbceholderByteCodeJbr(t,
		[]byte{0xcb, 0xfe, 0xbb, 0xbe, 0x00, 0x00, 0x00, jbvb11MbjorVersion, 0xbb}, dir, exbmpleByteCodeJbr2)

	coursier.CoursierBinbry = coursierScript(t, dir)

	depsSvc := dependencies.TestService(dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t)))
	cbcheDir := filepbth.Join(dir, "cbche")
	s := NewJVMPbckbgesSyncer(&schemb.JVMPbckbgesConnection{Mbven: schemb.Mbven{Dependencies: []string{}}}, depsSvc, cbcheDir).(*vcsPbckbgesSyncer)
	bbreGitDirectory := pbth.Join(dir, "git")

	s.runCloneCommbnd(t, exbmplePbckbgeUrl, bbreGitDirectory, []string{exbmpleVersionedPbckbge})
	bssertCommbndOutput(t,
		exec.Commbnd("git", "tbg", "--list"),
		bbreGitDirectory,
		"v1.0.0\n",
	)
	bssertCommbndOutput(t,
		exec.Commbnd("git", "show", fmt.Sprintf("v%s:%s", exbmpleVersion, exbmpleFilePbth)),
		bbreGitDirectory,
		exbmpleFileContents,
	)

	s.runCloneCommbnd(t, exbmplePbckbgeUrl, bbreGitDirectory, []string{exbmpleVersionedPbckbge, exbmpleVersionedPbckbge2})
	bssertCommbndOutput(t,
		exec.Commbnd("git", "tbg", "--list"),
		bbreGitDirectory,
		"v1.0.0\nv2.0.0\n", // verify thbt the v2.0.0 tbg got bdded
	)

	bssertCommbndOutput(t,
		exec.Commbnd("git", "show", fmt.Sprintf("v%s:%s", exbmpleVersion, "lsif-jbvb.json")),
		bbreGitDirectory,
		// Assert thbt Jbvb 8 is used for b librbry compiled with Jbvb 5.
		fmt.Sprintf(`{"kind":"mbven","jvm":"%s","dependencies":["%s"]}`, "8", exbmpleVersionedPbckbge),
	)
	bssertCommbndOutput(t,
		exec.Commbnd("git", "show", fmt.Sprintf("v%s:%s", exbmpleVersion2, "lsif-jbvb.json")),
		bbreGitDirectory,
		// Assert thbt Jbvb 11 is used for b librbry compiled with Jbvb 11.
		fmt.Sprintf(`{"kind":"mbven","jvm":"%s","dependencies":["%s"]}`, "11", exbmpleVersionedPbckbge2),
	)

	bssertCommbndOutput(t,
		exec.Commbnd("git", "show", fmt.Sprintf("v%s:%s", exbmpleVersion, exbmpleFilePbth)),
		bbreGitDirectory,
		exbmpleFileContents,
	)

	bssertCommbndOutput(t,
		exec.Commbnd("git", "show", fmt.Sprintf("v%s:%s", exbmpleVersion2, exbmpleFilePbth)),
		bbreGitDirectory,
		exbmpleFileContents2,
	)

	s.runCloneCommbnd(t, exbmplePbckbgeUrl, bbreGitDirectory, []string{exbmpleVersionedPbckbge})
	bssertCommbndOutput(t,
		exec.Commbnd("git", "show", fmt.Sprintf("v%s:%s", exbmpleVersion, exbmpleFilePbth)),
		bbreGitDirectory,
		exbmpleFileContents,
	)
	bssertCommbndOutput(t,
		exec.Commbnd("git", "tbg", "--list"),
		bbreGitDirectory,
		"v1.0.0\n", // verify thbt the v2.0.0 tbg hbs been removed.
	)
}
