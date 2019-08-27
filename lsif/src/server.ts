import bodyParser from 'body-parser'
import express from 'express'
import { ERRNOLSIFDATA, makeBackend } from './backend'
import { readEnvInt, hasErrorCode, readEnv } from './util'
import { ConnectionCache, DocumentCache } from './cache'
import { zlib } from 'mz'

/**
 * Which port to run the LSIF server on. Defaults to 3186.
 */
const HTTP_PORT = readEnvInt('LSIF_HTTP_PORT', 3186)

/**
 * The maximum size of an LSIF dump upload.
 */
const MAX_UPLOAD = readEnv('LSIF_MAX_UPLOAD', '100mb')

/**
 * The number of SQLite connections that can be opened at once. This
 * value may be exceeded for a short period if many handles are held
 * at once.
 */
const CONNECTION_CACHE_SIZE = readEnvInt('CONNECTION_CACHE_SIZE', 20)

/**
 * The maximum number of documents that can be held in memory at once.
 */
const DOCUMENT_CACHE_SIZE = readEnvInt('DOCUMENT_CACHE_SIZE', 100)

/**
 * Runs the HTTP server which accepts LSIF dump uploads and responds to LSIF requests.
 */
async function main(): Promise<void> {
    const connectionCache = new ConnectionCache(CONNECTION_CACHE_SIZE)
    const documentCache = new DocumentCache(DOCUMENT_CACHE_SIZE)
    const backend = await makeBackend(connectionCache, documentCache)
    const app = express()
    app.use(errorHandler)

    app.get('/ping', (_, res) => {
        res.send({ pong: 'pong' })
    })

    app.post(
        '/upload',
        bodyParser.raw({ limit: MAX_UPLOAD }),
        wrap(async (req, res) => {
            const { repository, commit } = req.query
            checkRepository(repository)
            checkCommit(commit)
            await backend.insertDump(req.pipe(zlib.createGunzip()), repository, commit)
            res.json(null)
        })
    )

    app.post(
        '/exists',
        wrap(async (req, res) => {
            const { repository, commit, file } = req.query
            checkRepository(repository)
            checkCommit(commit)

            try {
                const db = await backend.createDatabase(repository, commit)
                const result = !file || (await db.exists(file))
                res.json(result)
            } catch (e) {
                if (hasErrorCode(e, ERRNOLSIFDATA)) {
                    res.json(false)
                    return
                }

                throw e
            }
        })
    )

    app.post(
        '/definitions',
        bodyParser.json({ limit: '1mb' }),
        wrap(async (req, res) => {
            const { repository, commit } = req.query
            const { path, position } = req.body
            checkRepository(repository)
            checkCommit(commit)

            const db = await backend.createDatabase(repository, commit)
            const data = await db.definitions(path, position)
            res.json({ data })
        })
    )

    app.post(
        '/references',
        bodyParser.json({ limit: '1mb' }),
        wrap(async (req, res) => {
            const { repository, commit, page } = req.query
            const { path, position } = req.body
            checkRepository(repository)
            checkCommit(commit)

            const db = await backend.createDatabase(repository, commit)
            const { data, nextPage } = await db.references(path, position, page)
            res.json({ data, nextPage })
        })
    )

    app.post(
        '/hover',
        bodyParser.json({ limit: '1mb' }),
        wrap(async (req, res) => {
            // TODO - why are these split this way?
            const { repository, commit } = req.query
            const { path, position } = req.body
            checkRepository(repository)
            checkCommit(commit)

            const db = await backend.createDatabase(repository, commit)
            const data = await db.hover(path, position)
            res.json({ data })
        })
    )

    app.listen(HTTP_PORT, () => {
        console.log(`Listening for HTTP requests on port ${HTTP_PORT}`)
    })
}

/**
 * Turn an async handler function into an express handler. Catch any exceptions that
 * are thrown and pass them to the registered error handler. If the exception is an
 * ERRNOLSIFDATA exception, return with a 404 response.
 *
 * @param handler The two-argument handler.
 */
function wrap(
    handler: (req: express.Request, res: express.Response) => Promise<void>
): (req: express.Request, res: express.Response, next: express.NextFunction) => Promise<void> {
    return async (req: express.Request, res: express.Response, next: express.NextFunction) => {
        try {
            return await handler(req, res)
        } catch (e) {
            if (hasErrorCode(e, ERRNOLSIFDATA)) {
                res.status(404).send({ message: e.message })
                return
            }

            return next(e)
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

main().catch(e => {
    console.error(e)
})
