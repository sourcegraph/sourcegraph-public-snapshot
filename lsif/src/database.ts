import * as lsp from 'vscode-languageserver-protocol'
import { mustGet, hashKey } from './util'
import { Connection } from 'typeorm'
import { ConnectionCache, DocumentCache, EncodedJsonCacheValue, ResultChunkCache } from './cache'
import { gunzipJSON } from './encoding'
import { DefaultMap } from './default-map'
import {
    DefinitionModel,
    DocumentData,
    DocumentModel,
    MetaModel,
    MonikerData,
    RangeData,
    ReferenceModel,
    ResultChunkData,
    ResultChunkModel,
    DocumentPathRangeId,
    DefinitionReferenceResultId,
    RangeId,
} from './models.database'
import { isEqual, uniqWith } from 'lodash'
import { makeFilename } from './backend'
import { PackageModel } from './models.xrepo'
import { XrepoDatabase } from './xrepo'

/**
 * A wrapper around operations for single repository/commit pair.
 */
export class Database {
    /**
     * A static map of database paths to the `numResultChunks` value of their
     * metadata row. This map is populated lazily as the values are needed.
     */
    private static numResultChunks = new Map<string, number>()

    /**
     * Create a new `Database` with the given cross-repo database instance and the
     * filename of the database that contains data for a particular repository/commit.
     *
     * @param storageRoot The path where SQLite databases are stored.
     * @param xrepoDatabase The cross-repo database.
     * @param connectionCache The cache of SQLite connections.
     * @param documentCache The cache of loaded documents.
     * @param resultChunkCache The cache of loaded result chunks.
     * @param repository The repository for which this database answers queries.
     * @param commit The commit for which this database answers queries.
     * @param databasePath The path to the database file.
     */
    constructor(
        private storageRoot: string,
        private xrepoDatabase: XrepoDatabase,
        private connectionCache: ConnectionCache,
        private documentCache: DocumentCache,
        private resultChunkCache: ResultChunkCache,
        private repository: string,
        private commit: string,
        private databasePath: string
    ) {}

    /**
     * Determine if data exists for a particular document in this database.
     *
     * @param path The path of the document.
     */
    public async exists(path: string): Promise<boolean> {
        return (await this.getDocumentByPath(path)) !== undefined
    }

    /**
     * Return the location for the definition of the reference at the given position.
     *
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     */
    public async definitions(path: string, position: lsp.Position): Promise<lsp.Location[]> {
        const { document, range } = await this.getRangeByPosition(path, position)
        if (!document || !range) {
            return []
        }

        // First, we try to find the definition result attached to the range or one
        // of the result sets to which the range is attached.

        if (range.definitionResultId) {
            // We have a definition result in this database.
            const definitionResults = await this.getResultById(range.definitionResultId)

            // TODO - due to some bugs in tsc... this fixes the tests and some typescript examples
            // Not sure of a better way to do this right now until we work through how to patch
            // lsif-tsc to handle node_modules inclusion (or somehow blacklist it on import).

            if (!definitionResults.some(v => v.documentPath.includes('node_modules'))) {
                return await this.convertRangesToLspLocations(path, document, definitionResults)
            }
        }

        // Otherwise, we fall back to a moniker search. We get all the monikers attached
        // to the range or a result set to which the range is attached. We process each
        // moniker sequentially in order of priority, where import monikers, if any exist,
        // will be processed first.

        for (const moniker of sortMonikers(
            Array.from(range.monikerIds).map(id => mustGet(document.monikers, id, 'moniker'))
        )) {
            if (moniker.kind === 'import') {
                // This symbol was imported from another database. See if we have xrepo
                // definition for it.

                const remoteDefinitions = await this.remoteDefinitions(document, moniker)
                if (remoteDefinitions) {
                    return remoteDefinitions
                }
            } else {
                // This symbol was not imported from another database. We search the definitions
                // table of our own database in case there was a definition that wasn't properly
                // attached to a result set but did have the correct monikers attached.

                const localDefinitions = await Database.monikerResults(this, DefinitionModel, moniker, path => path)
                if (localDefinitions) {
                    return localDefinitions
                }
            }
        }

        return []
    }

    /**
     * Return a list of locations which reference the definition at the given position.
     *
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     */
    public async references(path: string, position: lsp.Position): Promise<lsp.Location[]> {
        const { document, range } = await this.getRangeByPosition(path, position)
        if (!document || !range) {
            return []
        }

        let locations: lsp.Location[] = []

        // First, we try to find the reference result attached to the range or one
        // of the result sets to which the range is attached.

        if (range.referenceResultId) {
            // We have references in this database.
            locations = locations.concat(
                await this.convertRangesToLspLocations(
                    path,
                    document,
                    await this.getResultById(range.referenceResultId)
                )
            )
        }

        // Next, we do a moniker search in two stages, described below. We process each
        // moniker sequentially in order of priority for each stage, where import monikers,
        // if any exist, will be processed first.

        const monikers = sortMonikers(
            Array.from(range.monikerIds).map(id => mustGet(document.monikers, id, 'monikers'))
        )

        // Next, we search the references table of our own database - this search is necessary,
        // but may be un-intuitive, but remember that a 'Find References' operation on a reference
        // should also return references to the definition. These are not necessarily fully linked
        // in the LSIF data.

        for (const moniker of monikers) {
            locations = locations.concat(await Database.monikerResults(this, ReferenceModel, moniker, path => path))
        }

        // Next, we perform an xrepo search for uses of each nonlocal moniker. We stop processing after
        // the first moniker for which we received results. As we process monikers in an order that
        // considers moniker schemes, the first one to get results should be the most desirable.

        for (const moniker of monikers) {
            if (moniker.kind === 'import') {
                // Get locations in the defining package
                locations = locations.concat(await this.remoteMoniker(document, moniker))
            }

            // Get locations in all packages
            const remoteResults = await this.remoteReferences(document, moniker)
            if (remoteResults) {
                // TODO - determine source of duplication (and below)
                return uniqWith(locations.concat(remoteResults), isEqual)
            }
        }

        return uniqWith(locations, isEqual)
    }

    /**
     * Return the hover content for the definition or reference at the given position.
     *
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     */
    public async hover(path: string, position: lsp.Position): Promise<lsp.Hover | null> {
        const { document, range } = await this.getRangeByPosition(path, position)
        if (!document || !range) {
            return null
        }

        // Try to find the hover content attached to the range or one of the result sets to
        // which the range is attached. There is no fall-back search via monikers for this
        // operation.

        if (range.hoverResultId) {
            return {
                contents: {
                    kind: lsp.MarkupKind.Markdown,
                    value: mustGet(document.hoverResults, range.hoverResultId, 'hoverResult'),
                },
                range: createRange(range),
            }
        }

        return null
    }

    //
    // Helper Functions

    /**
     * Convert a set of range-document pairs (from a definition or reference query) into
     * a set of LSP ranges. Each pair holds the range identifier as well as the document
     * path. For document paths matching the loaded document, find the range data locally.
     * For all other paths, find the document in this database and find the range in that
     * document.
     *
     * @param path The path of the document for this query.
     * @param document The document object for this query.
     * @param resultData A list of range ids and the document they belong to.
     */
    private async convertRangesToLspLocations(
        path: string,
        document: DocumentData,
        resultData: DocumentPathRangeId[]
    ): Promise<lsp.Location[]> {
        // Group by document path so we only have to load each document once
        const groupedResults = new DefaultMap<string, Set<RangeId>>(() => new Set())

        for (const { documentPath, rangeId } of resultData) {
            groupedResults.getOrDefault(documentPath).add(rangeId)
        }

        let results: lsp.Location[] = []
        for (const [documentPath, rangeIds] of groupedResults) {
            if (documentPath === path) {
                // If the document path is this document, convert the locations directly
                results = results.concat(mapRangesToLocations(document.ranges, path, rangeIds))
                continue
            }

            // Otherwise, we need to get the correct document
            const sibling = await this.getDocumentByPath(documentPath)
            if (!sibling) {
                continue
            }

            // Then finally convert the locations in the sibling document
            results = results.concat(mapRangesToLocations(sibling.ranges, documentPath, rangeIds))
        }

        return results
    }

    /**
     * Query the definitions or references table of `db` for items that match the given moniker.
     * Convert each result into an LSP location. The `pathTransformer` function is invoked on each
     * result item to modify the resulting locations.
     *
     * @param db The target database.
     * @param model The constructor for the model type.
     * @param moniker The target moniker.
     * @param pathTransformer The function used to alter location paths.
     */
    private static async monikerResults(
        db: Database,
        model: typeof DefinitionModel | typeof ReferenceModel,
        moniker: MonikerData,
        pathTransformer: (path: string) => string
    ): Promise<lsp.Location[]> {
        const results = await db.withConnection(connection =>
            connection.getRepository<DefinitionModel | ReferenceModel>(model).find({
                where: {
                    scheme: moniker.scheme,
                    identifier: moniker.identifier,
                },
            })
        )

        return results.map(result => lsp.Location.create(pathTransformer(result.documentPath), createRange(result)))
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
        if (!moniker.packageInformationId) {
            return null
        }

        const packageInformation = document.packageInformation.get(moniker.packageInformationId)
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

        const db = this.createNewDatabase(
            packageEntity.repository,
            packageEntity.commit,
            makeFilename(this.storageRoot, packageEntity.repository, packageEntity.commit)
        )

        const pathTransformer = (path: string): string => createRemoteUri(packageEntity, path)
        return await Database.monikerResults(db, DefinitionModel, moniker, pathTransformer)
    }

    /**
     * Find the references of of the target moniker inside the database where that moniker is defined.
     *
     * @param document The document containing the definition.
     * @param moniker The target moniker.
     */
    private async remoteMoniker(document: DocumentData, moniker: MonikerData): Promise<lsp.Location[]> {
        if (!moniker.packageInformationId) {
            return []
        }

        const packageInformation = document.packageInformation.get(moniker.packageInformationId)
        if (!packageInformation) {
            return []
        }

        const packageEntity = await this.xrepoDatabase.getPackage(
            moniker.scheme,
            packageInformation.name,
            packageInformation.version
        )

        if (!packageEntity) {
            return []
        }

        const db = this.createNewDatabase(
            packageEntity.repository,
            packageEntity.commit,
            makeFilename(this.storageRoot, packageEntity.repository, packageEntity.commit)
        )

        const pathTransformer = (path: string): string => createRemoteUri(packageEntity, path)
        return await Database.monikerResults(db, ReferenceModel, moniker, pathTransformer)
    }

    /**
     * Find the references of the target moniker outside of the current database. If the moniker
     * has attached package information, then the xrepo database is queried for the packages that
     * require this particular moniker identifier. These databases are opened, and their ref tables
     * are queried for the target moniker.
     *
     * @param document The document containing the definition.
     * @param moniker The target moniker.
     */
    private async remoteReferences(document: DocumentData, moniker: MonikerData): Promise<lsp.Location[]> {
        if (!moniker.packageInformationId) {
            return []
        }

        const packageInformation = document.packageInformation.get(moniker.packageInformationId)
        if (!packageInformation) {
            return []
        }

        const references = await this.xrepoDatabase.getReferences({
            scheme: moniker.scheme,
            name: packageInformation.name,
            version: packageInformation.version,
            value: moniker.identifier,
        })

        let allReferences: lsp.Location[] = []
        for (const reference of references) {
            // Skip the remote reference that show up for ourselves - we've already gathered
            // these in the previous step of the references query.
            if (reference.repository === this.repository && reference.commit === this.commit) {
                continue
            }

            const db = this.createNewDatabase(
                reference.repository,
                reference.commit,
                makeFilename(this.storageRoot, reference.repository, reference.commit)
            )

            const pathTransformer = (path: string): string => createRemoteUri(reference, path)
            const references = await Database.monikerResults(db, ReferenceModel, moniker, pathTransformer)
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
    private async getDocumentByPath(path: string): Promise<DocumentData | undefined> {
        const factory = async (): Promise<EncodedJsonCacheValue<DocumentData>> => {
            const document = await this.withConnection(connection =>
                connection.getRepository(DocumentModel).findOneOrFail(path)
            )

            return {
                size: document.data.length,
                data: await gunzipJSON<DocumentData>(document.data),
            }
        }

        return await this.documentCache.withValue(`${this.databasePath}::${path}`, factory, document =>
            Promise.resolve(document.data)
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
    private async getRangeByPosition(
        path: string,
        position: lsp.Position
    ): Promise<{ document: DocumentData | undefined; range: RangeData | undefined }> {
        const document = await this.getDocumentByPath(path)
        if (!document) {
            return { document: undefined, range: undefined }
        }

        for (const range of document.ranges.values()) {
            if (comparePosition(range, position) === 0) {
                return { document, range }
            }
        }

        return { document: undefined, range: undefined }
    }

    /**
     * Convert a list of ranges with document ids into a list of ranges with
     * document paths by looking into the result chunks table and parsing the
     * data associated with the given identifier.
     *
     * @param id The identifier of the definition or reference result.
     */
    private async getResultById(id: DefinitionReferenceResultId): Promise<DocumentPathRangeId[]> {
        const { documentPaths, documentIdRangeIds } = await this.getResultChunkByResultId(id)
        const ranges = mustGet(documentIdRangeIds, id, 'documentIdRangeId')

        return ranges.map(range => ({
            documentPath: mustGet(documentPaths, range.documentId, 'documentPath'),
            rangeId: range.rangeId,
        }))
    }

    /**
     * Return a parsed result chunk that contains the given identifier.
     *
     * @param id An identifier contained in the result chunk.
     */
    private async getResultChunkByResultId(id: DefinitionReferenceResultId): Promise<ResultChunkData> {
        // Find the result chunk index this id belongs to
        const index = hashKey(id, await this.getNumResultChunks())

        const factory = async (): Promise<EncodedJsonCacheValue<ResultChunkData>> => {
            const resultChunk = await this.withConnection(connection =>
                connection.getRepository(ResultChunkModel).findOneOrFail(index)
            )

            return {
                size: resultChunk.data.length,
                data: await gunzipJSON<ResultChunkData>(resultChunk.data),
            }
        }

        return await this.resultChunkCache.withValue(`${this.databasePath}::${index}`, factory, resultChunk =>
            Promise.resolve(resultChunk.data)
        )
    }

    /**
     * Get the `numResultChunks` value from this database's metadata row.
     */
    private async getNumResultChunks(): Promise<number> {
        const numResultChunks = Database.numResultChunks.get(this.databasePath)
        if (numResultChunks !== undefined) {
            return numResultChunks
        }

        // Not in the shared map, need to query it
        const meta = await this.withConnection(connection => connection.getRepository(MetaModel).findOneOrFail(1))
        Database.numResultChunks.set(this.databasePath, meta.numResultChunks)
        return meta.numResultChunks
    }

    /**
     * Create a new database with the same configuration but a different repository,
     * commit, and databasePath.
     *
     *
     * @param repository The repository for which this database answers queries.
     * @param commit The commit for which this database answers queries.
     * @param databasePath The path to the database file.
     */
    private createNewDatabase(repository: string, commit: string, databasePath: string): Database {
        return new Database(
            this.storageRoot,
            this.xrepoDatabase,
            this.connectionCache,
            this.documentCache,
            this.resultChunkCache,
            repository,
            commit,
            databasePath
        )
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
            [DefinitionModel, DocumentModel, MetaModel, ReferenceModel, ResultChunkModel],
            callback
        )
    }
}

/**
 * Compare a position against a range. Returns 0 if the position occurs
 * within the range (inclusive bounds), -1 if the position occurs after
 * it, and +1 if the position occurs before it.
 *
 * @param range The range.
 * @param position The position.
 */
export function comparePosition(range: RangeData, position: lsp.Position): number {
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

/**
 * Sort the monikers by kind, then scheme in order of the following
 * preferences.
 *
 *   - kind: import, local, export
 *   - scheme: npm, tsc
 *
 * @param monikers The list of monikers.
 */
export function sortMonikers(monikers: MonikerData[]): MonikerData[] {
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
 * Construct a URI that can be used by the frontend to switch to another
 * directory.
 *
 * @param pkg The target package.
 * @param path The path relative to the project root.
 */
export function createRemoteUri(pkg: PackageModel, path: string): string {
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
function createRange(result: {
    startLine: number
    startCharacter: number
    endLine: number
    endCharacter: number
}): lsp.Range {
    return lsp.Range.create(result.startLine, result.startCharacter, result.endLine, result.endCharacter)
}

/**
 * Convert the given range identifiers into LSP location objects.
 *
 * @param ranges The map of ranges of the document.
 * @param uri The location URI.
 * @param ids The set of range identifiers for each resulting location.
 */
export function mapRangesToLocations(ranges: Map<RangeId, RangeData>, uri: string, ids: Set<RangeId>): lsp.Location[] {
    const locations = []
    for (const id of ids) {
        locations.push(lsp.Location.create(uri, createRange(mustGet(ranges, id, 'range'))))
    }

    return locations
}
