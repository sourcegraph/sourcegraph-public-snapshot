import * as lsp from 'vscode-languageserver-protocol'
import { Connection } from 'typeorm'
import { ConnectionCache, DocumentCache, EncodedJsonCacheValue, ResultChunkCache } from './cache'
import { dbFilename, hashKey, mustGet, hasErrorCode } from './util'
import { instrument } from './metrics'
import { databaseQueryDurationHistogram, databaseQueryErrorsCounter } from './database.metrics'
import { DefaultMap } from './default-map'
import { gunzipJSON } from './encoding'
import * as fs from 'mz/fs'
import { isEqual, uniqWith } from 'lodash'
import { PackageModel, DumpID } from './xrepo.models'
import { XrepoDatabase } from './xrepo'
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
    entities,
} from './database.models'
import { TracingContext, logSpan } from './tracing'

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
     * @param dumpID The ID of the dump for which this database answers queries.
     * @param databasePath The path to the database file.
     * @param root The path into the repository at which this database is mounted.
     */
    constructor(
        private storageRoot: string,
        private xrepoDatabase: XrepoDatabase,
        private connectionCache: ConnectionCache,
        private documentCache: DocumentCache,
        private resultChunkCache: ResultChunkCache,
        private dumpID: DumpID,
        private databasePath: string,
        private root: string
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
     * @param ctx The tracing context.
     */
    public async definitions(path: string, position: lsp.Position, ctx: TracingContext = {}): Promise<lsp.Location[]> {
        const { document, ranges } = await this.getRangeByPosition(path, position, ctx)
        if (!document || ranges.length === 0) {
            return []
        }

        for (const range of ranges) {
            if (!range.definitionResultId) {
                continue
            }

            // We have a definition result in this database, return this first.
            const definitionResults = await this.getResultById(range.definitionResultId)
            this.logSpan(ctx, 'definition_results', { definitionResultId: range.definitionResultId, definitionResults })

            // TODO - due to some bugs in tsc... this fixes the tests and some typescript examples
            // Not sure of a better way to do this right now until we work through how to patch
            // lsif-tsc to handle node_modules inclusion (or somehow blacklist it on import).

            if (!definitionResults.some(v => v.documentPath.includes('node_modules'))) {
                return await this.convertRangesToLspLocations(path, document, definitionResults)
            }
        }

        // No definition result in this repo; fall back to a moniker search. First, we find
        // the monikers for each range, from innermost to outermost, such that the set of
        // monikers for reach range is sorted by priority. Then, we perform a search for
        // each moniker, in sequence, until valid results are found.

        for (const range of ranges) {
            const monikers = sortMonikers(
                Array.from(range.monikerIds).map(id => mustGet(document.monikers, id, 'moniker'))
            )

            for (const moniker of monikers) {
                if (moniker.kind === 'import') {
                    // This symbol was imported from another database. See if we have xrepo
                    // definition for it.

                    const remoteDefinitions = await this.remoteDefinitions(document, moniker, ctx)
                    if (remoteDefinitions) {
                        return remoteDefinitions
                    }
                } else {
                    // This symbol was not imported from another database. We search the definitions
                    // table of our own database in case there was a definition that wasn't properly
                    // attached to a result set but did have the correct monikers attached.

                    const localDefinitions = await this.monikerResults(DefinitionModel, moniker, path => path, ctx)
                    if (localDefinitions) {
                        return localDefinitions
                    }
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
     * @param ctx The tracing context.
     */
    public async references(path: string, position: lsp.Position, ctx: TracingContext = {}): Promise<lsp.Location[]> {
        const { document, ranges } = await this.getRangeByPosition(path, position, ctx)
        if (!document || ranges.length === 0) {
            return []
        }

        let locations: lsp.Location[] = []

        // First, we try to find the reference result attached to the range or one
        // of the result sets to which the range is attached.

        for (const range of ranges) {
            if (range.referenceResultId) {
                // We have references in this database.
                const referenceResults = await this.getResultById(range.referenceResultId)
                this.logSpan(ctx, 'reference_results', { referenceResultId: range.referenceResultId, referenceResults })
                locations = locations.concat(await this.convertRangesToLspLocations(path, document, referenceResults))
            }
        }

        // Next, we do a moniker search in two stages, described below. We process the
        // monikers for each range sequentially in order of priority for each stage, such
        // that import monikers, if any exist, will be processed first.

        for (const range of ranges) {
            const monikers = sortMonikers(
                Array.from(range.monikerIds).map(id => mustGet(document.monikers, id, 'monikers'))
            )

            // Next, we search the references table of our own database - this search is necessary,
            // but may be un-intuitive, but remember that a 'Find References' operation on a reference
            // should also return references to the definition. These are not necessarily fully linked
            // in the LSIF data.

            for (const moniker of monikers) {
                locations = locations.concat(await this.monikerResults(ReferenceModel, moniker, path => path, ctx))
            }

            // Next, we perform an xrepo search for uses of each nonlocal moniker. We stop processing after
            // the first moniker for which we received results. As we process monikers in an order that
            // considers moniker schemes, the first one to get results should be the most desirable.

            for (const moniker of monikers) {
                if (moniker.kind === 'import') {
                    // Get locations in the defining package
                    locations = locations.concat(await this.remoteMoniker(document, moniker, ctx))
                }

                // Get locations in all packages
                const remoteResults = await this.remoteReferences(document, moniker, ctx)
                if (remoteResults) {
                    // TODO - determine source of duplication (and below)
                    return uniqWith(locations.concat(remoteResults), isEqual)
                }
            }
        }

        return uniqWith(locations, isEqual)
    }

    /**
     * Return the hover content for the definition or reference at the given position.
     *
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     * @param ctx The tracing context.
     */
    public async hover(path: string, position: lsp.Position, ctx: TracingContext = {}): Promise<lsp.Hover | null> {
        const { document, ranges } = await this.getRangeByPosition(path, position, ctx)
        if (!document || ranges.length === 0) {
            return null
        }

        for (const range of ranges) {
            if (range.hoverResultId) {
                const contents = {
                    kind: lsp.MarkupKind.Markdown,
                    value: mustGet(document.hoverResults, range.hoverResultId, 'hoverResult'),
                }

                // Return first defined hover result for the inner-most range
                return { contents, range: createRange(range) }
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
     * @param model The constructor for the model type.
     * @param moniker The target moniker.
     * @param pathTransformer The function used to alter location paths.
     * @param ctx The tracing context.
     */
    private async monikerResults(
        model: typeof DefinitionModel | typeof ReferenceModel,
        moniker: MonikerData,
        pathTransformer: (path: string) => string,
        ctx: TracingContext
    ): Promise<lsp.Location[]> {
        const results = await this.withConnection(connection =>
            connection.getRepository<DefinitionModel | ReferenceModel>(model).find({
                where: {
                    scheme: moniker.scheme,
                    identifier: moniker.identifier,
                },
            })
        )

        this.logSpan(ctx, 'symbol_results', { moniker, symbol: results })
        return results.map(result => lsp.Location.create(pathTransformer(result.documentPath), createRange(result)))
    }

    /**
     * Find the definition of the target moniker outside of the current database.
     *
     * @param document The document containing the reference.
     * @param moniker The target moniker.
     * @param ctx The tracing context.
     */
    private remoteDefinitions(
        document: DocumentData,
        moniker: MonikerData,
        ctx: TracingContext
    ): Promise<lsp.Location[] | null> {
        return this.lookupMoniker(document, moniker, DefinitionModel, ctx)
    }

    /**
     * Find the references of the target moniker inside the database where that moniker is defined.
     *
     * @param document The document containing the definition.
     * @param moniker The target moniker.
     * @param ctx The tracing context.
     */
    private remoteMoniker(document: DocumentData, moniker: MonikerData, ctx: TracingContext): Promise<lsp.Location[]> {
        return this.lookupMoniker(document, moniker, ReferenceModel, ctx)
    }

    /**
     * Find the locations attached to the target moniker outside of the current database. If
     * the moniker has attached package information, then the cross-repo database is queried
     * for the target package. That database is opened, and its definitions table is queried
     * for the target moniker.
     *
     * @param document The document containing the definition.
     * @param moniker The target moniker.
     * @param model The target model.
     * @param ctx The tracing context.
     */
    private async lookupMoniker(
        document: DocumentData,
        moniker: MonikerData,
        model: typeof DefinitionModel | typeof ReferenceModel,
        ctx: TracingContext
    ): Promise<lsp.Location[]> {
        if (!moniker.packageInformationId) {
            return []
        }

        const packageInformation = document.packageInformation.get(moniker.packageInformationId)
        if (!packageInformation) {
            return []
        }

        this.logSpan(ctx, 'package_information', {
            moniker,
            packageInformation,
        })

        const packageEntity = await this.xrepoDatabase.getPackage(
            moniker.scheme,
            packageInformation.name,
            packageInformation.version
        )
        if (!packageEntity) {
            return []
        }

        this.logSpan(ctx, 'package_entity', {
            moniker,
            packageInformation,
            packageRepository: packageEntity.dump.repository,
            packageCommit: packageEntity.dump.commit,
        })

        const db = this.createNewDatabase(
            packageEntity.dump.id,
            dbFilename(
                this.storageRoot,
                packageEntity.dump.id,
                packageEntity.dump.repository,
                packageEntity.dump.commit
            )
        )

        const pathTransformer = (path: string): string => createRemoteUri(packageEntity, path)
        return await db.monikerResults(model, moniker, pathTransformer, ctx)
    }

    /**
     * Find the references of the target moniker outside of the current database. If the moniker
     * has attached package information, then the cross-repo database is queried for the packages
     * that require this particular moniker identifier. These databases are opened, and their
     * references tables are queried for the target moniker.
     *
     * @param document The document containing the definition.
     * @param moniker The target moniker.
     * @param ctx The tracing context.
     */
    private async remoteReferences(
        document: DocumentData,
        moniker: MonikerData,
        ctx: TracingContext
    ): Promise<lsp.Location[]> {
        if (!moniker.packageInformationId) {
            return []
        }

        const packageInformation = document.packageInformation.get(moniker.packageInformationId)
        if (!packageInformation) {
            return []
        }

        this.logSpan(ctx, 'package_information', {
            moniker,
            packageInformation,
        })

        const references = await this.xrepoDatabase.getReferences({
            scheme: moniker.scheme,
            name: packageInformation.name,
            version: packageInformation.version,
            value: moniker.identifier,
        })

        this.logSpan(ctx, 'package_references', {
            moniker,
            packageInformation,
            references: references.map(r => ({ repository: r.dump.repository, commit: r.dump.commit })),
        })

        let allReferences: lsp.Location[] = []
        for (const reference of references) {
            // Skip the remote reference that show up for ourselves - we've already gathered
            // these in the previous step of the references query.
            if (reference.dump.id === this.dumpID) {
                continue
            }

            const db = this.createNewDatabase(
                reference.dump.id,
                dbFilename(this.storageRoot, reference.dump.id, reference.dump.repository, reference.dump.commit)
            )

            const pathTransformer = (path: string): string => createRemoteUri(reference, path)
            const references = await db.monikerResults(ReferenceModel, moniker, pathTransformer, ctx)
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

        try {
            return await this.documentCache.withValue(`${this.databasePath}::${path}`, factory, document =>
                Promise.resolve(document.data)
            )
        } catch (error) {
            if (error.name === 'EntityNotFound') {
                return undefined
            }

            throw error
        }
    }

    /**
     * Return a parsed document that describes the given path as well as the ranges
     * from that document that contains the given position. If multiple ranges are
     * returned, then the inner-most ranges will occur before the outer-most ranges.
     *
     * @param path The path of the document.
     * @param position The user's hover position.
     * @param ctx The tracing context.
     */
    private async getRangeByPosition(
        path: string,
        position: lsp.Position,
        ctx: TracingContext
    ): Promise<{ document: DocumentData | undefined; ranges: RangeData[] }> {
        const document = await this.getDocumentByPath(path)
        if (!document) {
            return { document: undefined, ranges: [] }
        }

        const ranges = findRanges(document.ranges.values(), position)
        this.logSpan(ctx, 'matching_ranges', { ranges: cleanRanges(ranges) })
        return { document, ranges }
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
     * @param dumpID The ID of the dump for which this database answers queries.
     * @param databasePath The path to the database file.
     */
    private createNewDatabase(dumpID: DumpID, databasePath: string): Database {
        return new Database(
            this.storageRoot,
            this.xrepoDatabase,
            this.connectionCache,
            this.documentCache,
            this.resultChunkCache,
            dumpID,
            databasePath,
            this.root
        )
    }

    /**
     * Invoke `callback` with a SQLite connection object obtained from the
     * cache or created on cache miss.
     *
     * @param callback The function invoke with the SQLite connection.
     */
    private async withConnection<T>(callback: (connection: Connection) => Promise<T>): Promise<T> {
        return await this.connectionCache.withConnection(this.databasePath, entities, connection =>
            instrument(databaseQueryDurationHistogram, databaseQueryErrorsCounter, () => callback(connection))
        )
    }

    /**
     * Logs an event to the span of The tracing context, if its defined.
     *
     * @param ctx The tracing context.
     * @param event The name of the event.
     * @param pairs The values to log.
     */
    private logSpan(ctx: TracingContext, event: string, pairs: { [K: string]: any }): void {
        logSpan(ctx, event, { ...pairs, dbID: this.dumpID })
    }
}

/**
 * Return the set of ranges that contain the given position. If multiple ranges
 * are returned, then the inner-most ranges will occur before the outer-most
 * ranges.
 *
 * @param ranges The set of possible ranges.
 * @param position The user's hover position.
 */
export function findRanges(ranges: Iterable<RangeData>, position: lsp.Position): RangeData[] {
    const filtered = []
    for (const range of ranges) {
        if (comparePosition(range, position) === 0) {
            filtered.push(range)
        }
    }

    return filtered.sort((a, b) => {
        if (comparePosition(a, { line: b.startLine, character: b.startCharacter }) === 0) {
            return +1
        }

        return -1
    })
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
    const url = new URL(`git://${pkg.dump.repository}`)
    url.search = pkg.dump.commit
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

/**
 * Create a new `Database` backed by the given file. A databases is only
 * created if the file exists.
 *
 * @param storageRoot The path where SQLite databases are stored.
 * @param xrepoDatabase The cross-repo database.
 * @param connectionCache The cache of SQLite connections.
 * @param documentCache The cache of loaded documents.
 * @param resultChunkCache The cache of loaded result chunks.
 * @param dumpID The ID of the dump for which this database answers queries.
 * @param databasePath The path to the database file.
 */
export async function tryCreateDatabase(
    storageRoot: string,
    xrepoDatabase: XrepoDatabase,
    connectionCache: ConnectionCache,
    documentCache: DocumentCache,
    resultChunkCache: ResultChunkCache,
    dumpID: DumpID,
    databasePath: string,
    root: string
): Promise<Database | undefined> {
    try {
        await fs.stat(databasePath)
    } catch (e) {
        if (hasErrorCode(e, 'ENOENT')) {
            return undefined
        }

        throw e
    }

    return new Database(
        storageRoot,
        xrepoDatabase,
        connectionCache,
        documentCache,
        resultChunkCache,
        dumpID,
        databasePath,
        root
    )
}

/**
 * Format ranges to be serialized in opentracing logs.
 *
 * @param ranges The ranges to clean.
 */
function cleanRanges(ranges: RangeData[]): { [K: string]: any } {
    // We need to array-ize sets otherwise we get a "0 key" object
    return ranges.map(r => ({ ...r, monikerIds: Array.from(r.monikerIds) }))
}
