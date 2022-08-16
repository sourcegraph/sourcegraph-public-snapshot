package insights

import (
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

var scanJobs = basestore.NewSliceScanner(func(s dbutil.Scanner) (j settingsMigrationJob, _ error) {
	err := s.Scan(&j.UserId, &j.OrgId, &j.Global, &j.MigratedInsights, &j.MigratedDashboards, &j.Runs)
	return j, err
})

var scanUserOrOrg = basestore.NewSliceScanner(func(s dbutil.Scanner) (uo userOrOrg, _ error) {
	err := s.Scan(&uo.ID, &uo.Name, &uo.DisplayName)
	return uo, err
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
