import { URI } from 'vscode-uri'
import { Range, Id, MetaData, RangeBasedDocumentSymbol, Moniker, PackageInformation } from 'lsif-protocol'
import * as lsp from 'vscode-languageserver-protocol'
import { CorrelationDatabase } from './xrepo'
import { makeFilename } from './backend'
import { decodeJSON } from './encoding'
import * as entities from './entities'
import { Connection } from 'typeorm'
import { connectionCache, blobCache } from './cache'

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

export interface UriTransformer {
    toDatabase(uri: string): string
    fromDatabase(uri: string): string
}

export class Database {
    private uriTransformer!: UriTransformer
    private database!: string

    private async withConnection<T>(callback: (connection: Connection) => Promise<T>): Promise<T> {
        return await connectionCache.withConnection(
            this.database,
            [entities.Blob, entities.Def, entities.Document, entities.Hover, entities.Meta, entities.Ref],
            callback
        )
    }

    public constructor(private correlationDb: CorrelationDatabase) {}

    public async load(database: string): Promise<void> {
        this.database = database
        const root = await this.getProjectRoot()

        // TODO - flatten this, no need for an interface
        this.uriTransformer = {
            toDatabase: path => `${root}/${path}`,
            fromDatabase: uri => (uri.startsWith(root) ? uri.slice(`${root}/`.length) : uri),
        }
    }

    private async getProjectRoot(): Promise<string> {
        const metaData: MetaData = await this.withConnection(async connection =>
            JSON.parse((await connection.getRepository(entities.Meta).findOneOrFail()).value)
        )

        if (metaData.projectRoot) {
            return URI.parse(metaData.projectRoot).toString(true)
        }

        throw new Error('Failed to get project root from meta.')
    }

    //
    // Exists

    public exists(uri: string): boolean {
        return Boolean(this.findFile(this.uriTransformer.toDatabase(uri)))
    }

    //
    // Definitions

    public async definitions(uri: string, position: lsp.Position): Promise<lsp.Location | lsp.Location[] | undefined> {
        const { range, blob } = await this.findRangeFromPosition(this.uriTransformer.toDatabase(uri), position)
        if (!range || !blob || !blob.definitionResults) {
            return undefined
        }

        const resultData = this.findResult(blob.resultSets, blob.definitionResults, range, 'definitionResult')
        if (resultData) {
            return asLocations(blob.ranges, uri, resultData.values)
        }

        for (const moniker of this.findMonikers(blob.resultSets, blob.monikers, range)) {
            if (moniker.kind === 'import') {
                return await this.getRemoteDefinition(blob, moniker)
            }

            // TODO(efritz) - prove that this returns something useful
            // in some circumstances. I'm not sure if this was meant to
            // do what we're trying to do now...

            const defsResult = await this.withConnection(connection =>
                connection.getRepository(entities.Def).find({
                    where: {
                        scheme: moniker.scheme,
                        identifier: moniker.identifier,
                    },
                    relations: ['document'],
                })
            )

            if (defsResult && defsResult.length > 0) {
                return defsResult.map(item =>
                    lsp.Location.create(
                        this.uriTransformer.fromDatabase(item.document.uri),
                        lsp.Range.create(item.startLine, item.startCharacter, item.endLine, item.endCharacter)
                    )
                )
            }
        }

        return undefined
    }

    private async getRemoteDefinition(
        blob: DocumentBlob,
        moniker: MonikerData
    ): Promise<lsp.Location | lsp.Location[] | undefined> {
        if (!moniker.packageInformation) {
            return undefined
        }

        const packageInformation = blob.packageInformation![moniker.packageInformation]
        const packageEntity = await this.correlationDb.getPackage(
            moniker.scheme,
            packageInformation.name,
            packageInformation.version!
        )

        if (!packageEntity) {
            return undefined
        }

        // TODO(efritz) - determine why npm monikers are mismatched
        const parts = moniker.identifier.split(':')
        parts[1] = '' // WTF
        moniker.identifier = parts.join(':')

        // TODO - need to cache db handles instead of backends
        // if we end up going with SQLite databases
        const subDb = new Database(this.correlationDb)
        await subDb.load(makeFilename(packageEntity.repository, packageEntity.commit))

        const defsResult = await subDb.withConnection(connection =>
            connection.getRepository(entities.Def).find({
                where: {
                    scheme: moniker.scheme,
                    identifier: moniker.identifier,
                },
                relations: ['document'],
            })
        )

        for (const defxResult of defsResult) {
            const subUri = subDb.uriTransformer.fromDatabase(defxResult.document.uri)

            return lsp.Location.create(
                `git://${packageEntity.repository}?${packageEntity.commit}#${subUri}`,
                lsp.Range.create(
                    defxResult.startLine,
                    defxResult.startCharacter,
                    defxResult.endLine,
                    defxResult.endCharacter
                )
            )
        }

        return undefined
    }

    //
    // References

    public async references(uri: string, position: lsp.Position): Promise<lsp.Location[] | undefined> {
        const { range, blob } = await this.findRangeFromPosition(this.uriTransformer.toDatabase(uri), position)
        if (!range || !blob || !blob.referenceResults) {
            return undefined
        }

        const resultData = this.findResult(blob.resultSets, blob.referenceResults, range, 'referenceResult')
        const monikers = this.findMonikers(blob.resultSets, blob.monikers, range)
        const result = (await this.localReferences(uri, blob, resultData, monikers)) || []

        for (const moniker of monikers) {
            if (moniker.kind === 'export') {
                const remoteReferences = await this.remoteReferences(blob, moniker)
                if (remoteReferences !== undefined) {
                    result.push(...remoteReferences)
                }

                break
            }
        }

        return result
    }

    private async localReferences(
        uri: string,
        blob: DocumentBlob,
        resultData: ReferenceResultData | undefined,
        monikers: MonikerData[]
    ): Promise<lsp.Location[] | undefined> {
        if (resultData) {
            const result = []
            result.push(...asLocations(blob.ranges, uri, resultData.definitions || []))
            result.push(...asLocations(blob.ranges, uri, resultData.references || []))
            return result
        }

        for (const moniker of monikers) {
            const refsResult = await this.withConnection(connection =>
                connection.getRepository(entities.Ref).find({
                    where: {
                        scheme: moniker.scheme,
                        identifier: moniker.identifier,
                    },
                    relations: ['document'],
                })
            )

            if (refsResult && refsResult.length > 0) {
                return refsResult.map(item =>
                    lsp.Location.create(
                        this.uriTransformer.fromDatabase(item.document.uri),
                        lsp.Range.create(item.startLine, item.startCharacter, item.endLine, item.endCharacter)
                    )
                )
            }
        }

        return undefined
    }

    private async remoteReferences(blob: DocumentBlob, moniker: MonikerData): Promise<lsp.Location[] | undefined> {
        if (!moniker.packageInformation) {
            return undefined
        }

        const packageInformation = blob.packageInformation![moniker.packageInformation]
        const references = await this.correlationDb.getReferences(
            moniker.scheme,
            packageInformation.name!,
            packageInformation.version!,
            moniker.identifier
        )

        const allReferences = []
        for (const reference of references) {
            // TODO(efritz) - determine why npm monikers are mismatched
            const parts = moniker.identifier.split(':')
            parts[1] = '' // WTF
            moniker.identifier = parts.join(':')

            // TODO - need to cache db handles instead of backends
            // if we end up going with SQLite databases
            const subDb = new Database(this.correlationDb)
            await subDb.load(makeFilename(reference.repository, reference.commit))

            const refsResult = await subDb.withConnection(connection =>
                connection.getRepository(entities.Ref).find({
                    where: {
                        scheme: moniker.scheme,
                        identifier: moniker.identifier,
                    },
                    relations: ['document'],
                })
            )

            if (refsResult && refsResult.length > 0) {
                for (const refxResult of refsResult) {
                    const subUri = subDb.uriTransformer.fromDatabase(refxResult.document.uri)

                    allReferences.push(
                        lsp.Location.create(
                            `git://${reference.repository}?${reference.commit}#${subUri}`,
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
        }

        return allReferences.length > 0 ? allReferences : undefined
    }

    //
    // Hover

    public async hover(uri: string, position: lsp.Position): Promise<lsp.Hover | undefined> {
        const { range, blob } = await this.findRangeFromPosition(this.uriTransformer.toDatabase(uri), position)
        if (!range || !blob || !blob.hovers) {
            return undefined
        }

        const result = this.findResult(blob.resultSets, blob.hovers, range, 'hoverResult')
        if (result !== undefined) {
            return result
        }

        for (const moniker of this.findMonikers(blob.resultSets, blob.monikers, range)) {
            const hoverResults = await this.withConnection(connection =>
                connection.getRepository(entities.Hover).find({
                    where: {
                        scheme: moniker.scheme,
                        identifier: moniker.identifier,
                    },
                    relations: ['blob'],
                })
            )

            for (const hoverResult of hoverResults) {
                const result: lsp.Hover = JSON.parse(await decodeJSON(hoverResult.blob.value))
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

    private async findRangeFromPosition(
        uri: string,
        position: lsp.Position
    ): Promise<{ range: RangeData | undefined; blob: DocumentBlob | undefined }> {
        const documentId = await this.findFile(uri)
        if (!documentId) {
            return { range: undefined, blob: undefined }
        }

        const blob = await blobCache.withBlob(
            documentId,
            async () =>
                JSON.parse(
                    await decodeJSON(
                        (await this.withConnection(connection =>
                            connection.getRepository(entities.Blob).findOneOrFail(documentId)
                        )).value
                    )
                ) as DocumentBlob,
            async blob => blob
        )

        let candidate: RangeData | undefined
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

    private async findFile(uri: string): Promise<Id | undefined> {
        const result = await this.withConnection(connection =>
            connection.getRepository(entities.Document).findOne({ uri: uri })
        )
        return result && result.hash
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
