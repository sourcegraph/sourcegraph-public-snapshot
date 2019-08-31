import * as path from 'path'
import * as fs from 'mz/fs'
import * as readline from 'mz/readline'
import { Database } from './database'
import { hasErrorCode } from './util'
import { importLsif } from './importer'
import { XrepoDatabase } from './xrepo'
import { Readable } from 'stream'
import { ConnectionCache, DocumentCache } from './cache'
import { DefModel, MetaModel, RefModel, DocumentModel } from './models'
import { Edge, Vertex } from 'lsif-protocol'
import { EntityManager } from 'typeorm'

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
export class SQLiteBackend {
    constructor(
        private storageRoot: string,
        private xrepoDatabase: XrepoDatabase,
        private connectionCache: ConnectionCache,
        private documentCache: DocumentCache
    ) {}

    /**
     * Read the content of the temporary file containing a JSON-encoded LSIF
     * dump. Insert these contents into some storage with an encoding that
     * can be subsequently read by the `createRunner` method.
     */
    public async insertDump(input: Readable, repository: string, commit: string): Promise<void> {
        const outFile = makeFilename(this.storageRoot, repository, commit)

        try {
            await fs.unlink(outFile)
        } catch (e) {
            if (!hasErrorCode(e, 'ENOENT')) {
                throw e
            }
        }

        const { packages, references } = await this.connectionCache.withTransactionalEntityManager(
            outFile,
            [DefModel, DocumentModel, MetaModel, RefModel],
            (entityManager: EntityManager) => importLsif(entityManager, parseLines(readline.createInterface({ input })))
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
            yield JSON.parse(line) as Vertex | Edge
        } catch (e) {
            throw new Error(`Parsing failed for line ${i}: ${e}`)
        }

        i++
    }
}

export async function createBackend(
    storageRoot: string,
    connectionCache: ConnectionCache,
    documentCache: DocumentCache
): Promise<SQLiteBackend> {
    try {
        await fs.mkdir(storageRoot)
    } catch (e) {
        if (!hasErrorCode(e, 'EEXIST')) {
            throw e
        }
    }

    const filename = path.join(storageRoot, 'correlation.db')

    try {
        await fs.stat(filename)
    } catch (e) {
        if (!hasErrorCode(e, 'ENOENT')) {
            throw e
        }
    }

    return new SQLiteBackend(storageRoot, new XrepoDatabase(connectionCache, filename), connectionCache, documentCache)
}
