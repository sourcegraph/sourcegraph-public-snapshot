import * as constants from '../../shared/constants'
import * as fs from 'mz/fs'
import * as path from 'path'
import * as settings from '../settings'
import * as validation from '../middleware/validation'
import bodyParser from 'body-parser'
import express from 'express'
import pTimeout from 'p-timeout'
import uuid from 'uuid'
import { addTags, logAndTraceCall, TracingContext } from '../../shared/tracing'
import { Backend, ReferencePaginationCursor } from '../backend/backend'
import { checkSchema, ParamSchema } from 'express-validator'
import { encodeCursor } from '../pagination/cursor'
import { enqueue } from '../../shared/queue/queue'
import { Logger } from 'winston'
import { lsp } from 'lsif-protocol'
import { nextLink } from '../pagination/link'
import { pipeline as _pipeline } from 'stream'
import { promisify } from 'util'
import { Queue } from 'bull'
import { Span, Tracer } from 'opentracing'
import { wrap } from 'async-middleware'
import { extractLimitOffset } from '../pagination/limit-offset'

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

    // Used to validate commit hashes are 40 hex digits
    const commitPattern = /^[a-f0-9]{40}$/

    /**
     * Ensure roots end with a slash, unless it refers to the top-level directory.
     *
     * @param root The input root.
     */
    const sanitizeRoot = (root: string | undefined): string => {
        if (root === undefined || root === '/' || root === '') {
            return ''
        }

        return root.endsWith('/') ? root : root + '/'
    }

    interface UploadQueryArgs {
        repository: string
        commit: string
        root: string
        blocking: boolean
        maxWait: number
    }

    router.post(
        '/upload',
        validation.validationMiddleware([
            validation.validateNonEmptyString('repository'),
            validation.validateNonEmptyString('commit').matches(commitPattern),
            validation.validateOptionalString('root').customSanitizer(sanitizeRoot),
            validation.validateOptionalBoolean('blocking'),
            validation.validateOptionalInt('maxWait'),
        ]),
        wrap(
            async (req: express.Request & { span?: Span }, res: express.Response): Promise<void> => {
                const { repository, commit, root, blocking, maxWait }: UploadQueryArgs = req.query
                const ctx = createTracingContext(logger, req, { repository, commit, root })
                const filename = path.join(settings.STORAGE_ROOT, constants.UPLOADS_DIR, uuid.v4())
                const output = fs.createWriteStream(filename)
                await logAndTraceCall(ctx, 'uploading dump', () => pipeline(req, output))

                // Enqueue convert job
                logger.debug('enqueueing convert job', { repository, commit, root })
                const args = { repository, commit, root, filename }
                const job = await enqueue(queue, 'convert', args, {}, tracer, ctx.span)

                if (blocking) {
                    let promise = job.finished()
                    if (!isNaN(maxWait)) {
                        promise = pTimeout(promise, maxWait * 1000)
                    }

                    try {
                        await promise

                        // Job succeeded while blocked, send success
                        res.status(200).send({ id: job.id })
                        return
                    } catch (error) {
                        // Throw a job error, if one occurred. If we caught a timeout
                        // just fall-through and return a 202 response. The user can
                        // check the progress asynchronously with subsequent API calls.
                        if (!error.message.includes('Promise timed out')) {
                            throw error
                        }
                    }
                }

                // Job will complete asynchronously, send a 202: Accepted with
                // the job id so that the client can continue to track the progress
                // asynchronously.

                res.status(202).send({ id: job.id })
            }
        )
    )

    interface ExistsQueryArgs {
        repository: string
        commit: string
        file: string
    }

    router.post(
        '/exists',
        bodyParser.json({ limit: '1mb' }),
        validation.validationMiddleware([
            validation.validateNonEmptyString('repository'),
            validation.validateNonEmptyString('commit').matches(commitPattern),
            validation.validateNonEmptyString('file'),
        ]),
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { repository, commit, file }: ExistsQueryArgs = req.query
                const ctx = createTracingContext(logger, req, { repository, commit })
                res.json(await backend.exists(repository, commit, file, ctx))
            }
        )
    )

    const requestBodySchema: Record<string, ParamSchema> = {
        path: { isString: true, isEmpty: { negated: true } },
        'position.line': { isInt: true },
        'position.character': { isInt: true },
        method: { isIn: { options: [['definitions', 'references', 'hover']] } },
    }

    interface RequestQueryArgs {
        repository: string
        commit: string
        cursor: ReferencePaginationCursor | undefined
    }

    interface RequestBodyArgs {
        path: string
        position: lsp.Position
        method: string
    }

    router.post(
        '/request',
        bodyParser.json({ limit: '1mb' }),
        validation.validationMiddleware([
            validation.validateNonEmptyString('repository'),
            validation.validateNonEmptyString('commit').matches(commitPattern),
            validation.validateLimit,
            validation.validateCursor<ReferencePaginationCursor>(),
            ...checkSchema(requestBodySchema, ['body']),
        ]),
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { repository, commit, cursor }: RequestQueryArgs = req.query
                const { path: filePath, position, method }: RequestBodyArgs = req.body
                const { limit } = extractLimitOffset(req.query, settings.DEFAULT_REFERENCES_NUM_REMOTE_DUMPS)
                const ctx = createTracingContext(logger, req, { repository, commit })

                switch (method) {
                    case 'definitions':
                        res.json(await backend.definitions(repository, commit, filePath, position, ctx))
                        break

                    case 'references': {
                        const { locations, cursor: endCursor } = await backend.references(
                            repository,
                            commit,
                            filePath,
                            position,
                            { limit, cursor },
                            ctx
                        )

                        const encodedCursor = encodeCursor<ReferencePaginationCursor>(endCursor)
                        if (!encodedCursor) {
                            res.set('Link', nextLink(req, { limit, cursor: encodedCursor }))
                        }

                        res.json(locations)
                        break
                    }

                    case 'hover':
                        res.json(await backend.hover(repository, commit, filePath, position, ctx))
                        break
                }
            }
        )
    )

    return router
}
