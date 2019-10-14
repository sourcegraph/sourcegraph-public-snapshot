import * as lsp from 'vscode-languageserver-protocol'
import { ConnectionCache, DocumentCache, ResultChunkCache } from './cache'
import { dbFilename } from './util'
import { XrepoDatabase } from './xrepo'
import { TracingContext, logAndTraceCall, addTags } from './tracing'
import * as constants from './constants'
import { Database } from './database'
import { ConfigurationFetcher } from './config'

/**
 * A wrapper around code intelligence operations.
 */
export class Backend {
    private connectionCache = new ConnectionCache(constants.CONNECTION_CACHE_CAPACITY)
    private documentCache = new DocumentCache(constants.DOCUMENT_CACHE_CAPACITY)
    private resultChunkCache = new ResultChunkCache(constants.RESULT_CHUNK_CACHE_CAPACITY)

    /**
     * Create a new `Backend`.
     *
     * @param storageRoot The path where SQLite databases are stored.
     * @param xrepoDatabase The cross-repo database.
     * @param connectionCache The cache of SQLite connections.
     * @param documentCache The cache of loaded documents.
     * @param resultChunkCache The cache of loaded result chunks.
     */
    constructor(
        private storageRoot: string,
        private xrepoDatabase: XrepoDatabase,
        private fetchConfiguration: ConfigurationFetcher
    ) {}

    /**
     * Determine if data exists for a particular document in this database.
     *
     * @param path The path of the document.
     */
    public async exists(repository: string, commit: string, path: string, ctx: TracingContext = {}): Promise<boolean> {
        try {
            const { database, root } = await this.loadClosestDatabase(repository, commit, path, ctx)
            return await database.exists(this.pathToDatabase(root, path))
        } catch (e) {
            return false
        }
    }

    /**
     * Return the location for the definition of the reference at the given position.
     *
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     * @param ctx The tracing context.
     */
    public async definitions(
        repository: string,
        commit: string,
        path: string,
        position: lsp.Position,
        ctx: TracingContext = {}
    ): Promise<lsp.Location[]> {
        const { database, root, ctx: newCtx } = await this.loadClosestDatabase(repository, commit, path, ctx)
        return (await database.definitions(this.pathToDatabase(root, path), position, newCtx)).map(loc =>
            this.locationFromDatabase(root, loc)
        )
    }

    /**
     * Return a list of locations which reference the definition at the given position.
     *
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     * @param ctx The tracing context.
     */
    public async references(
        repository: string,
        commit: string,
        path: string,
        position: lsp.Position,
        ctx: TracingContext = {}
    ): Promise<lsp.Location[]> {
        const { database, root, ctx: newCtx } = await this.loadClosestDatabase(repository, commit, path, ctx)
        return (await database.references(this.pathToDatabase(root, path), position, newCtx)).map(loc =>
            this.locationFromDatabase(root, loc)
        )
    }

    /**
     * Return the hover content for the definition or reference at the given position.
     *
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     * @param ctx The tracing context.
     */
    public async hover(
        repository: string,
        commit: string,
        path: string,
        position: lsp.Position,
        ctx: TracingContext = {}
    ): Promise<lsp.Hover | null> {
        const { database, root, ctx: newCtx } = await this.loadClosestDatabase(repository, commit, path, ctx)
        return await database.hover(this.pathToDatabase(root, path), position, newCtx)
    }

    /**
     * Create a database instance for the given repository at the commit
     * closest to the target commit for which we have LSIF data. Returns
     * undefined if no such database can be created. Will also return a
     * tracing context tagged with the closest commit found. This new
     * tracing context should be used in all downstream requests so that
     * the original commit and the effective commit are both known.
     *
     * @param repository The repository name.
     * @param commit The target commit.
     * @param file One of the files in the dump.
     * @param ctx The tracing context.
     * @param gitserverUrls The set of ordered gitserver urls.
     */
    private async loadClosestDatabase(
        repository: string,
        commit: string,
        file: string,
        ctx: TracingContext
    ): Promise<{ database: Database; root: string; ctx: TracingContext }> {
        return await logAndTraceCall(ctx, 'loading closest database', async ctx => {
            // Determine the closest commit that we actually have LSIF data for. If the commit is
            // not tracked, then commit data is requested from gitserver and insert the ancestors
            // data for this commit.
            const dump = await logAndTraceCall(ctx, 'determining closest commit', (ctx: TracingContext) =>
                this.xrepoDatabase.findClosestDump(repository, commit, file, ctx, this.fetchConfiguration().gitServers)
            )
            if (!dump) {
                throw new Error('No LSIF data available.')
            }

            return {
                database: new Database(
                    this.storageRoot,
                    this.xrepoDatabase,
                    this.connectionCache,
                    this.documentCache,
                    this.resultChunkCache,
                    dump.id,
                    dbFilename(this.storageRoot, dump.id, dump.repository, dump.commit),
                    dump.root
                ),
                root: dump.root,
                ctx: addTags(ctx, { closestCommit: dump.commit }),
            }
        })
    }

    /**
     * Converts a file in the repository to the corresponding file in the
     * database.
     */
    private pathToDatabase(root: string, path: string): string {
        return stripPrefix(root, path)
    }

    /**
     * Converts a file in the database to the corresponding file in the
     * repository.
     */
    private pathFromDatabase(root: string, path: string): string {
        return `${root}${path}`
    }

    /**
     * Converts a location in the database to the corresponding location in the
     * repository.
     */
    private locationFromDatabase(root: string, { uri, range }: lsp.Location): lsp.Location {
        return lsp.Location.create(this.pathFromDatabase(root, uri), range)
    }
}

function stripPrefix(prefix: string, s: string): string {
    return s.startsWith(prefix) ? s.slice(prefix.length) : s
}
