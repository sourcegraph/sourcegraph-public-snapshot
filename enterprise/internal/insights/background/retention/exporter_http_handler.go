package retention

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

func NewExportHandler(db database.DB, insightsDB edb.InsightsDB) http.HandlerFunc {
	fmt.Println("export handler")
	insightPermStore := store.NewInsightPermissionStore(db)
	insightsStore := store.New(insightsDB, insightPermStore)

	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("response writer")
		id := mux.Vars(r)["id"]
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=\"CodeInsightsArchive.zip\"")

		archive, err := exportCodeInsightData(r.Context(), insightsStore, id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, _ = w.Write(archive)
	}
}

func exportCodeInsightData(ctx context.Context, store *store.Store, id string) ([]byte, error) {
	var insightViewId string
	if err := relay.UnmarshalSpec(graphql.ID(id), &insightViewId); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal insight view ID")
	}
	fmt.Println(id, insightViewId)
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	dataFile, err := zw.Create(fmt.Sprintf("CodeInsightData-%s.csv", insightViewId))
	if err != nil {
		return nil, err
	}

	dataWriter := csv.NewWriter(dataFile)

	dataPoint := []string{
		"check",
	}

	if err := dataWriter.Write(dataPoint); err != nil {
		return nil, err
	}

	dataPoints := []string{"this works"}

	for _, d := range dataPoints {
		dataPoint[0] = d

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
