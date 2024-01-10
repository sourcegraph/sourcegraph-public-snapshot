package usagestats

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestGetCodeHostIntegrationUsageStatistics(t *testing.T) {
	ctx := context.Background()

	defer func() {
		timeNow = time.Now
	}()

	now := time.Date(2022, 2, 9, 12, 55, 0, 0, time.UTC) // Feb 16 2022, Wednesday
	mockTimeNow(now)

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))

	_, err := db.ExecContext(context.Background(), `
		INSERT INTO event_logs
			(name, argument, url, user_id, anonymous_user_id, source, version, timestamp)
			VALUES
			`+insertBrowserExtensionValuesQueryFragment+`,
			`+insertNativeIntegrationValuesQueryFragment, now)
	if err != nil {
		t.Fatal(err)
	}

	have, err := GetCodeHostIntegrationUsageStatistics(ctx, db)
	if err != nil {
		t.Fatal(err)
	}

	want := &types.CodeHostIntegrationUsage{
		Day: types.CodeHostIntegrationUsagePeriod{
			StartTime: time.Date(2022, 2, 9, 0, 0, 0, 0, time.UTC),
			BrowserExtension: types.CodeHostIntegrationUsageType{
				TotalCount:   6,
				UniquesCount: 2,
				InboundTrafficToWeb: types.CodeHostIntegrationUsageInboundTrafficToWeb{
					TotalCount:   4,
					UniquesCount: 2,
				},
			},
			NativeIntegration: types.CodeHostIntegrationUsageType{
				TotalCount:   6,
				UniquesCount: 2,
				InboundTrafficToWeb: types.CodeHostIntegrationUsageInboundTrafficToWeb{
					TotalCount:   4,
					UniquesCount: 2,
				},
			},
		},
		Week: types.CodeHostIntegrationUsagePeriod{
			StartTime: time.Date(2022, 2, 7, 0, 0, 0, 0, time.UTC),
			BrowserExtension: types.CodeHostIntegrationUsageType{
				TotalCount:   12,
				UniquesCount: 2,
				InboundTrafficToWeb: types.CodeHostIntegrationUsageInboundTrafficToWeb{
					TotalCount:   8,
					UniquesCount: 2,
				},
			},
			NativeIntegration: types.CodeHostIntegrationUsageType{
				TotalCount:   12,
				UniquesCount: 2,
				InboundTrafficToWeb: types.CodeHostIntegrationUsageInboundTrafficToWeb{
					TotalCount:   8,
					UniquesCount: 2,
				},
			},
		},
		Month: types.CodeHostIntegrationUsagePeriod{
			StartTime: time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC),
			BrowserExtension: types.CodeHostIntegrationUsageType{
				TotalCount:   18,
				UniquesCount: 2,
				InboundTrafficToWeb: types.CodeHostIntegrationUsageInboundTrafficToWeb{
					TotalCount:   12,
					UniquesCount: 2,
				},
			},
			NativeIntegration: types.CodeHostIntegrationUsageType{
				TotalCount:   18,
				UniquesCount: 2,
				InboundTrafficToWeb: types.CodeHostIntegrationUsageInboundTrafficToWeb{
					TotalCount:   12,
					UniquesCount: 2,
				},
			},
		},
	}
	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatal(diff)
	}
}

const insertBrowserExtensionValuesQueryFragment = `
-- Current day event logs

-- registered user
('Anything', '{"platform": "chrome-extension"}', 'https://sourcegraph.test:3443/search', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestamp - interval '1 hour'),
('Anything', '{"platform": "firefox-extension"}', 'https://sourcegraph.test:3443/search', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestamp - interval '1 hour'),
('Anything', '{"platform": "safari-extension"}', 'https://sourcegraph.test:3443/search', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestamp - interval '1 hour'),
('UTMCodeHostIntegration', '{"utm_source": "chrome-extension"}', 'https://sourcegraph.test:3443/search', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 hour'),
('UTMCodeHostIntegration', '{"utm_source": "firefox-extension"}', 'https://sourcegraph.test:3443/search', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 hour'),

---- anonymous user
('Anything', '{"platform": "chrome-extension"}', 'https://sourcegraph.test:3443/search', 0, '320657f0-d443-4d16-ac7d-003d8cdc91gf', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestamp - interval '1 hour'),
('Anything', '{"platform": "firefox-extension"}', 'https://sourcegraph.test:3443/search', 0, '320657f0-d443-4d16-ac7d-003d8cdc91gf', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestamp - interval '1 hour'),
('Anything', '{"platform": "safari-extension"}', 'https://sourcegraph.test:3443/search', 0, '320657f0-d443-4d16-ac7d-003d8cdc91gf', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestamp - interval '1 hour'),
('UTMCodeHostIntegration', '{"utm_source": "chrome-extension"}', 'https://sourcegraph.test:3443/search', 0, '320657f0-d443-4d16-ac7d-003d8cdc91gf', 'WEB', '3.23.0', $1::timestamp - interval '1 hour'),
('UTMCodeHostIntegration', '{"utm_source": "firefox-extension"}', 'https://sourcegraph.test:3443/search', 0, '320657f0-d443-4d16-ac7d-003d8cdc91gf', 'WEB', '3.23.0', $1::timestamp - interval '1 hour'),


-- Previous day event logs

---- registered user
('Anything', '{"platform": "chrome-extension"}', 'https://sourcegraph.test:3443/search', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestamp - interval '1 day'),
('Anything', '{"platform": "firefox-extension"}', 'https://sourcegraph.test:3443/search', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestamp - interval '1 day'),
('Anything', '{"platform": "safari-extension"}', 'https://sourcegraph.test:3443/search', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestamp - interval '1 day'),
('UTMCodeHostIntegration', '{"utm_source": "chrome-extension"}', 'https://sourcegraph.test:3443/search', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
('UTMCodeHostIntegration', '{"utm_source": "firefox-extension"}', 'https://sourcegraph.test:3443/search', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),

---- anonymous user
('Anything', '{"platform": "chrome-extension"}', 'https://sourcegraph.test:3443/search', 0, '320657f0-d443-4d16-ac7d-003d8cdc91gf', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestamp - interval '1 day'),
('Anything', '{"platform": "firefox-extension"}', 'https://sourcegraph.test:3443/search', 0, '320657f0-d443-4d16-ac7d-003d8cdc91gf', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestamp - interval '1 day'),
('Anything', '{"platform": "safari-extension"}', 'https://sourcegraph.test:3443/search', 0, '320657f0-d443-4d16-ac7d-003d8cdc91gf', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestamp - interval '1 day'),
('UTMCodeHostIntegration', '{"utm_source": "chrome-extension"}', 'https://sourcegraph.test:3443/search', 0, '320657f0-d443-4d16-ac7d-003d8cdc91gf', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
('UTMCodeHostIntegration', '{"utm_source": "firefox-extension"}', 'https://sourcegraph.test:3443/search', 0, '320657f0-d443-4d16-ac7d-003d8cdc91gf', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),


-- Previous week event logs

---- registered user
('Anything', '{"platform": "chrome-extension"}', 'https://sourcegraph.test:3443/search', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestamp - interval '1 week'),
('Anything', '{"platform": "firefox-extension"}', 'https://sourcegraph.test:3443/search', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestamp - interval '1 week'),
('Anything', '{"platform": "safari-extension"}', 'https://sourcegraph.test:3443/search', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestamp - interval '1 week'),
('UTMCodeHostIntegration', '{"utm_source": "chrome-extension"}', 'https://sourcegraph.test:3443/search', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 week'),
('UTMCodeHostIntegration', '{"utm_source": "firefox-extension"}', 'https://sourcegraph.test:3443/search', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 week'),

---- anonymous user
('Anything', '{"platform": "chrome-extension"}', 'https://sourcegraph.test:3443/search', 0, '320657f0-d443-4d16-ac7d-003d8cdc91gf', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestamp - interval '1 week'),
('Anything', '{"platform": "firefox-extension"}', 'https://sourcegraph.test:3443/search', 0, '320657f0-d443-4d16-ac7d-003d8cdc91gf', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestamp - interval '1 week'),
('Anything', '{"platform": "safari-extension"}', 'https://sourcegraph.test:3443/search', 0, '320657f0-d443-4d16-ac7d-003d8cdc91gf', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestamp - interval '1 week'),
('UTMCodeHostIntegration', '{"utm_source": "chrome-extension"}', 'https://sourcegraph.test:3443/search', 0, '320657f0-d443-4d16-ac7d-003d8cdc91gf', 'WEB', '3.23.0', $1::timestamp - interval '1 week'),
('UTMCodeHostIntegration', '{"utm_source": "firefox-extension"}', 'https://sourcegraph.test:3443/search', 0, '320657f0-d443-4d16-ac7d-003d8cdc91gf', 'WEB', '3.23.0', $1::timestamp - interval '1 week')
`

var insertNativeIntegrationValuesQueryFragment = strings.NewReplacer("chrome-extension", "phabricator-integration", "firefox-extension", "bitbucket-integration", "safari-extension", "gitlab-integration").Replace(insertBrowserExtensionValuesQueryFragment)
