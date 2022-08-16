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
	err := s.Scan(&org.ID, &org.Name, &org.DisplayName, &org.CreatedAt, &org.UpdatedAt)
	return org, err
})

var scanUsers = basestore.NewSliceScanner(func(s dbutil.Scanner) (u user, _ error) {
	var displayName, avatarURL sql.NullString
	err := s.Scan(&u.ID, &u.Username, &displayName, &avatarURL, &u.CreatedAt, &u.UpdatedAt, &u.SiteAdmin, &u.BuiltinAuth, pq.Array(&u.Tags), &u.InvalidatedSessionsAt, &u.TosAccepted, &u.Searchable)
	u.DisplayName = displayName.String
	u.AvatarURL = avatarURL.String
	return u, err
})

var scanSettings = basestore.NewSliceScanner(func(scanner dbutil.Scanner) (s settings, _ error) {
	err := scanner.Scan(&s.ID, &s.Subject.Org, &s.Subject.User, &s.AuthorUserID, &s.Contents, &s.CreatedAt)
	if s.Subject.Org == nil && s.Subject.User == nil {
		s.Subject.Site = true
	}
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
		&s.Enabled,
		&s.SampleIntervalUnit,
		&s.SampleIntervalValue,
		&s.GeneratedFromCaptureGroups,
		&s.JustInTime,
		&s.GenerationMethod,
		pq.Array(&s.Repositories),
		&s.GroupBy,
		&s.BackfillAttempts,
	)
	return s, err
})
