import { fs } from 'mz'
import * as lsp from 'vscode-languageserver-protocol'
import * as readline from 'readline'
import * as uuid from 'uuid'
import * as db from './database'
import * as entities from './entities'
import { hashAndEncodeJSON, hashJSON } from './encoding'
import * as protocol from 'lsif-protocol'
import { Connection } from 'typeorm'
import { connectionCache } from './cache'

const INTERNAL_LSIF_VERSION = '0.1.0'

export interface ExportedPackage {
    scheme: string
    name: string
    version: string
}

export interface ImportedSymbol {
    scheme: string
    name: string
    version: string
    identifier: string
}

namespace LiteralMap {
    export function create<T = any>(): db.LiteralMap<T> {
        return Object.create(null)
    }
}

namespace Monikers {
    export function isLocal(moniker: db.MonikerData): boolean {
        return moniker.kind === protocol.MonikerKind.local
    }
}

type InlineRange = [number, number, number, number]

interface ExternalDefinition {
    scheme: string
    indentifier: string
    ranges: InlineRange[]
}

interface ExternalReference {
    scheme: string
    indentifier: string
    definitions?: InlineRange[]
    references?: InlineRange[]
}

interface DocumentDatabaseData {
    hash: string
    encoded: string
    definitions?: ExternalDefinition[]
    references?: ExternalReference[]
}

interface MonikerScopedResultData<T> {
    moniker: db.MonikerData
    data: T
}

export async function convertToBlob(
    inFile: string, // TODO - make a stream instead
    outFile: string
): Promise<{ exportedPackages: Set<ExportedPackage>; importedSymbols: Set<ImportedSymbol> }> {
    try {
        await fs.unlink(outFile)
    } catch (err) {
        // TODO
    }

    return await connectionCache.withConnection(
        outFile,
        [entities.Blob, entities.Def, entities.Document, entities.Hover, entities.Meta, entities.Ref],
        async connection => {
            const blobStore = new BlobStore(connection)
            const lines = readline.createInterface({ input: fs.createReadStream(inFile, { encoding: 'utf8' }) })

            for await (const line of lines) {
                let element: protocol.Vertex | protocol.Edge
                try {
                    element = JSON.parse(line)
                } catch (err) {
                    throw new Error(`Parsing failed for line:\n${line}`)
                }

                await blobStore.insert(element)
            }

            return {
                exportedPackages: blobStore.exportedPackages,
                importedSymbols: blobStore.importedSymbols,
            }
        }
    )
}

class BlobStore {
    private vertexHandlerMap: { [K: string]: (element: any) => Promise<void> } = {}
    private edgeHandlerMap: { [K: string]: (element: any) => Promise<void> } = {}

    private definitionDatas: Map<protocol.Id, Map<protocol.Id, db.DefinitionResultData>> = new Map()
    private documents: Map<protocol.Id, protocol.Document> = new Map()
    public hoverDatas: Map<protocol.Id, lsp.Hover> = new Map() // TODO - visibility
    public monikerDatas: Map<protocol.Id, db.MonikerData> = new Map() // TODO - visibility
    private packageInformationDatas: Map<protocol.Id, protocol.PackageInformation> = new Map()
    private rangeDatas: Map<protocol.Id, protocol.Range> = new Map()
    private referenceDatas: Map<protocol.Id, Map<protocol.Id, db.ReferenceResultData>> = new Map()
    public resultSetDatas: Map<protocol.Id, db.ResultSetData> = new Map() // TODO - visibility

    // TODO
    private projectRoot: string | undefined

    public exportedPackages: Set<ExportedPackage> = new Set()
    public importedSymbols: Set<ImportedSymbol> = new Set()

    private knownHashes: Set<string> = new Set()
    private insertedBlobs: Set<string> = new Set()
    private insertedHovers: Set<string> = new Set()
    private documentDatas: Map<protocol.Id, DocumentData | null> = new Map()
    private monikerSets: Map<protocol.Id, Set<protocol.Id>> = new Map()
    private monikerAttachments: Map<protocol.Id, protocol.Id> = new Map()
    private containsDatas: Map<protocol.Id, protocol.Id[]> = new Map()

    constructor(private connection: Connection) {
        const wrap = (f: (element: any) => void) => (element: any) => {
            f.bind(this)(element)
            return Promise.resolve()
        }

        this.vertexHandlerMap[protocol.VertexLabels.definitionResult] = wrap(this.handleDefinitionResult)
        this.vertexHandlerMap[protocol.VertexLabels.document] = wrap(this.handleDocument)
        this.vertexHandlerMap[protocol.VertexLabels.event] = this.handleEvent.bind(this)
        this.vertexHandlerMap[protocol.VertexLabels.hoverResult] = wrap(this.handleHover)
        this.vertexHandlerMap[protocol.VertexLabels.metaData] = this.handleMetaData.bind(this)
        this.vertexHandlerMap[protocol.VertexLabels.moniker] = wrap(this.handleMoniker)
        this.vertexHandlerMap[protocol.VertexLabels.packageInformation] = wrap(this.handlePackageInformation)
        this.vertexHandlerMap[protocol.VertexLabels.range] = wrap(this.handleRange)
        this.vertexHandlerMap[protocol.VertexLabels.referenceResult] = wrap(this.handleReferenceResult)
        this.vertexHandlerMap[protocol.VertexLabels.resultSet] = wrap(this.handleResultSet)

        this.edgeHandlerMap[protocol.EdgeLabels.contains] = wrap(this.handleContains)
        this.edgeHandlerMap[protocol.EdgeLabels.item] = wrap(this.handleItemEdge)
        this.edgeHandlerMap[protocol.EdgeLabels.moniker] = wrap(this.handleMonikerEdge)
        this.edgeHandlerMap[protocol.EdgeLabels.next] = wrap(this.handleNextEdge)
        this.edgeHandlerMap[protocol.EdgeLabels.nextMoniker] = wrap(this.handleNextMonikerEdge)
        this.edgeHandlerMap[protocol.EdgeLabels.packageInformation] = wrap(this.handlePackageInformationEdge)
        this.edgeHandlerMap[protocol.EdgeLabels.textDocument_definition] = wrap(this.handleDefinitionEdge)
        this.edgeHandlerMap[protocol.EdgeLabels.textDocument_hover] = wrap(this.handleHoverEdge)
        this.edgeHandlerMap[protocol.EdgeLabels.textDocument_references] = wrap(this.handleReferenceEdge)
    }

    public async insert(element: protocol.Vertex | protocol.Edge): Promise<void> {
        const handler =
            element.type === protocol.ElementTypes.vertex
                ? this.vertexHandlerMap[element.label]
                : this.edgeHandlerMap[element.label]

        if (handler) {
            await handler(element)
        }
    }

    //
    // Vertex Handlers

    private async handleMetaData(vertex: protocol.MetaData): Promise<void> {
        this.projectRoot = vertex.projectRoot

        await this.connection.getRepository(entities.Meta).save({
            lsifVersion: vertex.version,
            sourcegraphVersion: INTERNAL_LSIF_VERSION,
        })
    }

    private async handleEvent(event: protocol.Event): Promise<void> {
        if (event.scope === protocol.EventScope.document && event.kind === protocol.EventKind.begin) {
            this.handleDocumentBegin(event as protocol.DocumentEvent)
        }

        if (event.scope === protocol.EventScope.document && event.kind === protocol.EventKind.end) {
            await this.handleDocumentEnd(event as protocol.DocumentEvent)
        }
    }

    private handleDefinitionResult = this.setById(this.definitionDatas, () => new Map())
    private handleDocument = this.setById(this.documents, (e: protocol.Document) => e)
    private handleHover = this.setById(this.hoverDatas, (e: protocol.HoverResult) => e.result)
    private handleMoniker = this.setById(this.monikerDatas, e => e)
    private handlePackageInformation = this.setById(this.packageInformationDatas, e => e)
    private handleRange = this.setById(this.rangeDatas, e => e)
    private handleReferenceResult = this.setById(this.referenceDatas, e => new Map())
    private handleResultSet = this.setById(this.resultSetDatas, _ => ({}))

    private setById<K extends { id: protocol.Id }, V>(
        map: Map<protocol.Id, V>,
        factory: (element: K) => V
    ): (element: K) => void {
        return (element: K) => {
            map.set(element.id, factory(element))
        }
    }

    //
    // Edge Handlers

    private handleContains(contains: protocol.contains): void {
        let values = this.containsDatas.get(contains.outV)
        if (values === undefined) {
            values = []
            this.containsDatas.set(contains.outV, values)
        }

        values.push(...contains.inVs)
    }

    private handleItemEdge(edge: protocol.item): void {
        let property: protocol.ItemEdgeProperties | undefined = edge.property
        if (property === undefined) {
            const map: Map<protocol.Id, db.DefinitionResultData> = assertDefined(edge.outV, this.definitionDatas)
            let data: db.DefinitionResultData | undefined = map.get(edge.document)
            if (data === undefined) {
                data = { values: edge.inVs.slice() }
                map.set(edge.document, data)
            } else {
                data.values.push(...edge.inVs)
            }
        } else {
            const map: Map<protocol.Id, db.ReferenceResultData> = assertDefined(edge.outV, this.referenceDatas)
            let data: db.ReferenceResultData | undefined = map.get(edge.document)
            if (data === undefined) {
                data = {}
                map.set(edge.document, data)
            }

            if (property === protocol.ItemEdgeProperties.definitions) {
                if (data.definitions === undefined) {
                    data.definitions = edge.inVs.slice()
                } else {
                    data.definitions.push(...edge.inVs)
                }
            }

            if (property === protocol.ItemEdgeProperties.references) {
                if (data.references === undefined) {
                    data.references = edge.inVs.slice()
                } else {
                    data.references.push(...edge.inVs)
                }
            }
        }
    }

    private handleMonikerEdge(moniker: protocol.moniker): void {
        assertDefined<db.RangeData | db.ResultSetData>(moniker.outV, this.rangeDatas, this.resultSetDatas)
        assertDefined(moniker.inV, this.monikerDatas)
        this.monikerAttachments.set(moniker.outV, moniker.inV)
        this.updateMonikerSets([moniker.inV])
    }

    private handleNextEdge(edge: protocol.next): void {
        const outV = assertDefined<db.RangeData | db.ResultSetData>(edge.outV, this.rangeDatas, this.resultSetDatas)
        assertDefined(edge.inV, this.resultSetDatas)
        outV.next = edge.inV
    }

    private handleNextMonikerEdge(nextMoniker: protocol.nextMoniker): void {
        assertDefined(nextMoniker.inV, this.monikerDatas)
        assertDefined(nextMoniker.outV, this.monikerDatas)
        this.updateMonikerSets([nextMoniker.inV, nextMoniker.outV])
    }

    private handlePackageInformationEdge(packageInformation: protocol.packageInformation): void {
        const source: db.MonikerData = assertDefined(packageInformation.outV, this.monikerDatas)
        const packageInfo = assertDefined(packageInformation.inV, this.packageInformationDatas)
        source.packageInformation = packageInformation.inV

        if (source.kind === 'export') {
            this.exportedPackages.add({ scheme: source.scheme, name: packageInfo.name, version: packageInfo.version! })
        }
    }

    private handleDefinitionEdge(edge: protocol.textDocument_definition): void {
        const outV = assertDefined<db.RangeData | db.ResultSetData>(edge.outV, this.rangeDatas, this.resultSetDatas)
        this.ensureMoniker(outV)
        assertDefined(edge.inV, this.definitionDatas)
        outV.definitionResult = edge.inV
    }


    private handleHoverEdge(edge: protocol.textDocument_hover): void {
        const outV = assertDefined<db.RangeData | db.ResultSetData>(edge.outV, this.rangeDatas, this.resultSetDatas)
        this.ensureMoniker(outV)
        assertDefined(edge.inV, this.hoverDatas)
        outV.hoverResult = edge.inV
    }

    private handleReferenceEdge(edge: protocol.textDocument_references): void {
        const outV = assertDefined<db.RangeData | db.ResultSetData>(edge.outV, this.rangeDatas, this.resultSetDatas)
        this.ensureMoniker(outV)
        assertDefined(edge.inV, this.referenceDatas)
        outV.referenceResult = edge.inV
    }

    //
    // Event Handlers

    private handleDocumentBegin(event: protocol.DocumentEvent): void {
        this.getOrCreateDocumentData(assertDefined(event.data, this.documents))
        this.documents.delete(event.data)
    }

    private async handleDocumentEnd(event: protocol.DocumentEvent): Promise<void> {
        const documentData = assertDefined(event.data, this.documentDatas)

        for (const data of this.monikerDatas.values()) {
            if (data.kind === 'import' && data.packageInformation) {
                const packageInformation = assertDefined(data.packageInformation!, this.packageInformationDatas)

                this.importedSymbols.add({
                    scheme: data.scheme,
                    name: packageInformation!.name!,
                    version: packageInformation!.version!,
                    identifier: data.identifier,
                })
            }
        }

        for (const [key, value] of this.monikerAttachments.entries()) {
            const ids = this.monikerSets.get(value)
            if (!ids) {
                throw new Error('moniker set is empty')
            }

            const source = assertDefined<db.RangeData | db.ResultSetData>(key, this.rangeDatas, this.resultSetDatas)
            ids.forEach(id => assertDefined(id, this.monikerDatas))
            source.monikers = Array.from(ids)
        }

        const contains = this.containsDatas.get(event.data)
        if (contains !== undefined) {
            for (let id of contains) {
                const range = assertDefined(id, this.rangeDatas)
                await documentData.addRangeData(id, range)
            }
        }

        for (const [key, value] of this.packageInformationDatas) {
            documentData.addPackageInformation(key, value)
        }

        let data = await documentData.finalize()

        try {
            if (this.knownHashes.has(data.hash)) {
                // TODO - see how this can happen (empty files?)
                console.log('woah, duplicate?')
                return
            }
        } finally {
            this.documentDatas.set(event.id, null)
        }

        await this.connection.getRepository(entities.Blob).save({ hash: data.hash, value: data.encoded })

        await this.connection.getRepository(entities.Document).save({
            hash: data.hash,
            uri: documentData.getUri(),
        })

        const defs = []
        for (let definition of data.definitions || []) {
            for (let range of definition.ranges) {
                defs.push({
                    scheme: definition.scheme,
                    identifier: definition.indentifier,
                    startLine: range[0],
                    startCharacter: range[1],
                    endLine: range[2],
                    endCharacter: range[3],
                    documentHash: data.hash,
                })
            }
        }

        const refs = []
        for (let reference of data.references || []) {
            for (const arr of [reference.definitions, reference.references]) {
                for (const range of arr || []) {
                    refs.push({
                        scheme: reference.scheme,
                        identifier: reference.indentifier,
                        startLine: range[0],
                        startCharacter: range[1],
                        endLine: range[2],
                        endCharacter: range[3],
                        documentHash: data.hash,
                    })
                }
            }
        }

        await this.connection
            .createQueryBuilder()
            .insert()
            .into(entities.Def)
            .values(defs)
            .execute()

        await this.connection
            .createQueryBuilder()
            .insert()
            .into(entities.Ref)
            .values(refs)
            .execute()
    }

    //
    // TODO - categorize

    private getOrCreateDocumentData(document: protocol.Document): DocumentData {
        let result: DocumentData | undefined | null = this.documentDatas.get(document.id)
        if (result === null) {
            throw new Error(`The document ${document.uri} has already been processed`)
        }

        if (!this.projectRoot) {
            throw new Error(`No project root`)
        }

        result = new DocumentData(this.projectRoot!, document, this)
        this.documentDatas.set(document.id, result)
        return result
    }

    public getAndDeleteDefinitions(
        definitionResult: protocol.Id,
        documentId: protocol.Id
    ): db.DefinitionResultData | undefined {
        const map = assertDefined(definitionResult, this.definitionDatas)
        const result = map.get(documentId)
        map.delete(documentId)
        return result
    }

    public getAndDeleteReferences(
        referenceResult: protocol.Id,
        documentId: protocol.Id
    ): db.ReferenceResultData | undefined {
        const map = assertDefined(referenceResult, this.referenceDatas)
        const result = map.get(documentId)
        map.delete(documentId)
        return result
    }

    private updateMonikerSets(vals: protocol.Id[]): void {
        const combined = new Set<protocol.Id>()
        for (const val of vals) {
            combined.add(val)

            const otherSet = this.monikerSets.get(val)
            if (otherSet) {
                otherSet.forEach(v => combined.add(v))
            }
        }

        for (const val of combined) {
            this.monikerSets.set(val, combined)
        }
    }

    private ensureMoniker(data: db.RangeData | db.ResultSetData): void {
        if (data.monikers !== undefined) {
            return
        }

        const id = uuid.v4()
        const monikerData: db.MonikerData = { id: id, scheme: '$synthetic', identifier: id }
        data.monikers = [monikerData.identifier]
        this.monikerDatas.set(monikerData.identifier, monikerData)
    }

    public async storeHover(moniker: db.MonikerData, id: protocol.Id): Promise<void> {
        let hover = this.hoverDatas.get(id)
        if (hover === undefined) {
            // We have already processed the hover
            return
        }

        const { hash, encoded } = await hashAndEncodeJSON(hover)

        if (!this.knownHashes.has(hash)) {
            if (!this.insertedBlobs.has(hash)) {
                await this.connection.getRepository(entities.Blob).save({ hash, value: encoded })
                this.insertedBlobs.add(hash)
            }

            const hoverHash = hashJSON({ s: moniker.scheme, i: moniker.identifier, b: hash })

            if (!this.insertedHovers.has(hoverHash)) {
                await this.connection
                    .getRepository(entities.Hover)
                    .save({ scheme: moniker.scheme, identifier: moniker.identifier, blobHash: hash })
                this.insertedHovers.add(hoverHash)
            }
        }
    }
}

class DocumentData {
    private provider: BlobStore
    private id: protocol.Id
    private uri: string

    private blob: db.DocumentBlob = { ranges: {} }
    private definitions: MonikerScopedResultData<db.DefinitionResultData>[] = []
    private references: MonikerScopedResultData<db.ReferenceResultData>[] = []

    constructor(projectRoot: string, document: protocol.Document, provider: BlobStore) {
        this.provider = provider
        this.id = document.id
        this.uri = document.uri.slice(projectRoot.length + 1)
    }

    public getUri(): string {
        return this.uri
    }

    public async addRangeData(id: protocol.Id, data: db.RangeData): Promise<void> {
        this.blob.ranges[id] = data
        await this.addReferencedData(data)
    }

    private async addResultSetData(id: protocol.Id, resultSet: db.ResultSetData): Promise<void> {
        if (this.blob.resultSets === undefined) {
            this.blob.resultSets = LiteralMap.create()
        }

        // Many ranges can point to the same result set. Make sure
        // we only travers once.
        if (this.blob.resultSets[id] !== undefined) {
            return
        }

        this.blob.resultSets[id] = resultSet
        await this.addReferencedData(resultSet)
    }

    private async addReferencedData(item: db.RangeData | db.ResultSetData): Promise<void> {
        const monikers = []

        if (item.monikers !== undefined) {
            for (const itemMoniker of item.monikers) {
                const moniker = assertDefined(itemMoniker, this.provider.monikerDatas)
                this.addMoniker(itemMoniker, moniker)
                monikers.push(moniker)
            }
        }

        if (item.next !== undefined) {
            await this.addResultSetData(item.next, assertDefined(item.next, this.provider.resultSetDatas))
        }

        if (item.hoverResult !== undefined) {
            let local = true
            for (const moniker of monikers) {
                if (!Monikers.isLocal(moniker)) {
                    local = false
                    await this.provider.storeHover(moniker, item.hoverResult)
                }
            }

            if (local) {
                if (this.blob.hovers === undefined) {
                    this.blob.hovers = LiteralMap.create()
                }

                if (this.blob.hovers[item.hoverResult] === undefined) {
                    this.blob.hovers[item.hoverResult] = assertDefined(item.hoverResult, this.provider.hoverDatas)
                }
            }
        }

        if (item.definitionResult) {
            const definitions = this.provider.getAndDeleteDefinitions(item.definitionResult, this.id)

            if (definitions !== undefined) {
                let local = true
                for (const moniker of monikers) {
                    if (!Monikers.isLocal(moniker)) {
                        local = false
                        this.definitions.push({ moniker, data: definitions })
                    }

                    if (local) {
                        if (this.blob.definitionResults === undefined) {
                            this.blob.definitionResults = LiteralMap.create()
                        }

                        this.blob.definitionResults[item.definitionResult] = definitions
                    }
                }
            }
        }

        if (item.referenceResult) {
            const references = this.provider.getAndDeleteReferences(item.referenceResult, this.id)

            if (references !== undefined) {
                let local = true
                for (const moniker of monikers) {
                    if (!Monikers.isLocal(moniker)) {
                        local = false
                        this.references.push({ moniker, data: references })
                    }
                }

                if (local) {
                    if (this.blob.referenceResults === undefined) {
                        this.blob.referenceResults = LiteralMap.create()
                    }
                    this.blob.referenceResults[item.referenceResult] = references
                }
            }
        }
    }

    private addMoniker(id: protocol.Id, moniker: db.MonikerData): void {
        if (this.blob.monikers === undefined) {
            this.blob.monikers = LiteralMap.create()
        }
        this.blob.monikers![id] = moniker
    }

    public addPackageInformation(id: protocol.Id, packageInformation: db.PackageInformationData): void {
        if (this.blob.packageInformation === undefined) {
            this.blob.packageInformation = LiteralMap.create()
        }
        this.blob.packageInformation![id] = packageInformation
    }

    public async finalize(): Promise<DocumentDatabaseData> {
        const id2InlineRange = (id: protocol.Id): [number, number, number, number] => {
            const range = this.blob.ranges[id]
            return [range.start.line, range.start.character, range.end.line, range.end.character]
        }

        let externalDefinitions: ExternalDefinition[] = []
        let externalReferences: ExternalReference[] = []

        for (let definition of this.definitions) {
            externalDefinitions.push({
                scheme: definition.moniker.scheme,
                indentifier: definition.moniker.identifier,
                ranges: definition.data.values.map(id2InlineRange),
            })
        }

        for (let reference of this.references) {
            externalReferences.push({
                scheme: reference.moniker.scheme,
                indentifier: reference.moniker.identifier,
                definitions: reference.data.definitions ? reference.data.definitions.map(id2InlineRange) : undefined,
                references: reference.data.references ? reference.data.references.map(id2InlineRange) : undefined,
            })
        }

        const { hash, encoded } = await hashAndEncodeJSON(this.blob)

        return {
            hash: hash,
            encoded: encoded,
            definitions: externalDefinitions.length > 0 ? externalDefinitions : undefined,
            references: externalReferences.length > 0 ? externalReferences : undefined,
        }
    }
}

function assertDefined<T>(id: protocol.Id, ...maps: Map<protocol.Id, T | null>[]): T {
    for (const map of maps) {
        const value = map.get(id)
        if (value === null) {
            throw new Error(`Element ${id} has already been processed.`)
        }

        if (value !== undefined) {
            return value
        }
    }

    throw new Error(`Element '${id}' is referenced before its definition.`)
}
