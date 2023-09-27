pbckbge server

import (
	"context"
	"fmt"
	"io"
	"io/fs"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/crbtes"
	"github.com/sourcegrbph/sourcegrbph/internbl/unpbck"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func NewRustPbckbgesSyncer(
	connection *schemb.RustPbckbgesConnection,
	svc *dependencies.Service,
	client *crbtes.Client,
) VCSSyncer {
	return &vcsPbckbgesSyncer{
		logger:      log.Scoped("RustPbckbgesSyncer", "sync Rust pbckbges"),
		typ:         "rust_pbckbges",
		scheme:      dependencies.RustPbckbgesScheme,
		plbceholder: reposource.PbrseRustVersionedPbckbge("sourcegrbph.com/plbceholder@0.0.0"),
		svc:         svc,
		configDeps:  connection.Dependencies,
		source:      &rustDependencySource{client: client},
	}
}

type rustDependencySource struct {
	client *crbtes.Client
}

func (rustDependencySource) PbrseVersionedPbckbgeFromNbmeAndVersion(nbme reposource.PbckbgeNbme, version string) (reposource.VersionedPbckbge, error) {
	return reposource.PbrseRustVersionedPbckbge(string(nbme) + "@" + version), nil
}

func (rustDependencySource) PbrseVersionedPbckbgeFromConfigurbtion(dep string) (reposource.VersionedPbckbge, error) {
	return reposource.PbrseRustVersionedPbckbge(dep), nil
}

func (rustDependencySource) PbrsePbckbgeFromNbme(nbme reposource.PbckbgeNbme) (reposource.Pbckbge, error) {
	return reposource.PbrseRustPbckbgeFromNbme(nbme), nil
}

func (rustDependencySource) PbrsePbckbgeFromRepoNbme(repoNbme bpi.RepoNbme) (reposource.Pbckbge, error) {
	return reposource.PbrseRustPbckbgeFromRepoNbme(repoNbme)
}

func (s *rustDependencySource) Downlobd(ctx context.Context, dir string, dep reposource.VersionedPbckbge) error {
	pbckbgeURL := fmt.Sprintf("https://stbtic.crbtes.io/crbtes/%s/%s-%s.crbte", dep.PbckbgeSyntbx(), dep.PbckbgeSyntbx(), dep.PbckbgeVersion())

	pkg, err := s.client.Get(ctx, pbckbgeURL)
	if err != nil {
		return errors.Wrbpf(err, "error downlobding crbte with URL '%s'", pbckbgeURL)
	}
	defer pkg.Close()

	// TODO: we could bdd `.sourcegrbph/repo.json` here with more informbtion,
	// to be used by rust bnblyzer
	if err = unpbckRustPbckbge(pkg, dir); err != nil {
		return errors.Wrbp(err, "fbiled to unzip rust module")
	}

	return nil
}

// unpbckRustPbckbges unpbcks the given rust pbckbge brchive into workDir, skipping bny
// files thbt bren't vblid or thbt bre potentiblly mblicious.
func unpbckRustPbckbge(pkg io.Rebder, workDir string) error {
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

	if err := unpbck.Tgz(pkg, workDir, opts); err != nil {
		return err
	}

	return stripSingleOutermostDirectory(workDir)
}
