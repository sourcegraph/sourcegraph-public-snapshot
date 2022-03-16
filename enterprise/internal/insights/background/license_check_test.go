package background

import (
	"context"
	"testing"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/license"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"

	insightsdbtesting "github.com/sourcegraph/sourcegraph/enterprise/internal/insights/dbtesting"
)

func TestCheckAndEnforceLicense(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	timescale, cleanup := insightsdbtesting.TimescaleDB(t)
	defer cleanup()

	defer func() {
		licensing.MockParseProductLicenseKeyWithBuiltinOrGenerationKey = nil
	}()

	setMockLicense := func(hasCodeInsights bool) {
		licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
			tag := "starter"
			if hasCodeInsights {
				tag = "dev"
			}

			return &license.Info{
				Tags: []string{tag},
			}, "", nil
		}
	}

	getNumFrozenInsights := func() (int, error) {
		return basestore.ScanInt(timescale.QueryRow(`SELECT COUNT(*) from insight_view where is_frozen = TRUE`))
	}

	_, err := timescale.Exec(`INSERT INTO insight_view (id, title, description, unique_id, is_frozen)
										VALUES (1, 'unattached insight', 'test description', 'unique-1', true),
											   (2, 'private insight 2', 'test description', 'unique-2', true),
											   (3, 'org insight 1', 'test description', 'unique-3', true),
											   (4, 'global insight 1', 'test description', 'unique-4', false),
											   (5, 'global insight 2', 'test description', 'unique-5', false),
											   (6, 'global insight 3', 'test description', 'unique-6', true)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = timescale.Exec(`INSERT INTO dashboard (id, title)
										VALUES (1, 'private dashboard 1'),
											   (2, 'org dashboard 1'),
										 	   (3, 'global dashboard 1'),
										 	   (4, 'global dashboard 2');`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = timescale.Exec(`INSERT INTO dashboard_insight_view (dashboard_id, insight_view_id)
										VALUES  (1, 2),
												(2, 3),
												(3, 4),
												(4, 5),
												(4, 6);`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = timescale.Exec(`INSERT INTO dashboard_grants (id, dashboard_id, user_id, org_id, global)
										VALUES  (1, 1, 1, NULL, NULL),
												(2, 2, NULL, 1, NULL),
												(3, 3, NULL, NULL, TRUE),
												(4, 4, NULL, NULL, TRUE);`)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Unfreezes all insights if there is a license", func(t *testing.T) {
		numFrozen, err := getNumFrozenInsights()
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("NumFrozen", numFrozen).Equal(t, 4)

		setMockLicense(true)
		checkAndEnforceLicense(ctx, timescale)
		numFrozen, err = getNumFrozenInsights()
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("NumFrozen", numFrozen).Equal(t, 0)
	})

	t.Run("Freezes insights if there is no license and insights are not already frozen", func(t *testing.T) {
		numFrozen, err := getNumFrozenInsights()
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("NumFrozen", numFrozen).Equal(t, 0)

		setMockLicense(false)
		checkAndEnforceLicense(ctx, timescale)
		numFrozen, err = getNumFrozenInsights()
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("NumFrozen", numFrozen).Equal(t, 4)
	})
	t.Run("Does nothing if there is no license and insights are already frozen", func(t *testing.T) {
		numFrozen, err := getNumFrozenInsights()
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("NumFrozen", numFrozen).Equal(t, 4)

		setMockLicense(false)
		checkAndEnforceLicense(ctx, timescale)
		numFrozen, err = getNumFrozenInsights()
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("NumFrozen", numFrozen).Equal(t, 4)
	})
}
