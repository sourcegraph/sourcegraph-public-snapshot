package httpapi

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/grafana/regexp"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	edb "github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ExportHandler handles retrieving and exporting code insights data.
type ExportHandler struct {
	primaryDB database.DB

	seriesStore          *store.Store
	permStore            *store.InsightPermStore
	insightStore         *store.InsightStore
	searchContextHandler *store.SearchContextHandler
}

const pingName = "InsightsDataExportRequest"

func NewExportHandler(db database.DB, insightsDB edb.InsightsDB) *ExportHandler {
	insightPermStore := store.NewInsightPermissionStore(db)
	seriesStore := store.New(insightsDB, insightPermStore)
	insightsStore := store.NewInsightStore(insightsDB)
	searchContextHandler := store.NewSearchContextHandler(db)

	return &ExportHandler{
		primaryDB:            db,
		seriesStore:          seriesStore,
		permStore:            insightPermStore,
		insightStore:         insightsStore,
		searchContextHandler: searchContextHandler,
	}
}

func (h *ExportHandler) ExportFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]

		archive, err := h.exportCodeInsightData(r.Context(), id)
		if err != nil {
			if errors.Is(err, notFoundError) {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else if errors.Is(err, authenticationError) {
				http.Error(w, err.Error(), http.StatusUnauthorized)
			} else if errors.Is(err, invalidLicenseError) {
				http.Error(w, err.Error(), http.StatusForbidden)
			} else {
				http.Error(w, fmt.Sprintf("failed to export data: %v", err), http.StatusInternalServerError)
			}
			return
		}
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", archive.name))

		_, err = w.Write(archive.data)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to write data: %v", err), http.StatusInternalServerError)
		}
	}
}

type codeInsightsDataArchive struct {
	name string
	data []byte
}

var notFoundError = errors.New("insight not found")
var authenticationError = errors.New("authentication error")
var invalidLicenseError = errors.New("invalid license for code insights")

func (h *ExportHandler) exportCodeInsightData(ctx context.Context, id string) (*codeInsightsDataArchive, error) {
	currentActor := actor.FromContext(ctx)
	if !currentActor.IsAuthenticated() {
		return nil, authenticationError
	}
	userIDs, orgIDs, err := h.permStore.GetUserPermissions(ctx)
	if err != nil {
		return nil, authenticationError
	}

	//lint:ignore SA1019 existing usage of deprecated functionality.
	// Use EventRecorder from internal/telemetryrecorder instead.
	if err := h.primaryDB.EventLogs().Insert(ctx, &database.Event{
		Name:            pingName,
		UserID:          uint32(currentActor.UID),
		AnonymousUserID: "",
		Argument:        nil,
		Timestamp:       time.Now(),
		Source:          "BACKEND",
	}); err != nil {
		return nil, err
	}

	licenseError := licensing.Check(licensing.FeatureCodeInsights)
	if licenseError != nil {
		return nil, invalidLicenseError
	}

	var insightViewId string
	if err := relay.UnmarshalSpec(graphql.ID(id), &insightViewId); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal insight view ID")
	}

	visibleViewSeries, err := h.insightStore.GetAll(ctx, store.InsightQueryArgs{
		UniqueIDs:            []string{insightViewId},
		UserIDs:              userIDs,
		OrgIDs:               orgIDs,
		WithoutAuthorization: false,
	})
	if err != nil {
		return nil, errors.New("could not fetch insight information")
	}
	// 🚨 SECURITY: if the user context doesn't get any response here that means they should not be able to access this insight.
	if len(visibleViewSeries) == 0 {
		return nil, notFoundError
	}

	opts := store.ExportOpts{}
	includeRepo := func(regex ...string) {
		opts.IncludeRepoRegex = append(opts.IncludeRepoRegex, regex...)
	}
	excludeRepo := func(regex ...string) {
		opts.ExcludeRepoRegex = append(opts.ExcludeRepoRegex, regex...)
	}

	if visibleViewSeries[0].DefaultFilterIncludeRepoRegex != nil {
		includeRepo(*visibleViewSeries[0].DefaultFilterIncludeRepoRegex)
	}
	if visibleViewSeries[0].DefaultFilterExcludeRepoRegex != nil {
		includeRepo(*visibleViewSeries[0].DefaultFilterExcludeRepoRegex)
	}

	inc, exc, err := h.searchContextHandler.UnwrapSearchContexts(ctx, visibleViewSeries[0].DefaultFilterSearchContexts)
	if err != nil {
		return nil, errors.Wrap(err, "search context error")
	}
	includeRepo(inc...)
	excludeRepo(exc...)

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	timestamp := time.Now().Format(time.RFC3339)
	escapedInsightViewTitle := regexp.MustCompile(`\W+`).ReplaceAllString(visibleViewSeries[0].Title, "-")
	name := fmt.Sprintf("%s-%s", escapedInsightViewTitle, timestamp)

	dataFile, err := zw.Create(fmt.Sprintf("%s.csv", name))
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

	opts.InsightViewUniqueID = insightViewId
	dataPoints, err := h.seriesStore.GetAllDataForInsightViewID(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch all data for insight")
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
	return &codeInsightsDataArchive{
		name: name,
		data: buf.Bytes(),
	}, nil
}

func emptyStringIfNil(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
