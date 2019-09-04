import { EntityManager } from 'typeorm'
import { isEqual, uniqWith } from 'lodash'
import { DefinitionModel, DocumentModel, MetaModel, ReferenceModel } from './models.database'
import RelateUrl from 'relateurl'
import { encodeJSON } from './encoding'
import { TableInserter } from './inserter'
import { MonikerData, RangeData, PackageInformationData, DocumentData } from './entities'
import {
    Id,
    VertexLabels,
    EdgeLabels,
    Vertex,
    Edge,
    MonikerKind,
    ItemEdgeProperties,
    moniker,
    next,
    nextMoniker,
    textDocument_definition,
    textDocument_hover,
    textDocument_references,
    packageInformation,
    item,
    MetaData,
    ElementTypes,
    contains,
} from 'lsif-protocol'
import { Package, SymbolReferences } from './xrepo'
import { Hover, MarkupContent } from 'vscode-languageserver-types'

/**
 * The internal version of our SQLite databases. We need to keep this in case
 * we add something that can't be done transparently; if we change how we process
 * something in the future we'll need to consider a number of previous version
 * while we update or re-process the already-uploaded data.
 */
const INTERNAL_LSIF_VERSION = '0.1.0'

/**
 * An extension of `Map` that defines `getOrDefault` for a type of stunted
 * autovivification. This saves a bunch of code that needs to check if a
 * nested type within a map is undefined on first access.
 */
export class DefaultMap<K, V> extends Map {
    /**
     * Returns a new `DefaultMap`.
     *
     * @param defaultFactory The factory invoked when an undefined value is accessed.
     */
    constructor(private defaultFactory: () => V) {
        super()
    }

    /**
     * Get a key from the map. If the key does not exist, the default factory produces
     * a value and inserted into the map before being returned.
     *
     * @param key The key to retrieve.
     */
    public getOrDefault(key: K): V {
        let value = super.get(key)
        if (value !== undefined) {
            return value
        }

        value = this.defaultFactory()
        this.set(key, value)
        return value
    }
}

/**
 * An internal representation of a result set vertex. This is only used during import
 * as we flatten this data into the range vertices for faster queries.
 */
interface ResultSetData {
    /**
     * * The identifier of the definition result attached to this result set.
     */
    definitionResult?: Id

    /**
     * * The identifier of the reference result attached to this result set.
     */
    referenceResult?: Id

    /**
     * * The identifier of the hover result attached to this result set.
     */
    hoverResult?: Id

    /**
     * * The set of moniker identifiers directly attached to this result set.
     */
    monikers: Id[]
}

/**
 * Common state around the conversion of a single LSIF dump upload. This class
 * receives the parsed vertex or edge, line by line, from the caller, and adds it
 * into an in-memory structure that is later processed and converted into a SQLite
 * database on disk.
 */
class LsifCorrelator {
    /**
     * The LSIF version of the input. This is extracted from the metadata vertex at
     * the beginning of processing.
     */
    public lsifVersion?: string

    /**
     * The root of all document URIs. This is extracted from the metadata vertex at
     * the beginning of processing.
     */
    public projectRoot?: URL

    // Vertex data
    public documentPaths = new Map<Id, string>()
    public rangeData = new Map<Id, RangeData>()
    public resultSetData = new Map<Id, ResultSetData>()
    public hoverData = new Map<Id, string>()
    public monikerData = new Map<Id, MonikerData>()
    public packageInformationData = new Map<Id, PackageInformationData>()

    // Edge data
    public containsData = new Map<Id, Set<Id>>()
    public nextData = new Map<Id, Id>()
    public definitionData = new Map<Id, DefaultMap<Id, Id[]>>()
    public referenceData = new Map<Id, DefaultMap<Id, Id[]>>()

    /**
     * A mapping for the relation from moniker to the set of monikers that they are related
     * to via nextMoniker edges. This relation is symmetric (if `a` is in `MonikerSets[b]`,
     * then `b` is in `monikerSets[a]`).
     */
    public monikerSets = new DefaultMap<Id, Set<Id>>(() => new Set<Id>())

    /**
     * The set of exported moniker identifiers that have package information attached.
     */
    public importedMonikers = new Set<Id>()

    /**
     * The set of exported moniker identifiers that have package information attached.
     */
    public exportedMonikers = new Set<Id>()

    /**
     * Process a single vertex or edge.
     *
     * @param element A vertex or edge element from the LSIF dump.
     */
    public insert(element: Vertex | Edge): void {
        if (element.type === ElementTypes.vertex) {
            switch (element.label) {
                case VertexLabels.metaData:
                    this.handleMetaData(element)
                    break

                case VertexLabels.document: {
                    if (!this.projectRoot) {
                        throw new Error('No metadata defined.')
                    }

                    const path = RelateUrl.relate(this.projectRoot.href + '/', new URL(element.uri).href, {
                        defaultPorts: {},
                        output: RelateUrl.PATH_RELATIVE,
                        removeRootTrailingSlash: false,
                    })

                    this.documentPaths.set(element.id, path)
                    this.containsData.set(element.id, new Set<Id>())
                    break
                }

                // The remaining vertex handlers stash data into an appropriate map. This data
                // may be retrieved when an edge that references it is seen, or when a document
                // is finalized.

                case VertexLabels.range:
                    this.rangeData.set(element.id, {
                        startLine: element.start.line,
                        startCharacter: element.start.character,
                        endLine: element.end.line,
                        endCharacter: element.end.character,
                        monikers: [],
                    })
                    break
                case VertexLabels.resultSet:
                    this.resultSetData.set(element.id, { monikers: [] })
                    break
                case VertexLabels.definitionResult:
                    this.definitionData.set(element.id, new DefaultMap<Id, Id[]>(() => []))
                    break
                case VertexLabels.referenceResult:
                    this.referenceData.set(element.id, new DefaultMap<Id, Id[]>(() => []))
                    break
                case VertexLabels.hoverResult:
                    this.hoverData.set(element.id, normalizeHover(element.result))
                    break
                case VertexLabels.moniker:
                    this.monikerData.set(element.id, {
                        kind: element.kind || MonikerKind.local,
                        scheme: element.scheme,
                        identifier: element.identifier,
                    })
                    break
                case VertexLabels.packageInformation:
                    this.packageInformationData.set(element.id, {
                        name: element.name,
                        version: element.version || '$missing',
                    })
                    break
            }
        }

        if (element.type === ElementTypes.edge) {
            switch (element.label) {
                case EdgeLabels.contains:
                    this.handleContains(element)
                    break
                case EdgeLabels.next:
                    this.handleNextEdge(element)
                    break
                case EdgeLabels.item:
                    this.handleItemEdge(element)
                    break
                case EdgeLabels.textDocument_definition:
                    this.handleDefinitionEdge(element)
                    break
                case EdgeLabels.textDocument_references:
                    this.handleReferenceEdge(element)
                    break
                case EdgeLabels.textDocument_hover:
                    this.handleHoverEdge(element)
                    break
                case EdgeLabels.moniker:
                    this.handleMonikerEdge(element)
                    break
                case EdgeLabels.nextMoniker:
                    this.handleNextMonikerEdge(element)
                    break
                case EdgeLabels.packageInformation:
                    this.handlePackageInformationEdge(element)
                    break
            }
        }
    }

    //
    // Vertex Handlers

    /**
     * This should be the first vertex seen. Extract the project root so we
     * can create relative paths for documents. Insert a row in the meta
     * table with the LSIF protocol version.
     *
     * @param vertex The metadata vertex.
     */
    private handleMetaData(vertex: MetaData): void {
        this.lsifVersion = vertex.version
        this.projectRoot = new URL(vertex.projectRoot)
    }

    //
    // Edge Handlers

    /**
     * Add range data ids into the document in which they are contained. Ensures
     * all referenced vertices are defined.
     *
     * @param edge The contains edge.
     */
    private handleContains(edge: contains): void {
        // Do not track project contains
        if (!this.documentPaths.has(edge.outV)) {
            return
        }

        const set = assertDefined(edge.outV, 'contains', this.containsData)
        for (const inV of edge.inVs) {
            assertDefined(inV, 'range', this.rangeData)
            set.add(inV)
        }
    }

    /**
     * Update definition and reference fields from an item edge. Ensures all
     * referenced vertices are defined.
     *
     * @param edge The item edge.
     */
    private handleItemEdge(edge: item): void {
        switch (edge.property) {
            // `item` edges with a `property` refer to a referenceResult
            case ItemEdgeProperties.definitions:
            case ItemEdgeProperties.references: {
                const documentMap = assertDefined(edge.outV, 'referenceResult', this.referenceData)
                const rangeIds = documentMap.getOrDefault(edge.document)
                for (const inV of edge.inVs) {
                    assertDefined(inV, 'range', this.rangeData)
                    rangeIds.push(inV)
                }

                break
            }

            // `item` edges without a `property` refer to a definitionResult
            case undefined: {
                const documentMap = assertDefined(edge.outV, 'definitionResult', this.definitionData)
                const rangeIds = documentMap.getOrDefault(edge.document)
                for (const inV of edge.inVs) {
                    assertDefined(inV, 'range', this.rangeData)
                    rangeIds.push(inV)
                }

                break
            }
        }
    }

    /**
     * Attaches the specified moniker to the specified range or result set. Ensures all referenced
     * vertices are defined.
     *
     * @param edge The moniker edge.
     */
    private handleMonikerEdge(edge: moniker): void {
        const source = assertDefined<RangeData | ResultSetData>(
            edge.outV,
            'range/resultSet',
            this.rangeData,
            this.resultSetData
        )
        assertDefined(edge.inV, 'moniker', this.monikerData)
        source.monikers = [edge.inV]
    }

    /**
     * Sets the next field fo the specified range or result set. Ensures all referenced vertices
     * are defined.
     *
     * @param edge The next edge.
     */
    private handleNextEdge(edge: next): void {
        assertDefined<RangeData | ResultSetData>(edge.outV, 'range/resultSet', this.rangeData, this.resultSetData)
        assertDefined(edge.inV, 'resultSet', this.resultSetData)
        this.nextData.set(edge.outV, edge.inV)
    }

    /**
     * Correlates monikers together so that when one moniker is queried, each correlated moniker
     * is also returned as a strongly connected set. Ensures all referenced vertices are defined.
     *
     * @param edge The nextMoniker edge.
     */
    private handleNextMonikerEdge(edge: nextMoniker): void {
        assertDefined(edge.inV, 'moniker', this.monikerData)
        assertDefined(edge.outV, 'moniker', this.monikerData)
        this.monikerSets.getOrDefault(edge.inV).add(edge.outV) // Forward direction
        this.monikerSets.getOrDefault(edge.outV).add(edge.inV) // Backwards direction
    }

    /**
     * Sets the package information of the specified moniker. If the moniker is an export moniker,
     * then the package information will also be returned as an exported package by the `finalize`
     * method. Ensures all referenced vertices are defined.
     *
     * @param edge The packageInformation edge.
     */
    private handlePackageInformationEdge(edge: packageInformation): void {
        const source = assertDefined(edge.outV, 'moniker', this.monikerData)
        assertDefined(edge.inV, 'packageInformation', this.packageInformationData)
        source.packageInformation = edge.inV

        if (source.kind === 'export') {
            this.exportedMonikers.add(edge.outV)
        }

        if (source.kind === 'import') {
            this.importedMonikers.add(edge.outV)
        }
    }

    /**
     * Sets the definition result of the specified range or result set. Ensures all referenced
     * vertices are defined.
     *
     * @param edge The textDocument/definition edge.
     */
    private handleDefinitionEdge(edge: textDocument_definition): void {
        const outV = assertDefined<RangeData | ResultSetData>(
            edge.outV,
            'range/resultSet',
            this.rangeData,
            this.resultSetData
        )
        assertDefined(edge.inV, 'definitionResult', this.definitionData)
        outV.definitionResult = edge.inV
    }

    /**
     * Sets the hover result of the specified range or result set. Ensures all referenced
     * vertices are defined.
     *
     * @param edge The textDocument/hover edge.
     */
    private handleHoverEdge(edge: textDocument_hover): void {
        const outV = assertDefined<RangeData | ResultSetData>(
            edge.outV,
            'range/resultSet',
            this.rangeData,
            this.resultSetData
        )
        assertDefined(edge.inV, 'hoverResult', this.hoverData)
        outV.hoverResult = edge.inV
    }

    /**
     * Sets the reference result of the specified range or result set. Ensures all
     * referenced vertices are defined.
     *
     * @param edge The textDocument/references edge.
     */
    private handleReferenceEdge(edge: textDocument_references): void {
        const outV = assertDefined<RangeData | ResultSetData>(
            edge.outV,
            'range/resultSet',
            this.rangeData,
            this.resultSetData
        )
        assertDefined(edge.inV, 'referenceResult', this.referenceData)
        outV.referenceResult = edge.inV
    }
}

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
    const correlator = new LsifCorrelator()

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
    correlator: LsifCorrelator,
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
function canonicalizeRange(correlator: LsifCorrelator, id: Id, range: RangeData): void {
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
 * Return the value of `id`, or throw an exception if it is undefined.
 *
 * @param id The identifier.
 */
function assertId(id: Id | undefined): Id {
    if (id !== undefined) {
        return id
    }

    throw new Error('id is undefined')
}

/**
 * Return the value of the key `id` in one of the given maps. The first value
 * to exist is returned. If the key does not exist in any map, an exception is
 * thrown.
 *
 * @param id The id to search for.
 * @param name The type of element (used for exception message).
 * @param maps The set of maps to query.
 */
export function assertDefined<T>(id: Id, name: string, ...maps: Map<Id, T>[]): T {
    for (const map of maps) {
        const value = map.get(id)
        if (value !== undefined) {
            return value
        }
    }

    throw new Error(`Unknown ${name} '${id}'.`)
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

/**
 * Normalize an LSP hover object into a string.
 *
 * @param hover The hover object.
 */
export function normalizeHover(hover: Hover): string {
    const normalizeContent = (content: string | MarkupContent | { language: string; value: string }): string => {
        if (typeof content === 'string') {
            return content
        }

        if (MarkupContent.is(content)) {
            return content.value
        }

        const tick = '```'
        return `${tick}${content.language}\n${content.value}\n${tick}`
    }

    const separator = '\n\n---\n\n'
    const contents = Array.isArray(hover.contents) ? hover.contents : [hover.contents]
    return contents
        .map(c => normalizeContent(c).trim())
        .filter(s => s)
        .join(separator)
}
