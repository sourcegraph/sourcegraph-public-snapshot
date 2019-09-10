import * as fs from 'mz/fs'
import * as zlib from 'mz/zlib'
import exitHook from 'async-exit-hook'
import { Backend, createBackend } from './backend'
import { ConnectionCache, DocumentCache, ResultChunkCache } from './cache'
import { readEnvInt, logErrorAndExit } from './util'
import { Worker } from 'node-resque'

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
 * The maximum number of documents that can be held in memory at once.
 */
const DOCUMENT_CACHE_CAPACITY = readEnvInt('DOCUMENT_CACHE_CAPACITY', 1000)

/**
 * The maximum number of result chunks that can be held in memory at once.
 */
const RESULT_CHUNK_CACHE_SIZE = readEnvInt('RESULT_CHUNK_CACHE_SIZE', 1000)

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
    const documentCache = new DocumentCache(DOCUMENT_CACHE_CAPACITY)
    const resultChunkCache = new ResultChunkCache(RESULT_CHUNK_CACHE_SIZE)
    const backend = await createBackend(STORAGE_ROOT, connectionCache, documentCache, resultChunkCache)
    const convertJob = { perform: createConvertJob(backend) }

    const worker = new Worker({ connection: connectionOptions, queues: ['lsif'] }, { convert: convertJob })
    worker.on('error', logErrorAndExit)
    await worker.connect()
    exitHook(() => worker.end())
    worker.start().catch(logErrorAndExit)

    if (LOG_READY) {
        console.log('Listening for uploads')
    }
}

function createConvertJob(backend: Backend): (repository: string, commit: string, filename: string) => Promise<void> {
    return async (repository, commit, filename) => {
        console.log(`Converting ${repository}@${commit}`)
        const input = fs.createReadStream(filename).pipe(zlib.createGunzip())
        await backend.insertDump(input, repository, commit)
        await fs.unlink(filename)
    }
}

main().catch(logErrorAndExit)
