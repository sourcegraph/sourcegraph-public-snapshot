import * as constants from '../../shared/constants'
import * as fs from 'mz/fs'
import * as nodepath from 'path'
import * as settings from '../settings'
import * as validation from '../middleware/validation'
import bodyParser from 'body-parser'
import express from 'express'
import uuid from 'uuid'
import { addTags, logAndTraceCall, TracingContext } from '../../shared/tracing'
import { Backend, ReferencePaginationCursor } from '../backend/backend'
import { checkSchema, ParamSchema } from 'express-validator'
import { encodeCursor } from '../pagination/cursor'
import { Logger } from 'winston'
import { lsp } from 'lsif-protocol'
import { nextLink } from '../pagination/link'
import { pipeline as _pipeline } from 'stream'
import { promisify } from 'util'
import { Span, Tracer } from 'opentracing'
import { wrap } from 'async-middleware'
import { extractLimitOffset } from '../pagination/limit-offset'
import { UploadsManager } from '../../shared/uploads/uploads'
import { InternalLocation } from '../backend/database'

const pipeline = promisify(_pipeline)

/**
 * Create a router containing the LSIF upload and query endpoints.
 *
 * @param backend The backend instance.
 * @param uploadsManager The uploads manager instance.
 * @param logger The logger instance.
 * @param tracer The tracer instance.
 */
export function createLsifRouter(
    backend: Backend,
    uploadsManager: UploadsManager,
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
        repository: string
        commit: string
        root?: string
        blocking?: boolean
        maxWait?: number
    }

    router.post(
        '/upload',
        validation.validationMiddleware([
            validation.validateNonEmptyString('repository'),
            validation.validateNonEmptyString('commit').matches(commitPattern),
            validation.validateOptionalString('root'),
            validation.validateOptionalBoolean('blocking'),
            validation.validateOptionalInt('maxWait'),
        ]),
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { repository, commit, root: rootRaw, blocking, maxWait }: UploadQueryArgs = req.query
                const root = sanitizeRoot(rootRaw)
                const ctx = createTracingContext(req, { repository, commit, root })
                const filename = nodepath.join(settings.STORAGE_ROOT, constants.UPLOADS_DIR, uuid.v4())
                const output = fs.createWriteStream(filename)
                await logAndTraceCall(ctx, 'Uploading dump', () => pipeline(req, output))

                // Add upload record
                const upload = await uploadsManager.enqueue({ repository, commit, root, filename }, tracer, ctx.span)

                if (blocking) {
                    logger.debug('Blocking on upload conversion', { repository, commit, root })

                    if (await uploadsManager.waitForUploadToConvert(upload.id, maxWait)) {
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
        repository: string
        commit: string
        path: string
    }

    router.get(
        '/exists',
        validation.validationMiddleware([
            validation.validateNonEmptyString('repository'),
            validation.validateNonEmptyString('commit').matches(commitPattern),
            validation.validateNonEmptyString('path'),
        ]),
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { repository, commit, path }: ExistsQueryArgs = req.query
                const ctx = createTracingContext(req, { repository, commit })
                const dump = await backend.exists(repository, commit, path, undefined, ctx)
                res.json({ dump })
            }
        )
    )

    interface FilePositionArgs {
        repository: string
        commit: string
        path: string
        line: number
        character: number
        dumpId?: number
    }

    router.get(
        '/definitions',
        validation.validationMiddleware([
            validation.validateNonEmptyString('repository'),
            validation.validateNonEmptyString('commit'),
            validation.validateNonEmptyString('path'),
            validation.validateInt('line'),
            validation.validateInt('character'),
            validation.validateOptionalInt('dumpId'),
        ]),
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { repository, commit, path, line, character, dumpId }: FilePositionArgs = req.query
                const ctx = createTracingContext(req, { repository, commit, path })

                const locations = await backend.definitions(repository, commit, path, { line, character }, dumpId, ctx)
                if (locations === undefined) {
                    throw Object.assign(new Error('LSIF dump not found'), { status: 404 })
                }

                res.send({
                    locations: locations.map(l => ({
                        repository: l.dump.repository,
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
            validation.validateNonEmptyString('repository'),
            validation.validateNonEmptyString('commit'),
            validation.validateNonEmptyString('path'),
            validation.validateInt('line'),
            validation.validateInt('character'),
            validation.validateOptionalInt('dumpId'),
            validation.validateLimit,
            validation.validateCursor<ReferencePaginationCursor>(),
        ]),
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { repository, commit, path, line, character, dumpId, cursor }: ReferencesQueryArgs = req.query
                const { limit } = extractLimitOffset(req.query, settings.DEFAULT_REFERENCES_NUM_REMOTE_DUMPS)
                const ctx = createTracingContext(req, { repository, commit, path })

                const result = await backend.references(
                    repository,
                    commit,
                    path,
                    { line, character },
                    { limit, cursor },
                    dumpId,
                    ctx
                )
                if (result === undefined) {
                    throw Object.assign(new Error('LSIF dump not found'), { status: 404 })
                }

                const { locations, cursor: endCursor } = result
                const encodedCursor = encodeCursor<ReferencePaginationCursor>(endCursor)
                if (encodedCursor) {
                    res.set('Link', nextLink(req, { limit, cursor: encodedCursor }))
                }

                res.json({
                    locations: locations.map(l => ({
                        repository: l.dump.repository,
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
            validation.validateNonEmptyString('repository'),
            validation.validateNonEmptyString('commit'),
            validation.validateNonEmptyString('path'),
            validation.validateInt('line'),
            validation.validateInt('character'),
            validation.validateOptionalInt('dumpId'),
        ]),
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { repository, commit, path, line, character, dumpId }: FilePositionArgs = req.query
                const ctx = createTracingContext(req, { repository, commit, path })

                const result = await backend.hover(repository, commit, path, { line, character }, dumpId, ctx)
                if (result === undefined) {
                    throw Object.assign(new Error('LSIF dump not found'), { status: 404 })
                }

                res.json(result)
            }
        )
    )

    //
    // Legacy Endpoints

    interface ExistsQueryArgs {
        repository: string
        commit: string
        file: string
    }

    router.post(
        '/exists',
        validation.validationMiddleware([
            validation.validateNonEmptyString('repository'),
            validation.validateNonEmptyString('commit').matches(commitPattern),
            validation.validateNonEmptyString('file'),
        ]),
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { repository, commit, file }: ExistsQueryArgs = req.query
                const ctx = createTracingContext(req, { repository, commit })
                const dump = await backend.exists(repository, commit, file, undefined, ctx)
                res.json(dump !== undefined)
            }
        )
    )

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

    const requestBodySchema: Record<string, ParamSchema> = {
        path: { isString: true, isEmpty: { negated: true } },
        'position.line': { isInt: true },
        'position.character': { isInt: true },
        method: { isIn: { options: [['definitions', 'references', 'hover']] } },
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
                const ctx = createTracingContext(req, { repository, commit })

                switch (method) {
                    case 'definitions': {
                        const result = await backend.definitions(repository, commit, filePath, position, undefined, ctx)
                        if (result === undefined) {
                            res.status(404).send()
                            return
                        }

                        res.json(result.map(loc => internalLocationToLocation(repository, loc)))
                        break
                    }

                    case 'hover': {
                        const result = await backend.hover(repository, commit, filePath, position, undefined, ctx)
                        if (result === undefined) {
                            res.status(404).send()
                            return
                        }

                        res.json(result && { contents: result.text })
                        break
                    }

                    case 'references': {
                        const result = await backend.references(
                            repository,
                            commit,
                            filePath,
                            position,
                            { limit, cursor },
                            undefined,
                            ctx
                        )

                        if (result === undefined) {
                            res.status(404).send()
                            return
                        }

                        const { locations, cursor: endCursor } = result
                        const encodedCursor = encodeCursor<ReferencePaginationCursor>(endCursor)
                        if (encodedCursor) {
                            res.set('Link', nextLink(req, { limit, cursor: encodedCursor }))
                        }

                        res.json(locations.map(loc => internalLocationToLocation(repository, loc)))
                        break
                    }
                }
            }
        )
    )

    return router
}

/**
 * Convert an `InternalLocation` to an LSP location object. The URI of the resulting
 * location object will be a relative if the dump describes a location in the source
 * repository and wil be an absolute URI otherwise.
 *
 * @param repository The source repository.
 * @param location The location object.
 */
export function internalLocationToLocation(repository: string, { dump, path, range }: InternalLocation): lsp.Location {
    if (dump.repository !== repository) {
        const url = new URL(`git://${dump.repository}`)
        url.search = dump.commit
        url.hash = path
        path = url.href
    }

    return lsp.Location.create(path, range)
}
