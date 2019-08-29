import { EntityManager } from 'typeorm'
import { isEqual, uniqWith } from 'lodash'
import { DefModel, DocumentModel, MetaModel, RefModel } from './models'
import RelateUrl from 'relateurl'
import { encodeJSON } from './encoding'
import { TableInserter } from './inserter'
import { MonikerData, RangeData, ResultSetData, PackageInformationData, DocumentData, FlattenedRange } from './entities'
import {
    Id,
    VertexLabels,
    EdgeLabels,
    Vertex,
    Edge,
    MonikerKind,
    ItemEdgeProperties,
    Document,
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
    contains,
    HoverResult,
    EventKind,
    EventScope,
    MetaData,
    ElementTypes,
    Moniker,
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
 * A wrapper around `DocumentData` with additional context required during the
 * importing of an LSIF dump.
 */
export interface DecoratedDocumentData extends DocumentData {
    // The identifier of the document.
    id: Id

    // The root-relative path of the document.
    path: string

    /**
     * The running set of identifiers that have a contains edge to this document
     * in the LSIF dump.
     */
    contains: Id[]

    /**
     * A field that carries the data of definitionResult edges attached within
     * the document if there is a non-local moniker attached to it; otherwise,
     * the definition result data would be stored in `definitionResults` in the
     * superclass.
     */
    definitions: { ids: Id[]; moniker: MonikerData }[]

    /**
     * A field that carries the data of referenceResult edges attached within
     * the document if there is a non-local moniker attached to it; otherwise,
     * the reference result data would be stored in `referenceResults` in the
     * superclass.
     */
    references: { ids: Id[]; moniker: MonikerData }[]
}

/**
 * A mapping from vertex or edge labels to the function that can handle an object
 * of that particular type during import.
 */
interface HandlerMap {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    [K: string]: (element: any) => Promise<void>
}

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
    // Handler vtables
    private vertexHandlerMap: HandlerMap = {}
    private edgeHandlerMap: HandlerMap = {}

    // Bulk database inserters
    private metaInserter: TableInserter<MetaModel, new () => MetaModel>
    private documentInserter: TableInserter<DocumentModel, new () => DocumentModel>
    private defInserter: TableInserter<DefModel, new () => DefModel>
    private refInserter: TableInserter<RefModel, new () => RefModel>

    // Vertex data
    private definitionDatas: Map<Id, Map<Id, Id[]>> = new Map()
    private documents: Map<Id, string> = new Map()
    private hoverDatas: Map<Id, string> = new Map()
    private monikerDatas: Map<Id, MonikerData> = new Map()
    private packageInformationDatas: Map<Id, PackageInformationData> = new Map()
    private rangeDatas: Map<Id, RangeData> = new Map()
    private referenceDatas: Map<Id, Map<Id, Id[]>> = new Map()
    private resultSetDatas: Map<Id, ResultSetData> = new Map()

    /**
     * The root of all document URIs. This is extracted from the metadata vertex at
     * the beginning of processing.
     */
    private projectRoot?: URL

    // The set of exported moniker identifiers that have package information attached.
    private importedMonikers = new Set<Id>()

    // The set of exported moniker identifiers that have package information attached.
    private exportedMonikers = new Set<Id>()

    /**
     * A map of decorated `DocumentData` objects that are created on document begin events
     * and are inserted into the databse on document end events.
     */
    private documentDatas = new Map<Id, DecoratedDocumentData>()

    /**
     * A mapping for the relation from moniker to the set of monikers that they are related
     * to via nextMoniker edges. This relation is symmetric (if `a` is in `MonikerSets[b]`,
     * then `b` is in `monikerSets[a]`).
     */
    private monikerSets = new Map<Id, Set<Id>>()

    /**
     * Create a new `LsifImporter` with the given entity manager.
     *
     * @param entityManager A transactional SQLite entity manager.
     */
    constructor(private entityManager: EntityManager) {
        // Convert f into an async function
        const wrap = <T>(f: (element: T) => void) => (element: T) => Promise.resolve(f(element))

        // Register vertex handlers
        this.vertexHandlerMap[VertexLabels.definitionResult] = wrap(e => this.handleDefinitionResult(e))
        this.vertexHandlerMap[VertexLabels.document] = wrap(e => this.handleDocument(e))
        this.vertexHandlerMap[VertexLabels.event] = e => this.handleEvent(e)
        this.vertexHandlerMap[VertexLabels.hoverResult] = wrap(e => this.handleHover(e))
        this.vertexHandlerMap[VertexLabels.metaData] = e => this.handleMetaData(e)
        this.vertexHandlerMap[VertexLabels.moniker] = wrap(e => this.handleMoniker(e))
        this.vertexHandlerMap[VertexLabels.packageInformation] = wrap(e => this.handlePackageInformation(e))
        this.vertexHandlerMap[VertexLabels.range] = wrap(e => this.handleRange(e))
        this.vertexHandlerMap[VertexLabels.referenceResult] = wrap(e => this.handleReferenceResult(e))
        this.vertexHandlerMap[VertexLabels.resultSet] = wrap(e => this.handleResultSet(e))

        // Register edge handlers
        this.edgeHandlerMap[EdgeLabels.contains] = wrap(e => this.handleContains(e))
        this.edgeHandlerMap[EdgeLabels.item] = wrap(e => this.handleItemEdge(e))
        this.edgeHandlerMap[EdgeLabels.moniker] = wrap(e => this.handleMonikerEdge(e))
        this.edgeHandlerMap[EdgeLabels.next] = wrap(e => this.handleNextEdge(e))
        this.edgeHandlerMap[EdgeLabels.nextMoniker] = wrap(e => this.handleNextMonikerEdge(e))
        this.edgeHandlerMap[EdgeLabels.packageInformation] = wrap(e => this.handlePackageInformationEdge(e))
        this.edgeHandlerMap[EdgeLabels.textDocument_definition] = wrap(e => this.handleDefinitionEdge(e))
        this.edgeHandlerMap[EdgeLabels.textDocument_hover] = wrap(e => this.handleHoverEdge(e))
        this.edgeHandlerMap[EdgeLabels.textDocument_references] = wrap(e => this.handleReferenceEdge(e))

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
        const handler =
            element.type === ElementTypes.vertex
                ? this.vertexHandlerMap[element.label]
                : this.edgeHandlerMap[element.label]

        if (handler) {
            await handler(element)
        }
    }

    /**
     * Ensure that any outstanding records are flushed to the database. Also
     * returns the set of packages provided by the project analyzed by this
     * LSIF dump as well as the symbols imported into the LSIF dump from
     * external packages.
     */
    public async finalize(): Promise<{ packages: Package[]; references: SymbolReferences[] }> {
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
            const source = assertDefined(id, 'moniker', this.monikerDatas)
            const packageInformationId = assertId(source.packageInformation)
            const packageInfo = assertDefined(packageInformationId, 'packageInformation', this.packageInformationDatas)
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
            const source = assertDefined(id, 'moniker', this.monikerDatas)
            const packageInformationId = assertId(source.packageInformation)
            const packageInfo = assertDefined(packageInformationId, 'packageInformation', this.packageInformationDatas)
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
     * can create relative paths for documnets. Insert a row in the meta
     * table with the LSIF protocol version.
     *
     * @param vertex The metadata vertex.
     */
    private async handleMetaData(vertex: MetaData): Promise<void> {
        this.projectRoot = new URL(vertex.projectRoot)
        await this.metaInserter.insert(convertMetadata(vertex))
    }

    /**
     * Delegate document-scoped begin and end events.
     *
     * @param vertex The event vertex.
     */
    private async handleEvent(vertex: Event): Promise<void> {
        if (vertex.scope === EventScope.document && vertex.kind === EventKind.begin) {
            this.handleDocumentBegin(vertex as DocumentEvent)
        }

        if (vertex.scope === EventScope.document && vertex.kind === EventKind.end) {
            await this.handleDocumentEnd(vertex as DocumentEvent)
        }
    }

    // The remaining vertex handlers stash data into an appropriate map. This data
    // may be retrieved when an edge that references it is seen, or when a document
    // is finalized.

    private handleDefinitionResult = this.setById(this.definitionDatas, () => new Map())
    private handleDocument = this.setById(this.documents, (e: Document) => e.uri)
    private handleHover = this.setById(this.hoverDatas, (e: HoverResult) => normalizeHover(e.result))
    private handleMoniker = this.setById(this.monikerDatas, convertMoniker)
    private handlePackageInformation = this.setById(this.packageInformationDatas, convertPackageInformation)
    private handleRange = this.setById(this.rangeDatas, (e: Range) => convertRange(e))
    private handleReferenceResult = this.setById(this.referenceDatas, () => new Map())
    private handleResultSet = this.setById(this.resultSetDatas, () => ({ monikers: [] }))

    //
    // Edge Handlers

    /**
     * Add range data ids into the document in which they are contained. Ensures
     * all referenced vertices are defined.
     *
     * @param edge The contains edge.
     */
    private handleContains(edge: contains): void {
        if (this.documentDatas.has(edge.outV)) {
            const source = assertDefined(edge.outV, 'document', this.documentDatas)
            mapAssertDefined(edge.inVs, 'range', this.rangeDatas)
            source.contains = source.contains.concat(edge.inVs)
        }
    }

    /**
     * Update definition and reference fields from an item edge. Ensures all
     * referenced vertices are defined.
     *
     * @param edge The item edge.
     */
    private handleItemEdge(edge: item): void {
        if (edge.property === undefined) {
            this.handleGenericItemEdge(edge, 'definitionResult', this.definitionDatas)
        }

        if (edge.property === ItemEdgeProperties.definitions) {
            this.handleGenericItemEdge(edge, 'referenceResult', this.referenceDatas)
        }

        if (edge.property === ItemEdgeProperties.references) {
            this.handleGenericItemEdge(edge, 'referenceResult', this.referenceDatas)
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
            this.rangeDatas,
            this.resultSetDatas
        )
        assertDefined(edge.inV, 'moniker', this.monikerDatas)
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
            this.rangeDatas,
            this.resultSetDatas
        )
        assertDefined(edge.inV, 'resultSet', this.resultSetDatas)
        outV.next = edge.inV
    }

    /**
     * Correlates monikers together so that when one moniker is queried, each correlated moniker
     * is also returned as a strongly connected set. Ensures all referenced vertices are defined.
     *
     * @param edge The nextMoniker edge.
     */
    private handleNextMonikerEdge(edge: nextMoniker): void {
        assertDefined(edge.inV, 'moniker', this.monikerDatas)
        assertDefined(edge.outV, 'moniker', this.monikerDatas)
        this.correlateMonikers(edge.inV, edge.outV) // Forward direction
        this.correlateMonikers(edge.outV, edge.inV) // Backwards direction
    }

    /**
     * Sets the package information of the specified moniker. If the moniker is an export moniker,
     * then the package information will also be returned as an exported pacakge by the `finalize`
     * method. Ensures all referenced vertices are defined.
     *
     * @param edge The packageInformation edge.
     */
    private handlePackageInformationEdge(edge: packageInformation): void {
        const source = assertDefined(edge.outV, 'moniker', this.monikerDatas)
        assertDefined(edge.inV, 'packageInformation', this.packageInformationDatas)
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
            this.rangeDatas,
            this.resultSetDatas
        )
        assertDefined(edge.inV, 'definitionResult', this.definitionDatas)
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
            this.rangeDatas,
            this.resultSetDatas
        )
        assertDefined(edge.inV, 'hoverResult', this.hoverDatas)
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
            this.rangeDatas,
            this.resultSetDatas
        )
        assertDefined(edge.inV, 'referenceResult', this.referenceDatas)
        outV.referenceResult = edge.inV
    }

    //
    // Event Handlers

    /**
     * Initialize a blank document which will be fully populated on the invocation of
     * `handleDocumentEnd`. This document is created now so that we can stash the ids
     * of ranges refered to by `contains` edges we see before the document end event
     * occurs.
     *
     * @param event The document begin event.
     */
    private handleDocumentBegin(event: DocumentEvent): void {
        if (!this.projectRoot) {
            throw new Error('No project root has been defined.')
        }

        const uri = assertDefined(event.data, 'document', this.documents)

        const path = RelateUrl.relate(this.projectRoot.href + '/', new URL(uri).href, {
            defaultPorts: {},
            output: RelateUrl.PATH_RELATIVE,
            removeRootTrailingSlash: false,
        })

        this.documentDatas.set(event.data, {
            id: event.data,
            path,
            contains: [],
            definitions: [],
            references: [],
            ranges: new Map(),
            orderedRanges: [],
            resultSets: new Map(),
            definitionResults: new Map(),
            referenceResults: new Map(),
            hovers: new Map(),
            monikers: new Map(),
            packageInformation: new Map(),
        })
    }

    /**
     * Finalize the document by correlating and compressing any data reachable from a
     * range that it contains. This document, as well as its definitions and references,
     * will be submitted to the database for insertion.
     *
     * @param event The document end event.
     */
    private async handleDocumentEnd(event: DocumentEvent): Promise<void> {
        const document = assertDefined(event.data, 'document', this.documentDatas)

        // Finalize document
        await this.finalizeDocument(document)

        // Insert document record
        await this.documentInserter.insert({
            path: document.path,
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

        // Insert all related definitions
        for (const { ids, moniker } of document.definitions) {
            for (const range of lookupRanges(document, ids)) {
                await this.defInserter.insert({
                    scheme: moniker.scheme,
                    identifier: moniker.identifier,
                    documentPath: document.path,
                    ...range,
                })
            }
        }

        // Insert all related references
        for (const { ids, moniker } of document.references) {
            for (const range of lookupRanges(document, ids)) {
                await this.refInserter.insert({
                    scheme: moniker.scheme,
                    identifier: moniker.identifier,
                    documentPath: document.path,
                    ...range,
                })
            }
        }
    }

    //
    // Helper Functions

    /**
     * Creates a function that takes an `element`, then correlates that
     * element's identifier with the result from `factory` in `map`.
     *
     * @param map The map to populate.
     * @param factory The function that produces a value from `element`.
     */
    private setById<K extends { id: Id }, V>(map: Map<Id, V>, factory: (element: K) => V): (element: K) => void {
        return (element: K) => map.set(element.id, factory(element))
    }

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
     * Populate a document object (whose only populated value should be its `contains` array).
     * Each range that is contained in this document will be added to this object, as well as
     * any item reachable from that range. This lazily populates the document with the minimal
     * data, and keeps it self-contained within the document so that multiple queries are not
     * needed when asking about (non-xrepo) LSIF data.
     *
     * @param document The document object.
     */
    private async finalizeDocument(document: DecoratedDocumentData): Promise<void> {
        const orderedRanges: (RangeData & { id: Id })[] = []
        for (const id of document.contains) {
            const range = assertDefined(id, 'range', this.rangeDatas)
            orderedRanges.push({ id, ...range })
            await this.attachItemToDocument(document, id, range)
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
    }

    /**
     * Moves the data reachable from the given range or result set into the
     * given document. This walks the edges of next/item edges as seen in
     * one of the handler functions above.
     *
     * @param document The document object.
     * @param id The identifier of the range or result set.
     * @param item The range or result set.
     */
    private async attachItemToDocument(
        document: DecoratedDocumentData,
        id: Id,
        item: RangeData | ResultSetData
    ): Promise<void> {
        // Find monikers for an item and add them to the item and document.
        // This will also add any package information attached to a moniker
        // to the document.
        const monikers = this.attachItemMonikersToDocument(document, id, item)

        // Add result set to document, if it doesn't exist
        if (item.next && !document.resultSets.has(item.next)) {
            const resultSet = assertDefined(item.next, 'resultSet', this.resultSetDatas)
            document.resultSets.set(item.next, resultSet)
            await this.attachItemToDocument(document, item.next, resultSet)
        }

        // Add hover to document, if it doesn't exist
        if (item.hoverResult && !document.hovers.has(item.hoverResult)) {
            const hoverResult = assertDefined(item.hoverResult, 'hoverResult', this.hoverDatas)
            document.hovers.set(item.hoverResult, hoverResult)
        }

        // Attach definition and reference results results to the document.
        // This atatches some denormalized data on the `WrappedDocumentData`
        // object which will also be used to populate the defs and refs
        // tables.

        this.attachResultsToDocument(
            'definitionResult',
            this.definitionDatas,
            document,
            document.definitions,
            document.definitionResults,
            item.definitionResult,
            monikers
        )

        this.attachResultsToDocument(
            'referenceResult',
            this.referenceDatas,
            document,
            document.references,
            document.referenceResults,
            item.referenceResult,
            monikers
        )
    }

    /**
     * Find all monikers reachable from the given range or result set, and
     * add them to the item, and the document. If pacakge information is
     * also attached, it is also atatched to the document.
     *
     * @param document The document object.
     * @param id The identifier of the range or result set.
     * @param item The range or result set.
     */
    private attachItemMonikersToDocument(
        document: DecoratedDocumentData,
        id: Id,
        item: RangeData | ResultSetData
    ): MonikerData[] {
        if (item.monikers.length === 0) {
            return []
        }

        const monikers = []
        for (const id of reachableMonikers(this.monikerSets, item.monikers[0])) {
            const moniker = assertDefined(id, 'moniker', this.monikerDatas)
            monikers.push(moniker)
            item.monikers.push(id)
            document.monikers.set(id, moniker)

            if (moniker.packageInformation) {
                const packageInformation = assertDefined(
                    moniker.packageInformation,
                    'packageInformation',
                    this.packageInformationDatas
                )

                document.packageInformation.set(moniker.packageInformation, packageInformation)
            }
        }

        return monikers
    }

    /**
     * Attach definition or reference results to the document (with respect to a
     * particular item). This method retrieves the result data from the `sourceMap`
     * and assigns it to either the `documentArray` or the `documentMap`, depending
     * on whether or not the list of monikers has contains a non-local item. This
     * method will early-out if the given identifier is undefined.
     *
     * @param name The type of element (used for exception message).
     * @param sourceMap The map in `this` that holds the source data.
     * @param document The document object.
     * @param documentArray The list object in `WrappedDocumentData` to modify.
     * @param documentMap The set object in `DocumentData` to modify.
     * @param id The identifier of the item's result.
     * @param monikers The set of monikers attached to the item.
     */
    private attachResultsToDocument<T>(
        name: string,
        sourceMap: Map<Id, Map<Id, T>>,
        document: DecoratedDocumentData,
        documentArray: { ids: T; moniker: MonikerData }[],
        documentMap: Map<Id, T>,
        id: Id | undefined,
        monikers: MonikerData[]
    ): void {
        if (!id) {
            return
        }

        const innerMap = assertDefined(id, name, sourceMap)
        const data = innerMap.get(document.id)
        if (!data) {
            return
        }

        const nonlocalMonikers = monikers.filter(m => m.kind !== MonikerKind.local)
        for (const moniker of nonlocalMonikers) {
            documentArray.push({ ids: data, moniker })
        }

        if (nonlocalMonikers.length === 0) {
            documentMap.set(id, data)
        }
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
 * Call `assertDefined` over the given ids.
 *
 * @param ids The ids to map over.
 * @param name The type of element (used for exception message).
 * @param maps The set of maps to query.
 */
function mapAssertDefined<T>(ids: Id[], name: string, ...maps: Map<Id, T | null>[]): T[] {
    return ids.map(id => assertDefined(id, name, ...maps))
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
 * Convert a set of range identifers into the flattened range objects stored by
 * identifier in the given document. This requires that the document's `ranges`
 * and `orderedRanges` fields to be completely populated.
 *
 * @param document The document object.
 * @param ids The list of ids.
 */
function lookupRanges(document: DecoratedDocumentData, ids: Id[]): FlattenedRange[] {
    const ranges = []
    for (const id of ids) {
        const rangeIndex = document.ranges.get(id)
        if (rangeIndex === undefined) {
            continue
        }

        const range = document.orderedRanges[rangeIndex]
        ranges.push(range)
    }

    return ranges
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
function reachableMonikers(monikerSets: Map<Id, Set<Id>>, id: Id): Set<Id> {
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
 * Handle the lifecycle of an importer. Creates an `LsifImporter`. This will create
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
function normalizeHover(hover: Hover): string {
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
