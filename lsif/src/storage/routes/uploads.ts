import express from 'express'
import { Logger } from 'winston'
import { Span } from 'opentracing'
import { wrap } from 'async-middleware'
import { addTags, TracingContext, logAndTraceCall } from '../../shared/tracing'
import { pipeline as _pipeline } from 'stream'
import { promisify } from 'util'
import * as nodepath from 'path'
import * as fs from 'mz/fs'
import * as settings from '../settings'
import * as constants from '../../shared/constants'
import { dbFilename } from '../../shared/paths'

const pipeline = promisify(_pipeline)

/**
 * Create a router containing the upload endpoints.
 *
 * @param logger The logger instance.
 */
export function createUploadRouter(logger: Logger): express.Router {
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

    router.get(
        '/:id([0-9a-fA-F-]+)/raw',
        wrap(
            async (req: express.Request, res: express.Response<undefined>): Promise<void> => {
                const id = req.params.id
                const filename = nodepath.join(settings.STORAGE_ROOT, constants.UPLOADS_DIR, id)
                const output = fs.createReadStream(filename)
                output.pipe(res) // TODO - does not forward errors
            }
        )
    )

    router.post(
        '/:id([0-9a-fA-F-]+)/raw',
        wrap(
            async (req: express.Request, res: express.Response<undefined>): Promise<void> => {
                const id = req.params.id
                const ctx = createTracingContext(req, { id })

                const filename = nodepath.join(settings.STORAGE_ROOT, constants.UPLOADS_DIR, id)
                const output = fs.createWriteStream(filename)
                await logAndTraceCall(ctx, 'Uploading payload', () => pipeline(req, output))

                res.send()
            }
        )
    )

    router.delete(
        '/:id([0-9a-fA-F-]+)/raw',
        wrap(
            async (req: express.Request, res: express.Response<undefined>): Promise<void> => {
                const id = req.params.id
                const filename = nodepath.join(settings.STORAGE_ROOT, constants.UPLOADS_DIR, id)
                await fs.unlink(filename)
                res.send()
            }
        )
    )

    router.post(
        '/:id([0-9]+)',
        wrap(
            async (req: express.Request, res: express.Response<undefined>): Promise<void> => {
                const id = parseInt(req.params.id, 10)
                const ctx = createTracingContext(req, { id })

                const output = fs.createWriteStream(dbFilename(settings.STORAGE_ROOT, id))
                await logAndTraceCall(ctx, 'Uploading payload', () => pipeline(req, output))

                // try {
                //     // TODO - make this a helper method
                //     await fs.unlink(nodepath.join(settings.STORAGE_ROOT, constants.UPLOADS_DIR, `${id}`))
                // } catch (error) {
                //     if (!(error && error.code === 'ENOENT')) {
                //         throw error
                //     }
                // }

                res.send()
            }
        )
    )

    router.delete(
        '/:id([0-9]+)',
        wrap(
            async (req: express.Request, res: express.Response<undefined>): Promise<void> => {
                const id = parseInt(req.params.id, 10)
                await fs.unlink(dbFilename(settings.STORAGE_ROOT, id))
                res.send()
            }
        )
    )

    return router
}
