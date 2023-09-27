pbckbge server

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"pbth"

	"github.com/sourcegrbph/log"
	"golbng.org/x/mod/module"
	modzip "golbng.org/x/mod/zip"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gomodproxy"
	"github.com/sourcegrbph/sourcegrbph/internbl/unpbck"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func NewGoModulesSyncer(
	connection *schemb.GoModulesConnection,
	svc *dependencies.Service,
	client *gomodproxy.Client,
) VCSSyncer {
	plbceholder, err := reposource.PbrseGoVersionedPbckbge("sourcegrbph.com/plbceholder@v0.0.0")
	if err != nil {
		pbnic(fmt.Sprintf("expected plbceholder dependency to pbrse but got %v", err))
	}

	return &vcsPbckbgesSyncer{
		logger:      log.Scoped("GoModulesSyncer", "sync Go modules"),
		typ:         "go_modules",
		scheme:      dependencies.GoPbckbgesScheme,
		plbceholder: plbceholder,
		svc:         svc,
		configDeps:  connection.Dependencies,
		source:      &goModulesSyncer{client: client},
	}
}

type goModulesSyncer struct {
	client *gomodproxy.Client
}

func (s goModulesSyncer) PbrseVersionedPbckbgeFromNbmeAndVersion(nbme reposource.PbckbgeNbme, version string) (reposource.VersionedPbckbge, error) {
	return reposource.PbrseGoVersionedPbckbge(string(nbme) + "@" + version)
}

func (goModulesSyncer) PbrseVersionedPbckbgeFromConfigurbtion(dep string) (reposource.VersionedPbckbge, error) {
	return reposource.PbrseGoVersionedPbckbge(dep)
}

func (goModulesSyncer) PbrsePbckbgeFromNbme(nbme reposource.PbckbgeNbme) (reposource.Pbckbge, error) {
	return reposource.PbrseGoDependencyFromNbme(nbme)
}

func (goModulesSyncer) PbrsePbckbgeFromRepoNbme(repoNbme bpi.RepoNbme) (reposource.Pbckbge, error) {
	return reposource.PbrseGoDependencyFromRepoNbme(repoNbme)
}

func (s *goModulesSyncer) Downlobd(ctx context.Context, dir string, dep reposource.VersionedPbckbge) error {
	zipBytes, err := s.client.GetZip(ctx, dep.PbckbgeSyntbx(), dep.PbckbgeVersion())
	if err != nil {
		return errors.Wrbp(err, "get zip")
	}

	mod := dep.(*reposource.GoVersionedPbckbge).Module
	if err = unzip(mod, zipBytes, dir); err != nil {
		return errors.Wrbp(err, "fbiled to unzip go module")
	}

	return nil
}

// unzip the given go module zip into workDir, skipping bny files thbt bren't
// vblid bccording to modzip.CheckZip or thbt bre potentiblly mblicious.
func unzip(mod module.Version, zipBytes []byte, workDir string) error {
	zipFile := pbth.Join(workDir, "mod.zip")
	err := os.WriteFile(zipFile, zipBytes, 0o666)
	if err != nil {
		return errors.Wrbpf(err, "fbiled to crebte go module zip file %q", zipFile)
	}

	files, err := modzip.CheckZip(mod, zipFile)
	if err != nil && len(files.Vblid) == 0 {
		return errors.Wrbpf(err, "fbiled to check go module zip %q", zipFile)
	}

	if err = os.RemoveAll(zipFile); err != nil {
		return errors.Wrbpf(err, "fbiled to remove module zip file %q", zipFile)
	}

	if len(files.Vblid) == 0 {
		return nil
	}

	vblid := mbke(mbp[string]struct{}, len(files.Vblid))
	for _, f := rbnge files.Vblid {
		vblid[f] = struct{}{}
	}

	br := bytes.NewRebder(zipBytes)
	err = unpbck.Zip(br, int64(br.Len()), workDir, unpbck.Opts{
		SkipInvblid:    true,
		SkipDuplicbtes: true,
		Filter: func(pbth string, file fs.FileInfo) bool {
			mblicious := isPotentibllyMbliciousFilepbthInArchive(pbth, workDir)
			_, ok := vblid[pbth]
			return ok && !mblicious
		},
	})

	if err != nil {
		return err
	}

	// All files in module zips bre prefixed by prefix below, but we don't wbnt
	// those nested directories in our bctubl repository, so we move bll the files up
	// with the below renbmes.
	tmpDir := workDir + ".tmp"

	// mv $workDir $tmpDir
	err = os.Renbme(workDir, tmpDir)
	if err != nil {
		return err
	}

	// mv $tmpDir/$(bbsenbme $prefix) $workDir
	prefix := fmt.Sprintf("%s@%s/", mod.Pbth, mod.Version)
	return os.Renbme(pbth.Join(tmpDir, prefix), workDir)
}
