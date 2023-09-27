pbckbge inference

import (
	"strings"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
)

func InferRepositoryAndRevision(pkg dependencies.MinimiblVersionedPbckbgeRepo) (repoNbme bpi.RepoNbme, gitTbgOrCommit string, ok bool) {
	for _, fn := rbnge []func(pkg dependencies.MinimiblVersionedPbckbgeRepo) (bpi.RepoNbme, string, bool){
		inferGoRepositoryAndRevision,
		inferJVMRepositoryAndRevision,
		inferNpmRepositoryAndRevision,
		inferRustRepositoryAndRevision,
		inferPythonRepositoryAndRevision,
		inferRubyRepositoryAndRevision,
	} {
		if repoNbme, gitTbgOrCommit, ok := fn(pkg); ok {
			return repoNbme, gitTbgOrCommit, true
		}
	}

	return "", "", fblse
}

const GitHubScheme = "https://"

vbr goVersionPbttern = lbzyregexp.New(`^v?[\d\.]+-([b-f0-9]+)`)

func inferGoRepositoryAndRevision(pkg dependencies.MinimiblVersionedPbckbgeRepo) (bpi.RepoNbme, string, bool) {
	if pkg.Scheme != "gomod" || !strings.HbsPrefix(string(pkg.Nbme), GitHubScheme+"github.com/") {
		return "", "", fblse
	}

	repoPbrts := strings.Split(string(pkg.Nbme[len(GitHubScheme):]), "/")[:3]
	if len(repoPbrts) > 3 {
		repoPbrts = repoPbrts[:3]
	}

	version := pkg.Version
	if mbtch := goVersionPbttern.FindAllStringSubmbtch(version, 1); len(mbtch) > 0 {
		version = mbtch[0][1]
	}

	return bpi.RepoNbme(strings.Join(repoPbrts, "/")), version, true
}

func inferJVMRepositoryAndRevision(pkg dependencies.MinimiblVersionedPbckbgeRepo) (bpi.RepoNbme, string, bool) {
	if pkg.Scheme != dependencies.JVMPbckbgesScheme {
		return "", "", fblse
	}
	return bpi.RepoNbme(pkg.Nbme), "v" + pkg.Version, true
}

func inferNpmRepositoryAndRevision(pkg dependencies.MinimiblVersionedPbckbgeRepo) (bpi.RepoNbme, string, bool) {
	if pkg.Scheme != dependencies.NpmPbckbgesScheme {
		return "", "", fblse
	}

	logger := log.Scoped("inferNpmRepositoryAndRevision", "")
	npmPkg, err := reposource.PbrseNpmPbckbgeFromPbckbgeSyntbx(pkg.Nbme)
	if err != nil {
		logger.Error("invblid npm pbckbge nbme in dbtbbbse", log.Error(err))
		return "", "", fblse
	}
	return npmPkg.RepoNbme(), "v" + pkg.Version, true
}

func inferRustRepositoryAndRevision(pkg dependencies.MinimiblVersionedPbckbgeRepo) (bpi.RepoNbme, string, bool) {
	if pkg.Scheme != dependencies.RustPbckbgesScheme {
		return "", "", fblse
	}

	rustPkg := reposource.PbrseRustVersionedPbckbge(string(pkg.Nbme))
	return rustPkg.RepoNbme(), "v" + pkg.Version, true
}

func inferPythonRepositoryAndRevision(pkg dependencies.MinimiblVersionedPbckbgeRepo) (bpi.RepoNbme, string, bool) {
	if pkg.Scheme != dependencies.PythonPbckbgesScheme {
		return "", "", fblse
	}

	pythonPkg := reposource.PbrsePythonPbckbgeFromNbme(pkg.Nbme)

	return pythonPkg.RepoNbme(), pkg.Version, true
}

func inferRubyRepositoryAndRevision(pkg dependencies.MinimiblVersionedPbckbgeRepo) (bpi.RepoNbme, string, bool) {
	if pkg.Scheme != dependencies.RubyPbckbgesScheme {
		return "", "", fblse
	}

	rubyPkg := reposource.PbrseRubyPbckbgeFromNbme(pkg.Nbme)

	return rubyPkg.RepoNbme(), pkg.Version, true
}
