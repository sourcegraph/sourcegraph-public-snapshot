import * as path from 'path'
import bodyParser from 'body-parser'
import exitHook from 'async-exit-hook'
import express from 'express'
import uuid from 'uuid'
import { ConnectionCache, DocumentCache } from './cache'
import { Database } from './database'
import { fs, zlib } from 'mz'
import { hasErrorCode, makeFilename } from './util'
import { Queue, Scheduler } from 'node-resque'
import { XrepoDatabase } from './xrepo'
import {
    STORAGE_ROOT,
    CONNECTION_CACHE_SIZE,
    DOCUMENT_CACHE_SIZE,
    MAX_UPLOAD,
    HTTP_PORT,
    REDIS_PORT,
    REDIS_HOST,
} from './settings'

/**
 * Runs the HTTP server which accepts LSIF dump uploads and responds to LSIF requests.
 */
async function main(): Promise<void> {
    await createDirectory(STORAGE_ROOT)
    await createDirectory(path.join(STORAGE_ROOT, 'uploads'))

    const connectionCache = new ConnectionCache(CONNECTION_CACHE_SIZE)
    const documentCache = new DocumentCache(DOCUMENT_CACHE_SIZE)
    const xrepoDatabase = new XrepoDatabase(connectionCache)
    const queue = await setupQueue()

    const app = express()
    app.use(errorHandler)

    const createDatabase = async (repository: string, commit: string): Promise<Database | undefined> => {
        const file = makeFilename(repository, commit)

        try {
            await fs.stat(file)
        } catch (e) {
            if (hasErrorCode(e, 'ENOENT')) {
                return undefined
            }

            throw e
        }

        return new Database(xrepoDatabase, connectionCache, documentCache, file)
    }

    app.post('/upload', bodyParser.raw({ limit: MAX_UPLOAD }), async (req, res, next) => {
        try {
            const { repository, commit } = req.query
            checkRepository(repository)
            checkCommit(commit)
            const filename = path.join(STORAGE_ROOT, 'uploads', uuid.v4())
            const writeStream = fs.createWriteStream(filename)
            // TODO - do validation
            req.pipe(zlib.createGunzip())
                .pipe(zlib.createGzip())
                .pipe(writeStream)
            await new Promise(resolve => writeStream.on('finish', resolve))
            await queue.enqueue('lsif', 'convert', [repository, commit, filename])
            res.json(null)
        } catch (e) {
            return next(e)
        }
    })

    app.post('/exists', async (req, res, next) => {
        try {
            const { repository, commit, file } = req.query
            checkRepository(repository)
            checkCommit(commit)

            const db = await createDatabase(repository, commit)
            if (!db) {
                res.json(false)
                return
            }

            const result = !file || (await db.exists(file))
            res.json(result)
        } catch (e) {
            return next(e)
        }
    })

    app.post('/request', bodyParser.json({ limit: '1mb' }), async (req, res, next) => {
        try {
            const { repository, commit } = req.query
            const { path, position, method } = req.body
            checkRepository(repository)
            checkCommit(commit)
            checkMethod(method, ['definitions', 'references', 'hover'])
            const cleanMethod = method as 'definitions' | 'references' | 'hover'

            const db = await createDatabase(repository, commit)
            if (!db) {
                throw Object.assign(new Error('No LSIF data'), { status: 404 })
            }

            res.json(await db[cleanMethod](path, position))
        } catch (e) {
            return next(e)
        }
    })

    app.listen(HTTP_PORT, () => {
        console.log(`Listening for HTTP requests on port ${HTTP_PORT}`)
    })
}

/**
 * Connect and start an active connection to the worker queue. We also run a
 * node-resque scheduler on each server instance, as these are guaranteed to
 * always be up with a responsive system. The schedulers will do their own
 * master election via a redis key and will check for dead workers attached
 * to the queue.
 */
async function setupQueue(): Promise<Queue> {
    const connectionOptions = {
        host: REDIS_HOST,
        port: REDIS_PORT,
        namespace: 'lsif',
    }

    const queue = new Queue({ connection: connectionOptions })
    queue.on('error', e => console.error(e))
    await queue.connect()
    exitHook(() => queue.end())

    const scheduler = new Scheduler({ connection: connectionOptions })
    scheduler.on('error', e => console.error(e))
    await scheduler.connect()
    exitHook(() => scheduler.end())
    scheduler.start().catch(e => console.error(e))

    return queue
}

/**
 * Ensure the given directory path exists.
 *
 * @param path The directory path.
 */
async function createDirectory(path: string): Promise<void> {
    try {
        await fs.mkdir(path)
    } catch (e) {
        if (!hasErrorCode(e, 'EEXIST')) {
            throw e
        }
    }
}

/* eslint-disable @typescript-eslint/no-unused-vars */
/* eslint-disable @typescript-eslint/no-explicit-any */

/**
 * Middleware function used to convert uncaught exceptions into 500 responses.
 */
function errorHandler(err: any, req: express.Request, res: express.Response, next: express.NextFunction): void {
    if (err && err.status) {
        res.status(err.status).send({ message: err.message })
        return
    }

    console.error(err)
    res.status(500).send({ message: 'Unknown error' })
}

/**
 * Throws an error with status 400 if the repository string is invalid.
 */
export function checkRepository(repository: any): void {
    if (typeof repository !== 'string') {
        throw Object.assign(new Error('Must specify the repository (usually of the form github.com/user/repo)'), {
            status: 400,
        })
    }
}

/**
 * Throws an error with status 400 if the commit string is invalid.
 */
export function checkCommit(commit: any): void {
    if (typeof commit !== 'string' || commit.length !== 40 || !/^[0-9a-f]+$/.test(commit)) {
        throw Object.assign(new Error('Must specify the commit as a 40 character hash ' + commit), { status: 400 })
    }
}

/**
 * Throws an error with status 422 if the requested method is not supported.
 */
export function checkMethod(method: string, supportedMethods: string[]): void {
    if (!supportedMethods.includes(method)) {
        throw Object.assign(new Error(`Method must be one of ${Array.from(supportedMethods).join(', ')}`), {
            status: 422,
        })
    }
}

main().catch(e => console.error(e))
