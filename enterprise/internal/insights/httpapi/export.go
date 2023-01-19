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
	"github.com/sourcegraph/log"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewExportHandler(db database.DB, insightsDB edb.InsightsDB) http.HandlerFunc {
	insightPermStore := store.NewInsightPermissionStore(db)
	insightsStore := store.New(insightsDB, insightPermStore)
	logger := log.Scoped("Code insights data exporter", "")

	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=\"CodeInsightsArchive.zip\"")

		archive, err := exportCodeInsightData(r.Context(), insightsStore, id)
		if err != nil {
			logger.Error("exporting data errored", log.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = w.Write(archive)
		if err != nil {
			logger.Error("writing archive errored", log.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func exportCodeInsightData(ctx context.Context, store *store.Store, id string) ([]byte, error) {
	var insightViewId string
	if err := relay.UnmarshalSpec(graphql.ID(id), &insightViewId); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal insight view ID")
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	dataFile, err := zw.Create(fmt.Sprintf("CodeInsightData-%s.csv", insightViewId))
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

	dataPoints, err := store.GetAllDataForInsightViewID(ctx, insightViewId)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch all data for insight view with id %s", insightViewId)
	}

	for _, d := range dataPoints {
		dataPoint[0] = d.InsightViewTitle
		dataPoint[1] = d.SeriesLabel
		dataPoint[2] = d.SeriesQuery
		dataPoint[3] = d.RecordingTime.String()
		dataPoint[4] = emptyStringIfNil(d.RepoName)
		dataPoint[5] = fmt.Sprintf("%d", d.Value)
		dataPoint[6] = emptyStringIfNil(d.Capture)

		if err := dataWriter.Write(dataPoint); err != nil {
			return nil, err
		}
	}

	dataWriter.Flush()

	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func emptyStringIfNil(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
