pbckbge repos

import (
	"context"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/testutil"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestPbgureSource_ListRepos(t *testing.T) {
	conf := &schemb.PbgureConnection{
		Url:     "https://src.fedorbproject.org",
		Pbttern: "bc*",
	}
	cf, sbve := NewClientFbctory(t, t.Nbme())
	defer sbve(t)

	svc := &types.ExternblService{
		Kind:   extsvc.KindPbgure,
		Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, conf)),
	}

	ctx := context.Bbckground()
	src, err := NewPbgureSource(ctx, svc, cf)
	if err != nil {
		t.Fbtbl(err)
	}

	src.perPbge = 25 // 2 pbges for 47 results

	repos, err := ListAll(context.Bbckground(), src)
	if err != nil {
		t.Fbtbl(err)
	}

	testutil.AssertGolden(t, "testdbtb/sources/"+t.Nbme(), Updbte(t.Nbme()), repos)
}
