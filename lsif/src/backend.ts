import * as fs from 'mz/fs'
import * as path from 'path'
import * as readline from 'mz/readline'
import { ConnectionCache, DocumentCache, ResultChunkCache } from './cache'
import { Database } from './database'
import { DefinitionModel, DocumentModel, MetaModel, ReferenceModel, ResultChunkModel } from './models.database'
import { Edge, Vertex } from 'lsif-protocol'
import { hasErrorCode } from './util'
import { importLsif } from './importer'
import { Readable } from 'stream'
import { XrepoDatabase } from './xrepo'

export const ERRNOLSIFDATA = 'NoLSIFDataError'

/**
 * An error thrown when no LSIF database can be found on disk.
 */
export class NoLSIFDataError extends Error {
    public readonly name = ERRNOLSIFDATA

    constructor(repository: string, commit: string) {
        super(`No LSIF data available for ${repository}@${commit}.`)
    }
}

/**
 * Backend for LSIF dumps stored in SQLite.
 */
export class Backend {
    constructor(
        private storageRoot: string,
        private xrepoDatabase: XrepoDatabase,
        private connectionCache: ConnectionCache,
        private documentCache: DocumentCache,
        private resultChunkCache: ResultChunkCache
    ) {}

    /**
     * Read the content of the temporary file containing a JSON-encoded LSIF
     * dump. Insert these contents into some storage with an encoding that
     * can be subsequently read by the `createRunner` method.
     */
    public async insertDump(input: Readable, repository: string, commit: string): Promise<void> {
        const outFile = makeFilename(this.storageRoot, repository, commit)

        try {
            // Remove old databse file, if it exists
            await fs.unlink(outFile)
        } catch (e) {
            if (!hasErrorCode(e, 'ENOENT')) {
                throw e
            }
        }

        // Remove old data from xrepo database
        await this.xrepoDatabase.clearCommit(repository, commit)

        // Remove any connection in the cache to the file we just removed
        await this.connectionCache.bustKey(outFile)

        const { packages, references } = await this.connectionCache.withTransactionalEntityManager(
            outFile,
            [DefinitionModel, DocumentModel, MetaModel, ReferenceModel, ResultChunkModel],
            entityManager => importLsif(entityManager, parseLines(readline.createInterface({ input }))),
            async connection => {
                await connection.query('PRAGMA synchronous = OFF')
                await connection.query('PRAGMA journal_mode = OFF')
            }
        )

        // These needs to be done in sequence as SQLite can only have one
        // write txn at a time without causing the other one to abort with
        // a weird error.
        await this.xrepoDatabase.addPackages(repository, commit, packages)
        await this.xrepoDatabase.addReferences(repository, commit, references)
    }

    /**
     * Create a database relevant to the given repository and commit hash. This
     * assumes that data for this subset of data has already been inserted via
     * `insertDump` (otherwise this method is expected to throw).
     */
    public async createDatabase(repository: string, commit: string): Promise<Database> {
        const file = makeFilename(this.storageRoot, repository, commit)

        try {
            await fs.stat(file)
        } catch (e) {
            if (hasErrorCode(e, 'ENOENT')) {
                throw new NoLSIFDataError(repository, commit)
            }

            throw e
        }

        return new Database(
            this.storageRoot,
            this.xrepoDatabase,
            this.connectionCache,
            this.documentCache,
            this.resultChunkCache,
            repository,
            commit,
            file
        )
    }
}

/**
 * Create the path of the SQLite database file for the given repository and commit.
 *
 * @param storageRoot The path where SQLite databases are stored.
 * @param repository The repository name.
 * @param commit The repository commit.
 */
export function makeFilename(storageRoot: string, repository: string, commit: string): string {
    return path.join(storageRoot, `${encodeURIComponent(repository)}@${commit}.lsif.db`)
}

/**
 * Converts streaming JSON input into an iterable of vertex and edge objects.
 *
 * @param lines The stream of raw, uncompressed JSON lines.
 */
async function* parseLines(lines: AsyncIterable<string>): AsyncIterable<Vertex | Edge> {
    let i = 0
    for await (const line of lines) {
        try {
            yield JSON.parse(line)
        } catch (e) {
            throw Object.assign(
                new Error(`Failed to process line #${i + 1} (${JSON.stringify(line)}): Invalid JSON.`),
                { status: 422 }
            )
        }

        i++
    }
}

export async function createBackend(
    storageRoot: string,
    connectionCache: ConnectionCache,
    documentCache: DocumentCache,
    resultChunkCache: ResultChunkCache
): Promise<Backend> {
    try {
        await fs.mkdir(storageRoot)
    } catch (e) {
        if (!hasErrorCode(e, 'EEXIST')) {
            throw e
        }
    }

    return new Backend(
        storageRoot,
        new XrepoDatabase(connectionCache, path.join(storageRoot, 'xrepo.db')),
        connectionCache,
        documentCache,
        resultChunkCache
    )
}
