package worker

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"sort"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/correlation"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/uploadstore"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/metrics"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/util"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type handler struct {
	dbStore         DBStore
	lsifStore       LSIFStore
	uploadStore     uploadstore.Store
	gitserverClient GitserverClient
	enableBudget    bool
	budgetRemaining int64
}

var _ dbworker.Handler = &handler{}
var _ workerutil.WithPreDequeue = &handler{}
var _ workerutil.WithHooks = &handler{}

func (h *handler) Handle(ctx context.Context, tx dbworkerstore.Store, record workerutil.Record) error {
	_, err := h.handle(ctx, h.dbStore.With(tx), record.(store.Upload))
	return err
}

func (h *handler) PreDequeue(ctx context.Context) (bool, interface{}, error) {
	if !h.enableBudget {
		return true, nil, nil
	}

	budgetRemaining := atomic.LoadInt64(&h.budgetRemaining)
	if budgetRemaining <= 0 {
		return false, nil, nil
	}

	return true, []*sqlf.Query{sqlf.Sprintf("(upload_size IS NULL OR upload_size <= %s)", budgetRemaining)}, nil
}

func (h *handler) PreHandle(ctx context.Context, record workerutil.Record) {
	atomic.AddInt64(&h.budgetRemaining, -h.getSize(record))
}

func (h *handler) PostHandle(ctx context.Context, record workerutil.Record) {
	atomic.AddInt64(&h.budgetRemaining, +h.getSize(record))
}

func (h *handler) getSize(record workerutil.Record) int64 {
	if size := record.(store.Upload).UploadSize; size != nil {
		return *size
	}

	return 0
}

// handle converts a raw upload into a dump within the given transaction context. Returns true if the
// upload record was requeued and false otherwise.
func (h *handler) handle(ctx context.Context, dbStore DBStore, upload store.Upload) (requeued bool, err error) {
	if requeued, err := requeueIfCloning(ctx, dbStore, upload); err != nil || requeued {
		return requeued, err
	}

	getChildren := func(ctx context.Context, dirnames []string) (map[string][]string, error) {
		directoryChildren, err := h.gitserverClient.DirectoryChildren(ctx, upload.RepositoryID, upload.Commit, dirnames)
		if err != nil {
			return nil, errors.Wrap(err, "gitserverClient.DirectoryChildren")
		}
		return directoryChildren, nil
	}

	return false, withUploadData(ctx, h.uploadStore, upload.ID, func(r io.Reader) (err error) {
		groupedBundleData, err := correlation.Correlate(ctx, r, upload.ID, upload.Root, getChildren)
		if err != nil {
			return errors.Wrap(err, "correlation.Correlate")
		}

		if err := writeData(ctx, h.lsifStore, upload.ID, groupedBundleData); err != nil {
			return err
		}

		// Start a nested transaction. In the event that something after this point fails, we want to
		// update the upload record with an error message but do not want to alter any other data in
		// the database. Rolling back to this savepoint will allow us to discard any other changes
		// but still commit the transaction as a whole.

		// with Postgres savepoints. In the event that something after this point fails, we want to
		// update the upload record with an error message but do not want to alter any other data in
		// the database. Rolling back to this savepoint will allow us to discard any other changes
		// but still commit the transaction as a whole.
		tx, err := dbStore.Transact(ctx)
		if err != nil {
			return errors.Wrap(err, "store.Transact")
		}
		defer func() { err = tx.Done(err) }()

		// Update package and package reference data to support cross-repo queries.
		if err := tx.UpdatePackages(ctx, groupedBundleData.Packages); err != nil {
			return errors.Wrap(err, "store.UpdatePackages")
		}
		if err := tx.UpdatePackageReferences(ctx, groupedBundleData.PackageReferences); err != nil {
			return errors.Wrap(err, "store.UpdatePackageReferences")
		}

		// Before we mark the upload as complete, we need to delete any existing completed uploads
		// that have the same repository_id, commit, root, and indexer values. Otherwise the transaction
		// will fail as these values form a unique constraint.
		if err := tx.DeleteOverlappingDumps(ctx, upload.RepositoryID, upload.Commit, upload.Root, upload.Indexer); err != nil {
			return errors.Wrap(err, "store.DeleteOverlappingDumps")
		}

		// Almost-success: we need to mark this upload as complete at this point as the next step changes
		// the visibility of the dumps for this repository. This requires that the new dump be available in
		// the lsif_dumps view, which requires a change of state. In the event of a future failure we can
		// still roll back to the save point and mark the upload as errored.
		if err := tx.MarkComplete(ctx, upload.ID); err != nil {
			return errors.Wrap(err, "store.MarkComplete")
		}

		// Mark this repository so that the commit updater process will pull the full commit graph from
		// gitserver and recalculate the nearest upload for each commit as well as which uploads are visible
		// from the tip of the default branch. We don't do this inside of the transaction as we re-calcalute
		// the entire set of data from scratch and we want to be able to coalesce requests for the same
		// repository rather than having a set of uploads for the same repo re-calculate nearly identical
		// data multiple times.
		if err := tx.MarkRepositoryAsDirty(ctx, upload.RepositoryID); err != nil {
			return errors.Wrap(err, "store.MarkRepositoryDirty")
		}

		return nil
	})
}

// CloneInProgressDelay is the delay between processing attempts when a repo is currently being cloned.
const CloneInProgressDelay = time.Minute

// requeueIfCloning ensures that the repo and revision are resolvable. If the repo does not exist, or
// if the repo has finished cloning and the revision does not exist, then the upload will fail to process.
// If the repo is currently cloning, then we'll requeue the upload to be tried again later. This will not
// increase the reset count of the record (so this doesn't count against the upload as a legitimate attempt).
func requeueIfCloning(ctx context.Context, dbStore DBStore, upload store.Upload) (requeued bool, _ error) {
	repo, err := backend.Repos.Get(ctx, api.RepoID(upload.RepositoryID))
	if err != nil {
		return false, errors.Wrap(err, "Repos.Get")
	}

	if _, err := backend.Repos.ResolveRev(ctx, repo, upload.Commit); err != nil {
		if !vcs.IsCloneInProgress(err) {
			return false, errors.Wrap(err, "Repos.ResolveRev")
		}

		if err := dbStore.Requeue(ctx, upload.ID, time.Now().UTC().Add(CloneInProgressDelay)); err != nil {
			return false, errors.Wrap(err, "store.Requeue")
		}

		return true, nil
	}

	return false, nil
}

// withUploadData will invoke the given function with a reader of the upload's raw data. The
// consumer should expect raw newline-delimited JSON content. If the function returns without
// an error, the upload file will be deleted.
func withUploadData(ctx context.Context, uploadStore uploadstore.Store, id int, fn func(r io.Reader) error) error {
	uploadFilename := fmt.Sprintf("upload-%d.lsif.gz", id)

	// Pull raw uploaded data from bucket
	rc, err := uploadStore.Get(ctx, uploadFilename)
	if err != nil {
		return errors.Wrap(err, "uploadStore.Get")
	}
	defer rc.Close()

	rc, err = gzip.NewReader(rc)
	if err != nil {
		return errors.Wrap(err, "gzip.NewReader")
	}
	defer rc.Close()

	if err := fn(rc); err != nil {
		return err
	}

	if err := uploadStore.Delete(ctx, uploadFilename); err != nil {
		log15.Warn("Failed to delete upload file", "err", err, "filename", uploadFilename)
	}

	return nil
}

// writeData transactionally writes the given grouped bundle data into the given LSIF store.
func writeData(ctx context.Context, lsifStore LSIFStore, id int, groupedBundleData *correlation.GroupedBundleData) (err error) {
	tx, err := lsifStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.WriteMeta(ctx, id, groupedBundleData.Meta); err != nil {
		return errors.Wrap(err, "store.WriteMeta")
	}
	if err := tx.WriteDocuments(ctx, id, groupedBundleData.Documents); err != nil {
		return errors.Wrap(err, "store.WriteDocuments")
	}
	if err := tx.WriteResultChunks(ctx, id, groupedBundleData.ResultChunks); err != nil {
		return errors.Wrap(err, "store.WriteResultChunks")
	}
	if err := tx.WriteDefinitions(ctx, id, groupedBundleData.Definitions); err != nil {
		return errors.Wrap(err, "store.WriteDefinitions")
	}
	if err := tx.WriteReferences(ctx, id, groupedBundleData.References); err != nil {
		return errors.Wrap(err, "store.WriteReferences")

	return nil
}

func patchData(ctx context.Context, base persistence.Store, patch *correlation.GroupedBundleData, reindexedFiles []string, fileStatus map[string]gitserver.Status) (patched *correlation.GroupedBundleData, err error) {
	log15.Warn("loading patch data...")

	reindexed := make(map[string]struct{})
	for _, file := range reindexedFiles {
		reindexed[file] = struct{}{}
	}

	patchDocs := make(map[string]types.DocumentData)
	for keyedDocument := range patch.Documents {
		patchDocs[keyedDocument.Path] = keyedDocument.Document
	}

	patchChunks := make(map[int]types.ResultChunkData)
	for indexedChunk := range patch.ResultChunks {
		patchChunks[indexedChunk.Index] = indexedChunk.ResultChunk
	}

	basePathList, err := base.PathsWithPrefix(ctx, "")
	baseMeta, err := base.ReadMeta(ctx)

	log15.Warn("loading base documents...")
	baseDocs := make(map[string]types.DocumentData)
	for _, path := range basePathList {
		document, _, _ := base.ReadDocument(ctx, path)
		baseDocs[path] = document
	}

	log15.Warn("loading base result chunks...")
	baseChunks := make(map[int]types.ResultChunkData)
	for id := 0; id < baseMeta.NumResultChunks; id++ {
		resultChunk, _, _ := base.ReadResultChunk(ctx, id)
		baseChunks[id] = resultChunk
	}

	modifiedOrDeletedPaths := make(map[string]struct{})
	for path, status := range fileStatus {
		if status == gitserver.Modified || status == gitserver.Deleted {
			modifiedOrDeletedPaths[path] = struct{}{}
		}
	}
	removeRefsIn(modifiedOrDeletedPaths, baseMeta, baseDocs, baseChunks)

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
	unifyRangeIDs(baseDocs, patch.Meta, patchDocs, patchChunks, fileStatus)

	log15.Warn("indexing new data...")
	defResultsByPath := make(map[string]map[types.ID]types.RangeData)

	for path := range pathsToCopy {
		log15.Warn(fmt.Sprintf("finding all def results referenced in %v", path))
		for _, rng := range patchDocs[path].Ranges {
			if rng.DefinitionResultID == "" {
				continue
			}
			defs, defChunk := getDefRef(rng.DefinitionResultID, patch.Meta, patchChunks)
			for _, defLoc := range defs {
				defPath := defChunk.DocumentPaths[defLoc.DocumentID]
				def := patchDocs[defPath].Ranges[defLoc.RangeID]
				defResults, exists := defResultsByPath[defPath]
				if !exists {
					defResults = make(map[types.ID]types.RangeData)
					defResultsByPath[defPath] = defResults
				}
				if _, exists := defResults[defLoc.RangeID]; !exists {
					defResults[defLoc.RangeID] = def
				}
			}
		}
	}

	log15.Warn("merging data...")
	for path, defsMap := range defResultsByPath {
		baseDoc := baseDocs[path]
		doLog := path == "cmd/frontend/internal/app/updatecheck/handler.go"
		defIdxs := sortedRangeIDs(defsMap)
		for _, defRngID := range defIdxs {
			def := defsMap[defRngID]
			if doLog {
				log15.Warn(fmt.Sprintf("unifying def result defined in %v:%v:%v)", def.StartLine, def.StartCharacter, path))
			}
			var defID, refID types.ID
			if fileStatus[path] == gitserver.Unchanged {
				baseRng := baseDoc.Ranges[defRngID]

				defID = baseRng.DefinitionResultID
				refID = baseRng.ReferenceResultID
				if doLog {
					log15.Warn(fmt.Sprintf("unifying with existing result IDs %v, %v", defID, refID))
				}
			} else {
				defID, err = newID()
				if err != nil {
					return nil, err
				}
				refID, err = newID()
				if err != nil {
					return nil, err
				}
				if doLog {
					log15.Warn(fmt.Sprintf("using new result IDs %v, %v", defID, refID))
				}
			}

			patchRefs, patchRefChunk := getDefRef(def.ReferenceResultID, patch.Meta, patchChunks)

			patchDefs, patchDefChunk := getDefRef(def.DefinitionResultID, patch.Meta, patchChunks)
			baseRefs, baseRefChunk := getDefRef(refID, baseMeta, baseChunks)
			baseDefs, baseDefChunk := getDefRef(defID, baseMeta, baseChunks)

			baseRefDocumentIDs := make(map[string]types.ID)
			for id, path := range baseRefChunk.DocumentPaths {
				baseRefDocumentIDs[path] = id
			}
			baseDefDocumentIDs := make(map[string]types.ID)
			for id, path := range baseDefChunk.DocumentPaths {
				baseDefDocumentIDs[path] = id
			}
			for _, patchRef := range patchRefs {
				patchPath := patchRefChunk.DocumentPaths[patchRef.DocumentID]
				patchRng := patchDocs[patchPath].Ranges[patchRef.RangeID]
				if doLog {
					log15.Warn(fmt.Sprintf("processing ref %v:%v:%v", patchPath, patchRng.StartLine, patchRng.StartCharacter))
				}
				if fileStatus[patchPath] != gitserver.Unchanged {
					if doLog {
						log15.Warn(fmt.Sprintf("adding ref"))
					}
					baseRefDocumentID, exists := baseRefDocumentIDs[path]
					if !exists {
						baseRefDocumentID, err = newID()
						if err != nil {
							return nil, err
						}
						baseRefDocumentIDs[path] = baseRefDocumentID
						baseRefChunk.DocumentPaths[baseRefDocumentID] = path
					}
					patchRef.DocumentID = baseRefDocumentID
					baseRefs = append(baseRefs, patchRef)

				}

				if len(baseDefs) == 0 {
					var patchDef *types.DocumentIDRangeID
					for _, tmpDef := range patchDefs {
						patchDefPath := patchDefChunk.DocumentPaths[tmpDef.DocumentID]
						if patchDefPath == patchPath && tmpDef.RangeID == patchRef.RangeID {
							patchDef = &tmpDef
						}
					}
					if patchDef != nil {
						if doLog {
							log15.Warn(fmt.Sprintf("adding def"))
						}
						baseDefDocumentID, exists := baseDefDocumentIDs[path]
						if !exists {
							baseDefDocumentID, err = newID()
							if err != nil {
								return nil, err
							}
							baseDefDocumentIDs[path] = baseDefDocumentID
							baseDefChunk.DocumentPaths[baseDefDocumentID] = path
						}
						patchDef.DocumentID = baseDefDocumentID
						baseDefs = append(baseDefs, *patchDef)
					}
				}

				if _, exists := pathsToCopy[patchPath]; exists {
					rng := patchDocs[patchPath].Ranges[patchRef.RangeID]
					if doLog {
						log15.Warn(fmt.Sprintf("updating result ID"))
					}
					patchDocs[patchPath].Ranges[patchRef.RangeID] = types.RangeData{
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

			if doLog {
				log15.Warn("")
			}
		}
	}

	for path, status := range fileStatus {
		if status == gitserver.Deleted {
			log15.Warn(fmt.Sprintf("deleting path %v", path))
			delete(baseDocs, path)
		}
	}
	for path := range pathsToCopy {
		log15.Warn(fmt.Sprintf("copying document %v", path))
		baseDocs[path] = patchDocs[path]
	}

	log15.Warn("writing data...")
	documentChan := make(chan persistence.KeyedDocumentData, len(baseDocs))
	go func() {
		defer close(documentChan)
		for path, doc := range baseDocs {
			select {
			case documentChan <- persistence.KeyedDocumentData{
				Path:     path,
				Document: doc,
			}:
			case <-ctx.Done():
				return
			}
		}
	}()
	resultChunkChan := make(chan persistence.IndexedResultChunkData, len(baseChunks))
	go func() {
		defer close(resultChunkChan)

		for idx, chunk := range baseChunks {
			select {
			case resultChunkChan <- persistence.IndexedResultChunkData{
				Index:       idx,
				ResultChunk: chunk,
			}:
			case <-ctx.Done():
				return
			}
		}
	}()

	patched = &correlation.GroupedBundleData{
		Meta:              baseMeta,
		Documents:         documentChan,
		ResultChunks:      resultChunkChan,
		Definitions:       patch.Definitions,
		References:        patch.References,
		Packages:          patch.Packages,
		PackageReferences: patch.PackageReferences,
	}

	log15.Warn("done...")
	return
}

func removeRefsIn(paths map[string]struct{}, meta types.MetaData, docs map[string]types.DocumentData, chunks map[int]types.ResultChunkData) {
	deletedRefs := make(map[types.ID]struct{})

	for path := range paths {
		doc := docs[path]
		for _, rng := range doc.Ranges {
			if _, exists := deletedRefs[rng.ReferenceResultID]; exists {
				continue
			}

			refs, refChunk := getDefRef(rng.ReferenceResultID, meta, chunks)
			var filteredRefs []types.DocumentIDRangeID
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

func unifyRangeIDs(updateToDocs map[string]types.DocumentData, toUpdateMeta types.MetaData, toUpdateDocs map[string]types.DocumentData, toUpdateChunks map[int]types.ResultChunkData, fileStatus map[string]gitserver.Status) error {
	updatedRngIDs := make(map[types.ID]types.ID)
	resultsToUpdate := make(map[types.ID]struct{})

	for path, toUpdateDoc := range toUpdateDocs {
		pathUpdatedRngIDs := make(map[types.ID]types.ID)
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

				if util.CompareRanges(updateToRng, toUpdateRng) != 0 {
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
		var updated []types.DocumentIDRangeID
		for _, result := range results {
			if updatedID, exists := updatedRngIDs[result.RangeID]; exists {
				updated = append(updated, types.DocumentIDRangeID{
					RangeID: updatedID,
					DocumentID: result.DocumentID,
				})
			} else {
				updated = append(updated, types.DocumentIDRangeID{
					RangeID: result.RangeID,
					DocumentID: result.DocumentID,
				})
			}
		}
		chunk.DocumentIDRangeIDs[resultID] = updated
	}

	return nil
}

func sortedRangeIDs(ranges map[types.ID]types.RangeData) []types.ID {
	var rngIDs []types.ID
	for rngID := range ranges {
		rngIDs = append(rngIDs, rngID)
	}

	sort.Slice(rngIDs, func(i, j int) bool {
		return util.CompareRanges(ranges[rngIDs[i]], ranges[rngIDs[j]]) < 0
	})

	return rngIDs
}

func getDefRef(resultID types.ID, meta types.MetaData, resultChunks map[int]types.ResultChunkData) ([]types.DocumentIDRangeID, types.ResultChunkData) {
	chunkID := types.HashKey(resultID, meta.NumResultChunks)
	chunk := resultChunks[chunkID]
	docRngIDs := chunk.DocumentIDRangeIDs[resultID]
	return docRngIDs, chunk
}

func newID() (types.ID, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return types.ID(uuid.String()), nil
}

func (h *handler) sendDB(ctx context.Context, uploadID int, tempDir string) (err error) {
	ctx, endOperation := h.metrics.SendDBOperation.With(ctx, &err, observation.Args{})
	defer endOperation(1, observation.Args{})

	if err := h.bundleManagerClient.SendDB(ctx, uploadID, tempDir); err != nil {
		return errors.Wrap(err, "bundleManager.SendDB")
	}

	return nil
}
