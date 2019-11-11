import * as constants from '../../shared/constants'
import * as fs from 'mz/fs'
import * as path from 'path'
import * as settings from '../settings'
import bodyParser from 'body-parser'
import express from 'express'
import pTimeout from 'p-timeout'
import uuid from 'uuid'
import { addTags, logAndTraceCall, TracingContext } from '../../shared/tracing'
import { Backend, ReferencePaginationCursor } from '../backend/backend'
import { encodeCursor, parseCursor } from '../pagination/cursor'
import { enqueue } from '../../shared/queue/queue'
import { extractLimit } from '../pagination/limit-offset'
import { Logger } from 'winston'
import { nextLink } from '../pagination/link'
import { pipeline as _pipeline } from 'stream'
import { promisify } from 'util'
import { Queue } from 'bull'
import { Span, Tracer } from 'opentracing'
import { wrap } from 'async-middleware'

const pipeline = promisify(_pipeline)

/**
 * Create a tracing context from the request logger and tracing span
 * tagged with the given values.
 *
 * @param logger The logger instance.
 * @param req The express request.
 * @param tags The tags to apply to the logger and span.
 */
const createTracingContext = (
    logger: Logger,
    req: express.Request & { span?: Span },
    tags: { [K: string]: unknown }
): TracingContext => addTags({ logger, span: req.span }, tags)

/**
 * Adds a trailing slash to a root unless it refers to the top level.
 *
 * - 'foo' -> 'foo/'
 * - 'foo/' -> 'foo/'
 * - '/' -> ''
 * - '' -> ''
 */
const normalizeRoot = (root: unknown): string => {
    if (root === undefined || root === '/' || root === '') {
        return ''
    }

    if (typeof root !== 'string') {
        throw Object.assign(new Error('root must be a string'), {
            status: 422,
        })
    }

    return root.endsWith('/') ? root : root + '/'
}

/**
 * Throws an error with status 400 if the repository string is invalid.
 */
const validateRepository = (repository: unknown): void => {
    if (typeof repository !== 'string') {
        throw Object.assign(new Error(`Must specify a repository ${repository}`), {
            status: 400,
        })
    }
}

/**
 * Throws an error with status 400 if the commit string is invalid.
 */
const validateCommit = (commit: unknown): void => {
    if (typeof commit !== 'string' || commit.length !== 40 || !/^[0-9a-f]+$/.test(commit)) {
        throw Object.assign(new Error(`Must specify the commit as a 40 character hash ${commit}`), { status: 400 })
    }
}

/**
 * Throws an error with status 400 if the file is not present.
 */
const validateFile = (file: unknown): void => {
    if (typeof file !== 'string') {
        throw Object.assign(new Error(`Must specify a file ${file}`), { status: 400 })
    }
}

/**
 * Throws an error with status 422 if the requested method is not supported.
 */
const validateMethod = (method: string, supportedMethods: string[]): void => {
    if (!supportedMethods.includes(method)) {
        throw Object.assign(new Error(`Method must be one of ${Array.from(supportedMethods).join(', ')}`), {
            status: 422,
        })
    }
}

/**
 * Create a router containing the LSIF upload and query endpoints.
 *
 * @param backend The backend instance.
 * @param queue The queue containing LSIF jobs.
 * @param logger The logger instance.
 * @param tracer The tracer instance.
 */
export function createLsifRouter(
    backend: Backend,
    queue: Queue,
    logger: Logger,
    tracer: Tracer | undefined
): express.Router {
    const router = express.Router()

    router.post(
        '/upload',
        wrap(
            async (req: express.Request & { span?: Span }, res: express.Response): Promise<void> => {
                const { repository, commit, root: rootRaw, blocking, maxWait } = req.query
                const root = normalizeRoot(rootRaw)
                const timeout = parseInt(maxWait, 10) || 0
                validateRepository(repository)
                validateCommit(commit)

                const ctx = createTracingContext(logger, req, {
                    repository,
                    commit,
                    root,
                })
                const filename = path.join(settings.STORAGE_ROOT, constants.UPLOADS_DIR, uuid.v4())
                const output = fs.createWriteStream(filename)

                try {
                    await logAndTraceCall(ctx, 'uploading dump', () => pipeline(req, output))
                } catch (e) {
                    throw Object.assign(e, { status: 422 })
                }

                // Enqueue convert job
                logger.debug('enqueueing convert job', { repository, commit, root })
                const args = { repository, commit, root, filename }
                const job = await enqueue(queue, 'convert', args, {}, tracer, ctx.span)

                if (blocking) {
                    let promise = job.finished()
                    if (timeout >= 0) {
                        promise = pTimeout(promise, timeout * 1000)
                    }

                    try {
                        await promise

                        // Job succeeded while blocked, send success
                        res.status(200)
                        res.send({ id: job.id })
                        return
                    } catch (error) {
                        // Throw the job error, but not the timeout error
                        if (!error.message.includes('Promise timed out')) {
                            throw error
                        }
                    }
                }

                // Job will complete asynchronously, send a 202: Accepted with
                // the job id so that the client can continue to track the progress
                // asynchronously.

                res.status(202)
                res.send({ id: job.id })
            }
        )
    )

    router.post(
        '/exists',
        bodyParser.json({ limit: '1mb' }),
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { repository, commit, file } = req.query
                validateRepository(repository)
                validateCommit(commit)
                validateFile(file)

                const ctx = createTracingContext(logger, req, { repository, commit })
                res.json(await backend.exists(repository, commit, file, ctx))
            }
        )
    )

    router.post(
        '/request',
        bodyParser.json({ limit: '1mb' }),
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { repository, commit } = req.query
                const { path: filePath, position, method } = req.body
                validateRepository(repository)
                validateCommit(commit)
                validateMethod(method, ['definitions', 'references', 'hover'])

                const ctx = createTracingContext(logger, req, { repository, commit })

                switch (method) {
                    case 'definitions':
                        res.json(await backend.definitions(repository, commit, filePath, position, ctx))
                        break

                    case 'hover':
                        res.json(await backend.hover(repository, commit, filePath, position, ctx))
                        break

                    case 'references': {
                        const { cursor: cursorRaw } = req.query
                        const limit = extractLimit(req, settings.DEFAULT_REFERENCES_NUM_REMOTE_DUMPS)
                        const startCursor = parseCursor<ReferencePaginationCursor>(cursorRaw)

                        const { locations, cursor: endCursor } = await backend.references(
                            repository,
                            commit,
                            filePath,
                            position,
                            { limit, cursor: startCursor },
                            ctx
                        )

                        const encodedCursor = encodeCursor<ReferencePaginationCursor>(endCursor)
                        if (!encodedCursor) {
                            res.set('Link', nextLink(req, { limit, cursor: encodedCursor }))
                        }

                        res.json(locations)
                        break
                    }
                }
            }
        )
    )

    return router
}
