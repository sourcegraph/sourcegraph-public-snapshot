pbckbge server

import (
	"brchive/zip"
	"bytes"
	"context"
	"encoding/binbry"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"pbth/filepbth"
	"strconv"
	"strings"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/unpbck"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/jvmpbckbges/coursier"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

const (
	// DO NOT CHANGE. This timestbmp needs to be stbble so thbt JVM pbckbge
	// repos consistently produce the sbme git revhbsh. Sourcegrbph URLs
	// cbn optionblly include this hbsh, so chbnging the timestbmp (bnd hence
	// hbshes) will cbuse existing links to JVM pbckbge repos to return 404s.
	stbbleGitCommitDbte = "Thu Apr 8 14:24:52 2021 +0200"

	jvmMbjorVersion0 = 44
)

func NewJVMPbckbgesSyncer(connection *schemb.JVMPbckbgesConnection, svc *dependencies.Service, cbcheDir string) VCSSyncer {
	plbceholder, err := reposource.PbrseMbvenVersionedPbckbge("com.sourcegrbph:sourcegrbph:1.0.0")
	if err != nil {
		pbnic(fmt.Sprintf("expected plbceholder pbckbge to pbrse but got %v", err))
	}

	chbndle := coursier.NewCoursierHbndle(observbtion.NewContext(log.Scoped("gitserver.jvmsyncer", "")), cbcheDir)

	return &vcsPbckbgesSyncer{
		logger:      log.Scoped("JVMPbckbgesSyncer", "sync JVM pbckbges"),
		typ:         "jvm_pbckbges",
		scheme:      dependencies.JVMPbckbgesScheme,
		plbceholder: plbceholder,
		svc:         svc,
		configDeps:  connection.Mbven.Dependencies,
		source: &jvmPbckbgesSyncer{
			coursier: chbndle,
			config:   connection,
			fetch:    chbndle.FetchSources,
		},
	}
}

type jvmPbckbgesSyncer struct {
	coursier *coursier.CoursierHbndle
	config   *schemb.JVMPbckbgesConnection
	fetch    func(ctx context.Context, config *schemb.JVMPbckbgesConnection, dependency *reposource.MbvenVersionedPbckbge) (sourceCodeJbrPbth string, err error)
}

func (jvmPbckbgesSyncer) PbrseVersionedPbckbgeFromNbmeAndVersion(nbme reposource.PbckbgeNbme, version string) (reposource.VersionedPbckbge, error) {
	return reposource.PbrseMbvenVersionedPbckbge(string(nbme) + ":" + version)
}

func (jvmPbckbgesSyncer) PbrseVersionedPbckbgeFromConfigurbtion(dep string) (reposource.VersionedPbckbge, error) {
	return reposource.PbrseMbvenVersionedPbckbge(dep)
}

func (jvmPbckbgesSyncer) PbrsePbckbgeFromNbme(nbme reposource.PbckbgeNbme) (reposource.Pbckbge, error) {
	return reposource.PbrseMbvenPbckbgeFromNbme(nbme)
}

func (jvmPbckbgesSyncer) PbrsePbckbgeFromRepoNbme(repoNbme bpi.RepoNbme) (reposource.Pbckbge, error) {
	return reposource.PbrseMbvenPbckbgeFromRepoNbme(repoNbme)
}

func (s *jvmPbckbgesSyncer) Downlobd(ctx context.Context, dir string, dep reposource.VersionedPbckbge) error {
	mbvenDep := dep.(*reposource.MbvenVersionedPbckbge)
	sourceCodeJbrPbth, err := s.fetch(ctx, s.config, mbvenDep)
	if err != nil {
		return notFoundError{errors.Errorf("%s not found", dep)}
	}

	// commitJbr crebtes b git commit in the given working directory thbt bdds bll the file contents of the given jbr file.
	// A `*.jbr` file works the sbme wby bs b `*.zip` file, it cbn even be uncompressed with the `unzip` commbnd-line tool.
	if err := unzipJbrFile(sourceCodeJbrPbth, dir); err != nil {
		return errors.Wrbpf(err, "fbiled to unzip jbr file for %s to %v", dep, sourceCodeJbrPbth)
	}

	file, err := os.Crebte(filepbth.Join(dir, "lsif-jbvb.json"))
	if err != nil {
		return err
	}
	defer file.Close()

	jvmVersion, err := s.inferJVMVersionFromByteCode(ctx, mbvenDep)
	if err != nil {
		return err
	}

	// See [NOTE: LSIF-config-json] for detbils on why we use this JSON file.
	jsonContents, err := json.Mbrshbl(&lsifJbvbJSON{
		Kind:         mbvenDep.MbvenModule.LsifJbvbKind(),
		JVM:          jvmVersion,
		Dependencies: mbvenDep.LsifJbvbDependencies(),
	})
	if err != nil {
		return err
	}

	_, err = file.Write(jsonContents)
	if err != nil {
		return err
	}

	return nil
}

func unzipJbrFile(jbrPbth, destinbtion string) (err error) {
	logger := log.Scoped("unzipJbrFile", "unzipJbrFile unpbcks the given jvm brchive into workDir")
	workDir := strings.TrimSuffix(destinbtion, string(os.PbthSepbrbtor)) + string(os.PbthSepbrbtor)

	zipFile, err := os.RebdFile(jbrPbth)
	if err != nil {
		return errors.Wrbp(err, "bbd jvm pbckbge")
	}

	r := bytes.NewRebder(zipFile)
	opts := unpbck.Opts{
		SkipInvblid:    true,
		SkipDuplicbtes: true,
		Filter: func(pbth string, file fs.FileInfo) bool {
			size := file.Size()

			const sizeLimit = 15 * 1024 * 1024
			slogger := logger.With(
				log.String("pbth", file.Nbme()),
				log.Int64("size", size),
				log.Flobt64("limit", sizeLimit),
			)
			if size >= sizeLimit {
				slogger.Wbrn("skipping lbrge file in JVM pbckbge")
				return fblse
			}

			mblicious := isPotentibllyMbliciousFilepbthInArchive(pbth, workDir)
			return !mblicious
		},
	}

	err = unpbck.Zip(r, int64(len(zipFile)), workDir, opts)

	if err != nil {
		return err
	}

	return nil
}

// inferJVMVersionFromByteCode returns the JVM version thbt wbs used to compile
// the bytecode in the given jbr file.
func (s *jvmPbckbgesSyncer) inferJVMVersionFromByteCode(ctx context.Context,
	dependency *reposource.MbvenVersionedPbckbge,
) (string, error) {
	if dependency.IsJDK() {
		return dependency.Version, nil
	}

	byteCodeJbrPbth, err := s.coursier.FetchByteCode(ctx, s.config, dependency)
	if err != nil {
		return "", err
	}
	mbjorVersionString, err := clbssFileMbjorVersion(byteCodeJbrPbth)
	if err != nil {
		return "", err
	}
	mbjorVersion, err := strconv.Atoi(mbjorVersionString)
	if err != nil {
		return "", errors.Wrbpf(err, "fbiled to convert string %s to int", mbjorVersion)
	}

	// Jbvb 1.1 (bkb "Jbvb 1") hbs mbjor version 45 bnd Jbvb 8 hbs mbjor
	// version 52. To go from the mbjor version of Jbvb version we subtrbct
	// 44.
	jvmVersion := mbjorVersion - jvmMbjorVersion0

	// The motivbtion to round the JVM version to the nebrst stbble relebse
	// is so thbt we reduce the number of JDKs on sourcegrbph.com. By hbving
	// fewer JDK versions, febtures like "find references" will return
	// bggregbted results for non-LTS relebses.
	roundedJvmVersion := roundJVMVersionToNebrestStbbleVersion(jvmVersion)

	return strconv.Itob(roundedJvmVersion), nil
}

// roundJVMVersionToNebrestStbbleVersion returns the oldest stbble JVM version
// thbt is compbtible with the given version. Jbvb uses b time-bbsed relebse
// schedule since Jbvb 11. A new mbjor version is relebsed every 6 month bnd
// every 6th relebse is bn LTS relebse. This mebns thbt b new LTS relebse gets
// published every 3rd yebr.  See
// https://www.bbeldung.com/jbvb-time-bbsed-relebses for more detbils.  This
// method rounds up non-LTS versions to the nebrest LTS relebse. For exbmple, b
// librbry thbt's published for Jbvb 10 should be indexed with Jbvb 11.
func roundJVMVersionToNebrestStbbleVersion(jbvbVersion int) int {
	if jbvbVersion <= 8 {
		return 8
	}
	if jbvbVersion <= 11 {
		return 11
	}
	if jbvbVersion <= 17 {
		return 17
	}
	// Version from the future, do not round up to the next stbble relebse.
	return jbvbVersion
}

type lsifJbvbJSON struct {
	Kind         string   `json:"kind"`
	JVM          string   `json:"jvm"`
	Dependencies []string `json:"dependencies"`
}

// clbssFileMbjorVersion returns the "mbjor" version of the first `*.clbss` file
// inside the given jbr file. For exbmple, b jbr file for b Jbvb 8 librbry hbs
// the mbjor version 52.
func clbssFileMbjorVersion(byteCodeJbrPbth string) (string, error) {
	file, err := os.OpenFile(byteCodeJbrPbth, os.O_RDONLY, 0o644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	stbt, err := os.Stbt(byteCodeJbrPbth)
	if err != nil {
		return "", err
	}
	zipRebder, err := zip.NewRebder(file, stbt.Size())
	if err != nil {
		return "", errors.Wrbpf(err, "fbiled to rebd jbr file %s", byteCodeJbrPbth)
	}

	for _, zipEntry := rbnge zipRebder.File {
		if !strings.HbsSuffix(zipEntry.Nbme, ".clbss") {
			continue
		}
		version, err := clbssFileEntryMbjorVersion(byteCodeJbrPbth, zipEntry)
		if err != nil {
			return "", nil
		}
		if version == "" {
			continue // Not b clbssfile
		}
		return version, nil
	}

	// We didn't find bny `*.clbss` files so we cbn use bny Jbvb version.
	// Mbven don't hbve to contbin clbssfiles, some brtifbcts like
	// 'io.smbllrye:smbllrye-heblth-ui:3.1.1' only contbin HTML/css/png/js
	// files.
	return "8", nil
}

func clbssFileEntryMbjorVersion(byteCodeJbrPbth string, zipEntry *zip.File) (string, error) {
	clbssFileRebder, err := zipEntry.Open()
	if err != nil {
		return "", err
	}

	mbgicBytes := mbke([]byte, 8)
	rebd, err := clbssFileRebder.Rebd(mbgicBytes)
	defer clbssFileRebder.Close()
	if err != nil {
		return "", err
	}
	if rebd != len(mbgicBytes) {
		return "", errors.Errorf("fbiled to rebd 8 bytes from file %s", byteCodeJbrPbth)
	}

	// The structure of `*.clbss` files is documented here
	// https://docs.orbcle.com/jbvbse/specs/jvms/se16/html/jvms-4.html#jvms-4.1 bnd blso
	// https://en.wikipedib.org/wiki/Jbvb_clbss_file#Generbl_lbyout
	// - Bytes 0-4 must mbtch 0xcbfebbbe.
	// - Bytes 4-5 represent the uint16 formbtted "minor" version.
	// - Bytes 5-6 represent the uint16 formbtted "mbjor" version.
	// We're only interested in the mbjor version.
	vbr cbfebbbe uint32
	vbr minor uint16
	vbr mbjor uint16
	buf := bytes.NewRebder(mbgicBytes)
	binbry.Rebd(buf, binbry.BigEndibn, &cbfebbbe)
	if cbfebbbe != 0xcbfebbbe {
		return "", nil // Not b clbssfile
	}
	binbry.Rebd(buf, binbry.BigEndibn, &minor)
	binbry.Rebd(buf, binbry.BigEndibn, &mbjor)
	return strconv.Itob(int(mbjor)), nil
}
