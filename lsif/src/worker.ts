import * as fs from 'mz/fs'
import * as path from 'path'
import * as zlib from 'mz/zlib'
import exitHook from 'async-exit-hook'
import uuid from 'uuid'
import { addToXrepoDatabase, convertLsif } from './conversion'
import { ConnectionCache } from './cache'
import { createDatabaseFilename, logErrorAndExit, readEnvInt } from './util'
import { Worker } from 'node-resque'
import { XrepoDatabase } from './xrepo'

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
const CONNECTION_CACHE_CAPACITY = readEnvInt('CONNECTION_CACHE_CAPACITY', 1024 * 1024 * 1024)

/**
 * Whether or not to log a message when the HTTP server is ready and listening.
 */
const LOG_READY = process.env.DEPLOY_TYPE === 'dev'

/**
 * Where on the file system to store LSIF files.
 */
const STORAGE_ROOT = process.env.LSIF_STORAGE_ROOT || 'lsif-storage'

/**
 * Runs the worker which accepts LSIF conversion jobs from node-resque.
 */
async function main(): Promise<void> {
    const connectionOptions = {
        host: REDIS_HOST,
        port: REDIS_PORT,
        namespace: 'lsif',
    }

    const connectionCache = new ConnectionCache(CONNECTION_CACHE_CAPACITY)
    const filename = path.join(STORAGE_ROOT, 'correlation.db')
    const xrepoDatabase = new XrepoDatabase(connectionCache, filename)
    const convertJob = { perform: createConvertJob(xrepoDatabase) }

    const worker = new Worker({ connection: connectionOptions, queues: ['lsif'] }, { convert: convertJob })
    worker.on('error', logErrorAndExit)
    await worker.connect()
    exitHook(() => worker.end())
    worker.start().catch(logErrorAndExit)

    if (LOG_READY) {
        console.log('Listening for uploads')
    }
}

/**
 * Create a job that takes a repository, commit, and filename containing the gzipped
 * input of an LSIF dump and converts it to a SQLite database. This will also populate
 * the correlation database for this dump.
 *
 * @param xrepoDatabase The correlation database.
 */
function createConvertJob(
    xrepoDatabase: XrepoDatabase
): (repository: string, commit: string, filename: string) => Promise<void> {
    return async (repository, commit, filename) => {
        console.log(`Converting ${repository}@${commit}`)

        const input = fs.createReadStream(filename).pipe(zlib.createGunzip())
        const tempFile = path.join(STORAGE_ROOT, 'tmp', uuid.v4())

        try {
            // Create database in a temp path
            const { packages, references } = await convertLsif(input, tempFile)

            // Move the temp file where it can be found by the server
            await fs.rename(tempFile, createDatabaseFilename(STORAGE_ROOT, repository, commit))

            // Add the new database to the correlation db
            await addToXrepoDatabase(xrepoDatabase, packages, references, repository, commit)
        } catch (e) {
            // Don't leave busted artifacts
            await fs.unlink(tempFile)
            throw e
        }

        // Remove input
        await fs.unlink(filename)
    }
}

main().catch(logErrorAndExit)
