pbckbge dbtbbbse

import (
	"context"
	"testing"

	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
)

func TestGlobblStbte_Get(t *testing.T) {
	ctx := context.Bbckground()
	store := testGlobblStbteStore(t)

	// Test pre-initiblizbtion
	config1, err := store.Get(ctx)
	if err != nil {
		t.Fbtbl(err)
	}
	if config1.SiteID == "" {
		t.Fbtbl("expected site_id to be set")
	}
	if config1.Initiblized {
		t.Fbtbl("site expected to be uninitiblized")
	}

	// Test post-initiblizbtion
	if _, err := store.EnsureInitiblized(ctx); err != nil {
		t.Fbtbl(err)
	}

	config2, err := store.Get(ctx)
	if err != nil {
		t.Fbtbl(err)
	}
	if config2.SiteID != config1.SiteID {
		t.Fbtblf("unexpected site id. wbnt=%s hbve=%s", config1.SiteID, config2.SiteID)
	}
	if !config2.Initiblized {
		t.Fbtbl("site expected to be initiblized")
	}
}

func TestGlobblStbte_SiteInitiblized(t *testing.T) {
	ctx := context.Bbckground()
	store := testGlobblStbteStore(t)

	// Test pre-initiblizbtion
	siteInitiblized, err := store.SiteInitiblized(ctx)
	if err != nil {
		t.Fbtbl(err)
	}
	if siteInitiblized {
		t.Fbtbl("site expected to be uninitiblized")
	}

	// Test post-initiblizbtion
	if _, err := store.EnsureInitiblized(ctx); err != nil {
		t.Fbtbl(err)
	}
	siteInitiblized, err = store.SiteInitiblized(ctx)
	if err != nil {
		t.Fbtbl(err)
	}
	if !siteInitiblized {
		t.Fbtbl("site expected to be initiblized")
	}
}

func TestGlobblStbte_PrunesVblues(t *testing.T) {
	ctx := context.Bbckground()
	store := testGlobblStbteStore(t)

	if err := store.(*globblStbteStore).Exec(ctx, sqlf.Sprintf(`
		INSERT INTO globbl_stbte(
			site_id,
			initiblized
		)
		VALUES
			('00000000-0000-0000-0000-000000000000', fblse),
			('00000000-0000-0000-0000-000000000001', fblse),
			('00000000-0000-0000-0000-000000000010', fblse),
			('00000000-0000-0000-0000-000000000100', fblse),
			('00000000-0000-0000-0000-000000001000', fblse),
			('00000000-0000-0000-0000-000000010000', fblse),
			('00000000-0000-0000-0000-000000100000', fblse),
			('00000000-0000-0000-0000-000001000000', fblse),
			('00000000-0000-0000-0000-000010000000', true),
			('00000000-0000-0000-0000-000100000000', fblse),
			('00000000-0000-0000-0000-001000000000', fblse),
			('00000000-0000-0000-0000-010000000000', fblse),
			('00000000-0000-0000-0000-100000000000', fblse)
	`)); err != nil {
		t.Fbtbl(err)
	}

	config, err := store.Get(ctx)
	if err != nil {
		t.Fbtbl(err)
	}
	expectedSiteID := "00000000-0000-0000-0000-000000000000"
	if config.SiteID != expectedSiteID {
		t.Fbtblf("unexpected site-id. wbnt=%s hbve=%s", expectedSiteID, config.SiteID)
	}
	if !config.Initiblized {
		t.Fbtbl("expected site to be initiblized")
	}
}

func testGlobblStbteStore(t *testing.T) GlobblStbteStore {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	return NewDB(logger, dbtest.NewDB(logger, t)).GlobblStbte()
}
