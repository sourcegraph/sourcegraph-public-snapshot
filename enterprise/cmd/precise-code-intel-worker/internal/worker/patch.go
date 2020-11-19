package worker

import (
	"sort"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/correlation"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
)

func patchData(base, patch *correlation.GroupedBundleDataMaps, reindexedFiles []string, fileStatus map[string]gitserver.Status) (err error) {
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
	unifyRangeIDs(base.Documents, patch.Meta, patch.Documents, patch.ResultChunks, fileStatus)

	mergeData(pathsToCopy, fileStatus, base, patch)

	for path := range pathsToCopy {
		base.Documents[path] = patch.Documents[path]
	}

	for path, status := range fileStatus {
		if status == gitserver.Deleted {
			delete(base.Documents, path)
		}
	}

	return nil
}

func mergeData(pathsToCopy map[string]struct{}, fileStatus map[string]gitserver.Status, base, patch *correlation.GroupedBundleDataMaps) (err error) {
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
					patch.Documents[patchPath].Ranges[patchRef.RangeID] = lsifstore.RangeData{
						StartLine:          rng.StartLine,
						StartCharacter:     rng.StartCharacter,
						EndLine:            rng.EndLine,
						EndCharacter:       rng.EndCharacter,
						DefinitionResultID: defID,
						ReferenceResultID:  refID,
						HoverResultID:      rng.HoverResultID,
						MonikerIDs:         rng.MonikerIDs,
					}
				}
			}

			baseRefChunk.DocumentIDRangeIDs[refID] = baseRefs
			baseDefChunk.DocumentIDRangeIDs[defID] = baseDefs
		}
	}

	return nil
}

func removeRefsIn(paths map[string]struct{}, data *correlation.GroupedBundleDataMaps) {
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

func unifyRangeIDs(updateToDocs map[string]lsifstore.DocumentData, toUpdateMeta lsifstore.MetaData, toUpdateDocs map[string]lsifstore.DocumentData, toUpdateChunks map[int]lsifstore.ResultChunkData, fileStatus map[string]gitserver.Status) error {
	updatedRngIDs := make(map[lsifstore.ID]lsifstore.ID)
	resultsToUpdate := make(map[lsifstore.ID]struct{})

	for path, toUpdateDoc := range toUpdateDocs {
		pathUpdatedRngIDs := make(map[lsifstore.ID]lsifstore.ID)
		if fileStatus[path] == gitserver.Unchanged {
			updateToDoc := updateToDocs[path]

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
		results, chunk := getDefRef(resultID, toUpdateMeta, toUpdateChunks)
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

func sortedRangeIDs(ranges map[lsifstore.ID]lsifstore.RangeData) []lsifstore.ID {
	var rngIDs []lsifstore.ID
	for rngID := range ranges {
		rngIDs = append(rngIDs, rngID)
	}

	sort.Slice(rngIDs, func(i, j int) bool {
		return lsifstore.CompareRanges(ranges[rngIDs[i]], ranges[rngIDs[j]]) < 0
	})

	return rngIDs
}

func getDefRef(resultID lsifstore.ID, meta lsifstore.MetaData, resultChunks map[int]lsifstore.ResultChunkData) ([]lsifstore.DocumentIDRangeID, lsifstore.ResultChunkData) {
	chunkID := lsifstore.HashKey(resultID, meta.NumResultChunks)
	chunk := resultChunks[chunkID]
	docRngIDs := chunk.DocumentIDRangeIDs[resultID]
	return docRngIDs, chunk
}

func newID() (lsifstore.ID, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return lsifstore.ID(uuid.String()), nil
}
