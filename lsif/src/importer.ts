import { mustGet, assertId, hashKey, readEnvInt } from './util'
import { Correlator, ResultSetData, ResultSetId } from './correlator'
import { DefaultMap } from './default-map'
import {
    DefinitionModel,
    DocumentData,
    DocumentModel,
    MetaModel,
    MonikerData,
    PackageInformationData,
    RangeData,
    ReferenceModel,
    ResultChunkModel,
    DocumentIdRangeId,
    DefinitionResultId,
    MonikerId,
    DefinitionReferenceResultId,
    DocumentId,
    ReferenceResultId,
    PackageInformationId,
    HoverResultId,
    Ix,
    OrderedRanges,
} from './models.database'
import { Edge, MonikerKind, Vertex, RangeId } from 'lsif-protocol'
import { encodeJSON } from './encoding'
import { EntityManager } from 'typeorm'
import { isEqual, uniqWith } from 'lodash'
import { Package, SymbolReferences } from './xrepo'
import { TableInserter } from './inserter'

/**
 * The internal version of our SQLite databases. We need to keep this in case
 * we add something that can't be done transparently; if we change how we process
 * something in the future we'll need to consider a number of previous version
 * while we update or re-process the already-uploaded data.
 */
const INTERNAL_LSIF_VERSION = '0.1.0'

/**
 * The target results per result chunk. This is used to determine the number of chunks
 * created during conversion, but does not guarantee that the distribution of hash keys
 * will wbe even. In practice, chunks are fairly evenly filled.
 */
const RESULTS_PER_RESULT_CHUNK = readEnvInt('RESULTS_PER_RESULT_CHUNK', 500)

/**
 * The maximum number of result chunks that will be created during conversion.
 */
const MAX_NUM_RESULT_CHUNKS = readEnvInt('MAX_NUM_RESULT_CHUNKS', 1000)

/**
 * Correlate each vertex and edge together, then populate the provided entity manager
 * with the document, definition, and reference information. Returns the package and
 * external reference data needed to populate the correlation database.
 *
 * @param entityManager A transactional SQLite entity manager.
 * @param elements The stream of vertex and edge objects composing the LSIF dump.
 */
export async function importLsif(
    entityManager: EntityManager,
    elements: AsyncIterable<Vertex | Edge>
): Promise<{ packages: Package[]; references: SymbolReferences[] }> {
    const correlator = new Correlator()

    let line = 0
    for await (const element of elements) {
        try {
            correlator.insert(element)
        } catch (e) {
            throw Object.assign(
                new Error(`Failed to process line #${line + 1} (${JSON.stringify(element)}): ${e && e.message}`),
                { status: 422 }
            )
        }

        line++
    }

    if (correlator.lsifVersion === undefined) {
        throw new Error('No metadata defined.')
    }

    const numResults = correlator.definitionData.size + correlator.referenceData.size
    const numResultChunks = Math.min(MAX_NUM_RESULT_CHUNKS, Math.floor(numResults / RESULTS_PER_RESULT_CHUNK) || 1)

    // Insert metadata
    const metaInserter = new TableInserter(entityManager, MetaModel, getBatchSize(3))
    await populateMetadataTable(correlator, metaInserter, numResultChunks)
    await metaInserter.flush()

    // Insert documents
    const documentInserter = new TableInserter(entityManager, DocumentModel, getBatchSize(2))
    await populateDocumentsTable(correlator, documentInserter)
    await documentInserter.flush()

    // Insert result chunks
    const resultChunkInserter = new TableInserter(entityManager, ResultChunkModel, getBatchSize(2))
    await populateResultChunksTable(correlator, resultChunkInserter, numResultChunks)
    await resultChunkInserter.flush()

    // Insert definitions and references
    const definitionInserter = new TableInserter(entityManager, DefinitionModel, getBatchSize(8))
    const referenceInserter = new TableInserter(entityManager, ReferenceModel, getBatchSize(8))
    await populateDefinitionsAndReferencesTables(correlator, definitionInserter, referenceInserter)
    await definitionInserter.flush()
    await referenceInserter.flush()

    // Return correlation data to populate xrepo database
    return prepareCorrelationData(correlator)
}

/**
 * Determine the table inserter batch size for an entity given the number of
 * fields inserted for that entity. We cannot perform an insert operation with
 * more than 999 placeholder variables, so we need to flush our batch before
 * we reach that amount. If fields are added to the models, the argument to
 * this function also needs to change.
 *
 * @param numFields The number of fields for an entity.
 */
function getBatchSize(numFields: number): number {
    return Math.floor(999 / numFields)
}

/**
 * Correlate, encode, and insert all document entries for this dump.
 */
async function populateDocumentsTable(
    correlator: Correlator,
    documentInserter: TableInserter<DocumentModel, new () => DocumentModel>
): Promise<void> {
    // Collapse result sets data into the ranges that can reach them. The
    // remainder of this function assumes that we can completely ignore
    // the "next" edges coming from range data.
    for (const [rangeId, range] of correlator.rangeData) {
        canonicalizeItem(correlator, rangeId, range)
    }

    // Gather and insert document data that includes the ranges contained in the document,
    // any associated hover data, and any associated moniker data/package information.
    // Each range also has identifiers that correlate to a definition or reference result
    // which can be found in a result chunk, created in the next step.

    for (const [documentId, documentPath] of correlator.documentPaths) {
        // Create document record from the correlated information. This will also insert
        // external definitions and references into the maps initialized above, which are
        // inserted into the definitions and references table, respectively, below.
        const document = gatherDocument(correlator, documentId, documentPath)

        // Encode and insert document record
        await documentInserter.insert({
            path: documentPath,
            data: await encodeJSON({
                ranges: document.ranges,
                orderedRanges: document.orderedRanges,
                hoverResults: document.hoverResults,
                monikers: document.monikers,
                packageInformation: document.packageInformation,
            }),
        })
    }
}

/**
 * Correlate and insert all result chunk entries for this dump.
 */
async function populateResultChunksTable(
    correlator: Correlator,
    resultChunkInserter: TableInserter<ResultChunkModel, new () => ResultChunkModel>,
    numResultChunks: number
): Promise<void> {
    // Create all the result chunks we'll be populating and inserting up-front. Data will
    // be inserted into result chunks based on hash values (modulo the number of result chunks),
    // and we don't want to create them lazily.

    const resultChunks = new Array(numResultChunks).fill(null).map(() => ({
        paths: new Map<DocumentId, string>(),
        documentIdRangeIds: new Map<DefinitionReferenceResultId, DocumentIdRangeId[]>(),
    }))

    const chunkResults = (data: Map<DefinitionReferenceResultId, Map<DocumentId, RangeId[]>>): void => {
        for (const [id, documentRanges] of data) {
            // Flatten map into list of ranges
            let documentIdRangeIds: DocumentIdRangeId[] = []
            for (const [documentId, rangeIds] of documentRanges) {
                documentIdRangeIds = documentIdRangeIds.concat(rangeIds.map(rangeId => ({ documentId, rangeId })))
            }

            // Insert ranges into target result chunk
            const resultChunk = resultChunks[hashKey(id, resultChunks.length)]
            resultChunk.documentIdRangeIds.set(id, documentIdRangeIds)

            for (const documentId of documentRanges.keys()) {
                // Add paths into the result chunk where they are used
                resultChunk.paths.set(documentId, mustGet(correlator.documentPaths, documentId, 'documentPath'))
            }
        }
    }

    // Add definitions and references to result chunks
    chunkResults(correlator.definitionData)
    chunkResults(correlator.referenceData)

    for (let id = 0; id < resultChunks.length; id++) {
        // Empty chunk, no need to serialize as it will never be queried
        if (resultChunks[id].paths.size === 0 && resultChunks[id].documentIdRangeIds.size === 0) {
            continue
        }

        const data = await encodeJSON({
            documentPaths: resultChunks[id].paths,
            documentIdRangeIds: resultChunks[id].documentIdRangeIds,
        })

        // Encode and insert result chunk record
        await resultChunkInserter.insert({ id, data })
    }
}

/**
 * Correlate and insert all definition and reference entries for this dump.
 */
async function populateDefinitionsAndReferencesTables(
    correlator: Correlator,
    definitionInserter: TableInserter<DefinitionModel, new () => DefinitionModel>,
    referenceInserter: TableInserter<ReferenceModel, new () => ReferenceModel>
): Promise<void> {
    // Determine the set of monikers that are attached to a definition or a
    // reference result. Correlating information in this way has two benefits:
    //   (1) it reduces duplicates in the definitions and references tables
    //   (2) it stop us from re-iterating over the range data of the entire
    //       LSIF dump, which is by far the largest proportion of data.

    const definitionMonikers = new DefaultMap<DefinitionResultId, Set<MonikerId>>(() => new Set<MonikerId>())
    const referenceMonikers = new DefaultMap<ReferenceResultId, Set<MonikerId>>(() => new Set<MonikerId>())

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
        data: Map<DefinitionReferenceResultId, Map<DocumentId, RangeId[]>>,
        monikers: Map<MonikerId, Set<RangeId>>,
        inserter: TableInserter<DefinitionModel | ReferenceModel, new () => DefinitionModel | ReferenceModel>
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

    // Insert definitions and references records.
    await insertMonikerRanges(correlator.definitionData, definitionMonikers, definitionInserter)
    await insertMonikerRanges(correlator.referenceData, referenceMonikers, referenceInserter)
}

/**
 * Insert metadata row. This gives us a place to store the version of the converter that
 * created a database in case we have backwards-incompatible changes in the future that
 * require historic version flagging. This also stores the number of result chunks
 * determined above so that we can have stable hashes at query time.
 */
async function populateMetadataTable(
    correlator: Correlator,
    metaInserter: TableInserter<MetaModel, new () => MetaModel>,
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
 * Get the data needed to add this dump's package and dependency information into the
 * xrepo database.
 */
function prepareCorrelationData(correlator: Correlator): { packages: Package[]; references: SymbolReferences[] } {
    // Gather all package information that is referenced by an exported
    // moniker. These will be the packages that are provided by the repository
    // represented by this LSIF dump.

    const packageHashes: Package[] = []
    for (const monikerId of correlator.exportedMonikers) {
        const source = mustGet(correlator.monikerData, monikerId, 'moniker')
        const packageInformationId = assertId(source.packageInformationId)
        const packageInfo = mustGet(correlator.packageInformationData, packageInformationId, 'packageInformation')

        packageHashes.push({
            scheme: source.scheme,
            name: packageInfo.name,
            version: packageInfo.version,
        })
    }

    // Ensure packages are unique
    const exportedPackages = uniqWith(packageHashes, isEqual)

    // Gather all imported moniker identifiers along with their package
    // information. These will be the packages that are a dependency of the
    // repository represented by this LSIF dump.

    const packages = new Map<string, Package>()
    const packageIdentifiers = new DefaultMap<string, string[]>(() => [])
    for (const monikerId of correlator.importedMonikers) {
        const source = mustGet(correlator.monikerData, monikerId, 'moniker')
        const packageInformationId = assertId(source.packageInformationId)
        const packageInfo = mustGet(correlator.packageInformationData, packageInformationId, 'packageInformation')

        const key = `${source.scheme}::${packageInfo.name}::${packageInfo.version}`
        packages.set(key, { scheme: source.scheme, name: packageInfo.name, version: packageInfo.version })
        packageIdentifiers.getOrDefault(key).push(source.identifier)
    }

    // Create a unique list of package information and imported symbol pairs.
    // Ensure that each package is represented only once in the list.

    const importedReferences = Array.from(packages.keys()).map(key => ({
        package: mustGet(packages, key, 'package'),
        identifiers: mustGet(packageIdentifiers, key, 'packageIdentifier'),
    }))

    // Kick back the xrepo data needed to be inserted into the correlation database
    return { packages: exportedPackages, references: importedReferences }
}

/**
 * Flatten the definition result, reference result, hover results, and monikers of range
 * and result set items by following next links in the graph. This needs to be run over
 * each range before committing them to a document.
 *
 * @param correlator The correlator with all vertices and edges inserted.
 * @param id The item identifier.
 * @param item The range or result set item.
 */
function canonicalizeItem(correlator: Correlator, id: RangeId | ResultSetId, item: RangeData | ResultSetData): void {
    const monikers = new Set<MonikerId>()
    if (item.monikerIds.size > 0) {
        // If we have any monikers attached to this item, then we only need to look at the
        // monikers reachable from any attached moniker. All other attached monikers are
        // necessarily reachable, so we can choose any single value from the moniker set
        // as the source of the graph traversal.

        const candidateMoniker = item.monikerIds.keys().next().value

        for (const monikerId of reachableMonikers(correlator.monikerSets, candidateMoniker)) {
            if (mustGet(correlator.monikerData, monikerId, 'moniker').kind !== MonikerKind.local) {
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
        canonicalizeItem(correlator, nextId, nextItem)

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
function gatherDocument(correlator: Correlator, currentDocumentId: DocumentId, path: string): DocumentData {
    const document = {
        path,
        ranges: new Map<RangeId, Ix<OrderedRanges>>(),
        orderedRanges: [] as RangeData[],
        hoverResults: new Map<HoverResultId, string>(),
        monikers: new Map<MonikerId, MonikerData>(),
        packageInformation: new Map<PackageInformationId, PackageInformationData>(),
    }

    const addHover = (id: HoverResultId | undefined): void => {
        if (id === undefined || document.hoverResults.has(id)) {
            return
        }

        // Add hover result to the document, if defined and not a duplicate
        const data = mustGet(correlator.hoverData, id, 'hoverResult')
        document.hoverResults.set(id, data)
    }

    const addPackageInformation = (id: PackageInformationId | undefined): void => {
        if (id === undefined || document.packageInformation.has(id)) {
            return
        }

        // Add package information to the document, if defined and not a duplicate
        const data = mustGet(correlator.packageInformationData, id, 'packageInformation')
        document.packageInformation.set(id, data)
    }

    const addMoniker = (id: MonikerId | undefined): void => {
        if (id === undefined || document.monikers.has(id)) {
            return
        }

        // Add moniker to the document, if defined and not a duplicate
        const moniker = mustGet(correlator.monikerData, id, 'moniker')
        document.monikers.set(id, moniker)

        // Add related package information to document
        addPackageInformation(moniker.packageInformationId)
    }

    // Correlate range data with its id so after we sort we can pull out the ids in the
    // same order to make the identifier -> index mapping.
    const orderedRanges: (RangeData & { id: RangeId })[] = []

    for (const id of mustGet(correlator.containsData, currentDocumentId, 'contains')) {
        const range = mustGet(correlator.rangeData, id, 'range')
        orderedRanges.push({ id, ...range })
        addHover(range.hoverResultId)
        for (const id of range.monikerIds) {
            addMoniker(id)
        }
    }

    // Sort ranges by their starting position
    orderedRanges.sort((a, b) => a.startLine - b.startLine || a.startCharacter - b.startCharacter)

    // Populate a reverse lookup so ranges can be queried by id via `orderedRanges[range[id]]`.
    for (const [index, range] of orderedRanges.entries()) {
        document.ranges.set(range.id, index)
    }

    // eslint-disable-next-line require-atomic-updates
    document.orderedRanges = orderedRanges.map(({ id, ...range }) => range)

    return document
}

/**
 * Return the set of moniker identifiers which are reachable from the given value.
 * This relies on `monikerSets` being properly set up: each moniker edge `a -> b`
 * from the dump should ensure that `b` is a member of `monkerSets[a]`, and that
 * `a` is a member of `monikerSets[b]`.
 *
 * @param monikerSets A undirected graph of moniker ids.
 * @param id The initial moniker id.
 */
export function reachableMonikers(monikerSets: Map<MonikerId, Set<MonikerId>>, id: MonikerId): Set<MonikerId> {
    const monikerIds = new Set<MonikerId>()
    let frontier = [id]

    while (frontier.length > 0) {
        const val = assertId(frontier.pop())
        if (monikerIds.has(val)) {
            continue
        }

        monikerIds.add(val)

        const nextValues = monikerSets.get(val)
        if (nextValues) {
            frontier = frontier.concat(Array.from(nextValues))
        }
    }

    // TODO - (efritz) should we sort these ids here instead of at query time?
    return monikerIds
}
