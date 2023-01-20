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
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	searchquery "github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/searchcontexts"
	sctypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ExportHandler handles retrieving and exporting code insights data.
type ExportHandler struct {
	primaryDB database.DB

	seriesStore  *store.Store
	permStore    *store.InsightPermStore
	insightStore *store.InsightStore
}

func NewExportHandler(db database.DB, insightsDB edb.InsightsDB) *ExportHandler {
	insightPermStore := store.NewInsightPermissionStore(db)
	seriesStore := store.New(insightsDB, insightPermStore)
	insightsStore := store.NewInsightStore(insightsDB)

	return &ExportHandler{
		primaryDB:    db,
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
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"CodeInsightsDataExport-%s.zip\"", archive.name))

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

	scLoader := &scLoader{primary: h.primaryDB}
	inc, exc, err := unwrapSearchContexts(ctx, scLoader, visibleViewSeries[0].DefaultFilterSearchContexts)
	if err != nil {
		return nil, errors.Wrap(err, "search context error")
	}
	includeRepo(inc...)
	excludeRepo(exc...)

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	timestamp := time.Now().Format(time.RFC3339)
	name := fmt.Sprintf("%s-%s", insightViewId, timestamp)

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

type SearchContextLoader interface {
	GetByName(ctx context.Context, name string) (*sctypes.SearchContext, error)
}

type scLoader struct {
	primary database.DB
}

func (l *scLoader) GetByName(ctx context.Context, name string) (*sctypes.SearchContext, error) {
	return searchcontexts.ResolveSearchContextSpec(ctx, l.primary, name)
}

func unwrapSearchContexts(ctx context.Context, loader SearchContextLoader, rawContexts []string) ([]string, []string, error) {
	var include []string
	var exclude []string

	for _, rawContext := range rawContexts {
		searchContext, err := loader.GetByName(ctx, rawContext)
		if err != nil {
			return nil, nil, err
		}
		if searchContext.Query != "" {
			var plan searchquery.Plan
			plan, err := searchquery.Pipeline(
				searchquery.Init(searchContext.Query, searchquery.SearchTypeRegex),
			)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "failed to parse search query for search context: %s", rawContext)
			}
			inc, exc := plan.ToQ().Repositories()
			for _, repoFilter := range inc {
				if len(repoFilter.Revs) > 0 {
					return nil, nil, errors.Errorf("search context filters cannot include repo revisions: %s", rawContext)
				}
				include = append(include, repoFilter.Repo)
			}
			exclude = append(exclude, exc...)
		}
	}
	return include, exclude, nil
}
