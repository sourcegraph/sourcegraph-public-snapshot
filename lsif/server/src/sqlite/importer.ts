import { fs } from 'mz'
import * as lsp from 'vscode-languageserver-protocol'
import * as readline from 'readline'
import * as uuid from 'uuid'
import * as entities from './entities'
import { hashAndEncodeJSON, hashJSON } from './encoding'
import * as protocol from 'lsif-protocol'
import { Connection } from 'typeorm'
import { connectionCache } from './cache'
import {
    MonikerData,
    RangeData,
    ResultSetData,
    DefinitionResultData,
    ReferenceResultData,
    DocumentBlob,
} from './models'
import { Id } from 'lsif-protocol'

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
    definitions: ExternalDefinition[]
    references: ExternalReference[]
}

interface WrappedDocumentBlob extends DocumentBlob {
    id: protocol.Id
    uri: string
    definitions: MonikerScopedResultData<DefinitionResultData>[] // TODO - get rid of this class definition as well
    references: MonikerScopedResultData<ReferenceResultData>[] // TODO - get rid of this class definition as well
}

interface MonikerScopedResultData<T> {
    moniker: MonikerData
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

    // TODO - standardize
    private definitionDatas: Map<protocol.Id, Map<protocol.Id, DefinitionResultData>> = new Map()
    private documents: Map<protocol.Id, protocol.Document> = new Map()
    private hoverDatas: Map<protocol.Id, lsp.Hover> = new Map()
    private monikerDatas: Map<protocol.Id, MonikerData> = new Map()
    private packageInformationDatas: Map<protocol.Id, protocol.PackageInformation> = new Map()
    private rangeDatas: Map<protocol.Id, RangeData> = new Map()
    private referenceDatas: Map<protocol.Id, Map<protocol.Id, ReferenceResultData>> = new Map()
    private resultSetDatas: Map<protocol.Id, ResultSetData> = new Map()

    // TODO
    private projectRoot: string | undefined

    // TODO - expose via method?
    public exportedPackages: Set<ExportedPackage> = new Set()
    public importedSymbols: Set<ImportedSymbol> = new Set()

    private knownHashes: Set<string> = new Set()
    private insertedBlobs: Set<string> = new Set()
    private insertedHovers: Set<string> = new Set()
    private documentDatas: Map<protocol.Id, WrappedDocumentBlob | null> = new Map()
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

    private handleDefinitionResult = this.setById(this.definitionDatas, _ => new Map())
    private handleDocument = this.setById(this.documents, (e: protocol.Document) => e)
    private handleHover = this.setById(this.hoverDatas, (e: protocol.HoverResult) => e.result)
    private handleMoniker = this.setById(this.monikerDatas, e => e)
    private handlePackageInformation = this.setById(this.packageInformationDatas, e => e)
    private handleRange = this.setById(this.rangeDatas, (e: protocol.Range) => e)
    private handleReferenceResult = this.setById(this.referenceDatas, _ => new Map())
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
        if (!values) {
            this.containsDatas.set(contains.outV, contains.inVs.slice())
        } else {
            values.push(...contains.inVs)
        }
    }

    private handleItemEdge(edge: protocol.item): void {
        let property: protocol.ItemEdgeProperties | undefined = edge.property
        if (!property) {
            const map = assertDefined(edge.outV, this.definitionDatas)

            let data = map.get(edge.document)
            if (!data) {
                data = { values: edge.inVs.slice() }
                map.set(edge.document, data)
            } else {
                data.values.push(...edge.inVs)
            }
        } else {
            const map = assertDefined(edge.outV, this.referenceDatas)

            let data = map.get(edge.document)
            if (!data) {
                data = {}
                map.set(edge.document, data)
            }

            if (property === protocol.ItemEdgeProperties.definitions) {
                if (!data.definitions) {
                    data.definitions = edge.inVs.slice()
                } else {
                    data.definitions.push(...edge.inVs)
                }
            }

            if (property === protocol.ItemEdgeProperties.references) {
                if (!data.references) {
                    data.references = edge.inVs.slice()
                } else {
                    data.references.push(...edge.inVs)
                }
            }
        }
    }

    private handleMonikerEdge(moniker: protocol.moniker): void {
        assertDefined<RangeData | ResultSetData>(moniker.outV, this.rangeDatas, this.resultSetDatas)
        assertDefined(moniker.inV, this.monikerDatas)
        this.monikerAttachments.set(moniker.outV, moniker.inV)
        this.updateMonikerSets([moniker.inV])
    }

    private handleNextEdge(edge: protocol.next): void {
        const outV = assertDefined<RangeData | ResultSetData>(edge.outV, this.rangeDatas, this.resultSetDatas)
        assertDefined(edge.inV, this.resultSetDatas)
        outV.next = edge.inV
    }

    private handleNextMonikerEdge(nextMoniker: protocol.nextMoniker): void {
        assertDefined(nextMoniker.inV, this.monikerDatas)
        assertDefined(nextMoniker.outV, this.monikerDatas)
        this.updateMonikerSets([nextMoniker.inV, nextMoniker.outV])
    }

    private handlePackageInformationEdge(packageInformation: protocol.packageInformation): void {
        const source: MonikerData = assertDefined(packageInformation.outV, this.monikerDatas)
        const packageInfo = assertDefined(packageInformation.inV, this.packageInformationDatas)
        source.packageInformation = packageInformation.inV

        if (source.kind === 'export') {
            this.exportedPackages.add({ scheme: source.scheme, name: packageInfo.name, version: packageInfo.version! })
        }
    }

    private handleDefinitionEdge(edge: protocol.textDocument_definition): void {
        const outV = assertDefined<RangeData | ResultSetData>(edge.outV, this.rangeDatas, this.resultSetDatas)
        this.ensureMoniker(outV)
        assertDefined(edge.inV, this.definitionDatas)
        outV.definitionResult = edge.inV
    }

    private handleHoverEdge(edge: protocol.textDocument_hover): void {
        const outV = assertDefined<RangeData | ResultSetData>(edge.outV, this.rangeDatas, this.resultSetDatas)
        this.ensureMoniker(outV)
        assertDefined(edge.inV, this.hoverDatas)
        outV.hoverResult = edge.inV
    }

    private handleReferenceEdge(edge: protocol.textDocument_references): void {
        const outV = assertDefined<RangeData | ResultSetData>(edge.outV, this.rangeDatas, this.resultSetDatas)
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
        let data = await this.finalizeBlob(documentData)

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
            uri: documentData.uri,
        })

        const defs = []
        for (let definition of data.definitions) {
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
        for (let reference of data.references) {
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

    private getOrCreateDocumentData(document: protocol.Document): WrappedDocumentBlob {
        let result: WrappedDocumentBlob | undefined | null = this.documentDatas.get(document.id)
        if (result === null) {
            throw new Error(`The document ${document.uri} has already been processed`)
        }

        if (!this.projectRoot) {
            throw new Error(`No project root`)
        }

        result = {
            id: document.id,
            uri: document.uri.slice(this.projectRoot.length + 1),
            definitions: [],
            references: [],
            ranges: {},
            resultSets: {},
            definitionResults: {},
            referenceResults: {},
            hovers: {},
            monikers: {},
            packageInformation: {},
        }

        this.documentDatas.set(document.id, result)
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

    private ensureMoniker(data: RangeData | ResultSetData): void {
        if (!data.monikers) {
            const id = uuid.v4()
            const monikerData: MonikerData = { id: id, scheme: '$synthetic', identifier: id }
            data.monikers = [monikerData.identifier]
            this.monikerDatas.set(monikerData.identifier, monikerData)
        }
    }

    private async storeHover(moniker: MonikerData, id: protocol.Id): Promise<void> {
        let hover = this.hoverDatas.get(id)
        // if (!hover) {
        //     // Should never happen
        //     return
        // }

        const { hash, encoded } = await hashAndEncodeJSON(hover)
        if (this.knownHashes.has(hash)) {
            return
        }

        if (!this.insertedBlobs.has(hash)) {
            await this.connection.getRepository(entities.Blob).save({ hash, value: encoded })
            this.insertedBlobs.add(hash)
        }

        const hoverHash = hashJSON({ s: moniker.scheme, i: moniker.identifier, b: hash })
        if (this.insertedHovers.has(hoverHash)) {
            return
        }

        this.insertedHovers.add(hoverHash)

        await this.connection
            .getRepository(entities.Hover)
            .save({ scheme: moniker.scheme, identifier: moniker.identifier, blobHash: hash })
    }

    //
    // Blob things

    private async addReferencedDataToBlob(blob: WrappedDocumentBlob, item: RangeData | ResultSetData): Promise<void> {
        const monikers = []
        for (const itemMoniker of item.monikers || []) {
            const moniker = assertDefined(itemMoniker, this.monikerDatas)
            blob.monikers![itemMoniker] = moniker
            monikers.push(moniker)
        }

        // Many ranges can point to the same result set. Make sure we only travers once.
        if (item.next && !blob.resultSets[item.next]) {
            const resultSet = assertDefined(item.next, this.resultSetDatas)
            blob.resultSets[item.next] = resultSet
            await this.addReferencedDataToBlob(blob, resultSet)
        }

        if (item.hoverResult) this.addHoverResultsToBlob(blob, item.hoverResult, monikers)
        if (item.definitionResult) this.addDefinitionResultsToBlob(blob, item.definitionResult, monikers)
        if (item.referenceResult) this.addReferenceResultsToBlob(blob, item.referenceResult, monikers)
    }

    private async addHoverResultsToBlob(
        blob: WrappedDocumentBlob,
        hoverResult: Id,
        monikers: MonikerData[]
    ): Promise<void> {
        let local = true
        for (const moniker of monikers.filter(m => m.kind !== protocol.MonikerKind.local)) {
            local = false
            await this.storeHover(moniker, hoverResult)
        }

        if (local && !blob.hovers[hoverResult]) {
            blob.hovers[hoverResult] = assertDefined(hoverResult, this.hoverDatas)
        }
    }

    private async addDefinitionResultsToBlob(
        blob: WrappedDocumentBlob,
        definitionResult: Id,
        monikers: MonikerData[]
    ): Promise<void> {
        const map = assertDefined(definitionResult, this.definitionDatas)
        const definitions = map.get(blob.id)
        map.delete(blob.id) // TODO - can we be sure this is correct?
        if (!definitions) {
            return
        }

        let local = true
        for (const moniker of monikers.filter(m => m.kind !== protocol.MonikerKind.local)) {
            local = false
            blob.definitions.push({ moniker, data: definitions })
        }

        if (local) {
            blob.definitionResults[definitionResult] = definitions
        }
    }

    private async addReferenceResultsToBlob(
        blob: WrappedDocumentBlob,
        referenceResult: Id,
        monikers: MonikerData[]
    ): Promise<void> {
        const map = assertDefined(referenceResult, this.referenceDatas)
        const references = map.get(blob.id)
        map.delete(blob.id) // TODO - can we be sure this is correct?
        if (!references) {
            return
        }

        let local = true
        for (const moniker of monikers.filter(m => m.kind !== protocol.MonikerKind.local)) {
            local = false
            blob.references.push({ moniker, data: references })
        }

        if (local) {
            blob.referenceResults[referenceResult] = references
        }
    }

    private async finalizeBlob(blob: WrappedDocumentBlob): Promise<DocumentDatabaseData> {
        const convertRange = (id: protocol.Id): [number, number, number, number] => {
            const range = blob.ranges[id]
            return [range.start.line, range.start.character, range.end.line, range.end.character]
        }

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

            const source = assertDefined<RangeData | ResultSetData>(key, this.rangeDatas, this.resultSetDatas)
            ids.forEach(id => assertDefined(id, this.monikerDatas))
            source.monikers = Array.from(ids)
        }

        const contains = this.containsDatas.get(blob.id)
        if (contains) {
            for (let id of contains) {
                const range = assertDefined(id, this.rangeDatas)
                blob.ranges[id] = range
                await this.addReferencedDataToBlob(blob, range)
            }
        }

        for (const [key, value] of this.packageInformationDatas) {
            blob.packageInformation![key] = value
        }

        const definitions = []
        for (const definition of blob.definitions) {
            definitions.push({
                scheme: definition.moniker.scheme,
                indentifier: definition.moniker.identifier,
                ranges: definition.data.values.map(convertRange),
            })
        }

        const references = []
        for (const reference of blob.references) {
            references.push({
                scheme: reference.moniker.scheme,
                indentifier: reference.moniker.identifier,
                definitions: reference.data.definitions ? reference.data.definitions.map(convertRange) : undefined,
                references: reference.data.references ? reference.data.references.map(convertRange) : undefined,
            })
        }

        return { ...(await hashAndEncodeJSON(blob)), definitions, references }
    }
}

function assertDefined<T>(id: protocol.Id, ...maps: Map<protocol.Id, T | null>[]): T {
    for (const map of maps) {
        const value = map.get(id)
        if (value === null) {
            throw new Error(`Element ${id} has already been processed.`)
        }

        if (value) {
            return value
        }
    }

    throw new Error(`Element '${id}' is referenced before its definition.`)
}
