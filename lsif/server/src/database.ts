import * as lsp from 'vscode-languageserver-protocol'
import { connectionCache, blobCache } from './cache'
import { DocumentModel, DefModel, MetaModel, RefModel, PackageModel } from './models'
import { Connection } from 'typeorm'
import { decodeJSON } from './encoding'
import { DocumentBlob, MonikerData, RangeData, ReferenceResultData, ResultSetData } from './entities'
import { Id } from 'lsif-protocol'
import { makeFilename } from './backend'
import { XrepoDatabase } from './xrepo'

const MONIKER_KIND_PREFERENCES = ['import', 'local', 'export']
const MONIKER_SCHEME_PREFERENCES = ['npm', 'tsc']

/**
 * `Database` wraps operations around a single repository/commit pair.
 */
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
        return (await this.findBlob(uri)) !== undefined
    }

    /**
     * Return the location for the definition of the reference at the given position.
     *
     * @param uri The document to which the position belongs.
     * @param position The current hover position.
     */
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
                return await this.remoteDefinition(blob, moniker)
            }

            const defs = await this.localDefinition(moniker)
            if (defs) {
                return defs
            }
        }

        return undefined
    }

    /**
     * Return a list of locations which reference the definition at the given position.
     *
     * @param uri The document to which the position belongs.
     * @param position The current hover position.
     */
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

    /**
     * Return the hover content for the definition or reference at the given position.
     *
     * @param uri The document to which the position belongs.
     * @param position The current hover position.
     */
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

    // TODO - document
    private async localDefinition(moniker: MonikerData): Promise<lsp.Location | lsp.Location[] | undefined> {
        const defsResult = await this.withConnection(connection =>
            connection.getRepository(DefModel).find({
                where: {
                    scheme: moniker.scheme,
                    identifier: moniker.identifier,
                },
            })
        )

        if (!defsResult || defsResult.length === 0) {
            return undefined
        }

        return defsResult.map(item => lsp.Location.create(item.documentUri, makeRange(item)))
    }

    // TODO - document
    private async remoteDefinition(
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
            })
        )

        for (const result of defsResult) {
            return lsp.Location.create(makeRemoteUri(packageEntity, result), makeRange(result))
        }

        return undefined
    }

    // TODO - document
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
                })
            )

            if (!refsResult || refsResult.length === 0) {
                continue
            }

            return refsResult.map(item => lsp.Location.create(item.documentUri, makeRange(item)))
        }

        return []
    }

    // TODO - document
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
                })
            )

            if (refsResult && refsResult.length > 0) {
                for (const result of refsResult) {
                    allReferences.push(lsp.Location.create(makeRemoteUri(reference, result), makeRange(result)))
                }
            }
        }

        return allReferences
    }

    /**
     * Return a parsed document blob with the given URI. The result of this
     * method is cached across all database instances.
     *
     * @param uri The uri of the document.
     */
    private async findBlob(uri: string): Promise<DocumentBlob | undefined> {
        const blobFactory = async (): Promise<DocumentBlob> => {
            const document = await this.withConnection(connection =>
                connection.getRepository(DocumentModel).findOneOrFail(uri)
            )

            return await decodeJSON<DocumentBlob>(document.value)
        }

        return await blobCache.withBlob(`${this.database}::${uri}`, blobFactory, blob => Promise.resolve(blob))
    }

    /**
     * Invoke `callback` with a SQLite connection object obtained from the
     * cache or created on cache miss.
     *
     * @param callback The function invoke with the SQLite connection.
     */
    private async withConnection<T>(callback: (connection: Connection) => Promise<T>): Promise<T> {
        return await connectionCache.withConnection(
            this.database,
            [DefModel, DocumentModel, MetaModel, RefModel],
            callback
        )
    }
}

// TODO - document
// TODO - order ranges so we can search this efficiently
function findRange(blob: DocumentBlob, position: lsp.Position): RangeData | undefined {
    for (const range of blob.ranges.values()) {
        if (containsPosition(range, position)) {
            return range
        }
    }

    return undefined
}

// TODO - document
function findResult<T>(
    resultSets: Map<Id, ResultSetData> | undefined,
    map: Map<Id, T>,
    data: RangeData | ResultSetData,
    property: 'definitionResult' | 'referenceResult' | 'hoverResult'
): T | undefined {
    return withChain(resultSets, data, current => {
        const value = current[property]
        if (value) {
            return map.get(value)
        }

        return undefined
    })
}

// TODO - document
function findMonikers(
    resultSets: Map<Id, ResultSetData>,
    monikers: Map<Id, MonikerData>,
    data: RangeData | ResultSetData
): MonikerData[] {
    const monikerSet: MonikerData[] = []

    withChain(resultSets, data, current => {
        for (const id of current.monikers) {
            const moniker = monikers.get(id)
            if (moniker) {
                monikerSet.push(moniker)
            }
        }

        return undefined
    })

    return sortMonikers(monikerSet)
}

// TODO - document
function withChain<T>(
    resultSets: Map<Id, ResultSetData> | undefined,
    data: RangeData | ResultSetData,
    visitor: (current: RangeData | ResultSetData) => T | undefined
): T | undefined {
    let current: RangeData | ResultSetData | undefined = data
    while (current) {
        const value = visitor(current)
        if (value) {
            return value
        }

        if (!current.next || !resultSets) {
            break
        }

        current = resultSets.get(current.next)
    }

    return undefined
}

// TODO - document
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

// TODO - document
function asLocations(ranges: Map<Id, RangeData>, uri: string, ids: Id[]): lsp.Location[] {
    const locations = []
    for (const id of ids) {
        const range = ranges.get(id)
        if (range) {
            locations.push(
                lsp.Location.create(uri, {
                    start: range.start,
                    end: range.end,
                })
            )
        }
    }

    return locations
}

// TODO - document
function makeRemoteUri(pkg: PackageModel, result: DefModel | RefModel): string {
    return `git://${pkg.repository}?${pkg.commit}#${result.documentUri}`
}

// TODO - document
function makeRange(result: {
    startLine: number
    startCharacter: number
    endLine: number
    endCharacter: number
}): lsp.Range {
    return lsp.Range.create(result.startLine, result.startCharacter, result.endLine, result.endCharacter)
}

// TODO - document
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
