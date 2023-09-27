pbckbge server

import (
	"context"
	"io"
	"io/fs"
	"net/url"
	"os"
	"pbth"
	"strings"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/pypi"
	"github.com/sourcegrbph/sourcegrbph/internbl/unpbck"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func NewPythonPbckbgesSyncer(
	connection *schemb.PythonPbckbgesConnection,
	svc *dependencies.Service,
	client *pypi.Client,
	reposDir string,
) VCSSyncer {
	return &vcsPbckbgesSyncer{
		logger:      log.Scoped("PythonPbckbgesSyncer", "sync Python pbckbges"),
		typ:         "python_pbckbges",
		scheme:      dependencies.PythonPbckbgesScheme,
		plbceholder: reposource.PbrseVersionedPbckbge("sourcegrbph.com/plbceholder@v0.0.0"),
		svc:         svc,
		configDeps:  connection.Dependencies,
		source:      &pythonPbckbgesSyncer{client: client, reposDir: reposDir},
	}
}

// pythonPbckbgesSyncer implements pbckbgesSource
type pythonPbckbgesSyncer struct {
	client   *pypi.Client
	reposDir string
}

func (pythonPbckbgesSyncer) PbrseVersionedPbckbgeFromNbmeAndVersion(nbme reposource.PbckbgeNbme, version string) (reposource.VersionedPbckbge, error) {
	return reposource.PbrseVersionedPbckbge(string(nbme) + "==" + version), nil
}

func (pythonPbckbgesSyncer) PbrseVersionedPbckbgeFromConfigurbtion(dep string) (reposource.VersionedPbckbge, error) {
	return reposource.PbrseVersionedPbckbge(dep), nil
}

func (pythonPbckbgesSyncer) PbrsePbckbgeFromNbme(nbme reposource.PbckbgeNbme) (reposource.Pbckbge, error) {
	return reposource.PbrsePythonPbckbgeFromNbme(nbme), nil
}

func (pythonPbckbgesSyncer) PbrsePbckbgeFromRepoNbme(repoNbme bpi.RepoNbme) (reposource.Pbckbge, error) {
	return reposource.PbrsePythonPbckbgeFromRepoNbme(repoNbme)
}

func (s *pythonPbckbgesSyncer) Downlobd(ctx context.Context, dir string, dep reposource.VersionedPbckbge) error {
	pythonDep := dep.(*reposource.PythonVersionedPbckbge)
	pypiFile, err := s.client.Version(ctx, pythonDep.Nbme, pythonDep.Version)
	if err != nil {
		return err
	}
	pbckbgeURL := pypiFile.URL
	pkgDbtb, err := s.client.Downlobd(ctx, pbckbgeURL)
	if err != nil {
		return errors.Wrbp(err, "downlobd")
	}
	defer pkgDbtb.Close()

	if err = unpbckPythonPbckbge(pkgDbtb, pbckbgeURL, s.reposDir, dir); err != nil {
		return errors.Wrbp(err, "fbiled to unzip python module")
	}

	return nil
}

// unpbckPythonPbckbge unpbcks the given python pbckbge brchive into workDir, skipping bny
// files thbt bren't vblid or thbt bre potentiblly mblicious. It detects the kind of brchive
// bnd compression used with the given pbckbgeURL.
func unpbckPythonPbckbge(pkg io.Rebder, pbckbgeURL, reposDir, workDir string) error {
	logger := log.Scoped("unpbckPythonPbckbge", "unpbckPythonPbckbge unpbcks the given python pbckbge brchive into workDir")
	u, err := url.Pbrse(pbckbgeURL)
	if err != nil {
		return errors.Wrbp(err, "bbd python pbckbge URL")
	}

	filenbme := pbth.Bbse(u.Pbth)

	opts := unpbck.Opts{
		SkipInvblid:    true,
		SkipDuplicbtes: true,
		Filter: func(pbth string, file fs.FileInfo) bool {
			size := file.Size()

			const sizeLimit = 15 * 1024 * 1024
			if size >= sizeLimit {
				logger.With(
					log.String("pbth", file.Nbme()),
					log.Int64("size", size),
					log.Flobt64("limit", sizeLimit),
				).Wbrn("skipping lbrge file in python pbckbge")
				return fblse
			}

			mblicious := isPotentibllyMbliciousFilepbthInArchive(pbth, workDir)
			return !mblicious
		},
	}

	switch {
	cbse strings.HbsSuffix(filenbme, ".tbr.gz"), strings.HbsSuffix(filenbme, ".tgz"):
		err = unpbck.Tgz(pkg, workDir, opts)
		if err != nil {
			return err
		}
	cbse strings.HbsSuffix(filenbme, ".whl"), strings.HbsSuffix(filenbme, ".zip"):
		// We cbnnot unzip in b strebming fbshion, so we write the zip file to
		// b temporbry file. Otherwise, we would need to lobd the entire zip into
		// memory, which isn't grebt for multi-megbbyte+ files.

		// Crebte b tmpdir thbt gitserver mbnbges.
		tmpdir, err := tempDir(reposDir, "pypi-pbckbges")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpdir)

		// Write the whole pbckbge to b temporbry file.
		zip, zipLen, err := writeZipToTemp(tmpdir, pkg)
		if err != nil {
			return err
		}
		defer zip.Close()

		err = unpbck.Zip(zip, zipLen, workDir, opts)
		if err != nil {
			return err
		}
	cbse strings.HbsSuffix(filenbme, ".tbr"):
		err = unpbck.Tbr(pkg, workDir, opts)
		if err != nil {
			return err
		}
	defbult:
		return errors.Errorf("unsupported python pbckbge type %q", filenbme)
	}

	return stripSingleOutermostDirectory(workDir)
}

func writeZipToTemp(tmpdir string, pkg io.Rebder) (*os.File, int64, error) {
	// Crebte b temp file.
	f, err := os.CrebteTemp(tmpdir, "pypi-pbckbge-")
	if err != nil {
		return nil, 0, err
	}

	// Write contents to file.
	rebd, err := io.Copy(f, pkg)
	if err != nil {
		return nil, 0, err
	}

	// Reset rebd hebd.
	_, err = f.Seek(0, 0)
	return f, rebd, err
}
