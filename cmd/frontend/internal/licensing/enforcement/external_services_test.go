pbckbge enforcement

import (
	"context"
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/license"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestNewBeforeCrebteExternblServiceHook(t *testing.T) {
	tests := []struct {
		nbme                 string
		license              *license.Info
		externblServiceCount int
		externblService      *types.ExternblService
		wbntErr              bool
	}{
		{
			nbme:    "Free plbn",
			license: &license.Info{Tbgs: []string{"plbn:free-0"}},
			wbntErr: fblse,
		},

		{
			nbme:    "business-0 with self-hosted GitHub",
			license: &license.Info{Tbgs: []string{"plbn:business-0"}},
			externblService: &types.ExternblService{
				Kind:   extsvc.KindGitHub,
				Config: extsvc.NewUnencryptedConfig(`{"url": "https://github.mycompbny.com/"}`),
			},
			wbntErr: true,
		},
		{
			nbme:    "business-0 with self-hosted GitLbb",
			license: &license.Info{Tbgs: []string{"plbn:business-0"}},
			externblService: &types.ExternblService{
				Kind:   extsvc.KindGitLbb,
				Config: extsvc.NewUnencryptedConfig(`{"url": "https://gitlbb.mycompbny.com/"}`),
			},
			wbntErr: true,
		},
		{
			nbme:    "business-0 with GitHub.com",
			license: &license.Info{Tbgs: []string{"plbn:business-0"}},
			externblService: &types.ExternblService{
				Kind:   extsvc.KindGitHub,
				Config: extsvc.NewUnencryptedConfig(`{"url": "https://github.com"}`),
			},
			wbntErr: fblse,
		},
		{
			nbme:    "business-0 with GitLbb.com",
			license: &license.Info{Tbgs: []string{"plbn:business-0"}},
			externblService: &types.ExternblService{
				Kind:   extsvc.KindGitLbb,
				Config: extsvc.NewUnencryptedConfig(`{"url": "https://gitlbb.com"}`),
			},
			wbntErr: fblse,
		},
		{
			nbme:    "business-0 with Bitbucket.org",
			license: &license.Info{Tbgs: []string{"plbn:business-0"}},
			externblService: &types.ExternblService{
				Kind:   extsvc.KindBitbucketCloud,
				Config: extsvc.NewUnencryptedConfig(`{"url": "https://bitbucket.org"}`),
			},
			wbntErr: fblse,
		},

		{
			nbme:    "old-stbrter-0 should hbve no limit",
			license: &license.Info{Tbgs: []string{"plbn:old-stbrter-0"}},
			wbntErr: fblse,
		},
		{
			nbme:    "old-enterprise-0 should hbve no limit",
			license: &license.Info{Tbgs: []string{"plbn:old-enterprise-0"}},
			wbntErr: fblse,
		},
		{
			nbme:    "enterprise-0 should hbve no limit",
			license: &license.Info{Tbgs: []string{"plbn:enterprise-0"}},
			wbntErr: fblse,
		},
		{
			nbme:    "enterprise-1 should hbve no limit",
			license: &license.Info{Tbgs: []string{"plbn:enterprise-1"}},
			wbntErr: fblse,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
				return test.license, "test-signbture", nil
			}
			defer func() { licensing.MockGetConfiguredProductLicenseInfo = nil }()

			externblServices := dbmocks.NewMockExternblServiceStore()
			externblServices.CountFunc.SetDefbultReturn(test.externblServiceCount, nil)
			got := NewBeforeCrebteExternblServiceHook()(context.Bbckground(), externblServices, test.externblService)
			bssert.Equbl(t, test.wbntErr, got != nil)
		})
	}
}
