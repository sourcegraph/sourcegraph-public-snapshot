import { assertDefined, assertId } from './util'
import { Correlator, ResultSetData } from './correlator'
import { DefaultMap } from './default-map'
import { DefinitionModel, DocumentModel, MetaModel, ReferenceModel } from './models.database'
import { DocumentData, MonikerData, PackageInformationData, RangeData } from './entities'
import { Edge, Id, MonikerKind, Vertex } from 'lsif-protocol'
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
 * Handle the life-cycle of an importer. Creates an `LsifImporter`. This will create
 * a new importer, insert each vertex and edge in the given stream, then call the
 * importer's finalize method.
 *
 * @param entityManager A transactional SQLite entity manager.
 * @param elements The stream of vertex and edge objects composing the LSIF dump.
 */
export async function importLsif(
    entityManager: EntityManager,
    elements: AsyncIterable<Vertex | Edge>
): Promise<{ packages: Package[]; references: SymbolReferences[] }> {
    const correlator = new Correlator()

    let i = 0
    for await (const element of elements) {
        try {
            correlator.insert(element)
        } catch (e) {
            console.log(e)
            throw Object.assign(new Error(`Failed to process line:\n${i}`), { e })
        }

        i++
    }

    // Determine the max batch size of each model type. We cannot perform an
    // insert operation with more than 999 placeholder variables, so we need
    // to flush our batch before we reach that amount. The batch size for each
    // model is calculated based on the number of fields inserted. If fields
    // are added to the models, these numbers will also need to change.

    const metaInserter = new TableInserter(entityManager, MetaModel, Math.floor(999 / 3))
    const documentInserter = new TableInserter(entityManager, DocumentModel, Math.floor(999 / 2))
    const defInserter = new TableInserter(entityManager, DefinitionModel, Math.floor(999 / 8))
    const refInserter = new TableInserter(entityManager, ReferenceModel, Math.floor(999 / 8))

    if (correlator.lsifVersion === undefined) {
        throw new Error('No metadata defined.')
    }

    await metaInserter.insert({ lsifVersion: correlator.lsifVersion, sourcegraphVersion: INTERNAL_LSIF_VERSION })

    const definitions = new DefaultMap<Id, DefaultMap<Id, Set<Id>>>(
        () => new DefaultMap<Id, Set<Id>>(() => new Set<Id>())
    )

    const references = new DefaultMap<Id, DefaultMap<Id, Set<Id>>>(
        () => new DefaultMap<Id, Set<Id>>(() => new Set<Id>())
    )

    for (const [id, path] of correlator.documentPaths) {
        // Finalize document
        const document = finalizeDocument(correlator, definitions, references, id, path)

        const data = await encodeJSON({
            ranges: document.ranges,
            orderedRanges: document.orderedRanges,
            definitionResults: document.definitionResults,
            referenceResults: document.referenceResults,
            hoverResults: document.hoverResults,
            monikers: document.monikers,
            packageInformation: document.packageInformation,
        })

        // Insert document record
        await documentInserter.insert({ path, data })
    }

    // Insert all related definitions
    for (const [documentId, m] of definitions) {
        for (const [rangeId, monikerIds] of m) {
            for (const monikerId of monikerIds) {
                // TODO - clean this up
                const range = assertDefined(rangeId, 'range', correlator.rangeData)
                const moniker = assertDefined(monikerId, 'moniker', correlator.monikerData)
                const documentPath = assertDefined(documentId, 'documentPath', correlator.documentPaths)

                await defInserter.insert({
                    scheme: moniker.scheme,
                    identifier: moniker.identifier,
                    documentPath,
                    ...range,
                })
            }
        }
    }

    // Insert all related references
    for (const [documentId, m] of references) {
        for (const [rangeId, monikerIds] of m) {
            for (const monikerId of monikerIds) {
                const range = assertDefined(rangeId, 'range', correlator.rangeData)
                const moniker = assertDefined(monikerId, 'moniker', correlator.monikerData)
                const documentPath = assertDefined(documentId, 'documentPath', correlator.documentPaths)

                await refInserter.insert({
                    scheme: moniker.scheme,
                    identifier: moniker.identifier,
                    documentPath,
                    ...range,
                })
            }
        }
    }

    await metaInserter.finalize()
    await documentInserter.finalize()
    await defInserter.finalize()
    await refInserter.finalize()

    const packageHashes: Package[] = []
    for (const id of correlator.exportedMonikers) {
        const source = assertDefined(id, 'moniker', correlator.monikerData)
        const packageInformationId = assertId(source.packageInformation)
        const packageInfo = assertDefined(packageInformationId, 'packageInformation', correlator.packageInformationData)

        packageHashes.push({
            scheme: source.scheme,
            name: packageInfo.name,
            version: packageInfo.version,
        })
    }

    const packageIdentifiers = new DefaultMap<string, string[]>(() => [])
    for (const id of correlator.importedMonikers) {
        const source = assertDefined(id, 'moniker', correlator.monikerData)
        const packageInformationId = assertId(source.packageInformation)
        const packageInfo = assertDefined(packageInformationId, 'packageInformation', correlator.packageInformationData)

        // TODO - same issue as the lodash thing above?
        const pkg = JSON.stringify({
            scheme: source.scheme,
            name: packageInfo.name,
            version: packageInfo.version,
        })

        packageIdentifiers.getOrDefault(pkg).push(source.identifier)
    }

    return {
        packages: uniqWith(packageHashes, isEqual),
        references: Array.from(packageIdentifiers).map(([key, identifiers]) => ({
            package: JSON.parse(key) as Package,
            identifiers,
        })),
    }
}

/**
 * Create a self-contained document object.
 *
 * @param correlator The correlator with all vertices and edges inserted.
 * @param currentDocumentId The identifier of the document.
 * @param path The path of the document.
 */
function finalizeDocument(
    correlator: Correlator,
    definitions: DefaultMap<Id, DefaultMap<Id, Set<Id>>>,
    references: DefaultMap<Id, DefaultMap<Id, Set<Id>>>,
    currentDocumentId: Id,
    path: string
): DocumentData {
    const document = {
        path,
        ranges: new Map<Id, number>(),
        orderedRanges: [] as RangeData[],
        definitionResults: new Map<Id, { documentPath: string; id: Id }[]>(),
        referenceResults: new Map<Id, { documentPath: string; id: Id }[]>(),
        hoverResults: new Map<Id, string>(),
        monikers: new Map<Id, MonikerData>(),
        packageInformation: new Map<Id, PackageInformationData>(),
    }

    const addHover = (id: Id | undefined): void => {
        if (id !== undefined && !document.hoverResults.has(id)) {
            document.hoverResults.set(id, assertDefined(id, 'hoverResult', correlator.hoverData))
        }
    }

    const addPackageInformation = (id: Id | undefined): void => {
        if (id !== undefined && !document.packageInformation.has(id)) {
            document.packageInformation.set(
                id,
                assertDefined(id, 'packageInformation', correlator.packageInformationData)
            )
        }
    }

    const addMoniker = (id: Id | undefined): void => {
        if (id !== undefined && !document.monikers.has(id)) {
            const moniker = assertDefined(id, 'moniker', correlator.monikerData)
            document.monikers.set(id, moniker)
            addPackageInformation(moniker.packageInformation)
        }
    }

    const addGenericResult = (
        name: string,
        datas: Map<Id, Map<Id, Id[]>>,
        monikerResults: DefaultMap<Id, DefaultMap<Id, Set<Id>>>,
        results: Map<Id, { documentPath: string; id: Id }[]>,
        id: Id | undefined,
        monikers: Id[]
    ): void => {
        if (!id) {
            return
        }

        const m = monikerResults.getOrDefault(currentDocumentId)

        const values = []
        for (const [documentId, ids] of assertDefined(id, name, datas)) {
            // Resolve the "document" field from the "item" edge. This will correlate
            // the referenced range identifier with the document in which it belongs.
            const documentPath = assertDefined(documentId, 'documentPath', correlator.documentPaths)

            for (const id of ids) {
                values.push({ documentPath, id })
            }

            if (documentId === currentDocumentId) {
                // If this is results for the current document, construct the data that
                // will later be used to insert into the definitions table for this
                // document.

                for (const id of ids) {
                    const n = m.getOrDefault(id)
                    for (const moniker of monikers) {
                        n.add(moniker)
                    }
                }
            }
        }

        results.set(id, values)
    }

    const orderedRanges: (RangeData & { id: Id })[] = []
    for (const id of assertDefined(currentDocumentId, 'contains', correlator.containsData)) {
        const range = assertDefined(id, 'range', correlator.rangeData)
        canonicalizeRange(correlator, id, range)
        orderedRanges.push({ id, ...range })

        addHover(range.hoverResult)

        for (const id of range.monikers) {
            addMoniker(id)
        }

        addGenericResult(
            'definitionResult',
            correlator.definitionData,
            definitions,
            document.definitionResults,
            range.definitionResult,
            range.monikers
        )

        addGenericResult(
            'referenceResult',
            correlator.referenceData,
            references,
            document.referenceResults,
            range.referenceResult,
            range.monikers
        )
    }

    // Sort ranges by their starting position
    orderedRanges.sort((a, b) => a.startLine - b.startLine || a.startCharacter - b.startCharacter)

    // Populate a reverse lookup so ranges can be queried by id
    // via `orderedRanges[range[id]]`.
    for (const [index, range] of orderedRanges.entries()) {
        document.ranges.set(range.id, index)
    }

    // eslint-disable-next-line require-atomic-updates
    document.orderedRanges = orderedRanges.map(({ id, ...range }) => range)

    return document
}

/**
 * Update the definition result, reference result, hover result, and monikers of the
 * given range with respect to the result sets reachable from the range. This puts
 * all of the necessary data about a range in the range object itself so we do not
 * have to traverse the graph at query time.
 *
 * @param correlator The correlator with all vertices and edges inserted.
 * @param id The identifier of the range.
 * @param range The range to canonicalize.
 */
function canonicalizeRange(correlator: Correlator, id: Id, range: RangeData): void {
    let definitionResult: Id | undefined
    let referenceResult: Id | undefined
    let hoverResult: Id | undefined
    const monikers = new Set<Id>()

    let itemId: Id = id
    let item: RangeData | ResultSetData | undefined = range

    while (item) {
        definitionResult = definitionResult === undefined ? item.definitionResult : definitionResult
        referenceResult = referenceResult === undefined ? item.referenceResult : referenceResult
        hoverResult = hoverResult === undefined ? item.hoverResult : hoverResult

        if (item.monikers.length > 0) {
            for (const mon of reachableMonikers(correlator.monikerSets, item.monikers[0])) {
                if (assertDefined(mon, 'moniker', correlator.monikerData).kind !== MonikerKind.local) {
                    monikers.add(mon)
                }
            }
        }

        const nextId = correlator.nextData.get(itemId)
        if (nextId === undefined) {
            break
        }

        itemId = nextId
        item = assertDefined(nextId, 'resultSet', correlator.resultSetData)
    }

    range.definitionResult = definitionResult
    range.referenceResult = referenceResult
    range.hoverResult = hoverResult
    range.monikers = Array.from(monikers)
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
export function reachableMonikers(monikerSets: Map<Id, Set<Id>>, id: Id): Set<Id> {
    const combined = new Set<Id>()
    let frontier = [id]

    while (true) {
        const val = frontier.pop()
        if (val === undefined) {
            break
        }

        if (combined.has(val)) {
            continue
        }

        const nextValues = monikerSets.get(val)
        if (nextValues) {
            frontier = frontier.concat(Array.from(nextValues))
        }

        combined.add(val)
    }

    return combined
}
