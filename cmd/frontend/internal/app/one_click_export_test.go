package app

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/log/logtest"
	oce "github.com/sourcegraph/sourcegraph/cmd/frontend/oneclickexport"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestOneClickExportHandler(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))

	t.Run("non-admins cannot download the archive", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "", nil)
		rec := httptest.NewRecorder()
		oneClickExportHandler(db, logger)(rec, req)

		if have, want := rec.Code, http.StatusUnauthorized; have != want {
			t.Errorf("status code: have %d, want %d", have, want)
		}
	})

	t.Run("admins can download the archive", func(t *testing.T) {
		oce.GlobalExporter = oce.NewDataExporter(db, logger)
		t.Cleanup(func() {
			oce.GlobalExporter = nil
		})

		request := oce.ExportRequest{
			IncludeSiteConfig:     true,
			IncludeCodeHostConfig: true,
			DBQueries: []*oce.DBQueryRequest{{
				TableName: "external_services",
				Count:     10,
			}},
		}
		data, err := json.Marshal(request)
		if err != nil {
			t.Errorf("Failed to marshal ExportRequest: %s", err)
		}

		req, _ := http.NewRequest("POST", "", bytes.NewBuffer(data))
		rec := httptest.NewRecorder()
		oneClickExportHandler(db, logger)(rec, req.WithContext(actor.WithInternalActor(context.Background())))

		contentType := rec.Header().Get("Content-Type")
		if have, want := contentType, "application/zip"; have != want {
			t.Errorf("Content-Type: have %q, want %q", have, want)
		}

		contentDisposition := rec.Header().Get("Content-Disposition")
		if have, want := contentDisposition, "attachment; filename=\"SourcegraphDataExport.zip\""; have != want {
			t.Errorf("Content-Disposition: have %q, want %q", have, want)
		}

		zr, err := zip.NewReader(bytes.NewReader(rec.Body.Bytes()), int64(rec.Body.Len()))
		if err != nil {
			t.Errorf("Body: Failed to open ZIP: %s", err)
		}

		if len(zr.File) == 0 {
			t.Errorf("Zero files in ZIP archive")
		}
	})
}
