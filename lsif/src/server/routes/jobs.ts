import * as settings from '../settings'
import express from 'express'
import { ApiJobState, QUEUE_PREFIX, queueTypes, statesByQueue } from '../../shared/queue/queue'
import { chunk } from 'lodash'
import { Job, Queue } from 'bull'
import { limitOffset } from '../pagination/limit-offset'
import { nextLink } from '../pagination/link'
import { ScriptedRedis } from '../redis/redis'
import { wrap } from 'async-middleware'

/**
 * The representation of a job as returned by the API.
 */
interface ApiJob {
    id: string
    jobType: string
    arguments: object
    state: ApiJobState
    failure: { summary: string; stacktraces: string[] } | null
    queuedAt: string
    startedAt: string | null
    completedOrErroredAt: string | null
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
 * Format a job to return from the API.
 *
 * @param payload The job's JSON payload.
 */
const formatJobInternal = ({
    id,
    name,
    data,
    failedReason,
    stacktrace,
    timestamp,
    finishedOn,
    processedOn,
    state,
}: {
    id: string | number
    name: string
    data: any
    failedReason: any
    stacktrace: string[] | null
    timestamp: number
    finishedOn: number | null
    processedOn: number | null
    state: ApiJobState
}): ApiJob => {
    const failure =
        state === 'errored' && typeof failedReason === 'string' && stacktrace
            ? { summary: failedReason, stacktraces: stacktrace }
            : null

    return {
        id: `${id}`,
        jobType: name,
        arguments: data.args,
        state,
        failure,
        queuedAt: toDate(timestamp),
        startedAt: toMaybeDate(processedOn),
        completedOrErroredAt: toMaybeDate(finishedOn),
    }
}

/**
 * Format a job to return from the API.
 *
 * @param job The job to format.
 * @param state The job's state.
 */
const formatJob = (job: Job, state: ApiJobState): ApiJob => formatJobInternal({ state, ...job.toJSON() })

/**
 * Format a job to return from the API.
 *
 * @param values A map of values composing the job.
 * @param state The job's state.
 */
const formatJobFromMap = (values: Map<string, string>, state: ApiJobState): ApiJob => {
    const rawData = values.get('data')
    const rawStacktrace = values.get('stacktrace')
    const rawTimestamp = values.get('timestamp')
    const rawProcessedOn = values.get('processedOn')
    const rawFinishedOn = values.get('finishedOn')

    return formatJobInternal({
        id: values.get('id') || '',
        name: values.get('name') || '',
        data: JSON.parse(rawData || '{}'),
        failedReason: values.get('failedReason') || null,
        stacktrace: rawStacktrace ? JSON.parse(rawStacktrace) : null,
        timestamp: rawTimestamp ? parseInt(rawTimestamp, 10) : 0,
        processedOn: rawProcessedOn ? parseInt(rawProcessedOn, 10) : null,
        finishedOn: rawFinishedOn ? parseInt(rawFinishedOn, 10) : null,
        state,
    })
}

/**
 * Create a router containing the job endpoints.
 *
 * @param queue The queue instance.
 * @param scriptedClient The Redis client with scripts loaded.
 */
export function createJobRouter(queue: Queue, scriptedClient: ScriptedRedis): express.Router {
    const router = express.Router()

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

    router.get(
        `/jobs/:state(${Array.from(queueTypes.keys()).join('|')})`,
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { state } = req.params as { state: ApiJobState }
                const { query } = req.query
                const { limit, offset } = limitOffset(req, settings.DEFAULT_JOB_PAGE_SIZE)

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
