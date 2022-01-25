package usagestats

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestCTAMetricsUsageStatistics(t *testing.T) {
	ctx := context.Background()

	defer func() {
		timeNow = time.Now
	}()

	now := time.Date(2021, 1, 20, 0, 0, 0, 0, time.UTC)
	mockTimeNow(now)

	db := database.NewDB(dbtest.NewDB(t))

	_, err := db.ExecContext(context.Background(), `
		INSERT INTO event_logs
			(id, name, argument, url, user_id, anonymous_user_id, source, version, timestamp)
		VALUES
			-- Current month event logs
			(1, 'InstallBrowserExtensionCTAShown', '{"page": "file"}', 'https://sourcegraph.test:3443/search', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			(2, 'InstallBrowserExtensionCTAClicked', '{"page": "file"}', 'https://sourcegraph.test:3443/search', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			(3, 'InstallBrowserExtensionCTAShown', '{"page": "search"}', 'https://sourcegraph.test:3443/search', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			(4, 'InstallBrowserExtensionCTAClicked', '{"page": "search"}', 'https://sourcegraph.test:3443/search', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),

			-- Previous month event logs
			(5, 'InstallBrowserExtensionCTAShown', '{"page": "file"}', 'https://sourcegraph.test:3443/search', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 month'),
			(6, 'InstallBrowserExtensionCTAClicked', '{"page": "file"}', 'https://sourcegraph.test:3443/search', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 month'),
			(7, 'InstallBrowserExtensionCTAShown', '{"page": "search"}', 'https://sourcegraph.test:3443/search', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 month'),
			(8, 'InstallBrowserExtensionCTAClicked', '{"page": "search"}', 'https://sourcegraph.test:3443/search', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 month')
	`, now)
	if err != nil {
		t.Fatal(err)
	}

	have, err := GetCTAMetrics(ctx, db)
	if err != nil {
		t.Fatal(err)
	}

	want := &types.CTAMetrics{
		BextCtaShownCountOnFilePage:     1,
		BextCtaClickedCountOnFilePage:   1,
		BextCtaShownCountOnSearchPage:   1,
		BextCtaClickedCountOnSearchPage: 1,
	}
	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatal(diff)
	}
}
