package lsifstore

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GetMonikersByPosition returns all monikers attached ranges containing the given position. If multiple
// ranges contain the position, then this method will return multiple sets of monikers. Each slice
// of monikers are attached to a single range. The order of the output slice is "outside-in", so that
// the range attached to earlier monikers enclose the range attached to later monikers.
func (s *store) GetMonikersByPosition(ctx context.Context, uploadID int, path core.UploadRelPath, line, character int) (_ [][]precise.MonikerData, err error) {
	ctx, trace, endObservation := s.operations.getMonikersByPosition.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("uploadID", uploadID),
		attribute.String("path", path.RawValue()),
		attribute.Int("line", line),
		attribute.Int("character", character),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.db.Query(ctx, sqlf.Sprintf(
		monikersDocumentQuery,
		uploadID,
		path.RawValue(),
	)))
	if err != nil || !exists {
		return nil, err
	}

	trace.AddEvent("TODO Domain Owner", attribute.Int("numOccurrences", len(documentData.SCIPData.Occurrences)))
	occurrences := scip.FindOccurrences(documentData.SCIPData.Occurrences, int32(line), int32(character))
	trace.AddEvent("TODO Domain Owner", attribute.Int("numIntersectingOccurrences", len(occurrences)))

	// Make lookup map of symbol information by name
	symbolMap := map[string]*scip.SymbolInformation{}
	for _, symbol := range documentData.SCIPData.Symbols {
		symbolMap[symbol.Symbol] = symbol
	}

	monikerData := make([][]precise.MonikerData, 0, len(occurrences))
	for _, occurrence := range occurrences {
		if occurrence.Symbol == "" || scip.IsLocalSymbol(occurrence.Symbol) {
			continue
		}
		symbol, hasSymbol := symbolMap[occurrence.Symbol]

		kind := precise.Import
		if hasSymbol {
			for _, o := range documentData.SCIPData.Occurrences {
				if o.Symbol == occurrence.Symbol {
					// TODO - do we need to check additional documents here?
					if isDefinition := scip.SymbolRole_Definition.Matches(o); isDefinition {
						kind = precise.Export
					}

					break
				}
			}
		}

		moniker, err := symbolNameToQualifiedMoniker(occurrence.Symbol, kind)
		if err != nil {
			return nil, err
		}
		occurrenceMonikers := []precise.MonikerData{moniker}

		if hasSymbol {
			for _, rel := range symbol.Relationships {
				if rel.IsImplementation {
					relatedMoniker, err := symbolNameToQualifiedMoniker(rel.Symbol, precise.Implementation)
					if err != nil {
						return nil, err
					}

					occurrenceMonikers = append(occurrenceMonikers, relatedMoniker)
				}
			}
		}

		monikerData = append(monikerData, occurrenceMonikers)
	}
	trace.AddEvent("TODO Domain Owner", attribute.Int("numMonikers", len(monikerData)))

	return monikerData, nil
}

const monikersDocumentQuery = `
SELECT
	sd.id,
	sid.document_path,
	sd.raw_scip_payload
FROM codeintel_scip_document_lookup sid
JOIN codeintel_scip_documents sd ON sd.id = sid.document_id
WHERE
	sid.upload_id = %s AND
	sid.document_path = %s
LIMIT 1
`

// GetPackageInformation returns package information data by identifier.
func (s *store) GetPackageInformation(ctx context.Context, bundleID int, packageInformationID string) (_ precise.PackageInformationData, _ bool, err error) {
	_, _, endObservation := s.operations.getPackageInformation.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("bundleID", bundleID),
		attribute.String("packageInformationID", packageInformationID),
	}})
	defer endObservation(1, observation.Args{})

	if strings.HasPrefix(packageInformationID, "scip:") {
		packageInfo := strings.Split(packageInformationID, ":")
		if len(packageInfo) != 4 {
			return precise.PackageInformationData{}, false, errors.Newf("invalid package information ID %q", packageInformationID)
		}

		manager, err := base64.RawStdEncoding.DecodeString(packageInfo[1])
		if err != nil {
			return precise.PackageInformationData{}, false, err
		}
		name, err := base64.RawStdEncoding.DecodeString(packageInfo[2])
		if err != nil {
			return precise.PackageInformationData{}, false, err
		}
		version, err := base64.RawStdEncoding.DecodeString(packageInfo[3])
		if err != nil {
			return precise.PackageInformationData{}, false, err
		}

		return precise.PackageInformationData{
			Manager: string(manager),
			Name:    string(name),
			Version: string(version),
		}, true, nil
	}

	return precise.PackageInformationData{}, false, nil
}

func symbolNameToQualifiedMoniker(symbolName, kind string) (precise.MonikerData, error) {
	parsedSymbol, err := scip.ParseSymbol(symbolName)
	if err != nil {
		return precise.MonikerData{}, err
	}

	return precise.MonikerData{
		Scheme:     parsedSymbol.Scheme,
		Kind:       kind,
		Identifier: symbolName,
		PackageInformationID: precise.ID(strings.Join([]string{
			"scip",
			// Base64 encoding these components as names converted from LSIF contain colons as part of the
			// general moniker scheme. It's reasonable for manager and names in SCIP-land to also have colons,
			// so we'll just remove the ambiguity from the generated string here.
			base64.RawStdEncoding.EncodeToString([]byte(parsedSymbol.Package.Manager)),
			base64.RawStdEncoding.EncodeToString([]byte(parsedSymbol.Package.Name)),
			base64.RawStdEncoding.EncodeToString([]byte(parsedSymbol.Package.Version)),
		}, ":")),
	}, nil
}
