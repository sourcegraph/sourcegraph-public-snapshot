/* --------------------------------------------------------------------------------------------
 * Copyright (c) Sourcegraph and Microsoft Corporation. All rights reserved.
 * Licensed under the MIT License.
 * ------------------------------------------------------------------------------------------ */

import Sqlite from 'better-sqlite3'
import { URI } from 'vscode-uri'
import { Range, Id, MetaData, RangeBasedDocumentSymbol, Moniker, PackageInformation } from 'lsif-protocol'
import * as lsp from 'vscode-languageserver-protocol'
import { CorrelationDatabase } from './sqlite.xrepo'
import { makeFilename } from './sqlite'

export interface DocumentBlob {
    contents: string
    ranges: LiteralMap<RangeData>
    resultSets?: LiteralMap<ResultSetData>
    monikers?: LiteralMap<MonikerData>
    packageInformation?: LiteralMap<PackageInformationData>
    hovers?: LiteralMap<lsp.Hover>
    declarationResults?: LiteralMap<DeclarationResultData>
    definitionResults?: LiteralMap<DefinitionResultData>
    referenceResults?: LiteralMap<ReferenceResultData>
    foldingRanges?: lsp.FoldingRange[]
    documentSymbols?: lsp.DocumentSymbol[] | RangeBasedDocumentSymbol[]
    diagnostics?: lsp.Diagnostic[]
}

export interface RangeData extends Pick<Range, 'start' | 'end' | 'tag'> {
    monikers?: Id[]
    next?: Id
    hoverResult?: Id
    declarationResult?: Id
    definitionResult?: Id
    referenceResult?: Id
}

export interface ResultSetData {
    monikers?: Id[]
    next?: Id
    hoverResult?: Id
    declarationResult?: Id
    definitionResult?: Id
    referenceResult?: Id
}

export interface DeclarationResultData {
    values: Id[]
}

export interface DefinitionResultData {
    values: Id[]
}

export interface ReferenceResultData {
    declarations?: Id[]
    definitions?: Id[]
    references?: Id[]
}

export type MonikerData = Pick<Moniker, 'scheme' | 'identifier' | 'kind'> & {
    packageInformation?: Id
}

export type PackageInformationData = Pick<
    PackageInformation,
    'name' | 'manager' | 'uri' | 'contents' | 'version' | 'repository'
>

export interface LiteralMap<T> {
    [key: string]: T
    [key: number]: T
}

const MONIKER_KIND_PREFERENCES = ['import', 'local', 'export']
const MONIKER_SCHEME_PREFERENCES = ['npm', 'tsc']

interface MetaDataResult {
    id: number
    value: string
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

interface RefsResult {
    uri: string
    kind: number
    startLine: number
    startCharacter: number
    endLine: number
    endCharacter: number
}

export interface UriTransformer {
    toDatabase(uri: string): string
    fromDatabase(uri: string): string
}

export const noopTransformer: UriTransformer = {
    toDatabase: uri => uri,
    fromDatabase: uri => uri,
}

export class Database {
    private uriTransformer!: UriTransformer
    private db!: Sqlite.Database

    private allDocumentsStmt!: Sqlite.Statement
    private findDocumentStmt!: Sqlite.Statement
    private findBlobStmt!: Sqlite.Statement
    private findDefsStmt!: Sqlite.Statement
    private findRefsStmt!: Sqlite.Statement
    private findHoverStmt!: Sqlite.Statement

    private version!: string
    private blobs: Map<Id, DocumentBlob>

    public constructor(private correlationDb: CorrelationDatabase) {
        this.blobs = new Map()
    }

    public load(file: string): Promise<void> {
        const transformerFactory: (projectRoot: string) => UriTransformer = root => ({
            toDatabase: path => `${root}/${path}`,
            fromDatabase: uri => (uri.startsWith(root) ? uri.slice(`${root}/`.length) : uri),
        })

        this.db = new Sqlite(file, { readonly: true })
        let result: MetaDataResult[] = this.db.prepare('Select * from meta').all()
        if (result === undefined || result.length !== 1) {
            throw new Error('Failed to read meta data record.')
        }
        let metaData: MetaData = JSON.parse(result[0].value)
        if (metaData.projectRoot === undefined) {
            throw new Error('No project root provided.')
        }
        const projectRoot = URI.parse(metaData.projectRoot).toString(true)

        this.allDocumentsStmt = this.db.prepare(
            'Select d.* FROM documents d Inner Join versions v On v.hash = d.documentHash'
        )

        this.findDocumentStmt = this.db.prepare(
            [
                'Select d.documentHash From documents d',
                'Inner Join versions v On v.hash = d.documentHash',
                'Where v.version = $version and d.uri = $uri',
            ].join(' ')
        )
        this.findBlobStmt = this.db.prepare('Select content From blobs Where hash = ?')
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

        this.uriTransformer = transformerFactory ? transformerFactory(projectRoot) : noopTransformer
        return Promise.resolve()
    }

    public close(): void {
        this.db.close()
    }

    public exists(uri: string): boolean {
        return Boolean(this.findFile(this.uriTransformer.toDatabase(uri)))
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

    public definitions(uri: string, position: lsp.Position): lsp.Location | lsp.Location[] | undefined {
        const { range, blob } = this.findRangeFromPosition(this.toDatabase(uri), position)
        if (range === undefined || blob === undefined || blob.definitionResults === undefined) {
            return undefined
        }

        let resultData = this.findResult(blob.resultSets, blob.definitionResults, range, 'definitionResult')
        if (resultData === undefined) {
            const monikers = this.findMonikers(blob.resultSets, blob.monikers, range)

            for (const moniker of monikers) {
                
                if (moniker.kind === 'import') {
                    // TODO - clean this up
                    // for (const moniker of monikers) {
                        if (moniker.packageInformation) {
                            const packageInformation = blob.packageInformation![moniker.packageInformation]
                            const result = this.correlationDb.lookup(
                                moniker.scheme,
                                packageInformation.name,
                                packageInformation.version!
                            )

                            if (result) {
                                const { repository, commit } = result

                                // TODO(efritz) - use cached handle if open
                                const subDb = new Database(this.correlationDb)
                                subDb.load(makeFilename(repository, commit))

                                try {
                                    // TODO(efritz) - determine why npm monikers are mismatched
                                    const parts = moniker.identifier.split(':')
                                    parts[1] = '' // WTF
                                    moniker.identifier = parts.join(':')

                                    // TODO(efritz) - make this indexable in db
                                    for (const qResult of subDb.allDocumentsStmt.all()) {
                                        const blob = subDb.getBlob(qResult.documentHash)
                                        for (const id of Object.keys(blob.monikers!)) {
                                            const m = blob.monikers![id]
                                            // TODO(efritz) - skip non-definitions?
                                            if (m.scheme === moniker.scheme && m.identifier === moniker.identifier) {
                                                for (const otherId of Object.keys(blob.ranges)) {
                                                    const range = blob.ranges[otherId]
                                                    const monikers = subDb.findMonikers(
                                                        blob.resultSets,
                                                        blob.monikers,
                                                        range
                                                    )

                                                    if (monikers.includes(m)) {
                                                        const subUri = subDb.fromDatabase(qResult.uri)

                                                        return [
                                                            lsp.Location.create(
                                                                `git://${repository}?${commit}#${subUri}`,
                                                                lsp.Range.create(
                                                                    range.start.line,
                                                                    range.start.character,
                                                                    range.end.line,
                                                                    range.end.character
                                                                )
                                                            ),
                                                        ]
                                                    }
                                                }
                                            }
                                        }
                                    }
                                } finally {
                                    subDb.close()
                                }
                            }
                        }
                    // }

                    return undefined
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
            return asLocations(blob.ranges, uri, resultData.values)
        }
    }

    public references(uri: string, position: lsp.Position, context: lsp.ReferenceContext): lsp.Location[] | undefined {
        const { range, blob } = this.findRangeFromPosition(this.toDatabase(uri), position)
        if (range === undefined || blob === undefined || blob.referenceResults === undefined) {
            return undefined
        }
        let resultData = this.findResult(blob.resultSets, blob.referenceResults, range, 'referenceResult')
        console.log('ref data:', resultData)
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
                result.push(...asLocations(blob.ranges, uri, resultData.declarations))
            }
            if (context.includeDeclaration && resultData.definitions !== undefined) {
                result.push(...asLocations(blob.ranges, uri, resultData.definitions))
            }
            if (resultData.references !== undefined) {
                result.push(...asLocations(blob.ranges, uri, resultData.references))
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
    ): MonikerData[] {
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
            if (containsPosition(range, position)) {
                if (!candidate) {
                    candidate = range
                } else {
                    if (containsRange(candidate, range)) {
                        candidate = range
                    }
                }
            }
        }
        return { range: candidate, blob }
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

    private findFile(uri: string): Id | undefined {
        let result: DocumentResult = this.findDocumentStmt.get({ version: this.version, uri: uri })
        return result !== undefined ? result.documentHash : undefined
    }

    private toDatabase(uri: string): string {
        return this.uriTransformer.toDatabase(uri)
    }

    private fromDatabase(uri: string): string {
        return this.uriTransformer.fromDatabase(uri)
    }
}

function asLocations(ranges: LiteralMap<RangeData>, uri: string, ids: Id[]): lsp.Location[] {
    return ids.map(id => {
        let range = ranges[id]
        return lsp.Location.create(
            uri,
            lsp.Range.create(range.start.line, range.start.character, range.end.line, range.end.character)
        )
    })
}

function containsPosition(range: lsp.Range, position: lsp.Position): boolean {
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
function containsRange(range: lsp.Range, otherRange: lsp.Range): boolean {
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
