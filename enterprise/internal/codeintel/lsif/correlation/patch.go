package correlation

import (
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bloomfilter"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
)

func PatchData(base, patch *GroupedBundleDataMaps, reindexedFiles []string, fileStatus map[string]gitserver.Status) (err error) {
	reindexed := make(map[string]struct{})
	for _, file := range reindexedFiles {
		reindexed[file] = struct{}{}
	}

	modifiedOrDeletedPaths := make(map[string]struct{})
	for path, status := range fileStatus {
		if status == gitserver.Modified || status == gitserver.Deleted {
			modifiedOrDeletedPaths[path] = struct{}{}
		}
	}
	removeRefsIn(modifiedOrDeletedPaths, base)

	pathsToCopy := make(map[string]struct{})
	unmodifiedReindexedPaths := make(map[string]struct{})
	for path := range reindexed {
		pathsToCopy[path] = struct{}{}
		if fileStatus[path] == gitserver.Unchanged {
			unmodifiedReindexedPaths[path] = struct{}{}
		}
	}
	for path, status := range fileStatus {
		if status == gitserver.Added {
			pathsToCopy[path] = struct{}{}
		}
	}
	unifyRangeIDs(base, patch, fileStatus)
	unifyMonikerVersions(base, patch)
	unifyDefsRefs(base, patch, pathsToCopy, fileStatus)
	unifyMonikers(base.Definitions, patch.Definitions, pathsToCopy, fileStatus)
	unifyMonikers(base.References, patch.References, pathsToCopy, fileStatus)

	for path := range pathsToCopy {
		base.Documents[path] = patch.Documents[path]
	}

	for path, status := range fileStatus {
		if status == gitserver.Deleted {
			delete(base.Documents, path)
		}
	}

	recreatePackageData(base)

	return nil
}

func unifyDefsRefs(base, patch *GroupedBundleDataMaps, pathsToCopy map[string]struct{}, fileStatus map[string]gitserver.Status) (err error) {
	defResultsByPath := make(map[string]map[lsifstore.ID]lsifstore.RangeData)

	for path := range pathsToCopy {
		for _, rng := range patch.Documents[path].Ranges {
			if rng.DefinitionResultID == "" {
				continue
			}
			defs, defChunk := getDefRef(rng.DefinitionResultID, patch.Meta, patch.ResultChunks)
			for _, defLoc := range defs {
				defPath := defChunk.DocumentPaths[defLoc.DocumentID]
				def := patch.Documents[defPath].Ranges[defLoc.RangeID]
				defResults, exists := defResultsByPath[defPath]
				if !exists {
					defResults = make(map[lsifstore.ID]lsifstore.RangeData)
					defResultsByPath[defPath] = defResults
				}
				if _, exists := defResults[defLoc.RangeID]; !exists {
					defResults[defLoc.RangeID] = def
				}
			}
		}
	}

	for path, defsMap := range defResultsByPath {
		baseDoc := base.Documents[path]
		defIdxs := sortedRangeIDs(defsMap)
		for _, defRngID := range defIdxs {
			def := defsMap[defRngID]
			var defID, refID lsifstore.ID
			if fileStatus[path] == gitserver.Unchanged {
				baseRng := baseDoc.Ranges[defRngID]

				defID = baseRng.DefinitionResultID
				refID = baseRng.ReferenceResultID
			} else {
				defID, err = newID()
				if err != nil {
					return err
				}
				refID, err = newID()
				if err != nil {
					return err
				}
			}

			patchRefs, patchRefChunk := getDefRef(def.ReferenceResultID, patch.Meta, patch.ResultChunks)

			patchDefs, patchDefChunk := getDefRef(def.DefinitionResultID, patch.Meta, patch.ResultChunks)
			baseRefs, baseRefChunk := getDefRef(refID, base.Meta, base.ResultChunks)
			baseDefs, baseDefChunk := getDefRef(defID, base.Meta, base.ResultChunks)

			baseRefDocumentIDs := make(map[string]lsifstore.ID)
			for id, path := range baseRefChunk.DocumentPaths {
				baseRefDocumentIDs[path] = id
			}
			baseDefDocumentIDs := make(map[string]lsifstore.ID)
			for id, path := range baseDefChunk.DocumentPaths {
				baseDefDocumentIDs[path] = id
			}
			for _, patchRef := range patchRefs {
				patchPath := patchRefChunk.DocumentPaths[patchRef.DocumentID]
				if fileStatus[patchPath] != gitserver.Unchanged {
					baseRefDocumentID, exists := baseRefDocumentIDs[path]
					if !exists {
						baseRefDocumentID, err = newID()
						if err != nil {
							return err
						}
						baseRefDocumentIDs[path] = baseRefDocumentID
						baseRefChunk.DocumentPaths[baseRefDocumentID] = path
					}
					patchRef.DocumentID = baseRefDocumentID
					baseRefs = append(baseRefs, patchRef)

				}

				if len(baseDefs) == 0 {
					var patchDef *lsifstore.DocumentIDRangeID
					for _, tmpDef := range patchDefs {
						patchDefPath := patchDefChunk.DocumentPaths[tmpDef.DocumentID]
						if patchDefPath == patchPath && tmpDef.RangeID == patchRef.RangeID {
							patchDef = &tmpDef
						}
					}
					if patchDef != nil {
						baseDefDocumentID, exists := baseDefDocumentIDs[path]
						if !exists {
							baseDefDocumentID, err = newID()
							if err != nil {
								return err
							}
							baseDefDocumentIDs[path] = baseDefDocumentID
							baseDefChunk.DocumentPaths[baseDefDocumentID] = path
						}
						patchDef.DocumentID = baseDefDocumentID
						baseDefs = append(baseDefs, *patchDef)
					}
				}

				if _, exists := pathsToCopy[patchPath]; exists {
					rng := patch.Documents[patchPath].Ranges[patchRef.RangeID]
					rng.DefinitionResultID = defID
					rng.ReferenceResultID = refID
					patch.Documents[patchPath].Ranges[patchRef.RangeID] = rng
				}
			}

			baseRefChunk.DocumentIDRangeIDs[refID] = baseRefs
			baseDefChunk.DocumentIDRangeIDs[defID] = baseDefs
		}
	}

	return nil
}

func unifyMonikers(base, patch map[string]map[string][]lsifstore.LocationData, pathsToCopy map[string]struct{}, fileStatus map[string]gitserver.Status) {
	for _, identMap := range base {
		for ident, locations := range identMap {
			var filteredLocations []lsifstore.LocationData
			for _, location := range locations {
				if fileStatus[location.URI] != gitserver.Deleted && fileStatus[location.URI] != gitserver.Modified {
					filteredLocations = append(filteredLocations, location)
				}
			}
			if len(filteredLocations) == 0 {
				delete(identMap, ident)
			} else {
				identMap[ident] = filteredLocations
			}
		}
	}

	for scheme, identMap := range patch {
		baseIdentMap, exists := base[scheme]
		if !exists {
			base[scheme] = make(map[string][]lsifstore.LocationData)
			baseIdentMap = base[scheme]
		}

		for ident, locations := range identMap {
			baseLocations := baseIdentMap[ident]
			for _, location := range locations {
				baseLocations = append(baseLocations, location)
			}
			baseIdentMap[ident] = baseLocations
		}
	}
}

func removeRefsIn(paths map[string]struct{}, data *GroupedBundleDataMaps) {
	deletedRefs := make(map[lsifstore.ID]struct{})

	for path := range paths {
		doc := data.Documents[path]
		for _, rng := range doc.Ranges {
			if _, exists := deletedRefs[rng.ReferenceResultID]; exists {
				continue
			}

			refs, refChunk := getDefRef(rng.ReferenceResultID, data.Meta, data.ResultChunks)
			var filteredRefs []lsifstore.DocumentIDRangeID
			for _, ref := range refs {
				refPath := refChunk.DocumentPaths[ref.DocumentID]
				if _, exists := paths[refPath]; !exists {
					filteredRefs = append(filteredRefs, ref)
				}
			}
			refChunk.DocumentIDRangeIDs[rng.ReferenceResultID] = filteredRefs
			deletedRefs[rng.ReferenceResultID] = struct{}{}
		}
	}
}

var unequalUnmodifiedPathsErr = errors.New("The ranges of unmodified path in LSIF patch do not match ranges of the same path in the base LSIF dump.")

func unifyRangeIDs(updateTo, toUpdate *GroupedBundleDataMaps, fileStatus map[string]gitserver.Status) error {
	updatedRngIDs := make(map[lsifstore.ID]lsifstore.ID)
	resultsToUpdate := make(map[lsifstore.ID]struct{})

	for path, toUpdateDoc := range toUpdate.Documents {
		pathUpdatedRngIDs := make(map[lsifstore.ID]lsifstore.ID)
		if fileStatus[path] == gitserver.Unchanged {
			updateToDoc := updateTo.Documents[path]

			updateToRngIDs := sortedRangeIDs(updateToDoc.Ranges)
			toUpdateRng := sortedRangeIDs(toUpdateDoc.Ranges)
			if len(toUpdateRng) != len(updateToRngIDs) {
				return unequalUnmodifiedPathsErr
			}

			for idx, updateToRngID := range updateToRngIDs {
				updateToRng := updateToDoc.Ranges[updateToRngID]
				toUpdateRngID := toUpdateRng[idx]
				toUpdateRng := toUpdateDoc.Ranges[toUpdateRngID]

				if lsifstore.CompareRanges(updateToRng, toUpdateRng) != 0 {
					return unequalUnmodifiedPathsErr
				}

				pathUpdatedRngIDs[toUpdateRngID] = updateToRngID
			}
		} else {
			for rngID := range toUpdateDoc.Ranges {
				newRngID, err := newID()
				if err != nil {
					return err
				}
				updatedRngIDs[rngID] = newRngID
			}
		}

		for oldID, newID := range pathUpdatedRngIDs {
			rng := toUpdateDoc.Ranges[oldID]
			toUpdateDoc.Ranges[newID] = rng
			resultsToUpdate[rng.ReferenceResultID] = struct{}{}
			resultsToUpdate[rng.DefinitionResultID] = struct{}{}
			delete(toUpdateDoc.Ranges, oldID)
		}
	}

	for resultID := range resultsToUpdate {
		results, chunk := getDefRef(resultID, toUpdate.Meta, toUpdate.ResultChunks)
		var updated []lsifstore.DocumentIDRangeID
		for _, result := range results {
			if updatedID, exists := updatedRngIDs[result.RangeID]; exists {
				updated = append(updated, lsifstore.DocumentIDRangeID{
					RangeID:    updatedID,
					DocumentID: result.DocumentID,
				})
			} else {
				updated = append(updated, lsifstore.DocumentIDRangeID{
					RangeID:    result.RangeID,
					DocumentID: result.DocumentID,
				})
			}
		}
		chunk.DocumentIDRangeIDs[resultID] = updated
	}

	return nil
}

func unifyMonikerVersions(base, patch *GroupedBundleDataMaps) {
	schemeNameVersionMap := make(map[string]map[string]string)

	for _, document := range patch.Documents {
		for _, moniker := range document.Monikers {
			packageInfo, exists := document.PackageInformation[moniker.PackageInformationID]
			if !exists {
				continue
			}

			nameVersionMap, exists := schemeNameVersionMap[moniker.Scheme]
			if !exists {
				schemeNameVersionMap[moniker.Scheme] = make(map[string]string)
				nameVersionMap = schemeNameVersionMap[moniker.Scheme]
			}
			nameVersionMap[packageInfo.Name] = packageInfo.Version
		}
	}

	for _, document := range base.Documents {
		for _, moniker := range document.Monikers {
			nameVersionMap, exists := schemeNameVersionMap[moniker.Scheme]
			if !exists {
				continue
			}
			packageInfo, exists := document.PackageInformation[moniker.PackageInformationID]
			if !exists {
				continue
			}

			version, exists := nameVersionMap[packageInfo.Name]
			if !exists {
				continue
			}

			packageInfo.Version = version
			document.PackageInformation[moniker.PackageInformationID] = packageInfo
		}
	}
}

func recreatePackageData(base *GroupedBundleDataMaps) error {
	type ExpandedPackageReference struct {
		Scheme      string
		Name        string
		Version     string
		Identifiers []string
	}

	exports := make(map[string]lsifstore.Package)
	imports := make(map[string]ExpandedPackageReference)

	for _, document := range base.Documents {
		for _, moniker := range document.Monikers {
			packageInfo, exists := document.PackageInformation[moniker.PackageInformationID]
			if !exists {
				continue
			}

			key := makeKey(moniker.Scheme, packageInfo.Name, packageInfo.Version)
			if moniker.Kind == "import" {
				imports[key] = ExpandedPackageReference{
					Scheme:      moniker.Scheme,
					Name:        packageInfo.Name,
					Version:     packageInfo.Version,
					Identifiers: append(imports[key].Identifiers, moniker.Identifier),
				}
			} else if moniker.Kind == "export" {
				exports[key] = lsifstore.Package{
					DumpID:  0, // TODO
					Scheme:  moniker.Scheme,
					Name:    packageInfo.Name,
					Version: packageInfo.Version,
				}
			}
		}
	}

	exportList := make([]lsifstore.Package, 0, len(exports))
	for _, export := range exports {
		exportList = append(exportList, export)
	}

	importList := make([]lsifstore.PackageReference, 0, len(imports))
	for _, imp := range imports {
		filter, err := bloomfilter.CreateFilter(imp.Identifiers)
		if err != nil {
			return errors.Wrap(err, "bloomfilter.CreateFilter")
		}

		importList = append(importList, lsifstore.PackageReference{
			DumpID:  0, // TODO
			Scheme:  imp.Scheme,
			Name:    imp.Name,
			Version: imp.Version,
			Filter:  filter,
		})
	}

	base.Packages = exportList
	base.PackageReferences = importList

	return nil
}
