import * as constants from '../../shared/constants'
import * as fs from 'mz/fs'
import * as nodepath from 'path'
import * as settings from '../settings'
import * as validation from '../middleware/validation'
import express from 'express'
import uuid from 'uuid'
import { addTags, logAndTraceCall, TracingContext } from '../../shared/tracing'
import { Backend, ReferencePaginationCursor } from '../backend/backend'
import { encodeCursor } from '../pagination/cursor'
import { Logger } from 'winston'
import { nextLink } from '../pagination/link'
import { pipeline as _pipeline } from 'stream'
import { promisify } from 'util'
import { Span, Tracer } from 'opentracing'
import { wrap } from 'async-middleware'
import { extractLimitOffset } from '../pagination/limit-offset'
import { UploadManager } from '../../shared/store/uploads'

const pipeline = promisify(_pipeline)

/**
 * Create a router containing the LSIF upload and query endpoints.
 *
 * @param backend The backend instance.
 * @param uploadManager The uploads manager instance.
 * @param logger The logger instance.
 * @param tracer The tracer instance.
 */
export function createLsifRouter(
    backend: Backend,
    uploadManager: UploadManager,
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

    /**
     * Create a tracing context from the request logger and tracing span
     * tagged with the given values.
     *
     * @param req The express request.
     * @param tags The tags to apply to the logger and span.
     */
    const createTracingContext = (
        req: express.Request & { span?: Span },
        tags: { [K: string]: unknown }
    ): TracingContext => addTags({ logger, span: req.span }, tags)

    interface UploadQueryArgs {
        repositoryId: number
        commit: string
        root?: string
        blocking?: boolean
        maxWait?: number
    }

    router.post(
        '/upload',
        validation.validationMiddleware([
            validation.validateInt('repositoryId'),
            validation.validateNonEmptyString('commit').matches(commitPattern),
            validation.validateOptionalString('root'),
            validation.validateOptionalBoolean('blocking'),
            validation.validateOptionalInt('maxWait'),
        ]),
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { repositoryId, commit, root: rootRaw, blocking, maxWait }: UploadQueryArgs = req.query
                const root = sanitizeRoot(rootRaw)
                const ctx = createTracingContext(req, { repositoryId, commit, root })
                const filename = nodepath.join(settings.STORAGE_ROOT, constants.UPLOADS_DIR, uuid.v4())
                const output = fs.createWriteStream(filename)
                await logAndTraceCall(ctx, 'Uploading dump', () => pipeline(req, output))

                // Add upload record
                const upload = await uploadManager.enqueue({ repositoryId, commit, root, filename }, tracer, ctx.span)

                if (blocking) {
                    logger.debug('Blocking on upload conversion', { repositoryId, commit, root })

                    if (await uploadManager.waitForUploadToConvert(upload.id, maxWait)) {
                        // Upload converted successfully while blocked, send success
                        res.status(200).send({ id: upload.id })
                        return
                    }
                }

                // Upload conversion will complete asynchronously, send an accepted response
                // with the upload id so that the client can continue to track the progress
                // asynchronously.
                res.status(202).send({ id: upload.id })
            }
        )
    )

    interface ExistsQueryArgs {
        repositoryId: number
        commit: string
        path: string
    }

    router.get(
        '/exists',
        validation.validationMiddleware([
            validation.validateInt('repositoryId'),
            validation.validateNonEmptyString('commit').matches(commitPattern),
            validation.validateNonEmptyString('path'),
        ]),
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { repositoryId, commit, path }: ExistsQueryArgs = req.query
                const ctx = createTracingContext(req, { repositoryId, commit })
                const upload = await backend.exists(repositoryId, commit, path, undefined, ctx)
                res.json({ upload })
            }
        )
    )

    interface FilePositionArgs {
        repositoryId: number
        commit: string
        path: string
        line: number
        character: number
        uploadId?: number
    }

    router.get(
        '/definitions',
        validation.validationMiddleware([
            validation.validateInt('repositoryId'),
            validation.validateNonEmptyString('commit'),
            validation.validateNonEmptyString('path'),
            validation.validateInt('line'),
            validation.validateInt('character'),
            validation.validateOptionalInt('uploadId'),
        ]),
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { repositoryId, commit, path, line, character, uploadId }: FilePositionArgs = req.query
                const ctx = createTracingContext(req, { repositoryId, commit, path })

                const locations = await backend.definitions(
                    repositoryId,
                    commit,
                    path,
                    { line, character },
                    uploadId,
                    ctx
                )
                if (locations === undefined) {
                    throw Object.assign(new Error('LSIF upload not found'), { status: 404 })
                }

                res.send({
                    locations: locations.map(l => ({
                        repositoryId: l.dump.repositoryId,
                        commit: l.dump.commit,
                        path: l.path,
                        range: l.range,
                    })),
                })
            }
        )
    )

    interface ReferencesQueryArgs extends FilePositionArgs {
        commit: string
        cursor: ReferencePaginationCursor | undefined
    }

    router.get(
        '/references',
        validation.validationMiddleware([
            validation.validateInt('repositoryId'),
            validation.validateNonEmptyString('commit'),
            validation.validateNonEmptyString('path'),
            validation.validateInt('line'),
            validation.validateInt('character'),
            validation.validateOptionalInt('uploadId'),
            validation.validateLimit,
            validation.validateCursor<ReferencePaginationCursor>(),
        ]),
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { repositoryId, commit, path, line, character, uploadId, cursor }: ReferencesQueryArgs = req.query
                const { limit } = extractLimitOffset(req.query, settings.DEFAULT_REFERENCES_NUM_REMOTE_DUMPS)
                const ctx = createTracingContext(req, { repositoryId, commit, path })

                const result = await backend.references(
                    repositoryId,
                    commit,
                    path,
                    { line, character },
                    { limit, cursor },
                    uploadId,
                    ctx
                )
                if (result === undefined) {
                    throw Object.assign(new Error('LSIF upload not found'), { status: 404 })
                }

                const { locations, cursor: endCursor } = result
                const encodedCursor = encodeCursor<ReferencePaginationCursor>(endCursor)
                if (encodedCursor) {
                    res.set('Link', nextLink(req, { limit, cursor: encodedCursor }))
                }

                res.json({
                    locations: locations.map(l => ({
                        repositoryId: l.dump.repositoryId,
                        commit: l.dump.commit,
                        path: l.path,
                        range: l.range,
                    })),
                })
            }
        )
    )

    router.get(
        '/hover',
        validation.validationMiddleware([
            validation.validateInt('repositoryId'),
            validation.validateNonEmptyString('commit'),
            validation.validateNonEmptyString('path'),
            validation.validateInt('line'),
            validation.validateInt('character'),
            validation.validateOptionalInt('uploadId'),
        ]),
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { repositoryId, commit, path, line, character, uploadId }: FilePositionArgs = req.query
                const ctx = createTracingContext(req, { repositoryId, commit, path })

                const result = await backend.hover(repositoryId, commit, path, { line, character }, uploadId, ctx)
                if (result === undefined) {
                    throw Object.assign(new Error('LSIF upload not found'), { status: 404 })
                }

                res.json(result)
            }
        )
    )

    return router
}
