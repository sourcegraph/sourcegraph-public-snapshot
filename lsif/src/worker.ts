import * as fs from 'mz/fs'
import * as path from 'path'
import * as temp from 'temp'
import * as zlib from 'mz/zlib'
import exitHook from 'async-exit-hook'
import { ConnectionCache } from './cache'
import { DefinitionModel, DocumentModel, MetaModel, ReferenceModel, ResultChunkModel } from './models.database'
import { logErrorAndExit, readEnvInt, createDatabaseFilename } from './util'
import { Worker } from 'node-resque'
import { XrepoDatabase, Package, SymbolReferences } from './xrepo'
import { readline } from 'mz'
import { importLsif } from './importer'
import { Readable } from 'stream'
import { createConnection } from 'typeorm'
import { Vertex, Edge } from 'lsif-protocol'

/**
 * The host running the redis instance containing work queues. Defaults to localhost.
 */
const REDIS_HOST = process.env.LSIF_REDIS_HOST || 'localhost'

/**
 * The port of the redis instance containing work queues. Defaults to 6379.
 */
const REDIS_PORT = readEnvInt('LSIF_REDIS_PORT', 6379)

/**
 * The number of SQLite connections that can be opened at once. This
 * value may be exceeded for a short period if many handles are held
 * at once.
 */
const CONNECTION_CACHE_CAPACITY = readEnvInt('CONNECTION_CACHE_CAPACITY', 1000)

/**
 * Whether or not to log a message when the HTTP server is ready and listening.
 */
const LOG_READY = process.env.DEPLOY_TYPE === 'dev'

/**
 * Where on the file system to store LSIF files.
 */
const STORAGE_ROOT = process.env.LSIF_STORAGE_ROOT || 'lsif-storage'

async function main(): Promise<void> {
    const connectionOptions = {
        host: REDIS_HOST,
        port: REDIS_PORT,
        namespace: 'lsif',
    }

    const connectionCache = new ConnectionCache(CONNECTION_CACHE_CAPACITY)
    const filename = path.join(STORAGE_ROOT, 'correlation.db')
    const xrepoDatabase = new XrepoDatabase(connectionCache, filename)
    const convertJob = { perform: makeConvertJob(xrepoDatabase) }

    const worker = new Worker({ connection: connectionOptions, queues: ['lsif'] }, { convert: convertJob })
    worker.on('error', logErrorAndExit)
    await worker.connect()
    exitHook(() => worker.end())
    worker.start().catch(logErrorAndExit)

    if (LOG_READY) {
        console.log('Listening for uploads')
    }
}

function makeConvertJob(
    xrepoDatabase: XrepoDatabase
): (repository: string, commit: string, filename: string) => Promise<void> {
    return async (repository, commit, filename) => {
        const input = fs.createReadStream(filename).pipe(zlib.createGunzip())
        const tempFile = temp.path()

        try {
            const { packages, references } = await convertLsif(input, tempFile)
            // Move the temp file where it can be found by the server
            await fs.rename(tempFile, createDatabaseFilename(STORAGE_ROOT, repository, commit))
            await addToXrepoDatabase(xrepoDatabase, packages, references, repository, commit)
        } catch (e) {
            await fs.unlink(tempFile)
            throw e
        }

        await fs.unlink(filename)
    }
}

export async function convertLsif(
    input: Readable,
    database: string
): Promise<{ packages: Package[]; references: SymbolReferences[] }> {
    // TODO - standardize
    const connection = await createConnection({
        database,
        entities: [DefinitionModel, DocumentModel, MetaModel, ReferenceModel, ResultChunkModel],
        type: 'sqlite',
        name: database,
        synchronize: true,
        logging: ['error', 'warn'],
        maxQueryExecutionTime: 1000,
    })

    try {
        await connection.query('PRAGMA synchronous = OFF')
        await connection.query('PRAGMA journal_mode = OFF')

        return await connection.transaction(entityManager =>
            importLsif(entityManager, parseLines(readline.createInterface({ input })))
        )
    } finally {
        await connection.close()
    }
}

export async function addToXrepoDatabase(
    xrepoDatabase: XrepoDatabase,
    packages: Package[],
    references: SymbolReferences[],
    repository: string,
    commit: string
): Promise<void> {
    await xrepoDatabase.addPackages(repository, commit, packages)
    await xrepoDatabase.addReferences(repository, commit, references)
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

main().catch(logErrorAndExit)
