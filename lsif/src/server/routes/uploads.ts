import * as xrepoModels from '../../shared/models/xrepo'
import * as settings from '../settings'
import * as validation from '../middleware/validation'
import express from 'express'
import { nextLink } from '../pagination/link'
import { wrap } from 'async-middleware'
import { extractLimitOffset } from '../pagination/limit-offset'
import { Queue } from '../../shared/uploads/uploads'
import { omit } from 'lodash'

/**
 * The representation of an upload as returned by the API.
 */
type Upload = Omit<xrepoModels.LsifUpload, 'failureSummary' | 'failureStacktrace'> & {
    failure: { summary: string; stacktrace: string } | null
}

/**
 * Format a upload to return from the API.
 *
 * @param upload The upload to format.
 */
const formatUpload = (upload: xrepoModels.LsifUpload): Upload => ({
    ...omit(upload, 'failureSummary', 'failureStacktrace'),
    failure:
        upload.state === 'errored'
            ? {
                  summary: upload.failureSummary,
                  stacktrace: upload.failureStacktrace,
              }
            : null,
})

/**
 * Create a router containing the upload endpoints.
 *
 * @param queue The queue instance.
 */
export function createUploadRouter(queue: Queue): express.Router {
    const router = express.Router()

    router.get(
        '/uploads/stats',
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                res.send(await queue.getCounts())
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
                const { uploads, totalCount } = await queue.getUploads(
                    req.params.state as xrepoModels.LsifUploadState,
                    query,
                    limit,
                    offset
                )

                if (offset + uploads.length < totalCount) {
                    res.set('Link', nextLink(req, { limit, offset: offset + uploads.length }))
                }

                res.json({ uploads: uploads.map(formatUpload), totalCount })
            }
        )
    )

    router.get(
        '/uploads/:id([0-9]+)',
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const upload = await queue.getUpload(parseInt(req.params.id, 10))
                if (upload) {
                    res.send(formatUpload(upload))
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
                if (await queue.deleteUpload(parseInt(req.params.id, 10))) {
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
