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
	VulnerabilityByID(ctx context.Context, id int) (shared.Vulnerability, bool, error)
	VulnerabilityMatchByID(ctx context.Context, id int) (shared.VulnerabilityMatch, bool, error)
	GetVulnerabilitiesByIDs(ctx context.Context, ids ...int) ([]shared.Vulnerability, error)
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

func (s *store) VulnerabilityByID(ctx context.Context, id int) (_ shared.Vulnerability, _ bool, err error) {
	ctx, _, endObservation := s.operations.vulnerabilityByID.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	vulnerabilities, _, err := scanVulnerabilitiesAndCount(s.db.Query(ctx, sqlf.Sprintf(getVulnerabilityByIDQuery, id)))
	if err != nil || len(vulnerabilities) == 0 {
		return shared.Vulnerability{}, false, err
	}

	return vulnerabilities[0], true, nil
}

const getVulnerabilityByIDQuery = `
SELECT
	` + vulnerabilityFields + `,
	` + vulnerabilityAffectedPackageFields + `,
	` + vulnerabilityAffectedSymbolFields + `,
	0 AS count
FROM vulnerabilities v
LEFT JOIN vulnerability_affected_packages vap ON vap.vulnerability_id = v.id
LEFT JOIN vulnerability_affected_symbols vas ON vas.vulnerability_affected_package_id = vap.id
WHERE v.id = %s
ORDER BY vap.id, vas.id
`

const vulnerabilityFields = `
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
	v.withdrawn
`

const vulnerabilityAffectedPackageFields = `
	vap.package_name,
	vap.language,
	vap.namespace,
	vap.version_constraint,
	vap.fixed,
	vap.fixed_in
`

const vulnerabilityAffectedSymbolFields = `
	vas.path,
	vas.symbols
`

func (s *store) VulnerabilityMatchByID(ctx context.Context, id int) (_ shared.VulnerabilityMatch, _ bool, err error) {
	ctx, _, endObservation := s.operations.vulnerabilityMatchByID.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	matches, _, err := scanVulnerabilityMatchesAndCount(s.db.Query(ctx, sqlf.Sprintf(vulnerabilityMatchByIDQuery, id)))
	if err != nil || len(matches) == 0 {
		return shared.VulnerabilityMatch{}, false, err
	}

	return matches[0], true, nil
}

const vulnerabilityMatchByIDQuery = `
SELECT
	m.id,
	m.upload_id,
	vap.vulnerability_id,
	` + vulnerabilityAffectedPackageFields + `,
	` + vulnerabilityAffectedSymbolFields + `,
	0 AS count
FROM vulnerability_matches m
LEFT JOIN vulnerability_affected_packages vap ON vap.id = m.vulnerability_affected_package_id
LEFT JOIN vulnerability_affected_symbols vas ON vas.vulnerability_affected_package_id = vap.id
WHERE id = %s
`

func (s *store) GetVulnerabilitiesByIDs(ctx context.Context, ids ...int) (_ []shared.Vulnerability, err error) {
	ctx, _, endObservation := s.operations.getVulnerabilitiesByIDs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	vulnerabilities, _, err := scanVulnerabilitiesAndCount(s.db.Query(ctx, sqlf.Sprintf(getVulnerabilitiesByIDsQuery, pq.Array(ids))))
	return vulnerabilities, err
}

const getVulnerabilitiesByIDsQuery = `
SELECT
	` + vulnerabilityFields + `,
	` + vulnerabilityAffectedPackageFields + `,
	` + vulnerabilityAffectedSymbolFields + `,
	0 AS count
FROM vulnerabilities v
LEFT JOIN vulnerability_affected_packages vap ON vap.vulnerability_id = v.id
LEFT JOIN vulnerability_affected_symbols vas ON vas.vulnerability_affected_package_id = vap.id
WHERE v.id = ANY(%s)
ORDER BY v.id, vap.id, vas.id
`

func (s *store) GetVulnerabilities(ctx context.Context, args shared.GetVulnerabilitiesArgs) (_ []shared.Vulnerability, _ int, err error) {
	ctx, _, endObservation := s.operations.getVulnerabilities.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return scanVulnerabilitiesAndCount(s.db.Query(ctx, sqlf.Sprintf(getVulnerabilitiesQuery, args.Limit, args.Offset)))
}

const getVulnerabilitiesQuery = `
WITH limited_vulnerabilities AS (
	SELECT
		` + vulnerabilityFields + `,
		COUNT(*) OVER() AS count
	FROM vulnerabilities v
	ORDER BY id
	LIMIT %s
	OFFSET %s
)
SELECT
	` + vulnerabilityFields + `,
	` + vulnerabilityAffectedPackageFields + `,
	` + vulnerabilityAffectedSymbolFields + `,
	v.count
FROM limited_vulnerabilities v
LEFT JOIN vulnerability_affected_packages vap ON vap.vulnerability_id = v.id
LEFT JOIN vulnerability_affected_symbols vas ON vas.vulnerability_affected_package_id = vap.id
ORDER BY v.id, vap.id, vas.id
`

var scanSingleVulnerabilityAndCount = func(s dbutil.Scanner) (v shared.Vulnerability, count int, _ error) {
	var (
		vap     shared.AffectedPackage
		vas     shared.AffectedSymbol
		fixedIn string
	)

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
		return shared.Vulnerability{}, 0, err
	}

	if fixedIn != "" {
		vap.FixedIn = &fixedIn
	}
	if vas.Path != "" {
		vap.AffectedSymbols = append(vap.AffectedSymbols, vas)
	}
	if vap.PackageName != "" {
		v.AffectedPackages = append(v.AffectedPackages, vap)
	}

	return v, count, nil
}

var flattenPackages = func(packages []shared.AffectedPackage) []shared.AffectedPackage {
	flattened := []shared.AffectedPackage{}
	for _, pkg := range packages {
		i := len(flattened) - 1
		if len(flattened) == 0 || flattened[i].Namespace != pkg.Namespace || flattened[i].Language != pkg.Language || flattened[i].PackageName != flattened[i].PackageName {
			flattened = append(flattened, pkg)
		} else {
			flattened[i].AffectedSymbols = append(flattened[i].AffectedSymbols, pkg.AffectedSymbols...)
		}
	}

	return flattened
}

var flattenVulnerabilities = func(vs []shared.Vulnerability) []shared.Vulnerability {
	flattened := []shared.Vulnerability{}
	for _, v := range vs {
		i := len(flattened) - 1
		if len(flattened) == 0 || flattened[i].ID != v.ID {
			flattened = append(flattened, v)
		} else {
			flattened[i].AffectedPackages = flattenPackages(append(flattened[i].AffectedPackages, v.AffectedPackages...))
		}
	}

	return flattened
}

var scanVulnerabilitiesAndCount = func(rows basestore.Rows, queryErr error) ([]shared.Vulnerability, int, error) {
	values, totalCount, err := basestore.NewSliceWithCountScanner(func(s dbutil.Scanner) (shared.Vulnerability, int, error) {
		return scanSingleVulnerabilityAndCount(s)
	})(rows, queryErr)
	if err != nil {
		return nil, 0, err
	}

	return flattenVulnerabilities(values), totalCount, nil
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

	return scanVulnerabilityMatchesAndCount(s.db.Query(ctx, sqlf.Sprintf(getVulnerabilityMatchesQuery, args.Limit, args.Offset)))
}

const getVulnerabilityMatchesQuery = `
WITH limited_matches AS (
	SELECT
		m.id,
		m.upload_id,
		m.vulnerability_affected_package_id,
		COUNT(*) OVER() AS count
	FROM vulnerability_matches m
	ORDER BY id
	LIMIT %s
	OFFSET %s
)
SELECT
	m.id,
	m.upload_id,
	vap.vulnerability_id,
	` + vulnerabilityAffectedPackageFields + `,
	` + vulnerabilityAffectedSymbolFields + `,
	COUNT(*) OVER() AS count
FROM limited_matches m
LEFT JOIN vulnerability_affected_packages vap ON vap.id = m.vulnerability_affected_package_id
LEFT JOIN vulnerability_affected_symbols vas ON vas.vulnerability_affected_package_id = vap.id
ORDER BY m.id, vap.id, vas.id
`

var flattenMatches = func(ms []shared.VulnerabilityMatch) []shared.VulnerabilityMatch {
	flattened := []shared.VulnerabilityMatch{}
	for _, m := range ms {
		i := len(flattened) - 1
		if len(flattened) == 0 || flattened[i].ID != m.ID {
			flattened = append(flattened, m)
		} else {
			if flattened[i].AffectedPackage.PackageName == "" {
				flattened[i].AffectedPackage = m.AffectedPackage
			} else {
				symbols := flattened[i].AffectedPackage.AffectedSymbols
				symbols = append(symbols, m.AffectedPackage.AffectedSymbols...)
				flattened[i].AffectedPackage.AffectedSymbols = symbols
			}
		}
	}

	return flattened
}

var scanVulnerabilityMatchesAndCount = func(rows basestore.Rows, queryErr error) ([]shared.VulnerabilityMatch, int, error) {
	matches, totalCount, err := basestore.NewSliceWithCountScanner(func(s dbutil.Scanner) (match shared.VulnerabilityMatch, count int, _ error) {
		var (
			vap     shared.AffectedPackage
			vas     shared.AffectedSymbol
			fixedIn string
		)

		if err := s.Scan(
			&match.ID,
			&match.UploadID,
			&match.VulnerabilityID,
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
			return shared.VulnerabilityMatch{}, 0, err
		}

		if fixedIn != "" {
			vap.FixedIn = &fixedIn
		}
		if vas.Path != "" {
			vap.AffectedSymbols = append(vap.AffectedSymbols, vas)
		}
		if vap.PackageName != "" {
			match.AffectedPackage = vap
		}

		return match, count, nil
	})(rows, queryErr)
	if err != nil {
		return nil, 0, err
	}

	return flattenMatches(matches), totalCount, nil
}

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
