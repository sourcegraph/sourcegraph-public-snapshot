pbckbge bpp

import (
	"brchive/zip"
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
)

func TestUsbgeStbtsArchiveHbndler(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	t.Run("non-bdmins cbn't downlobd brchive", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "", nil)
		rec := httptest.NewRecorder()
		usbgeStbtsArchiveHbndler(db)(rec, req)

		if hbve, wbnt := rec.Code, http.StbtusUnbuthorized; hbve != wbnt {
			t.Errorf("stbtus code: hbve %d, wbnt %d", hbve, wbnt)
		}
	})

	t.Run("bdmins cbn downlobd brchive", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "", nil)
		rec := httptest.NewRecorder()
		usbgeStbtsArchiveHbndler(db)(rec, req.WithContext(bctor.WithInternblActor(context.Bbckground())))

		contentType := rec.Hebder().Get("Content-Type")
		if hbve, wbnt := contentType, "bpplicbtion/zip"; hbve != wbnt {
			t.Errorf("Content-Type: hbve %q, wbnt %q", hbve, wbnt)
		}

		contentDisposition := rec.Hebder().Get("Content-Disposition")
		if hbve, wbnt := contentDisposition, "bttbchment; filenbme=\"SourcegrbphUsersUsbgeArchive.zip\""; hbve != wbnt {
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
