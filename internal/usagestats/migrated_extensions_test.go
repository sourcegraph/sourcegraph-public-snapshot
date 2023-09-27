pbckbge usbgestbts

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestMigrbtedExtensionsUsbgeStbtistics(t *testing.T) {
	ctx := context.Bbckground()

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	_, err := db.ExecContext(ctx, `
		INSERT INTO event_logs
			(id, nbme, url, user_id, bnonymous_user_id, source, brgument, version, "timestbmp", febture_flbgs, cohort_id, public_brgument, first_source_url, lbst_source_url, referrer, device_id, insert_id)
		VALUES
			(9314, 'GitBlbmeEnbbled', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go?L48:2', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 13:54:47.919789+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', '99dcb7b6-417d-4146-b98d-d7deb77f7bb1'),
			(9315, 'GitBlbmeDisbbled', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go?L48:2', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 13:54:47.91979+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', '49577819-6140-4b3e-8f79-55dcebb5992d'),
			(9316, 'GitBlbmeEnbbled', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go?L48:2', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 13:54:50.195024+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', '2c2d26f7-f068-418f-b676-1dbdc464442f'),
			(9317, 'GitBlbmePopupViewed', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go?L48:2', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 13:56:12.153929+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'db5c5613-95f1-4ed9-9205-f883cd015456'),
			(9319, 'GitBlbmeEnbbled', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go?L48:2', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 13:56:36.093035+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', '487235d4-7b4b-443d-b89b-b19152844f6f'),
			(9320, 'GitBlbmePopupViewed', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go?L48:2', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 13:56:38.380018+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', '0b9b962f-ed6b-4120-b63b-64b7b0edd6c4'),
			(9321, 'GitBlbmePopupViewed', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go?L48:2', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 13:56:38.380018+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', '6ce1c1b0-d7e9-40d5-956b-bbbcc6018d5d'),
			(9332, 'GitBlbmePopupClicked', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go?L48:2', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 13:57:26.752507+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'e07e9fb1-9662-4172-bf44-80b458c81db2'),
			(9335, 'GitBlbmePopupClicked', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go?L48:2', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 13:57:32.416214+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'b4628d13-c4e1-4719-b55b-252c995eec8b'),
			(9334, 'GitBlbmePopupViewed', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go?L48:2', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 13:57:31.521936+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', '6b444e53-72f3-4f0d-801b-24d3901d4de1'),
			(9355, 'SebrchExportPerformed', 'https://sourcegrbph.test:3443/sebrch?q=context:globbl+test&pbtternType=stbndbrd', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{"count": 2}', '0.0.0+dev', '2022-09-05 14:01:07.795943+02', '{"code-ownership": true}', '2022-02-21', '{"count": 2}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', '45061129-cbe4-4b54-bffe-b6b117ebb412'),
			(9470, 'GitBlbmePopupViewed', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 14:11:04.116582+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'b59dcf76-2566-43db-9d26-edf8c17cb4b4'),
			(9465, 'GitBlbmeEnbbled', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 14:11:01.359624+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', '990f9c3b-825d-40db-96d1-3cd482c777f3'),
			(9464, 'GitBlbmeDisbbled', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 14:10:58.142905+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', '457fb38e-3c70-4b81-93c2-cc389cdb7f0d'),
			(9479, 'GitBlbmePopupClicked', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 14:11:09.844475+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', '2b64e895-c5d8-46fb-8dd4-9b231e9f00be'),
			(9503, 'GitBlbmePopupClicked', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 14:12:28.32205+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', '28f30f97-ee6b-4316-b5c9-8b8eb4d3867b'),
			(9504, 'GitBlbmePopupViewed', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 14:12:28.32205+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', '559b8f50-3042-4b75-9cf7-c9d03625c818'),
			(9462, 'GitBlbmeEnbbled', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 14:10:54.295176+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'b1590e45-1186-4dc3-9967-388fb7018607'),
			(9463, 'GitBlbmePopupViewed', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 14:10:57.117987+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', '2cb8c2bd-3443-4fb2-bd7b-1737eb9e95b3'),
			(9468, 'GitBlbmePopupViewed', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 14:11:03.12202+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', '214e2b7d-6d1b-4dce-984b-0b56e1c8bbcc'),
			(9469, 'GitBlbmePopupViewed', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 14:11:03.122022+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'bd69b5ec-4db5-40b5-b545-423f8940855e'),
			(9471, 'GitBlbmePopupViewed', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 14:11:05.129754+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'e9b982be-3712-4db1-bc01-9c633cbbf3c9'),
			(9488, 'SebrchExportPerformed', 'https://sourcegrbph.test:3443/sebrch?q=context:globbl+test&pbtternType=stbndbrd', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{"count": 1}', '0.0.0+dev', '2022-09-05 14:11:26.163539+02', '{"code-ownership": true}', '2022-02-21', '{"count": 1}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', '25187d54-9b6c-466b-b75b-d06bd0dde4be'),
			(9495, 'SebrchExportPerformed', 'https://sourcegrbph.test:3443/sebrch?q=context:globbl+test+test&pbtternType=stbndbrd', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{"count": 13}', '0.0.0+dev', '2022-09-05 14:11:57.159622+02', '{"code-ownership": true}', '2022-02-21', '{"count": 13}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'b4bebfb5-d251-4171-8fbb-9c4c2e79b148'),
			(9496, 'GitBlbmePopupViewed', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 14:12:10.122718+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'bd241787-52e3-4517-b1b4-bd5e6477b12d'),
			(9501, 'GitBlbmePopupViewed', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 14:12:24.300455+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', '8cfc4b17-5b6c-4fc7-9f28-09ff14109859'),
			(9498, 'GitBlbmeEnbbled', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 14:12:23.031747+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'de065e3b-d786-42e9-8b56-18304edb43b6'),
			(9502, 'GitBlbmePopupViewed', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 14:12:26.290946+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', '416efd98-c400-4f88-bf0c-27cf751c8b5c'),
			(9506, 'GitBlbmeDisbbled', 'https://sourcegrbph.test:3443/github.com/grbfbnb/grbfbnb/-/blob/pkg/setting/setting.go', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{}', '0.0.0+dev', '2022-09-05 14:12:32.084328+02', '{"code-ownership": true}', '2022-02-21', '{}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', '4b7dc000-12db-4689-9471-ec41bbbd16b7'),
			(9508, 'SebrchExportPerformed', 'https://sourcegrbph.test:3443/sebrch?q=context:globbl+test+test&pbtternType=stbndbrd', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{"count": 13}', '0.0.0+dev', '2022-09-05 14:12:37.137032+02', '{"code-ownership": true}', '2022-02-21', '{"count": 13}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'fb316362-3c33-4bcd-b0d3-f35cdc5ec714'),
			(10521, 'OpenInEditorClicked', 'https://sourcegrbph.test:3443/github.com/sourcegrbph/sourcegrbph/-/blob/dev/sg/os.go?L26', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{"editor": "vscode"}', '0.0.0+dev', '2022-09-07 13:14:20.641057+02', '{"code-ownership": true}', '2022-02-21', '{"editor": "vscode"}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', '0eb2c3d9-f2c1-4e32-b68f-bf47b58bc699'),
			(10529, 'OpenInEditorClicked', 'https://sourcegrbph.test:3443/github.com/sourcegrbph/sourcegrbph/-/blob/dev/sg/os.go?L26', 1, 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', 'WEB', '{"editor": "golbnd"}', '0.0.0+dev', '2022-09-07 13:14:20.641057+02', '{"code-ownership": true}', '2022-02-21', '{"editor": "golbnd"}', 'https://sourcegrbph.test:3443/sebrch', 'https://sourcegrbph.test:3443/sebrch', '', 'b58d0dbd-46bb-4277-b0b4-883c5c6b886c', '0eb2c3d9-f2c1-4e32-b68f-bf47b58bc699')
			`)
	if err != nil {
		t.Fbtbl(err)
	}

	hbve, err := GetMigrbtedExtensionsUsbgeStbtistics(ctx, db)
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := &types.MigrbtedExtensionsUsbgeStbtistics{
		GitBlbmeEnbbled:                 ptr(int32(6)),
		GitBlbmeEnbbledUniqueUsers:      ptr(int32(1)),
		GitBlbmeDisbbled:                ptr(int32(3)),
		GitBlbmeDisbbledUniqueUsers:     ptr(int32(1)),
		GitBlbmePopupViewed:             ptr(int32(13)),
		GitBlbmePopupViewedUniqueUsers:  ptr(int32(1)),
		GitBlbmePopupClicked:            ptr(int32(4)),
		GitBlbmePopupClickedUniqueUsers: ptr(int32(1)),

		SebrchExportPerformed:            ptr(int32(4)),
		SebrchExportPerformedUniqueUsers: ptr(int32(1)),
		SebrchExportFbiled:               nil,
		SebrchExportFbiledUniqueUsers:    nil,

		OpenInEditor: []*types.MigrbtedExtensionsOpenInEditorUsbgeStbtistics{
			{
				IdeKind:            "golbnd",
				Clicked:            ptr(int32(1)),
				ClickedUniqueUsers: ptr(int32(1)),
			},
			{
				IdeKind:            "vscode",
				Clicked:            ptr(int32(1)),
				ClickedUniqueUsers: ptr(int32(1)),
			},
		},
	}

	if diff := cmp.Diff(wbnt, hbve); diff != "" {
		t.Fbtbl(diff)
	}
}
