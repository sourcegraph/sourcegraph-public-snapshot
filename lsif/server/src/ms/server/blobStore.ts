/* --------------------------------------------------------------------------------------------
 * Copyright (c) Microsoft Corporation. All rights reserved.
 * Licensed under the MIT License. See License.txt in the project root for license information.
 * ------------------------------------------------------------------------------------------ */

import * as lsp from 'vscode-languageserver'
import Sqlite from 'better-sqlite3'
import { Database, UriTransformer } from './database'
import { DocumentInfo } from './files'
import { Id, MetaData, Moniker, Range, RangeBasedDocumentSymbol } from 'lsif-protocol'
import { URI } from 'vscode-uri'

const MONIKER_KIND_PREFERENCES = ['import', 'local', 'export']
const MONIKER_SCHEME_PREFERENCES = ['npm', 'tsc']

interface MetaDataResult {
    id: number
    value: string
}

interface LiteralMap<T> {
    [key: string]: T
    [key: number]: T
}

interface RangeData extends Pick<Range, 'start' | 'end' | 'tag'> {
    monikers?: Id[]
    next?: Id
    hoverResult?: Id
    declarationResult?: Id
    definitionResult?: Id
    referenceResult?: Id
}

interface ResultSetData {
    monikers?: Id[]
    next?: Id
    hoverResult?: Id
    declarationResult?: Id
    definitionResult?: Id
    referenceResult?: Id
}

interface DeclarationResultData {
    values: Id[]
}

interface DefinitionResultData {
    values: Id[]
}

interface ReferenceResultData {
    declarations?: Id[]
    definitions?: Id[]
    references?: Id[]
}

type MonikerData = Pick<Moniker, 'scheme' | 'identifier' | 'kind'>

interface DocumentBlob {
    contents: string
    ranges: LiteralMap<RangeData>
    resultSets?: LiteralMap<ResultSetData>
    monikers?: LiteralMap<MonikerData>
    hovers?: LiteralMap<lsp.Hover>
    declarationResults?: LiteralMap<DeclarationResultData>
    definitionResults?: LiteralMap<DefinitionResultData>
    referenceResults?: LiteralMap<ReferenceResultData>
    foldingRanges?: lsp.FoldingRange[]
    documentSymbols?: lsp.DocumentSymbol[] | RangeBasedDocumentSymbol[]
    diagnostics?: lsp.Diagnostic[]
}

interface DocumentsResult {
    documentHash: string
    uri: string
}

interface BlobResult {
    content: Buffer
}

interface DocumentResult {
    documentHash: string
}

interface DefsResult {
    uri: string
    startLine: number
    startCharacter: number
    endLine: number
    endCharacter: number
}

interface DeclsResult {
    uri: string
    startLine: number
    startCharacter: number
    endLine: number
    endCharacter: number
}

interface RefsResult {
    uri: string
    kind: number
    startLine: number
    startCharacter: number
    endLine: number
    endCharacter: number
}

export class BlobStore extends Database {
    private db!: Sqlite.Database

    private allDocumentsStmt!: Sqlite.Statement
    private findDocumentStmt!: Sqlite.Statement
    private findBlobStmt!: Sqlite.Statement
    private findDeclsStmt!: Sqlite.Statement
    private findDefsStmt!: Sqlite.Statement
    private findRefsStmt!: Sqlite.Statement
    private findHoverStmt!: Sqlite.Statement

    private version!: string
    private projectRoot!: URI
    private blobs: Map<Id, DocumentBlob>

    public constructor() {
        super()
        this.version
        this.blobs = new Map()
    }

    public load(file: string, transformerFactory: (projectRoot: string) => UriTransformer): Promise<void> {
        this.db = new Sqlite(file, { readonly: true })
        this.readMetaData()
        this.allDocumentsStmt = this.db.prepare(
            [
                'Select d.documentHash, d.uri From documents d',
                'Inner Join versions v On v.hash = d.documentHash',
                'Where v.version = ?',
            ].join(' ')
        )
        this.findDocumentStmt = this.db.prepare(
            [
                'Select d.documentHash From documents d',
                'Inner Join versions v On v.hash = d.documentHash',
                'Where v.version = $version and d.uri = $uri',
            ].join(' ')
        )
        this.findBlobStmt = this.db.prepare('Select content From blobs Where hash = ?')
        this.findDeclsStmt = this.db.prepare(
            [
                'Select doc.uri, d.startLine, d.startCharacter, d.endLine, d.endCharacter From decls d',
                'Inner Join versions v On d.documentHash = v.hash',
                'Inner Join documents doc On d.documentHash = doc.documentHash',
                'Where v.version = $version and d.scheme = $scheme and d.identifier = $identifier',
            ].join(' ')
        )
        this.findDefsStmt = this.db.prepare(
            [
                'Select doc.uri, d.startLine, d.startCharacter, d.endLine, d.endCharacter From defs d',
                'Inner Join versions v On d.documentHash = v.hash',
                'Inner Join documents doc On d.documentHash = doc.documentHash',
                'Where v.version = $version and d.scheme = $scheme and d.identifier = $identifier',
            ].join(' ')
        )
        this.findRefsStmt = this.db.prepare(
            [
                'Select doc.uri, r.kind, r.startLine, r.startCharacter, r.endLine, r.endCharacter From refs r',
                'Inner Join versions v On r.documentHash = v.hash',
                'Inner Join documents doc On r.documentHash = doc.documentHash',
                'Where v.version = $version and r.scheme = $scheme and r.identifier = $identifier',
            ].join(' ')
        )
        this.findHoverStmt = this.db.prepare(
            [
                'Select b.content From blobs b',
                'Inner Join versions v On b.hash = v.hash',
                'Inner Join hovers h On h.hoverHash = b.hash',
                'Where v.version = $version and h.scheme = $scheme and h.identifier = $identifier',
            ].join(' ')
        )
        this.version = this.db.prepare('Select * from versionTags Order by dateTime desc').get().tag
        if (typeof this.version !== 'string') {
            throw new Error('Version tag must be a string')
        }
        this.initialize(transformerFactory)
        return Promise.resolve()
    }

    private readMetaData(): void {
        let result: MetaDataResult[] = this.db.prepare('Select * from meta').all()
        if (result === undefined || result.length !== 1) {
            throw new Error('Failed to read meta data record.')
        }
        let metaData: MetaData = JSON.parse(result[0].value)
        if (metaData.projectRoot === undefined) {
            throw new Error('No project root provided.')
        }
        this.projectRoot = URI.parse(metaData.projectRoot)
    }

    public getProjectRoot(): URI {
        return this.projectRoot
    }

    public close(): void {
        this.db.close()
    }

    protected getDocumentInfos(): DocumentInfo[] {
        let result: DocumentsResult[] = this.allDocumentsStmt.all(this.version)
        if (result === undefined) {
            return []
        }
        return result.map(item => {
            return { id: item.documentHash, uri: item.uri }
        })
    }

    private getBlob(documentId: Id): DocumentBlob {
        let result = this.blobs.get(documentId)
        if (result === undefined) {
            const blobResult: BlobResult = this.findBlobStmt.get(documentId)
            result = JSON.parse(blobResult.content.toString('utf8')) as DocumentBlob
            this.blobs.set(documentId, result)
        }
        return result
    }

    protected findFile(uri: string): Id | undefined {
        let result: DocumentResult = this.findDocumentStmt.get({ version: this.version, uri: uri })
        return result !== undefined ? result.documentHash : undefined
    }

    protected fileContent(documentId: Id): string {
        const blob = this.getBlob(documentId)
        return Buffer.from(blob.contents).toString('base64')
    }

    public foldingRanges(uri: string): lsp.FoldingRange[] | undefined {
        return undefined
    }

    public documentSymbols(uri: string): lsp.DocumentSymbol[] | undefined {
        return undefined
    }

    public hover(uri: string, position: lsp.Position): lsp.Hover | undefined {
        const { range, blob } = this.findRangeFromPosition(this.toDatabase(uri), position)
        if (range === undefined || blob === undefined || blob.hovers === undefined) {
            return undefined
        }
        let result = this.findResult(blob.resultSets, blob.hovers, range, 'hoverResult')
        if (result !== undefined) {
            return result
        }
        const monikers = this.findMonikers(blob.resultSets, blob.monikers, range)
        for (const moniker of monikers) {
            const qResult: BlobResult = this.findHoverStmt.get({
                version: this.version,
                scheme: moniker.scheme,
                identifier: moniker.identifier,
            })
            if (qResult === undefined) {
                continue
            }
            result = JSON.parse(qResult.content.toString()) as lsp.Hover
            if (result.range === undefined) {
                result.range = lsp.Range.create(
                    range.start.line,
                    range.start.character,
                    range.end.line,
                    range.end.character
                )
            }
            return result
        }

        return undefined
    }

    public declarations(uri: string, position: lsp.Position): lsp.Location | lsp.Location[] | undefined {
        const { range, blob } = this.findRangeFromPosition(this.toDatabase(uri), position)
        if (range === undefined || blob === undefined || blob.declarationResults === undefined) {
            return undefined
        }
        let resultData = this.findResult(blob.resultSets, blob.declarationResults, range, 'declarationResult')
        if (resultData === undefined) {
            for (const moniker of this.findMonikers(blob.resultSets, blob.monikers, range)) {
                let qResult: DeclsResult[] = this.findDeclsStmt.all({
                    version: this.version,
                    scheme: moniker.scheme,
                    identifier: moniker.identifier,
                })
                if (qResult === undefined || qResult.length === 0) {
                    continue
                }
                return qResult.map(item => {
                    return lsp.Location.create(
                        this.fromDatabase(item.uri),
                        lsp.Range.create(item.startLine, item.startCharacter, item.endLine, item.endCharacter)
                    )
                })
            }

            return undefined
        } else {
            return BlobStore.asLocations(blob.ranges, uri, resultData.values)
        }
    }

    public definitions(uri: string, position: lsp.Position): lsp.Location | lsp.Location[] | undefined {
        const { range, blob } = this.findRangeFromPosition(this.toDatabase(uri), position)
        if (range === undefined || blob === undefined || blob.definitionResults === undefined) {
            return undefined
        }

        let resultData = this.findResult(blob.resultSets, blob.definitionResults, range, 'definitionResult')
        if (resultData === undefined) {
            const monikers = this.findMonikers(blob.resultSets, blob.monikers, range)

            for (const moniker of monikers) {
                if (moniker.kind === "import") {
                    // TODO(efritz) - implement xrepo find defs here
                    console.log("UNIMPLEMENTED")
                    console.log("MONIKERS:", monikers)
                    return undefined;
                }

                let qResult: DefsResult[] = this.findDefsStmt.all({
                    version: this.version,
                    scheme: moniker.scheme,
                    identifier: moniker.identifier,
                })
                if (qResult === undefined || qResult.length === 0) {
                    continue
                }

                return qResult.map(item => {
                    return lsp.Location.create(
                        this.fromDatabase(item.uri),
                        lsp.Range.create(item.startLine, item.startCharacter, item.endLine, item.endCharacter)
                    )
                })
            }

            return undefined
        } else {
            return BlobStore.asLocations(blob.ranges, uri, resultData.values)
        }
    }

    public references(uri: string, position: lsp.Position, context: lsp.ReferenceContext): lsp.Location[] | undefined {
        const { range, blob } = this.findRangeFromPosition(this.toDatabase(uri), position)
        if (range === undefined || blob === undefined || blob.referenceResults === undefined) {
            return undefined
        }
        let resultData = this.findResult(blob.resultSets, blob.referenceResults, range, 'referenceResult')
        if (resultData === undefined) {
            for (const moniker of this.findMonikers(blob.resultSets, blob.monikers, range)) {
                let qResult: RefsResult[] = this.findRefsStmt.all({
                    version: this.version,
                    scheme: moniker.scheme,
                    identifier: moniker.identifier,
                })
                if (qResult === undefined || qResult.length === 0) {
                    continue
                }
                let result: lsp.Location[] = []
                for (let item of qResult) {
                    if (context.includeDeclaration || item.kind === 2) {
                        result.push(
                            lsp.Location.create(
                                this.fromDatabase(item.uri),
                                lsp.Range.create(item.startLine, item.startCharacter, item.endLine, item.endCharacter)
                            )
                        )
                    }
                }
                return result
            }
            return undefined
        } else {
            let result: lsp.Location[] = []
            if (context.includeDeclaration && resultData.declarations !== undefined) {
                result.push(...BlobStore.asLocations(blob.ranges, uri, resultData.declarations))
            }
            if (context.includeDeclaration && resultData.definitions !== undefined) {
                result.push(...BlobStore.asLocations(blob.ranges, uri, resultData.definitions))
            }
            if (resultData.references !== undefined) {
                result.push(...BlobStore.asLocations(blob.ranges, uri, resultData.references))
            }
            return result
        }
    }

    private findResult<T>(
        resultSets: LiteralMap<ResultSetData> | undefined,
        map: LiteralMap<T>,
        data: RangeData | ResultSetData,
        property: 'next' | 'hoverResult' | 'declarationResult' | 'definitionResult' | 'referenceResult'
    ): T | undefined {
        let current: RangeData | ResultSetData | undefined = data
        while (current !== undefined) {
            let value = current[property]
            if (value !== undefined) {
                return map[value]
            }
            current =
                current.next !== undefined
                    ? resultSets !== undefined
                        ? resultSets[current.next]
                        : undefined
                    : undefined
        }
        return undefined
    }

    private findMonikers(
        resultSets: LiteralMap<ResultSetData> | undefined,
        monikers: LiteralMap<MonikerData> | undefined,
        data: RangeData | ResultSetData
    ): MonikerData[]  {
        if (monikers === undefined) {
            return []
        }

        let current: RangeData | ResultSetData | undefined = data
        const ids = []
        while (current !== undefined) {
            if (current.monikers !== undefined) {
                for (const id of current.monikers) {
                    ids.push(id)
                }
            }

            current =
                current.next !== undefined
                    ? resultSets !== undefined
                        ? resultSets[current.next]
                        : undefined
                    : undefined
        }

        const resultMonikers = ids.map(id => monikers[id])

        resultMonikers.sort((a, b) => {
            const ord = MONIKER_KIND_PREFERENCES.indexOf(a.kind!) - MONIKER_KIND_PREFERENCES.indexOf(b.kind!)
            if (ord !== 0) {
                return ord
            }

            return MONIKER_SCHEME_PREFERENCES.indexOf(a.scheme!) - MONIKER_SCHEME_PREFERENCES.indexOf(b.scheme!)
        })

        return resultMonikers
    }

    private findRangeFromPosition(
        uri: string,
        position: lsp.Position
    ): { range: RangeData | undefined; blob: DocumentBlob | undefined } {
        const documentId = this.findFile(uri)
        if (documentId === undefined) {
            return { range: undefined, blob: undefined }
        }
        const blob = this.getBlob(documentId)
        let candidate: RangeData | undefined
        for (let key of Object.keys(blob.ranges)) {
            let range = blob.ranges[key]
            if (BlobStore.containsPosition(range, position)) {
                if (!candidate) {
                    candidate = range
                } else {
                    if (BlobStore.containsRange(candidate, range)) {
                        candidate = range
                    }
                }
            }
        }
        return { range: candidate, blob }
    }

    private static asLocations(ranges: LiteralMap<RangeData>, uri: string, ids: Id[]): lsp.Location[] {
        return ids.map(id => {
            let range = ranges[id]
            return lsp.Location.create(
                uri,
                lsp.Range.create(range.start.line, range.start.character, range.end.line, range.end.character)
            )
        })
    }

    private static containsPosition(range: lsp.Range, position: lsp.Position): boolean {
        if (position.line < range.start.line || position.line > range.end.line) {
            return false
        }
        if (position.line === range.start.line && position.character < range.start.character) {
            return false
        }
        if (position.line === range.end.line && position.character > range.end.character) {
            return false
        }
        return true
    }

    /**
     * Test if `otherRange` is in `range`. If the ranges are equal, will return true.
     */
    public static containsRange(range: lsp.Range, otherRange: lsp.Range): boolean {
        if (otherRange.start.line < range.start.line || otherRange.end.line < range.start.line) {
            return false
        }
        if (otherRange.start.line > range.end.line || otherRange.end.line > range.end.line) {
            return false
        }
        if (otherRange.start.line === range.start.line && otherRange.start.character < range.start.character) {
            return false
        }
        if (otherRange.end.line === range.end.line && otherRange.end.character > range.end.character) {
            return false
        }
        return true
    }
}
