import * as readline from 'readline'
import * as uuid from 'uuid'
import { DocumentModel, DefModel, MetaModel, RefModel } from './models'
import { Connection } from 'typeorm'
import { connectionCache } from './cache'
import { fs } from 'mz'
import { hashAndEncodeJSON } from './encoding'
import {
    MonikerData,
    RangeData,
    ResultSetData,
    DefinitionResultData,
    ReferenceResultData,
    DocumentBlob,
    HoverData,
    PackageInformationData,
} from './entities'
import {
    Id,
    VertexLabels,
    EdgeLabels,
    Vertex,
    Edge,
    Uri,
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
} from 'lsif-protocol'
import { Readable } from 'stream'

const INTERNAL_LSIF_VERSION = '0.1.0'

export interface XrepoSymbols {
    exported: Set<Package>
    imported: Set<SymbolReference>
}

export interface Package {
    scheme: string
    name: string
    version: string
}

export interface SymbolReference {
    scheme: string
    name: string
    version: string
    identifier: string
}

interface DocumentData {
    id: Id
    uri: Uri
}

interface FlattenedRange {
    startLine: number
    startCharacter: number
    endLine: number
    endCharacter: number
}

interface ExternalDefinition {
    scheme: string
    indentifier: string
    ranges: FlattenedRange[]
}

interface ExternalReference {
    scheme: string
    indentifier: string
    definitions: FlattenedRange[]
    references: FlattenedRange[]
}

interface DocumentDatabaseData {
    hash: string
    encoded: string
    definitions: ExternalDefinition[]
    references: ExternalReference[]
}

interface WrappedDocumentBlob extends DocumentBlob {
    id: Id
    uri: string
    contains: Id[]
    definitions: MonikerScopedResultData<DefinitionResultData>[] // TODO - get rid of this class definition as well
    references: MonikerScopedResultData<ReferenceResultData>[] // TODO - get rid of this class definition as well
}

interface MonikerScopedResultData<T> {
    moniker: MonikerData
    data: T
}

export async function convertToBlob(input: Readable, outFile: string): Promise<XrepoSymbols> {
    try {
        await fs.unlink(outFile)
    } catch (err) {
        // TODO
    }

    return await connectionCache.withConnection(
        outFile,
        [DefModel, DocumentModel, MetaModel, RefModel],
        async connection => {
            const blobStore = new BlobStore(connection)
            const lines = readline.createInterface({ input })

            for await (const line of lines) {
                await blobStore.insert(parseLine(line))
            }

            return {
                exported: blobStore.exportedPackages,
                imported: blobStore.importedSymbols,
            }
        }
    )
}

function parseLine(line: string): Vertex | Edge {
    try {
        return JSON.parse(line)
    } catch (err) {
        throw new Error(`Parsing failed for line:\n${line}`)
    }
}

interface HandlerMap {
    [K: string]: (element: any) => Promise<void>
}

class BlobStore {
    // Handler vtables
    private vertexHandlerMap: HandlerMap = {}
    private edgeHandlerMap: HandlerMap = {}

    // Vertex data
    private definitionDatas: Map<Id, Map<Id, DefinitionResultData>> = new Map()
    private documents: Map<Id, DocumentData> = new Map()
    private hoverDatas: Map<Id, HoverData> = new Map()
    private monikerDatas: Map<Id, MonikerData> = new Map()
    private packageInformationDatas: Map<Id, PackageInformationData> = new Map()
    private rangeDatas: Map<Id, RangeData> = new Map()
    private referenceDatas: Map<Id, Map<Id, ReferenceResultData>> = new Map()
    private resultSetDatas: Map<Id, ResultSetData> = new Map()

    // Edge data
    private monikerSets: Map<Id, Set<Id>> = new Map()
    private monikerAttachments: Map<Id, Id> = new Map()

    // Documents under construction
    private documentDatas: Map<Id, WrappedDocumentBlob | null> = new Map()

    // TODO
    private projectRoot: string | undefined

    // TODO - expose via method?
    public exportedPackages: Set<Package> = new Set()
    public importedSymbols: Set<SymbolReference> = new Set()

    private insertedBlobs: Set<string> = new Set()
    // private insertedHovers: Set<string> = new Set()

    constructor(private connection: Connection) {
        const wrap = <T>(f: (element: T) => void) => (element: T) => Promise.resolve(f(element))

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

        this.edgeHandlerMap[EdgeLabels.contains] = wrap(e => this.handleContains(e))
        this.edgeHandlerMap[EdgeLabels.item] = wrap(e => this.handleItemEdge(e))
        this.edgeHandlerMap[EdgeLabels.moniker] = wrap(e => this.handleMonikerEdge(e))
        this.edgeHandlerMap[EdgeLabels.next] = wrap(e => this.handleNextEdge(e))
        this.edgeHandlerMap[EdgeLabels.nextMoniker] = wrap(e => this.handleNextMonikerEdge(e))
        this.edgeHandlerMap[EdgeLabels.packageInformation] = wrap(e => this.handlePackageInformationEdge(e))
        this.edgeHandlerMap[EdgeLabels.textDocument_definition] = wrap(e => this.handleDefinitionEdge(e))
        this.edgeHandlerMap[EdgeLabels.textDocument_hover] = wrap(e => this.handleHoverEdge(e))
        this.edgeHandlerMap[EdgeLabels.textDocument_references] = wrap(e => this.handleReferenceEdge(e))
    }

    public async insert(element: Vertex | Edge): Promise<void> {
        const handler =
            element.type === ElementTypes.vertex
                ? this.vertexHandlerMap[element.label]
                : this.edgeHandlerMap[element.label]

        if (handler) {
            await handler(element)
        }
    }

    //
    // Vertex Handlers

    private async handleMetaData(vertex: MetaData): Promise<void> {
        this.projectRoot = vertex.projectRoot

        await this.connection.getRepository(MetaModel).save({
            lsifVersion: vertex.version,
            sourcegraphVersion: INTERNAL_LSIF_VERSION,
        })
    }

    private async handleEvent(vertex: Event): Promise<void> {
        if (vertex.scope === EventScope.document && vertex.kind === EventKind.begin) {
            this.handleDocumentBegin(vertex as DocumentEvent)
        }

        if (vertex.scope === EventScope.document && vertex.kind === EventKind.end) {
            await this.handleDocumentEnd(vertex as DocumentEvent)
        }
    }

    private handleDefinitionResult = this.setById(this.definitionDatas, _ => new Map())
    private handleDocument = this.setById(this.documents, (e: Document) => ({ id: e.id, uri: e.uri }))
    private handleHover = this.setById(this.hoverDatas, (e: HoverResult) => e.result)
    private handleMoniker = this.setById(this.monikerDatas, e => e)
    private handlePackageInformation = this.setById(this.packageInformationDatas, convertPackageInformation)
    private handleRange = this.setById(this.rangeDatas, (e: Range) => ({ ...e, monikers: [] }))
    private handleReferenceResult = this.setById(this.referenceDatas, _ => new Map())
    private handleResultSet = this.setById(this.resultSetDatas, _ => ({}))

    private setById<K extends { id: Id }, V>(map: Map<Id, V>, factory: (element: K) => V): (element: K) => void {
        return (element: K) => map.set(element.id, factory(element))
    }

    //
    // Edge Handlers

    private handleContains(edge: contains): void {
        // Don't track projects
        if (!this.documentDatas.has(edge.outV)) {
            return
        }

        const source = assertDefined(edge.outV, this.documentDatas)
        for (const id of edge.inVs) {
            assertDefined(id, this.rangeDatas)
        }

        source.contains.push(...edge.inVs)
    }

    private handleItemEdge(edge: item): void {
        if (edge.property === undefined) {
            this.handleGenericItemEdge(edge, this.definitionDatas, { values: [] }, 'values')
        }

        if (edge.property === ItemEdgeProperties.definitions) {
            this.handleGenericItemEdge(edge, this.referenceDatas, { definitions: [], references: [] }, 'definitions')
        }

        if (edge.property === ItemEdgeProperties.references) {
            this.handleGenericItemEdge(edge, this.referenceDatas, { definitions: [], references: [] }, 'references')
        }
    }

    // TODO - move this
    private handleGenericItemEdge<T extends { [K in F]: Id[] }, F extends string>(
        edge: item,
        map: Map<Id, Map<Id, T>>,
        defaultValue: T,
        field: F
    ): void {
        const innerMap = assertDefined(edge.outV, map)
        let data = innerMap.get(edge.document)
        if (!data) {
            data = defaultValue
            innerMap.set(edge.document, data)
        }

        data[field].push(...edge.inVs)
    }

    private handleMonikerEdge(edge: moniker): void {
        assertDefined<RangeData | ResultSetData>(edge.outV, this.rangeDatas, this.resultSetDatas)
        assertDefined(edge.inV, this.monikerDatas)
        this.monikerAttachments.set(edge.outV, edge.inV)
        this.updateMonikerSets([edge.inV])
    }

    private handleNextEdge(edge: next): void {
        const outV = assertDefined<RangeData | ResultSetData>(edge.outV, this.rangeDatas, this.resultSetDatas)
        assertDefined(edge.inV, this.resultSetDatas)
        outV.next = edge.inV
    }

    private handleNextMonikerEdge(edge: nextMoniker): void {
        assertDefined(edge.inV, this.monikerDatas)
        assertDefined(edge.outV, this.monikerDatas)
        this.updateMonikerSets([edge.inV, edge.outV])
    }

    private handlePackageInformationEdge(edge: packageInformation): void {
        const source: MonikerData = assertDefined(edge.outV, this.monikerDatas)
        const packageInfo = assertDefined(edge.inV, this.packageInformationDatas)
        source.packageInformation = edge.inV

        if (source.kind === 'export') {
            this.exportedPackages.add({ scheme: source.scheme, name: packageInfo.name, version: packageInfo.version })
        }
    }

    private handleDefinitionEdge(edge: textDocument_definition): void {
        const outV = assertDefined<RangeData | ResultSetData>(edge.outV, this.rangeDatas, this.resultSetDatas)
        this.ensureMoniker(outV)
        assertDefined(edge.inV, this.definitionDatas)
        outV.definitionResult = edge.inV
    }

    private handleHoverEdge(edge: textDocument_hover): void {
        const outV = assertDefined<RangeData | ResultSetData>(edge.outV, this.rangeDatas, this.resultSetDatas)
        this.ensureMoniker(outV)
        assertDefined(edge.inV, this.hoverDatas)
        outV.hoverResult = edge.inV
    }

    private handleReferenceEdge(edge: textDocument_references): void {
        const outV = assertDefined<RangeData | ResultSetData>(edge.outV, this.rangeDatas, this.resultSetDatas)
        this.ensureMoniker(outV)
        assertDefined(edge.inV, this.referenceDatas)
        outV.referenceResult = edge.inV
    }

    //
    // Event Handlers

    private handleDocumentBegin(event: DocumentEvent): void {
        this.getOrCreateDocumentData(assertDefined(event.data, this.documents))
        this.documents.delete(event.data)
    }

    private async handleDocumentEnd(event: DocumentEvent): Promise<void> {
        console.log('finalizing document...')

        const documentData = assertDefined(event.data, this.documentDatas)
        const data = await this.finalizeBlob(documentData)
        if (this.insertedBlobs.has(data.hash)) {
            return
        }

        this.insertedBlobs.add(data.hash)
        await this.connection
            .getRepository(DocumentModel)
            .save({ hash: data.hash, value: data.encoded, uri: documentData.uri })

        const defs = []
        for (const definition of data.definitions) {
            for (const range of definition.ranges) {
                defs.push({
                    scheme: definition.scheme,
                    identifier: definition.indentifier,
                    documentUri: documentData.uri,
                    ...range,
                })
            }
        }

        const refs = []
        for (const reference of data.references) {
            for (const arr of [reference.definitions, reference.references]) {
                for (const range of arr || []) {
                    refs.push({
                        scheme: reference.scheme,
                        identifier: reference.indentifier,
                        documentUri: documentData.uri,
                        ...range,
                    })
                }
            }
        }

        const maxBatchSize = 124

        //
        // TODO - make chunk inserter a generic thing
        //

        const chunk = async <T>(items: T[], callback: (batch: T[]) => Promise<any>): Promise<void> => {
            while (items.length > maxBatchSize) {
                await callback(items.slice(0, maxBatchSize))
                items = items.slice(maxBatchSize)
            }

            if (items.length > 0) {
                await callback(items)
            }
        }

        await chunk(defs, batch =>
            this.connection
                .createQueryBuilder()
                .insert()
                .into(DefModel)
                .values(batch)
                .execute()
        )

        await chunk(refs, batch =>
            this.connection
                .createQueryBuilder()
                .insert()
                .into(RefModel)
                .values(batch)
                .execute()
        )
    }

    //
    // TODO - categorize

    private getOrCreateDocumentData(document: DocumentData): WrappedDocumentBlob {
        let result: WrappedDocumentBlob | undefined | null = this.documentDatas.get(document.id)
        if (result === null) {
            throw new Error(`The document ${document.uri} has already been processed`)
        }

        if (!this.projectRoot) {
            throw new Error('No project root.')
        }

        result = {
            id: document.id,
            uri: document.uri.slice(this.projectRoot.length + 1),
            contains: [],
            definitions: [],
            references: [],
            ranges: new Map(),
            resultSets: new Map(),
            definitionResults: new Map(),
            referenceResults: new Map(),
            hovers: new Map(),
            monikers: new Map(),
            packageInformation: new Map(),
        }

        this.documentDatas.set(document.id, result)
        return result
    }

    private updateMonikerSets(vals: Id[]): void {
        const combined = new Set<Id>()
        for (const val of vals) {
            combined.add(val)

            const otherSet = this.monikerSets.get(val)
            if (otherSet) {
                for (const v of otherSet) {
                    combined.add(v)
                }
            }
        }

        for (const val of combined) {
            this.monikerSets.set(val, combined)
        }
    }

    private ensureMoniker(data: RangeData | ResultSetData): void {
        if (!data.monikers) {
            const id = uuid.v4()
            const monikerData: MonikerData = { id, kind: MonikerKind.local, scheme: '$synthetic', identifier: id }
            data.monikers = [monikerData.identifier]
            this.monikerDatas.set(monikerData.identifier, monikerData)
        }
    }

    //
    // Blob things

    private async addReferencedDataToBlob(blob: WrappedDocumentBlob, item: RangeData | ResultSetData): Promise<void> {
        const monikers = []
        for (const itemMoniker of item.monikers || []) {
            const moniker = assertDefined(itemMoniker, this.monikerDatas)
            blob.monikers.set(itemMoniker, moniker)
            monikers.push(moniker)
        }

        // Many ranges can point to the same result set. Make sure we only travers once.
        if (item.next && !blob.resultSets.has(item.next)) {
            const resultSet = assertDefined(item.next, this.resultSetDatas)
            blob.resultSets.set(item.next, resultSet)
            await this.addReferencedDataToBlob(blob, resultSet)
        }

        if (item.hoverResult && !blob.hovers.has(item.hoverResult)) {
            const hoverResult = assertDefined(item.hoverResult, this.hoverDatas)
            blob.hovers.set(item.hoverResult, hoverResult)
        }

        if (item.definitionResult) {
            this.addDefinitionResultsToBlob(blob, item.definitionResult, monikers)
        }

        if (item.referenceResult) {
            this.addReferenceResultsToBlob(blob, item.referenceResult, monikers)
        }
    }

    private addDefinitionResultsToBlob(blob: WrappedDocumentBlob, definitionResult: Id, monikers: MonikerData[]): void {
        const map = assertDefined(definitionResult, this.definitionDatas)
        const definitions = map.get(blob.id)
        if (!definitions) {
            return
        }

        const nonlocalMonikers = monikers.filter(m => m.kind !== MonikerKind.local)
        for (const moniker of nonlocalMonikers) {
            blob.definitions.push({ moniker, data: definitions })
        }

        if (nonlocalMonikers.length === 0) {
            blob.definitionResults.set(definitionResult, definitions)
        }
    }

    private addReferenceResultsToBlob(blob: WrappedDocumentBlob, referenceResult: Id, monikers: MonikerData[]): void {
        const map = assertDefined(referenceResult, this.referenceDatas)
        const references = map.get(blob.id)
        if (!references) {
            return
        }

        const nonlocalMonikers = monikers.filter(m => m.kind !== MonikerKind.local)
        for (const moniker of nonlocalMonikers) {
            blob.references.push({ moniker, data: references })
        }

        if (nonlocalMonikers.length === 0) {
            blob.referenceResults.set(referenceResult, references)
        }
    }

    private async finalizeBlob(blob: WrappedDocumentBlob): Promise<DocumentDatabaseData> {
        for (const [key, value] of this.packageInformationDatas) {
            blob.packageInformation.set(key, value)
        }

        for (const data of this.monikerDatas.values()) {
            if (data.kind === 'import' && data.packageInformation) {
                const packageInformation = assertDefined(data.packageInformation, this.packageInformationDatas)

                this.importedSymbols.add({
                    scheme: data.scheme,
                    name: packageInformation.name,
                    version: packageInformation.version,
                    identifier: data.identifier,
                })
            }
        }

        for (const [key, value] of this.monikerAttachments.entries()) {
            const ids = assertDefined(value, this.monikerSets)
            for (const id of ids) {
                assertDefined(id, this.monikerDatas)
            }

            const source = assertDefined<RangeData | ResultSetData>(key, this.rangeDatas, this.resultSetDatas)
            source.monikers = Array.from(ids)
        }

        for (const id of blob.contains) {
            const range = assertDefined(id, this.rangeDatas)
            blob.ranges.set(id, range)
            await this.addReferencedDataToBlob(blob, range)
        }

        const definitions = []
        for (const definition of blob.definitions) {
            definitions.push({
                scheme: definition.moniker.scheme,
                indentifier: definition.moniker.identifier,
                ranges: flattenRanges(blob, definition.data.values),
            })
        }

        const references = []
        for (const reference of blob.references) {
            references.push({
                scheme: reference.moniker.scheme,
                indentifier: reference.moniker.identifier,
                definitions: flattenRanges(blob, reference.data.definitions),
                references: flattenRanges(blob, reference.data.references),
            })
        }

        return { ...(await hashAndEncodeJSON(blob)), definitions, references }
    }
}

function assertDefined<T>(id: Id, ...maps: Map<Id, T | null>[]): T {
    for (const map of maps) {
        const value = map.get(id)
        if (value === null) {
            // TODO - use a different tombstone value
            throw new Error(`Element ${id} has already been processed.`)
        }

        if (value) {
            return value
        }
    }

    throw new Error(`Element '${id}' is referenced before its definition.`)
}

function convertPackageInformation(e: PackageInformation): PackageInformationData {
    return e.version ? { name: e.name, version: e.version } : { name: e.name, version: '$missing' }
}

function flattenRanges(blob: WrappedDocumentBlob, ids: Id[]): FlattenedRange[] {
    const ranges = []
    for (const id of ids) {
        const range = blob.ranges.get(id)
        if (range) {
            ranges.push({
                startLine: range.start.line,
                startCharacter: range.start.character,
                endLine: range.end.line,
                endCharacter: range.end.character,
            })
        }
    }

    return ranges
}
