import * as lsp from 'vscode-languageserver-protocol'
import { blobCache, connectionCache } from './cache'
import { BlobModel, DefModel, DocumentModel, MetaModel, RefModel } from './models'
import { Connection } from 'typeorm'
import { decodeJSON } from './encoding'
import { DocumentBlob, MonikerData, RangeData, ReferenceResultData, ResultSetData } from './entities'
import { Id } from 'lsif-protocol'
import { makeFilename } from './backend'
import { XrepoDatabase } from './xrepo'

const MONIKER_KIND_PREFERENCES = ['import', 'local', 'export']
const MONIKER_SCHEME_PREFERENCES = ['npm', 'tsc']

export class Database {
    /**
     * Create a new `Database` with the given cross-repo database instance and the
     * filename of the database that contains data for a particular repository/commit.
     *
     * @param xrepoDatabase The cross-repo databse.
     * @param database The filename of the database.
     */
    constructor(private xrepoDatabase: XrepoDatabase, private database: string) {}

    /**
     * Determine if data exists for a particular document in this database.
     *
     * @param uri The URI of the document.
     */
    public async exists(uri: string): Promise<boolean> {
        return (await this.findFile(uri)) !== undefined
    }

    //
    // Definitions

    public async definitions(uri: string, position: lsp.Position): Promise<lsp.Location | lsp.Location[] | undefined> {
        const blob = await this.findBlob(uri)
        if (!blob) {
            return undefined
        }

        const range = findRange(blob, position)
        if (!range) {
            return undefined
        }

        const resultData = findResult(blob.resultSets, blob.definitionResults, range, 'definitionResult')
        if (resultData) {
            return asLocations(blob.ranges, uri, resultData.values)
        }

        for (const moniker of findMonikers(blob.resultSets, blob.monikers, range)) {
            if (moniker.kind === 'import') {
                return await this.getRemoteDefinition(blob, moniker)
            }

            const defsResult = await this.withConnection(connection =>
                connection.getRepository(DefModel).find({
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
                        item.document.uri,
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

        const packageInformation = blob.packageInformation.get(moniker.packageInformation)
        if (!packageInformation) {
            return undefined
        }

        const packageEntity = await this.xrepoDatabase.getPackage(
            moniker.scheme,
            packageInformation.name,
            packageInformation.version
        )

        if (!packageEntity) {
            return undefined
        }

        // TODO(efritz) - determine why npm monikers are mismatched
        const parts = moniker.identifier.split(':')
        parts[1] = '' // WTF
        moniker.identifier = parts.join(':')

        const subDb = new Database(this.xrepoDatabase, makeFilename(packageEntity.repository, packageEntity.commit))

        const defsResult = await subDb.withConnection(connection =>
            connection.getRepository(DefModel).find({
                where: {
                    scheme: moniker.scheme,
                    identifier: moniker.identifier,
                },
                relations: ['document'],
            })
        )

        for (const defxResult of defsResult) {
            return lsp.Location.create(
                `git://${packageEntity.repository}?${packageEntity.commit}#${defxResult.document.uri}`,
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
        const blob = await this.findBlob(uri)
        if (!blob) {
            return undefined
        }

        const range = findRange(blob, position)
        if (!range) {
            return undefined
        }

        const resultData = findResult(blob.resultSets, blob.referenceResults, range, 'referenceResult')
        const monikers = findMonikers(blob.resultSets, blob.monikers, range)
        const result = await this.localReferences(uri, blob, resultData, monikers)

        for (const moniker of monikers) {
            if (moniker.kind === 'export') {
                const moreResult = await this.remoteReferences(blob, moniker)
                result.push(...moreResult)
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
    ): Promise<lsp.Location[]> {
        if (resultData) {
            const result = []
            result.push(...asLocations(blob.ranges, uri, resultData.definitions))
            result.push(...asLocations(blob.ranges, uri, resultData.references))
            return result
        }

        for (const moniker of monikers) {
            const refsResult = await this.withConnection(connection =>
                connection.getRepository(RefModel).find({
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
                        item.document.uri,
                        lsp.Range.create(item.startLine, item.startCharacter, item.endLine, item.endCharacter)
                    )
                )
            }
        }

        return []
    }

    private async remoteReferences(blob: DocumentBlob, moniker: MonikerData): Promise<lsp.Location[]> {
        if (!moniker.packageInformation) {
            return []
        }

        const packageInformation = blob.packageInformation.get(moniker.packageInformation)
        if (!packageInformation) {
            return []
        }

        const references = await this.xrepoDatabase.getReferences(
            moniker.scheme,
            packageInformation.name,
            packageInformation.version,
            moniker.identifier
        )

        const allReferences = []
        for (const reference of references) {
            // TODO(efritz) - determine why npm monikers are mismatched
            const parts = moniker.identifier.split(':')
            parts[1] = '' // WTF
            moniker.identifier = parts.join(':')

            const subDb = new Database(this.xrepoDatabase, makeFilename(reference.repository, reference.commit))

            const refsResult = await subDb.withConnection(connection =>
                connection.getRepository(RefModel).find({
                    where: {
                        scheme: moniker.scheme,
                        identifier: moniker.identifier,
                    },
                    relations: ['document'],
                })
            )

            if (refsResult && refsResult.length > 0) {
                for (const refxResult of refsResult) {
                    allReferences.push(
                        lsp.Location.create(
                            `git://${reference.repository}?${reference.commit}#${refxResult.document.uri}`,
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

        return allReferences
    }

    //
    // Hover

    public async hover(uri: string, position: lsp.Position): Promise<lsp.Hover | undefined> {
        const blob = await this.findBlob(uri)
        if (!blob) {
            return undefined
        }

        const range = findRange(blob, position)
        if (!range) {
            return undefined
        }

        return findResult(blob.resultSets, blob.hovers, range, 'hoverResult')
    }

    //
    // TODO - categorize

    private async findFile(uri: string): Promise<Id | undefined> {
        // TODO - why not join?
        const result = await this.withConnection(connection => connection.getRepository(DocumentModel).findOne({ uri }))
        return result && result.hash
    }

    private async findBlob(uri: string): Promise<DocumentBlob | undefined> {
        const documentId = await this.findFile(uri)
        if (!documentId) {
            return undefined
        }

        const blobFactory = async (): Promise<DocumentBlob> =>
            await decodeJSON<DocumentBlob>(
                (await this.withConnection(connection => connection.getRepository(BlobModel).findOneOrFail(documentId)))
                    .value
            )

        return await blobCache.withBlob(documentId, blobFactory, blob => Promise.resolve(blob))
    }

    private async withConnection<T>(callback: (connection: Connection) => Promise<T>): Promise<T> {
        return await connectionCache.withConnection(
            this.database,
            [BlobModel, DefModel, DocumentModel, MetaModel, RefModel],
            callback
        )
    }
}

// TODO - order ranges so we can search this efficiently
function findRange(blob: DocumentBlob, position: lsp.Position): RangeData | undefined {
    for (const range of blob.ranges.values()) {
        if (containsPosition(range, position)) {
            return range
        }
    }

    return undefined
}

function findResult<T>(
    resultSets: Map<Id, ResultSetData> | undefined,
    map: Map<Id, T>,
    data: RangeData | ResultSetData,
    property: 'definitionResult' | 'referenceResult' | 'hoverResult'
): T | undefined {
    let current: RangeData | ResultSetData | undefined = data
    while (current) {
        const value = current[property]
        if (value) {
            return map.get(value)
        }

        if (!current.next || !resultSets) {
            break
        }

        current = resultSets.get(current.next)
    }

    return undefined
}

function findMonikers(
    resultSets: Map<Id, ResultSetData>,
    monikers: Map<Id, MonikerData>,
    data: RangeData | ResultSetData
): MonikerData[] {
    const monikerSet = []

    let current: RangeData | ResultSetData | undefined = data
    while (current) {
        for (const id of current.monikers) {
            const moniker = monikers.get(id)
            if (moniker) {
                monikerSet.push(moniker)
            }
        }

        if (!current.next || !resultSets) {
            break
        }

        current = resultSets.get(current.next)
    }

    return sortMonikers(monikerSet)
}

function sortMonikers(monikers: MonikerData[]): MonikerData[] {
    monikers.sort((a, b) => {
        const ord = MONIKER_KIND_PREFERENCES.indexOf(a.kind) - MONIKER_KIND_PREFERENCES.indexOf(b.kind)
        if (ord !== 0) {
            return ord
        }

        return MONIKER_SCHEME_PREFERENCES.indexOf(a.scheme) - MONIKER_SCHEME_PREFERENCES.indexOf(b.scheme)
    })

    return monikers
}

function asLocations(ranges: Map<Id, RangeData>, uri: string, ids: Id[]): lsp.Location[] {
    const locations = []
    for (const id of ids) {
        const range = ranges.get(id)
        if (range) {
            locations.push(
                lsp.Location.create(
                    uri,
                    lsp.Range.create(range.start.line, range.start.character, range.end.line, range.end.character)
                )
            )
        }
    }

    return locations
}

function containsPosition(range: lsp.Range, position: lsp.Position): boolean {
    if (position.line < range.start.line || range.end.line < position.line) {
        return false
    }

    if (position.line === range.start.line && position.character < range.start.character) {
        return false
    }

    if (position.line === range.end.line && range.end.character < position.character) {
        return false
    }

    return true
}
