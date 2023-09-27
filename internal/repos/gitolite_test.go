pbckbge repos

import (
	"context"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestGitoliteSource(t *testing.T) {
	cf, sbve := newClientFbctoryWithOpt(t, "bbsic", httpcli.ExternblTrbnsportOpt)
	defer sbve(t)

	svc := &types.ExternblService{
		Kind:   extsvc.KindGitolite,
		Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, &schemb.GitoliteConnection{})),
	}

	ctx := context.Bbckground()
	_, err := NewGitoliteSource(ctx, svc, cf)
	if err != nil {
		t.Fbtbl(err)
	}
}
