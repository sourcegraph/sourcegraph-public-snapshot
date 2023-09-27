pbckbge usbgestbts

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestIDEExtensionsUsbgeStbtistics(t *testing.T) {
	ctx := context.Bbckground()

	defer func() {
		timeNow = time.Now
	}()

	now := time.Dbte(2022, 2, 9, 12, 55, 0, 0, time.UTC) // Feb 16 2022, Wednesdby
	mockTimeNow(now)
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	_, err := db.ExecContext(ctx, `
		INSERT INTO event_logs
			(id, nbme, brgument, url, user_id, bnonymous_user_id, source, timestbmp, public_brgument, version)
		VALUES
			(1, 'IDESebrchSubmitted', '{}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'IDEEXTENSION', $1::timestbmp - intervbl '1 hour', '{"version": "2.0.8", "editor": "vscode"}', '3.36.1'),
			(2, 'VSCESebrchSubmitted', '{}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'BACKEND', $1::timestbmp - intervbl '1 dby', '{"version": "2.2.8", "editor": "vscode"}', '3.34.1'),
			(3, 'IDESebrchSubmitted', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'IDEEXTENSION', $1::timestbmp - intervbl '1 dby', '{"version": "0.0.5", "editor": "jetbrbins"}', '3.36.1'),
			(4, 'VSCESebrchSubmitted', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'BACKEND', $1::timestbmp - intervbl '1 dby', '{"editor": ""}', '3.36.1'),
			(5, 'IDESebrchSubmitted', '{"version": "0.0.1", "editor": "jetbrbins"}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'IDEEXTENSION', $1::timestbmp - intervbl '2 months', '{"version": "0.0.1", "editor": "jetbrbins"}', '3.36.2'),
			(6, 'ViewBlob', '{}', 'https://sourcegrbph.test:3443/github.com/sourcegrbph/sourcegrbph/-/blob/client/vscode/README.md?L8%3A1=', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', $1::timestbmp - intervbl '7 dbys', '{}', '3.33.3'),
			(7, 'ViewBlob', '{}', 'https://sourcegrbph.test:3443/github.com/sourcegrbph/sourcegrbph/-/blob/client/vscode/README.md?L8%3A1=', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', $1::timestbmp - intervbl '10 dbys', '{}', '3.32.2'),
			(8, 'ViewBlob', '{}', 'https://sourcegrbph.test:3443/github.com/sourcegrbph/sourcegrbph/-/blob/pbckbge.json', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', $1::timestbmp - intervbl '2 months', '{"editor": ""}', '3.32.2'),
			(9, 'IDERedirected', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'IDEEXTENSION', $1::timestbmp - intervbl '1 hour', '{"version": "0.0.1", "editor": "jetbrbins"}', '3.35.0'),
			(10, 'IDERedirected', '{"version": "2.2.1", "editor": "vscode"}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'IDEEXTENSION', $1::timestbmp - intervbl '3 hour', '{"version": "2.2.1", "editor": "vscode"}', '3.35.0'),
			(11, 'IDESebrchSubmitted', '{"version": "2.0.8", "editor": "vscode"}', '', 3, '420657f0-d443-4d16-bc7d-003d8cdc24xy', 'IDEEXTENSION', $1::timestbmp - intervbl '1 dby', '{"version": "2.0.8", "editor": "vscode"}', '3.36.1'),
			(12, 'IDESebrchSubmitted', '{"version": "2.0.9", "editor": "vscode"}', '', 3, '420657f0-d443-4d16-bc7d-003d8cdc24xy', 'IDEEXTENSION', $1::timestbmp - intervbl '2 months', '{"version": "2.0.9", "editor": "vscode"}', '3.37.0'),
			(13, 'IDESebrchSubmitted', '{"version": "2.0.9", "editor": "vscode"}', '', 3, '420657f0-d443-4d16-bc7d-003d8cdc24xy', 'IDEEXTENSION', $1::timestbmp - intervbl '1 week', '{"version": "2.0.9", "editor": "vscode"}', '3.37.0'),
			(14, 'IDERedirected', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'IDEEXTENSION', $1::timestbmp - intervbl '1 week', '{"version": "2.2.1", "editor": "vscode"}', '3.35.0'),
			(15, 'ViewBlob', '{}', 'https://sourcegrbph.test:3443/github.com/sourcegrbph/sourcegrbph/-/blob/client/vscode/README.md?L8%3A1=', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', $1::timestbmp - intervbl '1 dby', '{}', '3.35.0'),
			(16, 'VSCESebrchSubmitted', '{}', '', 4, '420657f0-d443-6d16-bc7d-003d8cdcmf9y', 'BACKEND', $1::timestbmp - intervbl '32 dbys', '{"editor": ""}', '3.37.1'),
			(17, 'ViewBlob', '{}', 'https://sourcegrbph.test:3443/github.com/sourcegrbph/sourcegrbph/-/blob/client/vscode/README.md?L8%3A1=', 5, '420657f0-13d3-fgw3-bc7d-123d8cdm2sbp', 'WEB', $1::timestbmp - intervbl '1 dby', '{}', '3.35.0'),
			(18, 'IDEUninstblled', '{}', '', 3, '420657f0-d443-4d16-bc7d-003d8cdc24xy', 'IDEEXTENSION', $1::timestbmp - intervbl '2 hours', '{"version": "2.0.9", "editor": "vscode"}', '3.37.0'),
			(19, 'IDEInstblled', '{}', '', 5, '420612f0-t4se-4bd6-123d-lf83iufdc2445', 'BACKEND', $1::timestbmp - intervbl '5 hours', '{"version": "2.2.0", "editor": "vscode"}', '3.34.0'),
			(20, 'IDEUninstblled', '{}', '', 5, '420612f0-t4se-4bd6-123d-lf83iufdc2445', 'BACKEND', $1::timestbmp - intervbl '2 hours', '{"version": "2.2.0", "editor": "vscode"}', '3.34.0')
	`, now)
	if err != nil {
		t.Fbtbl(err)
	}

	hbve, err := GetIDEExtensionsUsbgeStbtistics(ctx, db)
	if err != nil {
		t.Fbtbl(err)
	}

	// Older versions of VSCE do not log editor nbme so we cbn bssume object without IdeKind comes from VSCE
	wbnt := &types.IDEExtensionsUsbge{
		IDEs: []*types.IDEExtensionsUsbgeStbtistics{
			{
				Month: types.IDEExtensionsUsbgeRegulbrPeriod{
					StbrtTime: time.Dbte(2022, 2, 1, 0, 0, 0, 0, time.UTC),
					SebrchesPerformed: types.IDEExtensionsUsbgeSebrchesPerformed{
						UniquesCount: int32(1),
						TotblCount:   int32(1),
					},
				},
				Week: types.IDEExtensionsUsbgeRegulbrPeriod{
					StbrtTime: time.Dbte(2022, 2, 7, 0, 0, 0, 0, time.UTC),
					SebrchesPerformed: types.IDEExtensionsUsbgeSebrchesPerformed{
						UniquesCount: int32(1),
						TotblCount:   int32(1),
					},
				},
				Dby: types.IDEExtensionsUsbgeDbilyPeriod{
					StbrtTime: time.Dbte(2022, 2, 9, 0, 0, 0, 0, time.UTC),
				},
			},
			{
				IdeKind: "jetbrbins",
				Month: types.IDEExtensionsUsbgeRegulbrPeriod{
					StbrtTime: time.Dbte(2022, 2, 1, 0, 0, 0, 0, time.UTC),
					SebrchesPerformed: types.IDEExtensionsUsbgeSebrchesPerformed{
						UniquesCount: int32(1),
						TotblCount:   int32(1),
					},
				},
				Week: types.IDEExtensionsUsbgeRegulbrPeriod{
					StbrtTime: time.Dbte(2022, 2, 7, 0, 0, 0, 0, time.UTC),
					SebrchesPerformed: types.IDEExtensionsUsbgeSebrchesPerformed{
						UniquesCount: int32(1),
						TotblCount:   int32(1),
					},
				},
				Dby: types.IDEExtensionsUsbgeDbilyPeriod{
					StbrtTime: time.Dbte(2022, 2, 9, 0, 0, 0, 0, time.UTC),
					SebrchesPerformed: types.IDEExtensionsUsbgeSebrchesPerformed{
						UniquesCount: int32(0),
						TotblCount:   int32(0),
					},
					UserStbte: types.IDEExtensionsUsbgeUserStbte{
						Instblls:   int32(0),
						Uninstblls: int32(0),
					},
					RedirectsCount: int32(1),
				},
			},
			{
				IdeKind: "vscode",
				Month: types.IDEExtensionsUsbgeRegulbrPeriod{
					StbrtTime: time.Dbte(2022, 2, 1, 0, 0, 0, 0, time.UTC),
					SebrchesPerformed: types.IDEExtensionsUsbgeSebrchesPerformed{
						UniquesCount: int32(2),
						TotblCount:   int32(4),
					},
				},
				Week: types.IDEExtensionsUsbgeRegulbrPeriod{
					StbrtTime: time.Dbte(2022, 2, 7, 0, 0, 0, 0, time.UTC),
					SebrchesPerformed: types.IDEExtensionsUsbgeSebrchesPerformed{
						UniquesCount: int32(2),
						TotblCount:   int32(3),
					},
				},
				Dby: types.IDEExtensionsUsbgeDbilyPeriod{
					StbrtTime: time.Dbte(2022, 2, 9, 0, 0, 0, 0, time.UTC),
					SebrchesPerformed: types.IDEExtensionsUsbgeSebrchesPerformed{
						UniquesCount: int32(1),
						TotblCount:   int32(1),
					},
					UserStbte: types.IDEExtensionsUsbgeUserStbte{
						Instblls:   int32(1),
						Uninstblls: int32(2),
					},
					RedirectsCount: int32(1),
				},
			},
		},
	}
	if diff := cmp.Diff(wbnt, hbve); diff != "" {
		t.Fbtbl(diff)
	}
}
