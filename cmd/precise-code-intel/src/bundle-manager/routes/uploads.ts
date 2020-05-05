import express from 'express'
import { Logger } from 'winston'
import { Span } from 'opentracing'
import { wrap } from 'async-middleware'
import { addTags, TracingContext, logAndTraceCall } from '../../shared/tracing'
import { pipeline as _pipeline } from 'stream'
import { promisify } from 'util'
import * as fs from 'mz/fs'
import * as settings from '../settings'
import { dbFilename, uploadFilename } from '../../shared/paths'
import { ThrottleGroup, Throttle } from 'stream-throttle'

const pipeline = promisify(_pipeline)

/**
 * Create a router containing the upload endpoints.
 *
 * @param logger The logger instance.
 */
export function createUploadRouter(logger: Logger): express.Router {
    const router = express.Router()

    const makeServeThrottle = makeThrottleFactory(
        settings.MAXIMUM_SERVE_BYTES_PER_SECOND,
        settings.MAXIMUM_SERVE_CHUNK_BYTES
    )

    const makeUploadThrottle = makeThrottleFactory(
        settings.MAXIMUM_UPLOAD_BYTES_PER_SECOND,
        settings.MAXIMUM_UPLOAD_CHUNK_BYTES
    )

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
        '/uploads/:id([0-9]+)',
        wrap(
            async (req: express.Request, res: express.Response<unknown>): Promise<void> => {
                const id = parseInt(req.params.id, 10)
                const ctx = createTracingContext(req, { id })
                const filename = uploadFilename(settings.STORAGE_ROOT, id)
                const stream = fs.createReadStream(filename)
                await logAndTraceCall(ctx, 'Serving payload', () => pipeline(stream, makeServeThrottle(), res))
            }
        )
    )

    router.post(
        '/uploads/:id([0-9]+)',
        wrap(
            async (req: express.Request, res: express.Response<unknown>): Promise<void> => {
                const id = parseInt(req.params.id, 10)
                const ctx = createTracingContext(req, { id })
                const filename = uploadFilename(settings.STORAGE_ROOT, id)
                const stream = fs.createWriteStream(filename)
                await logAndTraceCall(ctx, 'Uploading payload', () => pipeline(req, makeUploadThrottle(), stream))
                res.send()
            }
        )
    )

    router.post(
        '/dbs/:id([0-9]+)',
        wrap(
            async (req: express.Request, res: express.Response<unknown>): Promise<void> => {
                const id = parseInt(req.params.id, 10)
                const ctx = createTracingContext(req, { id })
                const filename = dbFilename(settings.STORAGE_ROOT, id)
                const stream = fs.createWriteStream(filename)
                await logAndTraceCall(ctx, 'Uploading payload', () => pipeline(req, makeUploadThrottle(), stream))
                res.send()
            }
        )
    )

    return router
}

/**
 * Create a function that will create a throttle that can be used as a stream
 * transformer. This transformer can limit both readable and writable streams.
 *
 * @param rate The maximum bit second of the stream.
 * @param chunksize The size of chunks used to break down larger slices of data.
 */
function makeThrottleFactory(rate: number, chunksize: number): () => Throttle {
    const opts = { rate, chunksize }
    const throttleGroup = new ThrottleGroup(opts)
    return () => throttleGroup.throttle(opts)
}
