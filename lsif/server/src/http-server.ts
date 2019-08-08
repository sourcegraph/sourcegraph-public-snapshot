import * as Prometheus from 'prom-client'
import * as tmp from 'tmp-promise'
import bodyParser from 'body-parser'
import express from 'express'
import { Backend, ERRNOLSIFDATA, QueryRunner } from './backend'
import { Cache } from './cache'
import { createPrometheusReporters, emit } from './prometheus'
import { DgraphBackend } from './dgraph'
import { fs } from 'mz'
import { readEnvInt } from './env'
import { SQLiteBlobBackend, SQLiteGraphBackend } from './sqlite'
import { wrap } from 'async-middleware'
import {
    checkRepository,
    checkCommit,
    checkContentLength,
    validateContent,
    checkMethod,
    MAX_FILE_SIZE,
} from './http-server.validation'

/**
 * Which port to run the LSIF server on. Defaults to 3186.
 */
const PORT = readEnvInt({ key: 'LSIF_HTTP_PORT', defaultValue: 3186 })

/**
 * A list of available backend factories by name.
 */
const AVAILABLE_BACKENDS: { [k: string]: () => Promise<Backend<QueryRunner>> } = {
    'sqlite-graph': async () => new SQLiteGraphBackend(),
    'sqlite-blob': async () => new SQLiteBlobBackend(),
    dgraph: async () => {
        const db = new DgraphBackend()
        await db.initialize()
        return db
    },
}

/**
 * The name of the backend that the server initializes on startup.
 */
const DEFAULT_BACKEND = process.env['DEFAULT_BACKEND'] || 'sqlite-blob'

/**
 * Whether or not JSON-schema validation is performed when uploading LSIF dumps.
 */
const VALIDATE_INPUT = Boolean(process.env['LSIF_VALIDATE_INPUT'])

/**
 * Runs the HTTP server which accepts LSIF dump uploads and responds to LSIF requests.
 */
async function main(): Promise<void> {
    const app = express()
    const cache = new Cache()
    const schema = JSON.parse((await fs.readFile('./src/lsif.schema.json')).toString())
    const prometheusReporters = createPrometheusReporters()

    // Initialize default backend
    let currentBackend = DEFAULT_BACKEND
    let backend = await AVAILABLE_BACKENDS[currentBackend]()

    app.use(errorHandler)

    app.get('/ping', (req, res) => {
        res.send({ pong: 'pong' })
    })

    app.get('/metrics', (req, res) => {
        res.set('Content-Type', Prometheus.register.contentType)
        res.end(Prometheus.register.metrics())
    })

    app.get(
        '/backend',
        wrap(async (req, res) => {
            res.send({ data: currentBackend })
        })
    )

    app.post(
        '/backend',
        wrap(async (req, res) => {
            const { backendName } = req.query
            backend.close()
            currentBackend = backendName
            backend = await AVAILABLE_BACKENDS[backendName]()
            cache.reset()
            res.send({ data: null })
        })
    )

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
                const contentLength = await readContent(req, tempFile.path)

                if (VALIDATE_INPUT) {
                    await validateContent(tempFile.path, schema)
                }

                const { insertStats } = await backend.insertDump(tempFile.path, repository, commit, contentLength)

                // Bust the cache
                cache.delete(repository, commit)

                // Emit metrics
                emit(prometheusReporters, insertStats)

                res.send({
                    data: null,
                    stats: {
                        insertStats: insertStats,
                    },
                })
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

                // Emit metrics
                emit(prometheusReporters, cacheStats)
                if (createRunnerStats) {
                    emit(prometheusReporters, createRunnerStats)
                }

                res.send({
                    data: result,
                    stats: {
                        cacheStats: cacheStats,
                        createRunnerStats: createRunnerStats,
                    },
                })
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
            checkMethod(method, backend.availableQueries())

            try {
                const {
                    result: { result, queryStats },
                    cacheStats,
                } = await cache.withDB(backend, repository, commit, async queryRunner => {
                    return await queryRunner.query(method, path, position)
                })

                // Emit metrics
                emit(prometheusReporters, cacheStats)
                emit(prometheusReporters, queryStats)

                res.json({
                    data: result || null,
                    stats: {
                        cacheStats: cacheStats,
                        queryStats: queryStats,
                    },
                })
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
 * Read the request body into the given file path and return the contenLength
 * of the input stream.
 */
async function readContent(req: express.Request, tempPath: string): Promise<number> {
    let contentLength = 0
    const tempFileWriteStream = fs.createWriteStream(tempPath)

    const validateSize = (line: string, reject: (_: any) => void) => {
        contentLength += line.length
        if (contentLength > MAX_FILE_SIZE) {
            reject(
                Object.assign(
                    new Error(
                        `The size of the given LSIF file (${contentLength} bytes so far) exceeds the max of ${MAX_FILE_SIZE}`
                    ),
                    { status: 413 }
                )
            )
        }
    }

    try {
        await new Promise((resolve, reject) => {
            req.on('data', chunk => {
                validateSize(chunk, reject)
            }).pipe(tempFileWriteStream)

            tempFileWriteStream.on('close', () => resolve(contentLength))
            tempFileWriteStream.on('error', reject)
        })
    } catch (e) {
        tempFileWriteStream.destroy()
    }

    return contentLength
}

main().catch(console.error)
