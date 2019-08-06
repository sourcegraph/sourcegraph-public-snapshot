import * as Prometheus from 'prom-client'
import * as tmp from 'tmp-promise'
import bodyParser from 'body-parser'
import express from 'express'
import { Cache } from './cache'
import { createPrometheusReporters, emit } from './prometheus'
import { ERRNOLSIFDATA } from './backend'
import { fs } from 'mz'
import { JsonDatabase } from './ms/json'
import { noopTransformer } from './ms/database'
import { readEnvInt } from './env'
import { SQLiteGraphBackend } from './sqlite'
import { wrap } from 'async-middleware'

/**
 * Which port to run the LSIF server on. Defaults to 3186.
 */
const PORT = readEnvInt({ key: 'LSIF_HTTP_PORT', defaultValue: 3186 })

/**
 * Limit on the file size accepted by the /upload endpoint. Defaults to 100MB.
 */
const MAX_FILE_SIZE = readEnvInt({ key: 'LSIF_MAX_FILE_SIZE', defaultValue: 100 * 1024 * 1024 })

/**
 * List of supported LSIF methods that can be passed to query runners.
 */
type SupportedMethods = 'hover' | 'definitions' | 'references'

const SUPPORTED_METHODS: Set<SupportedMethods> = new Set(['hover', 'definitions', 'references'])

/**
 * Runs the HTTP server which accepts LSIF dump uploads and responds to LSIF requests.
 */
function main(): void {
    const app = express()
    const backend = new SQLiteGraphBackend()
    const cache = new Cache()
    const prometheusReporters = createPrometheusReporters()

    app.use(errorHandler)

    app.get('/ping', (req, res) => {
        res.send({ pong: 'pong' })
    })

    app.get('/metrics', (req, res) => {
        res.set('Content-Type', Prometheus.register.contentType)
        res.end(Prometheus.register.metrics())
    })

    app.post(
        '/upload',
        wrap(async (req, res) => {
            const { repository, commit } = req.query
            checkRepository(repository)
            checkCommit(commit)
            checkContentLength(req.header('Content-Length'))

            // Create temp file to receive the request body
            const tempFile = await tmp.file()

            try {
                // Read the content and ensure the body is a valid LSIF dump
                const contentLength = await readContent(req, tempFile.path)
                await new JsonDatabase().load(tempFile.path, () => noopTransformer)
                const { insertStats } = await backend.insertDump(tempFile.path, repository, commit, contentLength)

                // Bust the cache
                cache.delete(repository, commit)

                res.send({
                    data: null,
                    stats: {
                        insertStats: insertStats,
                    },
                })

                // Emit metrics
                emit(prometheusReporters, insertStats)
            } finally {
                // Temp files are cleaned up on process exit, but we want to do it
                // proactively and in the event of exceptions so we do not fill up
                // the temporary directory on the machine.
                await fs.unlink(tempFile.path)
            }
        })
    )

    app.post(
        '/exists',
        wrap(async (req, res) => {
            const { repository, commit, file } = req.query
            checkRepository(repository)
            checkCommit(commit)

            try {
                const { result, cacheStats, createRunnerStats } = await cache.withDB(
                    backend,
                    repository,
                    commit,
                    async queryRunner => {
                        return !file || queryRunner.exists(file)
                    }
                )

                res.send({
                    data: result,
                    stats: {
                        cacheStats: cacheStats,
                        createRunnerStats: createRunnerStats,
                    },
                })

                // Emit metrics
                emit(prometheusReporters, cacheStats)

                if (createRunnerStats) {
                    emit(prometheusReporters, createRunnerStats)
                }
            } catch (e) {
                if ('code' in e && e.code === ERRNOLSIFDATA) {
                    // TODO(efritz) - emit stats
                    res.send({ data: false, stats: {} })
                } else {
                    throw e
                }
            }
        })
    )

    app.post(
        '/request',
        bodyParser.json({ limit: '1mb' }),
        wrap(async (req, res) => {
            const { repository, commit } = req.query
            const { path, position, method } = req.body
            checkRepository(repository)
            checkCommit(commit)
            checkMethod(method)

            try {
                const {
                    result: { result, queryStats },
                    cacheStats,
                } = await cache.withDB(backend, repository, commit, async queryRunner => {
                    return await queryRunner.query(method, path, position)
                })

                res.json({
                    data: result || null,
                    stats: {
                        cacheStats: cacheStats,
                        queryStats: queryStats,
                    },
                })

                // Emit metrics
                emit(prometheusReporters, cacheStats)
                emit(prometheusReporters, queryStats)
            } catch (e) {
                if ('code' in e && e.code === ERRNOLSIFDATA) {
                    throw Object.assign(e, { status: 404 })
                }

                throw e
            }
        })
    )

    app.listen(PORT, () => {
        console.log(`Listening for HTTP requests on port ${PORT}`)
    })
}

//
// Helpers

/**
 * Middleware functino used to convert uncaught exceptions into 500 responses.
 */
function errorHandler(err: any, req: express.Request, res: express.Response, next: express.NextFunction) {
    if (err && err.status) {
        res.status(err.status).send({ message: err.message })
        return
    }

    console.error(err)
    res.status(500).send({ message: 'Unknown error' })
}

/**
 * Read the request body into the given file path.
 */
async function readContent(req: express.Request, tempPath: string): Promise<number> {
    let contentLength = 0
    const tempFileWriteStream = fs.createWriteStream(tempPath)

    return new Promise((resolve, reject) => {
        req.on('data', chunk => {
            contentLength += chunk.length
            if (contentLength > MAX_FILE_SIZE) {
                tempFileWriteStream.destroy()

                reject(
                    Object.assign(
                        new Error(
                            `The size of the given LSIF file (${contentLength} bytes so far) exceeds the max of ${MAX_FILE_SIZE}`
                        ),
                        { status: 413 }
                    )
                )
            }
        }).pipe(tempFileWriteStream)

        tempFileWriteStream.on('close', () => resolve(contentLength))
        tempFileWriteStream.on('error', reject)
    })
}

//
// Validation

/**
 * Type guard for SupportedMethods.
 */
function isSupportedMethod(method: string): method is SupportedMethods {
    return (SUPPORTED_METHODS as Set<string>).has(method)
}

/**
 * Throws an error with status 422 if the method is invalid.
 */
function checkMethod(method: string): void {
    if (!isSupportedMethod(method)) {
        throw Object.assign(new Error(`Method must be one of ${Array.from(SUPPORTED_METHODS.keys()).join(', ')}`), {
            status: 422,
        })
    }
}

/**
 * Throws an error with status 400 if the repository is invalid.
 */
function checkRepository(repository: any): void {
    if (typeof repository !== 'string') {
        throw Object.assign(new Error('Must specify the repository (usually of the form github.com/user/repo)'), {
            status: 400,
        })
    }
}

/**
 * Throws an error with status 400 if the commit is invalid.
 */
function checkCommit(commit: any): void {
    if (typeof commit !== 'string' || commit.length !== 40 || !/^[0-9a-f]+$/.test(commit)) {
        throw Object.assign(new Error('Must specify the commit as a 40 character hash ' + commit), { status: 400 })
    }
}

/**
 * Throws an error with status 413 if the content length is too large.
 */
function checkContentLength(rawContentLength: string | undefined): void {
    if (rawContentLength && parseInt(rawContentLength || '', 10) > MAX_FILE_SIZE) {
        throw Object.assign(
            new Error(
                `The size of the given LSIF file (${rawContentLength} bytes) exceeds the max of ${MAX_FILE_SIZE}`
            ),
            { status: 413 }
        )
    }
}

main()
