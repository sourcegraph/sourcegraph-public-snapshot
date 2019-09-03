import { EntityManager } from 'typeorm'
import { isEqual, uniqWith } from 'lodash'
import { DefModel, DocumentModel, MetaModel, RefModel } from './models'
import RelateUrl from 'relateurl'
import { encodeJSON } from './encoding'
import { TableInserter } from './inserter'
import { MonikerData, RangeData, ResultSetData, PackageInformationData, FlattenedRange, DocumentData } from './entities'
import {
    Id,
    VertexLabels,
    EdgeLabels,
    Vertex,
    Edge,
    MonikerKind,
    ItemEdgeProperties,
    DocumentEvent,
    Event,
    moniker,
    next,
    nextMoniker,
    textDocument_definition,
    textDocument_hover,
    Range,
    textDocument_references,
    packageInformation,
    PackageInformation,
    item,
    EventKind,
    EventScope,
    MetaData,
    ElementTypes,
    Moniker,
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
 * Common state around the conversion of a single LSIF dump upload. This class
 * receives the parsed vertex or edge, line by line, from the caller, and adds it
 * into a new database file on disk. Once finalized, the database is ready for use
 * and relevant cross-repository metadata is returned to the caller, which
 * is used to populate the xrepo database.
 *
 * This class should not be used directly - use the `importLsif` function instead.
 */
class LsifImporter {
    // Bulk database inserters
    private metaInserter: TableInserter<MetaModel, new () => MetaModel>
    private documentInserter: TableInserter<DocumentModel, new () => DocumentModel>
    private defInserter: TableInserter<DefModel, new () => DefModel>
    private refInserter: TableInserter<RefModel, new () => RefModel>

    // Vertex data
    private definitionData: Map<Id, Map<Id, Id[]>> = new Map()
    private documentPaths: Map<Id, string> = new Map()
    private containsData = new Map<Id, Set<Id>>()
    private hoverData: Map<Id, string> = new Map()
    private monikerData: Map<Id, MonikerData> = new Map()
    private packageInformationData: Map<Id, PackageInformationData> = new Map()
    private rangeData: Map<Id, RangeData> = new Map()
    private referenceData: Map<Id, Map<Id, Id[]>> = new Map()
    private resultSetData: Map<Id, ResultSetData> = new Map()

    /**
     * The root of all document URIs. This is extracted from the metadata vertex at
     * the beginning of processing.
     */
    private projectRoot?: URL

    /**
     * A mapping for the relation from moniker to the set of monikers that they are related
     * to via nextMoniker edges. This relation is symmetric (if `a` is in `MonikerSets[b]`,
     * then `b` is in `monikerSets[a]`).
     */
    private monikerSets = new Map<Id, Set<Id>>()

    /**
     * The set of exported moniker identifiers that have package information attached.
     */
    private importedMonikers = new Set<Id>()

    /**
     * The set of exported moniker identifiers that have package information attached.
     */
    private exportedMonikers = new Set<Id>()

    /**
     * A map from document identifiers to range identifiers to the set of nonlocal
     * monikers attached to that range. These values will be used to populate the
     * Defs table.
     */
    private definitions = new Map<Id, Map<Id, Set<Id>>>()

    /**
     * A map from document identifiers to range identifiers to the set of nonlocal
     * monikers attached to that range. These values will be used to populate the
     * Refs table.
     */
    private references = new Map<Id, Map<Id, Set<Id>>>()

    /**
     * Create a new `LsifImporter` with the given entity manager.
     *
     * @param entityManager A transactional SQLite entity manager.
     */
    constructor(private entityManager: EntityManager) {
        // Determine the max batch size of each model type. We cannot perform an
        // insert operation with more than 999 placeholder variables, so we need
        // to flush our batch before we reach that amount. The batch size for each
        // model is calculated based on the number of fields inserted. If fields
        // are added to the models, these numbers will also need to change.

        this.metaInserter = new TableInserter(this.entityManager, MetaModel, Math.floor(999 / 3))
        this.documentInserter = new TableInserter(this.entityManager, DocumentModel, Math.floor(999 / 2))
        this.defInserter = new TableInserter(this.entityManager, DefModel, Math.floor(999 / 8))
        this.refInserter = new TableInserter(this.entityManager, RefModel, Math.floor(999 / 8))
    }

    /**
     * Process a single vertex or edge.
     *
     * @param element A vertex or edge element from the LSIF dump.
     */
    public async insert(element: Vertex | Edge): Promise<void> {
        if (element.type === ElementTypes.vertex) {
            switch (element.label) {
                case VertexLabels.metaData:
                    await this.handleMetaData(element)
                    break
                case VertexLabels.event:
                    await this.handleEvent(element)
                    break

                // The remaining vertex handlers stash data into an appropriate map. This data
                // may be retrieved when an edge that references it is seen, or when a document
                // is finalized.

                case VertexLabels.definitionResult:
                    this.definitionData.set(element.id, new Map())
                    break
                case VertexLabels.document: {
                    if (!this.projectRoot) {
                        throw new Error('No project root has been defined.')
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

                case VertexLabels.hoverResult:
                    this.hoverData.set(element.id, normalizeHover(element.result))
                    break
                case VertexLabels.moniker:
                    this.monikerData.set(element.id, convertMoniker(element))
                    break
                case VertexLabels.packageInformation:
                    this.packageInformationData.set(element.id, convertPackageInformation(element))
                    break
                case VertexLabels.range:
                    this.rangeData.set(element.id, convertRange(element))
                    break
                case VertexLabels.referenceResult:
                    this.referenceData.set(element.id, new Map())
                    break
                case VertexLabels.resultSet:
                    this.resultSetData.set(element.id, { monikers: [] })
                    break
            }
        }

        if (element.type === ElementTypes.edge) {
            switch (element.label) {
                case EdgeLabels.contains:
                    this.handleContains(element)
                    break
                case EdgeLabels.item:
                    this.handleItemEdge(element)
                    break
                case EdgeLabels.moniker:
                    this.handleMonikerEdge(element)
                    break
                case EdgeLabels.next:
                    this.handleNextEdge(element)
                    break
                case EdgeLabels.nextMoniker:
                    this.handleNextMonikerEdge(element)
                    break
                case EdgeLabels.packageInformation:
                    this.handlePackageInformationEdge(element)
                    break
                case EdgeLabels.textDocument_definition:
                    this.handleDefinitionEdge(element)
                    break
                case EdgeLabels.textDocument_hover:
                    this.handleHoverEdge(element)
                    break
                case EdgeLabels.textDocument_references:
                    this.handleReferenceEdge(element)
                    break
            }
        }
    }

    /**
     * Ensure that any outstanding records are flushed to the database. Also
     * returns the set of packages provided by the project analyzed by this
     * LSIF dump as well as the symbols imported into the LSIF dump from
     * external packages.
     */
    public async finalize(): Promise<{ packages: Package[]; references: SymbolReferences[] }> {
        // Insert all related definitions
        for (const [documentId, m] of this.definitions) {
            for (const [rangeId, monikerIds] of m) {
                for (const monikerId of monikerIds) {
                    // TODO - clean this up
                    const range = assertDefined(rangeId, 'range', this.rangeData)
                    const moniker = assertDefined(monikerId, 'moniker', this.monikerData)
                    const documentPath = assertDefined(documentId, 'documentPath', this.documentPaths)

                    await this.defInserter.insert({
                        scheme: moniker.scheme,
                        identifier: moniker.identifier,
                        documentPath,
                        ...range,
                    })
                }
            }
        }

        // Insert all related references
        for (const [documentId, m] of this.references) {
            for (const [rangeId, monikerIds] of m) {
                for (const monikerId of monikerIds) {
                    const range = assertDefined(rangeId, 'range', this.rangeData)
                    const moniker = assertDefined(monikerId, 'moniker', this.monikerData)
                    const documentPath = assertDefined(documentId, 'documentPath', this.documentPaths)

                    await this.refInserter.insert({
                        scheme: moniker.scheme,
                        identifier: moniker.identifier,
                        documentPath,
                        ...range,
                    })
                }
            }
        }

        await this.metaInserter.finalize()
        await this.documentInserter.finalize()
        await this.defInserter.finalize()
        await this.refInserter.finalize()

        return { packages: this.getPackages(), references: this.getReferences() }
    }

    /**
     * Return the set of packages provided by the project analyzed by this LSIF dump.
     */
    private getPackages(): Package[] {
        const packageHashes: Package[] = []
        for (const id of this.exportedMonikers) {
            const source = assertDefined(id, 'moniker', this.monikerData)
            const packageInformationId = assertId(source.packageInformation)
            const packageInfo = assertDefined(packageInformationId, 'packageInformation', this.packageInformationData)

            packageHashes.push({
                scheme: source.scheme,
                name: packageInfo.name,
                version: packageInfo.version,
            })
        }

        return uniqWith(packageHashes, isEqual)
    }

    /**
     * Return the symbols imported into the LSIF dump from external packages.
     */
    private getReferences(): SymbolReferences[] {
        const packageIdentifiers: Map<string, string[]> = new Map()
        for (const id of this.importedMonikers) {
            const source = assertDefined(id, 'moniker', this.monikerData)
            const packageInformationId = assertId(source.packageInformation)
            const packageInfo = assertDefined(packageInformationId, 'packageInformation', this.packageInformationData)
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

    //
    // Vertex Handlers

    /**
     * This should be the first vertex seen. Extract the project root so we
     * can create relative paths for documents. Insert a row in the meta
     * table with the LSIF protocol version.
     *
     * @param vertex The metadata vertex.
     */
    private async handleMetaData(vertex: MetaData): Promise<void> {
        this.projectRoot = new URL(vertex.projectRoot)
        await this.metaInserter.insert(convertMetadata(vertex))
    }

    /**
     * Insert documents when we see a document-scoped end event. This constructs
     * a `DocumentData` object that correlates data related to a range contained
     * in the document and inserts its encoded, compressed representation into
     * the database.
     *
     * @param vertex The event vertex.
     */
    private async handleEvent(vertex: Event): Promise<void> {
        if (vertex.scope !== EventScope.document || vertex.kind !== EventKind.end) {
            return
        }

        const event = vertex as DocumentEvent
        const path = assertDefined(event.data, 'documentPath', this.documentPaths)

        // Finalize document
        const document = this.finalizeDocument(event.data, path)

        // Insert document record
        await this.documentInserter.insert({
            path,
            value: await encodeJSON({
                ranges: document.ranges,
                orderedRanges: document.orderedRanges,
                resultSets: document.resultSets,
                definitionResults: document.definitionResults,
                referenceResults: document.referenceResults,
                hovers: document.hovers,
                monikers: document.monikers,
                packageInformation: document.packageInformation,
            }),
        })
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
        if (!this.containsData.has(edge.outV)) {
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
            case ItemEdgeProperties.references:
                this.handleGenericItemEdge(edge, 'referenceResult', this.referenceData)
                break

            // `item` edges without a `property` refer to a definitionResult
            case undefined:
                this.handleGenericItemEdge(edge, 'definitionResult', this.definitionData)
                break
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
        const outV = assertDefined<RangeData | ResultSetData>(
            edge.outV,
            'range/resultSet',
            this.rangeData,
            this.resultSetData
        )
        assertDefined(edge.inV, 'resultSet', this.resultSetData)
        outV.next = edge.inV
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
        this.correlateMonikers(edge.inV, edge.outV) // Forward direction
        this.correlateMonikers(edge.outV, edge.inV) // Backwards direction
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

    //
    // Helper Functions

    /**
     * Concatenate `edge.inVs` to the array at `map[edge.outV][edge.document]`.
     * If any field is undefined, it is created on the fly.
     *
     * @param edge The edge.
     * @param name The type of map (used for exception message).
     * @param map The map to populate.
     */
    private handleGenericItemEdge(edge: item, name: string, map: Map<Id, Map<Id, Id[]>>): void {
        const innerMap = assertDefined(edge.outV, name, map)
        const data = innerMap.get(edge.document)
        if (!data) {
            innerMap.set(edge.document, edge.inVs)
        } else {
            for (const inV of edge.inVs) {
                data.push(inV)
            }
        }
    }

    /**
     * Add `b` as a neighbor of `a` in `monikerSets`.
     *
     * @param a A moniker.
     * @param b A second moniker.
     */
    private correlateMonikers(a: Id, b: Id): void {
        const neighbors = this.monikerSets.get(a)
        if (neighbors) {
            neighbors.add(b)
        } else {
            this.monikerSets.set(a, new Set<Id>([b]))
        }
    }

    /**
     * Create a self-contained document object.
     *
     * @param currentDocumentId The identifier of the document.
     * @param path The path of the document.
     */
    private finalizeDocument(currentDocumentId: Id, path: string): DocumentData {
        const document = {
            path,
            ranges: new Map<Id, number>(),
            orderedRanges: [] as RangeData[],
            resultSets: new Map(), // TODO - remove
            definitionResults: new Map<Id, { documentPath: string; id: Id }[]>(),
            referenceResults: new Map<Id, { documentPath: string; id: Id }[]>(),
            hovers: new Map<Id, string>(),
            monikers: new Map<Id, MonikerData>(),
            packageInformation: new Map<Id, PackageInformationData>(),
        }

        const addHover = (id: Id | undefined): void => {
            if (id !== undefined && !document.hovers.has(id)) {
                document.hovers.set(id, assertDefined(id, 'hoverResult', this.hoverData))
            }
        }

        const addPackageInformation = (id: Id | undefined): void => {
            if (id !== undefined && !document.packageInformation.has(id)) {
                document.packageInformation.set(
                    id,
                    assertDefined(id, 'packageInformation', this.packageInformationData)
                )
            }
        }

        const addMoniker = (id: Id | undefined): void => {
            if (id !== undefined && !document.monikers.has(id)) {
                const moniker = assertDefined(id, 'moniker', this.monikerData)
                document.monikers.set(id, moniker)
                addPackageInformation(moniker.packageInformation)
            }
        }

        const addMonikerIds = (map: Map<Id, Map<Id, Set<Id>>>, ids: Id[], monikers: Id[]): void => {
            for (const moniker of monikers) {
                for (const id of ids) {
                    let m = map.get(currentDocumentId)
                    if (!m) {
                        m = new Map<Id, Set<Id>>()
                        map.set(currentDocumentId, m)
                    }

                    let n = m.get(id)
                    if (!n) {
                        n = new Set<Id>()
                        m.set(id, n)
                    }

                    n.add(moniker)
                }
            }
        }

        const addGenericResult = (
            name: string,
            datas: Map<Id, Map<Id, Id[]>>,
            monikerResults: Map<Id, Map<Id, Set<Id>>>,
            results: Map<Id, { documentPath: string; id: Id }[]>,
            id: Id | undefined,
            monikers: Id[]
        ): void => {
            if (!id) {
                return
            }

            const values = []
            for (const [documentId, ids] of assertDefined(id, name, datas)) {
                // Resolve the "document" field from the "item" edge. This will correlate
                // the referenced range identifier with the document in which it belongs.
                const documentPath = assertDefined(documentId, 'documentPath', this.documentPaths)

                for (const id of ids) {
                    values.push({ documentPath, id })
                }

                if (documentId === currentDocumentId) {
                    // If this is results for the current document, construct the data that
                    // will later be used to insert into the Defs table for this document.
                    addMonikerIds(monikerResults, ids, monikers)
                }
            }

            results.set(id, values)
        }

        const orderedRanges: (RangeData & { id: Id })[] = []
        for (const id of assertDefined(currentDocumentId, 'contains', this.containsData)) {
            const range = assertDefined(id, 'range', this.rangeData)
            this.canonicalizeRange(range)
            orderedRanges.push({ id, ...range })

            addHover(range.hoverResult)

            for (const id of range.monikers) {
                addMoniker(id)
            }

            addGenericResult(
                'definitionResult',
                this.definitionData,
                this.definitions,
                document.definitionResults,
                range.definitionResult,
                range.monikers
            )

            addGenericResult(
                'referenceResult',
                this.referenceData,
                this.references,
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
     * @param range The range to canonicalize.
     */
    private canonicalizeRange(range: RangeData): void {
        let definitionResult: Id | undefined
        let referenceResult: Id | undefined
        let hoverResult: Id | undefined
        const monikers = new Set<Id>()

        let item: RangeData | ResultSetData | undefined = range
        while (item) {
            definitionResult = definitionResult === undefined ? item.definitionResult : definitionResult
            referenceResult = referenceResult === undefined ? item.referenceResult : referenceResult
            hoverResult = hoverResult === undefined ? item.hoverResult : hoverResult

            if (item.monikers.length > 0) {
                for (const mon of reachableMonikers(this.monikerSets, item.monikers[0])) {
                    if (assertDefined(mon, 'moniker', this.monikerData).kind !== MonikerKind.local) {
                        monikers.add(mon)
                    }
                }
            }

            if (item.next === undefined) {
                break
            }

            item = assertDefined(item.next, 'resultSet', this.resultSetData)
        }

        range.definitionResult = definitionResult
        range.referenceResult = referenceResult
        range.hoverResult = hoverResult
        range.monikers = Array.from(monikers)
        range.next = undefined // TODO - just remove this
    }
}

/**
 * Return the value of `id`, or throw an exception if it is undefined.
 *
 * @param id The identifier.
 */
function assertId(id: Id | undefined): Id {
    if (id) {
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
function assertDefined<T>(id: Id, name: string, ...maps: Map<Id, T | null>[]): T {
    for (const map of maps) {
        const value = map.get(id)
        if (value) {
            return value
        }
    }

    throw new Error(`Unknown ${name} '${id}'.`)
}

/**
 * Extract the version from a protocol `MetaData` object.
 *
 * @param meta The protocol object.
 */
function convertMetadata(meta: MetaData): { lsifVersion: string; sourcegraphVersion: string } {
    return {
        lsifVersion: meta.version,
        sourcegraphVersion: INTERNAL_LSIF_VERSION,
    }
}

/**
 * Convert a protocol `Moniker` object into a `MonikerData` object.
 *
 * @param moniker The moniker object.
 */
function convertMoniker(moniker: Moniker): MonikerData {
    return { kind: moniker.kind || MonikerKind.local, scheme: moniker.scheme, identifier: moniker.identifier }
}

/**
 * Convert a protocol `PackageInformation` object into a `PackgeInformationData` object.
 *
 * @param info The protocol object.
 */
function convertPackageInformation(info: PackageInformation): PackageInformationData {
    return { name: info.name, version: info.version || '$missing' }
}

/**
 * Convert a protocol `Range` object into a `RangeData` object.
 *
 * @param range The range object.
 */
function convertRange(range: Range): RangeData {
    return { ...flattenRange(range), monikers: [] }
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
    const importer = new LsifImporter(entityManager)

    let i = 0
    for await (const element of elements) {
        try {
            await importer.insert(element)
        } catch (e) {
            throw Object.assign(new Error(`Failed to process line:\n${i}`), { e })
        }

        i++
    }

    return await importer.finalize()
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

/**
 * Construct a flattened four-tuple of numbers from an LSP range.
 *
 * @param range The LSP range.
 */
function flattenRange(range: Range): FlattenedRange {
    return {
        startLine: range.start.line,
        startCharacter: range.start.character,
        endLine: range.end.line,
        endCharacter: range.end.character,
    }
}
