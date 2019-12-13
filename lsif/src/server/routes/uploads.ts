import * as xrepoModels from '../../shared/models/xrepo'
import * as settings from '../settings'
import * as validation from '../middleware/validation'
import express from 'express'
import { nextLink } from '../pagination/link'
import { wrap } from 'async-middleware'
import { extractLimitOffset } from '../pagination/limit-offset'
import { UploadsManager } from '../../shared/uploads/uploads'

/**
 * Create a router containing the upload endpoints.
 *
 * @param uploadsManager The uploads manager instance.
 */
export function createUploadRouter(uploadsManager: UploadsManager): express.Router {
    const router = express.Router()

    router.get(
        '/uploads/stats',
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                res.send(await uploadsManager.getCounts())
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
                const { uploads, totalCount } = await uploadsManager.getUploads(
                    req.params.state as xrepoModels.LsifUploadState,
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
                const upload = await uploadsManager.getUpload(parseInt(req.params.id, 10))
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
                if (await uploadsManager.deleteUpload(parseInt(req.params.id, 10))) {
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
