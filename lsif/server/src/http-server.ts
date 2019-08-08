import * as Prometheus from 'prom-client'
import * as tmp from 'tmp-promise'
import bodyParser from 'body-parser'
import express from 'express'
import split2 from 'split2'
import through2 from 'through2'
import { Backend, ERRNOLSIFDATA, QueryRunner } from './backend'
import { Cache } from './cache'
import { createPrometheusReporters, emit } from './prometheus'
import { DgraphBackend } from './dgraph'
import { fs } from 'mz'
import { readEnvInt } from './env'
import { SQLiteBlobBackend, SQLiteGraphBackend } from './sqlite'
import { Validator } from 'jsonschema'
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
 * Limit the size of each line for a JSON-line encoded LSIF dump. Defaults to 1MB.
 */
const MAX_LINE_SIZE = readEnvInt({ key: 'LSIF_MAX_LINE_SIZE', defaultValue: 1024 * 1024 })

/**
 * Whether or not JSON-schema validation is performed when uploading LSIF dumps.
 */
const VALIDATE_INPUT = Boolean(process.env['LSIF_VALIDATE_INPUT'])

/**
 * List of supported LSIF methods that can be passed to query runners.
 */
type SupportedMethods = 'hover' | 'definitions' | 'references'

const SUPPORTED_METHODS: Set<SupportedMethods> = new Set(['hover', 'definitions', 'references'])

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
const DEFAULT_BACKEND = process.env['DEFAULT_BACKEND'] || 'sqlite-graph'

/**
 * Runs the HTTP server which accepts LSIF dump uploads and responds to LSIF requests.
 */
async function main(): Promise<void> {
    const app = express()
    const cache = new Cache()
    const schema = JSON.parse((await fs.readFile('./src/lsif.schema.json')).toString())
    const prometheusReporters = createPrometheusReporters()

    // Choose a default backend
    let backend = await AVAILABLE_BACKENDS[DEFAULT_BACKEND]()

    app.use(errorHandler)

    app.get('/ping', (req, res) => {
        res.send({ pong: 'pong' })
    })

    app.get('/metrics', (req, res) => {
        res.set('Content-Type', Prometheus.register.contentType)
        res.end(Prometheus.register.metrics())
    })

    app.post(
        '/switch-backend',
        wrap(async (req, res) => {
            const { backendName } = req.query
            backend.close()
            backend = await AVAILABLE_BACKENDS[backendName]()
            cache.reset()
            res.send({ status: 'ok' })
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
                const contentLength = VALIDATE_INPUT
                    ? await readAndValidateContent(req, tempFile.path, schema)
                    : await readContent(req, tempFile.path)

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
            checkMethod(method)

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

//
// TODO(efritz) - figure out where validation should live. It is VERY slow to do
// in process here.
//

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

/**
 * Like readContent, but also validate each JSON line against the JSON schema on disk.
 */
async function readAndValidateContent(req: express.Request, tempPath: string, schema: any): Promise<number> {
    let lineno = 0
    let contentLength = 0
    const validator = new Validator()
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

    const validateJSON = (line: string, lineno: number, reject: (_: any) => void) => {
        let data: any
        try {
            data = JSON.parse(line)
        } catch (e) {
            Object.assign(new Error(`Malformed JSON on line ${lineno}`), { status: 422 })
        }

        const result = validator.validate(data, schema)
        if (result.errors.length > 0) {
            reject(
                Object.assign(
                    // TODO(efritz) - better validation response
                    new Error(`Invalid JSON data on line ${lineno} ${result.errors.join(', ')}`),
                    { status: 422 }
                )
            )
        }
    }

    try {
        await new Promise((resolve, reject) => {
            const appendNewLine = through2((data, _, callback) => {
                callback(null, new Buffer(data + '\n'))
            })

            req.pipe(split2({ maxLength: MAX_LINE_SIZE }))
                .on('data', line => {
                    if (VALIDATE_INPUT) {
                        validateJSON(line, lineno, reject)
                    }

                    validateSize(line, reject)
                    lineno++
                })
                .pipe(appendNewLine)
                .pipe(tempFileWriteStream)

            tempFileWriteStream.on('close', () => resolve(contentLength))
            tempFileWriteStream.on('error', reject)
        })
    } catch (e) {
        tempFileWriteStream.destroy()
    }

    return contentLength
}

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

main().catch(console.error)
