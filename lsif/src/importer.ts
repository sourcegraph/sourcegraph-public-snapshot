import { EntityManager } from 'typeorm'
import { DefModel, DocumentModel, MetaModel, RefModel } from './models'
import { encodeJSON } from './encoding'
import { TableInserter } from './inserter'
import {
    MonikerData,
    RangeData,
    ResultSetData,
    DefinitionResultData,
    ReferenceResultData,
    HoverData,
    PackageInformationData,
    DocumentData,
    FlattenedRange,
} from './entities'
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

/**
 * The internal version of our SQLite databases. We need to keep this in case
 * we add something that can't be done transparently; if we change how we process
 * something in the future we'll need to consider a number of previous version
 * while we update or re-process the already-uploaded data.
 */
const INTERNAL_LSIF_VERSION = '0.1.0'

/**
 * `DecoratedDocumentData` is a `DocumentData` instance with additional context
 * during the importing of an LSIF dump.
 */
export interface DecoratedDocumentData extends DocumentData {
    /**
     * `id` is the identifier of the document.
     */
    id: Id

    /**
     * `uri` is the URI of the document.
     */
    uri: string

    /**
     * `contains` is the running set of identifiers that have a contains edge
     * to this document in the LSIF dump.
     */
    contains: Id[]

    /**
     * `definitions` carries the data of definitionResult edges attached within
     * the document if there is a non-local moniker attached to it; otherwise,
     * the definition result data would be stored in `definitionResults` in the
     * superclass.
     */
    definitions: { data: DefinitionResultData; moniker: MonikerData }[]

    /**
     * `references` carries the data of referenceResult edges attached within
     * the document if there is a non-local moniker attached to it; otherwise,
     * the reference result data would be stored in `referenceResults` in the
     * superclass.
     */
    references: { data: ReferenceResultData; moniker: MonikerData }[]
}

/**
 * `HandlerMap` is a mapping from vertex or edge labels to the function that
 * can handle an object of that particular type during import.
 */
interface HandlerMap {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    [K: string]: (element: any) => Promise<void>
}

/**
 * `LsifImporter` processes an upload of an LSIF dump. This class receives the
 * parsed vertex or edge, line by line, from the caller, and adds it into a
 * new database file on disk. Once finalized, the database is ready for use
 * and relevant cross-repository metadata is returned to the caller, which
 * is used to populate the xrepo database.
 */
export class LsifImporter {
    // Handler vtables
    private vertexHandlerMap: HandlerMap = {}
    private edgeHandlerMap: HandlerMap = {}

    // Bulk database inserters
    private metaInserter: TableInserter<MetaModel, new () => MetaModel>
    private documentInserter: TableInserter<DocumentModel, new () => DocumentModel>
    private defInserter: TableInserter<DefModel, new () => DefModel>
    private refInserter: TableInserter<RefModel, new () => RefModel>

    // Vertex data
    private definitionDatas: Map<Id, Map<Id, DefinitionResultData>> = new Map()
    private documents: Map<Id, string> = new Map()
    private hoverDatas: Map<Id, HoverData> = new Map()
    private monikerDatas: Map<Id, MonikerData> = new Map()
    private packageInformationDatas: Map<Id, PackageInformationData> = new Map()
    private rangeDatas: Map<Id, RangeData> = new Map()
    private referenceDatas: Map<Id, Map<Id, ReferenceResultData>> = new Map()
    private resultSetDatas: Map<Id, ResultSetData> = new Map()

    /**
     * `projectRoot` is the root of all document URIs. This is extracted from
     * the metadata vertex at the beginning of processing.
     */
    private projectRoot = ''

    /**
     * `importedMonikers` is the set of exported moniker identifiers that have
     * package information attached.
     */
    private importedMonikers = new Set<Id>()

    /**
     * `exportedMonikers` is the set of exported moniker identifiers that have
     * package information attached.
     */
    private exportedMonikers = new Set<Id>()

    /**
     * `documentDatas` are decorated `DocumentData` objects that are created on
     * document begin events and are inserted into the databse on document end
     * events.
     */
    private documentDatas = new Map<Id, DecoratedDocumentData>()

    /**
     * `monikerSets` holds the relation from moniker to the set of monikers that
     * they are related to via nextMoniker edges. This relation is symmetric.
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
        const packageHashes: Set<string> = new Set()
        for (const id of this.exportedMonikers) {
            const source = assertDefined(id, 'moniker', this.monikerDatas)
            const packageInformationId = assertId(source.packageInformation)
            const packageInfo = assertDefined(packageInformationId, 'packageInformation', this.packageInformationDatas)
            packageHashes.add(
                JSON.stringify({
                    scheme: source.scheme,
                    name: packageInfo.name,
                    version: packageInfo.version,
                })
            )
        }

        return Array.from(packageHashes).map(value => JSON.parse(value) as Package)
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
     * can create relative URIs for documnets. Insert a row in the meta
     * table with the LSIF protocol version.
     *
     * @param vertex The metadata vertex.
     */
    private async handleMetaData(vertex: MetaData): Promise<void> {
        this.projectRoot = vertex.projectRoot
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
    private handleHover = this.setById(this.hoverDatas, (e: HoverResult) => e.result)
    private handleMoniker = this.setById(this.monikerDatas, convertMoniker)
    private handlePackageInformation = this.setById(this.packageInformationDatas, convertPackageInformation)
    private handleRange = this.setById(this.rangeDatas, (e: Range) => ({ ...e, monikers: [] }))
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
            const defaultValue = { values: [] }
            this.handleGenericItemEdge(edge, 'definitionResult', this.definitionDatas, defaultValue, 'values')
        }

        if (edge.property === ItemEdgeProperties.definitions) {
            const defaultValue = { definitions: [], references: [] }
            this.handleGenericItemEdge(edge, 'referenceResult', this.referenceDatas, defaultValue, 'definitions')
        }

        if (edge.property === ItemEdgeProperties.references) {
            const defaultValue = { definitions: [], references: [] }
            this.handleGenericItemEdge(edge, 'referenceResult', this.referenceDatas, defaultValue, 'references')
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
        const uri = assertDefined(event.data, 'document', this.documents)

        this.documentDatas.set(event.data, {
            id: event.data,
            uri: uri.slice(this.projectRoot.length + 1),
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
        await this.documentInserter.insert({ uri: document.uri, value: await encodeJSON(document) })

        // Insert all related definitions
        for (const { data, moniker } of document.definitions) {
            for (const range of flattenRanges(document, data.values)) {
                await this.defInserter.insert({
                    scheme: moniker.scheme,
                    identifier: moniker.identifier,
                    documentUri: document.uri,
                    ...range,
                })
            }
        }

        // Insert all related references
        for (const { data, moniker } of document.references) {
            for (const ids of [data.definitions, data.references]) {
                for (const range of flattenRanges(document, ids)) {
                    await this.refInserter.insert({
                        scheme: moniker.scheme,
                        identifier: moniker.identifier,
                        documentUri: document.uri,
                        ...range,
                    })
                }
            }
        }
    }

    //
    // Helper Functions

    /**
     * Creates a function that takes an `element`, then correlates that
     * element's  identifier with the result from `factory` in `map`.
     *
     * @param map The map to populate.
     * @param factory The function that produces a value from `element`.
     */
    private setById<K extends { id: Id }, V>(map: Map<Id, V>, factory: (element: K) => V): (element: K) => void {
        return (element: K) => map.set(element.id, factory(element))
    }

    /**
     * Adds data to a nested array within the two-tier `map`. Let `outV`
     * and `inVs` be the source and destinations of `edge`, such that
     * `map` is indexed in the outer level by `outV` and indexed in the
     * inner level by document. This method adds the destination edges
     * to `map[outV][document][field]`, and creates any data structure
     * on th epath that has not yet been constructed.
     *
     * @param edge The edge.
     * @param name The type of map (used for exception message).
     * @param map The map to populate.
     * @param defaultValue The value to use if inner map is not populated.
     * @param field The field containing the target array.
     */
    private handleGenericItemEdge<T extends { [K in F]: Id[] }, F extends string>(
        edge: item,
        name: string,
        map: Map<Id, Map<Id, T>>,
        defaultValue: T,
        field: F
    ): void {
        const innerMap = assertDefined(edge.outV, name, map)
        let data = innerMap.get(edge.document)
        if (!data) {
            data = defaultValue
            innerMap.set(edge.document, data)
        }

        // TODO - use concat instead
        data[field].push(...edge.inVs)
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
        orderedRanges.sort((a, b) => a.start.line - b.start.line || a.start.character - b.start.character)

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
        documentArray: { data: T; moniker: MonikerData }[],
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
            documentArray.push({ moniker, data })
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
 * Convert a set of range identifiers into flattened range objects. This requires
 * the document's `ranges` and `orderedRanges` fields to be completely populated.
 *k
 * @param document The document object.
 * @param ids The list of ids.
 */
function flattenRanges(document: DecoratedDocumentData, ids: Id[]): FlattenedRange[] {
    const ranges = []
    for (const id of ids) {
        const rangeIndex = document.ranges.get(id)
        if (rangeIndex === undefined) {
            continue
        }

        const range = document.orderedRanges[rangeIndex]
        ranges.push({
            startLine: range.start.line,
            startCharacter: range.start.character,
            endLine: range.end.line,
            endCharacter: range.end.character,
        })
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
