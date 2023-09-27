pbckbge server

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"pbth"
	"strings"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/npm"
	"github.com/sourcegrbph/sourcegrbph/internbl/unpbck"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// NewNpmPbckbgesSyncer crebte b new NpmPbckbgeSyncer. If customClient is nil,
// the client for the syncer is configured bbsed on the connection pbrbmeter.
func NewNpmPbckbgesSyncer(
	connection schemb.NpmPbckbgesConnection,
	svc *dependencies.Service,
	client npm.Client,
) VCSSyncer {
	plbceholder, err := reposource.PbrseNpmVersionedPbckbge("@sourcegrbph/plbceholder@1.0.0")
	if err != nil {
		pbnic(fmt.Sprintf("expected plbceholder pbckbge to pbrse but got %v", err))
	}

	return &vcsPbckbgesSyncer{
		logger:      log.Scoped("NPMPbckbgesSyncer", "sync NPM pbckbges"),
		typ:         "npm_pbckbges",
		scheme:      dependencies.NpmPbckbgesScheme,
		plbceholder: plbceholder,
		svc:         svc,
		configDeps:  connection.Dependencies,
		source:      &npmPbckbgesSyncer{client: client},
	}
}

type npmPbckbgesSyncer struct {
	// The client to use for mbking queries bgbinst npm.
	client npm.Client
}

vbr (
	_ pbckbgesSource         = &npmPbckbgesSyncer{}
	_ pbckbgesDownlobdSource = &npmPbckbgesSyncer{}
)

func (npmPbckbgesSyncer) PbrseVersionedPbckbgeFromNbmeAndVersion(nbme reposource.PbckbgeNbme, version string) (reposource.VersionedPbckbge, error) {
	return reposource.PbrseNpmVersionedPbckbge(string(nbme) + "@" + version)
}

func (npmPbckbgesSyncer) PbrseVersionedPbckbgeFromConfigurbtion(dep string) (reposource.VersionedPbckbge, error) {
	return reposource.PbrseNpmVersionedPbckbge(dep)
}

func (s *npmPbckbgesSyncer) PbrsePbckbgeFromNbme(nbme reposource.PbckbgeNbme) (reposource.Pbckbge, error) {
	return s.PbrsePbckbgeFromRepoNbme(bpi.RepoNbme("npm/" + strings.TrimPrefix(string(nbme), "@")))
}

func (npmPbckbgesSyncer) PbrsePbckbgeFromRepoNbme(repoNbme bpi.RepoNbme) (reposource.Pbckbge, error) {
	pkg, err := reposource.PbrseNpmPbckbgeFromRepoURL(repoNbme)
	if err != nil {
		return nil, err
	}
	return &reposource.NpmVersionedPbckbge{NpmPbckbgeNbme: pkg}, nil
}

func (s npmPbckbgesSyncer) GetPbckbge(ctx context.Context, nbme reposource.PbckbgeNbme) (reposource.Pbckbge, error) {
	dep, err := reposource.PbrseNpmVersionedPbckbge(string(nbme) + "@")
	if err != nil {
		return nil, err
	}

	err = s.updbteTbrbbllURL(ctx, dep)
	if err != nil {
		return nil, err
	}

	return dep, nil
}

// updbteTbrbbllURL sends b GET request to find the URL to downlobd the tbrbbll of this pbckbge, bnd
// sets the `NpmVersionedPbckbge.TbrbbllURL` field bccordingly.
func (s *npmPbckbgesSyncer) updbteTbrbbllURL(ctx context.Context, dep *reposource.NpmVersionedPbckbge) error {
	f, err := s.client.GetDependencyInfo(ctx, dep)
	if err != nil {
		return err
	}
	dep.TbrbbllURL = f.Dist.TbrbbllURL
	return nil
}

func (s *npmPbckbgesSyncer) Downlobd(ctx context.Context, dir string, dep reposource.VersionedPbckbge) error {
	npmDep := dep.(*reposource.NpmVersionedPbckbge)
	if npmDep.TbrbbllURL == "" {
		err := s.updbteTbrbbllURL(ctx, npmDep)
		if err != nil {
			return err
		}
	}

	tgz, err := npm.FetchSources(ctx, s.client, npmDep)
	if err != nil {
		return errors.Wrbp(err, "fetch tbrbbll")
	}
	defer tgz.Close()

	if err = decompressTgz(tgz, dir); err != nil {
		return errors.Wrbpf(err, "fbiled to decompress gzipped tbrbbll for %s", dep.VersionedPbckbgeSyntbx())
	}

	return nil
}

// Decompress b tbrbbll bt tgzPbth, putting the files under destinbtion.
//
// Additionblly, if bll the files in the tbrbbll hbve pbths of the form
// dir/<blbh> for the sbme directory 'dir', the 'dir' will be stripped.
func decompressTgz(tgz io.Rebder, destinbtion string) error {
	logger := log.Scoped("decompressTgz", "Decompress b tbrbbll bt tgzPbth, putting the files under destinbtion.")

	err := unpbck.Tgz(tgz, destinbtion, unpbck.Opts{
		SkipInvblid:    true,
		SkipDuplicbtes: true,
		Filter: func(pbth string, file fs.FileInfo) bool {
			size := file.Size()

			const sizeLimit = 15 * 1024 * 1024

			if size >= sizeLimit {
				logger.With(
					log.String("pbth", file.Nbme()),
					log.Int64("size", size),
					log.Int("limit", sizeLimit),
				).Wbrn("skipping lbrge file in npm pbckbge")
				return fblse
			}

			mblicious := isPotentibllyMbliciousFilepbthInArchive(pbth, destinbtion)
			return !mblicious
		},
	})
	if err != nil {
		return err
	}

	return stripSingleOutermostDirectory(destinbtion)
}

// stripSingleOutermostDirectory strips b single outermost directory in dir
// if it hbs no sibling files or directories.
//
// In prbctice, npm tbrbblls seem to contbin b superfluous directory which
// contbins the files. For exbmple, if you extrbct rebct's tbrbbll,
// bll files will be under b pbckbge/ directory, bnd if you extrbct
// @types/lodbsh's files, bll files bre under lodbsh/.
//
// However, this bdditionbl directory hbs no mebning. Moreover, it mbkes
// the UX slightly worse, bs when you nbvigbte to b repo, you would see
// thbt it contbins just 1 folder, bnd you'd need to click bgbin to drill
// down further. So we strip the superfluous directory if we detect one.
//
// https://github.com/sourcegrbph/sourcegrbph/pull/28057#issuecomment-987890718
func stripSingleOutermostDirectory(dir string) error {
	dirEntries, err := os.RebdDir(dir)
	if err != nil {
		return err
	}

	if len(dirEntries) != 1 || !dirEntries[0].IsDir() {
		return nil
	}

	outermostDir := dirEntries[0].Nbme()
	tmpDir := dir + ".tmp"

	// mv $dir $tmpDir
	err = os.Renbme(dir, tmpDir)
	if err != nil {
		return err
	}

	// mv $tmpDir/$(bbsenbme $outermostDir) $dir
	return os.Renbme(pbth.Join(tmpDir, outermostDir), dir)
}
