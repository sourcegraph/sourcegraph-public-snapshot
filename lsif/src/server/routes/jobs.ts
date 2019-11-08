import express from 'express'
import { chunk } from 'lodash'
import { formatJob, formatJobFromMap } from '../../api-job'
import { Logger } from 'winston'
import { pipeline as _pipeline } from 'stream'
import { Queue } from 'bull'
import { Tracer } from 'opentracing'
import { wrap } from 'async-middleware'
import { queueTypes, QUEUE_PREFIX, statesByQueue, ApiJobState } from '../../queue'
import { nextLink } from '../pagination/link'
import { limitOffset } from '../pagination/limit-offset'
import { ScriptedRedis } from '../redis/redis'
import { MAX_JOB_SEARCH, DEFAULT_JOB_PAGE_SIZE } from '../settings'

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
                const { limit, offset } = limitOffset(req, DEFAULT_JOB_PAGE_SIZE)

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
                        MAX_JOB_SEARCH,
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
