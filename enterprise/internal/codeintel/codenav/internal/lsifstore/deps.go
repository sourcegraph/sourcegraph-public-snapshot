package lsifstore

import (
	"context"
	"fmt"
	"sort"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type DependencyDescription struct {
	Manager string
	Name    string
	Version string
}

// GetDependencies returns a list of dependencies for the given index.
func (s *store) GetDependencies(ctx context.Context, bundleIDs []int) (_ []DependencyDescription, err error) {
	ctx, _, endObservation := s.operations.getExists.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("bundleIDs", intsToString(bundleIDs)),
	}})
	defer endObservation(1, observation.Args{})

	symbolNames, err := basestore.ScanStrings(s.db.Query(ctx, sqlf.Sprintf(
		dependenciesQuery,
		pq.Array(bundleIDs),
	)))
	if err != nil {
		return nil, err
	}

	symbolsByKey := map[string]*scip.Symbol{}
	for _, symbolName := range symbolNames {
		symbol, err := scip.ParseSymbol(symbolName)
		if err != nil {
			return nil, err
		}

		symbolsByKey[fmt.Sprintf("%s:%s:%s", symbol.Package.Manager, symbol.Package.Name, symbol.Package.Version)] = symbol
	}

	var descriptions []DependencyDescription
	for _, symbol := range symbolsByKey {
		descriptions = append(descriptions, DependencyDescription{
			Manager: symbol.Package.Manager,
			Name:    symbol.Package.Name,
			Version: symbol.Package.Version,
		})
	}
	sort.Slice(descriptions, func(i, j int) bool {
		di := descriptions[i]
		dj := descriptions[j]

		if di.Manager == dj.Manager {
			if di.Name == dj.Name {
				return di.Version < dj.Version
			}

			return di.Name < dj.Name
		}

		return di.Manager < dj.Manager
	})

	return descriptions, nil
}

//
// TODO - build top down instead?

const dependenciesQuery = `
WITH RECURSIVE
all_prefixes(upload_id, id, prefix) AS (
	(
		SELECT
			ssn.upload_id,
			ssn.id,
			ssn.name_segment
		FROM codeintel_scip_symbol_names ssn
		WHERE
			ssn.upload_id = ANY(%s) AND
			ssn.prefix_id IS NULL
	) UNION (
		SELECT
			ssn.upload_id,
			ssn.id,
			mp.prefix || ssn.name_segment
		FROM all_prefixes mp
		JOIN codeintel_scip_symbol_names ssn ON
			ssn.upload_id = mp.upload_id AND
			ssn.prefix_id = mp.id
	)
)

SELECT mp.prefix FROM all_prefixes mp
WHERE NOT EXISTS (
	SELECT 1
	FROM codeintel_scip_symbol_names ssn
	WHERE ssn.prefix_id = mp.id
)
`
