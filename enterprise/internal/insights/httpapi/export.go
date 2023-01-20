package httpapi

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ExportHandler handles retrieving and exporting code insights data.
type ExportHandler struct {
	seriesStore  *store.Store
	permStore    *store.InsightPermStore
	insightStore *store.InsightStore
}

func NewExportHandler(db database.DB, insightsDB edb.InsightsDB) *ExportHandler {
	insightPermStore := store.NewInsightPermissionStore(db)
	seriesStore := store.New(insightsDB, insightPermStore)
	insightsStore := store.NewInsightStore(insightsDB)

	return &ExportHandler{
		seriesStore:  seriesStore,
		permStore:    insightPermStore,
		insightStore: insightsStore,
	}
}

func (h *ExportHandler) ExportFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]

		archive, err := h.exportCodeInsightData(r.Context(), id)
		if err != nil {
			if errors.Is(err, notFoundError) {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, fmt.Sprintf("failed to export data: %v", err), http.StatusInternalServerError)
			}
			return
		}
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"CodeInsightsDataExport-%s.zip\"", archive.insightName))

		_, err = w.Write(archive.data)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to write data: %v", err), http.StatusInternalServerError)
		}
	}
}

type codeInsightsDataArchive struct {
	insightName string
	data        []byte
}

var notFoundError = errors.New("insight not found")

func (h *ExportHandler) exportCodeInsightData(ctx context.Context, id string) (*codeInsightsDataArchive, error) {
	var insightViewId string
	if err := relay.UnmarshalSpec(graphql.ID(id), &insightViewId); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal insight view ID")
	}

	userID, orgIDs, err := h.permStore.GetUserPermissions(ctx)
	if err != nil {
		return nil, errors.New("error with session")
	}

	visibleViewSeries, err := h.insightStore.GetAll(ctx, store.InsightQueryArgs{
		UniqueIDs:            []string{insightViewId},
		UserID:               userID,
		OrgID:                orgIDs,
		WithoutAuthorization: false,
	})
	if err != nil {
		return nil, errors.New("could not fetch insight information")
	}
	// 🚨 SECURITY: if the user context doesn't get any response here that means they should not be able to access this insight.
	if len(visibleViewSeries) == 0 {
		return nil, notFoundError
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	dataFile, err := zw.Create(fmt.Sprintf("%s.csv", visibleViewSeries[0].Title))
	if err != nil {
		return nil, err
	}

	dataWriter := csv.NewWriter(dataFile)

	// this needs to be the same number of elements as the number of columns in store.GetAllDataForInsightViewID
	dataPoint := []string{
		"title",
		"label",
		"query",
		"recording_time",
		"repository",
		"value",
		"capture",
	}

	if err := dataWriter.Write(dataPoint); err != nil {
		return nil, errors.Wrap(err, "failed to write csv header")
	}

	dataPoints, err := h.seriesStore.GetAllDataForInsightViewID(ctx, insightViewId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch all data for insight")
	}

	var insightName string
	for _, d := range dataPoints {
		dataPoint[0] = d.InsightViewTitle
		dataPoint[1] = d.SeriesLabel
		dataPoint[2] = d.SeriesQuery
		dataPoint[3] = d.RecordingTime.String()
		dataPoint[4] = emptyStringIfNil(d.RepoName)
		dataPoint[5] = fmt.Sprintf("%d", d.Value)
		dataPoint[6] = emptyStringIfNil(d.Capture)
		insightName = d.InsightViewTitle

		if err := dataWriter.Write(dataPoint); err != nil {
			return nil, err
		}
	}
	dataWriter.Flush()

	if err := zw.Close(); err != nil {
		return nil, err
	}
	return &codeInsightsDataArchive{
		insightName: insightName,
		data:        buf.Bytes(),
	}, nil
}

func emptyStringIfNil(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
