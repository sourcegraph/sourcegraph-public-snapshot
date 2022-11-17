package insights

import (
	"time"

	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type insightsMigrationJob struct {
	userID             *int32
	orgID              *int32
	migratedInsights   int
	migratedDashboards int
}

var scanJobs = basestore.NewSliceScanner(func(s dbutil.Scanner) (j insightsMigrationJob, _ error) {
	err := s.Scan(&j.userID, &j.orgID, &j.migratedInsights, &j.migratedDashboards)
	return j, err
})

type settings struct {
	id       int32
	org      *int32
	user     *int32
	contents string
}

var scanSettings = basestore.NewSliceScanner(func(scanner dbutil.Scanner) (s settings, _ error) {
	err := scanner.Scan(&s.id, &s.org, &s.user, &s.contents)
	return s, err
})

type userOrOrg struct {
	name        string
	displayName *string
}

var scanFirstUserOrOrg = basestore.NewFirstScanner(func(s dbutil.Scanner) (uo userOrOrg, _ error) {
	err := s.Scan(&uo.name, &uo.displayName)
	return uo, err
})

type insightSeries struct {
	id                         int
	seriesID                   string
	query                      string
	createdAt                  time.Time
	oldestHistoricalAt         time.Time
	lastRecordedAt             time.Time
	nextRecordingAfter         time.Time
	lastSnapshotAt             time.Time
	nextSnapshotAfter          time.Time
	repositories               []string
	sampleIntervalUnit         string
	sampleIntervalValue        int
	generatedFromCaptureGroups bool
	justInTime                 bool
	generationMethod           string
	groupBy                    *string
}

var scanFirstSeries = basestore.NewFirstScanner(func(scanner dbutil.Scanner) (s insightSeries, _ error) {
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

type insightSeriesWithMetadata struct {
	insightSeries
	label  string
	stroke string
}
