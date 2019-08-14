import Sqlite from 'better-sqlite3'
import { URI } from 'vscode-uri'
import { Range, Id, MetaData, RangeBasedDocumentSymbol, Moniker, PackageInformation } from 'lsif-protocol'
import * as lsp from 'vscode-languageserver-protocol'
import { CorrelationDatabase } from './xrepo'
import { makeFilename } from './backend'
import { gunzipSync } from 'mz/zlib'

export interface DocumentBlob {
    ranges: LiteralMap<RangeData>
    resultSets?: LiteralMap<ResultSetData>
    monikers?: LiteralMap<MonikerData>
    packageInformation?: LiteralMap<PackageInformationData>
    hovers?: LiteralMap<lsp.Hover>
    definitionResults?: LiteralMap<DefinitionResultData>
    referenceResults?: LiteralMap<ReferenceResultData>
    foldingRanges?: lsp.FoldingRange[]
    documentSymbols?: lsp.DocumentSymbol[] | RangeBasedDocumentSymbol[]
}

export interface RangeData extends Pick<Range, 'start' | 'end' | 'tag'> {
    monikers?: Id[]
    next?: Id
    hoverResult?: Id
    definitionResult?: Id
    referenceResult?: Id
}

export interface ResultSetData {
    monikers?: Id[]
    next?: Id
    hoverResult?: Id
    definitionResult?: Id
    referenceResult?: Id
}

export interface DefinitionResultData {
    values: Id[]
}

export interface ReferenceResultData {
    definitions?: Id[]
    references?: Id[]
}

export type MonikerData = Pick<Moniker, 'scheme' | 'identifier' | 'kind'> & {
    packageInformation?: Id
}

export type PackageInformationData = Pick<PackageInformation, 'name' | 'manager' | 'uri' | 'version' | 'repository'>

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

export class Database {
    private db!: Sqlite.Database
    private uriTransformer!: UriTransformer
    private findDocumentStmt!: Sqlite.Statement
    private findBlobStmt!: Sqlite.Statement
    private findDefsStmt!: Sqlite.Statement
    private findRefsStmt!: Sqlite.Statement
    private findHoverStmt!: Sqlite.Statement

    // TODO - put this in a (shared) LRU cache as well
    private blobs: Map<Id, DocumentBlob> = new Map()

    public constructor(private correlationDb: CorrelationDatabase) {}

    public load(file: string): Promise<void> {
        this.db = new Sqlite(file, { readonly: true })

        const root = this.getProjectRoot()

        this.uriTransformer = {
            toDatabase: path => `${root}/${path}`,
            fromDatabase: uri => (uri.startsWith(root) ? uri.slice(`${root}/`.length) : uri),
        }

        // TODO - find a way to declare these as well?
        this.findDocumentStmt = this.db.prepare('select d.documentHash from documents d where d.uri = $uri')
        this.findBlobStmt = this.db.prepare('select content from blobs where hash = ?')
        this.findDefsStmt = this.db.prepare(
            [
                'select doc.uri, d.startLine, d.startCharacter, d.endLine, d.endCharacter from defs d',
                'inner join documents doc on d.documentHash = doc.documentHash',
                'where d.scheme = $scheme and d.identifier = $identifier',
            ].join(' ')
        )
        this.findRefsStmt = this.db.prepare(
            [
                'select doc.uri, r.kind, r.startLine, r.startCharacter, r.endLine, r.endCharacter from refs r',
                'inner join documents doc on r.documentHash = doc.documentHash',
                'where r.scheme = $scheme and r.identifier = $identifier',
            ].join(' ')
        )
        this.findHoverStmt = this.db.prepare(
            [
                'select b.content from blobs b',
                'inner join hovers h on h.hoverHash = b.hash',
                'where h.scheme = $scheme and h.identifier = $identifier',
            ].join(' ')
        )

        return Promise.resolve()
    }

    private getProjectRoot(): string {
        const result: MetaDataResult[] = this.db.prepare('select * from meta').all()
        if (result && result.length === 1) {
            const metaData: MetaData = JSON.parse(result[0].value)
            if (metaData.projectRoot) {
                return URI.parse(metaData.projectRoot).toString(true)
            }
        }

        throw new Error('Failed to get project root from meta.')
    }

    public close(): void {
        this.db.close()
    }

    //
    // Exists

    public exists(uri: string): boolean {
        return Boolean(this.findFile(this.uriTransformer.toDatabase(uri)))
    }

    //
    // Definitions

    public definitions(uri: string, position: lsp.Position): lsp.Location | lsp.Location[] | undefined {
        const { range, blob } = this.findRangeFromPosition(this.uriTransformer.toDatabase(uri), position)
        if (!range || !blob || !blob.definitionResults) {
            return undefined
        }

        const resultData = this.findResult(blob.resultSets, blob.definitionResults, range, 'definitionResult')
        if (resultData) {
            return asLocations(blob.ranges, uri, resultData.values)
        }

        for (const moniker of this.findMonikers(blob.resultSets, blob.monikers, range)) {
            if (moniker.kind === 'import') {
                return this.getRemoteDefinition(blob, moniker)
            }

            // TODO(efritz) - prove that this returns something useful
            // in some circumstances. I'm not sure if this was meant to
            // do what we're trying to do now...

            const defsResult: DefsResult[] = this.findDefsStmt.all({
                scheme: moniker.scheme,
                identifier: moniker.identifier,
            })

            if (defsResult && defsResult.length > 0) {
                return defsResult.map(item => {
                    return lsp.Location.create(
                        this.uriTransformer.fromDatabase(item.uri),
                        lsp.Range.create(item.startLine, item.startCharacter, item.endLine, item.endCharacter)
                    )
                })
            }
        }

        return undefined
    }

    private getRemoteDefinition(blob: DocumentBlob, moniker: MonikerData): lsp.Location | lsp.Location[] | undefined {
        if (!moniker.packageInformation) {
            return undefined
        }

        const packageInformation = blob.packageInformation![moniker.packageInformation]
        const repositoryCommit = this.correlationDb.lookupRepositoryCommitByPackage(
            moniker.scheme,
            packageInformation.name,
            packageInformation.version!
        )

        if (!repositoryCommit) {
            return undefined
        }

        // TODO(efritz) - determine why npm monikers are mismatched
        const parts = moniker.identifier.split(':')
        parts[1] = '' // WTF
        moniker.identifier = parts.join(':')

        // TODO - need to cache db handles instead of backends
        // if we end up going with SQLite databases
        const subDb = new Database(this.correlationDb)
        subDb.load(makeFilename(repositoryCommit.repository, repositoryCommit.commit))

        try {
            const defsResult: DefsResult[] = subDb.findDefsStmt.all({
                scheme: moniker.scheme,
                identifier: moniker.identifier,
            })

            for (const defxResult of defsResult || []) {
                const subUri = subDb.uriTransformer.fromDatabase(defxResult.uri)

                return lsp.Location.create(
                    `git://${repositoryCommit.repository}?${repositoryCommit.commit}#${subUri}`,
                    lsp.Range.create(
                        defxResult.startLine,
                        defxResult.startCharacter,
                        defxResult.endLine,
                        defxResult.endCharacter
                    )
                )
            }
        } finally {
            subDb.close()
        }

        return undefined
    }

    //
    // References

    public references(uri: string, position: lsp.Position): lsp.Location[] | undefined {
        const { range, blob } = this.findRangeFromPosition(this.uriTransformer.toDatabase(uri), position)
        if (!range || !blob || !blob.referenceResults) {
            return undefined
        }

        const resultData = this.findResult(blob.resultSets, blob.referenceResults, range, 'referenceResult')
        const monikers = this.findMonikers(blob.resultSets, blob.monikers, range)
        const result = this.localReferences(uri, blob, resultData, monikers) || []

        for (const moniker of monikers) {
            if (moniker.kind === 'export') {
                const remoteReferences = this.remoteReferences(blob, moniker)
                if (remoteReferences !== undefined) {
                    result.push(...remoteReferences)
                }

                break
            }
        }

        return result
    }

    private localReferences(
        uri: string,
        blob: DocumentBlob,
        resultData: ReferenceResultData | undefined,
        monikers: MonikerData[]
    ): lsp.Location[] | undefined {
        if (resultData) {
            const result = []
            result.push(...asLocations(blob.ranges, uri, resultData.definitions || []))
            result.push(...asLocations(blob.ranges, uri, resultData.references || []))
            return result
        }

        for (const moniker of monikers) {
            const refsResult: RefsResult[] = this.findRefsStmt.all({
                scheme: moniker.scheme,
                identifier: moniker.identifier,
            })

            if (refsResult && refsResult.length > 0) {
                return refsResult.map(item =>
                    lsp.Location.create(
                        this.uriTransformer.fromDatabase(item.uri),
                        lsp.Range.create(item.startLine, item.startCharacter, item.endLine, item.endCharacter)
                    )
                )
            }
        }

        return undefined
    }

    private remoteReferences(blob: DocumentBlob, moniker: MonikerData): lsp.Location[] | undefined {
        if (!moniker.packageInformation) {
            return undefined
        }

        const packageInformation = blob.packageInformation![moniker.packageInformation]
        const repositoryCommits = this.correlationDb.getAllRepositoryCommitReferences(
            moniker.scheme,
            packageInformation.name!,
            packageInformation.version!,
            moniker.identifier
        )

        const allReferences = []
        for (const repositoryCommit of repositoryCommits) {
            // TODO(efritz) - determine why npm monikers are mismatched
            const parts = moniker.identifier.split(':')
            parts[1] = '' // WTF
            moniker.identifier = parts.join(':')

            // TODO - need to cache db handles instead of backends
            // if we end up going with SQLite databases
            const subDb = new Database(this.correlationDb)
            subDb.load(makeFilename(repositoryCommit.repository, repositoryCommit.commit))

            try {
                const refsResult: RefsResult[] = subDb.findRefsStmt.all({
                    scheme: moniker.scheme,
                    identifier: moniker.identifier,
                })

                if (refsResult && refsResult.length > 0) {
                    for (const refxResult of refsResult) {
                        const subUri = subDb.uriTransformer.fromDatabase(refxResult.uri)

                        allReferences.push(
                            lsp.Location.create(
                                `git://${repositoryCommit.repository}?${repositoryCommit.commit}#${subUri}`,
                                lsp.Range.create(
                                    refxResult.startLine,
                                    refxResult.startCharacter,
                                    refxResult.endLine,
                                    refxResult.endCharacter
                                )
                            )
                        )
                    }
                }
            } finally {
                subDb.close()
            }
        }

        return allReferences.length > 0 ? allReferences : undefined
    }

    //
    // Hover

    public hover(uri: string, position: lsp.Position): lsp.Hover | undefined {
        const { range, blob } = this.findRangeFromPosition(this.uriTransformer.toDatabase(uri), position)
        if (!range || !blob || !blob.hovers) {
            return undefined
        }

        const result = this.findResult(blob.resultSets, blob.hovers, range, 'hoverResult')
        if (result !== undefined) {
            return result
        }

        for (const moniker of this.findMonikers(blob.resultSets, blob.monikers, range)) {
            const hoversResult: BlobResult = this.findHoverStmt.get({
                scheme: moniker.scheme,
                identifier: moniker.identifier,
            })

            if (hoversResult) {
                // TODO(efritz) - convert to promises
                const result: lsp.Hover = JSON.parse(gunzipSync(hoversResult.content).toString())
                if (!result.range) {
                    result.range = lsp.Range.create(
                        range.start.line,
                        range.start.character,
                        range.end.line,
                        range.end.character
                    )
                }

                return result
            }
        }

        return undefined
    }

    //
    // TODO - categorize

    private findResult<T>(
        resultSets: LiteralMap<ResultSetData> | undefined,
        map: LiteralMap<T>,
        data: RangeData | ResultSetData,
        property: 'definitionResult' | 'referenceResult' | 'hoverResult'
    ): T | undefined {
        let current: RangeData | ResultSetData | undefined = data
        while (current) {
            const value = current[property]
            if (value) {
                return map[value]
            }

            if (!current.next || !resultSets) {
                break
            }

            current = resultSets[current.next]
        }

        return undefined
    }

    private findMonikers(
        resultSets: LiteralMap<ResultSetData> | undefined,
        monikers: LiteralMap<MonikerData> | undefined,
        data: RangeData | ResultSetData
    ): MonikerData[] {
        if (!monikers) {
            return []
        }

        const ids = []

        let current: RangeData | ResultSetData | undefined = data
        while (current) {
            if (current.monikers) {
                ids.push(...current.monikers)
            }

            if (!current.next || !resultSets) {
                break
            }

            current = resultSets[current.next]
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
        if (!documentId) {
            return { range: undefined, blob: undefined }
        }

        let candidate: RangeData | undefined
        const blob = this.getBlob(documentId)

        for (const key of Object.keys(blob.ranges)) {
            const range = blob.ranges[key]
            if (containsPosition(range, position)) {
                if (!candidate || containsRange(candidate, range)) {
                    candidate = range
                }
            }
        }

        return { range: candidate, blob }
    }

    private getBlob(documentId: Id): DocumentBlob {
        let result = this.blobs.get(documentId)
        if (result) {
            return result
        }

        // TODO(efritz) - convert to promises
        const blobResult: BlobResult = this.findBlobStmt.get(documentId)
        result = JSON.parse(gunzipSync(blobResult.content).toString()) as DocumentBlob
        this.blobs.set(documentId, result)
        return result
    }

    private findFile(uri: string): Id | undefined {
        let result: DocumentResult = this.findDocumentStmt.get({ uri: uri })
        return result && result.documentHash
    }
}

function asLocations(ranges: LiteralMap<RangeData>, uri: string, ids: Id[]): lsp.Location[] {
    return ids.map(id =>
        lsp.Location.create(
            uri,
            lsp.Range.create(
                ranges[id].start.line,
                ranges[id].start.character,
                ranges[id].end.line,
                ranges[id].end.character
            )
        )
    )
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
