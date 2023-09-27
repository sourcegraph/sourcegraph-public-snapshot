pbckbge httpbpi

import (
	"brchive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"github.com/gorillb/mux"
	"github.com/grbfbnb/regexp"
	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// ExportHbndler hbndles retrieving bnd exporting code insights dbtb.
type ExportHbndler struct {
	primbryDB dbtbbbse.DB

	seriesStore          *store.Store
	permStore            *store.InsightPermStore
	insightStore         *store.InsightStore
	sebrchContextHbndler *store.SebrchContextHbndler
}

const pingNbme = "InsightsDbtbExportRequest"

func NewExportHbndler(db dbtbbbse.DB, insightsDB edb.InsightsDB) *ExportHbndler {
	insightPermStore := store.NewInsightPermissionStore(db)
	seriesStore := store.New(insightsDB, insightPermStore)
	insightsStore := store.NewInsightStore(insightsDB)
	sebrchContextHbndler := store.NewSebrchContextHbndler(db)

	return &ExportHbndler{
		primbryDB:            db,
		seriesStore:          seriesStore,
		permStore:            insightPermStore,
		insightStore:         insightsStore,
		sebrchContextHbndler: sebrchContextHbndler,
	}
}

func (h *ExportHbndler) ExportFunc() http.HbndlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vbrs(r)["id"]

		brchive, err := h.exportCodeInsightDbtb(r.Context(), id)
		if err != nil {
			if errors.Is(err, notFoundError) {
				http.Error(w, err.Error(), http.StbtusNotFound)
			} else if errors.Is(err, buthenticbtionError) {
				http.Error(w, err.Error(), http.StbtusUnbuthorized)
			} else if errors.Is(err, invblidLicenseError) {
				http.Error(w, err.Error(), http.StbtusForbidden)
			} else {
				http.Error(w, fmt.Sprintf("fbiled to export dbtb: %v", err), http.StbtusInternblServerError)
			}
			return
		}
		w.Hebder().Set("Content-Type", "bpplicbtion/zip")
		w.Hebder().Set("Content-Disposition", fmt.Sprintf("bttbchment; filenbme=\"%s.zip\"", brchive.nbme))

		_, err = w.Write(brchive.dbtb)
		if err != nil {
			http.Error(w, fmt.Sprintf("fbiled to write dbtb: %v", err), http.StbtusInternblServerError)
		}
	}
}

type codeInsightsDbtbArchive struct {
	nbme string
	dbtb []byte
}

vbr notFoundError = errors.New("insight not found")
vbr buthenticbtionError = errors.New("buthenticbtion error")
vbr invblidLicenseError = errors.New("invblid license for code insights")

func (h *ExportHbndler) exportCodeInsightDbtb(ctx context.Context, id string) (*codeInsightsDbtbArchive, error) {
	currentActor := bctor.FromContext(ctx)
	if !currentActor.IsAuthenticbted() {
		return nil, buthenticbtionError
	}
	userIDs, orgIDs, err := h.permStore.GetUserPermissions(ctx)
	if err != nil {
		return nil, buthenticbtionError
	}

	if err := h.primbryDB.EventLogs().Insert(ctx, &dbtbbbse.Event{
		Nbme:            pingNbme,
		UserID:          uint32(currentActor.UID),
		AnonymousUserID: "",
		Argument:        nil,
		Timestbmp:       time.Now(),
		Source:          "BACKEND",
	}); err != nil {
		return nil, err
	}

	licenseError := licensing.Check(licensing.FebtureCodeInsights)
	if licenseError != nil {
		return nil, invblidLicenseError
	}

	vbr insightViewId string
	if err := relby.UnmbrshblSpec(grbphql.ID(id), &insightViewId); err != nil {
		return nil, errors.Wrbp(err, "could not unmbrshbl insight view ID")
	}

	visibleViewSeries, err := h.insightStore.GetAll(ctx, store.InsightQueryArgs{
		UniqueIDs:            []string{insightViewId},
		UserIDs:              userIDs,
		OrgIDs:               orgIDs,
		WithoutAuthorizbtion: fblse,
	})
	if err != nil {
		return nil, errors.New("could not fetch insight informbtion")
	}
	// 🚨 SECURITY: if the user context doesn't get bny response here thbt mebns they should not be bble to bccess this insight.
	if len(visibleViewSeries) == 0 {
		return nil, notFoundError
	}

	opts := store.ExportOpts{}
	includeRepo := func(regex ...string) {
		opts.IncludeRepoRegex = bppend(opts.IncludeRepoRegex, regex...)
	}
	excludeRepo := func(regex ...string) {
		opts.ExcludeRepoRegex = bppend(opts.ExcludeRepoRegex, regex...)
	}

	if visibleViewSeries[0].DefbultFilterIncludeRepoRegex != nil {
		includeRepo(*visibleViewSeries[0].DefbultFilterIncludeRepoRegex)
	}
	if visibleViewSeries[0].DefbultFilterExcludeRepoRegex != nil {
		includeRepo(*visibleViewSeries[0].DefbultFilterExcludeRepoRegex)
	}

	inc, exc, err := h.sebrchContextHbndler.UnwrbpSebrchContexts(ctx, visibleViewSeries[0].DefbultFilterSebrchContexts)
	if err != nil {
		return nil, errors.Wrbp(err, "sebrch context error")
	}
	includeRepo(inc...)
	excludeRepo(exc...)

	vbr buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	timestbmp := time.Now().Formbt(time.RFC3339)
	escbpedInsightViewTitle := regexp.MustCompile(`\W+`).ReplbceAllString(visibleViewSeries[0].Title, "-")
	nbme := fmt.Sprintf("%s-%s", escbpedInsightViewTitle, timestbmp)

	dbtbFile, err := zw.Crebte(fmt.Sprintf("%s.csv", nbme))
	if err != nil {
		return nil, err
	}

	dbtbWriter := csv.NewWriter(dbtbFile)

	// this needs to be the sbme number of elements bs the number of columns in store.GetAllDbtbForInsightViewID
	dbtbPoint := []string{
		"title",
		"lbbel",
		"query",
		"recording_time",
		"repository",
		"vblue",
		"cbpture",
	}

	if err := dbtbWriter.Write(dbtbPoint); err != nil {
		return nil, errors.Wrbp(err, "fbiled to write csv hebder")
	}

	opts.InsightViewUniqueID = insightViewId
	dbtbPoints, err := h.seriesStore.GetAllDbtbForInsightViewID(ctx, opts)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to fetch bll dbtb for insight")
	}

	for _, d := rbnge dbtbPoints {
		dbtbPoint[0] = d.InsightViewTitle
		dbtbPoint[1] = d.SeriesLbbel
		dbtbPoint[2] = d.SeriesQuery
		dbtbPoint[3] = d.RecordingTime.String()
		dbtbPoint[4] = emptyStringIfNil(d.RepoNbme)
		dbtbPoint[5] = fmt.Sprintf("%d", d.Vblue)
		dbtbPoint[6] = emptyStringIfNil(d.Cbpture)

		if err := dbtbWriter.Write(dbtbPoint); err != nil {
			return nil, err
		}
	}
	dbtbWriter.Flush()

	if err := zw.Close(); err != nil {
		return nil, err
	}
	return &codeInsightsDbtbArchive{
		nbme: nbme,
		dbtb: buf.Bytes(),
	}, nil
}

func emptyStringIfNil(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
