import * as sqliteModels from '../../shared/models/sqlite'
import * as lsp from 'vscode-languageserver-protocol'
import * as settings from '../settings'
import * as pgModels from '../../shared/models/pg'
import { addTags, logSpan, TracingContext } from '../../shared/tracing'
import { ConnectionCache, DocumentCache, ResultChunkCache } from './cache'
import { Database, sortMonikers, InternalLocation } from './database'
import { dbFilename } from '../../shared/paths'
import { mustGet } from '../../shared/maps'
import { DumpManager } from '../../shared/store/dumps'
import { DependencyManager } from '../../shared/store/dependencies'
import { isDefined } from '../../shared/util'
import {
    ReferencePaginationContext,
    ReferencePaginationCursor,
    RemoteDumpReferenceCursor,
    makeInitialSameDumpCursor,
    makeInitialSameDumpMonikersCursor,
    makeInitialSameRepoCursor,
    makeInitialRemoteRepoCursor,
    SameDumpReferenceCursor,
} from './cursor'
import { uniqWith, isEqual } from 'lodash'

/**
 * A wrapper around code intelligence operations.
 */
export class Backend {
    private connectionCache = new ConnectionCache(settings.CONNECTION_CACHE_CAPACITY)
    private documentCache = new DocumentCache(settings.DOCUMENT_CACHE_CAPACITY)
    private resultChunkCache = new ResultChunkCache(settings.RESULT_CHUNK_CACHE_CAPACITY)

    /**
     * Create a new `Backend`.
     *
     * @param storageRoot The path where SQLite databases are stored.
     * @param dumpManager The dumps manager instance.
     * @param dependencyManager The dependency manager instance.
     * @param frontendUrl The url of the frontend internal API.
     */
    constructor(
        private storageRoot: string,
        private dumpManager: DumpManager,
        private dependencyManager: DependencyManager,
        private frontendUrl: string
    ) {}

    /**
     * Determine if data exists for a particular document.
     *
     * @param repositoryId The repository identifier.
     * @param commit The commit.
     * @param path The path of the document.
     * @param ctx The tracing context.
     */
    public async exists(
        repositoryId: number,
        commit: string,
        path: string,
        ctx: TracingContext = {}
    ): Promise<pgModels.LsifDump[]> {
        return (await this.findClosestDatabases(repositoryId, commit, path, ctx)).map(({ dump }) => dump)
    }

    /**
     * Return the location for the symbol at the given position. Returns undefined if no dump can
     * be loaded to answer this query.
     *
     * @param repositoryId The repository identifier.
     * @param commit The commit.
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     * @param dumpId The identifier of the dump to load. If not supplied, the closest dump will be used.
     * @param ctx The tracing context.
     */
    public async definitions(
        repositoryId: number,
        commit: string,
        path: string,
        position: lsp.Position,
        dumpId?: number,
        ctx: TracingContext = {}
    ): Promise<InternalLocation[] | undefined> {
        const closestDatabaseAndDump = await this.closestDatabase(repositoryId, commit, path, dumpId, ctx)
        if (!closestDatabaseAndDump) {
            if (ctx.logger) {
                ctx.logger.warn('No database could be loaded', { repositoryId, commit, path })
            }

            return undefined
        }
        const { database, dump, ctx: newCtx } = closestDatabaseAndDump

        // Construct path within dump
        const pathInDb = pathToDatabase(dump.root, path)

        // Try to find definitions in the same dump
        const dbDefinitions = await database.definitions(pathInDb, position, newCtx)
        const definitions = dbDefinitions.map(loc => locationFromDatabase(dump.root, loc))
        if (definitions.length > 0) {
            return definitions
        }

        // Try to find definitions in other dumps
        const { document, ranges } = await database.getRangeByPosition(pathInDb, position, ctx)
        if (!document || ranges.length === 0) {
            return []
        }

        // First, we find the monikers for each range, from innermost to
        // outermost, such that the set of monikers for reach range is sorted by
        // priority. Then, we perform a search for each moniker, in sequence,
        // until valid results are found.
        for (const range of ranges) {
            const monikers = sortMonikers(
                Array.from(range.monikerIds).map(id => mustGet(document.monikers, id, 'moniker'))
            )

            for (const moniker of monikers) {
                if (moniker.kind === 'import') {
                    // This symbol was imported from another database. See if we have
                    // an remote definition for it.

                    const remoteDefinitions = await this.lookupMoniker(
                        document,
                        moniker,
                        sqliteModels.DefinitionModel,
                        {},
                        ctx
                    )
                    if (remoteDefinitions.length > 0) {
                        return remoteDefinitions
                    }
                } else {
                    // This symbol was not imported from another database. We search the definitions
                    // table of our own database in case there was a definition that wasn't properly
                    // attached to a result set but did have the correct monikers attached.

                    const { locations: monikerResults } = await database.monikerResults(
                        sqliteModels.DefinitionModel,
                        moniker,
                        {},
                        ctx
                    )
                    const localDefinitions = monikerResults.map(loc => locationFromDatabase(dump.root, loc))
                    if (localDefinitions.length > 0) {
                        return localDefinitions
                    }
                }
            }
        }
        return []
    }

    /**
     * Return a list of locations which reference the symbol at the given position. Returns
     * undefined if no dump can be loaded to answer this query.
     *
     * @param repositoryId The repository identifier.
     * @param commit The commit.
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     * @param paginationContext Context describing the current request for paginated results.
     * @param dumpId The identifier of the dump to load. If not supplied, the closest dump will be used.
     * @param ctx The tracing context.
     */
    public async references(
        repositoryId: number,
        commit: string,
        path: string,
        position: lsp.Position,
        paginationContext: ReferencePaginationContext = { limit: 10 },
        dumpId?: number,
        ctx: TracingContext = {}
    ): Promise<{ locations: InternalLocation[]; newCursor?: ReferencePaginationCursor } | undefined> {
        if (paginationContext.cursor) {
            return this.handleReferencePaginationCursor(
                repositoryId,
                commit,
                paginationContext.limit,
                paginationContext.cursor,
                ctx
            )
        }

        const closestDatabaseAndDump = await this.closestDatabase(repositoryId, commit, path, dumpId, ctx)
        if (!closestDatabaseAndDump) {
            if (ctx.logger) {
                ctx.logger.warn('No database could be loaded', { repositoryId, commit, path })
            }

            return undefined
        }
        const { database, dump, ctx: newCtx } = closestDatabaseAndDump

        // Construct path within dump
        const pathInDb = pathToDatabase(dump.root, path)

        // Try to find references in the same dump
        const dbReferences = await database.references(pathInDb, position, newCtx)
        const locations = dbReferences.map(loc => locationFromDatabase(dump.root, loc))

        // Get the ranges of for this position and the document in which they occur
        const { document, ranges } = await database.getRangeByPosition(pathInDb, position, ctx)
        if (!document || ranges.length === 0) {
            return { locations }
        }

        // Find and normalize the monikers attached to each range
        const monikers: sqliteModels.MonikerData[] = []
        for (const range of ranges) {
            for (const moniker of Array.from(range.monikerIds).map(id => mustGet(document.monikers, id, 'monikers'))) {
                monikers.push(moniker)
            }
        }

        // Immediately request the next page of results
        const { locations: remoteLocations, newCursor } = await this.handleReferencePaginationCursor(
            repositoryId,
            commit,
            paginationContext.limit,
            makeInitialSameDumpCursor({ dumpId: dump.id, path: pathInDb, monikers: sortMonikers(monikers) }),
            ctx
        )

        // TODO - is this still a problem?
        // TODO - find out where else we're calling uniqWith
        return { locations: uniqWith(locations.concat(remoteLocations), isEqual), newCursor }
    }

    /**
     * Return the hover content for the symbol at the given position. Returns undefined if no dump can
     * be loaded to answer this query.
     *
     * @param repositoryId The repository identifier.
     * @param commit The commit.
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     * @param dumpId The identifier of the dump to load. If not supplied, the closest dump will be used.
     * @param ctx The tracing context.
     */
    public async hover(
        repositoryId: number,
        commit: string,
        path: string,
        position: lsp.Position,
        dumpId?: number,
        ctx: TracingContext = {}
    ): Promise<{ text: string; range: lsp.Range } | null | undefined> {
        const closestDatabaseAndDump = await this.closestDatabase(repositoryId, commit, path, dumpId, ctx)
        if (!closestDatabaseAndDump) {
            if (ctx.logger) {
                ctx.logger.warn('No database could be loaded', { repositoryId, commit, path })
            }

            return undefined
        }
        const { database, dump, ctx: newCtx } = closestDatabaseAndDump

        // Try to find hover in the same dump
        const hover = await database.hover(pathToDatabase(dump.root, path), position, newCtx)
        if (hover !== null) {
            return hover
        }

        // If we don't have a local hover, lookup the definitions of the range and read the hover
        // data from the remote database. This can happen when the indexer only gives a moniker but
        // does not give hover data for externally defined symbols.

        const locations = await this.definitions(repositoryId, commit, path, position, dumpId, ctx)
        if (!locations || locations.length === 0) {
            return null
        }

        const { dump: definitionDump, path: definitionPath, range } = locations[0]
        const definitionDatabase = this.createDatabase(definitionDump)
        return definitionDatabase.hover(pathToDatabase(definitionDump.root, definitionPath), range.start, newCtx)
    }

    /**
     * Using the current state of the pagination cursor, determine what we need to query next. The
     * four major phases of a reference request are:
     *
     *   (1) 'same-dump': query the original dump's references table
     *   (2) 'same-dump-monikers': query the original dump's monikers
     *   (3) 'same-repo': open additional dumps that belong to the same repository
     *   (4) 'remote-repo': open additional dumps that belong to a different repository
     *
     * This method will return any locations found in this page of results as well as a cursor
     * indicating how to execute the next page of results. If the cursor is undefined there are no
     * more results.
     *
     * @param repositoryId The repository identifier.
     * @param commit The target commit.
     * @param limit The maximum number of dumps to open.
     * @param cursor The pagination cursor.
     * @param ctx The tracing context.
     */
    private async handleReferencePaginationCursor(
        repositoryId: number,
        commit: string,
        limit: number,
        cursor: ReferencePaginationCursor,
        ctx: TracingContext = {}
    ): Promise<{ locations: InternalLocation[]; newCursor?: ReferencePaginationCursor }> {
        switch (cursor.phase) {
            case 'same-dump': {
                return this.handleReferencePaginationCursorRecursive(
                    repositoryId,
                    commit,
                    limit,
                    50,
                    () => this.performSameDumpReferences(cursor, ctx),
                    () => makeInitialSameDumpMonikersCursor(cursor),
                    ctx
                )
            }

            case 'same-dump-monikers': {
                const makeCursor = async () => {
                    const document = await this.getDocumentByPath(cursor.dumpId, cursor.path, ctx)
                    if (!document) {
                        return undefined
                    }

                    for (const moniker of cursor.monikers) {
                        const packageInformation = this.lookupPackageInformation(document, moniker, ctx)
                        if (packageInformation) {
                            return makeInitialSameRepoCursor(cursor, moniker, packageInformation)
                        }
                    }
                    return undefined
                }

                return this.handleReferencePaginationCursorRecursive(
                    repositoryId,
                    commit,
                    limit,
                    50,
                    () => this.performSameDumpMonikerReferences(cursor, ctx),
                    makeCursor,
                    ctx
                )
            }

            case 'same-repo': {
                const makeCursor = async () => {
                    if (await this.hasRemoteReferences(repositoryId, cursor, ctx)) {
                        // If there are no remote consumers of this symbol, do not
                        // make a cursor for the next phase as it would only be a
                        // single empty page.
                        return makeInitialRemoteRepoCursor(cursor)
                    }

                    return undefined
                }

                return this.handleReferencePaginationCursorRecursive(
                    repositoryId,
                    commit,
                    limit,
                    0,
                    () => this.performSameRepositoryRemoteReferences(repositoryId, commit, limit, cursor, ctx),
                    makeCursor,
                    ctx
                )
            }

            case 'remote-repo': {
                return this.performRemoteReferences(repositoryId, limit, cursor, ctx)
            }
        }
    }

    /**
     * A helper function used by `handleReferencePaginationCursor`. This method takes a handler
     * that executes the current page of results and returns a new cursor for the **same phase**
     * of results. If there are no more results in that phase of the result set, the cursor is
     * undefined. In this case, we call the `makeCursor` factory function to construct the cursor
     * for the next phase of pagination.
     *
     * @param repositoryId The repository identifier.
     * @param commit The target commit.
     * @param limit The maximum number of dumps to open.
     * @param threshold Request the next page of results immediately if the number of results
     *     in the previous page is below this threshold. Concatenate the results together.
     * @param handler The handler for the current page of results.
     * @param makeCursor A factory that creates a cursor for the next phase of pagination.
     * @param ctx The tracing context.
     */
    private async handleReferencePaginationCursorRecursive(
        repositoryId: number,
        commit: string,
        limit: number,
        threshold: number,
        handler: () => Promise<{ locations: InternalLocation[]; newCursor?: ReferencePaginationCursor }>,
        makeCursor: () => Promise<ReferencePaginationCursor | undefined> | ReferencePaginationCursor | undefined,
        ctx: TracingContext = {}
    ): Promise<{ locations: InternalLocation[]; newCursor?: ReferencePaginationCursor }> {
        const { locations, newCursor: originalCursor } = await handler()
        const newCursor = originalCursor || (await makeCursor())
        if (locations.length > threshold || !newCursor) {
            return { locations, newCursor }
        }

        const {
            locations: nextPageLocations,
            newCursor: nextPageNewCursor,
        } = await this.handleReferencePaginationCursor(repositoryId, commit, limit, newCursor, ctx)

        return { locations: locations.concat(nextPageLocations), newCursor: nextPageNewCursor }
    }

    /**
     * Search the references table of the current dump.
     *
     * This search is necessary, but may be un-intuitive. A 'Find References' operation on a
     * reference should also return references to the definition. These are not necessarily
     * fully linked in the LSIF data.
     *
     * If there are any locations in the result set, this method returns the new cursor. This
     * method returns undefined if there are no remaining results for the same repository.
     *
     * @param repositoryId The repository identifier.
     * @param commit The target commit.
     * @param limit The maximum number of dumps to open.
     * @param cursor The pagination cursor.
     * @param ctx The tracing context.
     */
    private async performSameDumpReferences(
        cursor: SameDumpReferenceCursor,
        ctx: TracingContext = {}
    ): Promise<{ locations: InternalLocation[]; newCursor?: ReferencePaginationCursor }> {
        const dumpAndDatabase = await this.getDumpAndDatabaseById(cursor.dumpId)
        if (!dumpAndDatabase) {
            return { locations: [] }
        }
        const { dump, database } = dumpAndDatabase

        let locations: InternalLocation[] = []
        for (const moniker of cursor.monikers) {
            const { locations: monikerResults } = await database.monikerResults(
                sqliteModels.ReferenceModel,
                moniker,
                {}, // TODO - paginate
                ctx
            )
            locations = locations.concat(monikerResults.map(loc => locationFromDatabase(dump.root, loc)))
        }

        return { locations } // TODO - cursor
    }

    // TODO - fix documentation (we do not early-out monikers here - should we actually?)

    /**
 * Perform a remote search for uses of each nonlocal moniker. We stop processing after the first
 * moniker for which we received results. As we process monikers in an order that considers moniker
 * schemes, the first one to get results should be the most desirable.

 * If there are any locations in the result set, this method returns the new cursor. This method
 * returns undefined if there are no remaining results for the same repository.
 *
 * @param repositoryId The repository identifier.
 * @param commit The target commit.
 * @param limit The maximum number of dumps to open.
 * @param cursor The pagination cursor.
 * @param ctx The tracing context.
 */
    private async performSameDumpMonikerReferences(
        cursor: SameDumpReferenceCursor,
        ctx: TracingContext = {}
    ): Promise<{ locations: InternalLocation[]; newCursor?: ReferencePaginationCursor }> {
        const document = await this.getDocumentByPath(cursor.dumpId, cursor.path, ctx)
        if (!document) {
            return { locations: [] }
        }

        let locations: InternalLocation[] = []

        for (const moniker of cursor.monikers) {
            if (moniker.kind !== 'import') {
                continue
            }

            // Get locations in the defining package
            const monikerLocations = await this.lookupMoniker(
                document,
                moniker,
                sqliteModels.ReferenceModel,
                {}, // TODO - paginate
                ctx
            )
            locations = locations.concat(monikerLocations)
        }

        return { locations } // TODO - cursor
    }

    /**
     * Find the references of the target moniker outside of the current dump but within a dump of
     * the same repository. If the moniker has attached package information, then the dependency
     * database is queried for the packages that require this particular moniker identifier. These
     * dumps are opened, and their references tables are queried for the target moniker. If there
     * are any locations in the result set, this method returns the new cursor. This method returns
     * undefined if there are no remaining results for the same repository.
     *
     * @param repositoryId The repository identifier.
     * @param commit The target commit.
     * @param limit The maximum number of dumps to open.
     * @param cursor The pagination cursor.
     * @param ctx The tracing context.
     */
    private async performSameRepositoryRemoteReferences(
        repositoryId: number,
        commit: string,
        limit: number,
        cursor: RemoteDumpReferenceCursor,
        ctx: TracingContext = {}
    ): Promise<{ locations: InternalLocation[]; newCursor?: ReferencePaginationCursor }> {
        const { references, totalCount, newOffset } = await this.dependencyManager.getSameRepoRemoteReferences({
            ...cursor,
            repositoryId,
            commit,
            limit,
            ctx,
        })

        const locations = await this.locationsFromRemoteReferences(
            cursor.dumpId,
            { scheme: cursor.scheme, identifier: cursor.identifier },
            references.map(r => r.dump),
            ctx
        )

        return {
            locations,
            newCursor: newOffset < totalCount ? { ...cursor, phase: 'same-repo', offset: newOffset } : undefined,
        }
    }

    /**
     * Find the references of the target moniker outside of the current repository. If the moniker
     * has attached package information, then Postgres is queried for the packages that require
     * this particular moniker identifier. These dumps are opened, and their references tables are
     * queried for the target moniker. If there are any locations in the result set, this method
     * returns the new cursor. This method returns undefined if there are no remaining results for
     * the same repository.
     *
     * @param repositoryId The repository identifier.
     * @param limit The maximum number of dumps to open.
     * @param cursor The pagination cursor.
     * @param ctx The tracing context.
     */
    private async performRemoteReferences(
        repositoryId: number,
        limit: number,
        cursor: RemoteDumpReferenceCursor,
        ctx: TracingContext = {}
    ): Promise<{ locations: InternalLocation[]; newCursor?: ReferencePaginationCursor }> {
        const { references, totalCount, newOffset } = await this.dependencyManager.getReferences({
            ...cursor,
            repositoryId,
            limit,
            ctx,
        })

        const locations = await this.locationsFromRemoteReferences(
            cursor.dumpId,
            { scheme: cursor.scheme, identifier: cursor.identifier },
            references.map(r => r.dump),
            ctx
        )

        return {
            locations,
            newCursor: newOffset < totalCount ? { ...cursor, phase: 'remote-repo', offset: newOffset } : undefined,
        }
    }

    /**
     * Determine if the moniker and package identified by the pagination cursor has at least one
     * remote repository. containing that definition. We use this to determine if we should move
     * on to the next phase without doing it unconditionally and yielding an empty last page.
     *
     * @param repositoryId The repository identifier.
     * @param cursor The pagination cursor.
     * @param ctx The tracing context.
     */
    private async hasRemoteReferences(
        repositoryId: number,
        cursor: RemoteDumpReferenceCursor,
        ctx: TracingContext = {}
    ): Promise<boolean> {
        const { totalCount: remoteTotalCount } = await this.dependencyManager.getReferences({
            ...cursor,
            repositoryId,
            limit: 1,
            offset: 0,
            ctx,
        })

        return remoteTotalCount > 0
    }

    /**
     * Query the given dumps for references to the given moniker.
     *
     * @param dumpId The ID of the dump for which this database answers queries.
     * @param moniker The target moniker.
     * @param dumps The dumps to open.
     * @param ctx The tracing context.
     */
    private async locationsFromRemoteReferences(
        dumpId: pgModels.DumpId,
        moniker: Pick<sqliteModels.MonikerData, 'scheme' | 'identifier'>,
        dumps: pgModels.LsifDump[],
        ctx: TracingContext = {}
    ): Promise<InternalLocation[]> {
        logSpan(ctx, 'package_references', {
            references: dumps.map(d => ({ repositoryId: d.repositoryId, commit: d.commit })),
        })

        let locations: InternalLocation[] = []
        for (const dump of dumps) {
            // Skip the remote reference that show up for ourselves - we've already gathered
            // these in the previous step of the references query.
            if (dump.id === dumpId) {
                continue
            }

            const { locations: monikerResults } = await this.createDatabase(dump).monikerResults(
                sqliteModels.ReferenceModel,
                moniker,
                {}, // TODO - paginate
                ctx
            )
            const references = monikerResults.map(loc => locationFromDatabase(dump.root, loc))
            locations = locations.concat(references)
        }

        return locations
    }

    /**
     * Find the locations attached to the target moniker outside of the current database. If
     * the moniker has attached package information, then Postgres is queried for the target
     * package. That database is opened, and its definitions table is queried for the target
     * moniker.
     *
     * @param document The document containing the definition.
     * @param moniker The target moniker.
     * @param model The target model.
     * @param pagination A limit and offset to use for the query.
     * @param ctx The tracing context.
     */
    private async lookupMoniker(
        document: sqliteModels.DocumentData,
        moniker: sqliteModels.MonikerData,
        model: typeof sqliteModels.DefinitionModel | typeof sqliteModels.ReferenceModel,
        pagination: { skip?: number; take?: number },
        ctx: TracingContext = {}
    ): Promise<InternalLocation[]> {
        const packageInformation = this.lookupPackageInformation(document, moniker, ctx)
        if (!packageInformation) {
            return []
        }

        const packageEntity = await this.dependencyManager.getPackage(
            moniker.scheme,
            packageInformation.name,
            packageInformation.version
        )
        if (!packageEntity) {
            return []
        }

        logSpan(ctx, 'package_entity', {
            moniker,
            packageInformation,
            packageRepositoryId: packageEntity.dump.repositoryId,
            packageCommit: packageEntity.dump.commit,
        })

        const { locations: monikerResults } = await this.createDatabase(packageEntity.dump).monikerResults(
            model,
            moniker,
            pagination,
            ctx
        )
        return monikerResults.map(loc => locationFromDatabase(packageEntity.dump.root, loc))
    }

    /**
     * Retrieve the package information from associated with the given moniker.
     *
     * @param document The document containing an instance of the moniker.
     * @param moniker The target moniker.
     * @param ctx The tracing context.
     */
    private lookupPackageInformation(
        document: sqliteModels.DocumentData,
        moniker: sqliteModels.MonikerData,
        ctx: TracingContext = {}
    ): sqliteModels.PackageInformationData | undefined {
        if (!moniker.packageInformationId) {
            return undefined
        }

        const packageInformation = document.packageInformation.get(moniker.packageInformationId)
        if (!packageInformation) {
            return undefined
        }

        logSpan(ctx, 'package_information', {
            moniker,
            packageInformation,
        })

        return packageInformation
    }

    /**
     * Create a database instance for the dump identifier. This identifier should have ben retrieved
     * from a call to the `exists` route, which would have this identifier from `findClosestDatabase`.
     * Also returns the dump instance backing the database. Returns an undefined database and dump if
     * no such dump can be found. Will also return a tracing context tagged with the closest commit
     * found. This new tracing context should be used in all downstream requests so that the original
     * commit and the effective commit are both known.
     *
     * If no dumpId is supplied, the first database from `findClosestDatabase` is used. Note that this
     * functionality does not happen in the application and only in tests, as an uploadId is a required
     * parameter on all routes into the API.
     *
     * @param repositoryId The repository identifier.
     * @param commit The target commit.
     * @param path One of the files in the dump.
     * @param dumpId The identifier of the dump to load.
     * @param ctx The tracing context.
     */
    private async closestDatabase(
        repositoryId: number,
        commit: string,
        path: string,
        dumpId?: number,
        ctx: TracingContext = {}
    ): Promise<{ database: Database; dump: pgModels.LsifDump; ctx: TracingContext } | undefined> {
        if (!dumpId) {
            const databases = await this.findClosestDatabases(repositoryId, commit, path)
            return databases.length > 0 ? databases[0] : undefined
        }

        const dumpAndDatabase = await this.getDumpAndDatabaseById(dumpId)
        if (!dumpAndDatabase) {
            return undefined
        }
        const { dump, database } = dumpAndDatabase

        return { dump, database, ctx: addTags(ctx, { closestCommit: dump.commit }) }
    }

    /**
     * Create a set of database instances for the given repository at the closest commits to the
     * target commit. This method returns only databases that contain the given file. Also returns
     * the dump instance backing the database. Returns an undefined database and dump if no such
     * dump can be found. Will also return a tracing context tagged with the closest commit found.
     * This new tracing context should be used in all downstream requests so that the original
     * commit and the effective commit are both known.
     *
     * This method returns databases ordered by commit distance (nearest first).
     *
     * @param repositoryId The repository identifier.
     * @param commit The target commit.
     * @param path One of the files in the dump.
     * @param ctx The tracing context
     */
    private async findClosestDatabases(
        repositoryId: number,
        commit: string,
        path: string,
        ctx: TracingContext = {}
    ): Promise<{ database: Database; dump: pgModels.LsifDump; ctx: TracingContext }[]> {
        // Find all closest dumps. Each database is guaranteed to have a root that is a
        // prefix of the given path, but does not guarantee that the path actually exists
        // in that dump.

        const closestDumps = await this.dumpManager.findClosestDumps(repositoryId, commit, path, ctx, this.frontendUrl)

        // Concurrently ensure that each database contains the target file. If it does
        // not contain data for that file, return undefined and filter it from the list
        // before returning.

        return (
            await Promise.all(
                closestDumps.map(async dump => {
                    const database = this.createDatabase(dump)
                    const taggedCtx = addTags(ctx, { closestCommit: dump.commit })

                    return (await database.exists(pathToDatabase(dump.root, path), taggedCtx))
                        ? { database, dump, ctx: taggedCtx }
                        : undefined
                })
            )
        ).filter(isDefined)
    }

    /**
     * Create a database instance backed by the given dump.
     *
     * @param dump The dump.
     */
    private createDatabase(dump: pgModels.LsifDump): Database {
        return new Database(
            this.connectionCache,
            this.documentCache,
            this.resultChunkCache,
            dump,
            dbFilename(this.storageRoot, dump.id)
        )
    }

    /**
     * Create a database for the dump with the given identifier.
     *
     * @param dumpId The dump id.
     * @param ctx The tracing context.
     */
    private async getDumpAndDatabaseById(
        dumpId: number
    ): Promise<{ dump: pgModels.LsifDump; database: Database } | undefined> {
        const dump = await this.dumpManager.getDumpById(dumpId)
        if (!dump) {
            return undefined
        }

        return { dump, database: this.createDatabase(dump) }
    }

    /**
     * Create a database for the dump with the given identifier and return the document
     * with the given path.
     *
     * @param dumpId The dump id.
     * @param path The document path.
     * @param ctx The tracing context.
     */
    private async getDocumentByPath(
        dumpId: number,
        path: string,
        ctx: TracingContext = {}
    ): Promise<sqliteModels.DocumentData | undefined> {
        const dumpAndDatabase = await this.getDumpAndDatabaseById(dumpId)
        if (!dumpAndDatabase) {
            return undefined
        }
        const { database } = dumpAndDatabase

        return database.getDocumentByPath(path, ctx)
    }
}

/**
 * Converts a file in the repository to the corresponding file in the
 * database.
 *
 * @param root The root of all files in the dump.
 * @param path The path within the dump.
 */
function pathToDatabase(root: string, path: string): string {
    return path.startsWith(root) ? path.slice(root.length) : path
}

/**
 * Converts a location in a dump to the corresponding location in the repository.
 *
 * @param root The root of all files in the dump.
 * @param location The original location.
 */
function locationFromDatabase(root: string, { dump, path, range }: InternalLocation): InternalLocation {
    return {
        dump,
        path: `${root}${path}`,
        range,
    }
}
