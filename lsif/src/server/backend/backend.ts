import * as sqliteModels from '../../shared/models/sqlite'
import * as lsp from 'vscode-languageserver-protocol'
import * as pgModels from '../../shared/models/pg'
import { addTags, logSpan, TracingContext } from '../../shared/tracing'
import { Database } from './database'
import { dbFilename } from '../../shared/paths'
import { mustGet } from '../../shared/maps'
import { DumpManager } from '../../shared/store/dumps'
import { DEFAULT_REFERENCES_REMOTE_DUMP_LIMIT } from '../../shared/constants'
import { DependencyManager } from '../../shared/store/dependencies'
import { isDefined } from '../../shared/util'
import {
    DefinitionMonikersReferenceCursor,
    ReferencePaginationContext,
    ReferencePaginationCursor,
    RemoteDumpReferenceCursor,
    SameDumpReferenceCursor,
} from './cursor'
import { InternalLocation, ResolvedInternalLocation } from './location'
import { isEqual, uniqWith } from 'lodash'

interface PaginatedInternalLocations {
    locations: ResolvedInternalLocation[]
    newCursor?: ReferencePaginationCursor
}

/**
 * A wrapper around code intelligence operations. This class deals with logic that spans
 * multiple repositories or commits. For single-dump logic, see the `Database` class.
 */
export class Backend {
    /**
     * Create a new `Backend`.
     *
     * @param storageRoot The path where SQLite databases are stored.
     * @param dumpManager The dumps manager instance.
     * @param dependencyManager The dependency manager instance.
     * @param frontendUrl The url of the frontend internal API.
     * @param createDatabase Function used to create a database instance from a dump.
     */
    constructor(
        private storageRoot: string,
        private dumpManager: DumpManager,
        private dependencyManager: DependencyManager,
        private frontendUrl: string,
        private createDatabase: (dump: pgModels.LsifDump) => Database = dump =>
            new Database(dump.id, dbFilename(this.storageRoot, dump.id))
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
    ): Promise<ResolvedInternalLocation[] | undefined> {
        const closestDumpAndDatabase = await this.closestDatabase(repositoryId, commit, path, dumpId, ctx)
        if (!closestDumpAndDatabase) {
            if (ctx.logger) {
                ctx.logger.warn('No database could be loaded', { repositoryId, commit, path })
            }

            return undefined
        }
        const { dump, database, ctx: newCtx } = closestDumpAndDatabase

        // Construct path within dump
        const pathInDb = pathToDatabase(dump.root, path)

        // Try to find definitions in the same dump
        const dbDefinitions = await database.definitions(pathInDb, position, newCtx)
        const definitions = dbDefinitions.map(loc => locationFromDatabase(dump.root, loc))
        if (definitions.length > 0) {
            return this.resolveLocations(definitions)
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
                    // a remote definition for it.

                    const { locations: remoteDefinitions } = await this.lookupMoniker(
                        document,
                        moniker,
                        sqliteModels.DefinitionModel,
                        {},
                        ctx
                    )
                    if (remoteDefinitions.length > 0) {
                        return this.resolveLocations(remoteDefinitions)
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
                        return this.resolveLocations(localDefinitions)
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
     * @param remoteDumpLimit The maximum number of remote dumps to query in one operation.
     * @param dumpId The identifier of the dump to load. If not supplied, the closest dump will be used.
     * @param ctx The tracing context.
     */
    public async references(
        repositoryId: number,
        commit: string,
        path: string,
        position: lsp.Position,
        paginationContext: ReferencePaginationContext = { limit: 10 },
        remoteDumpLimit = DEFAULT_REFERENCES_REMOTE_DUMP_LIMIT,
        dumpId?: number,
        ctx: TracingContext = {}
    ): Promise<PaginatedInternalLocations | undefined> {
        if (paginationContext.cursor) {
            return this.handleReferencePaginationCursor(
                repositoryId,
                commit,
                remoteDumpLimit,
                paginationContext.limit,
                paginationContext.cursor,
                ctx
            )
        }

        const closestDumpAndDatabase = await this.closestDatabase(repositoryId, commit, path, dumpId, ctx)
        if (!closestDumpAndDatabase) {
            if (ctx.logger) {
                ctx.logger.warn('No database could be loaded', { repositoryId, commit, path })
            }

            return undefined
        }
        const { dump, database, ctx: newCtx } = closestDumpAndDatabase

        // Construct path within dump
        const pathInDb = pathToDatabase(dump.root, path)

        // Get the ranges of for this position and the document in which they occur
        const { document, ranges } = await database.getRangeByPosition(pathInDb, position, ctx)

        // Find and normalize the monikers attached to each range
        const monikers: sqliteModels.MonikerData[] = []
        if (document) {
            for (const range of ranges) {
                for (const monikerId of range.monikerIds) {
                    monikers.push(mustGet(document.monikers, monikerId, 'monikers'))
                }
            }
        }

        const cursor: ReferencePaginationCursor = {
            phase: 'same-dump',
            dumpId: dump.id,
            path: pathInDb,
            position,
            monikers: sortMonikers(monikers),
            skipResults: 0,
        }

        // Request the first page of results
        return this.handleReferencePaginationCursor(
            repositoryId,
            commit,
            remoteDumpLimit,
            paginationContext.limit,
            cursor,
            newCtx
        )
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
        const closestDumpAndDatabase = await this.closestDatabase(repositoryId, commit, path, dumpId, ctx)
        if (!closestDumpAndDatabase) {
            if (ctx.logger) {
                ctx.logger.warn('No database could be loaded', { repositoryId, commit, path })
            }

            return undefined
        }
        const { dump, database, ctx: newCtx } = closestDumpAndDatabase

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
     *   (1) 'same-dump': query the original dump's LSIF reference results and references table
     *   (2) 'definition-monikers': query the monikers in the dump that defines them
     *   (3) 'same-repo': open additional dumps that belong to the same repository
     *   (4) 'remote-repo': open additional dumps that belong to a different repository
     *
     * This method will return any locations found in this page of results as well as a cursor
     * indicating how to execute the next page of results. If the cursor is undefined there are no
     * more results.
     *
     * @param repositoryId The repository identifier.
     * @param commit The target commit.
     * @param remoteDumpLimit The maximum number of remote dumps to query in one operation.
     * @param limit The maximum number of locations to return on this page.
     * @param cursor The pagination cursor.
     * @param ctx The tracing context.
     */
    private async handleReferencePaginationCursor(
        repositoryId: number,
        commit: string,
        remoteDumpLimit: number,
        limit: number,
        cursor: ReferencePaginationCursor,
        ctx: TracingContext = {}
    ): Promise<PaginatedInternalLocations> {
        /**
         * This method takes a handler that executes the current page of results and returns a new
         * cursor for the **same phase** of results. If there are no more results in that phase of
         * the result set, the cursor is undefined. In this case, we call the `makeCursor` factory
         * to construct the cursor for the next phase of pagination. When no further data are
         * available in any phase, the factory returns undefined.
         *
         * If the locations from the handler function do not produce a full page of results, the
         * next page of results are evaluated with a modified limit.
         *
         * @param handler The handler for the current page of results.
         * @param makeCursor A factory that creates a cursor for the next phase of pagination.
         */
        const recur = async (
            handler: () => Promise<PaginatedInternalLocations>,
            makeCursor: () => Promise<ReferencePaginationCursor | undefined> | ReferencePaginationCursor | undefined
        ): Promise<PaginatedInternalLocations> => {
            const { locations, newCursor: originalCursor } = await handler()
            const newCursor = originalCursor || (await makeCursor())
            if (!newCursor) {
                return { locations }
            }

            limit -= locations.length
            if (limit <= 0) {
                return { locations, newCursor }
            }

            const {
                locations: nextPageLocations,
                newCursor: nextPageNewCursor,
            } = await this.handleReferencePaginationCursor(repositoryId, commit, remoteDumpLimit, limit, newCursor, ctx)

            return { locations: locations.concat(nextPageLocations), newCursor: nextPageNewCursor }
        }

        switch (cursor.phase) {
            case 'same-dump': {
                return recur(
                    () => this.performSameDumpReferences(limit, cursor, ctx),
                    () => ({
                        dumpId: cursor.dumpId,
                        phase: 'definition-monikers',
                        path: cursor.path,
                        position: cursor.position,
                        monikers: cursor.monikers,
                        skipResults: 0,
                    })
                )
            }

            case 'definition-monikers': {
                return recur(
                    () => this.performDefinitionMonikersReferences(limit, cursor, ctx),
                    async (): Promise<ReferencePaginationCursor | undefined> => {
                        const document = await this.getDocumentByPath(cursor.dumpId, cursor.path, ctx)
                        if (!document) {
                            return undefined
                        }

                        for (const moniker of cursor.monikers) {
                            const packageInformation = this.lookupPackageInformation(document, moniker, ctx)
                            if (packageInformation) {
                                return {
                                    dumpId: cursor.dumpId,
                                    phase: 'same-repo',
                                    scheme: moniker.scheme,
                                    identifier: moniker.identifier,
                                    name: packageInformation.name,
                                    version: packageInformation.version,
                                    dumpIds: [],
                                    totalDumpsWhenBatching: 0,
                                    skipDumpsWhenBatching: 0,
                                    skipDumpsInBatch: 0,
                                    skipResultsInDump: 0,
                                }
                            }
                        }
                        return undefined
                    }
                )
            }

            case 'same-repo': {
                return recur(
                    () =>
                        this.performSameRepositoryRemoteReferences(
                            repositoryId,
                            commit,
                            remoteDumpLimit,
                            limit,
                            cursor,
                            ctx
                        ),
                    (): ReferencePaginationCursor | undefined => ({
                        dumpId: cursor.dumpId,
                        phase: 'remote-repo',
                        scheme: cursor.scheme,
                        identifier: cursor.identifier,
                        name: cursor.name,
                        version: cursor.version,
                        dumpIds: [],
                        totalDumpsWhenBatching: 0,
                        skipDumpsWhenBatching: 0,
                        skipDumpsInBatch: 0,
                        skipResultsInDump: 0,
                    })
                )
            }

            case 'remote-repo': {
                return recur(
                    () => this.performRemoteReferences(repositoryId, remoteDumpLimit, limit, cursor, ctx),
                    () => undefined
                )
            }
        }
    }

    /**
     * Search the LSIF reference results and the references table of the current dump. This method
     * returns a cursor if there are reference results locations remaining for a subsequent page.
     *
     * Implementation detail: this method brings all of the LSIF and references table results into
     * memory so that they can be reasonable deduplicated. We splice the combined results by index
     * and return pages in this manner rather than pulling back a set of results from the SQLite
     * database. This step turns out to be quite necessary, as a single LSIF dump can have a fair
     * amount of reference duplication when looking at both result set edges and attached monikers.
     * Skipping this step may return several pages of duplicated results if not batched in this
     * manner.
     *
     * @param limit The maximum number of locations to return on this page.
     * @param cursor The pagination cursor.
     * @param ctx The tracing context.
     */
    private async performSameDumpReferences(
        limit: number,
        cursor: SameDumpReferenceCursor,
        ctx: TracingContext = {}
    ): Promise<PaginatedInternalLocations> {
        const dumpAndDatabase = await this.getDumpAndDatabaseById(cursor.dumpId)
        if (!dumpAndDatabase) {
            return { locations: [] }
        }
        const { dump, database } = dumpAndDatabase

        // First get all LSIF reference result locations for the given position.
        const locationSet = await database.references(cursor.path, cursor.position, ctx)

        // Search the references table of the current dump. This search is necessary because
        // we want a 'Find References' operation on a reference to also return references to
        // the governing definition, and those may not be fully linked in the LSIF data. This
        // method returns a cursor if there are reference rows remaining for a subsequent page.
        for (const moniker of cursor.monikers) {
            const { locations: monikerLocations } = await database.monikerResults(
                sqliteModels.ReferenceModel,
                moniker,
                {},
                ctx
            )

            for (const location of monikerLocations) {
                locationSet.push(location)
            }
        }

        // Get the page's slice of results
        const slicedLocations = locationSet.locations.slice(cursor.skipResults, cursor.skipResults + limit)
        const newOffset = cursor.skipResults + limit
        const newCursor = { ...cursor, skipResults: cursor.skipResults + limit }

        return {
            locations: await this.resolveLocations(slicedLocations.map(loc => locationFromDatabase(dump.root, loc))),
            newCursor: newOffset < locationSet.locations.length ? newCursor : undefined,
        }
    }

    /**
     *
     * Perform a search for uses of each nonlocal moniker in the dump where they are defined. We stop
     * processing after the first moniker for which we received results. As we process monikers in an
     * order that considers moniker schemes, the first one to get results should be the most desirable.
     * This method returns a cursor if there are additional monikers to process on a subsequent page.
     *
     * @param limit The maximum number of locations to return on this page.
     * @param cursor The pagination cursor.
     * @param ctx The tracing context.
     */
    private async performDefinitionMonikersReferences(
        limit: number,
        cursor: DefinitionMonikersReferenceCursor,
        ctx: TracingContext = {}
    ): Promise<PaginatedInternalLocations> {
        const document = await this.getDocumentByPath(cursor.dumpId, cursor.path, ctx)
        if (!document) {
            return { locations: [] }
        }

        for (const moniker of cursor.monikers) {
            if (moniker.kind !== 'import') {
                continue
            }

            // Get locations in the defining package
            const { locations, count } = await this.lookupMoniker(
                document,
                moniker,
                sqliteModels.ReferenceModel,
                { take: limit, skip: cursor.skipResults },
                ctx
            )

            if (locations.length > 0) {
                const newOffset = cursor.skipResults + locations.length
                const newCursor = { ...cursor, skipResults: cursor.skipResults + limit }

                return {
                    locations: await this.resolveLocations(locations),
                    newCursor: newOffset < count ? newCursor : undefined,
                }
            }
        }

        return { locations: [] }
    }

    /**
     * Find the references of the target moniker outside of the current dump but within a dump of
     * the same repository. If the moniker has attached package information, then the dependency
     * database is queried for the packages that require this particular moniker identifier. These
     * dumps are opened, and their references tables are queried for the target moniker. This method
     * returns a cursor if there are additional dumps to process on a subsequent page.
     *
     * @param repositoryId The repository identifier.
     * @param commit The target commit.
     * @param remoteDumpLimit The maximum number of remote dumps to query in one operation.
     * @param limit The maximum number of locations to return on this page.
     * @param cursor The pagination cursor.
     * @param ctx The tracing context.
     */
    private async performSameRepositoryRemoteReferences(
        repositoryId: number,
        commit: string,
        remoteDumpLimit: number,
        limit: number,
        cursor: RemoteDumpReferenceCursor,
        ctx: TracingContext = {}
    ): Promise<PaginatedInternalLocations> {
        const getPackageReferences = (): ReturnType<DependencyManager['getSameRepoRemotePackageReferences']> =>
            this.dependencyManager.getSameRepoRemotePackageReferences({
                repositoryId,
                commit,
                scheme: cursor.scheme,
                name: cursor.name,
                version: cursor.version,
                identifier: cursor.identifier,
                limit: remoteDumpLimit,
                offset: cursor.skipDumpsWhenBatching,
                ctx,
            })

        return this.locationsFromRemoteReferences({
            dumpId: cursor.dumpId,
            moniker: { scheme: cursor.scheme, identifier: cursor.identifier },
            getPackageReferences,
            limit,
            cursor,
            ctx,
        })
    }

    /**
     * Find the references of the target moniker outside of the current repository. If the moniker
     * has attached package information, then Postgres is queried for the packages that require
     * this particular moniker identifier. These dumps are opened, and their references tables are
     * queried for the target moniker. This method returns a cursor if there are additional dumps
     * to process on a subsequent page.
     *
     * @param repositoryId The repository identifier.
     * @param remoteDumpLimit The maximum number of remote dumps to query in one operation.
     * @param limit The maximum number of locations to return on this page.
     * @param cursor The pagination cursor.
     * @param ctx The tracing context.
     */
    private async performRemoteReferences(
        repositoryId: number,
        remoteDumpLimit: number,
        limit: number,
        cursor: RemoteDumpReferenceCursor,
        ctx: TracingContext = {}
    ): Promise<PaginatedInternalLocations> {
        const getPackageReferences = (): ReturnType<DependencyManager['getPackageReferences']> =>
            this.dependencyManager.getPackageReferences({
                repositoryId,
                scheme: cursor.scheme,
                name: cursor.name,
                version: cursor.version,
                identifier: cursor.identifier,
                limit: remoteDumpLimit,
                offset: cursor.skipDumpsWhenBatching,
                ctx,
            })

        return this.locationsFromRemoteReferences({
            dumpId: cursor.dumpId,
            moniker: { scheme: cursor.scheme, identifier: cursor.identifier },
            getPackageReferences,
            limit,
            cursor,
            ctx,
        })
    }

    /**
     * Query the given dumps for references to the given moniker.
     *
     * @param args Parameter bag.
     */
    private async locationsFromRemoteReferences({
        dumpId,
        moniker,
        getPackageReferences,
        limit,
        cursor,
        ctx = {},
    }: {
        /** The ID of the dump for which this database answers queries. */
        dumpId: pgModels.DumpId
        /** The target moniker. */
        moniker: Pick<sqliteModels.MonikerData, 'scheme' | 'identifier'>
        /** A function that retrieves the next batch of references. */
        getPackageReferences: () => Promise<{
            packageReferences: pgModels.ReferenceModel[]
            newOffset: number
            totalCount: number
        }>
        /** The maximum number of locations to return on this page. */
        limit: number
        /** The pagination cursor. */
        cursor: RemoteDumpReferenceCursor
        /** The tracing context. */
        ctx: TracingContext
    }): Promise<PaginatedInternalLocations> {
        if (cursor.dumpIds.length === 0) {
            const { packageReferences, newOffset, totalCount } = await getPackageReferences()

            logSpan(ctx, 'package_references', {
                references: packageReferences.map(r => ({ repositoryId: r.dump.repositoryId, commit: r.dump.commit })),
            })

            cursor.dumpIds = packageReferences.map(r => r.dump.id)
            cursor.skipDumpsWhenBatching = newOffset
            cursor.totalDumpsWhenBatching = totalCount
        }

        for (const [i, batchDumpId] of cursor.dumpIds.entries()) {
            if (i < cursor.skipDumpsInBatch) {
                continue
            }

            // Skip the remote reference that show up for ourselves - we've already gathered
            // these in the previous step of the references query.
            if (batchDumpId === dumpId) {
                continue
            }

            const dumpAndDatabase = await this.getDumpAndDatabaseById(batchDumpId)
            if (!dumpAndDatabase) {
                continue
            }
            const { dump, database } = dumpAndDatabase

            const { locations, count } = await database.monikerResults(
                sqliteModels.ReferenceModel,
                moniker,
                { take: limit, skip: cursor.skipResultsInDump },
                ctx
            )

            if (locations.length > 0) {
                const newResultOffset = cursor.skipResultsInDump + locations.length
                const moreDumps = i + 1 < cursor.dumpIds.length
                const nextCursor = { ...cursor, skipResultsInDump: cursor.skipResultsInDump + limit }
                const nextDumpCursor = { ...cursor, skipDumpsInBatch: i + 1, skipResultsInDump: 0 }
                const nextBatchCursor = {
                    ...cursor,
                    dumpIds: [],
                    skipDumpsInBatch: 0,
                    skipResultsInDump: 0,
                }

                return {
                    locations: await this.resolveLocations(locations.map(loc => locationFromDatabase(dump.root, loc))),
                    newCursor:
                        newResultOffset < count
                            ? nextCursor
                            : moreDumps
                            ? nextDumpCursor
                            : cursor.skipDumpsWhenBatching < cursor.totalDumpsWhenBatching
                            ? nextBatchCursor
                            : undefined,
                }
            }
        }

        return { locations: [] }
    }

    /**
     * Find the locations attached to the target moniker in the dump where it is defined. If
     * the moniker has attached package information, then query Postgres for the target
     * package. Open that package's database and query its definitions or references table
     * for the target moniker (depending on the given model).
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
    ): Promise<{ locations: InternalLocation[]; count: number }> {
        const packageInformation = this.lookupPackageInformation(document, moniker, ctx)
        if (!packageInformation) {
            return { locations: [], count: 0 }
        }

        const packageEntity = await this.dependencyManager.getPackage(
            moniker.scheme,
            packageInformation.name,
            packageInformation.version
        )
        if (!packageEntity) {
            return { locations: [], count: 0 }
        }

        logSpan(ctx, 'package_entity', {
            moniker,
            packageInformation,
            packageRepositoryId: packageEntity.dump.repositoryId,
            packageCommit: packageEntity.dump.commit,
        })

        const { locations, count } = await this.createDatabase(packageEntity.dump).monikerResults(
            model,
            moniker,
            pagination,
            ctx
        )
        return { locations: locations.map(loc => locationFromDatabase(packageEntity.dump.root, loc)), count }
    }

    /**
     * Retrieve the package information associated with the given moniker.
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
     * TODO - remove test-specific logic
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
    ): Promise<{ dump: pgModels.LsifDump; database: Database; ctx: TracingContext } | undefined> {
        if (!dumpId) {
            const databases = await this.findClosestDatabases(repositoryId, commit, path)
            return databases.length > 0 ? databases[0] : undefined
        }

        const dumpAndDatabase = await this.getDumpAndDatabaseById(dumpId)
        if (!dumpAndDatabase) {
            return undefined
        }

        return { ...dumpAndDatabase, ctx: addTags(ctx, { closestCommit: dumpAndDatabase.dump.commit }) }
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
    ): Promise<{ dump: pgModels.LsifDump; database: Database; ctx: TracingContext }[]> {
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
     * Create a database for the dump with the given identifier.
     *
     * @param dumpId The dump id.
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
        return dumpAndDatabase?.database.getDocumentByPath(path, ctx)
    }

    /** Bulk populate the dump model for internal locations. */
    private async resolveLocations(locations: InternalLocation[]): Promise<ResolvedInternalLocation[]> {
        const dumps = await this.dumpManager.getDumpsByIds(Array.from(new Set(locations.map(({ dumpId }) => dumpId))))

        const resolvedLocations: ResolvedInternalLocation[] = []
        for (const { dumpId, path, range } of locations) {
            const dump = dumps.get(dumpId)
            if (!dump) {
                continue
            }

            resolvedLocations.push({ dump, path, range })
        }

        return resolvedLocations
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
 * Converts a location in a dump to the corresponding location in the repository.2
 *
 * @param root The root of all files in the dump.
 * @param location The original location.
 */
function locationFromDatabase(root: string, { dumpId, path, range }: InternalLocation): InternalLocation {
    return {
        dumpId,
        path: `${root}${path}`,
        range,
    }
}

// The order to present monikers in when organized by kinds
const monikerKindPreferences = ['import', 'local', 'export']

// A map from moniker schemes to schemes that subsume them. The schemes
// identified by keys should be removed from the sets of monikers that
// also contain the scheme identified by that key's value.
const subsumedMonikers = new Map([
    ['go', 'gomod'],
    ['tsc', 'npm'],
])

/**
 * Normalize the set of monikers by filtering, sorting, and removing
 * duplicates from the list based on the moniker kind and scheme values.
 *
 * @param monikers The list of monikers.
 */
export function sortMonikers(monikers: sqliteModels.MonikerData[]): sqliteModels.MonikerData[] {
    // Deduplicate monikers. This can happen with long chains of result
    // sets where monikers are applied several times to an aliased symbol.
    monikers = uniqWith(monikers, isEqual)

    // Remove monikers subsumed by the presence of another. For example,
    // if we have an `npm` moniker in this list, we want to remove all
    // `tsc` monikers as they are duplicate by construction in lsif-tsc.
    monikers = monikers.filter(a => {
        const by = subsumedMonikers.get(a.scheme)
        return !(by && monikers.some(b => b.scheme === by))
    })

    // Sort monikers by kind
    monikers.sort((a, b) => monikerKindPreferences.indexOf(a.kind) - monikerKindPreferences.indexOf(b.kind))
    return monikers
}
