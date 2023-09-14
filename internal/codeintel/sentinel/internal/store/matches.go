package store

import (
	"context"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/sentinel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *store) VulnerabilityMatchByID(ctx context.Context, id int) (_ shared.VulnerabilityMatch, _ bool, err error) {
	ctx, _, endObservation := s.operations.vulnerabilityMatchByID.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("id", id),
	}})
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
	vap.package_name,
	vap.language,
	vap.namespace,
	vap.version_constraint,
	vap.fixed,
	vap.fixed_in,
	vas.path,
	vas.symbols,
	vul.severity,
	0 AS count
FROM vulnerability_matches m
LEFT JOIN vulnerability_affected_packages vap ON vap.id = m.vulnerability_affected_package_id
LEFT JOIN vulnerability_affected_symbols vas ON vas.vulnerability_affected_package_id = vap.id
LEFT JOIN vulnerabilities vul ON vap.vulnerability_id = vul.id
WHERE m.id = %s
`

func (s *store) GetVulnerabilityMatches(ctx context.Context, args shared.GetVulnerabilityMatchesArgs) (_ []shared.VulnerabilityMatch, _ int, err error) {
	ctx, _, endObservation := s.operations.getVulnerabilityMatches.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("limit", args.Limit),
		attribute.Int("offset", args.Offset),
		attribute.String("severity", args.Severity),
		attribute.String("language", args.Language),
		attribute.String("repositoryName", args.RepositoryName),
	}})
	defer endObservation(1, observation.Args{})

	var conds []*sqlf.Query
	if args.Language != "" {
		conds = append(conds, sqlf.Sprintf("vap.language = %s", args.Language))
	}
	if args.Severity != "" {
		conds = append(conds, sqlf.Sprintf("vul.severity = %s", args.Severity))
	}
	if args.RepositoryName != "" {
		conds = append(conds, sqlf.Sprintf("r.name = %s", args.RepositoryName))
	}
	if len(conds) == 0 {
		conds = append(conds, sqlf.Sprintf("TRUE"))
	}

	return scanVulnerabilityMatchesAndCount(s.db.Query(ctx, sqlf.Sprintf(getVulnerabilityMatchesQuery, sqlf.Join(conds, " AND "), args.Limit, args.Offset)))
}

const getVulnerabilityMatchesQuery = `
WITH limited_matches AS (
	SELECT
		m.id,
		m.upload_id,
		m.vulnerability_affected_package_id
	FROM vulnerability_matches m
	ORDER BY id
)
SELECT
	m.id,
	m.upload_id,
	vap.vulnerability_id,
	vap.package_name,
	vap.language,
	vap.namespace,
	vap.version_constraint,
	vap.fixed,
	vap.fixed_in,
	vas.path,
	vas.symbols,
	vul.severity,
	COUNT(*) OVER() AS count
FROM limited_matches m
LEFT JOIN vulnerability_affected_packages vap ON vap.id = m.vulnerability_affected_package_id
LEFT JOIN vulnerability_affected_symbols vas ON vas.vulnerability_affected_package_id = vap.id
LEFT JOIN vulnerabilities vul ON vap.vulnerability_id = vul.id
LEFT JOIN lsif_uploads lu ON m.upload_id = lu.id
LEFT JOIN repo r ON r.id = lu.repository_id
WHERE %s
ORDER BY m.id, vap.id, vas.id
LIMIT %s OFFSET %s
`

func (s *store) GetVulnerabilityMatchesSummaryCount(ctx context.Context) (counts shared.GetVulnerabilityMatchesSummaryCounts, err error) {
	ctx, _, endObservation := s.operations.getVulnerabilityMatchesSummaryCount.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	row := s.db.QueryRow(ctx, sqlf.Sprintf(getVulnerabilityMatchesSummaryCounts))
	err = row.Scan(
		&counts.High,
		&counts.Medium,
		&counts.Low,
		&counts.Critical,
		&counts.Repositories,
	)
	if err != nil {
		return shared.GetVulnerabilityMatchesSummaryCounts{}, err
	}

	return counts, nil
}

const getVulnerabilityMatchesSummaryCounts = `
	WITH limited_matches AS (
	SELECT
		m.id,
		m.upload_id,
		m.vulnerability_affected_package_id
	FROM vulnerability_matches m
	ORDER BY id
)
SELECT
  sum(case when vul.severity = 'HIGH' then 1 else 0 end) as high,
  sum(case when vul.severity = 'MEDIUM' then 1 else 0 end) as medium,
  sum(case when vul.severity = 'LOW' then 1 else 0 end) as low,
  sum(case when vul.severity = 'CRITICAL' then 1 else 0 end) as critical,
  count(distinct r.name) as repositories
FROM limited_matches m
LEFT JOIN vulnerability_affected_packages vap ON vap.id = m.vulnerability_affected_package_id
LEFT JOIN vulnerability_affected_symbols vas ON vas.vulnerability_affected_package_id = vap.id
LEFT JOIN vulnerabilities vul ON vap.vulnerability_id = vul.id
LEFT JOIN lsif_uploads lu ON lu.id = m.upload_id
LEFT JOIN repo r ON r.id = lu.repository_id
`

func (s *store) GetVulnerabilityMatchesCountByRepository(ctx context.Context, args shared.GetVulnerabilityMatchesCountByRepositoryArgs) (_ []shared.VulnerabilityMatchesByRepository, _ int, err error) {
	ctx, _, endObservation := s.operations.getVulnerabilityMatchesCountByRepository.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("limit", args.Limit),
		attribute.Int("offset", args.Offset),
		attribute.String("repositoryName", args.RepositoryName),
	}})
	defer endObservation(1, observation.Args{})

	var conds []*sqlf.Query
	if args.RepositoryName != "" {
		conds = append(conds, sqlf.Sprintf("r.name ILIKE %s", "%"+args.RepositoryName+"%"))
	}
	if len(conds) == 0 {
		conds = append(conds, sqlf.Sprintf("TRUE"))
	}

	rows, err := s.db.Query(ctx, sqlf.Sprintf(getVulnerabilityMatchesGroupedByRepos, sqlf.Join(conds, " AND "), args.Limit, args.Offset))
	if err != nil {
		return nil, 0, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var matches []shared.VulnerabilityMatchesByRepository
	var totalCount int
	for rows.Next() {
		var match shared.VulnerabilityMatchesByRepository
		if err := rows.Scan(&match.ID, &match.RepositoryName, &match.MatchCount, &totalCount); err != nil {
			return nil, 0, err
		}

		matches = append(matches, match)
	}

	return matches, totalCount, nil
}

const getVulnerabilityMatchesGroupedByRepos = `
select
	r.id,
	r.name,
	count(*) as count,
	COUNT(*) OVER() AS total_count
from vulnerability_matches vm
join lsif_uploads lu on lu.id = vm.upload_id
join repo r on r.id = lu.repository_id
where %s
group by r.name, r.id
order by count DESC
limit %s offset %s
`

//
//

func (s *store) ScanMatches(ctx context.Context, batchSize int) (numReferencesScanned int, numVulnerabilityMatches int, err error) {
	ctx, _, endObservation := s.operations.scanMatches.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("batchSize", batchSize),
	}})
	defer endObservation(1, observation.Args{})

	var a, b int
	err = s.db.WithTransact(ctx, func(tx *basestore.Store) error {
		type vulnerabilityMatch struct {
			UploadID                       int
			VulnerabilityAffectedPackageID int
		}
		numScanned := 0
		scanFilteredVulnerabilityMatches := basestore.NewFilteredSliceScanner(func(s dbutil.Scanner) (m vulnerabilityMatch, _ bool, _ error) {
			var (
				version            string
				versionConstraints []string
			)

			if err := s.Scan(&m.UploadID, &m.VulnerabilityAffectedPackageID, &version, pq.Array(&versionConstraints)); err != nil {
				return vulnerabilityMatch{}, false, err
			}

			numScanned++
			matches, valid := versionMatchesConstraints(version, versionConstraints)
			_ = valid // TODO - log un-parseable versions

			return m, matches, nil
		})

		matches, err := scanFilteredVulnerabilityMatches(tx.Query(ctx, sqlf.Sprintf(
			scanMatchesQuery,
			batchSize,
			sqlf.Join(makeSchemeTtoVulnerabilityLanguageMappingConditions(), " OR "),
		)))
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

		numMatched, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(scanMatchesUpdateQuery)))
		if err != nil {
			return err
		}

		a = numScanned
		b = numMatched
		return nil
	})

	return a, b, err
}

const scanMatchesQuery = `
WITH
candidates AS (
	SELECT u.id
	FROM lsif_uploads u
	JOIN repo r ON r.id = u.repository_id
	WHERE
		u.state = 'completed' AND
		r.deleted_at IS NULL AND
		r.blocked IS NULL AND
		NOT EXISTS (
			SELECT 1
			FROM lsif_uploads_vulnerability_scan uvs
			WHERE
				uvs.upload_id = u.id AND
				-- TODO: we'd rather compare this against vuln update times
				uvs.last_scanned_at < NOW()
		)
	ORDER BY u.id
	LIMIT %s
),
locked_candidates AS (
	INSERT INTO lsif_uploads_vulnerability_scan (upload_id, last_scanned_at)
	SELECT id, NOW() FROM candidates
	ON CONFLICT DO NOTHING
	RETURNING upload_id
)
SELECT
	r.dump_id,
	vap.id,
	r.version,
	vap.version_constraint
FROM locked_candidates lc
JOIN lsif_references r ON r.dump_id = lc.upload_id
JOIN vulnerability_affected_packages vap ON
	-- NOTE: This is currently a bit of a hack that works to find some
	-- good matches with the dataset we have. We should have a better
	-- way to match on a normalized name here, or have rules per types
	-- of language ecosystem.
	r.name LIKE '%%' || vap.package_name || '%%'
WHERE %s
`

const scanMatchesTemporaryTableQuery = `
CREATE TEMPORARY TABLE t_vulnerability_affected_packages (
	upload_id                          INT NOT NULL,
	vulnerability_affected_package_id  INT NOT NULL
) ON COMMIT DROP
`

const scanMatchesUpdateQuery = `
WITH ins AS (
	INSERT INTO vulnerability_matches (upload_id, vulnerability_affected_package_id)
	SELECT upload_id, vulnerability_affected_package_id FROM t_vulnerability_affected_packages
	ON CONFLICT DO NOTHING
	RETURNING 1
)
SELECT COUNT(*) FROM ins
`

//
//

var scanVulnerabilityMatchesAndCount = func(rows basestore.Rows, queryErr error) ([]shared.VulnerabilityMatch, int, error) {
	matches, totalCount, err := basestore.NewSliceWithCountScanner(func(s dbutil.Scanner) (match shared.VulnerabilityMatch, count int, _ error) {
		var (
			vap     shared.AffectedPackage
			vas     shared.AffectedSymbol
			vul     shared.Vulnerability
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
			&dbutil.NullString{S: &vul.Severity},
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

func versionMatchesConstraints(versionString string, constraints []string) (matches, valid bool) {
	v, err := version.NewVersion(versionString)
	if err != nil {
		return false, false
	}

	constraint, err := version.NewConstraint(strings.Join(constraints, ","))
	if err != nil {
		return false, false
	}

	return constraint.Check(v), true
}

var scipSchemeToVulnerabilityLanguage = map[string]string{
	"gomod": "go",
	"npm":   "Javascript",
	// TODO - java mapping
}

func makeSchemeTtoVulnerabilityLanguageMappingConditions() []*sqlf.Query {
	schemes := make([]string, 0, len(scipSchemeToVulnerabilityLanguage))
	for scheme := range scipSchemeToVulnerabilityLanguage {
		schemes = append(schemes, scheme)
	}
	sort.Strings(schemes)

	mappings := make([]*sqlf.Query, 0, len(schemes))
	for _, scheme := range schemes {
		mappings = append(mappings, sqlf.Sprintf("(r.scheme = %s AND vap.language = %s)", scheme, scipSchemeToVulnerabilityLanguage[scheme]))
	}

	return mappings
}
