import * as crypto from 'crypto'
import * as fs from 'fs'
import * as lsp from 'vscode-languageserver-protocol'
import * as readline from 'readline'
import * as uuid from 'uuid'
import { gzipSync } from 'mz/zlib'
import Sqlite from 'better-sqlite3'
import * as db from './database'
import * as protocol from 'lsif-protocol'
import { SchemaTable, Inserter, createIndices, createTables } from './sqlite'

const SCHEMA_TABLES: { [K: string]: SchemaTable } = {
    meta: {
        columns: [
            { name: 'id', type: 'integer unique primary key', indexed: false },
            { name: 'value', type: 'text not null', indexed: false },
        ],
        batchInsertSize: 1,
    },
    blobs: {
        columns: [
            { name: 'hash', type: 'text unique primary key', indexed: true },
            { name: 'content', type: 'blob not null', indexed: false },
        ],
        batchInsertSize: 16,
    },
    documents: {
        columns: [
            { name: 'documentHash', type: 'text not null', indexed: true },
            { name: 'uri', type: 'text not null', indexed: false },
        ],
        batchInsertSize: 16,
    },
    defs: {
        columns: [
            { name: 'scheme', type: 'text not null', indexed: true },
            { name: 'identifier', type: 'text not null', indexed: true },
            { name: 'documentHash', type: 'text not null', indexed: true },
            { name: 'startLine', type: 'integer not null', indexed: false },
            { name: 'startCharacter', type: 'integer not null', indexed: false },
            { name: 'endLine', type: 'integer not null', indexed: false },
            { name: 'endCharacter', type: 'integer not null', indexed: false },
        ],
        batchInsertSize: 128,
    },
    refs: {
        columns: [
            { name: 'scheme', type: 'text not null', indexed: true },
            { name: 'identifier', type: 'text not null', indexed: true },
            { name: 'documentHash', type: 'text not null', indexed: true },
            { name: 'kind', type: 'integer not null', indexed: false },
            { name: 'startLine', type: 'integer not null', indexed: false },
            { name: 'startCharacter', type: 'integer not null', indexed: false },
            { name: 'endLine', type: 'integer not null', indexed: false },
            { name: 'endCharacter', type: 'integer not null', indexed: false },
        ],
        batchInsertSize: 64,
    },
    hovers: {
        columns: [
            { name: 'scheme', type: 'text not null', indexed: true },
            { name: 'identifier', type: 'text not null', indexed: true },
            { name: 'hoverHash', type: 'text not null', indexed: true },
        ],
        batchInsertSize: 128,
    },
}

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

export async function convertToBlob(
    inFile: string,
    outFile: string
): Promise<{ exportedPackages: Set<ExportedPackage>; importedSymbols: Set<ImportedSymbol> }> {
    const db = new BlobStore(outFile)
    const input = fs.createReadStream(inFile, { encoding: 'utf8' }) // TODO - go directly from input stream

    return new Promise((resolve, reject) => {
        const rd = readline.createInterface(input)
        rd.on('line', line => {
            if (!line) {
                return
            }

            let element: protocol.Vertex | protocol.Edge
            try {
                element = JSON.parse(line)
            } catch (err) {
                throw new Error(`Parsing failed for line:\n${line}`)
            }

            db.insert(element)
        })

        rd.on('close', () => {
            db.close()
            resolve({
                exportedPackages: db.exportedPackages,
                importedSymbols: db.importedSymbols,
            })
        })

        rd.on('error', reject)
    })
}

class BlobStore {
    private db: Sqlite.Database
    private metaInserter: Inserter
    private documentInserter: Inserter
    private blobInserter: Inserter
    private definitionInserter: Inserter
    private referenceInserter: Inserter
    private hoverInserter: Inserter

    private vertexHandlerMap: { [K: string]: (element: any) => void } = {}
    private edgeHandlerMap: { [K: string]: (element: any) => void } = {}

    public exportedPackages: Set<ExportedPackage> = new Set()
    public importedSymbols: Set<ImportedSymbol> = new Set()

    private knownHashes: Set<string> = new Set()
    private insertedBlobs: Set<string> = new Set()
    private insertedHovers: Set<string> = new Set()
    private documents: Map<protocol.Id, protocol.Document> = new Map()
    private documentDatas: Map<protocol.Id, DocumentData | null> = new Map()
    private foldingRanges: Map<protocol.Id, lsp.FoldingRange[]> = new Map()
    private documentSymbols: Map<protocol.Id, lsp.DocumentSymbol[] | protocol.RangeBasedDocumentSymbol[]> = new Map()
    private rangeDatas: Map<protocol.Id, db.RangeData> = new Map()
    private resultSetDatas: Map<protocol.Id, db.ResultSetData> = new Map()
    private monikerDatas: Map<protocol.Id, db.MonikerData> = new Map()
    private monikerSets: Map<protocol.Id, Set<protocol.Id>> = new Map()
    private monikerAttachments: Map<protocol.Id, protocol.Id> = new Map()
    private packageInformationDatas: Map<protocol.Id, db.PackageInformationData> = new Map()
    private hoverDatas: Map<protocol.Id, lsp.Hover> = new Map()
    private definitionDatas: Map<protocol.Id, Map<protocol.Id, db.DefinitionResultData>> = new Map()
    private referenceDatas: Map<protocol.Id, Map<protocol.Id, db.ReferenceResultData>> = new Map()
    private containsDatas: Map<protocol.Id, protocol.Id[]> = new Map()

    constructor(filename: string) {
        this.vertexHandlerMap[protocol.VertexLabels.definitionResult] = this.handleDefinitionResult
        this.vertexHandlerMap[protocol.VertexLabels.document] = this.handleDocument
        this.vertexHandlerMap[protocol.VertexLabels.documentSymbolResult] = this.handleDocumentSymbols
        this.vertexHandlerMap[protocol.VertexLabels.event] = this.handleEvent
        this.vertexHandlerMap[protocol.VertexLabels.foldingRangeResult] = this.handleFoldingRange
        this.vertexHandlerMap[protocol.VertexLabels.hoverResult] = this.handleHover
        this.vertexHandlerMap[protocol.VertexLabels.metaData] = this.handleMetaData
        this.vertexHandlerMap[protocol.VertexLabels.moniker] = this.handleMoniker
        this.vertexHandlerMap[protocol.VertexLabels.packageInformation] = this.handlePackageInformation
        this.vertexHandlerMap[protocol.VertexLabels.range] = this.handleRange
        this.vertexHandlerMap[protocol.VertexLabels.referenceResult] = this.handleReferenceResult
        this.vertexHandlerMap[protocol.VertexLabels.resultSet] = this.handleResultSet

        this.edgeHandlerMap[protocol.EdgeLabels.contains] = this.handleContains
        this.edgeHandlerMap[protocol.EdgeLabels.item] = this.handleItemEdge
        this.edgeHandlerMap[protocol.EdgeLabels.moniker] = this.handleMonikerEdge
        this.edgeHandlerMap[protocol.EdgeLabels.next] = this.handleNextEdge
        this.edgeHandlerMap[protocol.EdgeLabels.nextMoniker] = this.handleNextMonikerEdge
        this.edgeHandlerMap[protocol.EdgeLabels.packageInformation] = this.handlePackageInformationEdge
        this.edgeHandlerMap[protocol.EdgeLabels.textDocument_definition] = this.handleDefinitionEdge
        this.edgeHandlerMap[protocol.EdgeLabels.textDocument_documentSymbol] = this.handleDocumentSymbolEdge
        this.edgeHandlerMap[protocol.EdgeLabels.textDocument_foldingRange] = this.handleFoldingRangeEdge
        this.edgeHandlerMap[protocol.EdgeLabels.textDocument_hover] = this.handleHoverEdge
        this.edgeHandlerMap[protocol.EdgeLabels.textDocument_references] = this.handleReferenceEdge

        try {
            fs.unlinkSync(filename)
        } catch (err) {}

        this.db = new Sqlite(filename)
        this.db.pragma('synchronous = OFF')
        this.db.pragma('journal_mode = MEMORY')

        createTables(this.db, SCHEMA_TABLES)

        this.metaInserter = new Inserter(this.db, 'meta', SCHEMA_TABLES['meta'])
        this.blobInserter = new Inserter(this.db, 'blobs', SCHEMA_TABLES['blobs'])
        this.documentInserter = new Inserter(this.db, 'documents', SCHEMA_TABLES['documents'])
        this.definitionInserter = new Inserter(this.db, 'defs', SCHEMA_TABLES['defs'])
        this.referenceInserter = new Inserter(this.db, 'refs', SCHEMA_TABLES['refs'])
        this.hoverInserter = new Inserter(this.db, 'hovers', SCHEMA_TABLES['hovers'])
    }

    public insert(element: protocol.Vertex | protocol.Edge): void {
        const handler =
            element.type === protocol.ElementTypes.vertex
                ? this.vertexHandlerMap[element.label]
                : this.edgeHandlerMap[element.label]

        if (handler) {
            handler.bind(this)(element)
        }
    }

    public close(): void {
        this.blobInserter.finish()
        this.documentInserter.finish()
        this.definitionInserter.finish()
        this.referenceInserter.finish()
        this.hoverInserter.finish()
        createIndices(this.db, SCHEMA_TABLES)
        this.db.close()
    }

    //
    // Vertex Handlers

    private handleDefinitionResult(result: protocol.DefinitionResult): void {
        this.definitionDatas.set(result.id, new Map())
    }

    private handleDocument(element: protocol.Document): void {
        this.documents.set(element.id, element)
    }

    private handleDocumentSymbols(symbols: protocol.DocumentSymbolResult): void {
        this.documentSymbols.set(symbols.id, symbols.result)
    }

    private handleFoldingRange(folding: protocol.FoldingRangeResult): void {
        this.foldingRanges.set(folding.id, folding.result)
    }

    private handleEvent(event: protocol.Event): void {
        if (event.scope === protocol.EventScope.document) {
            if (event.kind === protocol.EventKind.begin) {
                this.handleDocumentBegin(event as protocol.DocumentEvent)
            }

            if (event.kind === protocol.EventKind.end) {
                this.handleDocumentEnd(event as protocol.DocumentEvent)
            }
        }
    }

    private handleHover(hover: protocol.HoverResult): void {
        this.hoverDatas.set(hover.id, hover.result)
    }

    private handleMetaData(vertex: protocol.MetaData): void {
        this.metaInserter.do(vertex.id, JSON.stringify(vertex, undefined, 0))
    }

    private handleMoniker(moniker: protocol.Moniker): void {
        this.monikerDatas.set(moniker.id, {
            scheme: moniker.scheme,
            identifier: moniker.identifier,
            kind: moniker.kind,
        })
    }

    private handlePackageInformation(packageInformation: protocol.PackageInformation): void {
        this.packageInformationDatas.set(packageInformation.id, {
            name: packageInformation.name,
            manager: packageInformation.manager,
            uri: packageInformation.uri,
            version: packageInformation.version,
            repository: packageInformation.repository,
        })
    }

    private handleRange(range: protocol.Range): void {
        this.rangeDatas.set(range.id, { start: range.start, end: range.end, tag: range.tag })
    }

    private handleReferenceResult(result: protocol.ReferenceResult): void {
        this.referenceDatas.set(result.id, new Map())
    }

    private handleResultSet(set: protocol.ResultSet): void {
        this.resultSetDatas.set(set.id, {})
    }

    //
    // Edge Handlers

    private handleContains(contains: protocol.contains): boolean {
        let values = this.containsDatas.get(contains.outV)
        if (values === undefined) {
            values = []
            this.containsDatas.set(contains.outV, values)
        }

        values.push(...contains.inVs)
        return true
    }

    private handleItemEdge(edge: protocol.item): void {
        let property: protocol.ItemEdgeProperties | undefined = edge.property
        if (property === undefined) {
            const map: Map<protocol.Id, db.DefinitionResultData> = assertDefined(this.definitionDatas.get(edge.outV))
            let data: db.DefinitionResultData | undefined = map.get(edge.document)
            if (data === undefined) {
                data = { values: edge.inVs.slice() }
                map.set(edge.document, data)
            } else {
                data.values.push(...edge.inVs)
            }
        } else {
            const map: Map<protocol.Id, db.ReferenceResultData> = assertDefined(this.referenceDatas.get(edge.outV))
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
        assertDefined(this.rangeDatas.get(moniker.outV) || this.resultSetDatas.get(moniker.outV))
        assertDefined(this.monikerDatas.get(moniker.inV))
        this.monikerAttachments.set(moniker.outV, moniker.inV)
        this.updateMonikerSets([moniker.inV])
    }

    private handleNextEdge(edge: protocol.next): void {
        const outV: db.RangeData | db.ResultSetData = assertDefined(
            this.rangeDatas.get(edge.outV) || this.resultSetDatas.get(edge.outV)
        )
        assertDefined(this.resultSetDatas.get(edge.inV))
        outV.next = edge.inV
    }

    private handleNextMonikerEdge(nextMoniker: protocol.nextMoniker): void {
        assertDefined(this.monikerDatas.get(nextMoniker.inV))
        assertDefined(this.monikerDatas.get(nextMoniker.outV))
        this.updateMonikerSets([nextMoniker.inV, nextMoniker.outV])
    }

    private handlePackageInformationEdge(packageInformation: protocol.packageInformation): void {
        const source: db.MonikerData = assertDefined(this.monikerDatas.get(packageInformation.outV))
        const packageInfo = assertDefined(this.packageInformationDatas.get(packageInformation.inV))
        source.packageInformation = packageInformation.inV

        if (source.kind === 'export') {
            this.exportedPackages.add({ scheme: source.scheme, name: packageInfo.name, version: packageInfo.version! })
        }
    }

    private handleDefinitionEdge(edge: protocol.textDocument_definition): void {
        const outV: db.RangeData | db.ResultSetData = assertDefined(
            this.rangeDatas.get(edge.outV) || this.resultSetDatas.get(edge.outV)
        )
        this.ensureMoniker(outV)
        assertDefined(this.definitionDatas.get(edge.inV))
        outV.definitionResult = edge.inV
    }

    private handleDocumentSymbolEdge(edge: protocol.textDocument_documentSymbol): void {
        const source = assertDefined(this.getDocumentData(edge.outV))
        source.addDocumentSymbolResult(assertDefined(this.documentSymbols.get(edge.inV)))
    }

    private handleFoldingRangeEdge(edge: protocol.textDocument_foldingRange): void {
        const source = assertDefined(this.getDocumentData(edge.outV))
        source.addFoldingRangeResult(assertDefined(this.foldingRanges.get(edge.inV)))
    }

    private handleHoverEdge(edge: protocol.textDocument_hover): void {
        const outV: db.RangeData | db.ResultSetData = assertDefined(
            this.rangeDatas.get(edge.outV) || this.resultSetDatas.get(edge.outV)
        )

        this.ensureMoniker(outV)
        assertDefined(this.hoverDatas.get(edge.inV))
        outV.hoverResult = edge.inV
    }

    private handleReferenceEdge(edge: protocol.textDocument_references): void {
        const outV: db.RangeData | db.ResultSetData = assertDefined(
            this.rangeDatas.get(edge.outV) || this.resultSetDatas.get(edge.outV)
        )

        this.ensureMoniker(outV)
        assertDefined(this.referenceDatas.get(edge.inV))
        outV.referenceResult = edge.inV
    }

    //
    // Event Handlers

    private handleDocumentBegin(event: protocol.DocumentEvent) {
        const document = this.documents.get(event.data)
        if (document === undefined) {
            throw new Error(`Document with id ${event.data} not known`)
        }

        this.getOrCreateDocumentData(document)
        this.documents.delete(event.data)
    }

    private handleDocumentEnd(event: protocol.DocumentEvent) {
        for (const data of this.monikerDatas.values()) {
            if (data.kind === 'import' && data.packageInformation) {
                const packageInformation = assertDefined(this.packageInformationDatas.get(data.packageInformation!))

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

            const source: db.RangeData | db.ResultSetData = assertDefined(
                this.rangeDatas.get(key) || this.resultSetDatas.get(key)
            )
            ids.forEach(id => assertDefined(this.monikerDatas.get(id)))
            source.monikers = Array.from(ids)
        }

        const documentData = this.getEnsureDocumentData(event.data)
        const contains = this.containsDatas.get(event.data)
        if (contains !== undefined) {
            for (let id of contains) {
                const range = assertDefined(this.rangeDatas.get(id))
                documentData.addRangeData(id, range)
            }
        }

        for (const [key, value] of this.packageInformationDatas) {
            documentData.addPackageInformation(key, value)
        }

        let data = documentData.finalize()
        if (this.knownHashes.has(data.hash)) {
            // TODO - see how this can happen
            console.log('woah, duplicate?')
            this.documentDatas.set(event.id, null)
        }

        // TODO(efritz) - convert to promises
        this.blobInserter.do(data.hash, gzipSync(Buffer.from(data.blob, 'utf8')))
        this.documentInserter.do(data.hash, documentData.getUri())

        for (let definition of data.definitions || []) {
            for (let range of definition.ranges) {
                this.definitionInserter.do(
                    definition.scheme,
                    definition.indentifier,
                    data.hash,
                    range[0],
                    range[1],
                    range[2],
                    range[3]
                )
            }
        }

        for (let reference of data.references || []) {
            for (let range of reference.definitions || []) {
                this.referenceInserter.do(
                    reference.scheme,
                    reference.indentifier,
                    data.hash,
                    1,
                    range[0],
                    range[1],
                    range[2],
                    range[3]
                )
            }

            for (let range of reference.references || []) {
                this.referenceInserter.do(
                    reference.scheme,
                    reference.indentifier,
                    data.hash,
                    2,
                    range[0],
                    range[1],
                    range[2],
                    range[3]
                )
            }
        }
    }

    //
    // TODO - categorize

    private getEnsureDocumentData(id: protocol.Id): DocumentData {
        let result: DocumentData | undefined | null = this.documentDatas.get(id)
        if (result === undefined) {
            throw new Error(`No document data found for id ${id}`)
        }

        if (result === null) {
            throw new Error(`The document with Id ${id} has already been processed.`)
        }

        return result
    }

    public getMonikerData(id: protocol.Id): db.MonikerData | undefined {
        return this.monikerDatas.get(id)
    }

    public getResultData(id: protocol.Id): db.ResultSetData | undefined {
        return this.resultSetDatas.get(id)
    }

    private getOrCreateDocumentData(document: protocol.Document): DocumentData {
        let result: DocumentData | undefined | null = this.documentDatas.get(document.id)
        if (result === null) {
            throw new Error(`The document ${document.uri} has already been processed`)
        }

        result = new DocumentData(document, this)
        this.documentDatas.set(document.id, result)
        return result
    }

    public getAndDeleteDefinitions(
        definitionResult: protocol.Id,
        documentId: protocol.Id
    ): db.DefinitionResultData | undefined {
        const map = assertDefined(this.definitionDatas.get(definitionResult))
        const result = map.get(documentId)
        map.delete(documentId)
        return result
    }

    public getAndDeleteReferences(
        referenceResult: protocol.Id,
        documentId: protocol.Id
    ): db.ReferenceResultData | undefined {
        const map = assertDefined(this.referenceDatas.get(referenceResult))
        const result = map.get(documentId)
        map.delete(documentId)
        return result
    }

    public getAndDeleteHoverData(id: protocol.Id): lsp.Hover | undefined {
        let result = this.hoverDatas.get(id)
        if (result !== undefined) {
            // We don't delete the hover right now.
            // See https://github.com/microsoft/lsif-node/issues/57
            // this.hoverDatas.delete(id);
        }
        return result
    }

    private getDocumentData(id: protocol.Id): DocumentData | undefined {
        let result: DocumentData | undefined | null = this.documentDatas.get(id)
        if (result === null) {
            throw new Error(`The document with Id ${id} has already been processed.`)
        }

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

    public removeResultSetData(id: protocol.Id): void {
        this.resultSetDatas.delete(id)
    }

    public removeMonikerData(id: protocol.Id): void {
        this.monikerDatas.delete(id)
    }

    private ensureMoniker(data: db.RangeData | db.ResultSetData): void {
        if (data.monikers !== undefined) {
            return
        }

        const monikerData: db.MonikerData = { scheme: '$synthetic', identifier: uuid.v4() }
        data.monikers = [monikerData.identifier]
        this.monikerDatas.set(monikerData.identifier, monikerData)
    }

    public storeHover(moniker: db.MonikerData, id: protocol.Id): void {
        let hover = this.getAndDeleteHoverData(id)
        if (hover === undefined) {
            // We have already processed the hover
            return
        }

        const blob = JSON.stringify(hover, undefined, 0)
        const blobHash = crypto
            .createHash('md5')
            .update(blob)
            .digest('base64')

        if (!this.knownHashes.has(blobHash)) {
            if (!this.insertedBlobs.has(blobHash)) {
                // TODO(efritz) - convert to promises
                this.blobInserter.do(blobHash, gzipSync(Buffer.from(blob, 'utf8')))
                this.insertedBlobs.add(blobHash)
            }

            const hoverHash = crypto
                .createHash('md5')
                .update(JSON.stringify({ s: moniker.scheme, i: moniker.identifier, b: blobHash }, undefined, 0))
                .digest('base64')

            if (!this.insertedHovers.has(hoverHash)) {
                this.hoverInserter.do(moniker.scheme, moniker.identifier, blobHash)
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

    constructor(document: protocol.Document, provider: BlobStore) {
        this.provider = provider
        this.id = document.id
        this.uri = document.uri
    }

    public getUri(): string {
        return this.uri
    }

    public addRangeData(id: protocol.Id, data: db.RangeData): void {
        this.blob.ranges[id] = data
        this.addReferencedData(id, data)
    }

    private addResultSetData(id: protocol.Id, resultSet: db.ResultSetData): void {
        if (this.blob.resultSets === undefined) {
            this.blob.resultSets = LiteralMap.create()
        }

        // Many ranges can point to the same result set. Make sure
        // we only travers once.
        if (this.blob.resultSets[id] !== undefined) {
            return
        }

        this.blob.resultSets[id] = resultSet
        this.addReferencedData(id, resultSet)
    }

    private addReferencedData(id: protocol.Id, item: db.RangeData | db.ResultSetData): void {
        const monikers = []

        if (item.monikers !== undefined) {
            for (const itemMoniker of item.monikers) {
                const moniker = assertDefined(this.provider.getMonikerData(itemMoniker))
                this.addMoniker(itemMoniker, moniker)
                monikers.push(moniker)
            }
        }

        if (item.next !== undefined) {
            this.addResultSetData(item.next, assertDefined(this.provider.getResultData(item.next)))
        }

        for (const moniker of monikers) {
            if (item.hoverResult !== undefined) {
                if (Monikers.isLocal(moniker)) {
                    if (this.blob.hovers === undefined) {
                        this.blob.hovers = LiteralMap.create()
                    }

                    if (this.blob.hovers[item.hoverResult] === undefined) {
                        this.blob.hovers[item.hoverResult] = assertDefined(
                            this.provider.getAndDeleteHoverData(item.hoverResult)
                        )
                    }
                } else {
                    this.provider.storeHover(moniker, item.hoverResult)
                }
            }

            if (item.definitionResult) {
                const definitions = this.provider.getAndDeleteDefinitions(item.definitionResult, this.id)
                if (definitions !== undefined) {
                    if (Monikers.isLocal(moniker)) {
                        if (this.blob.definitionResults === undefined) {
                            this.blob.definitionResults = LiteralMap.create()
                        }
                        this.blob.definitionResults[item.definitionResult] = definitions
                    } else {
                        this.definitions.push({ moniker, data: definitions })
                    }
                }
            }

            if (item.referenceResult) {
                const references = this.provider.getAndDeleteReferences(item.referenceResult, this.id)
                if (references !== undefined) {
                    if (Monikers.isLocal(moniker)) {
                        if (this.blob.referenceResults === undefined) {
                            this.blob.referenceResults = LiteralMap.create()
                        }
                        this.blob.referenceResults[item.referenceResult] = references
                    } else {
                        this.references.push({ moniker, data: references })
                    }
                }
            }
        }
    }

    public addFoldingRangeResult(value: lsp.FoldingRange[]): void {
        this.blob.foldingRanges = value
    }

    public addDocumentSymbolResult(value: lsp.DocumentSymbol[] | protocol.RangeBasedDocumentSymbol[]): void {
        this.blob.documentSymbols = value
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

    public finalize(): DocumentDatabaseData {
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

        return {
            hash: this.computeHash(),
            blob: JSON.stringify(this.blob, undefined, 0),
            definitions: externalDefinitions.length > 0 ? externalDefinitions : undefined,
            references: externalReferences.length > 0 ? externalReferences : undefined,
        }
    }

    private computeHash(): string {
        const hash = crypto.createHash('md5')
        hash.update(JSON.stringify(this.blob))
        return hash.digest('base64')
    }
}

function assertDefined<T>(value: T | undefined | null): T {
    if (value === undefined || value === null) {
        throw new Error(`Element must be defined`)
    }
    return value
}

namespace LiteralMap {
    export function create<T = any>(): db.LiteralMap<T> {
        return Object.create(null)
    }

    export function values<T>(map: db.LiteralMap<T>): T[] {
        let result: T[] = []
        for (let key of Object.keys(map)) {
            result.push(map[key])
        }
        return result
    }
}

namespace Strings {
    export function compare(s1: string, s2: string): number {
        return s1 == s2 ? 0 : s1 > s2 ? 1 : -1
    }
}

namespace Monikers {
    export function compare(m1: db.MonikerData, m2: db.MonikerData): number {
        let result = Strings.compare(m1.identifier, m2.identifier)
        if (result !== 0) {
            return result
        }
        result = Strings.compare(m1.scheme, m2.scheme)
        if (result !== 0) {
            return result
        }
        if (m1.kind === m2.kind) {
            return 0
        }
        const k1 = m1.kind !== undefined ? m1.kind : protocol.MonikerKind.import
        const k2 = m2.kind !== undefined ? m2.kind : protocol.MonikerKind.import
        if (k1 === protocol.MonikerKind.import && k2 === protocol.MonikerKind.export) {
            return -1
        }
        if (k1 === protocol.MonikerKind.export && k2 === protocol.MonikerKind.import) {
            return 1
        }
        return 0
    }

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
    blob: string
    definitions?: ExternalDefinition[]
    references?: ExternalReference[]
}

interface MonikerScopedResultData<T> {
    moniker: db.MonikerData
    data: T
}
