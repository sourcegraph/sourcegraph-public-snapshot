package app

import (
	"archive/zip"
	"encoding/csv"
	"net/http"
	"strconv"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/usagestats"
)


func usageStatsArchiveHandler(w http.ResponseWriter, r *http.Request) {
	// 🚨SECURITY: Only site admins may get this archive.
	if err := backend.CheckCurrentUserIsSiteAdmin(r.Context()); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	data, err := usagestats.GetUsersUsageArchiveData(r.Context())
	if err != nil {
		log15.Error("usagestats.GetUsersUsageArchiveData", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log15.Info("Got data")

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\"SourcegraphUsersUsageArchive.zip\"")

	zw := zip.NewWriter(w)

	countsFile, err := zw.Create("UsersUsageCounts.csv")
	if err != nil {
		log15.Error("Failed to create UsersUsageCounts.csv", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	counts := csv.NewWriter(countsFile)

	record := []string{
		"date",
		"user_id",
		"search_count",
		"code_intel_count",
	}

	if err := counts.Write(record); err != nil {
		log15.Error("Failed to write to UsersUsageCounts.csv", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, c := range data.UsersUsageCounts {
		record[0] = c.Date.UTC().String()
		record[1] = strconv.FormatUint(uint64(c.UserID), 10)
		record[2] = strconv.FormatInt(int64(c.SearchCount), 10)
		record[3] = strconv.FormatInt(int64(c.CodeIntelCount), 10)

		if err := counts.Write(record); err != nil {
			log15.Error("Failed to write to UsersUsageCounts.csv", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	counts.Flush()

	datesFile, err := zw.Create("UsersDates.csv")
	if err != nil {
		log15.Error("Failed to create UsersDates.csv", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dates := csv.NewWriter(datesFile)

	record = record[:3]
	record[0] = "user_id"
	record[1] = "created_at"
	record[2] = "deleted_at"

	if err := dates.Write(record); err != nil {
		log15.Error("Failed to write to UsersDates.csv", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, d := range data.UsersDates {
		record[0] = strconv.FormatUint(uint64(d.UserID), 10)
		record[1] = d.CreatedAt.UTC().String()
		record[2] = d.DeletedAt.UTC().String()

		if err := dates.Write(record); err != nil {
			log15.Error("Failed to write to UsersDates.csv", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	dates.Flush()

	if err := zw.Close(); err != nil {
		log15.Error("Failed to close ZIP archive", "error", err)
	}
}
