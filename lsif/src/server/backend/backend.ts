import * as dumpModels from '../../shared/models/dump'
import * as lsp from 'vscode-languageserver-protocol'
import * as settings from '../settings'
import * as xrepoModels from '../../shared/models/xrepo'
import { addTags, logSpan, TracingContext } from '../../shared/tracing'
import { ConnectionCache, DocumentCache, ResultChunkCache } from './cache'
import { Database, sortMonikers, InternalLocation } from './database'
import { dbFilename } from '../../shared/paths'
import { isEqual, uniqWith } from 'lodash'
import { mustGet } from '../../shared/maps'
import { XrepoDatabase } from '../../shared/xrepo/xrepo'

/**
 * Context describing the current request for paginated results.
 */
export interface ReferencePaginationContext {
    /**
     * The maximum number of remote dumps to search.
     */
    limit: number

    /**
     * Context describing the previous page of results.
     */
    cursor?: ReferencePaginationCursor
}

/**
 * Reference pagination happens in two distinct phases:
 *
 *   (1) open a slice of dumps for the same repositories, and
 *   (2) open a slice of dumps for other repositories.
 */
export type ReferencePaginationPhase = 'same-repo' | 'remote-repo'

/**
 * Context describing the previous page of results.
 */
export interface ReferencePaginationCursor {
    /**
     * The identifier of the dump that contains the target range.
     */
    dumpId: number

    /**
     * The scheme of the moniker that has remote results.
     */
    scheme: string

    /**
     * The identifier of the moniker that has remote results.
     */
    identifier: string

    /**
     * The name of the package that has remote results.
     */
    name: string

    /**
     * The version of the package that has remote results.
     */
    version: string | null

    /**
     * The phase of the pagination.
     */
    phase: ReferencePaginationPhase

    /**
     * The number of remote dumps to skip.
     */
    offset: number
}

/**
 * Converts a file in the repository to the corresponding file in the
 * database.
 *
 * @param root The root of the dump.
 * @param path The path within the dump.
 */
const pathToDatabase = (root: string, path: string): string => (path.startsWith(root) ? path.slice(root.length) : path)

/**
 * Converts a location in a dump to the corresponding location in the repository.
 *
 * @param root The root of the dump.
 * @param location The original location.
 */
const locationFromDatabase = (root: string, { dump, path, range }: InternalLocation): InternalLocation => ({
    dump,
    path: `${root}${path}`,
    range,
})

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
     * @param xrepoDatabase The cross-repo database.
     * @param fetchConfiguration A function that returns the current configuration.
     */
    constructor(
        private storageRoot: string,
        private xrepoDatabase: XrepoDatabase,
        private fetchConfiguration: () => { gitServers: string[] }
    ) {}

    /**
     * Get the set of dumps for a repository.
     *
     * @param repository The repository.
     * @param query A search query.
     * @param visibleAtTip If true, only return dumps visible at tip.
     * @param limit The maximum number of dumps to return.
     * @param offset The number of dumps to skip.
     */
    public dumps(
        repository: string,
        query: string,
        visibleAtTip: boolean,
        limit: number,
        offset: number
    ): Promise<{ dumps: xrepoModels.LsifDump[]; totalCount: number }> {
        return this.xrepoDatabase.getDumps(repository, query, visibleAtTip, limit, offset)
    }

    /**
     * Get a dump by identifier.
     *
     * @param id The dump identifier.
     */
    public dump(id: xrepoModels.DumpId): Promise<xrepoModels.LsifDump | undefined> {
        return this.xrepoDatabase.getDumpById(id)
    }

    /**
     * Delete a dump.
     *
     * @param dump The dump.
     */
    public deleteDump(dump: xrepoModels.LsifDump): Promise<void> {
        return this.xrepoDatabase.deleteDump(dump)
    }

    /**
     * Determine if data exists for a particular document in this database.
     *
     * @param repository The repository name.
     * @param commit The commit.
     * @param path The path of the document.
     * @param ctx The tracing context.
     */
    public async exists(
        repository: string,
        commit: string,
        path: string,
        dumpId?: number,
        ctx: TracingContext = {}
    ): Promise<xrepoModels.LsifDump | undefined> {
        const closestDatabaseAndDump = await this.loadClosestDatabase(repository, commit, path, dumpId, ctx)
        if (!closestDatabaseAndDump) {
            return undefined
        }
        const { database, dump } = closestDatabaseAndDump
        return (await database.exists(pathToDatabase(dump.root, path))) ? dump : undefined
    }

    /**
     * Return the location for the definition of the reference at the given position. Returns
     * undefined if no dump can be loaded to answer this query.
     *
     * @param repository The repository name.
     * @param commit The commit.
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     * @param ctx The tracing context.
     */
    public async definitions(
        repository: string,
        commit: string,
        path: string,
        position: lsp.Position,
        dumpId?: number,
        ctx: TracingContext = {}
    ): Promise<InternalLocation[] | undefined> {
        const result = await this.internalDefinitions(repository, commit, path, position, dumpId, ctx)
        if (result === undefined) {
            return undefined
        }

        return result.locations
    }

    private async internalDefinitions(
        repository: string,
        commit: string,
        path: string,
        position: lsp.Position,
        dumpId?: number,
        ctx: TracingContext = {}
    ): Promise<{ dump: xrepoModels.LsifDump; locations: InternalLocation[] } | undefined> {
        const closestDatabaseAndDump = await this.loadClosestDatabase(repository, commit, path, dumpId, ctx)
        if (!closestDatabaseAndDump) {
            if (ctx.logger) {
                ctx.logger.warn('No database could be loaded', { repository, commit, path })
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
            return { dump, locations: definitions }
        }

        // Try to find definitions in other dumps
        const { document, ranges } = await database.getRangeByPosition(pathInDb, position, ctx)
        if (!document || ranges.length === 0) {
            return { dump, locations: [] }
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
                    // This symbol was imported from another database. See if we have xrepo
                    // definition for it.

                    const remoteDefinitions = await this.lookupMoniker(
                        document,
                        moniker,
                        dumpModels.DefinitionModel,
                        ctx
                    )
                    if (remoteDefinitions.length > 0) {
                        return { dump, locations: remoteDefinitions }
                    }
                } else {
                    // This symbol was not imported from another database. We search the definitions
                    // table of our own database in case there was a definition that wasn't properly
                    // attached to a result set but did have the correct monikers attached.

                    const monikerResults = await database.monikerResults(dumpModels.DefinitionModel, moniker, ctx)
                    const localDefinitions = monikerResults.map(loc => locationFromDatabase(dump.root, loc))
                    if (localDefinitions.length > 0) {
                        return { dump, locations: localDefinitions }
                    }
                }
            }
        }
        return { dump, locations: [] }
    }

    /**
     * Retrieve the package information from associated with the given moniker.
     *
     * @param document The document containing an instance of the moniker.
     * @param moniker The target moniker.
     * @param ctx The tracing context.
     */
    private lookupPackageInformation(
        document: dumpModels.DocumentData,
        moniker: dumpModels.MonikerData,
        ctx: TracingContext = {}
    ): dumpModels.PackageInformationData | undefined {
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
        document: dumpModels.DocumentData,
        moniker: dumpModels.MonikerData,
        model: typeof dumpModels.DefinitionModel | typeof dumpModels.ReferenceModel,
        ctx: TracingContext = {}
    ): Promise<InternalLocation[]> {
        const packageInformation = this.lookupPackageInformation(document, moniker, ctx)
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

        logSpan(ctx, 'package_entity', {
            moniker,
            packageInformation,
            packageRepository: packageEntity.dump.repository,
            packageCommit: packageEntity.dump.commit,
        })

        return (await this.createDatabase(packageEntity.dump).monikerResults(model, moniker, ctx)).map(loc =>
            locationFromDatabase(packageEntity.dump.root, loc)
        )
    }

    /**
     * Find the references of the target moniker outside of the current repository. If the moniker
     * has attached package information, then the cross-repo database is queried for the packages
     * that require this particular moniker identifier. These dumps are opened, and their
     * references tables are queried for the target moniker.
     *
     * @param dumpId The ID of the dump for which this database answers queries.
     * @param repository The repository for which this database answers queries.
     * @param moniker The target moniker.
     * @param packageInformation The target package.
     * @param limit The maximum number of remote dumps to search.
     * @param offset The number of remote dumps to skip.
     * @param ctx The tracing context.
     */
    private async remoteReferences(
        dumpId: xrepoModels.DumpId,
        repository: string,
        moniker: Pick<dumpModels.MonikerData, 'scheme' | 'identifier'>,
        packageInformation: Pick<dumpModels.PackageInformationData, 'name' | 'version'>,
        limit: number,
        offset: number,
        ctx: TracingContext = {}
    ): Promise<{ locations: InternalLocation[]; totalCount: number; newOffset: number }> {
        const { references, totalCount, newOffset } = await this.xrepoDatabase.getReferences({
            repository,
            scheme: moniker.scheme,
            identifier: moniker.identifier,
            name: packageInformation.name,
            version: packageInformation.version,
            limit,
            offset,
        })

        const dumps = references.map(r => r.dump)
        const locations = await this.locationsFromRemoteReferences(dumpId, moniker, dumps, ctx)
        return { locations, totalCount, newOffset }
    }

    /**
     * Find the references of the target moniker outside of the current dump but within a dump of
     * the same repository. If the moniker has attached package information, then the cross-repo
     * database is queried for the packages that require this particular moniker identifier. These
     * dumps are opened, and their references tables are queried for the target moniker.
     *
     * @param dumpId The ID of the dump for which this database answers queries.
     * @param repository The repository for which this database answers queries.
     * @param commit The commit of the references query.
     * @param moniker The target moniker.
     * @param packageInformation The target package.
     * @param limit The maximum number of remote dumps to search.
     * @param offset The number of remote dumps to skip.
     * @param ctx The tracing context.
     */
    private async sameRepositoryRemoteReferences(
        dumpId: xrepoModels.DumpId,
        repository: string,
        commit: string,
        moniker: Pick<dumpModels.MonikerData, 'scheme' | 'identifier'>,
        packageInformation: Pick<dumpModels.PackageInformationData, 'name' | 'version'>,
        limit: number,
        offset: number,
        ctx: TracingContext = {}
    ): Promise<{ locations: InternalLocation[]; totalCount: number; newOffset: number }> {
        const { references, totalCount, newOffset } = await this.xrepoDatabase.getSameRepoRemoteReferences({
            repository,
            commit,
            scheme: moniker.scheme,
            identifier: moniker.identifier,
            name: packageInformation.name,
            version: packageInformation.version,
            limit,
            offset,
        })

        const dumps = references.map(r => r.dump)
        const locations = await this.locationsFromRemoteReferences(dumpId, moniker, dumps, ctx)
        return { locations, totalCount, newOffset }
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
        dumpId: xrepoModels.DumpId,
        moniker: Pick<dumpModels.MonikerData, 'scheme' | 'identifier'>,
        dumps: xrepoModels.LsifDump[],
        ctx: TracingContext = {}
    ): Promise<InternalLocation[]> {
        logSpan(ctx, 'package_references', {
            references: dumps.map(d => ({ repository: d.repository, commit: d.commit })),
        })

        let locations: InternalLocation[] = []
        for (const dump of dumps) {
            // Skip the remote reference that show up for ourselves - we've already gathered
            // these in the previous step of the references query.
            if (dump.id === dumpId) {
                continue
            }

            const references = (
                await this.createDatabase(dump).monikerResults(dumpModels.ReferenceModel, moniker, ctx)
            ).map(loc => locationFromDatabase(dump.root, loc))
            locations = locations.concat(references)
        }

        return locations
    }

    /**
     * Create a database instance backed by the given dump.
     *
     * @param dump The dump.
     */
    private createDatabase(dump: xrepoModels.LsifDump): Database {
        return new Database(
            this.connectionCache,
            this.documentCache,
            this.resultChunkCache,
            dump,
            dbFilename(this.storageRoot, dump.id, dump.repository, dump.commit)
        )
    }

    /**
     * Return a list of locations which reference the definition at the given position. Returns
     * undefined if no dump can be loaded to answer this query.
     *
     * @param repository The repository name.
     * @param commit The commit.
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     * @param paginationContext Context describing the current request for paginated results.
     * @param ctx The tracing context.
     */
    public async references(
        repository: string,
        commit: string,
        path: string,
        position: lsp.Position,
        paginationContext: ReferencePaginationContext = { limit: 10 },
        dumpId?: number,
        ctx: TracingContext = {}
    ): Promise<{ locations: InternalLocation[]; cursor?: ReferencePaginationCursor } | undefined> {
        return this.internalReferences(repository, commit, path, position, paginationContext, dumpId, ctx)
    }

    private async internalReferences(
        repository: string,
        commit: string,
        path: string,
        position: lsp.Position,
        paginationContext: ReferencePaginationContext = { limit: 10 },
        dumpId?: number,
        ctx: TracingContext = {}
    ): Promise<
        { dump: xrepoModels.LsifDump; locations: InternalLocation[]; cursor?: ReferencePaginationCursor } | undefined
    > {
        if (paginationContext.cursor) {
            const dump = await this.xrepoDatabase.getDumpById(paginationContext.cursor.dumpId)
            if (dump === undefined) {
                return undefined
            }

            // Continue from previous page
            const results = await this.performRemoteReferences(
                repository,
                commit,
                paginationContext.limit,
                paginationContext.cursor,
                ctx
            )
            if (results !== undefined) {
                return { dump, ...results }
            }

            // Do not fall through
            return { dump, locations: [] }
        }

        const closestDatabaseAndDump = await this.loadClosestDatabase(repository, commit, path, dumpId, ctx)
        if (!closestDatabaseAndDump) {
            if (ctx.logger) {
                ctx.logger.warn('No database could be loaded', { repository, commit, path })
            }

            return undefined
        }
        const { database, dump, ctx: newCtx } = closestDatabaseAndDump

        // Construct path within dump
        const pathInDb = pathToDatabase(dump.root, path)

        // Try to find references in the same dump
        const dbReferences = await database.references(pathInDb, position, newCtx)
        let locations = dbReferences.map(loc => locationFromDatabase(dump.root, loc))

        // Next, we do a moniker search in two stages, described below. We process the
        // monikers for each range sequentially in order of priority for each stage, such
        // that import monikers, if any exist, will be processed first.

        const { document, ranges } = await database.getRangeByPosition(pathInDb, position, ctx)
        if (!document || ranges.length === 0) {
            return { dump, locations: [] }
        }

        for (const range of ranges) {
            const monikers = sortMonikers(
                Array.from(range.monikerIds).map(id => mustGet(document.monikers, id, 'monikers'))
            )

            // Next, we search the references table of our own database - this search is necessary,
            // but may be un-intuitive, but remember that a 'Find References' operation on a reference
            // should also return references to the definition. These are not necessarily fully linked
            // in the LSIF data.

            for (const moniker of monikers) {
                const monikerResults = await database.monikerResults(dumpModels.ReferenceModel, moniker, ctx)
                locations = locations.concat(monikerResults.map(loc => locationFromDatabase(dump.root, loc)))
            }

            // Next, we perform an xrepo search for uses of each nonlocal moniker. We stop processing after
            // the first moniker for which we received results. As we process monikers in an order that
            // considers moniker schemes, the first one to get results should be the most desirable.

            for (const moniker of monikers) {
                if (moniker.kind === 'import') {
                    // Get locations in the defining package
                    const monikerLocations = await this.lookupMoniker(document, moniker, dumpModels.ReferenceModel, ctx)
                    locations = locations.concat(monikerLocations)
                }

                const packageInformation = this.lookupPackageInformation(document, moniker, ctx)
                if (!packageInformation) {
                    continue
                }

                // Build pagination cursor that will start scanning results from
                // the beginning of the set of results: first, scan dumps of the same
                // repository, then scan dumps from remote repositories.

                const cursor: ReferencePaginationCursor = {
                    dumpId: dump.id,
                    scheme: moniker.scheme,
                    identifier: moniker.identifier,
                    name: packageInformation.name,
                    version: packageInformation.version,
                    phase: 'same-repo',
                    offset: 0,
                }

                const results = await this.performRemoteReferences(
                    repository,
                    commit,
                    paginationContext.limit,
                    cursor,
                    ctx
                )

                if (results !== undefined) {
                    return {
                        dump,
                        ...results,
                        // TODO - determine source of duplication
                        locations: uniqWith(locations.concat(results.locations), isEqual),
                    }
                }
            }
        }

        // TODO - determine source of duplication
        return { dump, locations: uniqWith(locations, isEqual) }
    }

    /**
     * Perform a remote reference lookup on the dumps of the same repository, then on dumps of
     * other repositories. The offset into the set of results (as well as the target set of dumps)
     * depends on the exact values of the pagination cursor. This method returns the new cursor.
     *
     * @param repository The repository name.
     * @param commit The target commit.
     * @param limit The maximum number of dumps to open.
     * @param cursor The pagination cursor.
     * @param ctx The tracing context.
     */
    private async performRemoteReferences(
        repository: string,
        commit: string,
        limit: number,
        cursor: ReferencePaginationCursor,
        ctx: TracingContext = {}
    ): Promise<{ locations: InternalLocation[]; cursor?: ReferencePaginationCursor } | undefined> {
        const moniker = { scheme: cursor.scheme, identifier: cursor.identifier }
        const packageInformation = { name: cursor.name, version: cursor.version }

        if (cursor.phase === 'same-repo') {
            const { locations, totalCount, newOffset } = await this.sameRepositoryRemoteReferences(
                cursor.dumpId,
                repository,
                commit,
                moniker,
                packageInformation,
                limit,
                cursor.offset,
                ctx
            )

            if (locations.length > 0) {
                let newCursor: ReferencePaginationCursor | undefined
                if (newOffset < totalCount) {
                    newCursor = {
                        ...cursor,
                        offset: newOffset,
                    }
                } else {
                    // Determine if there are any valid remote dumps we will open if
                    // we move onto a next page.
                    const { totalCount: remoteTotalCount } = await this.xrepoDatabase.getReferences({
                        repository,
                        scheme: moniker.scheme,
                        name: packageInformation.name,
                        version: packageInformation.version,
                        identifier: moniker.identifier,
                        limit: 1,
                        offset: 0,
                    })

                    // Only construct a cursor that will be valid on a subsequent
                    // request. We don't want the situation where there are no uses
                    // of a symbol outside of the current repository and we give a
                    // "load more" option that yields no additional results.

                    if (remoteTotalCount > 0) {
                        newCursor = {
                            ...cursor,
                            phase: 'remote-repo',
                            offset: 0,
                        }
                    }
                }

                return { locations, cursor: newCursor }
            }
        }

        const { locations, totalCount, newOffset } = await this.remoteReferences(
            cursor.dumpId,
            repository,
            moniker,
            packageInformation,
            limit,
            cursor.offset,
            ctx
        )

        if (locations.length > 0) {
            let newCursor: ReferencePaginationCursor | undefined
            if (newOffset < totalCount) {
                newCursor = {
                    ...cursor,
                    phase: 'remote-repo',
                    offset: newOffset,
                }
            }

            return { locations, cursor: newCursor }
        }

        return undefined
    }

    /**
     * Return the hover content for the definition or reference at the given position. Returns
     * undefined if no dump can be loaded to answer this query.
     *
     * @param repository The repository name.
     * @param commit The commit.
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     * @param ctx The tracing context.
     */
    public async hover(
        repository: string,
        commit: string,
        path: string,
        position: lsp.Position,
        dumpId?: number,
        ctx: TracingContext = {}
    ): Promise<{ text: string; range: lsp.Range } | null | undefined> {
        const closestDatabaseAndDump = await this.loadClosestDatabase(repository, commit, path, dumpId, ctx)
        if (!closestDatabaseAndDump) {
            if (ctx.logger) {
                ctx.logger.warn('No database could be loaded', { repository, commit, path })
            }

            return undefined
        }
        const { database, dump, ctx: newCtx } = closestDatabaseAndDump

        // Try to find hover in the same dump
        const hover = await database.hover(pathToDatabase(dump.root, path), position, newCtx)
        if (hover !== null) {
            return hover
        }

        // If we don't have a local hover, lookup the definitions of the
        // range and read the hover data from the remote database. This
        // can happen when the indexer only gives a moniker but does not
        // give hover data for externally defined symbols.

        const result = await this.internalDefinitions(repository, commit, path, position, dumpId, ctx)
        if (result === undefined || result.locations.length === 0) {
            return null
        }

        return this.createDatabase(result.locations[0].dump).hover(
            pathToDatabase(result.locations[0].dump.root, result.locations[0].path),
            result.locations[0].range.start,
            newCtx
        )
    }

    /**
     * Create a database instance for the given repository at the commit closest to the target
     * commit for which we have LSIF data. Also returns the dump instance backing the database.
     * Returns an undefined database and dump if no such dump can be found. Will also return a
     * tracing context tagged with the closest commit found. This new tracing context should
     * be used in all downstream requests so that the original commit and the effective commit
     * are both known.
     *
     * @param repository The repository name.
     * @param commit The target commit.
     * @param file One of the files in the dump.
     * @param ctx The tracing context.
     */
    private async loadClosestDatabase(
        repository: string,
        commit: string,
        file: string,
        dumpId?: number,
        ctx: TracingContext = {}
    ): Promise<{ database: Database; dump: xrepoModels.LsifDump; ctx: TracingContext } | undefined> {
        // Determine the closest commit that we actually have LSIF data for. If the commit is
        // not tracked, then commit data is requested from gitserver and insert the ancestors
        // data for this commit.
        const dump = await (dumpId
            ? this.xrepoDatabase.getDumpById(dumpId)
            : this.xrepoDatabase.findClosestDump(repository, commit, file, ctx, this.fetchConfiguration().gitServers))

        if (dump) {
            return { database: this.createDatabase(dump), dump, ctx: addTags(ctx, { closestCommit: dump.commit }) }
        }

        return undefined
    }
}
