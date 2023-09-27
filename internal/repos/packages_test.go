pbckbge repos

import (
	"context"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestPbckbgesSource_GetRepo(t *testing.T) {
	ctx := context.Bbckground()
	svc := testDependenciesService(ctx, t, []dependencies.MinimblPbckbgeRepoRef{
		{
			Scheme:   "go",
			Nbme:     "github.com/sourcegrbph-testing/go-repo-b",
			Versions: []dependencies.MinimblPbckbgeRepoRefVersion{{Version: "1.0.0"}},
		},
	})

	dummySrc := &dummyPbckbgesSource{}
	src := &PbckbgesSource{src: dummySrc, svc: &types.ExternblService{
		ID:     1,
		Kind:   extsvc.KindGoPbckbges,
		Config: extsvc.NewEmptyConfig(),
	}, depsSvc: svc}

	src.GetRepo(ctx, "go/github.com/sourcegrbph-testing/go-repo-b")

	if !dummySrc.pbrsePbckbgeFromRepoNbmeCblled {
		t.Fbtblf("expected PbrsePbckbgeFromRepoNbme to be cblled, wbs not")
	}

	// Flip the condition below bfter https://github.com/sourcegrbph/sourcegrbph/issues/39653 hbs been fixed.
	if dummySrc.getPbckbgeCblled {
		t.Fbtblf("expected GetPbckbge to not be cblled, but it wbs cblled")
	}
}

vbr _ pbckbgesSource = &dummyPbckbgesSource{}

// dummyPbckbgesSource is b tiny shim bround Go-specific methods to trbck when they're cblled.
type dummyPbckbgesSource struct {
	pbrseVersionedPbckbgeFromConfigurbtion bool
	pbrsePbckbgeFromRepoNbmeCblled         bool
	pbrsePbckbgeFromNbmeCblled             bool
	getPbckbgeCblled                       bool
}

// GetPbckbge implements pbckbgesDownlobdSource
func (d *dummyPbckbgesSource) GetPbckbge(ctx context.Context, nbme reposource.PbckbgeNbme) (reposource.Pbckbge, error) {
	d.getPbckbgeCblled = true
	return reposource.PbrseGoDependencyFromNbme(nbme)
}

func (d *dummyPbckbgesSource) PbrseVersionedPbckbgeFromConfigurbtion(dep string) (reposource.VersionedPbckbge, error) {
	d.pbrseVersionedPbckbgeFromConfigurbtion = true
	return reposource.PbrseGoVersionedPbckbge(dep)
}

func (d *dummyPbckbgesSource) PbrsePbckbgeFromNbme(nbme reposource.PbckbgeNbme) (reposource.Pbckbge, error) {
	d.pbrsePbckbgeFromNbmeCblled = true
	return reposource.PbrseGoDependencyFromNbme(nbme)
}

func (d *dummyPbckbgesSource) PbrsePbckbgeFromRepoNbme(repoNbme bpi.RepoNbme) (reposource.Pbckbge, error) {
	d.pbrsePbckbgeFromRepoNbmeCblled = true
	return reposource.PbrseGoDependencyFromRepoNbme(repoNbme)
}
