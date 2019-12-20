import * as pgModels from '../../shared/models/pg'
import * as settings from '../settings'
import * as validation from '../middleware/validation'
import express from 'express'
import { nextLink } from '../pagination/link'
import { wrap } from 'async-middleware'
import { extractLimitOffset } from '../pagination/limit-offset'
import { UploadManager } from '../../shared/store/uploads'

/**
 * Create a router containing the upload endpoints.
 *
 * @param uploadManager The uploads manager instance.
 */
export function createUploadRouter(uploadManager: UploadManager): express.Router {
    const router = express.Router()

    router.get(
        '/uploads/stats',
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                res.send(await uploadManager.getCounts())
            }
        )
    )

    interface UploadsQueryArgs {
        query: string
    }

    router.get(
        '/uploads/:state(queued|completed|errored|processing)',
        validation.validationMiddleware([
            validation.validateQuery,
            validation.validateLimit,
            validation.validateOffset,
        ]),
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { query }: UploadsQueryArgs = req.query
                const { limit, offset } = extractLimitOffset(req.query, settings.DEFAULT_UPLOAD_PAGE_SIZE)
                const { uploads, totalCount } = await uploadManager.getUploads(
                    req.params.state as pgModels.LsifUploadState,
                    query,
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

    router.get(
        '/uploads/:id([0-9]+)',
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
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
            async (req: express.Request, res: express.Response): Promise<void> => {
                if (await uploadManager.deleteUpload(parseInt(req.params.id, 10))) {
                    res.status(204).send()
                    return
                }

                throw Object.assign(new Error('Upload not found'), {
                    status: 404,
                })
            }
        )
    )

    return router
}
