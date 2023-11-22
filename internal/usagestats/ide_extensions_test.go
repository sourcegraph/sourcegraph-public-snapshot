package usagestats

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestIDEExtensionsUsageStatistics(t *testing.T) {
	ctx := context.Background()

	defer func() {
		timeNow = time.Now
	}()

	now := time.Date(2022, 2, 9, 12, 55, 0, 0, time.UTC) // Feb 16 2022, Wednesday
	mockTimeNow(now)
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))

	_, err := db.ExecContext(ctx, `
		INSERT INTO event_logs
			(id, name, argument, url, user_id, anonymous_user_id, source, timestamp, public_argument, version)
		VALUES
			(1, 'IDESearchSubmitted', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'IDEEXTENSION', $1::timestamp - interval '1 hour', '{"version": "2.0.8", "editor": "vscode"}', '3.36.1'),
			(2, 'VSCESearchSubmitted', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'BACKEND', $1::timestamp - interval '1 day', '{"version": "2.2.8", "editor": "vscode"}', '3.34.1'),
			(3, 'IDESearchSubmitted', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'IDEEXTENSION', $1::timestamp - interval '1 day', '{"version": "0.0.5", "editor": "jetbrains"}', '3.36.1'),
			(4, 'VSCESearchSubmitted', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'BACKEND', $1::timestamp - interval '1 day', '{"editor": ""}', '3.36.1'),
			(5, 'IDESearchSubmitted', '{"version": "0.0.1", "editor": "jetbrains"}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'IDEEXTENSION', $1::timestamp - interval '2 months', '{"version": "0.0.1", "editor": "jetbrains"}', '3.36.2'),
			(6, 'ViewBlob', '{}', 'https://sourcegraph.test:3443/github.com/sourcegraph/sourcegraph/-/blob/client/vscode/README.md?L8%3A1=', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', $1::timestamp - interval '7 days', '{}', '3.33.3'),
			(7, 'ViewBlob', '{}', 'https://sourcegraph.test:3443/github.com/sourcegraph/sourcegraph/-/blob/client/vscode/README.md?L8%3A1=', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', $1::timestamp - interval '10 days', '{}', '3.32.2'),
			(8, 'ViewBlob', '{}', 'https://sourcegraph.test:3443/github.com/sourcegraph/sourcegraph/-/blob/package.json', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', $1::timestamp - interval '2 months', '{"editor": ""}', '3.32.2'),
			(9, 'IDERedirected', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'IDEEXTENSION', $1::timestamp - interval '1 hour', '{"version": "0.0.1", "editor": "jetbrains"}', '3.35.0'),
			(10, 'IDERedirected', '{"version": "2.2.1", "editor": "vscode"}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'IDEEXTENSION', $1::timestamp - interval '3 hour', '{"version": "2.2.1", "editor": "vscode"}', '3.35.0'),
			(11, 'IDESearchSubmitted', '{"version": "2.0.8", "editor": "vscode"}', '', 3, '420657f0-d443-4d16-ac7d-003d8cdc24xy', 'IDEEXTENSION', $1::timestamp - interval '1 day', '{"version": "2.0.8", "editor": "vscode"}', '3.36.1'),
			(12, 'IDESearchSubmitted', '{"version": "2.0.9", "editor": "vscode"}', '', 3, '420657f0-d443-4d16-ac7d-003d8cdc24xy', 'IDEEXTENSION', $1::timestamp - interval '2 months', '{"version": "2.0.9", "editor": "vscode"}', '3.37.0'),
			(13, 'IDESearchSubmitted', '{"version": "2.0.9", "editor": "vscode"}', '', 3, '420657f0-d443-4d16-ac7d-003d8cdc24xy', 'IDEEXTENSION', $1::timestamp - interval '1 week', '{"version": "2.0.9", "editor": "vscode"}', '3.37.0'),
			(14, 'IDERedirected', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'IDEEXTENSION', $1::timestamp - interval '1 week', '{"version": "2.2.1", "editor": "vscode"}', '3.35.0'),
			(15, 'ViewBlob', '{}', 'https://sourcegraph.test:3443/github.com/sourcegraph/sourcegraph/-/blob/client/vscode/README.md?L8%3A1=', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', $1::timestamp - interval '1 day', '{}', '3.35.0'),
			(16, 'VSCESearchSubmitted', '{}', '', 4, '420657f0-d443-6d16-ac7d-003d8cdcmf9y', 'BACKEND', $1::timestamp - interval '32 days', '{"editor": ""}', '3.37.1'),
			(17, 'ViewBlob', '{}', 'https://sourcegraph.test:3443/github.com/sourcegraph/sourcegraph/-/blob/client/vscode/README.md?L8%3A1=', 5, '420657f0-13d3-fgw3-ac7d-123d8cdm2sap', 'WEB', $1::timestamp - interval '1 day', '{}', '3.35.0'),
			(18, 'IDEUninstalled', '{}', '', 3, '420657f0-d443-4d16-ac7d-003d8cdc24xy', 'IDEEXTENSION', $1::timestamp - interval '2 hours', '{"version": "2.0.9", "editor": "vscode"}', '3.37.0'),
			(19, 'IDEInstalled', '{}', '', 5, '420612f0-t4se-4bd6-123d-lf83iufdc2445', 'BACKEND', $1::timestamp - interval '5 hours', '{"version": "2.2.0", "editor": "vscode"}', '3.34.0'),
			(20, 'IDEUninstalled', '{}', '', 5, '420612f0-t4se-4bd6-123d-lf83iufdc2445', 'BACKEND', $1::timestamp - interval '2 hours', '{"version": "2.2.0", "editor": "vscode"}', '3.34.0')
	`, now)
	if err != nil {
		t.Fatal(err)
	}

	have, err := GetIDEExtensionsUsageStatistics(ctx, db)
	if err != nil {
		t.Fatal(err)
	}

	// Older versions of VSCE do not log editor name so we can assume object without IdeKind comes from VSCE
	want := &types.IDEExtensionsUsage{
		IDEs: []*types.IDEExtensionsUsageStatistics{
			{
				Month: types.IDEExtensionsUsageRegularPeriod{
					StartTime: time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC),
					SearchesPerformed: types.IDEExtensionsUsageSearchesPerformed{
						UniquesCount: int32(1),
						TotalCount:   int32(1),
					},
				},
				Week: types.IDEExtensionsUsageRegularPeriod{
					StartTime: time.Date(2022, 2, 7, 0, 0, 0, 0, time.UTC),
					SearchesPerformed: types.IDEExtensionsUsageSearchesPerformed{
						UniquesCount: int32(1),
						TotalCount:   int32(1),
					},
				},
				Day: types.IDEExtensionsUsageDailyPeriod{
					StartTime: time.Date(2022, 2, 9, 0, 0, 0, 0, time.UTC),
				},
			},
			{
				IdeKind: "jetbrains",
				Month: types.IDEExtensionsUsageRegularPeriod{
					StartTime: time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC),
					SearchesPerformed: types.IDEExtensionsUsageSearchesPerformed{
						UniquesCount: int32(1),
						TotalCount:   int32(1),
					},
				},
				Week: types.IDEExtensionsUsageRegularPeriod{
					StartTime: time.Date(2022, 2, 7, 0, 0, 0, 0, time.UTC),
					SearchesPerformed: types.IDEExtensionsUsageSearchesPerformed{
						UniquesCount: int32(1),
						TotalCount:   int32(1),
					},
				},
				Day: types.IDEExtensionsUsageDailyPeriod{
					StartTime: time.Date(2022, 2, 9, 0, 0, 0, 0, time.UTC),
					SearchesPerformed: types.IDEExtensionsUsageSearchesPerformed{
						UniquesCount: int32(0),
						TotalCount:   int32(0),
					},
					UserState: types.IDEExtensionsUsageUserState{
						Installs:   int32(0),
						Uninstalls: int32(0),
					},
					RedirectsCount: int32(1),
				},
			},
			{
				IdeKind: "vscode",
				Month: types.IDEExtensionsUsageRegularPeriod{
					StartTime: time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC),
					SearchesPerformed: types.IDEExtensionsUsageSearchesPerformed{
						UniquesCount: int32(2),
						TotalCount:   int32(4),
					},
				},
				Week: types.IDEExtensionsUsageRegularPeriod{
					StartTime: time.Date(2022, 2, 7, 0, 0, 0, 0, time.UTC),
					SearchesPerformed: types.IDEExtensionsUsageSearchesPerformed{
						UniquesCount: int32(2),
						TotalCount:   int32(3),
					},
				},
				Day: types.IDEExtensionsUsageDailyPeriod{
					StartTime: time.Date(2022, 2, 9, 0, 0, 0, 0, time.UTC),
					SearchesPerformed: types.IDEExtensionsUsageSearchesPerformed{
						UniquesCount: int32(1),
						TotalCount:   int32(1),
					},
					UserState: types.IDEExtensionsUsageUserState{
						Installs:   int32(1),
						Uninstalls: int32(2),
					},
					RedirectsCount: int32(1),
				},
			},
		},
	}
	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatal(diff)
	}
}
