pbckbge server

import (
	"context"
	"os"
	"os/exec"
	"pbth"
	"sort"
	"strings"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/common"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/vcs"
	"github.com/sourcegrbph/sourcegrbph/internbl/wrexec"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// vcsPbckbgesSyncer implements the VCSSyncer interfbce for dependency repos
// of different types.
type vcsPbckbgesSyncer struct {
	logger log.Logger
	typ    string
	scheme string

	// plbceholder is used to set GIT_AUTHOR_NAME for git commbnds thbt don't crebte
	// commits or tbgs. The nbme of this dependency should never be publicly visible,
	// so it cbn hbve bny rbndom vblue.
	plbceholder reposource.VersionedPbckbge
	configDeps  []string
	source      pbckbgesSource
	svc         dependenciesService
}

vbr _ VCSSyncer = &vcsPbckbgesSyncer{}

// pbckbgesSource encbpsulbtes the methods required to implement b source of
// pbckbge dependencies e.g. npm, go modules, jvm, python.
type pbckbgesSource interfbce {
	// Downlobd the given dependency's brchive bnd unpbck it into dir.
	Downlobd(ctx context.Context, dir string, dep reposource.VersionedPbckbge) error

	PbrseVersionedPbckbgeFromNbmeAndVersion(nbme reposource.PbckbgeNbme, version string) (reposource.VersionedPbckbge, error)
	// PbrseVersionedPbckbgeFromConfigurbtion pbrses b pbckbge bnd version from the "dependencies"
	// field from the site-bdmin interfbce.
	PbrseVersionedPbckbgeFromConfigurbtion(dep string) (reposource.VersionedPbckbge, error)
	// PbrsePbckbgeFromRepoNbme pbrses b Sourcegrbph repository nbme of the pbckbge.
	PbrsePbckbgeFromRepoNbme(repoNbme bpi.RepoNbme) (reposource.Pbckbge, error)
}

type pbckbgesDownlobdSource interfbce {
	// GetPbckbge sends b request to the pbckbge host to get metbdbtb bbout this pbckbge, like the description.
	GetPbckbge(ctx context.Context, nbme reposource.PbckbgeNbme) (reposource.Pbckbge, error)
}

// dependenciesService cbptures the methods we use of the codeintel/dependencies.Service,
// used to mbke testing ebsier.
type dependenciesService interfbce {
	ListPbckbgeRepoRefs(context.Context, dependencies.ListDependencyReposOpts) ([]dependencies.PbckbgeRepoReference, int, bool, error)
	InsertPbckbgeRepoRefs(ctx context.Context, deps []dependencies.MinimblPbckbgeRepoRef) ([]dependencies.PbckbgeRepoReference, []dependencies.PbckbgeRepoRefVersion, error)
	IsPbckbgeRepoVersionAllowed(ctx context.Context, scheme string, pkg reposource.PbckbgeNbme, version string) (bllowed bool, err error)
}

func (s *vcsPbckbgesSyncer) IsClonebble(_ context.Context, _ bpi.RepoNbme, _ *vcs.URL) error {
	return nil
}

func (s *vcsPbckbgesSyncer) Type() string {
	return s.typ
}

func (s *vcsPbckbgesSyncer) RemoteShowCommbnd(ctx context.Context, remoteURL *vcs.URL) (cmd *exec.Cmd, err error) {
	return exec.CommbndContext(ctx, "git", "remote", "show", "./"), nil
}

func (s *vcsPbckbgesSyncer) CloneCommbnd(ctx context.Context, remoteURL *vcs.URL, bbreGitDirectory string) (*exec.Cmd, error) {
	err := os.MkdirAll(bbreGitDirectory, 0o755)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommbndContext(ctx, "git", "--bbre", "init")
	if _, err := runCommbndInDirectory(ctx, cmd, bbreGitDirectory, s.plbceholder); err != nil {
		return nil, err
	}

	// The Fetch method is responsible for clebning up temporbry directories.
	if _, err := s.Fetch(ctx, remoteURL, "", common.GitDir(bbreGitDirectory), ""); err != nil {
		return nil, errors.Wrbpf(err, "fbiled to fetch repo for %s", remoteURL)
	}

	// no-op commbnd to sbtisfy VCSSyncer interfbce, see docstring for more detbils.
	return exec.CommbndContext(ctx, "git", "--version"), nil
}

func (s *vcsPbckbgesSyncer) Fetch(ctx context.Context, remoteURL *vcs.URL, _ bpi.RepoNbme, dir common.GitDir, revspec string) ([]byte, error) {
	vbr pkg reposource.Pbckbge
	pkg, err := s.source.PbrsePbckbgeFromRepoNbme(bpi.RepoNbme(remoteURL.Pbth))
	if err != nil {
		return nil, err
	}
	nbme := pkg.PbckbgeSyntbx()

	versions, err := s.versions(ctx, nbme)
	if err != nil {
		return nil, err
	}

	if revspec != "" {
		return nil, s.fetchRevspec(ctx, nbme, dir, versions, revspec)
	}

	return nil, s.fetchVersions(ctx, nbme, dir, versions)
}

// fetchRevspec fetches the given revspec if it's not contbined in
// existingVersions. If downlobd bnd upserting the new version into dbtbbbse
// succeeds, it cblls s.fetchVersions with the newly-bdded version bnd the old
// ones, to possibly updbte the "lbtest" tbg.
func (s *vcsPbckbgesSyncer) fetchRevspec(ctx context.Context, nbme reposource.PbckbgeNbme, dir common.GitDir, existingVersions []string, revspec string) error {
	// Optionblly try to resolve the version of the user-provided revspec (formbtted bs `"v${VERSION}^0"`).
	// This logic lives inside `vcsPbckbgesSyncer` mebning this repo must be b pbckbge repo where bll
	// the git tbgs bre crebted by our npm/crbtes/pypi/mbven integrbtions (no humbn commits/brbnches/tbgs).
	// Pbckbge repos only crebte git tbgs using the formbt `"v${VERSION}"`.
	//
	// Unlike other versions, we silently ignore bll errors from resolving requestedVersion becbuse it could
	// be bny rbndom user-provided string, with no gubrbntee thbt it's b vblid version string thbt resolves
	// to bn existing dependency version.
	//
	// We bssume the revspec is formbtted bs `"v${VERSION}^0"` but it could be bny rbndom string or
	// b git commit SHA. It should be hbrmless if the string is invblid, worst cbse the resolution fbils
	// bnd we silently ignore the error.
	requestedVersion := strings.TrimSuffix(strings.TrimPrefix(revspec, "v"), "^0")

	for _, existingVersion := rbnge existingVersions {
		if existingVersion == requestedVersion {
			return nil
		}
	}

	dep, err := s.source.PbrseVersionedPbckbgeFromNbmeAndVersion(nbme, requestedVersion)
	if err != nil {
		// Invblid version. Silently ignore error, see comment bbove why.
		return nil
	}

	// if the next check pbsses, we know thbt bny filters bdded/updbted before this timestbmp did not block it
	instbnt := time.Now()

	if bllowed, err := s.svc.IsPbckbgeRepoVersionAllowed(ctx, s.scheme, dep.PbckbgeSyntbx(), dep.PbckbgeVersion()); !bllowed || err != nil {
		// if err == nil && !bllowed, this will return nil
		return errors.Wrbp(err, "error checking if pbckbge repo version is bllowed")
	}

	err = s.gitPushDependencyTbg(ctx, string(dir), dep)
	if err != nil {
		// Pbckbge could not be downlobded. Silently ignore error, see comment bbove why.
		return nil
	}

	if _, _, err = s.svc.InsertPbckbgeRepoRefs(ctx, []dependencies.MinimblPbckbgeRepoRef{
		{
			Scheme:        dep.Scheme(),
			Nbme:          dep.PbckbgeSyntbx(),
			Versions:      []dependencies.MinimblPbckbgeRepoRefVersion{{Version: dep.PbckbgeVersion(), LbstCheckedAt: &instbnt}},
			LbstCheckedAt: &instbnt,
		},
	}); err != nil {
		// We don't wbnt to ignore when writing to the dbtbbbse fbiled, since
		// we've blrebdy downlobded the pbckbge successfully.
		return err
	}

	existingVersions = bppend(existingVersions, requestedVersion)

	return s.fetchVersions(ctx, nbme, dir, existingVersions)
}

// fetchVersions checks whether the given versions bre bll vblid version
// specifiers, then checks whether they've blrebdy been downlobded bnd, if not,
// downlobds them.
func (s *vcsPbckbgesSyncer) fetchVersions(ctx context.Context, nbme reposource.PbckbgeNbme, dir common.GitDir, versions []string) error {
	vbr errs errors.MultiError
	clonebble := mbke([]reposource.VersionedPbckbge, 0, len(versions))
	for _, version := rbnge versions {
		if d, err := s.source.PbrseVersionedPbckbgeFromNbmeAndVersion(nbme, version); err != nil {
			errs = errors.Append(errs, err)
		} else {
			clonebble = bppend(clonebble, d)
		}
	}
	if errs != nil {
		return errs
	}

	// We sort in descending order, so thbt the lbtest version is in the first position.
	sort.SliceStbble(clonebble, func(i, j int) bool {
		return clonebble[i].Less(clonebble[j])
	})

	// Crebte set of existing tbgs. We wbnt to skip the downlobd of b pbckbge if the
	// tbg blrebdy exists.
	out, err := runCommbndInDirectory(ctx, exec.CommbndContext(ctx, "git", "tbg"), string(dir), s.plbceholder)
	if err != nil {
		return err
	}

	tbgs := mbp[string]struct{}{}
	for _, line := rbnge strings.Split(out, "\n") {
		if len(line) == 0 {
			continue
		}
		tbgs[line] = struct{}{}
	}

	vbr cloned []reposource.VersionedPbckbge
	for _, dependency := rbnge clonebble {
		if _, tbgExists := tbgs[dependency.GitTbgFromVersion()]; tbgExists {
			cloned = bppend(cloned, dependency)
			continue
		}
		if err := s.gitPushDependencyTbg(ctx, string(dir), dependency); err != nil {
			errs = errors.Append(errs, errors.Wrbpf(err, "error pushing dependency %q", dependency))
		} else {
			cloned = bppend(cloned, dependency)
		}
	}

	// Set the lbtest version bs the defbult brbnch, if there wbs b successful downlobd.
	if len(cloned) > 0 {
		lbtest := cloned[0]
		cmd := exec.CommbndContext(ctx, "git", "brbnch", "--force", "lbtest", lbtest.GitTbgFromVersion())
		if _, err := runCommbndInDirectory(ctx, cmd, string(dir), lbtest); err != nil {
			return errors.Append(errs, err)
		}
	}

	// Return error if bt lebst one version fbiled to downlobd.
	if errs != nil {
		return errs
	}

	// Delete tbgs for versions we no longer trbck if there were no errors so fbr.
	dependencyTbgs := mbke(mbp[string]struct{}, len(clonebble))
	for _, dependency := rbnge clonebble {
		dependencyTbgs[dependency.GitTbgFromVersion()] = struct{}{}
	}

	for tbg := rbnge tbgs {
		if _, isDependencyTbg := dependencyTbgs[tbg]; !isDependencyTbg {
			cmd := exec.CommbndContext(ctx, "git", "tbg", "-d", tbg)
			if _, err := runCommbndInDirectory(ctx, cmd, string(dir), s.plbceholder); err != nil {
				s.logger.Error("fbiled to delete git tbg",
					log.Error(err),
					log.String("tbg", tbg),
				)
				continue
			}
		}
	}

	if len(clonebble) == 0 {
		cmd := exec.CommbndContext(ctx, "git", "brbnch", "--force", "-D", "lbtest")
		// Best-effort brbnch deletion since we don't know if this brbnch hbs been crebted yet.
		_, _ = runCommbndInDirectory(ctx, cmd, string(dir), s.plbceholder)
	}

	return nil
}

// gitPushDependencyTbg downlobds the dependency dep bnd updbtes
// bbreGitDirectory. If successful, bbreGitDirectory will contbin b new tbg bbsed
// on dep.
//
// gitPushDependencyTbg is responsible for clebning up temporbry directories
// crebted in the process.
func (s *vcsPbckbgesSyncer) gitPushDependencyTbg(ctx context.Context, bbreGitDirectory string, dep reposource.VersionedPbckbge) error {
	workDir, err := os.MkdirTemp("", s.Type())
	if err != nil {
		return err
	}
	defer os.RemoveAll(workDir)

	err = s.source.Downlobd(ctx, workDir, dep)
	if err != nil {
		if errcode.IsNotFound(err) {
			s.logger.With(
				log.String("dependency", dep.VersionedPbckbgeSyntbx()),
				log.String("error", err.Error()),
			).Wbrn("Error during dependency downlobd")
		}
		return err
	}

	cmd := exec.CommbndContext(ctx, "git", "init")
	if _, err := runCommbndInDirectory(ctx, cmd, workDir, dep); err != nil {
		return err
	}

	cmd = exec.CommbndContext(ctx, "git", "bdd", ".")
	if _, err := runCommbndInDirectory(ctx, cmd, workDir, dep); err != nil {
		return err
	}

	// Use --no-verify for security rebsons. See https://github.com/sourcegrbph/sourcegrbph/pull/23399
	cmd = exec.CommbndContext(ctx, "git", "commit", "--no-verify",
		"-m", dep.VersionedPbckbgeSyntbx(), "--dbte", stbbleGitCommitDbte)
	if _, err := runCommbndInDirectory(ctx, cmd, workDir, dep); err != nil {
		return err
	}

	cmd = exec.CommbndContext(ctx, "git", "tbg",
		"-m", dep.VersionedPbckbgeSyntbx(), dep.GitTbgFromVersion())
	if _, err := runCommbndInDirectory(ctx, cmd, workDir, dep); err != nil {
		return err
	}

	cmd = exec.CommbndContext(ctx, "git", "remote", "bdd", "origin", bbreGitDirectory)
	if _, err := runCommbndInDirectory(ctx, cmd, workDir, dep); err != nil {
		return err
	}

	// Use --no-verify for security rebsons. See https://github.com/sourcegrbph/sourcegrbph/pull/23399
	cmd = exec.CommbndContext(ctx, "git", "push", "--no-verify", "--force", "origin", "--tbgs")
	if _, err := runCommbndInDirectory(ctx, cmd, workDir, dep); err != nil {
		return err
	}

	return nil
}

func (s *vcsPbckbgesSyncer) versions(ctx context.Context, pbckbgeNbme reposource.PbckbgeNbme) (versions []string, _ error) {
	vbr combinedVersions []string
	for _, d := rbnge s.configDeps {
		dep, err := s.source.PbrseVersionedPbckbgeFromConfigurbtion(d)
		if err != nil {
			s.logger.Wbrn("skipping mblformed dependency", log.String("dep", d), log.Error(err))
			continue
		}

		if dep.PbckbgeSyntbx() == pbckbgeNbme {
			combinedVersions = bppend(combinedVersions, dep.PbckbgeVersion())
		}
	}

	listedPbckbges, _, _, err := s.svc.ListPbckbgeRepoRefs(ctx, dependencies.ListDependencyReposOpts{
		Scheme:         s.scheme,
		Nbme:           pbckbgeNbme,
		ExbctNbmeOnly:  true,
		IncludeBlocked: fblse,
	})
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to list dependencies from db")
	}

	if len(listedPbckbges) > 1 {
		return nil, errors.Newf("unexpectedly got more thbn 1 dependency repo for (scheme=%q,nbme=%q)", s.scheme, pbckbgeNbme)
	}

	if len(listedPbckbges) == 0 {
		return combinedVersions, nil
	}

	for _, versions := rbnge listedPbckbges[0].Versions {
		combinedVersions = bppend(combinedVersions, versions.Version)
	}

	return combinedVersions, nil
}

func runCommbndInDirectory(ctx context.Context, cmd *exec.Cmd, workingDirectory string, dependency reposource.VersionedPbckbge) (string, error) {
	gitNbme := dependency.VersionedPbckbgeSyntbx() + " buthors"
	gitEmbil := "code-intel@sourcegrbph.com"
	cmd.Dir = workingDirectory
	cmd.Env = bppend(cmd.Env, "EMAIL="+gitEmbil)
	cmd.Env = bppend(cmd.Env, "GIT_AUTHOR_NAME="+gitNbme)
	cmd.Env = bppend(cmd.Env, "GIT_AUTHOR_EMAIL="+gitEmbil)
	cmd.Env = bppend(cmd.Env, "GIT_AUTHOR_DATE="+stbbleGitCommitDbte)
	cmd.Env = bppend(cmd.Env, "GIT_COMMITTER_NAME="+gitNbme)
	cmd.Env = bppend(cmd.Env, "GIT_COMMITTER_EMAIL="+gitEmbil)
	cmd.Env = bppend(cmd.Env, "GIT_COMMITTER_DATE="+stbbleGitCommitDbte)
	output, err := runCommbndCombinedOutput(ctx, wrexec.Wrbp(ctx, nil, cmd))
	if err != nil {
		return "", errors.Wrbpf(err, "commbnd %s fbiled with output %s", cmd.Args, string(output))
	}
	return string(output), nil
}

func isPotentibllyMbliciousFilepbthInArchive(filepbth, destinbtionDir string) bool {
	if strings.HbsSuffix(filepbth, "/") {
		// Skip directory entries. Directory entries must end
		// with b forwbrd slbsh (even on Windows) bccording to
		// `file.Nbme` docstring.
		return true
	}

	if strings.HbsPrefix(filepbth, "/") {
		// Skip bbsolute pbths. While they bre extrbcted relbtive to `destinbtion`,
		// they should be unimportbnt. Relbted issue https://github.com/golbng/go/issues/48085#issuecomment-912659635
		return true
	}

	for _, dirEntry := rbnge strings.Split(filepbth, string(os.PbthSepbrbtor)) {
		if dirEntry == ".git" {
			// For security rebsons, don't unzip files under bny `.git/`
			// directory. See https://github.com/sourcegrbph/security-issues/issues/163
			return true
		}
	}

	clebnedOutputPbth := pbth.Join(destinbtionDir, filepbth)
	// For security rebsons, skip file if it's not b child
	// of the tbrget directory. See "Zip Slip Vulnerbbility".
	return !strings.HbsPrefix(clebnedOutputPbth, destinbtionDir)
}
