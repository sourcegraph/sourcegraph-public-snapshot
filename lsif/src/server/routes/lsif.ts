import * as constants from '../../shared/constants'
import * as fs from 'mz/fs'
import * as lsp from 'vscode-languageserver-protocol'
import * as nodepath from 'path'
import * as settings from '../settings'
import * as validation from '../../shared/api/middleware/validation'
import express from 'express'
import * as uuid from 'uuid'
import { addTags, logAndTraceCall, TracingContext } from '../../shared/tracing'
import { Backend } from '../backend/backend'
import { encodeCursor } from '../../shared/api/pagination/cursor'
import { Logger } from 'winston'
import { nextLink } from '../../shared/api/pagination/link'
import { pipeline as _pipeline } from 'stream'
import { promisify } from 'util'
import { Span, Tracer } from 'opentracing'
import { wrap } from 'async-middleware'
import { extractLimitOffset } from '../../shared/api/pagination/limit-offset'
import { UploadManager } from '../../shared/store/uploads'
import { readGzippedJsonElementsFromFile } from '../../shared/input'
import * as lsif from 'lsif-protocol'
import { ReferencePaginationCursor } from '../backend/cursor'
import { LsifUpload } from '../../shared/models/pg'

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
        indexerName?: string
    }

    interface UploadResponse {
        id: number
    }

    router.post(
        '/upload',
        validation.validationMiddleware([
            validation.validateInt('repositoryId'),
            validation.validateNonEmptyString('commit').matches(commitPattern),
            validation.validateOptionalString('root'),
            validation.validateOptionalString('indexerName'),
        ]),
        wrap(
            async (req: express.Request, res: express.Response<UploadResponse>): Promise<void> => {
                const { repositoryId, commit, root: rootRaw, indexerName }: UploadQueryArgs = req.query

                const root = sanitizeRoot(rootRaw)
                const ctx = createTracingContext(req, { repositoryId, commit, root })
                const filename = nodepath.join(settings.STORAGE_ROOT, constants.UPLOADS_DIR, uuid.v4())
                const output = fs.createWriteStream(filename)
                await logAndTraceCall(ctx, 'Uploading dump', () => pipeline(req, output))

                const indexer = indexerName || (await findIndexer(filename))
                if (!indexer) {
                    throw new Error('Could not find tool type in metadata vertex at the start of the dump.')
                }

                // Add upload record
                const upload = await uploadManager.enqueue(
                    { repositoryId, commit, root, filename, indexer },
                    tracer,
                    ctx.span
                )

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

    interface ExistsResponse {
        uploads: LsifUpload[]
    }

    router.get(
        '/exists',
        validation.validationMiddleware([
            validation.validateInt('repositoryId'),
            validation.validateNonEmptyString('commit').matches(commitPattern),
            validation.validateNonEmptyString('path'),
        ]),
        wrap(
            async (req: express.Request, res: express.Response<ExistsResponse>): Promise<void> => {
                const { repositoryId, commit, path }: ExistsQueryArgs = req.query
                const ctx = createTracingContext(req, { repositoryId, commit })
                const uploads = await backend.exists(repositoryId, commit, path, ctx)
                res.json({ uploads })
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

    interface LocationsResponse {
        locations: { repositoryId: number; commit: string; path: string; range: lsp.Range }[]
    }

    router.get(
        '/definitions',
        validation.validationMiddleware([
            validation.validateInt('repositoryId'),
            validation.validateNonEmptyString('commit'),
            validation.validateNonEmptyString('path'),
            validation.validateInt('line'),
            validation.validateInt('character'),
            validation.validateInt('uploadId'),
        ]),
        wrap(
            async (req: express.Request, res: express.Response<LocationsResponse>): Promise<void> => {
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
            validation.validateInt('uploadId'),
            validation.validateLimit,
            validation.validateCursor<ReferencePaginationCursor>(),
        ]),
        wrap(
            async (req: express.Request, res: express.Response<LocationsResponse>): Promise<void> => {
                const { repositoryId, commit, path, line, character, uploadId, cursor }: ReferencesQueryArgs = req.query
                const { limit } = extractLimitOffset(req.query, settings.DEFAULT_REFERENCES_PAGE_SIZE)
                const ctx = createTracingContext(req, { repositoryId, commit, path })

                const result = await backend.references(
                    repositoryId,
                    commit,
                    path,
                    { line, character },
                    { limit, cursor },
                    constants.DEFAULT_REFERENCES_REMOTE_DUMP_LIMIT,
                    uploadId,
                    ctx
                )
                if (result === undefined) {
                    throw Object.assign(new Error('LSIF upload not found'), { status: 404 })
                }

                const { locations, newCursor } = result
                const encodedCursor = encodeCursor<ReferencePaginationCursor>(newCursor)
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

    type HoverResponse = { text: string; range: lsp.Range } | null

    router.get(
        '/hover',
        validation.validationMiddleware([
            validation.validateInt('repositoryId'),
            validation.validateNonEmptyString('commit'),
            validation.validateNonEmptyString('path'),
            validation.validateInt('line'),
            validation.validateInt('character'),
            validation.validateInt('uploadId'),
        ]),
        wrap(
            async (req: express.Request, res: express.Response<HoverResponse>): Promise<void> => {
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

/**
 * Read and decode the first entry of the dump. If the entry exists, encodes a metadata vertex,
 * and contains a tool info name field, return the contents of that field; otherwise undefined.
 *
 * @param filename The filename to read.
 */
async function findIndexer(filename: string): Promise<string | undefined> {
    for await (const element of readGzippedJsonElementsFromFile(filename) as AsyncIterable<lsif.Vertex | lsif.Edge>) {
        if (element.type === lsif.ElementTypes.vertex && element.label === lsif.VertexLabels.metaData) {
            return element.toolInfo?.name
        }
        break
    }

    return undefined
}
