import * as lsp from 'vscode-languageserver-protocol'
import { DocumentModel, DefModel, MetaModel, RefModel, PackageModel } from './models'
import { Connection } from 'typeorm'
import { decodeJSON } from './encoding'
import { MonikerData, RangeData, ResultSetData, DocumentData, FlattenedRange } from './entities'
import { Id } from 'lsif-protocol'
import { makeFilename } from './backend'
import { XrepoDatabase } from './xrepo'
import { ConnectionCache, DocumentCache } from './cache'

/**
 * A wrapper around operations for single repository/commit pair.
 */
export class Database {
    /**
     * Create a new `Database` with the given cross-repo database instance and the
     * filename of the database that contains data for a particular repository/commit.
     *
     * @param xrepoDatabase The cross-repo databse.
     * @param connectionCache The cache of SQLite connections.
     * @param documentCache The cache of loaded document.
     * @param databasePath The path to the database file.
     */
    constructor(
        private xrepoDatabase: XrepoDatabase,
        private connectionCache: ConnectionCache,
        private documentCache: DocumentCache,
        private databasePath: string
    ) {}

    /**
     * Determine if data exists for a particular document in this database.
     *
     * @param path The path of the document.
     */
    public async exists(path: string): Promise<boolean> {
        return (await this.findDocument(path)) !== undefined
    }

    /**
     * Return the location for the definition of the reference at the given position.
     *
     * @param path The path fo the document to which the position belongs.
     * @param position The current hover position.
     */
    public async definitions(path: string, position: lsp.Position): Promise<lsp.Location[] | null> {
        const { document, range } = await this.findRange(path, position)
        if (!document || !range) {
            return null
        }

        const resultData = findResult(document.resultSets, document.definitionResults, range, 'definitionResult')
        if (resultData) {
            return asLocations(document.ranges, document.orderedRanges, path, resultData)
        }

        // TODO - vet this logic of finding the first result
        for (const moniker of findMonikers(document.resultSets, document.monikers, range)) {
            if (moniker.kind === 'import') {
                return await this.remoteDefinitions(document, moniker)
            }

            const defs = await Database.monikerResults(this, DefModel, moniker, path => path)
            if (defs) {
                return defs
            }
        }

        return null
    }

    /**
     * Return a list of locations which reference the definition at the given position.
     *
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     */
    public async references(path: string, position: lsp.Position): Promise<lsp.Location[] | undefined> {
        const { document, range } = await this.findRange(path, position)
        if (!document || !range) {
            return undefined
        }

        // TODO - vet logic - should xrepo search be explicitly enabled via query flag?
        const resultData = findResult(document.resultSets, document.referenceResults, range, 'referenceResult')
        const monikers = findMonikers(document.resultSets, document.monikers, range)

        let result: lsp.Location[] = []
        if (resultData) {
            result = result.concat(asLocations(document.ranges, document.orderedRanges, path, resultData))
        } else {
            for (const moniker of monikers) {
                result = result.concat(await Database.monikerResults(this, RefModel, moniker, path => path))
            }
        }

        for (const moniker of monikers) {
            if (moniker.kind === 'import' || moniker.kind === 'export') {
                const remoteResults = await this.remoteReferences(document, moniker)
                result = result.concat(remoteResults)
                break
            }
        }

        return result
    }

    /**
     * Return the hover content for the definition or reference at the given position.
     *
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     */
    public async hover(path: string, position: lsp.Position): Promise<lsp.Hover | null> {
        const { document, range } = await this.findRange(path, position)
        if (!document || !range) {
            return null
        }

        // All hover contents should be contained in the document.
        const contents = findResult(document.resultSets, document.hovers, range, 'hoverResult')
        if (!contents) {
            return null
        }

        return { contents }
    }

    //
    // Helper Functions

    /**
     * Query the defs or refs table of `db` for items that match the given moniker. Convert
     * each result into an LSP location. The `pathTransformer` function is invoked on each
     * result item to modify the resulting locations.
     *
     * @param db The target database.
     * @param model The constructor for the model type.
     * @param moniker The target moniker.
     * @param pathTransformer The function used to alter location paths.
     */
    private static async monikerResults(
        db: Database,
        model: typeof DefModel | typeof RefModel,
        moniker: MonikerData,
        pathTransformer: (path: string) => string
    ): Promise<lsp.Location[]> {
        const results = await db.withConnection(connection =>
            connection.getRepository<DefModel | RefModel>(model).find({
                where: {
                    scheme: moniker.scheme,
                    identifier: moniker.identifier,
                },
            })
        )

        return results.map(result => lsp.Location.create(pathTransformer(result.documentPath), makeRange(result)))
    }

    /**
     * Find the definition of the target moniker outside of the current database. If the
     * moniker has attached package information, then the xrepo database is queried for
     * the target package. That database is opened, and its def table is queried for the
     * target moniker.
     *
     * @param document The document containing the reference.
     * @param moniker The target moniker.
     */
    private async remoteDefinitions(document: DocumentData, moniker: MonikerData): Promise<lsp.Location[] | null> {
        if (!moniker.packageInformation) {
            return null
        }

        const packageInformation = document.packageInformation.get(moniker.packageInformation)
        if (!packageInformation) {
            return null
        }

        const packageEntity = await this.xrepoDatabase.getPackage(
            moniker.scheme,
            packageInformation.name,
            packageInformation.version
        )

        if (!packageEntity) {
            return null
        }

        const db = new Database(
            this.xrepoDatabase,
            this.connectionCache,
            this.documentCache,
            makeFilename(packageEntity.repository, packageEntity.commit)
        )

        const pathTransformer = (path: string): string => makeRemoteUri(packageEntity, path)
        return await Database.monikerResults(db, DefModel, moniker, pathTransformer)
    }

    /**
     * Find the references of the target moniker outside of the current database. If the moniker
     * has attached package information, then the xrepo database is queried for the packages that
     * require this particular moniker identifier. These databases are opened, and their ref tables
     * are queried for the target moniker.
     *
     * @param document The document containing the definition.
     * @param moniker THe target moniker.
     */
    private async remoteReferences(document: DocumentData, moniker: MonikerData): Promise<lsp.Location[]> {
        if (!moniker.packageInformation) {
            return []
        }

        const packageInformation = document.packageInformation.get(moniker.packageInformation)
        if (!packageInformation) {
            return []
        }

        const references = await this.xrepoDatabase.getReferences(
            moniker.scheme,
            packageInformation.name,
            packageInformation.version,
            moniker.identifier
        )

        let allReferences: lsp.Location[] = []
        for (const reference of references) {
            const db = new Database(
                this.xrepoDatabase,
                this.connectionCache,
                this.documentCache,
                makeFilename(reference.repository, reference.commit)
            )

            const pathTransformer = (path: string): string => makeRemoteUri(reference, path)
            const references = await Database.monikerResults(db, RefModel, moniker, pathTransformer)
            allReferences = allReferences.concat(references)
        }

        return allReferences
    }

    /**
     * Return a parsed document that describes the given path. The result of this
     * method is cached across all database instances.
     *
     * @param path The path of the document.
     */
    private async findDocument(path: string): Promise<DocumentData | undefined> {
        const factory = async (): Promise<DocumentData> => {
            const document = await this.withConnection(connection =>
                connection.getRepository(DocumentModel).findOneOrFail(path)
            )

            return await decodeJSON<DocumentData>(document.value)
        }

        return await this.documentCache.withDocument(`${this.databasePath}::${path}`, factory, document =>
            Promise.resolve(document)
        )
    }

    /**
     * Return a parsed document that describes the given path as well as the range
     * from that document that contains the given position. Returns undefined for
     * both values if one cannot be loaded.
     *
     * @param path The path of the document.
     * @param position The user's hover position.
     */
    private async findRange(
        path: string,
        position: lsp.Position
    ): Promise<{ document: DocumentData | undefined; range: RangeData | undefined }> {
        const document = await this.findDocument(path)
        if (!document) {
            return { document: undefined, range: undefined }
        }

        const range = findRange(document.orderedRanges, position)
        if (!range) {
            return { document: undefined, range: undefined }
        }

        return { document, range }
    }

    /**
     * Invoke `callback` with a SQLite connection object obtained from the
     * cache or created on cache miss.
     *
     * @param callback The function invoke with the SQLite connection.
     */
    private async withConnection<T>(callback: (connection: Connection) => Promise<T>): Promise<T> {
        return await this.connectionCache.withConnection(
            this.databasePath,
            [DefModel, DocumentModel, MetaModel, RefModel],
            callback
        )
    }
}

/**
 * Perform binary search over the ordered ranges of a document, returning
 * the range that includes it (if it exists). LSIF requires that no ranges
 * overlap in a single document. Then, we can compare a position against a
 * range by saying that it's contained within it (what we want), occurs
 * before it, or occurs after it. These later two results let us cut our
 * search space by half each time.
 *
 * @param orderedRanges The ranges of the document, ordered by startLine/startCharacter.
 * @param position The user's hover position.
 */
function findRange(orderedRanges: RangeData[], position: lsp.Position): RangeData | undefined {
    let lo = 0
    let hi = orderedRanges.length - 1

    while (lo <= hi) {
        const mid = Math.floor((lo + hi) / 2)
        const range = orderedRanges[mid]

        const cmp = comparePosition(range, position)
        if (cmp === 0) {
            return range
        }

        if (cmp < 0) {
            lo = mid + 1
        } else {
            hi = mid - 1
        }
    }

    return undefined
}

/**
 * Return the closest defined `property` related to the given range
 * or result set. This method will walk the `next` chains of the item
 * to find the property on an attached result set if it's not set
 * on the range itself. Note that the `property` on the range and
 * result set objects are simply identifiers, so the real value must
 * be looked up in a secondary data structure `map`.
 *
 * @param resultSets The map of results sets of the document.
 * @param map The map from which to return the property value.
 * @param data The range or result set object.
 * @param property The target property.
 */
function findResult<T>(
    resultSets: Map<Id, ResultSetData>,
    map: Map<Id, T>,
    data: RangeData | ResultSetData,
    property: 'definitionResult' | 'referenceResult' | 'hoverResult'
): T | undefined {
    for (const current of walkChain(resultSets, data)) {
        const value = current[property]
        if (value) {
            return map.get(value)
        }
    }

    return undefined
}

/**
 * Retrieve all monikers attached to a range or result set.
 *
 * @param resultSets The map of results sets of the document.
 * @param monikers The map of monikers of the document.
 * @param data The range or restult set object.
 */
function findMonikers(
    resultSets: Map<Id, ResultSetData>,
    monikers: Map<Id, MonikerData>,
    data: RangeData | ResultSetData
): MonikerData[] {
    const monikerSet: MonikerData[] = []
    for (const current of walkChain(resultSets, data)) {
        for (const id of current.monikers) {
            const moniker = monikers.get(id)
            if (moniker) {
                monikerSet.push(moniker)
            }
        }
    }

    return sortMonikers(monikerSet)
}

/**
 * Return an iterabel of the range and result set items that are attached
 * to the given initial data. The initial data is yielded immediately.
 *
 * @param resultSets The map of results sets of the document.
 * @param data The range or result set object.
 */
function* walkChain<T>(
    resultSets: Map<Id, ResultSetData>,
    data: RangeData | ResultSetData
): Iterable<RangeData | ResultSetData> {
    let current: RangeData | ResultSetData | undefined = data

    while (current) {
        yield current
        if (!current.next) {
            return
        }

        current = resultSets.get(current.next)
    }
}

/**
 * Sort the monikers by kind, then scheme in order of the following
 * preferences.
 *
 *   - kind: import, local, export
 *   - scheme: npm, tsc
 *
 * @param monikers The list of monikers.
 */
function sortMonikers(monikers: MonikerData[]): MonikerData[] {
    const monikerKindPreferences = ['import', 'local', 'export']
    const monikerSchemePreferences = ['npm', 'tsc']

    monikers.sort((a, b) => {
        const ord = monikerKindPreferences.indexOf(a.kind) - monikerKindPreferences.indexOf(b.kind)
        if (ord !== 0) {
            return ord
        }

        return monikerSchemePreferences.indexOf(a.scheme) - monikerSchemePreferences.indexOf(b.scheme)
    })

    return monikers
}

/**
 * Convert the given range identifiers into LSP location objects.
 *
 * @param ranges The map of ranges of the document (from identifier to the range's index inorderedRanges).
 * @param orderedRanges The ordered ranges of the document.
 * @param uri The location URI.
 * @param ids The set of range identifiers for each resulting location.
 */
function asLocations(ranges: Map<Id, number>, orderedRanges: RangeData[], uri: string, ids: Id[]): lsp.Location[] {
    const locations = []
    for (const id of ids) {
        const rangeIndex = ranges.get(id)
        if (!rangeIndex) {
            continue
        }

        const range = orderedRanges[rangeIndex]
        locations.push(
            lsp.Location.create(uri, {
                start: { line: range.startLine, character: range.startCharacter },
                end: { line: range.endLine, character: range.endCharacter },
            })
        )
    }

    return locations
}

/**
 * Construct a URI that can be used by the frontend to switch to another
 * directory.
 *
 * @param pkg The target package.
 * @param path The path relative to the project root.
 */
function makeRemoteUri(pkg: PackageModel, path: string): string {
    const url = new URL(`git://${pkg.repository}`)
    url.search = pkg.commit
    url.hash = path
    return url.href
}

/**
 * Construct an LSP range from a flat range.
 *
 * @param result The start/end line/character of the range.
 */
function makeRange(result: {
    startLine: number
    startCharacter: number
    endLine: number
    endCharacter: number
}): lsp.Range {
    return lsp.Range.create(result.startLine, result.startCharacter, result.endLine, result.endCharacter)
}

/**
 * Compare a position against a range. Returns 0 if the position occurs
 * within the range (inclusive bounds), -1 if the position occurs after
 * it, and +1 if the position occurs before it.
 *
 * @param range The range.
 * @param position The position.
 */
function comparePosition(range: FlattenedRange, position: lsp.Position): number {
    if (position.line < range.startLine) {
        return +1
    }

    if (position.line > range.endLine) {
        return -1
    }

    if (position.line === range.startLine && position.character < range.startCharacter) {
        return +1
    }

    if (position.line === range.endLine && position.character > range.endCharacter) {
        return -1
    }

    return 0
}
