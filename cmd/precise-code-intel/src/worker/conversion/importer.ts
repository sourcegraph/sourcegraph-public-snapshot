import * as sqliteModels from '../../shared/models/sqlite'
import * as lsif from 'lsif-protocol'
import { Correlator, ResultSetData, ResultSetId } from './correlator'
import { createSqliteConnection } from '../../shared/database/sqlite'
import { databaseInsertionDurationHistogram, databaseInsertionErrorsCounter } from '../metrics'
import { DefaultMap } from '../../shared/datastructures/default-map'
import { EntityManager } from 'typeorm'
import { gzipJSON } from '../../shared/encoding/json'
import { hashKey } from '../../shared/models/hash'
import { isEqual, uniqWith } from 'lodash'
import { logAndTraceCall, TracingContext } from '../../shared/tracing'
import { mustGet } from '../../shared/maps'
import { Package, SymbolReferences } from '../../shared/store/dependencies'
import { readGzippedJsonElementsFromFile } from '../../shared/input'
import { TableInserter } from '../../shared/database/inserter'
import { createSilentLogger } from '../../shared/logging'
import { PathExistenceChecker } from './existence'
import * as settings from '../settings'

/** The insertion metrics for the database. */
const inserterMetrics = {
    durationHistogram: databaseInsertionDurationHistogram,
    errorsCounter: databaseInsertionErrorsCounter,
}

/**
 * The internal version of our SQLite databases. We need to keep this in case
 * we add something that can't be done transparently; if we change how we process
 * something in the future we'll need to consider a number of previous version
 * while we update or re-process the already-uploaded data.
 */
const INTERNAL_LSIF_VERSION = '0.1.0'

/**
 * Populate a SQLite database with the given input stream. Returns the
 * data required to populate the dependency tables in Postgres.
 *
 * @param args Parameter bag.
 */
export async function convertLsif({
    path,
    root,
    database,
    pathExistenceChecker,
    ctx: { logger = createSilentLogger(), span } = {},
}: {
    /** The filepath containing a gzipped compressed stream of JSON lines composing the LSIF dump. */
    path: string
    /** The root of all files that are in the dump. */
    root: string
    /** The filepath of the database to populate. */
    database: string
    /** An object that tracks whether a path is visible within the LSIF dump. */
    pathExistenceChecker: PathExistenceChecker
    /** The tracing context. */
    ctx?: TracingContext
}): Promise<{ packages: Package[]; references: SymbolReferences[] }> {
    const connection = await createSqliteConnection(database, sqliteModels.entities, logger)

    try {
        await connection.query('PRAGMA synchronous = OFF')
        await connection.query('PRAGMA journal_mode = OFF')

        return await connection.transaction(entityManager =>
            importLsif(entityManager, path, root, pathExistenceChecker, { logger, span })
        )
    } finally {
        await connection.close()
    }
}

/**
 * Correlate each vertex and edge together, then populate the provided entity manager
 * with the document, definition, and reference information. Returns the package and
 * external reference data needed to populate the dependency tables in Postgres.
 *
 * @param entityManager A transactional SQLite entity manager.
 * @param path The filepath containing a gzipped compressed stream of JSON lines composing the LSIF dump.
 * @param root The root of all files that are in the dump.
 * @param pathExistenceChecker An object that tracks whether a path is visible within the LSIF dump.
 * @param ctx The tracing context.
 */
export async function importLsif(
    entityManager: EntityManager,
    path: string,
    root: string,
    pathExistenceChecker: PathExistenceChecker,
    ctx: TracingContext
): Promise<{ packages: Package[]; references: SymbolReferences[] }> {
    // Correlate input data into in-memory maps
    const correlator = new Correlator(root, ctx.logger)
    await logAndTraceCall(ctx, 'Correlating LSIF data', async () => {
        for await (const element of readGzippedJsonElementsFromFile(path) as AsyncIterable<lsif.Vertex | lsif.Edge>) {
            correlator.insert(element)
        }
    })

    if (correlator.lsifVersion === undefined) {
        throw new Error('No metadata defined.')
    }

    // Determine if multiple documents are defined with the same URI. This happens in
    // some indexers (such as lsif-tsc) that index dependent projects into the same
    // dump as the target project. For each set of documents that share a path, we
    // choose one document to be the canonical representative and merge the contains,
    // definition, and reference data into the unique canonical document.
    await logAndTraceCall(ctx, 'Merging documents', () => mergeDocuments(correlator))

    // Determine which reference results are linked together. Determine a canonical
    // reference result for each set so that we can remap all identifiers to the
    // chosen one.
    const canonicalReferenceResultIds = await logAndTraceCall(ctx, 'Canonicalizing reference results', () =>
        canonicalizeReferenceResults(correlator)
    )

    // Make all necessary visibility queries to gitserver here. This allows us to batch
    // the requests to reduce the number of network roundtrips, and also allows us to
    // time the total cost of fetching this data from gitserver by doing it all in one
    // place. If we perform the queries lazily, we would need to add the timings for
    // each individual span in the resulting trace.
    await pathExistenceChecker.warmCache(Array.from(correlator.documentPaths.values()))

    // Calculate the number of result chunks that we'll attempt to populate
    const numResults = correlator.definitionData.size + correlator.referenceData.size
    const numResultChunks = Math.min(
        settings.MAX_NUM_RESULT_CHUNKS,
        Math.floor(numResults / settings.RESULTS_PER_RESULT_CHUNK) || 1
    )

    // Insert metadata
    const metaInserter = new TableInserter(
        entityManager,
        sqliteModels.MetaModel,
        sqliteModels.MetaModel.BatchSize,
        inserterMetrics
    )
    await populateMetadataTable(correlator, metaInserter, numResultChunks)
    await metaInserter.flush()

    // Insert documents
    await logAndTraceCall(ctx, 'Populating documents', async () => {
        const documentInserter = new TableInserter(
            entityManager,
            sqliteModels.DocumentModel,
            sqliteModels.DocumentModel.BatchSize,
            inserterMetrics
        )
        await populateDocumentsTable(correlator, documentInserter, canonicalReferenceResultIds, pathExistenceChecker)
        await documentInserter.flush()
    })

    // Insert result chunks
    await logAndTraceCall(ctx, 'Populating result chunks', async () => {
        const resultChunkInserter = new TableInserter(
            entityManager,
            sqliteModels.ResultChunkModel,
            sqliteModels.ResultChunkModel.BatchSize,
            inserterMetrics
        )
        await populateResultChunksTable(correlator, resultChunkInserter, numResultChunks, pathExistenceChecker)
        await resultChunkInserter.flush()
    })

    // Insert definitions and references
    await logAndTraceCall(ctx, 'Populating definitions and references', async () => {
        const definitionInserter = new TableInserter(
            entityManager,
            sqliteModels.DefinitionModel,
            sqliteModels.DefinitionModel.BatchSize,
            inserterMetrics
        )
        const referenceInserter = new TableInserter(
            entityManager,
            sqliteModels.ReferenceModel,
            sqliteModels.ReferenceModel.BatchSize,
            inserterMetrics
        )
        await populateDefinitionsAndReferencesTables(
            correlator,
            definitionInserter,
            referenceInserter,
            pathExistenceChecker
        )
        await definitionInserter.flush()
        await referenceInserter.flush()
    })

    // Return data to populate dependency tables in Postgres
    return { packages: getPackages(correlator), references: getReferences(correlator) }
}

/**
 * Correlate, encode, and insert all document entries for this dump.
 *
 * @param correlator The correlator with all vertices and edges inserted.
 * @param documentInserter The inserter for the documents table.
 * @param canonicalReferenceResultIds A map from reference result identifiers to its canonical identifier.
 * @param pathExistenceChecker An object that tracks whether a path is visible within the LSIF dump.
 */
async function populateDocumentsTable(
    correlator: Correlator,
    documentInserter: TableInserter<sqliteModels.DocumentModel, new () => sqliteModels.DocumentModel>,
    canonicalReferenceResultIds: Map<sqliteModels.ReferenceResultId, sqliteModels.ReferenceResultId>,
    pathExistenceChecker: PathExistenceChecker
): Promise<void> {
    // Collapse result sets data into the ranges that can reach them. The
    // remainder of this function assumes that we can completely ignore
    // the "next" edges coming from range data.
    for (const [rangeId, range] of correlator.rangeData) {
        canonicalizeItem(correlator, canonicalReferenceResultIds, rangeId, range)
    }

    // Gather and insert document data that includes the ranges contained in the document,
    // any associated hover data, and any associated moniker data/package information.
    // Each range also has identifiers that correlate to a definition or reference result
    // which can be found in a result chunk, created in the next step.

    for (const [documentId, documentPath] of correlator.documentPaths) {
        // Do not gather any document that is not within the dump root or does not exist
        // in git. If the path is outside of the dump root, then it will never be queried
        // as the current text document path and the dump root are compared to determine
        // which dump to open. If the path does not exist in git, it will also never be
        // queried.
        if (!pathExistenceChecker.shouldIncludePath(documentPath)) {
            continue
        }

        // Create document record from the correlated information. This will also insert
        // external definitions and references into the maps initialized above, which are
        // inserted into the definitions and references table, respectively, below.
        const document = gatherDocument(correlator, documentId, documentPath)

        // Encode and insert document record
        await documentInserter.insert({
            path: documentPath,
            data: await gzipJSON({
                ranges: document.ranges,
                hoverResults: document.hoverResults,
                monikers: document.monikers,
                packageInformation: document.packageInformation,
            }),
        })
    }
}

/**
 * Correlate and insert all result chunk entries for this dump.
 *
 * @param correlator The correlator with all vertices and edges inserted.
 * @param resultChunkInserter The inserter for the result chunks table.
 * @param numResultChunks The number of result chunks used to hash compute the result identifier hash.
 * @param pathExistenceChecker An object that tracks whether a path is visible within the LSIF dump.
 */
async function populateResultChunksTable(
    correlator: Correlator,
    resultChunkInserter: TableInserter<sqliteModels.ResultChunkModel, new () => sqliteModels.ResultChunkModel>,
    numResultChunks: number,
    pathExistenceChecker: PathExistenceChecker
): Promise<void> {
    // Create all the result chunks we'll be populating and inserting up-front. Data will
    // be inserted into result chunks based on hash values (modulo the number of result chunks),
    // and we don't want to create them lazily.

    const resultChunks = new Array(numResultChunks).fill(null).map(() => ({
        paths: new Map<sqliteModels.DocumentId, string>(),
        documentIdRangeIds: new Map<sqliteModels.DefinitionReferenceResultId, sqliteModels.DocumentIdRangeId[]>(),
    }))

    const chunkResults = (
        data: Map<sqliteModels.DefinitionReferenceResultId, Map<sqliteModels.DocumentId, lsif.RangeId[]>>
    ): void => {
        for (const [id, documentRanges] of data) {
            // Flatten map into list of ranges
            let flattenedRangeList: (sqliteModels.DocumentIdRangeId & { documentPath: string })[] = []
            for (const [documentId, rangeIds] of documentRanges) {
                const documentPath = mustGet(correlator.documentPaths, documentId, 'documentPath')

                // Skip pointing to locations that are not available in git. This can occur
                // with indexers that point to generated files or dependencies that are not
                // committed (e.g. node_modules). Keeping these in the dump can cause the
                // UI to redirect to a path that doesn't exist.
                if (!pathExistenceChecker.shouldIncludePath(documentPath, false)) {
                    continue
                }

                flattenedRangeList = flattenedRangeList.concat(
                    rangeIds.map(rangeId => ({ documentId, documentPath, rangeId }))
                )
            }

            // Insert ranges into target result chunk
            const resultChunk = resultChunks[hashKey(id, resultChunks.length)]
            const documentIdRangeIds = flattenedRangeList.map(({ documentId, rangeId }) => ({ documentId, rangeId }))
            resultChunk.documentIdRangeIds.set(id, documentIdRangeIds)

            for (const { documentId } of documentIdRangeIds) {
                // Add paths into the result chunk where they are used
                resultChunk.paths.set(documentId, mustGet(correlator.documentPaths, documentId, 'documentPath'))
            }
        }
    }

    // Add definitions and references to result chunks
    chunkResults(correlator.definitionData)
    chunkResults(correlator.referenceData)

    for (const [id, resultChunk] of resultChunks.entries()) {
        // Empty chunk, no need to serialize as it will never be queried
        if (resultChunk.paths.size === 0 && resultChunk.documentIdRangeIds.size === 0) {
            continue
        }

        const data = await gzipJSON({
            documentPaths: resultChunk.paths,
            documentIdRangeIds: resultChunk.documentIdRangeIds,
        })

        // Encode and insert result chunk record
        await resultChunkInserter.insert({ id, data })
    }
}

/**
 * Correlate and insert all definition and reference entries for this dump.
 *
 * @param correlator The correlator with all vertices and edges inserted.
 * @param definitionInserter The inserter for the definitions table.
 * @param referenceInserter The inserter for the references table.
 * @param pathExistenceChecker An object that tracks whether a path is visible within the LSIF dump.
 */
async function populateDefinitionsAndReferencesTables(
    correlator: Correlator,
    definitionInserter: TableInserter<sqliteModels.DefinitionModel, new () => sqliteModels.DefinitionModel>,
    referenceInserter: TableInserter<sqliteModels.ReferenceModel, new () => sqliteModels.ReferenceModel>,
    pathExistenceChecker: PathExistenceChecker
): Promise<void> {
    // Determine the set of monikers that are attached to a definition or a
    // reference result. Correlating information in this way has two benefits:
    //   (1) it reduces duplicates in the definitions and references tables
    //   (2) it stop us from re-iterating over the range data of the entire
    //       LSIF dump, which is by far the largest proportion of data.

    const definitionMonikers = new DefaultMap<sqliteModels.DefinitionResultId, Set<sqliteModels.MonikerId>>(
        () => new Set()
    )
    const referenceMonikers = new DefaultMap<sqliteModels.ReferenceResultId, Set<sqliteModels.MonikerId>>(
        () => new Set()
    )

    for (const range of correlator.rangeData.values()) {
        if (range.monikerIds.size === 0) {
            continue
        }

        if (range.definitionResultId !== undefined) {
            const set = definitionMonikers.getOrDefault(range.definitionResultId)
            for (const monikerId of range.monikerIds) {
                set.add(monikerId)
            }
        }

        if (range.referenceResultId !== undefined) {
            const set = referenceMonikers.getOrDefault(range.referenceResultId)
            for (const monikerId of range.monikerIds) {
                set.add(monikerId)
            }
        }
    }

    const insertMonikerRanges = async (
        data: Map<sqliteModels.DefinitionReferenceResultId, Map<sqliteModels.DocumentId, lsif.RangeId[]>>,
        monikers: Map<sqliteModels.DefinitionReferenceResultId, Set<lsif.RangeId>>,
        inserter: TableInserter<
            sqliteModels.DefinitionModel | sqliteModels.ReferenceModel,
            new () => sqliteModels.DefinitionModel | sqliteModels.ReferenceModel
        >
    ): Promise<void> => {
        for (const [id, documentRanges] of data) {
            // Get monikers. Nothing to insert if we don't have any.
            const monikerIds = monikers.get(id)
            if (monikerIds === undefined) {
                continue
            }

            // Correlate each moniker with the document/range pairs stored in
            // the result set provided by the data argument of this function.

            for (const monikerId of monikerIds) {
                const moniker = mustGet(correlator.monikerData, monikerId, 'moniker')

                for (const [documentId, rangeIds] of documentRanges) {
                    const documentPath = mustGet(correlator.documentPaths, documentId, 'documentPath')

                    // Skip definitions or references that point to a document that are not
                    // present in the dump. Including this would cause a query that always
                    // fails when it cannot resolve the missing document data.
                    if (!pathExistenceChecker.shouldIncludePath(documentPath)) {
                        continue
                    }

                    for (const rangeId of rangeIds) {
                        const range = mustGet(correlator.rangeData, rangeId, 'range')

                        await inserter.insert({
                            scheme: moniker.scheme,
                            identifier: moniker.identifier,
                            documentPath,
                            ...range,
                        })
                    }
                }
            }
        }
    }

    // Insert definitions and references records
    await insertMonikerRanges(correlator.definitionData, definitionMonikers, definitionInserter)
    await insertMonikerRanges(correlator.referenceData, referenceMonikers, referenceInserter)
}

/**
 * Insert metadata row. This gives us a place to store the version of the converter that
 * created a database in case we have backwards-incompatible changes in the future that
 * require historic version flagging. This also stores the number of result chunks
 * determined above so that we can have stable hashes at query time.
 *
 * @param correlator The correlator with all vertices and edges inserted.
 * @param metaInserter The inserter for the meta table.
 * @param numResultChunks The number of result chunks used to hash compute the result identifier hash.
 */
async function populateMetadataTable(
    correlator: Correlator,
    metaInserter: TableInserter<sqliteModels.MetaModel, new () => sqliteModels.MetaModel>,
    numResultChunks: number
): Promise<void> {
    await metaInserter.insert({
        id: 1,
        lsifVersion: correlator.lsifVersion,
        sourcegraphVersion: INTERNAL_LSIF_VERSION,
        numResultChunks,
    })
}

/**
 * Gather all package information that is referenced by an exported
 * moniker. These will be the packages that are provided by the repository
 * represented by this LSIF dump.
 *
 * @param correlator The correlator with all vertices and edges inserted.
 */
function getPackages(correlator: Correlator): Package[] {
    const packages: Package[] = []
    for (const id of correlator.exportedMonikers) {
        const source = mustGet(correlator.monikerData, id, 'moniker')
        const packageInformationId = assertId(source.packageInformationId)
        const packageInfo = mustGet(correlator.packageInformationData, packageInformationId, 'packageInformation')
        packages.push({
            scheme: source.scheme,
            name: packageInfo.name,
            version: packageInfo.version,
        })
    }

    return uniqWith(packages, isEqual)
}

/**
 * Gather all imported moniker identifiers along with their package
 * information. These will be the packages that are a dependency of the
 * repository represented by this LSIF dump.
 *
 * @param correlator The correlator with all vertices and edges inserted.
 */
function getReferences(correlator: Correlator): SymbolReferences[] {
    const packageIdentifiers: Map<string, string[]> = new Map()
    for (const id of correlator.importedMonikers) {
        const source = mustGet(correlator.monikerData, id, 'moniker')
        const packageInformationId = assertId(source.packageInformationId)
        const packageInfo = mustGet(correlator.packageInformationData, packageInformationId, 'packageInformation')
        const pkg = JSON.stringify({
            scheme: source.scheme,
            name: packageInfo.name,
            version: packageInfo.version,
        })

        const list = packageIdentifiers.get(pkg)
        if (list) {
            list.push(source.identifier)
        } else {
            packageIdentifiers.set(pkg, [source.identifier])
        }
    }

    return Array.from(packageIdentifiers).map(([key, identifiers]) => ({
        package: JSON.parse(key) as Package,
        identifiers,
    }))
}

/**
 * Merge the data in the correlator of all documents that share the same path. This
 * function works by moving the contains, definition, and reference data keyed by a
 * document with a duplicate path into a canonical document with that path. The first
 * document inserted for a path is the canonical document for that path. This function
 * guarantees that duplicate document ids are removed from these maps.
 *
 * @param correlator The correlator with all vertices and edges inserted.
 */
function mergeDocuments(correlator: Correlator): void {
    const uriMap = new Map<string, sqliteModels.DocumentId>()
    for (const [id, path] of correlator.documentPaths.entries()) {
        const canonicalId = uriMap.get(path)
        if (canonicalId === undefined) {
            uriMap.set(path, id)
            continue
        }

        mergeContains(id, canonicalId, correlator.containsData)
        mergeDefinitionReferences(id, canonicalId, correlator.definitionData)
        mergeDefinitionReferences(id, canonicalId, correlator.referenceData)

        // Discard the document data as a flag to prevent inserting one
        // of the documents subsumed by the canonical representative.
        correlator.documentPaths.delete(id)
    }
}

/**
 * Move the contains data for document `id` into the contains data of document
 * `canonicalId`, then delete the reference to document `id.`
 *
 * @param id The id of the document to replace.
 * @param canonicalId The id of the document to subsume the other.
 * @param containsData The containment map of the correlator.
 */
function mergeContains(
    id: sqliteModels.DocumentId,
    canonicalId: sqliteModels.DocumentId,
    containsData: Map<sqliteModels.DocumentId, Set<lsif.RangeId>>
): void {
    const contains = mustGet(containsData, id, 'contains')
    const canonicalContains = mustGet(containsData, canonicalId, 'contains')
    for (const rangeId of contains) {
        canonicalContains.add(rangeId)
    }

    // Do not keep refs to document id we're removing
    containsData.delete(id)
}

/**
 * Move the definition or reference data for document `id` into the definition or
 * reference data of document `canonicalId`, then delete the reference to document
 * `id.`
 *
 * @param id The id of the document to replace.
 * @param canonicalId The id of the document to subsume the other.
 * @param map The definitions or references map of the correlator.
 */
function mergeDefinitionReferences<K>(
    id: sqliteModels.DocumentId,
    canonicalId: sqliteModels.DocumentId,
    map: Map<K, DefaultMap<sqliteModels.DocumentId, lsif.RangeId[]>>
): void {
    for (const value of map.values()) {
        const data = value.get(id)
        if (data !== undefined) {
            for (const rangeId of data || []) {
                value.getOrDefault(canonicalId).push(rangeId)
            }

            // Do not keep refs to document id we're removing
            value.delete(id)
        }
    }
}

/**
 * Determine which reference result sets are linked via item edges. Choose a canonical
 * reference result from each batch. Merge all data into the canonical result and remove
 * all non-canonical results from the correlator (note: this leave unlinked results alone).
 * Return a map from reference result identifier to the identifier of the canonical result.
 *
 * @param correlator The correlator with all vertices and edges inserted.
 */
function canonicalizeReferenceResults(
    correlator: Correlator
): Map<sqliteModels.ReferenceResultId, sqliteModels.ReferenceResultId> {
    const canonicalReferenceResultIds = new Map<sqliteModels.ReferenceResultId, sqliteModels.ReferenceResultId>()

    for (const referenceResultId of correlator.linkedReferenceResults.keys()) {
        // Don't re-process the same set of linked reference results
        if (canonicalReferenceResultIds.has(referenceResultId)) {
            continue
        }

        // Find all reachable items and order them deterministically
        const linkedIds = Array.from(correlator.linkedReferenceResults.extractSet(referenceResultId))
        linkedIds.sort()

        // Choose arbitrary canonical id
        const canonicalId = linkedIds[0]
        const canonicalReferenceResult = mustGet(correlator.referenceData, canonicalId, 'referenceResult')

        for (const linkedId of linkedIds) {
            // Link each id to its canonical representation. We do this for
            // the `linkedId === canonicalId` case so we can reliably detect
            // duplication at the start of this loop.

            canonicalReferenceResultIds.set(linkedId, canonicalId)

            if (linkedId !== canonicalId) {
                // If it's a different identifier, then normalize all data from the linked result
                // set into the canonical one.
                for (const [documentId, rangeIds] of mustGet(correlator.referenceData, linkedId, 'referenceResult')) {
                    canonicalReferenceResult.getOrDefault(documentId).push(...rangeIds)
                }
            }
        }
    }

    // Remove all non-canonical but linked result sets
    const keys = new Set(canonicalReferenceResultIds.keys())
    const vals = new Set(canonicalReferenceResultIds.values())
    for (const key of keys) {
        if (!vals.has(key)) {
            correlator.referenceData.delete(key)
        }
    }

    return canonicalReferenceResultIds
}
/**
 * Flatten the definition result, reference result, hover results, and monikers of range
 * and result set items by following next links in the graph. This needs to be run over
 * each range before committing them to a document.
 *
 * @param correlator The correlator with all vertices and edges inserted.
 * @param canonicalReferenceResultIds A map from reference result identifiers to its canonical identifier.
 * @param id The item identifier.
 * @param item The range or result set item.
 */
function canonicalizeItem(
    correlator: Correlator,
    canonicalReferenceResultIds: Map<sqliteModels.ReferenceResultId, sqliteModels.ReferenceResultId>,
    id: lsif.RangeId | ResultSetId,
    item: sqliteModels.RangeData | ResultSetData
): void {
    const monikers = new Set<sqliteModels.MonikerId>()
    if (item.monikerIds.size > 0) {
        // Find arbitrary moniker attached to item
        const candidateMoniker = item.monikerIds.keys().next().value

        // Get all monikers reachable from this one
        for (const monikerId of correlator.linkedMonikers.extractSet(candidateMoniker)) {
            if (mustGet(correlator.monikerData, monikerId, 'moniker').kind !== lsif.MonikerKind.local) {
                monikers.add(monikerId)
            }
        }
    }

    const nextId = correlator.nextData.get(id)
    if (nextId !== undefined) {
        // If we have a next edge to a result set, get it and canonicalize it first. This
        // will recursively look at any result that that it can reach that hasn't yet been
        // canonicalized.

        const nextItem = mustGet(correlator.resultSetData, nextId, 'resultSet')
        canonicalizeItem(correlator, canonicalReferenceResultIds, nextId, nextItem)

        // Add each moniker of the next set to this item
        for (const monikerId of nextItem.monikerIds) {
            monikers.add(monikerId)
        }

        // If we do not have a definition, reference, or hover result, take the result
        // value from the next item.

        if (item.definitionResultId === undefined) {
            item.definitionResultId = nextItem.definitionResultId
        }

        if (item.referenceResultId === undefined) {
            item.referenceResultId = nextItem.referenceResultId
        }

        if (item.hoverResultId === undefined) {
            item.hoverResultId = nextItem.hoverResultId
        }
    }

    if (item.referenceResultId && canonicalReferenceResultIds.has(item.referenceResultId)) {
        // If there is a canonical version of this reference result, use that instead
        item.referenceResultId = canonicalReferenceResultIds.get(item.referenceResultId)
    }

    // Update our moniker sets (our normalized sets and any monikers of our next item)
    item.monikerIds = monikers

    // Remove the next edge so we don't traverse it a second time
    correlator.nextData.delete(id)
}

/**
 * Create a self-contained document object from the data in the given correlator. This
 * includes hover and moniker results, as well as identifiers to definition and reference
 * results (but not the actual ranges). See result chunk table for details.
 *
 * @param correlator The correlator with all vertices and edges inserted.
 * @param currentDocumentId The identifier of the document.
 * @param path The path of the document.
 */
function gatherDocument(
    correlator: Correlator,
    currentDocumentId: sqliteModels.DocumentId,
    path: string
): sqliteModels.DocumentData {
    const document = {
        path,
        ranges: new Map<lsif.RangeId, sqliteModels.RangeData>(),
        hoverResults: new Map<sqliteModels.HoverResultId, string>(),
        monikers: new Map<sqliteModels.MonikerId, sqliteModels.MonikerData>(),
        packageInformation: new Map<sqliteModels.PackageInformationId, sqliteModels.PackageInformationData>(),
    }

    const addHover = (id: sqliteModels.HoverResultId | undefined): void => {
        if (id === undefined || document.hoverResults.has(id)) {
            return
        }

        // Add hover result to the document, if defined and not a duplicate
        const data = mustGet(correlator.hoverData, id, 'hoverResult')
        document.hoverResults.set(id, data)
    }

    const addPackageInformation = (id: sqliteModels.PackageInformationId | undefined): void => {
        if (id === undefined || document.packageInformation.has(id)) {
            return
        }

        // Add package information to the document, if defined and not a duplicate
        const data = mustGet(correlator.packageInformationData, id, 'packageInformation')
        document.packageInformation.set(id, data)
    }

    const addMoniker = (id: sqliteModels.MonikerId | undefined): void => {
        if (id === undefined || document.monikers.has(id)) {
            return
        }

        // Add moniker to the document, if defined and not a duplicate
        const moniker = mustGet(correlator.monikerData, id, 'moniker')
        document.monikers.set(id, moniker)

        // Add related package information to document
        addPackageInformation(moniker.packageInformationId)
    }

    for (const id of mustGet(correlator.containsData, currentDocumentId, 'contains')) {
        const range = mustGet(correlator.rangeData, id, 'range')
        addHover(range.hoverResultId)
        for (const monikerId of range.monikerIds) {
            addMoniker(monikerId)
        }

        document.ranges.set(id, range)
    }

    return document
}

/**
 * Return the value of `id`, or throw an exception if it is undefined.
 *
 * @param id The identifier.
 */
function assertId<T extends lsif.Id>(id: T | undefined): T {
    if (id !== undefined) {
        return id
    }

    throw new Error('Id is undefined')
}
