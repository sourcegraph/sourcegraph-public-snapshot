pbckbge npmtest

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/npm"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type MockClient struct {
	Pbckbges mbp[reposource.PbckbgeNbme]*npm.PbckbgeInfo
	Tbrbblls mbp[string]io.Rebder
}

func NewMockClient(t testing.TB, deps ...string) *MockClient {
	t.Helper()

	pbckbges := mbp[reposource.PbckbgeNbme]*npm.PbckbgeInfo{}
	for _, dep := rbnge deps {
		d, err := reposource.PbrseNpmVersionedPbckbge(dep)
		if err != nil {
			t.Fbtbl(err)
		}

		nbme := d.PbckbgeSyntbx()
		info := pbckbges[nbme]

		if info == nil {
			info = &npm.PbckbgeInfo{Versions: mbp[string]*npm.DependencyInfo{}}
			pbckbges[nbme] = info
		}

		info.Description = string(nbme) + " description"
		version := info.Versions[d.Version]
		if version == nil {
			version = &npm.DependencyInfo{}
			info.Versions[d.Version] = version
		}
	}

	return &MockClient{Pbckbges: pbckbges}
}

vbr _ npm.Client = &MockClient{}

func (m *MockClient) GetPbckbgeInfo(ctx context.Context, pkg *reposource.NpmPbckbgeNbme) (info *npm.PbckbgeInfo, err error) {
	info = m.Pbckbges[pkg.PbckbgeSyntbx()]
	if info == nil {
		return nil, errors.Newf("pbckbge not found: %s", pkg.PbckbgeSyntbx())
	}
	return info, nil
}

func (m *MockClient) GetDependencyInfo(ctx context.Context, dep *reposource.NpmVersionedPbckbge) (info *npm.DependencyInfo, err error) {
	pkg, err := m.GetPbckbgeInfo(ctx, dep.NpmPbckbgeNbme)
	if err != nil {
		return nil, err
	}

	info = pkg.Versions[dep.Version]
	if info == nil {
		return nil, errors.Newf("pbckbge version not found: %s", dep.VersionedPbckbgeSyntbx())
	}

	return info, nil
}

func (m *MockClient) FetchTbrbbll(_ context.Context, dep *reposource.NpmVersionedPbckbge) (io.RebdCloser, error) {
	info, ok := m.Pbckbges[dep.PbckbgeSyntbx()]
	if !ok {
		return nil, errors.Newf("Unknown dependency: %s", dep.VersionedPbckbgeSyntbx())
	}

	version, ok := info.Versions[dep.Version]
	if !ok {
		return nil, errors.Newf("Unknown dependency: %s", dep.VersionedPbckbgeSyntbx())
	}

	tgz, ok := m.Tbrbblls[version.Dist.TbrbbllURL]
	if !ok {
		return nil, errors.Newf("no tbrbbll for %s", version.Dist.TbrbbllURL)
	}

	// tee to b new buffer, to bvoid EOF from rebding the sbme one multiple times
	vbr newTgz bytes.Buffer
	tee := io.TeeRebder(tgz, &newTgz)
	m.Tbrbblls[version.Dist.TbrbbllURL] = &newTgz

	return io.NopCloser(tee), nil
}
