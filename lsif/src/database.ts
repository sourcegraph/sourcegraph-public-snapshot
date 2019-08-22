import * as lsp from 'vscode-languageserver-protocol'
import { DocumentModel, DefModel, MetaModel, RefModel, PackageModel } from './models'
import { Connection } from 'typeorm'
import { decodeJSON } from './encoding'
import { MonikerData, RangeData, ResultSetData, DocumentData } from './entities'
import { Id } from 'lsif-protocol'
import { makeFilename } from './backend'
import { XrepoDatabase } from './xrepo'
import { ConnectionCache, DocumentCache } from './cache'

/**
 * `Database` wraps operations around a single repository/commit pair.
 */
export class Database {
    /**
     * Create a new `Database` with the given cross-repo database instance and the
     * filename of the database that contains data for a particular repository/commit.
     *
     * @param xrepoDatabase The cross-repo databse.
     * @param connectionCache The cache of SQLite connections.
     * @param documentCache The cache of loaded document.
     * @param database The filename of the database.
     */
    constructor(
        private xrepoDatabase: XrepoDatabase,
        private connectionCache: ConnectionCache,
        private documentCache: DocumentCache,
        private database: string
    ) {}

    /**
     * Determine if data exists for a particular document in this database.
     *
     * @param uri The URI of the document.
     */
    public async exists(uri: string): Promise<boolean> {
        return (await this.findDocument(uri)) !== undefined
    }

    /**
     * Return the location for the definition of the reference at the given position.
     *
     * @param uri The document to which the position belongs.
     * @param position The current hover position.
     */
    public async definitions(uri: string, position: lsp.Position): Promise<lsp.Location | lsp.Location[] | undefined> {
        const { document, range } = await this.findRange(uri, position)
        if (!document || !range) {
            return undefined
        }

        const resultData = findResult(document.resultSets, document.definitionResults, range, 'definitionResult')
        if (resultData) {
            return asLocations(document.ranges, document.orderedRanges, uri, resultData.values)
        }

        // TODO - vet this logic of finding the first result
        for (const moniker of findMonikers(document.resultSets, document.monikers, range)) {
            if (moniker.kind === 'import') {
                return await this.remoteDefinitions(document, moniker)
            }

            const defs = await Database.monikerResults(this, DefModel, moniker, uri => uri)
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
        const { document, range } = await this.findRange(uri, position)
        if (!document || !range) {
            return undefined
        }

        // TODO - vet logic - should xrepo search be explicitly enabled via query flag?
        const resultData = findResult(document.resultSets, document.referenceResults, range, 'referenceResult')
        const monikers = findMonikers(document.resultSets, document.monikers, range)

        const result = []
        if (resultData) {
            result.push(...asLocations(document.ranges, document.orderedRanges, uri, resultData.definitions))
            result.push(...asLocations(document.ranges, document.orderedRanges, uri, resultData.references))
        } else {
            for (const moniker of monikers) {
                result.push(...(await Database.monikerResults(this, RefModel, moniker, uri => uri)))
            }
        }

        for (const moniker of monikers) {
            if (moniker.kind === 'export') {
                const moreResult = await this.remoteReferences(document, moniker)
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
        const { document, range } = await this.findRange(uri, position)
        if (!document || !range) {
            return undefined
        }

        // All hover contents should be contained in the document.
        return findResult(document.resultSets, document.hovers, range, 'hoverResult')
    }

    //
    // Helper Functions

    /**
     * Query the defs or refs table of `db` for items that match the given moniker. Convert
     * each result into an LSP location. The `uriFilter` function is invoked on each result
     * item to modify the resulting locations.
     *
     * @param db The target database.
     * @param model The constructor for `T`.
     * @param moniker The target moniker.
     * @param uriFilter The function used to alter location uris.
     */
    private static async monikerResults<T extends DefModel | RefModel>(
        db: Database,
        model: Function,
        moniker: MonikerData,
        uriFilter: (uri: string) => string
    ): Promise<lsp.Location[]> {
        const results = await db.withConnection(connection =>
            connection.getRepository<T>(model).find({
                where: {
                    scheme: moniker.scheme,
                    identifier: moniker.identifier,
                },
            })
        )

        return results.map(result => lsp.Location.create(uriFilter(result.documentUri), makeRange(result)))
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
    private async remoteDefinitions(
        document: DocumentData,
        moniker: MonikerData
    ): Promise<lsp.Location | lsp.Location[] | undefined> {
        if (!moniker.packageInformation) {
            return undefined
        }

        const packageInformation = document.packageInformation.get(moniker.packageInformation)
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

        const db = new Database(
            this.xrepoDatabase,
            this.connectionCache,
            this.documentCache,
            makeFilename(packageEntity.repository, packageEntity.commit)
        )

        // FIXME
        fixMonikerIdentifier(moniker)
        const uriFilter = (uri: string): string => makeRemoteUri(packageEntity, uri)
        return await Database.monikerResults(db, DefModel, moniker, uriFilter)
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

        const allReferences = []
        for (const reference of references) {
            const db = new Database(
                this.xrepoDatabase,
                this.connectionCache,
                this.documentCache,
                makeFilename(reference.repository, reference.commit)
            )

            fixMonikerIdentifier(moniker)
            const uriFilter = (uri: string): string => makeRemoteUri(reference, uri)
            const references = await Database.monikerResults(db, RefModel, moniker, uriFilter)
            allReferences.push(...references)
        }

        return allReferences
    }

    /**
     * Return a parsed document that describes the given URI. The result of this
     * method is cached across all database instances.
     *
     * @param uri The uri of the document.
     */
    private async findDocument(uri: string): Promise<DocumentData | undefined> {
        const factory = async (): Promise<DocumentData> => {
            const document = await this.withConnection(connection =>
                connection.getRepository(DocumentModel).findOneOrFail(uri)
            )

            return await decodeJSON<DocumentData>(document.value)
        }

        return await this.documentCache.withDocument(`${this.database}::${uri}`, factory, document =>
            Promise.resolve(document)
        )
    }

    /**
     * Return a parsed document that describes the given URI as well as the range
     * from that document that contains the given position. Returns undefined for
     * both values if one cannot be loaded.
     *
     * @param uri The uri of the document.
     * @param position The uri's hover position.
     */
    private async findRange(
        uri: string,
        position: lsp.Position
    ): Promise<{ document: DocumentData | undefined; range: RangeData | undefined }> {
        const document = await this.findDocument(uri)
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
            this.database,
            [DefModel, DocumentModel, MetaModel, RefModel],
            callback
        )
    }
}

/**
 * Perform binary search over the ordered rnages of a document, returning
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
    return withChain(resultSets, data, current => {
        const value = current[property]
        if (value) {
            return map.get(value)
        }

        return undefined
    })
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

/**
 * Invoke the `visitor` function on each object in the chain of `next`
 * edges starting at `data`. If the `visitor` function returns a non
 * undefined value, it will be returned immediately.
 *
 * @param resultSets The map of results sets of the document.
 * @param data The range or result set object.
 * @param visitor The visitor function to invoke.
 */
function withChain<T>(
    resultSets: Map<Id, ResultSetData>,
    data: RangeData | ResultSetData,
    visitor: (current: RangeData | ResultSetData) => T | undefined
): T | undefined {
    let current: RangeData | ResultSetData | undefined = data
    while (current) {
        const value = visitor(current)
        if (value) {
            return value
        }

        if (!current.next) {
            break
        }

        current = resultSets.get(current.next)
    }

    return undefined
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
 * @param ranges The map of ranges of the document.
 * @param orderedRanges The ordered ranges of the document.
 * @param uri The location URI.
 * @param ids The set of ids.
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
                start: range.start,
                end: range.end,
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
    return `git://${pkg.repository}?${pkg.commit}#${path}`
}

/**
 * Construct an LSP range from a flattened four-tuple of numbers.
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
function comparePosition(range: lsp.Range, position: lsp.Position): number {
    if (position.line < range.start.line) {
        return +1
    }

    if (position.line > range.end.line) {
        return -1
    }

    if (position.line === range.start.line && position.character < range.start.character) {
        return +1
    }

    if (position.line === range.end.line && position.character > range.end.character) {
        return -1
    }

    return 0
}

// TODO - make this unnecessary, or figure out why it needs to stay
function fixMonikerIdentifier(moniker: MonikerData): void {
    const parts = moniker.identifier.split(':')
    parts[1] = ''
    moniker.identifier = parts.join(':')
}
