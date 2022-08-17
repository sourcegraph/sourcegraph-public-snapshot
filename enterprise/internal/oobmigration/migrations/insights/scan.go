package insights

import (
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

var scanJobs = basestore.NewSliceScanner(func(s dbutil.Scanner) (j settingsMigrationJob, _ error) {
	err := s.Scan(&j.userID, &j.orgID, &j.global, &j.migratedInsights, &j.migratedDashboards)
	return j, err
})

var scanUserOrOrgs = basestore.NewSliceScanner(func(s dbutil.Scanner) (uo userOrOrg, _ error) {
	err := s.Scan(&uo.id, &uo.name, &uo.displayName)
	return uo, err
})

var scanSettings = basestore.NewSliceScanner(func(scanner dbutil.Scanner) (s settings, _ error) {
	err := scanner.Scan(&s.ID, &s.Subject.Org, &s.Subject.User, &s.Contents)
	return s, err
})

var scanSeries = basestore.NewSliceScanner(func(scanner dbutil.Scanner) (s insightSeries, _ error) {
	err := scanner.Scan(
		&s.id,
		&s.seriesID,
		&s.query,
		&s.createdAt,
		&s.oldestHistoricalAt,
		&s.lastRecordedAt,
		&s.nextRecordingAfter,
		&s.lastSnapshotAt,
		&s.nextSnapshotAfter,
		&s.sampleIntervalUnit,
		&s.sampleIntervalValue,
		&s.generatedFromCaptureGroups,
		&s.justInTime,
		&s.generationMethod,
		pq.Array(&s.repositories),
		&s.groupBy,
	)
	return s, err
})
