import * as Prometheus from 'prom-client';
import * as tmp from 'tmp-promise';
import bodyParser from 'body-parser';
import express from 'express';
import { Cache } from './cache';
import { ERRNOLSIFDATA } from './backend';
import { fs } from 'mz';
import { JsonDatabase } from './ms/json';
import { noopTransformer } from './ms/database';
import { readEnvInt } from './env';
import { SQLiteGraphBackend } from './sqlite';
import { wrap } from 'async-middleware';

/**
 * Which port to run the LSIF server on. Defaults to 3186.
 */
const PORT = readEnvInt({ key: 'LSIF_HTTP_PORT', defaultValue: 3186 })

/**
 * Limit on the file size accepted by the /upload endpoint. Defaults to 100MB.
 */
const MAX_FILE_SIZE = readEnvInt({ key: 'LSIF_MAX_FILE_SIZE', defaultValue: 100 * 1024 * 1024 })

/**
 * An object that identifies a commit of a repository.
 */
interface RepositoryCommit {
    // An opaque repository ID.
    repository: string
    // A 40-character commit hash.
    commit: string
}

/**
 * List of supported `Database` methods.
 */
type SupportedMethods = 'hover' | 'definitions' | 'references'

const SUPPORTED_METHODS: Set<SupportedMethods> = new Set(['hover', 'definitions', 'references'])

/**
 * Runs the HTTP server which accepts LSIF file uploads and responds to
 * hover/defs/refs requests.
 */
function main(): void {
    const app = express()
    const backend = new SQLiteGraphBackend()
    const cache = new Cache()

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
            const key = cacheKey({ repository, commit })

            checkRepository(repository)
            checkCommit(commit)
            checkContentLength(req.header('Content-Length'))

            // Create temp file to receive the request body
            const tempFile = await tmp.file()

            try {
                // Read the content and ensure the body is a valid LSIF dump
                const contentLength = await readContent(req, tempFile.path)
                await new JsonDatabase().load(tempFile.path, () => noopTransformer)
                const { encodingStats } = await backend.createDB(tempFile.path, key, contentLength)

                // Bust the cache
                cache.delete(key)

                // TODO(efritz) - add stats to prometheus reporter
                res.send({
                    data: null,
                    stats: {
                        encodingStats: encodingStats,
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
            const key = cacheKey({ repository, commit })

            checkRepository(repository)
            checkCommit(commit)

            try {
                const { result, cacheStats, handleStats } = await cache.withDB(backend, key, async db => {
                    return !file || Boolean(db.stat(file))
                })

                // TODO(efritz) - add stats to prometheus reporter
                res.send({
                    data: result,
                    stats: {
                        cacheStats: cacheStats,
                        handleStats: handleStats,
                    },
                })
            } catch (e) {
                if ('code' in e && e.code === ERRNOLSIFDATA) {
                    // TODO(efritz) - add stats to prometheus reporter
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
            const key = cacheKey({ repository, commit })

            checkRepository(repository)
            checkCommit(commit)
            checkMethod(method)

            try {
                const {
                    result: { result, queryStats },
                    cacheStats,
                } = await cache.withDB(backend, key, async db => {
                    switch (method) {
                        case 'hover':
                            return backend.hover(db, path, position)
                        case 'definitions':
                            return backend.definitions(db, path, position)
                        case 'references':
                            return backend.references(db, path, position, { includeDeclaration: false })
                        default:
                            throw new Error(`Unknown method ${method}`)
                    }
                })

                // TODO(efritz) - add stats to prometheus reporter
                res.json({
                    data: result || null,
                    stats: {
                        cacheStats: cacheStats,
                        queryStats: queryStats,
                    },
                })
            } catch (e) {
                if ('code' in e && e.code === ERRNOLSIFDATA) {
                    throw Object.assign(new Error(`No LSIF data available for ${repository}@${commit}.`), {
                        status: 404,
                    })
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
 * Computes the cache key that contains LSIF data for the given repository@commit.
 */
function cacheKey({ repository, commit }: RepositoryCommit): string {
    const urlEncodedRepository = encodeURIComponent(repository)
    return `${urlEncodedRepository}@${commit}.lsif`
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
