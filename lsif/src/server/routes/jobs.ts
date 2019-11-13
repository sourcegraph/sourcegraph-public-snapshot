import * as settings from '../settings'
import * as validation from '../middleware/validation'
import express from 'express'
import { addTags, TracingContext } from '../../shared/tracing'
import { ApiJobState, enqueue, QUEUE_PREFIX, queueTypes, statesByQueue } from '../../shared/queue/queue'
import { checkSchema, ParamSchema } from 'express-validator'
import { chunk } from 'lodash'
import { Job, Queue } from 'bull'
import { Logger } from 'winston'
import { nextLink } from '../pagination/link'
import { ScriptedRedis } from '../redis/redis'
import { Span, Tracer } from 'opentracing'
import { wrap } from 'async-middleware'
import bodyParser from 'body-parser'
import { waitForJob } from '../jobs/blocking'
import { extractLimitOffset } from '../pagination/limit-offset'

/**
 * The representation of a job as returned by the API.
 */
interface ApiJob {
    id: string
    name: string
    args: object
    state: ApiJobState
    progress: number
    failedReason: string | null
    stacktrace: string[] | null
    timestamp: string
    processedOn: string | null
    finishedOn: string | null
}

/**
 * Convert a timestamp into an ISO string.
 *
 * @param timestamp The millisecond POSIX timestamp.
 */
const toDate = (timestamp: number): string => new Date(timestamp).toISOString()

/**
 * Attempt to convert a timestamp into an ISO string.
 *
 * @param timestamp The millisecond POSIX timestamp.
 */
const toMaybeDate = (timestamp: number | null): string | null => (timestamp ? toDate(timestamp) : null)

/**
 * Attempt to convert a string into an integer.
 *
 * @param value The int-y string.
 */
const toMaybeInt = (value: string | undefined): number | null => (value ? parseInt(value, 10) : null)

/**
 * Format a job to return from the API.
 *
 * @param job The job to format.
 * @param state The job's state.
 */
const formatJob = (job: Job, state: ApiJobState): ApiJob => {
    const payload = job.toJSON()

    return {
        id: `${payload.id}`,
        name: payload.name,
        args: payload.data.args,
        state,
        progress: payload.progress,
        failedReason: payload.failedReason,
        stacktrace: payload.stacktrace,
        timestamp: toDate(payload.timestamp),
        processedOn: toMaybeDate(payload.processedOn),
        finishedOn: toMaybeDate(payload.finishedOn),
    }
}

/**
 * Format a job to return from the API.
 *
 * @param values A map of values composing the job.
 * @param state The job's state.
 */
const formatJobFromMap = (values: Map<string, string>, state: ApiJobState): ApiJob => {
    const rawData = values.get('data')
    const rawStacktrace = values.get('stacktrace')

    return {
        id: values.get('id') || '',
        name: values.get('name') || '',
        args: rawData ? JSON.parse(rawData).args : {},
        state,
        progress: toMaybeInt(values.get('progress')) || 0,
        failedReason: values.get('failedReason') || null,
        stacktrace: rawStacktrace ? JSON.parse(rawStacktrace) : null,
        timestamp: toMaybeDate(toMaybeInt(values.get('timestamp'))) || '',
        processedOn: toMaybeDate(toMaybeInt(values.get('processedOn'))),
        finishedOn: toMaybeDate(toMaybeInt(values.get('finishedOn'))),
    }
}

/**
 * Create a router containing the job endpoints.
 *
 * @param queue The queue instance.
 * @param scriptedClient The Redis client with scripts loaded.
 * @param logger The logger instance.
 * @param tracer The tracer instance.
 */
export function createJobRouter(
    queue: Queue,
    scriptedClient: ScriptedRedis,
    logger: Logger,
    tracer: Tracer | undefined
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

    interface EnqueueQueryArgs {
        blocking: boolean
        maxWait: number
    }

    interface EnqueueBodyArgs {
        name: string
    }

    const enqueueBodySchema: Record<string, ParamSchema> = {
        name: { isIn: { options: [['update-tips', 'clean-old-jobs', 'clean-failed-jobs']] } },
    }

    router.post(
        '/jobs',
        bodyParser.json({ limit: '1mb' }),
        validation.validationMiddleware([
            validation.validateOptionalBoolean('blocking'),
            validation.validateOptionalInt('maxWait'),
            ...checkSchema(enqueueBodySchema, ['body']),
        ]),
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { blocking, maxWait }: EnqueueQueryArgs = req.query
                const { name }: EnqueueBodyArgs = req.body

                // Enqueue job
                const ctx = createTracingContext(req, { name })
                logger.debug(`enqueueing ${name} job`)
                const job = await enqueue(queue, name, {}, {}, tracer, ctx.span)

                if (blocking && (await waitForJob(job, maxWait))) {
                    // Job succeeded while blocked, send success
                    res.status(200).send({ id: job.id })
                    return
                }

                // Job will complete asynchronously, send an accepted response with
                // the job id so that the client can continue to track the progress
                // asynchronously.
                res.status(202).send({ id: job.id })
            }
        )
    )

    router.get(
        '/jobs/stats',
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const counts = await queue.getJobCounts()

                res.send({
                    processingCount: counts.active,
                    erroredCount: counts.failed,
                    completedCount: counts.completed,
                    queuedCount: counts.waiting,
                    scheduledCount: counts.delayed,
                })
            }
        )
    )

    interface JobsQueryArgs {
        query: string
    }

    router.get(
        `/jobs/:state(${Array.from(queueTypes.keys()).join('|')})`,
        validation.validationMiddleware([
            validation.validateQuery,
            validation.validateLimit,
            validation.validateOffset,
        ]),
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { state } = req.params as { state: ApiJobState }
                const { query }: JobsQueryArgs = req.query
                const { limit, offset } = extractLimitOffset(req.query, settings.DEFAULT_JOB_PAGE_SIZE)

                const queueName = queueTypes.get(state)
                if (!queueName) {
                    throw new Error(`Unknown job state ${state}`)
                }

                if (!query) {
                    const rawJobs = await queue.getJobs([queueName], offset, offset + limit - 1)
                    const jobs = rawJobs.map(job => formatJob(job, state))
                    const totalCount = (await queue.getJobCountByTypes([queueName])) as never

                    if (offset + jobs.length < totalCount) {
                        res.set('Link', nextLink(req, { limit, offset: offset + jobs.length }))
                    }

                    res.send({ jobs, totalCount })
                } else {
                    const [payloads, nextOffset] = await scriptedClient.searchJobs([
                        QUEUE_PREFIX,
                        queueName,
                        query,
                        offset,
                        limit,
                        settings.MAX_JOB_SEARCH,
                    ])

                    const jobs = payloads
                        // Convert each hgetall response into a map
                        .map(payload => new Map(chunk(payload, 2) as [string, string][]))
                        // Format each job
                        .map(payload => formatJobFromMap(payload, state))

                    if (nextOffset) {
                        res.set('Link', nextLink(req, { limit, offset: nextOffset }))
                    }

                    res.send({ jobs })
                }
            }
        )
    )

    router.get(
        '/jobs/:id',
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const job = await queue.getJob(req.params.id)
                if (!job) {
                    throw Object.assign(new Error('Job not found'), {
                        status: 404,
                    })
                }

                const rawState = await job.getState()
                const state = statesByQueue.get(rawState === 'waiting' ? 'wait' : rawState)
                if (!state) {
                    throw new Error(`Unknown job state ${state}.`)
                }

                res.send(formatJob(job, state))
            }
        )
    )

    return router
}
