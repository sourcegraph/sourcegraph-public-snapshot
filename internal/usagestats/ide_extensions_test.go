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

func TestIdeExtensionsUsageStatistics(t *testing.T) {
	ctx := context.Background()

	defer func() {
		timeNow = time.Now
	}()

	// monthStart := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2022, 2, 9, 12, 55, 0, 0, time.UTC) // Feb 16 2022, Wednesday
	mockTimeNow(now)

	db := database.NewDB(dbtest.NewDB(t))

	// TODO add better events
	_, err := db.ExecContext(ctx, `
		INSERT INTO event_logs
			(id, name, argument, url, user_id, anonymous_user_id, source, timestamp, public_argument, version)
		VALUES
			(1, 'IDESearchSubmitted', '{"version": "2.0.8", "platform": "vscode"}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'IDEEXTENSION', $1::timestamp - interval '1 hour', '{}', '3.36.1'),
			(2, 'VSCESearchSubmitted', '{"version": "2.2.8", "platform": "vscode"}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'BACKEND', $1::timestamp - interval '1 day', '{}', '3.34.1'),
			(3, 'IDESearchSubmitted', '{"version": "0.0.5", "platform": "jetbrains"}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'IDEEXTENSION', $1::timestamp - interval '1 day', '{}', '3.36.1'),
			(4, 'VSCESearchSubmitted', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'BACKEND', $1::timestamp - interval '1 day', '{}', '3.36.1'),
			(5, 'IDESearchSubmitted', '{"version": "0.0.1", "platform": "jetbrains"}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'IDEEXTENSION', $1::timestamp - interval '2 months', '{}', '3.36.2'),
			(6, 'ViewBlob', '{}', 'https://sourcegraph.test:3443/github.com/sourcegraph/sourcegraph/-/blob/client/vscode/README.md?L8%3A1=&utm_source=VSCode-2.0.9', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', $1::timestamp - interval '7 days', '{}', '3.33.3'),
			(7, 'ViewBlob', '{}', 'https://sourcegraph.test:3443/github.com/sourcegraph/sourcegraph/-/blob/client/vscode/README.md?L8%3A1=&utm_source=VSCode-2.0.5', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', $1::timestamp - interval '10 days', '{}', '3.32.2'),
			(8, 'ViewBlob', '{}', 'https://sourcegraph.test:3443/github.com/sourcegraph/sourcegraph/-/blob/package.json', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', $1::timestamp - interval '2 months', '{}', '3.32.2'),
			(9, 'IDERedirects', '{"version": "0.0.1", "platform": "jetbrains"}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'IDEEXTENSION', $1::timestamp - interval '1 hour', '{}', '3.35.0'),
			(10, 'IDERedirects', '{"version": "2.2.1", "platform": "vscode"}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'IDEEXTENSION', $1::timestamp - interval '3 hour', '{}', '3.35.0'),
			(11, 'IDESearchSubmitted', '{"version": "2.0.8", "platform": "vscode"}', '', 3, '420657f0-d443-4d16-ac7d-003d8cdc24xy', 'IDEEXTENSION', $1::timestamp - interval '1 day', '{}', '3.36.1'),
			(12, 'IDESearchSubmitted', '{"version": "2.0.9", "platform": "vscode"}', '', 3, '420657f0-d443-4d16-ac7d-003d8cdc24xy', 'IDEEXTENSION', $1::timestamp - interval '2 months', '{}', '3.37.0'),
			(13, 'IDESearchSubmitted', '{"version": "2.0.9", "platform": "vscode"}', '', 3, '420657f0-d443-4d16-ac7d-003d8cdc24xy', 'IDEEXTENSION', $1::timestamp - interval '1 week', '{}', '3.37.0'),
			(14, 'IDERedirects', '{"version": "2.2.1", "platform": "vscode"}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'IDEEXTENSION', $1::timestamp - interval '1 week', '{}', '3.35.0'),
			(15, 'ViewBlob', '{}', 'https://sourcegraph.test:3443/github.com/sourcegraph/sourcegraph/-/blob/client/vscode/README.md?L8%3A1=&utm_source=VSCode-2.0.9', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', $1::timestamp - interval '1 day', '{}', '3.35.0'),
			(16, 'VSCESearchSubmitted', '{}', '', 4, '420657f0-d443-6d16-ac7d-003d8cdcmf9y', 'BACKEND', $1::timestamp - interval '32 days', '{}', '3.37.1'),
			(17, 'ViewBlob', '{}', 'https://sourcegraph.test:3443/github.com/sourcegraph/sourcegraph/-/blob/client/vscode/README.md?L8%3A1=&utm_source=VSCode-2.0.9', 5, '420657f0-13d3-fgw3-ac7d-123d8cdm2sap', 'WEB', $1::timestamp - interval '1 day', '{}', '3.35.0'),
			(18, 'IDEUninstalled', '{"version": "2.0.9", "platform": "vscode"}', '', 3, '420657f0-d443-4d16-ac7d-003d8cdc24xy', 'IDEEXTENSION', $1::timestamp - interval '2 hours', '{}', '3.37.0')
	`, now)
	if err != nil {
		t.Fatal(err)
	}

	have, err := GetIdeExtensionsUsageStatistics(ctx, db)
	if err != nil {
		t.Fatal(err)
	}

	want := &types.IdeExtensionsUsage{
		IDEs: []*types.IdeExtensionsUsageStatistics{
			{
				IdeKind: "jetbrains",
				Month: types.IdeExtensionsUsagePeriod{
					StartTime: time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC),
					UserCount: int32(1),
					SearchPerformed: types.IdeExtensionsUsageSearchPerformed{
						UniqueCount: int32(1),
						TotalCount:  int32(1),
					},
					MonthlyUserState: types.IdeExtensionsUserStateMonthlyCount{
						Installs: int32(1),
					},
					RedirectCount: int32(1),
				},
				Week: types.IdeExtensionsUsagePeriod{
					StartTime: time.Date(2022, 2, 7, 0, 0, 0, 0, time.UTC),
					UserCount: int32(1),
					SearchPerformed: types.IdeExtensionsUsageSearchPerformed{
						UniqueCount: int32(1),
						TotalCount:  int32(1),
					},
					RedirectCount: int32(1),
				},
				Day: types.IdeExtensionsUsagePeriod{
					StartTime: time.Date(2022, 2, 9, 0, 0, 0, 0, time.UTC),
					UserCount: int32(1),
					SearchPerformed: types.IdeExtensionsUsageSearchPerformed{
						UniqueCount: int32(0),
						TotalCount:  int32(0),
					},
					RedirectCount: int32(1),
				},
			},
			{
				IdeKind: "vscode",
				Month: types.IdeExtensionsUsagePeriod{
					StartTime: time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC),
					UserCount: int32(3),
					SearchPerformed: types.IdeExtensionsUsageSearchPerformed{
						UniqueCount: int32(2),
						TotalCount:  int32(3),
					},
					MonthlyUserState: types.IdeExtensionsUserStateMonthlyCount{
						Installs:   int32(3),
						Uninstalls: int32(1),
					},
					RedirectCount: int32(2),
				},
				Week: types.IdeExtensionsUsagePeriod{
					StartTime: time.Date(2022, 2, 7, 0, 0, 0, 0, time.UTC),
					UserCount: int32(3),
					SearchPerformed: types.IdeExtensionsUsageSearchPerformed{
						UniqueCount: int32(2),
						TotalCount:  int32(2),
					},
					RedirectCount: int32(1),
				},
				Day: types.IdeExtensionsUsagePeriod{
					StartTime: time.Date(2022, 2, 9, 0, 0, 0, 0, time.UTC),
					UserCount: int32(3),
					SearchPerformed: types.IdeExtensionsUsageSearchPerformed{
						UniqueCount: int32(1),
						TotalCount:  int32(1),
					},
					RedirectCount: int32(1),
				},
			},
			{
				IdeKind: "vscode_pre_ide",
				Month: types.IdeExtensionsUsagePeriod{
					StartTime: time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC),
					UserCount: int32(3),
					SearchPerformed: types.IdeExtensionsUsageSearchPerformed{
						UniqueCount: int32(2),
						TotalCount:  int32(2),
					},
					MonthlyUserState: types.IdeExtensionsUserStateMonthlyCount{
						Installs: int32(3),
					},
					RedirectCount: int32(3),
				},
				Week: types.IdeExtensionsUsagePeriod{
					StartTime: time.Date(2022, 2, 7, 0, 0, 0, 0, time.UTC),
					UserCount: int32(3),
					SearchPerformed: types.IdeExtensionsUsageSearchPerformed{
						UniqueCount: int32(2),
						TotalCount:  int32(2),
					},
					RedirectCount: int32(2),
				},
				Day: types.IdeExtensionsUsagePeriod{
					StartTime: time.Date(2022, 2, 9, 0, 0, 0, 0, time.UTC),
					UserCount: int32(0),
				},
			},
		},
	}
	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatal(diff)
	}
}
