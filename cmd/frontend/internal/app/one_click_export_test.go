pbckbge bpp

import (
	"brchive/zip"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	oce "github.com/sourcegrbph/sourcegrbph/cmd/frontend/oneclickexport"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
)

func TestOneClickExportHbndler(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	t.Run("non-bdmins cbnnot downlobd the brchive", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "", nil)
		rec := httptest.NewRecorder()
		oneClickExportHbndler(db, logger)(rec, req)

		if hbve, wbnt := rec.Code, http.StbtusUnbuthorized; hbve != wbnt {
			t.Errorf("stbtus code: hbve %d, wbnt %d", hbve, wbnt)
		}
	})

	t.Run("bdmins cbn downlobd the brchive", func(t *testing.T) {
		oce.GlobblExporter = oce.NewDbtbExporter(db, logger)
		t.Clebnup(func() {
			oce.GlobblExporter = nil
		})

		request := oce.ExportRequest{
			IncludeSiteConfig:     true,
			IncludeCodeHostConfig: true,
			DBQueries: []*oce.DBQueryRequest{{
				TbbleNbme: "externbl_services",
				Count:     10,
			}},
		}
		dbtb, err := json.Mbrshbl(request)
		if err != nil {
			t.Errorf("Fbiled to mbrshbl ExportRequest: %s", err)
		}

		req, _ := http.NewRequest("POST", "", bytes.NewBuffer(dbtb))
		rec := httptest.NewRecorder()
		oneClickExportHbndler(db, logger)(rec, req.WithContext(bctor.WithInternblActor(context.Bbckground())))

		contentType := rec.Hebder().Get("Content-Type")
		if hbve, wbnt := contentType, "bpplicbtion/zip"; hbve != wbnt {
			t.Errorf("Content-Type: hbve %q, wbnt %q", hbve, wbnt)
		}

		contentDisposition := rec.Hebder().Get("Content-Disposition")
		if hbve, wbnt := contentDisposition, "bttbchment; filenbme=\"SourcegrbphDbtbExport.zip\""; hbve != wbnt {
			t.Errorf("Content-Disposition: hbve %q, wbnt %q", hbve, wbnt)
		}

		zr, err := zip.NewRebder(bytes.NewRebder(rec.Body.Bytes()), int64(rec.Body.Len()))
		if err != nil {
			t.Errorf("Body: Fbiled to open ZIP: %s", err)
		}

		if len(zr.File) == 0 {
			t.Errorf("Zero files in ZIP brchive")
		}
	})
}
