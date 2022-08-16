package insights

import (
	"database/sql"

	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

var scanJobs = basestore.NewSliceScanner(func(s dbutil.Scanner) (j settingsMigrationJob, _ error) {
	err := s.Scan(&j.UserId, &j.OrgId, &j.Global, &j.MigratedInsights, &j.MigratedDashboards, &j.Runs)
	return j, err
})

var scanOrgs = basestore.NewSliceScanner(func(s dbutil.Scanner) (org organization, _ error) {
	err := s.Scan(&org.ID, &org.Name, &org.DisplayName)
	return org, err
})

var scanUsers = basestore.NewSliceScanner(func(s dbutil.Scanner) (u user, _ error) {
	var displayName sql.NullString
	err := s.Scan(&u.ID, &u.Username, &displayName)
	u.DisplayName = displayName.String
	return u, err
})

var scanSettings = basestore.NewSliceScanner(func(scanner dbutil.Scanner) (s settings, _ error) {
	err := scanner.Scan(&s.ID, &s.Subject.Org, &s.Subject.User, &s.Contents)
	return s, err
})

var scanSeries = basestore.NewSliceScanner(func(scanner dbutil.Scanner) (s insightSeries, _ error) {
	err := scanner.Scan(
		&s.ID,
		&s.SeriesID,
		&s.Query,
		&s.CreatedAt,
		&s.OldestHistoricalAt,
		&s.LastRecordedAt,
		&s.NextRecordingAfter,
		&s.LastSnapshotAt,
		&s.NextSnapshotAfter,
		&s.SampleIntervalUnit,
		&s.SampleIntervalValue,
		&s.GeneratedFromCaptureGroups,
		&s.JustInTime,
		&s.GenerationMethod,
		pq.Array(&s.Repositories),
		&s.GroupBy,
	)
	return s, err
})
