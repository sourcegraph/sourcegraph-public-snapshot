import bodyParser from 'body-parser'
import express from 'express'
import { ERRNOLSIFDATA, makeBackend } from './backend'
import { readEnvInt, hasErrorCode, readEnv } from './util'
import { wrap } from 'async-middleware'

/**
 * Which port to run the LSIF server on. Defaults to 3186.
 */
const HTTP_PORT = readEnvInt('LSIF_HTTP_PORT', 3186)

const MAX_UPLOAD = readEnv('LSIF_MAX_UPLOAD', '100mb')

/**
 * Runs the HTTP server which accepts LSIF dump uploads and responds to LSIF requests.
 */
async function main(): Promise<void> {
    const backend = await makeBackend()
    const app = express()
    app.use(errorHandler)

    app.get('/ping', (req, res) => {
        res.send({ pong: 'pong' })
    })

    app.post(
        '/upload',
        bodyParser.raw({ limit: MAX_UPLOAD }),
        wrap(async (req, res) => {
            const { repository, commit } = req.query
            checkRepository(repository)
            checkCommit(commit)

            await backend.insertDump(req, repository, commit)
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
                const db = await backend.createRunner(repository, commit)
                const result = !file || (await db.exists(file))
                res.json(result)
            } catch (e) {
                if (hasErrorCode(e, ERRNOLSIFDATA)) {
                    res.json(false)
                } else {
                    throw e
                }
            }
        })
    )

    app.post(
        '/request',
        bodyParser.json({ limit: '1mb' }), // TODO - this seems HUGE for this request
        wrap(async (req, res) => {
            const { repository, commit } = req.query
            const { path, position, method } = req.body
            checkRepository(repository)
            checkCommit(commit)
            checkMethod(method, ['definitions', 'references', 'hover'])
            const cleanMethod = method as 'definitions' | 'references' | 'hover'

            try {
                const db = await backend.createRunner(repository, commit)
                res.json(await db[cleanMethod](path, position))
            } catch (e) {
                if (hasErrorCode(e, ERRNOLSIFDATA)) {
                    throw Object.assign(e, { status: 404 })
                }

                throw e
            }
        })
    )

    app.listen(HTTP_PORT, () => {
        console.log(`Listening for HTTP requests on port ${HTTP_PORT}`)
    })
}

/**
 * Middleware functino used to convert uncaught exceptions into 500 responses.
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

main().catch(e => {
    console.error(e)
})
