package background

import (
	"context"
	"fmt"
	"testing"

	"github.com/hexops/autogold/v2"

	"github.com/sourcegraph/log/logtest"

	edb "github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestCheckAndEnforceLicense(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)

	defer func() {
		licensing.MockParseProductLicenseKeyWithBuiltinOrGenerationKey = nil
	}()

	setMockLicenseCheck := func(hasCodeInsights bool) {
		licensing.MockCheckFeature = func(feature licensing.Feature) error {
			if hasCodeInsights {
				return nil
			}
			return errors.New("error")
		}
	}

	getNumFrozenInsights := func() (int, error) {
		return basestore.ScanInt(insightsDB.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM insight_view WHERE is_frozen = TRUE`))
	}
	getLAMDashboardCount := func() (int, error) {
		return basestore.ScanInt(insightsDB.QueryRowContext(context.Background(), fmt.Sprintf("SELECT COUNT(*) FROM dashboard WHERE type = '%s'", store.LimitedAccessMode)))
	}

	_, err := insightsDB.ExecContext(context.Background(), `INSERT INTO insight_view (id, title, description, unique_id, is_frozen)
										VALUES (1, 'unattached insight', 'test description', 'unique-1', true),
											   (2, 'private insight 2', 'test description', 'unique-2', true),
											   (3, 'org insight 1', 'test description', 'unique-3', true),
											   (4, 'global insight 1', 'test description', 'unique-4', false),
											   (5, 'global insight 2', 'test description', 'unique-5', false),
											   (6, 'global insight 3', 'test description', 'unique-6', true)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO dashboard (title)
										VALUES ('private dashboard 1'),
											   ('org dashboard 1'),
										 	   ('global dashboard 1'),
										 	   ('global dashboard 2');`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO dashboard_insight_view (dashboard_id, insight_view_id)
										VALUES  (1, 2),
												(2, 3),
												(3, 4),
												(4, 5),
												(4, 6);`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO dashboard_grants (dashboard_id, user_id, org_id, global)
										VALUES  (1, 1, NULL, NULL),
												(2, NULL, 1, NULL),
												(3, NULL, NULL, TRUE),
												(4, NULL, NULL, TRUE);`)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Unfreezes all insights if there is a license", func(t *testing.T) {
		numFrozen, err := getNumFrozenInsights()
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect(numFrozen).Equal(t, 4)

		setMockLicenseCheck(true)
		err = checkAndEnforceLicense(ctx, insightsDB, logger)
		if err != nil {
			t.Fatal(err)
		}
		numFrozen, err = getNumFrozenInsights()
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect(numFrozen).Equal(t, 0)
	})

	t.Run("Freezes insights if there is no license and insights are not already frozen", func(t *testing.T) {
		numFrozen, err := getNumFrozenInsights()
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect(numFrozen).Equal(t, 0)

		setMockLicenseCheck(false)
		checkAndEnforceLicense(ctx, insightsDB, logger)
		numFrozen, err = getNumFrozenInsights()
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect(numFrozen).Equal(t, 4)

		lamDashboardCount, err := getLAMDashboardCount()
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect(lamDashboardCount).Equal(t, 1)
	})
	t.Run("Does nothing if there is no license and insights are already frozen", func(t *testing.T) {
		numFrozen, err := getNumFrozenInsights()
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect(numFrozen).Equal(t, 4)

		setMockLicenseCheck(false)
		checkAndEnforceLicense(ctx, insightsDB, logger)
		numFrozen, err = getNumFrozenInsights()
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect(numFrozen).Equal(t, 4)

		lamDashboardCount, err := getLAMDashboardCount()
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect(lamDashboardCount).Equal(t, 1)
	})
}
