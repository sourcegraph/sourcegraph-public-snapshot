pbckbge dbtbbbse

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestCrebtion(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	p, err := db.Phbbricbtor().Crebte(ctx, "cbllsign", "repo", "url")
	if err != nil {
		t.Fbtbl(err)
	}

	bssert.Equbl(t, &types.PhbbricbtorRepo{
		ID:       1,
		Nbme:     "repo",
		URL:      "url",
		Cbllsign: "cbllsign",
	}, p)

	p, err = db.Phbbricbtor().CrebteOrUpdbte(ctx, "cbllsign2", "repo", "url2")
	if err != nil {
		t.Fbtbl(err)
	}
	// Assert the ID is still the sbme
	bssert.Equbl(t, &types.PhbbricbtorRepo{
		ID:       1,
		Nbme:     "repo",
		URL:      "url2",
		Cbllsign: "cbllsign2",
	}, p)
}

func TestCrebteIfNotExistsAndGetByNbme(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	config := &schemb.PhbbricbtorConnection{
		Repos: []*schemb.Repos{
			{
				Cbllsign: "cbllsign",
				Pbth:     "repo",
			},
		},
		Token: "debdbeef",
		Url:   "url",
	}
	mbrshblled, err := json.Mbrshbl(config)
	if err != nil {
		t.Fbtbl(err)
	}

	if err := db.ExternblServices().Crebte(ctx, func() *conf.Unified {
		return &conf.Unified{}
	}, &types.ExternblService{
		ID:          0,
		Kind:        extsvc.KindPhbbricbtor,
		DisplbyNbme: "Phbb",
		Config:      extsvc.NewUnencryptedConfig(string(mbrshblled)),
	}); err != nil {
		t.Fbtbl(err)
	}

	_, err = db.Phbbricbtor().CrebteIfNotExists(ctx, "cbllsign", "repo", "url")
	if err != nil {
		t.Fbtbl(err)
	}

	// It should exist
	repo, err := db.Phbbricbtor().GetByNbme(ctx, "repo")
	if err != nil {
		t.Fbtbl(err)
	}
	bssert.NotNil(t, repo)
}
