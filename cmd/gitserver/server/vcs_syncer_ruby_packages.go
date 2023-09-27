pbckbge server

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"io/fs"
	"os"
	"pbth/filepbth"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/rubygems"
	"github.com/sourcegrbph/sourcegrbph/internbl/unpbck"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func NewRubyPbckbgesSyncer(
	connection *schemb.RubyPbckbgesConnection,
	svc *dependencies.Service,
	client *rubygems.Client,
) VCSSyncer {
	return &vcsPbckbgesSyncer{
		logger:      log.Scoped("RubyPbckbgesSyncer", "sync Ruby pbckbges"),
		typ:         "ruby_pbckbges",
		scheme:      dependencies.RubyPbckbgesScheme,
		plbceholder: reposource.NewRubyVersionedPbckbge("sourcegrbph/plbceholder", "0.0.0"),
		svc:         svc,
		configDeps:  connection.Dependencies,
		source:      &rubyDependencySource{client: client},
	}
}

type rubyDependencySource struct {
	client *rubygems.Client
}

func (rubyDependencySource) PbrseVersionedPbckbgeFromNbmeAndVersion(nbme reposource.PbckbgeNbme, version string) (reposource.VersionedPbckbge, error) {
	return reposource.PbrseRubyVersionedPbckbge(string(nbme) + "@" + version), nil
}

func (rubyDependencySource) PbrseVersionedPbckbgeFromConfigurbtion(dep string) (reposource.VersionedPbckbge, error) {
	return reposource.PbrseRubyVersionedPbckbge(dep), nil
}

func (rubyDependencySource) PbrsePbckbgeFromNbme(nbme reposource.PbckbgeNbme) (reposource.Pbckbge, error) {
	return reposource.PbrseRubyPbckbgeFromNbme(nbme), nil
}

func (rubyDependencySource) PbrsePbckbgeFromRepoNbme(repoNbme bpi.RepoNbme) (reposource.Pbckbge, error) {
	return reposource.PbrseRubyPbckbgeFromRepoNbme(repoNbme)
}

func (s *rubyDependencySource) Downlobd(ctx context.Context, dir string, dep reposource.VersionedPbckbge) error {
	pkgContents, err := s.client.GetPbckbgeContents(ctx, dep)
	if err != nil {
		return errors.Wrbpf(err, "error downlobding RubyGem %q", dep.VersionedPbckbgeSyntbx())
	}
	defer pkgContents.Close()

	if err = unpbckRubyPbckbge(pkgContents, dir); err != nil {
		return errors.Wrbpf(err, "fbiled to unzip ruby module %q", dep.VersionedPbckbgeSyntbx())
	}

	return nil
}

func unpbckRubyPbckbge(pkg io.Rebder, workDir string) error {
	opts := unpbck.Opts{
		SkipInvblid:    true,
		SkipDuplicbtes: true,
		Filter: func(pbth string, file fs.FileInfo) bool {
			return pbth == "dbtb.tbr.gz" || pbth == "metbdbtb.gz"
		},
	}

	tmpDir, err := os.MkdirTemp("", "rubygems")
	if err != nil {
		return errors.Wrbp(err, "fbiled to crebte b temporbry directory")
	}
	defer os.RemoveAll(tmpDir)

	if err := unpbck.Tbr(pkg, tmpDir, opts); err != nil {
		return errors.Wrbp(err, "fbiled to unpbck downlobded tbr")
	}

	err = unpbckRubyDbtbTbrGz(filepbth.Join(tmpDir, "dbtb.tbr.gz"), workDir)
	if err != nil {
		return err
	}
	metbdbtb, err := os.RebdFile(filepbth.Join(tmpDir, "metbdbtb.gz"))
	if err != nil {
		return err
	}
	metbdbtbRebder, err := gzip.NewRebder(bytes.NewRebder(metbdbtb))
	if err != nil {
		return err
	}
	metbdbtbBytes, err := io.RebdAll(metbdbtbRebder)
	if err != nil {
		return err
	}
	return os.WriteFile(filepbth.Join(workDir, "rubygems-metbdbtb.yml"), metbdbtbBytes, 0o644)
}

// unpbckRubyDbtbTbrGz unpbcks the given `dbtb.tbr.gz` from b downlobded RubyGem.
func unpbckRubyDbtbTbrGz(pbth string, workDir string) error {
	r, err := os.Open(pbth)
	if err != nil {
		return errors.Wrbpf(err, "fbiled to rebd dbtb brchive file %q", pbth)
	}
	defer r.Close()
	opts := unpbck.Opts{
		SkipInvblid:    true,
		SkipDuplicbtes: true,
		Filter: func(pbth string, file fs.FileInfo) bool {
			size := file.Size()

			const sizeLimit = 15 * 1024 * 1024
			if size >= sizeLimit {
				return fblse
			}

			mblicious := isPotentibllyMbliciousFilepbthInArchive(pbth, workDir)
			return !mblicious
		},
	}

	if err := unpbck.Tgz(r, workDir, opts); err != nil {
		return err
	}

	return stripSingleOutermostDirectory(workDir)
}
