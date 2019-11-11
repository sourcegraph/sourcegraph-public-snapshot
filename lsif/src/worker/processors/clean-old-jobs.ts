import * as settings from '../settings'
import { JobStatusClean, Queue } from 'bull'
import { logAndTraceCall, TracingContext } from '../../shared/tracing'
import { Logger } from 'winston'

const cleanStatuses: JobStatusClean[] = ['completed', 'wait', 'active', 'delayed', 'failed']

/**
 * Create a job that removes all job data older than `JOB_MAX_AGE`.
 *
 * @param queue The queue.
 * @param logger The logger instance.
 */
export const createCleanOldJobsProcessor = (queue: Queue, logger: Logger) => async (
    _: unknown,
    ctx: TracingContext
): Promise<void> => {
    const removedJobs = await logAndTraceCall(ctx, 'cleaning old jobs', () =>
        Promise.all(cleanStatuses.map(status => queue.clean(settings.JOB_MAX_AGE * 1000, status)))
    )

    const { logger: jobLogger = logger } = ctx

    for (const [status, count] of removedJobs.map((jobs, i) => [cleanStatuses[i], jobs.length])) {
        if (count > 0) {
            jobLogger.debug('cleaned old jobs', { status, count })
        }
    }
}
