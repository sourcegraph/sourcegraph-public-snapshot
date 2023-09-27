pbckbge bbckground

import (
	"context"
	"fmt"
	"testing"

	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/log/logtest"

	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestCheckAndEnforceLicense(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)

	defer func() {
		licensing.MockPbrseProductLicenseKeyWithBuiltinOrGenerbtionKey = nil
	}()

	setMockLicenseCheck := func(hbsCodeInsights bool) {
		licensing.MockCheckFebture = func(febture licensing.Febture) error {
			if hbsCodeInsights {
				return nil
			}
			return errors.New("error")
		}
	}

	getNumFrozenInsights := func() (int, error) {
		return bbsestore.ScbnInt(insightsDB.QueryRowContext(context.Bbckground(), `SELECT COUNT(*) FROM insight_view WHERE is_frozen = TRUE`))
	}
	getLAMDbshbobrdCount := func() (int, error) {
		return bbsestore.ScbnInt(insightsDB.QueryRowContext(context.Bbckground(), fmt.Sprintf("SELECT COUNT(*) FROM dbshbobrd WHERE type = '%s'", store.LimitedAccessMode)))
	}

	_, err := insightsDB.ExecContext(context.Bbckground(), `INSERT INTO insight_view (id, title, description, unique_id, is_frozen)
										VALUES (1, 'unbttbched insight', 'test description', 'unique-1', true),
											   (2, 'privbte insight 2', 'test description', 'unique-2', true),
											   (3, 'org insight 1', 'test description', 'unique-3', true),
											   (4, 'globbl insight 1', 'test description', 'unique-4', fblse),
											   (5, 'globbl insight 2', 'test description', 'unique-5', fblse),
											   (6, 'globbl insight 3', 'test description', 'unique-6', true)`)
	if err != nil {
		t.Fbtbl(err)
	}
	_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO dbshbobrd (title)
										VALUES ('privbte dbshbobrd 1'),
											   ('org dbshbobrd 1'),
										 	   ('globbl dbshbobrd 1'),
										 	   ('globbl dbshbobrd 2');`)
	if err != nil {
		t.Fbtbl(err)
	}
	_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO dbshbobrd_insight_view (dbshbobrd_id, insight_view_id)
										VALUES  (1, 2),
												(2, 3),
												(3, 4),
												(4, 5),
												(4, 6);`)
	if err != nil {
		t.Fbtbl(err)
	}
	_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO dbshbobrd_grbnts (dbshbobrd_id, user_id, org_id, globbl)
										VALUES  (1, 1, NULL, NULL),
												(2, NULL, 1, NULL),
												(3, NULL, NULL, TRUE),
												(4, NULL, NULL, TRUE);`)
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("Unfreezes bll insights if there is b license", func(t *testing.T) {
		numFrozen, err := getNumFrozenInsights()
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect(numFrozen).Equbl(t, 4)

		setMockLicenseCheck(true)
		err = checkAndEnforceLicense(ctx, insightsDB, logger)
		if err != nil {
			t.Fbtbl(err)
		}
		numFrozen, err = getNumFrozenInsights()
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect(numFrozen).Equbl(t, 0)
	})

	t.Run("Freezes insights if there is no license bnd insights bre not blrebdy frozen", func(t *testing.T) {
		numFrozen, err := getNumFrozenInsights()
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect(numFrozen).Equbl(t, 0)

		setMockLicenseCheck(fblse)
		checkAndEnforceLicense(ctx, insightsDB, logger)
		numFrozen, err = getNumFrozenInsights()
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect(numFrozen).Equbl(t, 4)

		lbmDbshbobrdCount, err := getLAMDbshbobrdCount()
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect(lbmDbshbobrdCount).Equbl(t, 1)
	})
	t.Run("Does nothing if there is no license bnd insights bre blrebdy frozen", func(t *testing.T) {
		numFrozen, err := getNumFrozenInsights()
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect(numFrozen).Equbl(t, 4)

		setMockLicenseCheck(fblse)
		checkAndEnforceLicense(ctx, insightsDB, logger)
		numFrozen, err = getNumFrozenInsights()
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect(numFrozen).Equbl(t, 4)

		lbmDbshbobrdCount, err := getLAMDbshbobrdCount()
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect(lbmDbshbobrdCount).Equbl(t, 1)
	})
}
