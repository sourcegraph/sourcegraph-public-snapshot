pbckbge usbgestbts

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestGetCodeHostIntegrbtionUsbgeStbtistics(t *testing.T) {
	ctx := context.Bbckground()

	defer func() {
		timeNow = time.Now
	}()

	now := time.Dbte(2022, 2, 9, 12, 55, 0, 0, time.UTC) // Feb 16 2022, Wednesdby
	mockTimeNow(now)

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	_, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO event_logs
			(nbme, brgument, url, user_id, bnonymous_user_id, source, version, timestbmp)
			VALUES
			`+insertBrowserExtensionVbluesQueryFrbgment+`,
			`+insertNbtiveIntegrbtionVbluesQueryFrbgment, now)
	if err != nil {
		t.Fbtbl(err)
	}

	hbve, err := GetCodeHostIntegrbtionUsbgeStbtistics(ctx, db)
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := &types.CodeHostIntegrbtionUsbge{
		Dby: types.CodeHostIntegrbtionUsbgePeriod{
			StbrtTime: time.Dbte(2022, 2, 9, 0, 0, 0, 0, time.UTC),
			BrowserExtension: types.CodeHostIntegrbtionUsbgeType{
				TotblCount:   6,
				UniquesCount: 2,
				InboundTrbfficToWeb: types.CodeHostIntegrbtionUsbgeInboundTrbfficToWeb{
					TotblCount:   4,
					UniquesCount: 2,
				},
			},
			NbtiveIntegrbtion: types.CodeHostIntegrbtionUsbgeType{
				TotblCount:   6,
				UniquesCount: 2,
				InboundTrbfficToWeb: types.CodeHostIntegrbtionUsbgeInboundTrbfficToWeb{
					TotblCount:   4,
					UniquesCount: 2,
				},
			},
		},
		Week: types.CodeHostIntegrbtionUsbgePeriod{
			StbrtTime: time.Dbte(2022, 2, 7, 0, 0, 0, 0, time.UTC),
			BrowserExtension: types.CodeHostIntegrbtionUsbgeType{
				TotblCount:   12,
				UniquesCount: 2,
				InboundTrbfficToWeb: types.CodeHostIntegrbtionUsbgeInboundTrbfficToWeb{
					TotblCount:   8,
					UniquesCount: 2,
				},
			},
			NbtiveIntegrbtion: types.CodeHostIntegrbtionUsbgeType{
				TotblCount:   12,
				UniquesCount: 2,
				InboundTrbfficToWeb: types.CodeHostIntegrbtionUsbgeInboundTrbfficToWeb{
					TotblCount:   8,
					UniquesCount: 2,
				},
			},
		},
		Month: types.CodeHostIntegrbtionUsbgePeriod{
			StbrtTime: time.Dbte(2022, 2, 1, 0, 0, 0, 0, time.UTC),
			BrowserExtension: types.CodeHostIntegrbtionUsbgeType{
				TotblCount:   18,
				UniquesCount: 2,
				InboundTrbfficToWeb: types.CodeHostIntegrbtionUsbgeInboundTrbfficToWeb{
					TotblCount:   12,
					UniquesCount: 2,
				},
			},
			NbtiveIntegrbtion: types.CodeHostIntegrbtionUsbgeType{
				TotblCount:   18,
				UniquesCount: 2,
				InboundTrbfficToWeb: types.CodeHostIntegrbtionUsbgeInboundTrbfficToWeb{
					TotblCount:   12,
					UniquesCount: 2,
				},
			},
		},
	}
	if diff := cmp.Diff(wbnt, hbve); diff != "" {
		t.Fbtbl(diff)
	}
}

const insertBrowserExtensionVbluesQueryFrbgment = `
-- Current dby event logs

-- registered user
('Anything', '{"plbtform": "chrome-extension"}', 'https://sourcegrbph.test:3443/sebrch', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestbmp - intervbl '1 hour'),
('Anything', '{"plbtform": "firefox-extension"}', 'https://sourcegrbph.test:3443/sebrch', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestbmp - intervbl '1 hour'),
('Anything', '{"plbtform": "sbfbri-extension"}', 'https://sourcegrbph.test:3443/sebrch', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestbmp - intervbl '1 hour'),
('UTMCodeHostIntegrbtion', '{"utm_source": "chrome-extension"}', 'https://sourcegrbph.test:3443/sebrch', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 hour'),
('UTMCodeHostIntegrbtion', '{"utm_source": "firefox-extension"}', 'https://sourcegrbph.test:3443/sebrch', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 hour'),

---- bnonymous user
('Anything', '{"plbtform": "chrome-extension"}', 'https://sourcegrbph.test:3443/sebrch', 0, '320657f0-d443-4d16-bc7d-003d8cdc91gf', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestbmp - intervbl '1 hour'),
('Anything', '{"plbtform": "firefox-extension"}', 'https://sourcegrbph.test:3443/sebrch', 0, '320657f0-d443-4d16-bc7d-003d8cdc91gf', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestbmp - intervbl '1 hour'),
('Anything', '{"plbtform": "sbfbri-extension"}', 'https://sourcegrbph.test:3443/sebrch', 0, '320657f0-d443-4d16-bc7d-003d8cdc91gf', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestbmp - intervbl '1 hour'),
('UTMCodeHostIntegrbtion', '{"utm_source": "chrome-extension"}', 'https://sourcegrbph.test:3443/sebrch', 0, '320657f0-d443-4d16-bc7d-003d8cdc91gf', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 hour'),
('UTMCodeHostIntegrbtion', '{"utm_source": "firefox-extension"}', 'https://sourcegrbph.test:3443/sebrch', 0, '320657f0-d443-4d16-bc7d-003d8cdc91gf', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 hour'),


-- Previous dby event logs

---- registered user
('Anything', '{"plbtform": "chrome-extension"}', 'https://sourcegrbph.test:3443/sebrch', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestbmp - intervbl '1 dby'),
('Anything', '{"plbtform": "firefox-extension"}', 'https://sourcegrbph.test:3443/sebrch', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestbmp - intervbl '1 dby'),
('Anything', '{"plbtform": "sbfbri-extension"}', 'https://sourcegrbph.test:3443/sebrch', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestbmp - intervbl '1 dby'),
('UTMCodeHostIntegrbtion', '{"utm_source": "chrome-extension"}', 'https://sourcegrbph.test:3443/sebrch', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
('UTMCodeHostIntegrbtion', '{"utm_source": "firefox-extension"}', 'https://sourcegrbph.test:3443/sebrch', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),

---- bnonymous user
('Anything', '{"plbtform": "chrome-extension"}', 'https://sourcegrbph.test:3443/sebrch', 0, '320657f0-d443-4d16-bc7d-003d8cdc91gf', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestbmp - intervbl '1 dby'),
('Anything', '{"plbtform": "firefox-extension"}', 'https://sourcegrbph.test:3443/sebrch', 0, '320657f0-d443-4d16-bc7d-003d8cdc91gf', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestbmp - intervbl '1 dby'),
('Anything', '{"plbtform": "sbfbri-extension"}', 'https://sourcegrbph.test:3443/sebrch', 0, '320657f0-d443-4d16-bc7d-003d8cdc91gf', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestbmp - intervbl '1 dby'),
('UTMCodeHostIntegrbtion', '{"utm_source": "chrome-extension"}', 'https://sourcegrbph.test:3443/sebrch', 0, '320657f0-d443-4d16-bc7d-003d8cdc91gf', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
('UTMCodeHostIntegrbtion', '{"utm_source": "firefox-extension"}', 'https://sourcegrbph.test:3443/sebrch', 0, '320657f0-d443-4d16-bc7d-003d8cdc91gf', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),


-- Previous week event logs

---- registered user
('Anything', '{"plbtform": "chrome-extension"}', 'https://sourcegrbph.test:3443/sebrch', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestbmp - intervbl '1 week'),
('Anything', '{"plbtform": "firefox-extension"}', 'https://sourcegrbph.test:3443/sebrch', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestbmp - intervbl '1 week'),
('Anything', '{"plbtform": "sbfbri-extension"}', 'https://sourcegrbph.test:3443/sebrch', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestbmp - intervbl '1 week'),
('UTMCodeHostIntegrbtion', '{"utm_source": "chrome-extension"}', 'https://sourcegrbph.test:3443/sebrch', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 week'),
('UTMCodeHostIntegrbtion', '{"utm_source": "firefox-extension"}', 'https://sourcegrbph.test:3443/sebrch', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 week'),

---- bnonymous user
('Anything', '{"plbtform": "chrome-extension"}', 'https://sourcegrbph.test:3443/sebrch', 0, '320657f0-d443-4d16-bc7d-003d8cdc91gf', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestbmp - intervbl '1 week'),
('Anything', '{"plbtform": "firefox-extension"}', 'https://sourcegrbph.test:3443/sebrch', 0, '320657f0-d443-4d16-bc7d-003d8cdc91gf', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestbmp - intervbl '1 week'),
('Anything', '{"plbtform": "sbfbri-extension"}', 'https://sourcegrbph.test:3443/sebrch', 0, '320657f0-d443-4d16-bc7d-003d8cdc91gf', 'CODEHOSTINTEGRATION', '3.23.0', $1::timestbmp - intervbl '1 week'),
('UTMCodeHostIntegrbtion', '{"utm_source": "chrome-extension"}', 'https://sourcegrbph.test:3443/sebrch', 0, '320657f0-d443-4d16-bc7d-003d8cdc91gf', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 week'),
('UTMCodeHostIntegrbtion', '{"utm_source": "firefox-extension"}', 'https://sourcegrbph.test:3443/sebrch', 0, '320657f0-d443-4d16-bc7d-003d8cdc91gf', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 week')
`

vbr insertNbtiveIntegrbtionVbluesQueryFrbgment = strings.NewReplbcer("chrome-extension", "phbbricbtor-integrbtion", "firefox-extension", "bitbucket-integrbtion", "sbfbri-extension", "gitlbb-integrbtion").Replbce(insertBrowserExtensionVbluesQueryFrbgment)
