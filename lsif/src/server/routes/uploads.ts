import * as pgModels from '../../shared/models/pg'
import * as settings from '../settings'
import * as validation from '../../shared/api/middleware/validation'
import express from 'express'
import { nextLink } from '../../shared/api/pagination/link'
import { wrap } from 'async-middleware'
import { extractLimitOffset } from '../../shared/api/pagination/limit-offset'
import { UploadManager, LsifUploadWithPlaceInQueue } from '../../shared/store/uploads'
import { DumpManager } from '../../shared/store/dumps'
import { EntityManager } from 'typeorm'
import { SRC_FRONTEND_INTERNAL } from '../../shared/config/settings'
import { TracingContext, addTags } from '../../shared/tracing'
import { Span } from 'opentracing'
import { Logger } from 'winston'
import { updateCommitsAndDumpsVisibleFromTip } from '../../shared/visibility'

/**
 * Create a router containing the upload endpoints.
 *
 * @param dumpManager The dumps manager instance.
 * @param uploadManager The uploads manager instance.
 * @param logger The logger instance.
 */
export function createUploadRouter(
    dumpManager: DumpManager,
    uploadManager: UploadManager,
    logger: Logger
): express.Router {
    const router = express.Router()

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

    interface UploadsQueryArgs {
        query: string
        state?: pgModels.LsifUploadState
        visibleAtTip?: boolean
    }

    type UploadResponse = LsifUploadWithPlaceInQueue

    router.get(
        '/uploads/:id([0-9]+)',
        wrap(
            async (req: express.Request, res: express.Response<UploadResponse>): Promise<void> => {
                const upload = await uploadManager.getUpload(parseInt(req.params.id, 10))
                if (upload) {
                    res.send(upload)
                    return
                }

                throw Object.assign(new Error('Upload not found'), {
                    status: 404,
                })
            }
        )
    )

    router.delete(
        '/uploads/:id([0-9]+)',
        wrap(
            async (req: express.Request, res: express.Response<never>): Promise<void> => {
                const id = parseInt(req.params.id, 10)
                const ctx = createTracingContext(req, { id })

                const updateVisibility = (entityManager: EntityManager, repositoryId: number): Promise<void> =>
                    updateCommitsAndDumpsVisibleFromTip({
                        entityManager,
                        dumpManager,
                        frontendUrl: SRC_FRONTEND_INTERNAL,
                        repositoryId,
                        ctx,
                    })

                if (await uploadManager.deleteUpload(id, updateVisibility)) {
                    res.status(204).send()
                    return
                }

                throw Object.assign(new Error('Upload not found'), {
                    status: 404,
                })
            }
        )
    )

    interface UploadsResponse {
        uploads: LsifUploadWithPlaceInQueue[]
        totalCount: number
    }

    router.get(
        '/uploads/repository/:id([0-9]+)',
        validation.validationMiddleware([
            validation.validateQuery,
            validation.validateLsifUploadState,
            validation.validateOptionalBoolean('visibleAtTip'),
            validation.validateLimit,
            validation.validateOffset,
        ]),
        wrap(
            async (req: express.Request, res: express.Response<UploadsResponse>): Promise<void> => {
                const { query, state, visibleAtTip }: UploadsQueryArgs = req.query
                const { limit, offset } = extractLimitOffset(req.query, settings.DEFAULT_UPLOAD_PAGE_SIZE)
                const { uploads, totalCount } = await uploadManager.getUploads(
                    parseInt(req.params.id, 10),
                    state,
                    query,
                    !!visibleAtTip,
                    limit,
                    offset
                )

                if (offset + uploads.length < totalCount) {
                    res.set('Link', nextLink(req, { limit, offset: offset + uploads.length }))
                }

                res.json({ uploads, totalCount })
            }
        )
    )

    return router
}
