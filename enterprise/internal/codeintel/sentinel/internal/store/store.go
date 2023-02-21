package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Store interface {
	GetVulnerabilities(ctx context.Context, args shared.GetVulnerabilitiesArgs) ([]shared.Vulnerability, int, error)
	InsertVulnerabilities(ctx context.Context, vulnerabilities []shared.Vulnerability) error
	GetVulnerabilityMatches(ctx context.Context, args shared.GetVulnerabilityMatchesArgs) ([]shared.VulnerabilityMatch, int, error)
	ScanMatches(ctx context.Context) error
}

type store struct {
	db         *basestore.Store
	logger     logger.Logger
	operations *operations
}

// New returns a new sentinel store.
func New(observationCtx *observation.Context, db database.DB) Store {
	return &store{
		db:         basestore.NewWithHandle(db.Handle()),
		logger:     logger.Scoped("sentinel.store", ""),
		operations: newOperations(observationCtx),
	}
}

func (s *store) GetVulnerabilities(ctx context.Context, args shared.GetVulnerabilitiesArgs) (_ []shared.Vulnerability, _ int, err error) {
	ctx, _, endObservation := s.operations.getVulnerabilities.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return scanVulnerabilities(s.db.Query(ctx, sqlf.Sprintf(getVulnerabilitiesQuery, args.Limit, args.Offset)))
}

const getVulnerabilitiesQuery = `
WITH limited_vulnerabilities AS (
	SELECT
		v.id,
		v.source_id,
		v.summary,
		v.details,
		v.cpes,
		v.cwes,
		v.aliases,
		v.related,
		v.data_source,
		v.urls,
		v.severity,
		v.cvss_vector,
		v.cvss_score,
		v.published,
		v.modified,
		v.withdrawn,
		COUNT(*) OVER() AS count
	FROM vulnerabilities v
	LIMIT %s
	OFFSET %s
)
SELECT
	v.id,
	v.source_id,
	v.summary,
	v.details,
	v.cpes,
	v.cwes,
	v.aliases,
	v.related,
	v.data_source,
	v.urls,
	v.severity,
	v.cvss_vector,
	v.cvss_score,
	v.published,
	v.modified,
	v.withdrawn,
	vap.package_name,
	vap.language,
	vap.namespace,
	vap.version_constraint,
	vap.fixed,
	vap.fixed_in,
	vas.path,
	vas.symbols,
	v.count
FROM limited_vulnerabilities v
LEFT JOIN vulnerability_affected_packages vap ON vap.vulnerability_id = v.id
LEFT JOIN vulnerability_affected_symbols vas ON vas.vulnerability_affected_package_id = vap.id
`

var scanVulnerabilityTriple = func(s dbutil.Scanner) (v shared.Vulnerability, vap shared.AffectedPackage, vas shared.AffectedSymbol, count int, _ error) {
	var fixedIn string

	if err := s.Scan(
		&v.ID,
		&v.SourceID,
		&v.Summary,
		&v.Details,
		pq.Array(&v.CPEs),
		pq.Array(&v.CWEs),
		pq.Array(&v.Aliases),
		pq.Array(&v.Related),
		&v.DataSource,
		pq.Array(&v.URLs),
		&v.Severity,
		&v.CVSSVector,
		&v.CVSSScore,
		&v.Published,
		&v.Modified,
		&v.Withdrawn,
		// RHS(s) of left join (may be null)
		&dbutil.NullString{S: &vap.PackageName},
		&dbutil.NullString{S: &vap.Language},
		&dbutil.NullString{S: &vap.Namespace},
		pq.Array(&vap.VersionConstraint),
		&dbutil.NullBool{B: &vap.Fixed},
		&dbutil.NullString{S: &fixedIn},
		&dbutil.NullString{S: &vas.Path},
		pq.Array(vas.Symbols),
		&count,
	); err != nil {
		return shared.Vulnerability{}, shared.AffectedPackage{}, shared.AffectedSymbol{}, 0, err
	}

	if fixedIn != "" {
		vap.FixedIn = &fixedIn
	}

	return v, vap, vas, count, nil
}

var scanVulnerabilities = func(rows basestore.Rows, queryErr error) (values []shared.Vulnerability, totalCount int, _ error) {
	scanner := func(s dbutil.Scanner) (bool, error) {
		value, affectedPackage, affectedSymbol, count, err := scanVulnerabilityTriple(s)
		if err != nil {
			return false, err
		}

		lastValuesIndex := len(values) - 1
		if len(values) == 0 || values[lastValuesIndex].ID != value.ID {
			// New data from LHS of join
			values = append(values, value)
			lastValuesIndex++
		}

		// RHS of join is non-NULL
		if affectedPackage.PackageName != "" {
			value := values[lastValuesIndex]
			{
				if len(value.AffectedPackages) == 0 {
					// New data for affected packages (case 1)
					value.AffectedPackages = append(value.AffectedPackages, affectedPackage)
				} else {
					lastPackageIndex := len(value.AffectedPackages) - 1
					lastPackage := value.AffectedPackages[lastPackageIndex]

					// New data for affected packages (case 2)
					if lastPackage.Namespace != affectedPackage.Namespace || lastPackage.Language != affectedPackage.Language || lastPackage.PackageName != affectedPackage.PackageName {
						value.AffectedPackages = append(value.AffectedPackages, affectedPackage)
					}
				}

				if affectedSymbol.Path != "" {
					lastPackageIndex := len(value.AffectedPackages) - 1
					lastPackage := value.AffectedPackages[lastPackageIndex]
					{
						lastPackage.AffectedSymbols = append(lastPackage.AffectedSymbols, affectedSymbol)
					}
					value.AffectedPackages[lastPackageIndex] = lastPackage
				}
			}
			values[lastValuesIndex] = value
		}

		totalCount = count
		return true, nil
	}

	err := basestore.NewCallbackScanner(scanner)(rows, queryErr)
	return values, totalCount, err
}

func (s *store) InsertVulnerabilities(ctx context.Context, vulnerabilities []shared.Vulnerability) (err error) {
	ctx, _, endObservation := s.operations.insertVulnerabilities.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.Exec(ctx, sqlf.Sprintf(insertVulnerabilitiesTemporaryVulnerabilitiesTableQuery)); err != nil {
		return err
	}
	if err := tx.Exec(ctx, sqlf.Sprintf(insertVulnerabilitiesTemporaryVulnerabilityAffectedPackagesTableQuery)); err != nil {
		return err
	}

	if err := batch.WithInserter(
		ctx,
		tx.Handle(),
		"t_vulnerabilities",
		batch.MaxNumPostgresParameters,
		[]string{
			"source_id",
			"summary",
			"details",
			"cpes",
			"cwes",
			"aliases",
			"related",
			"data_source",
			"urls",
			"severity",
			"cvss_vector",
			"cvss_score",
			"published",
			"modified",
			"withdrawn",
		},
		func(inserter *batch.Inserter) error {
			for _, v := range vulnerabilities {
				if v.CPEs == nil {
					v.CPEs = []string{}
				}
				if v.CWEs == nil {
					v.CWEs = []string{}
				}
				if v.Aliases == nil {
					v.Aliases = []string{}
				}
				if v.Related == nil {
					v.Related = []string{}
				}
				if v.URLs == nil {
					v.URLs = []string{}
				}

				if err := inserter.Insert(
					ctx,
					v.SourceID,
					v.Summary,
					v.Details,
					v.CPEs,
					v.CWEs,
					v.Aliases,
					v.Related,
					v.DataSource,
					v.URLs,
					v.Severity,
					v.CVSSVector,
					v.CVSSScore,
					v.Published,
					dbutil.NullTime{Time: v.Modified},
					dbutil.NullTime{Time: v.Withdrawn},
				); err != nil {
					return err
				}
			}

			return nil
		}); err != nil {
		return err
	}

	if err := batch.WithInserter(
		ctx,
		tx.Handle(),
		"t_vulnerability_affected_packages",
		batch.MaxNumPostgresParameters,
		[]string{
			"source_id",
			"package_name",
			"language",
			"namespace",
			"version_constraint",
			"fixed",
			"fixed_in",
			"affected_symbols",
		},
		func(inserter *batch.Inserter) error {
			for _, v := range vulnerabilities {
				for _, ap := range v.AffectedPackages {
					if ap.VersionConstraint == nil {
						ap.VersionConstraint = []string{}
					}
					if ap.AffectedSymbols == nil {
						ap.AffectedSymbols = []shared.AffectedSymbol{}
					}

					serialized, err := json.Marshal(ap.AffectedSymbols)
					if err != nil {
						return err
					}

					if err := inserter.Insert(
						ctx,
						v.SourceID,
						ap.PackageName,
						ap.Language,
						ap.Namespace,
						ap.VersionConstraint,
						ap.Fixed,
						ap.FixedIn,
						serialized,
					); err != nil {
						return err
					}
				}
			}

			return nil
		}); err != nil {
		return err
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(insertVulnerabilitiesUpdateQuery)); err != nil {
		return err
	}
	if err := tx.Exec(ctx, sqlf.Sprintf(insertVulnerabilitiesAffectedPackagesUpdateQuery)); err != nil {
		return err
	}
	if err := tx.Exec(ctx, sqlf.Sprintf(insertVulnerabilitiesAffectedSymbolsUpdateQuery)); err != nil {
		return err
	}

	return nil
}

const insertVulnerabilitiesTemporaryVulnerabilitiesTableQuery = `
CREATE TEMPORARY TABLE t_vulnerabilities (
	source_id    TEXT NOT NULL,
	summary      TEXT NOT NULL,
	details      TEXT NOT NULL,
	cpes         TEXT[] NOT NULL,
	cwes         TEXT[] NOT NULL,
	aliases      TEXT[] NOT NULL,
	related      TEXT[] NOT NULL,
	data_source  TEXT NOT NULL,
	urls         TEXT[] NOT NULL,
	severity     TEXT NOT NULL,
	cvss_vector  TEXT NOT NULL,
	cvss_score   TEXT NOT NULL,
	published    TIMESTAMP WITH TIME ZONE NOT NULL,
	modified     TIMESTAMP WITH TIME ZONE,
	withdrawn    TIMESTAMP WITH TIME ZONE
) ON COMMIT DROP
`

const insertVulnerabilitiesTemporaryVulnerabilityAffectedPackagesTableQuery = `
CREATE TEMPORARY TABLE t_vulnerability_affected_packages (
	source_id           TEXT NOT NULL,
	package_name        TEXT NOT NULL,
	language            TEXT NOT NULL,
	namespace           TEXT NOT NULL,
	version_constraint  TEXT[] NOT NULL,
	fixed               boolean NOT NULL,
	fixed_in            TEXT,
	affected_symbols    JSON NOT NULL
) ON COMMIT DROP
`

const insertVulnerabilitiesUpdateQuery = `
INSERT INTO vulnerabilities (
	source_id,
	summary,
	details,
	cpes,
	cwes,
	aliases,
	related,
	data_source,
	urls,
	severity,
	cvss_vector,
	cvss_score,
	published,
	modified,
	withdrawn
)
SELECT
	source_id,
	summary,
	details,
	cpes,
	cwes,
	aliases,
	related,
	data_source,
	urls,
	severity,
	cvss_vector,
	cvss_score,
	published,
	modified,
	withdrawn
FROM t_vulnerabilities

-- TODO - update instead
ON CONFLICT DO NOTHING
`

const insertVulnerabilitiesAffectedPackagesUpdateQuery = `
INSERT INTO vulnerability_affected_packages(
	vulnerability_id,
	package_name,
	language,
	namespace,
	version_constraint,
	fixed,
	fixed_in
)
SELECT
	(SELECT v.id FROM vulnerabilities v WHERE v.source_id = vap.source_id),
	package_name,
	language,
	namespace,
	version_constraint,
	fixed,
	fixed_in
FROM t_vulnerability_affected_packages vap

-- TODO - update instead
ON CONFLICT DO NOTHING
`

const insertVulnerabilitiesAffectedSymbolsUpdateQuery = `
WITH
json_candidates AS (
	SELECT
		vap.id,
		json_array_elements(tvap.affected_symbols) AS affected_symbol
	FROM t_vulnerability_affected_packages tvap
	JOIN vulnerability_affected_packages vap ON vap.package_name = tvap.package_name
	JOIN vulnerabilities v ON v.id = vap.vulnerability_id
	WHERE
		v.source_id = tvap.source_id
),
candidates AS (
	SELECT
		c.id,
		c.affected_symbol->'path'::text AS path,
		ARRAY(SELECT json_array_elements_text(c.affected_symbol->'symbols'))::text[] AS symbols
	FROM json_candidates c
)
INSERT INTO vulnerability_affected_symbols(vulnerability_affected_package_id, path, symbols)
SELECT c.id, c.path, c.symbols FROM candidates c

-- TODO - update instead
ON CONFLICT DO NOTHING
`

func (s *store) GetVulnerabilityMatches(ctx context.Context, args shared.GetVulnerabilityMatchesArgs) (_ []shared.VulnerabilityMatch, _ int, err error) {
	ctx, _, endObservation := s.operations.getVulnerabilityMatches.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return scanVulnerabilityMatch(s.db.Query(ctx, sqlf.Sprintf(getVulnerabilityMatchesQuery, args.Limit, args.Offset)))
}

const getVulnerabilityMatchesQuery = `
SELECT
	m.id,
	m.upload_id,
	m.vulnerability_affected_package_id
FROM vulnerability_matches m
ORDER BY id
LIMIT %s
OFFSET %s
`

var scanVulnerabilityMatch = basestore.NewSliceWithCountScanner(func(s dbutil.Scanner) (match shared.VulnerabilityMatch, count int, _ error) {
	if err := s.Scan(
		&match.UploadID,
		&match.VulnerabilityAffectedPackage,
		&count,
	); err != nil {
		return shared.VulnerabilityMatch{}, 0, err
	}
	return match, count, nil
})

func (s *store) ScanMatches(ctx context.Context) (err error) {
	ctx, _, endObservation := s.operations.scanMatches.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	matches, err := scanFilteredVulnerabilityMatches(tx.Query(ctx, sqlf.Sprintf(scanMatchesQuery)))
	if err != nil {
		return err
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(scanMatchesTemporaryTableQuery)); err != nil {
		return err
	}

	if err := batch.WithInserter(
		ctx,
		tx.Handle(),
		"t_vulnerability_affected_packages",
		batch.MaxNumPostgresParameters,
		[]string{
			"upload_id",
			"vulnerability_affected_package_id",
		},
		func(inserter *batch.Inserter) error {
			for _, match := range matches {
				if err := inserter.Insert(
					ctx,
					match.UploadID,
					match.VulnerabilityAffectedPackageID,
				); err != nil {
					return err
				}
			}

			return nil
		},
	); err != nil {
		return err
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(scanMatchesUpdateQuery)); err != nil {
		return err
	}

	return nil
}

const scanMatchesQuery = `
SELECT
	r.dump_id,
	vap.id,
	r.version,
	vap.version_constraint
FROM vulnerability_affected_packages vap
-- TODO - do we need to do the inverse as well?
JOIN lsif_references r ON r.name LIKE '%%' || vap.package_name || '%%'
-- TODO - refine this match
WHERE
	(r.scheme = 'gomod' AND vap.language = 'go') OR
	(r.scheme = 'npm' AND vap.language = 'Javascript')
	-- TODO - java mapping
`

const scanMatchesTemporaryTableQuery = `
CREATE TEMPORARY TABLE t_vulnerability_affected_packages (
	upload_id                          INT NOT NULL,
	vulnerability_affected_package_id  INT NOT NULL
) ON COMMIT DROP
`

const scanMatchesUpdateQuery = `
INSERT INTO vulnerability_matches (upload_id, vulnerability_affected_package_id)
SELECT upload_id, vulnerability_affected_package_id FROM t_vulnerability_affected_packages
ON CONFLICT DO NOTHING
`

type VulnerabilityMatch struct {
	UploadID                       int
	VulnerabilityAffectedPackageID int
}

var scanFilteredVulnerabilityMatches = basestore.NewFilteredSliceScanner(func(s dbutil.Scanner) (m VulnerabilityMatch, _ bool, _ error) {
	var (
		version            string
		versionConstraints []string
	)

	if err := s.Scan(&m.UploadID, &m.VulnerabilityAffectedPackageID, &version, pq.Array(&versionConstraints)); err != nil {
		return VulnerabilityMatch{}, false, err
	}

	return m, versionMatchesConstraints(version, versionConstraints), nil
})

func versionMatchesConstraints(versionString string, constraints []string) bool {
	v, err := version.NewVersion(versionString)
	if err != nil {
		// TODO - log like an adult, you idiot.
		fmt.Printf("CANNOT PARSE VERSION: %q\n", versionString)
		return false
	}

	constraint, err := version.NewConstraint(strings.Join(constraints, ","))
	if err != nil {
		// TODO - log like an adult, you idiot.
		fmt.Printf("CANNOT PARSE CONSTRAINT: %q\n", versionString)
		return false
	}

	return constraint.Check(v) || true // TODO - true only for testing
}
