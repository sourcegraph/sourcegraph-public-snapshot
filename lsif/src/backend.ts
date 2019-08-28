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

export const ERRNOLSIFDATA = 'NoLSIFData'

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
 * Where on the file system to store LSIF files.
 */
const STORAGE_ROOT = process.env.LSIF_STORAGE_ROOT || 'lsif-storage'

/**
 * Backend for LSIF dumps stored in SQLite.
 */
export class SQLiteBackend {
    constructor(
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
        const outFile = makeFilename(repository, commit)

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

        // These needs to be done in sequence as SQLite can only have one write txn at a time.
        await this.xrepoDatabase.addPackages(repository, commit, packages)
        await this.xrepoDatabase.addReferences(repository, commit, references)
    }

    /**
     * Create a database relevant to the given repository and commit hash. This
     * assumes that data for this subset of data has already been inserted via
     * `insertDump` (otherwise this method is expected to throw).
     */
    public async createDatabase(repository: string, commit: string): Promise<Database> {
        const file = makeFilename(repository, commit)

        try {
            await fs.stat(file)
        } catch (e) {
            if (hasErrorCode(e, 'ENOENT')) {
                throw new NoLSIFDataError(repository, commit)
            }

            throw e
        }

        return new Database(this.xrepoDatabase, this.connectionCache, this.documentCache, file)
    }
}

/**
 *.Computes the filename of the LSIF dump from the given repository and commit hash.
 */
export function makeFilename(repository: string, commit: string): string {
    return path.join(STORAGE_ROOT, `${encodeURIComponent(repository)}@${commit}.lsif.db`)
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
            throw new Error(`Parsing failed for line:\n${i}`)
        }

        i++
    }
}

export async function createBackend(
    connectionCache: ConnectionCache,
    documentCache: DocumentCache
): Promise<SQLiteBackend> {
    try {
        await fs.mkdir(STORAGE_ROOT)
    } catch (e) {
        if (!hasErrorCode(e, 'EEXIST')) {
            throw e
        }
    }

    const filename = path.join(STORAGE_ROOT, 'correlation.db')

    try {
        await fs.stat(filename)
    } catch (e) {
        if (!hasErrorCode(e, 'ENOENT')) {
            throw e
        }
    }

    return new SQLiteBackend(new XrepoDatabase(connectionCache, filename), connectionCache, documentCache)
}
