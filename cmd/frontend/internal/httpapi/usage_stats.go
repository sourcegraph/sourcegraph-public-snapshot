package httpapi

import (
	"archive/zip"
	"encoding/csv"
	"net/http"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/usagestats"
)


func usageStatsDownloadServe(w http.ResponseWriter, r *http.Request) error {
	data, err := usagestats.GetUsersUsageArchiveData(r.Context())
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\"SourcegraphUsersUsageArchive.zip\"")

	zw := zip.NewWriter(w)

	countsFile, err := zw.Create("UsersUsageCounts.csv")
	if err != nil {
		return err
	}

	counts := csv.NewWriter(countsFile)

	record := []string{
		"date",
		"user_id",
		"search_count",
		"code_intel_count",
	}

	if err := counts.Write(record); err != nil {
		return err
	}

	for _, c := range data.UsersUsageCounts {
		record[0] = c.Date.UTC().String()
		record[1] = strconv.FormatUint(uint64(c.UserID), 10)
		record[2] = strconv.FormatInt(int64(c.SearchCount), 10)
		record[3] = strconv.FormatInt(int64(c.CodeIntelCount), 10)

		if err := counts.Write(record); err != nil {
			return err
		}
	}

	counts.Flush()

	datesFile, err := zw.Create("UsersDates.csv")
	if err != nil {
		return err
	}

	dates := csv.NewWriter(datesFile)

	record = record[:3]
	record[0] = "user_id"
	record[1] = "created_at"
	record[2] = "deleted_at"

	if err := dates.Write(record); err != nil {
		return err
	}

	for _, d := range data.UsersDates {
		record[0] = strconv.FormatUint(uint64(d.UserID), 10)
		record[1] = d.CreatedAt.UTC().String()
		record[2] = d.DeletedAt.UTC().String()

		if err := dates.Write(record); err != nil {
			return err
		}
	}

	dates.Flush()

	return nil
}
