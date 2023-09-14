package store

import (
	"context"
	"encoding/json"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/sentinel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *store) VulnerabilityByID(ctx context.Context, id int) (_ shared.Vulnerability, _ bool, err error) {
	ctx, _, endObservation := s.operations.vulnerabilityByID.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("id", id),
	}})
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
	v.published_at,
	v.modified_at,
	v.withdrawn_at
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

func (s *store) GetVulnerabilitiesByIDs(ctx context.Context, ids ...int) (_ []shared.Vulnerability, err error) {
	ctx, _, endObservation := s.operations.getVulnerabilitiesByIDs.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("numIDs", len(ids)),
	}})
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
	ctx, _, endObservation := s.operations.getVulnerabilities.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("limit", args.Limit),
		attribute.Int("offset", args.Offset),
	}})
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

func (s *store) InsertVulnerabilities(ctx context.Context, vulnerabilities []shared.Vulnerability) (_ int, err error) {
	ctx, _, endObservation := s.operations.insertVulnerabilities.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("numVulnerabilities", len(vulnerabilities)),
	}})
	defer endObservation(1, observation.Args{})

	vulnerabilities = canonicalizeVulnerabilities(vulnerabilities)

	var a int
	err = s.db.WithTransact(ctx, func(tx *basestore.Store) error {
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
				"published_at",
				"modified_at",
				"withdrawn_at",
			},
			func(inserter *batch.Inserter) error {
				for _, v := range vulnerabilities {
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
						v.PublishedAt,
						dbutil.NullTime{Time: v.ModifiedAt},
						dbutil.NullTime{Time: v.WithdrawnAt},
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

		count, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(insertVulnerabilitiesUpdateQuery)))
		if err != nil {
			return err
		}
		if err := tx.Exec(ctx, sqlf.Sprintf(insertVulnerabilitiesAffectedPackagesUpdateQuery)); err != nil {
			return err
		}
		if err := tx.Exec(ctx, sqlf.Sprintf(insertVulnerabilitiesAffectedSymbolsUpdateQuery)); err != nil {
			return err
		}

		a = count
		return nil
	})

	return a, err
}

const insertVulnerabilitiesTemporaryVulnerabilitiesTableQuery = `
CREATE TEMPORARY TABLE t_vulnerabilities (
	source_id     TEXT NOT NULL,
	summary       TEXT NOT NULL,
	details       TEXT NOT NULL,
	cpes          TEXT[] NOT NULL,
	cwes          TEXT[] NOT NULL,
	aliases       TEXT[] NOT NULL,
	related       TEXT[] NOT NULL,
	data_source   TEXT NOT NULL,
	urls          TEXT[] NOT NULL,
	severity      TEXT NOT NULL,
	cvss_vector   TEXT NOT NULL,
	cvss_score    TEXT NOT NULL,
	published_at  TIMESTAMP WITH TIME ZONE NOT NULL,
	modified_at   TIMESTAMP WITH TIME ZONE,
	withdrawn_at  TIMESTAMP WITH TIME ZONE
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
WITH ins AS (
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
		published_at,
		modified_at,
		withdrawn_at
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
		published_at,
		modified_at,
		withdrawn_at
	FROM t_vulnerabilities
	-- TODO - we'd prefer to update rather than keep first write
	ON CONFLICT DO NOTHING
	RETURNING 1
)
SELECT COUNT(*) FROM ins
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
-- TODO - we'd prefer to update rather than keep first write
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
-- TODO - we'd prefer to update rather than keep first write
ON CONFLICT DO NOTHING
`

//
//

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
		&v.PublishedAt,
		&v.ModifiedAt,
		&v.WithdrawnAt,
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
		if len(flattened) == 0 || flattened[i].Namespace != pkg.Namespace || flattened[i].Language != pkg.Language || flattened[i].PackageName != pkg.PackageName {
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

func canonicalizeVulnerabilities(vs []shared.Vulnerability) []shared.Vulnerability {
	for i, v := range vs {
		vs[i] = canonicalizeVulnerability(v)
	}

	return vs
}

func canonicalizeVulnerability(v shared.Vulnerability) shared.Vulnerability {
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
	for i, ap := range v.AffectedPackages {
		v.AffectedPackages[i] = canonicalizeAffectedPackage(ap)
	}

	return v
}

func canonicalizeAffectedPackage(ap shared.AffectedPackage) shared.AffectedPackage {
	if ap.VersionConstraint == nil {
		ap.VersionConstraint = []string{}
	}
	if ap.AffectedSymbols == nil {
		ap.AffectedSymbols = []shared.AffectedSymbol{}
	}

	return ap
}
