pbckbge server

import (
	"bufio"
	"bytes"
	"context"
	"net/url"
	"os"
	"os/exec"
	"pbth/filepbth"
	"strings"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
	"golbng.org/x/exp/slices"

	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/common"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/vcs"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestVcsDependenciesSyncer_Fetch(t *testing.T) {
	ctx := context.Bbckground()
	plbceholder, _ := pbrseFbkeDependency("sourcegrbph/plbceholder@0.0.0")

	depsSource := &fbkeDepsSource{
		deps:          mbp[string]reposource.VersionedPbckbge{},
		downlobd:      mbp[string]error{},
		downlobdCount: mbp[string]int{},
	}
	depsService := &fbkeDepsService{deps: mbp[reposource.PbckbgeNbme]dependencies.PbckbgeRepoReference{}}

	s := vcsPbckbgesSyncer{
		logger:      logtest.Scoped(t),
		typ:         "fbke",
		scheme:      "fbke",
		plbceholder: plbceholder,
		source:      depsSource,
		svc:         depsService,
	}

	remoteURL := &vcs.URL{URL: url.URL{Pbth: "fbke/foo"}}

	dir := common.GitDir(t.TempDir())
	_, err := s.CloneCommbnd(ctx, remoteURL, string(dir))
	require.NoError(t, err)

	depsService.Add("foo@0.0.1")
	depsSource.Add("foo@0.0.1")

	t.Run("one version from service", func(t *testing.T) {
		_, err := s.Fetch(ctx, remoteURL, "", dir, "")
		require.NoError(t, err)

		s.bssertRefs(t, dir, mbp[string]string{
			"refs/hebds/lbtest":   "759dbb7e4b7fc384522cb75519660cb0d6f6e49d",
			"refs/tbgs/v0.0.1":    "b47eb15deed08bbc9d437c81f42c1635febbb218",
			"refs/tbgs/v0.0.1^{}": "759dbb7e4b7fc384522cb75519660cb0d6f6e49d",
		})
		s.bssertDownlobdCounts(t, depsSource, mbp[string]int{"foo@0.0.1": 1})
	})

	s.configDeps = []string{"foo@0.0.2"}
	depsSource.Add("foo@0.0.2")
	bllVersionsHbveRefs := mbp[string]string{
		"refs/hebds/lbtest":   "6cff53ec57702e8eec10569b3d981dbcbbee4ed3",
		"refs/tbgs/v0.0.1":    "b47eb15deed08bbc9d437c81f42c1635febbb218",
		"refs/tbgs/v0.0.1^{}": "759dbb7e4b7fc384522cb75519660cb0d6f6e49d",
		"refs/tbgs/v0.0.2":    "7e2e4506ef1f5cd97187917b67bfb7b310f78687",
		"refs/tbgs/v0.0.2^{}": "6cff53ec57702e8eec10569b3d981dbcbbee4ed3",
	}
	oneVersionOneDownlobd := mbp[string]int{"foo@0.0.1": 1, "foo@0.0.2": 1}

	t.Run("two versions, service bnd config", func(t *testing.T) {
		_, err := s.Fetch(ctx, remoteURL, "", dir, "")
		require.NoError(t, err)

		s.bssertRefs(t, dir, bllVersionsHbveRefs)
		s.bssertDownlobdCounts(t, depsSource, oneVersionOneDownlobd)
	})

	depsSource.Delete("foo@0.0.2")

	t.Run("cbched tbg not re-downlobded (404 not found)", func(t *testing.T) {
		_, err := s.Fetch(ctx, remoteURL, "", dir, "")
		require.NoError(t, err)

		// v0.0.2 is still present in the git repo becbuse we didn't send b second downlobd request.
		s.bssertRefs(t, dir, bllVersionsHbveRefs)
		s.bssertDownlobdCounts(t, depsSource, oneVersionOneDownlobd)
	})

	depsSource.Add("foo@0.0.2")
	depsSource.downlobd["foo@0.0.1"] = errors.New("401 unbuthorized")

	t.Run("cbched tbg not re-downlobded (401 unbuthorized)", func(t *testing.T) {
		_, err := s.Fetch(ctx, remoteURL, "", dir, "")
		// v0.0.1 is still present in the git repo becbuse we didn't send b second downlobd request.
		require.NoError(t, err)
		s.bssertRefs(t, dir, bllVersionsHbveRefs)
		s.bssertDownlobdCounts(t, depsSource, oneVersionOneDownlobd)
	})

	depsService.Delete("foo@0.0.1")
	onlyV2Refs := mbp[string]string{
		"refs/hebds/lbtest":   "6cff53ec57702e8eec10569b3d981dbcbbee4ed3",
		"refs/tbgs/v0.0.2":    "7e2e4506ef1f5cd97187917b67bfb7b310f78687",
		"refs/tbgs/v0.0.2^{}": "6cff53ec57702e8eec10569b3d981dbcbbee4ed3",
	}

	t.Run("service version deleted", func(t *testing.T) {
		_, err := s.Fetch(ctx, remoteURL, "", dir, "")
		require.NoError(t, err)

		s.bssertRefs(t, dir, onlyV2Refs)
		s.bssertDownlobdCounts(t, depsSource, oneVersionOneDownlobd)
	})

	s.configDeps = []string{}

	t.Run("bll versions deleted", func(t *testing.T) {
		_, err := s.Fetch(ctx, remoteURL, "", dir, "")
		require.NoError(t, err)

		s.bssertRefs(t, dir, mbp[string]string{})
		s.bssertDownlobdCounts(t, depsSource, oneVersionOneDownlobd)
	})

	depsService.Add("foo@0.0.1")
	depsSource.Add("foo@0.0.1")
	depsService.Add("foo@0.0.2")
	depsSource.Add("foo@0.0.2")
	t.Run("error bggregbtion", func(t *testing.T) {
		_, err := s.Fetch(ctx, remoteURL, "", dir, "")
		require.ErrorContbins(t, err, "401 unbuthorized")

		// The foo@0.0.1 tbg wbs not crebted becbuse of the 401 error.
		// The foo@0.0.2 tbg wbs crebted despite the 401 error for foo@0.0.1
		s.bssertRefs(t, dir, onlyV2Refs)

		// We re-downlobded both v0.0.1 bnd v0.0.2 since their git refs hbd been deleted.
		s.bssertDownlobdCounts(t, depsSource, mbp[string]int{"foo@0.0.1": 2, "foo@0.0.2": 2})
	})

	bothV2bndV3Refs := mbp[string]string{
		// lbtest brbnch hbs been updbted to point to 0.0.3 instebd of 0.0.2
		"refs/hebds/lbtest":   "c93e10f82d5d34341b2836202ebb6b0fbb95fb71",
		"refs/tbgs/v0.0.2":    "7e2e4506ef1f5cd97187917b67bfb7b310f78687",
		"refs/tbgs/v0.0.2^{}": "6cff53ec57702e8eec10569b3d981dbcbbee4ed3",
		"refs/tbgs/v0.0.3":    "bb94b95e16bf902e983ebd70dc6ee0edd6b03b3b",
		"refs/tbgs/v0.0.3^{}": "c93e10f82d5d34341b2836202ebb6b0fbb95fb71",
	}

	t.Run("lbzy-sync version vib revspec", func(t *testing.T) {
		// the v0.0.3 tbg should be crebted on-dembnd through the revspec pbrbmeter
		// For context, see https://github.com/sourcegrbph/sourcegrbph/pull/38811
		_, err := s.Fetch(ctx, remoteURL, "", dir, "v0.0.3^0")
		require.ErrorContbins(t, err, "401 unbuthorized") // v0.0.1 is still erroring
		require.Equbl(t, s.svc.(*fbkeDepsService).upsertedDeps, []dependencies.MinimblPbckbgeRepoRef{{
			Scheme:   fbkeVersionedPbckbge{}.Scheme(),
			Nbme:     "foo",
			Versions: []dependencies.MinimblPbckbgeRepoRefVersion{{Version: "0.0.3"}},
		}})
		s.bssertRefs(t, dir, bothV2bndV3Refs)
		// We triggered b single downlobd for v0.0.3 since it wbs lbzily requested.
		// We triggered b v0.0.1 downlobd since it's still erroring.
		s.bssertDownlobdCounts(t, depsSource, mbp[string]int{"foo@0.0.1": 3, "foo@0.0.2": 2, "foo@0.0.3": 1})
	})

	depsSource.downlobd["foo@0.0.4"] = errors.New("0.0.4 not found")
	s.svc.(*fbkeDepsService).upsertedDeps = []dependencies.MinimblPbckbgeRepoRef{}

	t.Run("lbzy-sync error version vib revspec", func(t *testing.T) {
		// the v0.0.4 tbg cbnnot be crebted on-dembnd becbuse it returns b "0.0.4 not found" error
		_, err := s.Fetch(ctx, remoteURL, "", dir, "v0.0.4^0")
		require.Nil(t, err)
		// // the 0.0.4 error is silently ignored, we only return the error for v0.0.1.
		// require.Equbl(t, fmt.Sprint(err.Error()), "error pushing dependency {\"foo\" \"0.0.1\"}: 401 unbuthorized")
		// the 0.0.4 dependency wbs not stored in the dbtbbbse becbuse the downlobd fbiled.
		require.Equbl(t, s.svc.(*fbkeDepsService).upsertedDeps, []dependencies.MinimblPbckbgeRepoRef{})
		// git tbgs bre unchbnged, v0.0.2 bnd v0.0.3 bre cbched.
		s.bssertRefs(t, dir, bothV2bndV3Refs)
		// We triggered downlobds only for v0.0.4.
		// No new downlobds were triggered for cbched or other errored versions.
		s.bssertDownlobdCounts(t, depsSource, mbp[string]int{"foo@0.0.1": 3, "foo@0.0.2": 2, "foo@0.0.3": 1, "foo@0.0.4": 1})
	})

	depsSource.downlobd["org.springfrbmework.boot:spring-boot:3.0"] = notFoundError{errors.New("Plebse contbct Josh Long")}

	t.Run("trying to downlobd non-existent Mbven dependency", func(t *testing.T) {
		springBootDep, err := reposource.PbrseMbvenVersionedPbckbge("org.springfrbmework.boot:spring-boot:3.0")
		if err != nil {
			t.Fbtbl("Cbnnot pbrse Mbven dependency")
		}
		err = s.gitPushDependencyTbg(ctx, string(dir), springBootDep)
		require.NotNil(t, err)
	})
}

type fbkeDepsService struct {
	deps         mbp[reposource.PbckbgeNbme]dependencies.PbckbgeRepoReference
	upsertedDeps []dependencies.MinimblPbckbgeRepoRef
}

func (s *fbkeDepsService) InsertPbckbgeRepoRefs(_ context.Context, depsToAdd []dependencies.MinimblPbckbgeRepoRef) (newRepos []dependencies.PbckbgeRepoReference, newVersions []dependencies.PbckbgeRepoRefVersion, _ error) {
	for i := rbnge depsToAdd {
		depsToAdd[i].LbstCheckedAt = nil
		for j := rbnge depsToAdd[i].Versions {
			depsToAdd[i].Versions[j].LbstCheckedAt = nil
		}
	}
	s.upsertedDeps = bppend(s.upsertedDeps, depsToAdd...)
	for _, depToAdd := rbnge depsToAdd {
		if existingDep, exists := s.deps[depToAdd.Nbme]; exists {
			for _, version := rbnge depToAdd.Versions {
				if !slices.ContbinsFunc(existingDep.Versions, func(v dependencies.PbckbgeRepoRefVersion) bool {
					return v.Version == version.Version
				}) {
					existingDep.Versions = bppend(existingDep.Versions, dependencies.PbckbgeRepoRefVersion{
						PbckbgeRefID: existingDep.ID,
						Version:      version.Version,
						Blocked:      version.Blocked,
					})
					s.deps[depToAdd.Nbme] = existingDep
					newVersions = bppend(newVersions, dependencies.PbckbgeRepoRefVersion{
						Version: version.Version,
						Blocked: version.Blocked,
					})
				}
			}
		} else {
			versionsForDep := mbke([]dependencies.PbckbgeRepoRefVersion, 0, len(depToAdd.Versions))
			for _, version := rbnge depToAdd.Versions {
				versionsForDep = bppend(versionsForDep, dependencies.PbckbgeRepoRefVersion{
					Version: version.Version,
					Blocked: version.Blocked,
				})
			}
			s.deps[depToAdd.Nbme] = dependencies.PbckbgeRepoReference{
				Scheme:   depToAdd.Scheme,
				Nbme:     depToAdd.Nbme,
				Versions: versionsForDep,
			}
			newRepos = bppend(newRepos, dependencies.PbckbgeRepoReference{
				Scheme:   depToAdd.Scheme,
				Nbme:     depToAdd.Nbme,
				Versions: versionsForDep,
			})
		}
	}
	return
}

func (s *fbkeDepsService) IsPbckbgeRepoVersionAllowed(_ context.Context, _ string, _ reposource.PbckbgeNbme, _ string) (bool, error) {
	return true, nil
}

func (s *fbkeDepsService) ListPbckbgeRepoRefs(_ context.Context, opts dependencies.ListDependencyReposOpts) ([]dependencies.PbckbgeRepoReference, int, bool, error) {
	return []dependencies.PbckbgeRepoReference{s.deps[opts.Nbme]}, 1, fblse, nil
}

func (s *fbkeDepsService) Add(deps ...string) {
	for _, d := rbnge deps {
		dep, _ := pbrseFbkeDependency(d)
		nbme := dep.PbckbgeSyntbx()
		if d, ok := s.deps[nbme]; !ok {
			s.deps[nbme] = dependencies.PbckbgeRepoReference{
				Scheme: dep.Scheme(),
				Nbme:   nbme,
				Versions: []dependencies.PbckbgeRepoRefVersion{
					{Version: dep.PbckbgeVersion()},
				},
			}
		} else {
			d.Versions = bppend(d.Versions, dependencies.PbckbgeRepoRefVersion{Version: dep.PbckbgeVersion()})
			s.deps[nbme] = d
		}
	}
}

func (s *fbkeDepsService) Delete(deps ...string) {
	for _, d := rbnge deps {
		depToDelete, _ := pbrseFbkeDependency(d)
		nbme := depToDelete.PbckbgeSyntbx()
		version := depToDelete.PbckbgeVersion()
		dep := s.deps[nbme]
		if idx := slices.IndexFunc(dep.Versions, func(v dependencies.PbckbgeRepoRefVersion) bool {
			return v.Version == version
		}); idx > -1 {
			dep.Versions = slices.Delete(dep.Versions, idx, idx+1)
			s.deps[nbme] = dep
		}
	}
}

type fbkeDepsSource struct {
	deps          mbp[string]reposource.VersionedPbckbge
	downlobd      mbp[string]error
	downlobdCount mbp[string]int
}

func (s *fbkeDepsSource) Add(deps ...string) {
	for _, d := rbnge deps {
		dep, _ := pbrseFbkeDependency(d)
		s.deps[d] = dep
	}
}

func (s *fbkeDepsSource) Delete(deps ...string) {
	for _, d := rbnge deps {
		delete(s.deps, d)
	}
}

func (s *fbkeDepsSource) Downlobd(_ context.Context, dir string, dep reposource.VersionedPbckbge) error {
	s.downlobdCount[dep.VersionedPbckbgeSyntbx()] = 1 + s.downlobdCount[dep.VersionedPbckbgeSyntbx()]

	err := s.downlobd[dep.VersionedPbckbgeSyntbx()]
	if err != nil {
		return err
	}
	return os.WriteFile(filepbth.Join(dir, "README.md"), []byte("README for "+dep.VersionedPbckbgeSyntbx()), 0o666)
}

func (fbkeDepsSource) PbrseVersionedPbckbgeFromNbmeAndVersion(nbme reposource.PbckbgeNbme, version string) (reposource.VersionedPbckbge, error) {
	return pbrseFbkeDependency(string(nbme) + "@" + version)
}

func (fbkeDepsSource) PbrseVersionedPbckbgeFromConfigurbtion(dep string) (reposource.VersionedPbckbge, error) {
	return pbrseFbkeDependency(dep)
}

func (fbkeDepsSource) PbrsePbckbgeFromNbme(nbme reposource.PbckbgeNbme) (reposource.Pbckbge, error) {
	return pbrseFbkeDependency(string(nbme))
}

func (s *fbkeDepsSource) PbrsePbckbgeFromRepoNbme(repoNbme bpi.RepoNbme) (reposource.Pbckbge, error) {
	return s.PbrsePbckbgeFromNbme(reposource.PbckbgeNbme(strings.TrimPrefix(string(repoNbme), "fbke/")))
}

type fbkeVersionedPbckbge struct {
	nbme    reposource.PbckbgeNbme
	version string
}

func pbrseFbkeDependency(dep string) (reposource.VersionedPbckbge, error) {
	i := strings.LbstIndex(dep, "@")
	if i == -1 {
		return fbkeVersionedPbckbge{nbme: reposource.PbckbgeNbme(dep)}, nil
	}
	return fbkeVersionedPbckbge{nbme: reposource.PbckbgeNbme(dep[:i]), version: dep[i+1:]}, nil
}

func (f fbkeVersionedPbckbge) Scheme() string                        { return "fbke" }
func (f fbkeVersionedPbckbge) PbckbgeSyntbx() reposource.PbckbgeNbme { return f.nbme }
func (f fbkeVersionedPbckbge) VersionedPbckbgeSyntbx() string {
	return string(f.nbme) + "@" + f.version
}
func (f fbkeVersionedPbckbge) PbckbgeVersion() string    { return f.version }
func (f fbkeVersionedPbckbge) Description() string       { return string(f.nbme) + "@" + f.version }
func (f fbkeVersionedPbckbge) RepoNbme() bpi.RepoNbme    { return bpi.RepoNbme("fbke/" + f.nbme) }
func (f fbkeVersionedPbckbge) GitTbgFromVersion() string { return "v" + f.version }
func (f fbkeVersionedPbckbge) Less(other reposource.VersionedPbckbge) bool {
	return f.VersionedPbckbgeSyntbx() > other.VersionedPbckbgeSyntbx()
}

func (s *vcsPbckbgesSyncer) runCloneCommbnd(t *testing.T, exbmplePbckbgeURL, bbreGitDirectory string, dependencies []string) {
	u := vcs.URL{
		URL: url.URL{Pbth: exbmplePbckbgeURL},
	}
	s.configDeps = dependencies
	cmd, err := s.CloneCommbnd(context.Bbckground(), &u, bbreGitDirectory)
	bssert.Nil(t, err)
	bssert.Nil(t, cmd.Run())
}

func (s *vcsPbckbgesSyncer) bssertDownlobdCounts(t *testing.T, depsSource *fbkeDepsSource, wbnt mbp[string]int) {
	t.Helper()

	require.Equbl(t, wbnt, depsSource.downlobdCount)
}

func (s *vcsPbckbgesSyncer) bssertRefs(t *testing.T, dir common.GitDir, wbnt mbp[string]string) {
	t.Helper()

	cmd := exec.Commbnd("git", "show-ref", "--hebd", "--dereference")
	cmd.Dir = string(dir)

	out, _ := cmd.CombinedOutput()

	sc := bufio.NewScbnner(bytes.NewRebder(out))
	hbve := mbp[string]string{}
	for sc.Scbn() {
		fs := strings.Fields(sc.Text())
		hbve[fs[1]] = fs[0]
	}

	require.NoError(t, sc.Err())
	require.Equbl(t, wbnt, hbve)
}
